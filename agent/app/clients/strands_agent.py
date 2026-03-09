from __future__ import annotations

import functools
import hashlib
import inspect
import json
import logging
import re
import time
from collections import OrderedDict
from collections.abc import Callable
from dataclasses import dataclass
from datetime import datetime, timezone
from threading import Lock
from typing import Any, Protocol

from strands import Agent, tool
from strands.session import RepositorySessionManager
from strands.types.content import Messages
from tenacity import (
    retry,
    retry_if_exception,
    stop_after_attempt,
    stop_after_delay,
    wait_exponential,
    wait_random,
)

from app.clients.conversation_manager import SafeSlidingWindowConversationManager
from app.clients.k8s import KubernetesClient
from app.clients.llm_providers import LLMProvider, ModelConfig, create_model
from app.clients.prometheus import PrometheusClient
from app.clients.session_repository import PostgresSessionRepository
from app.clients.strands_patch import apply_gemini_thought_signature_patch
from app.clients.tempo import TempoClient, build_traceql_query
from app.core.config import Settings
from app.core.masking import RegexMasker

logger = logging.getLogger(__name__)

_PROM_DURATION_RE = re.compile(r"^(?P<value>\d+)(?P<unit>ms|s|m|h|d|w|y)$")
_QUERY_PREVIEW_LIMIT = 120
_HASH_LENGTH = 12


def _compact_text(value: str | None, *, limit: int = _QUERY_PREVIEW_LIMIT) -> str:
    if not value:
        return ""
    compact = " ".join(value.split())
    if len(compact) <= limit:
        return compact
    return f"{compact[: limit - 3]}..."


def _hash_text(value: str | None) -> str:
    if not value:
        return ""
    digest = hashlib.sha256(value.encode("utf-8")).hexdigest()
    return digest[:_HASH_LENGTH]


def _safe_scalar(value: Any) -> Any:
    if value is None or isinstance(value, (bool, int, float)):
        return value
    if isinstance(value, str):
        return _compact_text(value)
    return type(value).__name__


def _parse_time(value: str | None) -> datetime | None:
    if not value:
        return None
    raw = value.strip()
    if not raw:
        return None
    try:
        if raw.replace(".", "", 1).isdigit():
            return datetime.fromtimestamp(float(raw), tz=timezone.utc)
        if raw.endswith("Z"):
            raw = f"{raw[:-1]}+00:00"
        parsed = datetime.fromisoformat(raw)
        if parsed.tzinfo is None:
            return parsed.replace(tzinfo=timezone.utc)
        return parsed.astimezone(timezone.utc)
    except ValueError:
        return None


def _seconds_between(start: str | None, end: str | None) -> float | None:
    start_time = _parse_time(start)
    end_time = _parse_time(end)
    if start_time is None or end_time is None:
        return None
    return max((end_time - start_time).total_seconds(), 0.0)


def _parse_prom_duration(value: str | None) -> float | None:
    if not value:
        return None
    raw = value.strip()
    if not raw:
        return None
    if raw.replace(".", "", 1).isdigit():
        return float(raw)

    match = _PROM_DURATION_RE.match(raw)
    if not match:
        return None

    amount = float(match.group("value"))
    multipliers = {
        "ms": 0.001,
        "s": 1.0,
        "m": 60.0,
        "h": 3600.0,
        "d": 86400.0,
        "w": 604800.0,
        "y": 31536000.0,
    }
    return amount * multipliers[match.group("unit")]


def _bind_arguments(
    fn: Callable[..., Any],
    args: tuple[Any, ...],
    kwargs: dict[str, Any],
) -> dict[str, Any]:
    try:
        bound = inspect.signature(fn).bind_partial(*args, **kwargs)
    except TypeError:
        return {}
    return dict(bound.arguments)


def _default_arg_summary(arguments: dict[str, Any]) -> dict[str, object]:
    return {key: _safe_scalar(value) for key, value in arguments.items()}


def _default_result_summary(result: Any) -> dict[str, object]:
    if isinstance(result, list):
        return {"result_type": "list", "item_count": len(result)}
    if isinstance(result, dict):
        summary: dict[str, object] = {
            "result_type": "dict",
            "keys": sorted(result.keys())[:8],
        }
        if isinstance(result.get("warning"), str):
            summary["warning"] = _compact_text(result["warning"])
        if isinstance(result.get("error"), str):
            summary["error"] = _compact_text(result["error"])
        return summary
    return {"result_type": type(result).__name__}


def _pod_lookup_summary(arguments: dict[str, Any]) -> dict[str, object]:
    summary = {
        "namespace": arguments.get("namespace"),
        "pod_name": arguments.get("pod_name"),
    }
    if "container" in arguments and arguments.get("container"):
        summary["container"] = arguments.get("container")
    if "tail_lines" in arguments and arguments.get("tail_lines") is not None:
        summary["tail_lines"] = arguments.get("tail_lines")
    if "since_seconds" in arguments and arguments.get("since_seconds") is not None:
        summary["since_seconds"] = arguments.get("since_seconds")
    return summary


def _manifest_summary(arguments: dict[str, Any]) -> dict[str, object]:
    summary = {
        "namespace": arguments.get("namespace"),
        "api_version": arguments.get("api_version"),
        "resource": arguments.get("resource"),
    }
    if "name" in arguments:
        summary["name"] = arguments.get("name")
    if "label_selector" in arguments and arguments.get("label_selector"):
        summary["label_selector"] = arguments.get("label_selector")
    if "limit" in arguments and arguments.get("limit") is not None:
        summary["limit"] = arguments.get("limit")
    return summary


def _list_pods_summary(arguments: dict[str, Any]) -> dict[str, object]:
    return {
        "namespace": arguments.get("namespace"),
        "label_selector": arguments.get("label_selector"),
    }


def _namespace_name_summary(arguments: dict[str, Any]) -> dict[str, object]:
    return {
        "namespace": arguments.get("namespace"),
        "name": arguments.get("name"),
    }


def _prometheus_query_summary(arguments: dict[str, Any]) -> dict[str, object]:
    query = arguments.get("query")
    summary: dict[str, object] = {
        "query_hash": _hash_text(query if isinstance(query, str) else None),
        "query_preview": _compact_text(query if isinstance(query, str) else None),
        "query_length": len(query) if isinstance(query, str) else 0,
    }
    if arguments.get("time"):
        summary["time"] = arguments.get("time")
    return summary


def _prometheus_range_summary(arguments: dict[str, Any]) -> dict[str, object]:
    query = arguments.get("query")
    start = arguments.get("start")
    end = arguments.get("end")
    step = arguments.get("step")
    return {
        "query_hash": _hash_text(query if isinstance(query, str) else None),
        "query_preview": _compact_text(query if isinstance(query, str) else None),
        "query_length": len(query) if isinstance(query, str) else 0,
        "start": start,
        "end": end,
        "step": step,
        "window_seconds": _seconds_between(
            start if isinstance(start, str) else None,
            end if isinstance(end, str) else None,
        ),
        "step_seconds": _parse_prom_duration(step if isinstance(step, str) else None),
    }


def _tempo_search_summary(arguments: dict[str, Any]) -> dict[str, object]:
    query = arguments.get("query")
    start = arguments.get("start")
    end = arguments.get("end")
    return {
        "service_name": arguments.get("service_name"),
        "namespace": arguments.get("namespace"),
        "query_hash": _hash_text(query if isinstance(query, str) else None),
        "query_preview": _compact_text(query if isinstance(query, str) else None),
        "query_length": len(query) if isinstance(query, str) else 0,
        "start": start,
        "end": end,
        "window_seconds": _seconds_between(
            start if isinstance(start, str) else None,
            end if isinstance(end, str) else None,
        ),
        "limit": arguments.get("limit"),
    }


def _log_result_summary(result: Any) -> dict[str, object]:
    summary = _default_result_summary(result)
    if not isinstance(result, list):
        return summary
    log_line_count = 0
    for item in result:
        if not isinstance(item, dict):
            continue
        logs = item.get("logs")
        if isinstance(logs, list):
            log_line_count += len(logs)
    summary["log_line_count"] = log_line_count
    return summary


def _prometheus_result_summary(result: Any) -> dict[str, object]:
    summary = _default_result_summary(result)
    if not isinstance(result, dict):
        return summary
    payload = result.get("data")
    if not isinstance(payload, dict):
        return summary
    data = payload.get("data")
    if not isinstance(data, dict):
        return summary
    series = data.get("result")
    if isinstance(series, list):
        summary["series_count"] = len(series)
    result_type = data.get("resultType")
    if isinstance(result_type, str):
        summary["prom_result_type"] = result_type
    return summary


def _prometheus_range_result_summary(result: Any) -> dict[str, object]:
    summary = _prometheus_result_summary(result)
    if not isinstance(result, dict):
        return summary
    payload = result.get("data")
    if not isinstance(payload, dict):
        return summary
    data = payload.get("data")
    if not isinstance(data, dict):
        return summary
    series = data.get("result")
    if not isinstance(series, list):
        return summary

    sample_count = 0
    for entry in series:
        if not isinstance(entry, dict):
            continue
        values = entry.get("values")
        if isinstance(values, list):
            sample_count += len(values)
    summary["sample_count"] = sample_count
    return summary


def _tempo_result_summary(result: Any) -> dict[str, object]:
    summary = _default_result_summary(result)
    if not isinstance(result, dict):
        return summary
    trace_count = result.get("trace_count")
    if isinstance(trace_count, int):
        summary["trace_count"] = trace_count
    return summary


def _log_tool_event(
    event: str,
    tool_name: str,
    *,
    elapsed_ms: float | None = None,
    args_summary: dict[str, object] | None = None,
    result_summary: dict[str, object] | None = None,
    error: BaseException | None = None,
) -> None:
    parts = [event, f"tool={tool_name}"]
    if elapsed_ms is not None:
        parts.append(f"elapsed_ms={elapsed_ms:.1f}")
    if args_summary:
        parts.append(
            "args=" + json.dumps(args_summary, ensure_ascii=True, sort_keys=True, default=str)
        )
    if result_summary:
        parts.append(
            "result="
            + json.dumps(result_summary, ensure_ascii=True, sort_keys=True, default=str)
        )
    if error is not None:
        parts.append(f"error_type={type(error).__name__}")
        parts.append(f'error="{_compact_text(str(error), limit=200)}"')
    logger.info(" ".join(parts))


def _logged_tool(
    *,
    arg_formatter: Callable[[dict[str, Any]], dict[str, object]] | None = None,
    result_formatter: Callable[[Any], dict[str, object]] | None = None,
) -> Callable[[Callable[..., Any]], Any]:
    def decorator(fn: Callable[..., Any]) -> Any:
        formatter = arg_formatter or _default_arg_summary
        result_summary = result_formatter or _default_result_summary

        @functools.wraps(fn)
        def wrapped(*args: Any, **kwargs: Any) -> Any:
            bound_arguments = _bind_arguments(fn, args, kwargs)
            args_summary = formatter(bound_arguments)
            _log_tool_event("tool_started", fn.__name__, args_summary=args_summary)
            start_time = time.perf_counter()
            try:
                result = fn(*args, **kwargs)
            except Exception as exc:
                elapsed_ms = (time.perf_counter() - start_time) * 1000
                _log_tool_event(
                    "tool_failed",
                    fn.__name__,
                    elapsed_ms=elapsed_ms,
                    args_summary=args_summary,
                    error=exc,
                )
                raise

            elapsed_ms = (time.perf_counter() - start_time) * 1000
            _log_tool_event(
                "tool_finished",
                fn.__name__,
                elapsed_ms=elapsed_ms,
                args_summary=args_summary,
                result_summary=result_summary(result),
            )
            return result

        return tool(wrapped)

    return decorator


def _iter_exception_chain(exc: BaseException) -> list[BaseException]:
    """Return chained exceptions (__cause__/__context__/original_exception)."""
    errors: list[BaseException] = []
    stack: list[BaseException | None] = [exc]
    visited: set[int] = set()
    while stack:
        current = stack.pop()
        if current is None:
            continue
        marker = id(current)
        if marker in visited:
            continue
        visited.add(marker)
        errors.append(current)
        stack.append(getattr(current, "__cause__", None))
        stack.append(getattr(current, "__context__", None))
        stack.append(getattr(current, "original_exception", None))
    return errors


def _is_invalid_conversation_manager_state(exc: BaseException) -> bool:
    """Return True when Strands session restore rejects conversation manager state."""
    for err in _iter_exception_chain(exc):
        if isinstance(err, ValueError) and "invalid conversation manager state" in str(err).lower():
            return True
    return False


def _is_retryable(exc: BaseException) -> bool:
    """Return True for transient server errors (5xx, 429) that warrant a retry."""
    # Gemini SDK — google.api_core.exceptions.ServerError
    if type(exc).__name__ == "ServerError" and "google" in getattr(type(exc), "__module__", ""):
        return True
    # OpenAI / Anthropic (stainless-based SDKs expose status_code)
    status = getattr(exc, "status_code", None)
    if status is not None and status in (429, 500, 502, 503, 504):
        return True
    # Strands may wrap the original exception as __cause__
    cause = getattr(exc, "__cause__", None)
    if cause is not None:
        return _is_retryable(cause)
    return False


def _is_gemini_invalid_turn_order(exc: BaseException) -> bool:
    """Return True if *exc* is a Gemini 400 caused by invalid function-call turn order."""
    for err in (exc, getattr(exc, "__cause__", None), getattr(exc, "original_exception", None)):
        if err is None:
            continue
        module = getattr(type(err), "__module__", "") or ""
        if "google" not in module:
            continue
        status = getattr(err, "code", None) or getattr(err, "status_code", None)
        if status == 400 and "function call turn" in str(err).lower():
            return True
    return False


def _sanitize_message_order(messages: Messages) -> None:
    """Remove leading non-user messages so the conversation starts with a user turn."""
    while messages and messages[0].get("role") != "user":
        messages.pop(0)


def _log_retry(retry_state: object) -> None:
    """Log each retry attempt at WARNING level."""
    attempt = getattr(retry_state, "attempt_number", "?")
    outcome = getattr(retry_state, "outcome", None)
    exc = outcome.exception() if outcome else None
    wait = getattr(retry_state, "next_action", None)
    sleep_seconds = getattr(wait, "sleep", None) if wait else None
    logger.warning(
        "LLM call retry attempt=%s, wait=%.1fs, error=%s: %s",
        attempt,
        sleep_seconds if sleep_seconds is not None else 0.0,
        type(exc).__name__ if exc else "unknown",
        exc,
    )


@dataclass
class _AgentCacheEntry:
    agent: Agent
    lock: Lock
    last_access: float


class AnalysisEngine(Protocol):
    def analyze(self, prompt: str, incident_id: str | None = None) -> str:
        raise NotImplementedError


class StrandsAnalysisEngine:
    def __init__(
        self,
        settings: Settings,
        k8s_client: KubernetesClient,
        prometheus_client: PrometheusClient | None = None,
        tempo_client: TempoClient | None = None,
        masker: RegexMasker | None = None,
        model_config: ModelConfig | None = None,
    ) -> None:
        # Apply Gemini patch only when using Gemini provider
        if model_config is None or model_config.provider == LLMProvider.GEMINI:
            apply_gemini_thought_signature_patch()

        if not settings.session_store_dsn:
            raise ValueError(
                "SESSION_DB_HOST, SESSION_DB_USER, SESSION_DB_NAME are required "
                "for Postgres session storage"
            )
        self._settings = settings
        self._model_config = model_config
        self._masker = masker or RegexMasker()
        self._tools = _build_tools(k8s_client, prometheus_client, tempo_client, self._masker)
        self._cache_lock = Lock()
        self._agent_cache: OrderedDict[str, _AgentCacheEntry] = OrderedDict()
        self._cache_size = max(settings.agent_cache_size, 1)
        self._cache_ttl_seconds = settings.agent_cache_ttl_seconds
        self._session_repo = PostgresSessionRepository(settings.session_store_dsn)

        # LLM retry settings
        self._retry_max_attempts = max(1, settings.llm_retry_max_attempts)
        self._retry_min_wait = settings.llm_retry_min_wait
        self._retry_max_wait = settings.llm_retry_max_wait
        self._retry_total_timeout = settings.llm_retry_total_timeout

        # Log provider info
        if model_config:
            logger.info(
                "Analysis engine initialized with provider=%s, model=%s",
                model_config.provider.value,
                model_config.model_id,
            )

    def analyze(self, prompt: str, incident_id: str | None = None) -> str:
        session_id = self._resolve_session_id(incident_id)
        try:
            return self._analyze_once(prompt, session_id)
        except Exception as exc:
            if not _is_invalid_conversation_manager_state(exc):
                raise

            expected_manager = SafeSlidingWindowConversationManager.__name__
            found_manager = self._read_stored_manager_name(session_id)
            logger.warning(
                "chat_session_recovery_started "
                "session_id=%s error_type=%s expected_manager=%s found_manager=%s",
                session_id,
                type(exc).__name__,
                expected_manager,
                found_manager or "unknown",
            )
            self._reset_session_state(session_id)

            try:
                result = self._analyze_once(prompt, session_id)
            except Exception:
                logger.exception(
                    "chat_session_recovery_failed "
                    "session_id=%s expected_manager=%s found_manager=%s",
                    session_id,
                    expected_manager,
                    found_manager or "unknown",
                )
                raise

            logger.info(
                "chat_session_recovery_succeeded "
                "session_id=%s expected_manager=%s found_manager=%s",
                session_id,
                expected_manager,
                found_manager or "unknown",
            )
            return result

    def _analyze_once(self, prompt: str, session_id: str) -> str:
        entry = self._get_cache_entry(session_id)
        with self._session_repo.session_lock(session_id):
            with entry.lock:
                return self._invoke_with_retry(entry.agent, prompt)

    def _read_stored_manager_name(self, session_id: str) -> str | None:
        try:
            return self._session_repo.read_conversation_manager_name(session_id)
        except Exception:  # noqa: BLE001
            logger.warning(
                "Failed to read stored conversation manager state for session=%s",
                session_id,
            )
            return None

    def _reset_session_state(self, session_id: str) -> None:
        with self._cache_lock:
            self._agent_cache.pop(session_id, None)

        with self._session_repo.session_lock(session_id):
            self._session_repo.delete_session(session_id)

    def _invoke_with_retry(self, agent: Agent, prompt: str) -> str:
        """Invoke the LLM agent with tenacity retry on transient errors."""

        @retry(
            retry=retry_if_exception(_is_retryable),
            wait=wait_exponential(multiplier=1, min=self._retry_min_wait, max=self._retry_max_wait)
            + wait_random(0, 2),
            stop=stop_after_attempt(self._retry_max_attempts)
            | stop_after_delay(self._retry_total_timeout),
            before_sleep=_log_retry,
        )
        def _call() -> str:
            return str(agent(prompt))

        try:
            return _call()
        except Exception as exc:
            if _is_gemini_invalid_turn_order(exc):
                logger.warning("Gemini turn-order error detected, sanitizing messages and retrying")
                _sanitize_message_order(agent.messages)
                return str(agent(prompt))
            raise

    def _resolve_session_id(self, incident_id: str | None) -> str:
        cache_key = incident_id.strip() if incident_id else ""
        return cache_key or "default"

    def _get_cache_entry(self, session_id: str) -> _AgentCacheEntry:
        cache_key = session_id

        with self._cache_lock:
            now = time.time()
            self._evict_expired_entries(now)
            entry = self._agent_cache.get(cache_key)
            if entry is not None:
                entry.last_access = now
                self._agent_cache.move_to_end(cache_key)
                return entry

            entry = _AgentCacheEntry(
                agent=self._build_agent(cache_key),
                lock=Lock(),
                last_access=now,
            )
            self._agent_cache[cache_key] = entry
            if len(self._agent_cache) > self._cache_size:
                self._agent_cache.popitem(last=False)
            return entry

    def _evict_expired_entries(self, now: float) -> None:
        if self._cache_ttl_seconds <= 0:
            return

        expired_keys: list[str] = []
        for key, entry in self._agent_cache.items():
            if now - entry.last_access > self._cache_ttl_seconds:
                expired_keys.append(key)
            else:
                break
        for key in expired_keys:
            self._agent_cache.pop(key, None)

    def _build_agent(self, session_id: str) -> Agent:
        model = self._create_model()
        session_manager = RepositorySessionManager(
            session_id=session_id,
            session_repository=self._session_repo,
        )
        conversation_manager = SafeSlidingWindowConversationManager(window_size=40)
        return Agent(
            model=model,
            tools=self._tools,
            session_manager=session_manager,
            conversation_manager=conversation_manager,
        )

    def _create_model(self) -> object:
        """Create the LLM model based on configuration.

        Returns model using the multi-provider factory if model_config is set,
        otherwise falls back to Gemini for backward compatibility.
        """
        if self._model_config is not None:
            return create_model(self._model_config)

        # Fallback to legacy Gemini configuration for backward compatibility
        from strands.models.gemini import GeminiModel

        return GeminiModel(
            client_args={"api_key": self._settings.gemini_api_key},
            model_id=self._settings.gemini_model_id,
        )


def _build_tools(
    k8s_client: KubernetesClient,
    prometheus_client: PrometheusClient | None,
    tempo_client: TempoClient | None,
    masker: RegexMasker,
) -> list[object]:
    def _mask(data: Any) -> Any:
        return masker.mask_object(data)

    @_logged_tool(arg_formatter=_pod_lookup_summary)
    def get_pod_status(namespace: str, pod_name: str) -> dict[str, object]:
        """Fetch pod status details."""
        status = k8s_client.get_pod_status(namespace, pod_name)
        if status is None:
            return _mask({"warning": "pod status not found"})
        return _mask(status.to_dict())

    @_logged_tool(arg_formatter=_pod_lookup_summary)
    def get_pod_spec(namespace: str, pod_name: str) -> dict[str, object]:
        """Fetch a summarized pod spec (containers/resources/probes)."""
        spec = k8s_client.get_pod_spec_summary(namespace, pod_name)
        if spec is None:
            return _mask({"warning": "pod spec not found"})
        return _mask(spec)

    @_logged_tool(arg_formatter=_pod_lookup_summary, result_formatter=_default_result_summary)
    def list_pod_events(namespace: str, pod_name: str) -> list[dict[str, object]]:
        """List recent events for a pod."""
        return _mask([event.to_dict() for event in k8s_client.list_pod_events(namespace, pod_name)])

    @_logged_tool(result_formatter=_default_result_summary)
    def list_namespace_events(namespace: str) -> list[dict[str, object]]:
        """List recent events for a namespace."""
        return _mask([event.to_dict() for event in k8s_client.list_namespace_events(namespace)])

    @_logged_tool(result_formatter=_default_result_summary)
    def list_cluster_events() -> list[dict[str, object]]:
        """List recent events across all namespaces."""
        return _mask([event.to_dict() for event in k8s_client.list_cluster_events()])

    @_logged_tool(arg_formatter=_pod_lookup_summary, result_formatter=_log_result_summary)
    def get_previous_pod_logs(
        namespace: str,
        pod_name: str,
    ) -> list[dict[str, object]]:
        """Return previous container logs (tail)."""
        return _mask(
            [snippet.to_dict() for snippet in k8s_client.get_previous_logs(namespace, pod_name)]
        )

    @_logged_tool(arg_formatter=_pod_lookup_summary, result_formatter=_log_result_summary)
    def get_pod_logs(
        namespace: str,
        pod_name: str,
        container: str | None = None,
        tail_lines: int | None = None,
        since_seconds: int | None = None,
    ) -> list[dict[str, object]]:
        """Return current container logs (tail)."""
        return _mask(
            [
                snippet.to_dict()
                for snippet in k8s_client.get_pod_logs(
                    namespace,
                    pod_name,
                    container=container,
                    tail_lines=tail_lines,
                    since_seconds=since_seconds,
                )
            ]
        )

    @_logged_tool(arg_formatter=_pod_lookup_summary)
    def get_workload_status(namespace: str, pod_name: str) -> dict[str, object]:
        """Fetch top-level workload status (Deployment/StatefulSet/Job)."""
        summary = k8s_client.get_workload_summary(namespace, pod_name)
        if summary is None:
            return _mask({"warning": "workload not found"})
        return _mask(summary)

    @_logged_tool(arg_formatter=_namespace_name_summary)
    def get_daemonset_manifest(namespace: str, name: str) -> dict[str, object]:
        """Fetch DaemonSet manifest (metadata/spec) for updateStrategy/selector."""
        manifest = k8s_client.get_daemon_set_manifest(namespace, name)
        if manifest is None:
            return _mask({"warning": "daemonset not found"})
        return _mask(manifest)

    @_logged_tool()
    def get_node_status(node_name: str) -> dict[str, object]:
        """Fetch node status details."""
        status = k8s_client.get_node_status(node_name)
        if status is None:
            return _mask({"warning": "node not found"})
        return _mask(status)

    @_logged_tool(arg_formatter=_pod_lookup_summary)
    def get_pod_metrics(namespace: str, pod_name: str) -> dict[str, object]:
        """Fetch pod CPU/memory usage from metrics-server."""
        metrics = k8s_client.get_pod_metrics(namespace, pod_name)
        if metrics is None:
            return _mask({"warning": "pod metrics not available"})
        return _mask(metrics)

    @_logged_tool()
    def get_node_metrics(node_name: str) -> dict[str, object]:
        """Fetch node CPU/memory usage from metrics-server."""
        metrics = k8s_client.get_node_metrics(node_name)
        if metrics is None:
            return _mask({"warning": "node metrics not available"})
        return _mask(metrics)

    @_logged_tool()
    def discover_prometheus() -> dict[str, object]:
        """Return the configured Prometheus endpoint (PROMETHEUS_URL)."""
        if prometheus_client is None:
            return _mask({"warning": "prometheus client not configured"})
        return _mask(prometheus_client.describe_endpoint())

    @_logged_tool()
    def list_prometheus_metrics(match: str | None = None) -> dict[str, object]:
        """List available metric names from Prometheus.

        Use this to discover what metrics are available before querying.
        Args:
            match: Optional regex pattern to filter metrics.
                   Examples: 'kube_pod.*', 'istio.*', 'container_.*', 'http_.*'
        """
        if prometheus_client is None:
            return _mask({"warning": "prometheus client not configured"})
        return _mask(prometheus_client.list_metrics(match=match))

    @_logged_tool(
        arg_formatter=_prometheus_query_summary,
        result_formatter=_prometheus_result_summary,
    )
    def query_prometheus(query: str, time: str | None = None) -> dict[str, object]:
        """Run a Prometheus instant query."""
        if prometheus_client is None:
            return _mask({"warning": "prometheus client not configured"})
        return _mask(prometheus_client.query(query, time=time))

    @_logged_tool(
        arg_formatter=_prometheus_range_summary,
        result_formatter=_prometheus_range_result_summary,
    )
    def query_prometheus_range(
        query: str,
        start: str,
        end: str,
        step: str = "1m",
    ) -> dict[str, object]:
        """Run a Prometheus range query to get time-series (history) data.

        Use this to analyze metric trends over time, such as memory usage history
        before an OOMKilled event.

        Args:
            query: PromQL query (e.g., 'container_memory_usage_bytes{pod="my-pod"}').
            start: Start time (RFC3339 e.g., '2024-01-01T00:00:00Z' or Unix timestamp).
            end: End time (RFC3339 or Unix timestamp).
            step: Query resolution (e.g., '1m', '5m', '1h'). Default '1m'.

        Returns:
            Matrix result with timestamp-value pairs for each time series.
        """
        if prometheus_client is None:
            return _mask({"warning": "prometheus client not configured"})
        return _mask(prometheus_client.query_range(query, start=start, end=end, step=step))

    @_logged_tool()
    def discover_tempo() -> dict[str, object]:
        """Return the configured Tempo endpoint (TEMPO_URL)."""
        if tempo_client is None:
            return _mask({"warning": "tempo client not configured"})
        return _mask(tempo_client.describe_endpoint())

    @_logged_tool(
        arg_formatter=_tempo_search_summary,
        result_formatter=_tempo_result_summary,
    )
    def search_tempo_traces(
        start: str,
        end: str,
        service_name: str | None = None,
        namespace: str | None = None,
        query: str | None = None,
        limit: int = 5,
    ) -> dict[str, object]:
        """Search traces from Tempo.

        Use either:
        - query: custom TraceQL string
        - service_name/namespace: auto-build TraceQL query
        """
        if tempo_client is None:
            return _mask({"warning": "tempo client not configured"})

        traceql = query or build_traceql_query(service_name, namespace)
        return _mask(
            tempo_client.search_traces(
                query=traceql,
                start=start,
                end=end,
                limit=limit,
            )
        )

    @_logged_tool()
    def get_tempo_trace(trace_id: str) -> dict[str, object]:
        """Fetch a full Tempo trace payload by trace ID."""
        if tempo_client is None:
            return _mask({"warning": "tempo client not configured"})
        return _mask(tempo_client.get_trace(trace_id))

    @_logged_tool(arg_formatter=_list_pods_summary, result_formatter=_default_result_summary)
    def list_pods_in_namespace(
        namespace: str, label_selector: str | None = None
    ) -> list[dict[str, object]]:
        """List pods in a namespace with optional label filtering.

        Use this when pod name is not in alert labels to discover affected pods.
        Args:
            namespace: The Kubernetes namespace to list pods from.
            label_selector: Optional label selector (e.g., 'app=nginx', 'version=v1').
        """
        pods = k8s_client.list_pods_in_namespace(namespace, label_selector=label_selector)
        return _mask([pod.to_dict() for pod in pods])

    @_logged_tool(arg_formatter=_manifest_summary)
    def get_manifest(
        namespace: str,
        api_version: str,
        resource: str,
        name: str,
    ) -> dict[str, object]:
        """Fetch a namespaced manifest from core or custom resources.

        Examples:
        - api_version='v1', resource='services'
        - api_version='networking.istio.io/v1', resource='virtualservices'
        """
        manifest = k8s_client.get_manifest(
            namespace,
            api_version=api_version,
            resource=resource,
            name=name,
        )
        if manifest is None:
            return _mask({"warning": "manifest not found or unsupported"})
        return _mask(manifest)

    @_logged_tool(arg_formatter=_manifest_summary, result_formatter=_default_result_summary)
    def list_manifests(
        namespace: str,
        api_version: str,
        resource: str,
        label_selector: str | None = None,
        limit: int = 20,
    ) -> list[dict[str, object]]:
        """List namespaced manifests from core or custom resources."""
        return _mask(
            k8s_client.list_manifests(
                namespace,
                api_version=api_version,
                resource=resource,
                label_selector=label_selector,
                limit=limit,
            )
        )

    tools: list[object] = [
        get_pod_status,
        get_pod_spec,
        list_pod_events,
        list_namespace_events,
        list_cluster_events,
        list_pods_in_namespace,
        get_previous_pod_logs,
        get_pod_logs,
        get_workload_status,
        get_daemonset_manifest,
        get_node_status,
        get_pod_metrics,
        get_node_metrics,
        get_manifest,
        list_manifests,
    ]
    if prometheus_client is not None:
        tools.extend(
            [
                discover_prometheus,
                list_prometheus_metrics,
                query_prometheus,
                query_prometheus_range,
            ]
        )
    if tempo_client is not None:
        tools.extend(
            [
                discover_tempo,
                search_tempo_traces,
                get_tempo_trace,
            ]
        )
    return tools
