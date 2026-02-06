from __future__ import annotations

import json
import logging
from typing import Any, cast
from uuid import uuid4

from app.clients.k8s import KubernetesClient, extract_pod_target
from app.clients.strands_agent import AnalysisEngine
from app.clients.summary_store import SummaryStore
from app.core.masking import RegexMasker
from app.models.k8s import K8sContext
from app.schemas.analysis import AlertAnalysisRequest, IncidentSummaryRequest


class AnalysisService:
    def __init__(
        self,
        k8s_client: KubernetesClient,
        analysis_engine: AnalysisEngine | None,
        masker: RegexMasker | None = None,
        prometheus_enabled: bool = False,
        summary_store: SummaryStore | None = None,
        summary_history_size: int = 3,
        prompt_token_budget: int = 32000,
        prompt_max_log_lines: int = 25,
        prompt_max_events: int = 25,
    ) -> None:
        self._logger = logging.getLogger(__name__)
        self._k8s_client = k8s_client
        self._analysis_engine = analysis_engine
        self._masker = masker or RegexMasker()
        self._prometheus_enabled = prometheus_enabled
        self._summary_store = summary_store
        self._summary_history_size = max(1, summary_history_size)
        self._prompt_token_budget = max(0, prompt_token_budget)
        self._prompt_max_log_lines = max(0, prompt_max_log_lines)
        self._prompt_max_events = max(0, prompt_max_events)

    def analyze(
        self, request: AlertAnalysisRequest
    ) -> tuple[str, str, str, dict[str, object], list[dict[str, object]]]:
        namespace, pod_name, workload = extract_pod_target(request.alert.labels)
        k8s_context = self._k8s_client.collect_context(namespace, pod_name, workload)
        context = k8s_context.to_dict()
        artifacts = _build_alert_artifacts(k8s_context)
        masked_context = cast(dict[str, object], self._masker.mask_object(context))
        masked_artifacts = cast(list[dict[str, object]], self._masker.mask_object(artifacts))

        if self._analysis_engine is None:
            analysis = self._masker.mask_text(
                _fallback_summary(request, k8s_context, "analysis engine not configured")
            )
            summary, detail = _split_alert_analysis(analysis)
            return analysis, summary, detail, masked_context, masked_artifacts

        summary_key = _resolve_alert_session_id(request)
        recent_summaries = self._load_recent_summaries(summary_key)
        prompt = _build_prompt(
            request,
            k8s_context,
            self._prometheus_enabled,
            recent_summaries,
            self._prompt_token_budget,
            self._prompt_max_log_lines,
            self._prompt_max_events,
            self._masker,
        )
        try:
            session_id = _build_runtime_session_id(summary_key)
            analysis = self._analysis_engine.analyze(prompt, session_id)
            if not isinstance(analysis, str):
                analysis = ""
            analysis = self._masker.mask_text(analysis)
            if not analysis.strip():
                self._logger.warning("Strands analysis returned empty response")
                analysis = self._masker.mask_text(
                    _fallback_summary(
                        request,
                        k8s_context,
                        "analysis engine returned empty response",
                    )
                )
                summary, detail = _split_alert_analysis(analysis)
                return analysis, summary, detail, masked_context, masked_artifacts
            summary, detail = _split_alert_analysis(analysis)
            self._store_summary(summary_key, summary)
            return analysis, summary, detail, masked_context, masked_artifacts
        except Exception as exc:  # noqa: BLE001
            self._logger.exception("Strands analysis failed")
            analysis = self._masker.mask_text(
                _fallback_summary(request, k8s_context, f"analysis failed: {exc}")
            )
            summary, detail = _split_alert_analysis(analysis)
            return analysis, summary, detail, masked_context, masked_artifacts

    def summarize_incident(self, request: IncidentSummaryRequest) -> tuple[str, str, str]:
        """Synthesize final RCA summary for a resolved incident.

        Returns:
            tuple[str, str, str]: (title, summary, detail)
        """
        if self._analysis_engine is None:
            return self._mask_incident_result(
                _fallback_incident_summary(request, "analysis engine not configured")
            )

        prompt = _build_incident_summary_prompt(request, self._masker)
        try:
            session_id = _resolve_summary_session_id(request)
            result = self._analysis_engine.analyze(prompt, session_id)
            if not isinstance(result, str):
                result = ""
            masked_result = self._masker.mask_text(result)
            return self._mask_incident_result(_parse_incident_summary(masked_result, request.title))
        except Exception as exc:  # noqa: BLE001
            self._logger.exception("Incident summary analysis failed")
            return self._mask_incident_result(
                _fallback_incident_summary(request, f"analysis failed: {exc}")
            )

    def _mask_incident_result(self, result: tuple[str, str, str]) -> tuple[str, str, str]:
        title, summary, detail = result
        return (
            self._masker.mask_text(title),
            self._masker.mask_text(summary),
            self._masker.mask_text(detail),
        )

    def _load_recent_summaries(self, session_id: str) -> list[str]:
        if self._summary_store is None or self._summary_history_size <= 0:
            return []
        try:
            summaries = self._summary_store.list_summaries(session_id, self._summary_history_size)
            return [self._masker.mask_text(summary) for summary in summaries]
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to load session summaries: %s", exc)
            return []

    def _store_summary(self, session_id: str, summary: str) -> None:
        if self._summary_store is None or self._summary_history_size <= 0:
            return
        masked_summary = self._masker.mask_text(summary)
        compact = _compact_summary(masked_summary, limit=300) or masked_summary.strip()
        if not compact:
            return
        try:
            self._summary_store.append_summary(
                session_id,
                compact,
                max_items=self._summary_history_size,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to store session summary: %s", exc)


def _build_prompt(
    request: AlertAnalysisRequest,
    k8s_context: K8sContext,
    prometheus_enabled: bool,
    recent_summaries: list[str],
    prompt_token_budget: int,
    prompt_max_log_lines: int,
    prompt_max_events: int,
    masker: RegexMasker,
) -> str:
    alert_payload = cast(
        dict[str, Any],
        masker.mask_object(request.alert.model_dump(by_alias=True, mode="json")),
    )
    payload = {
        "alert": alert_payload,
        "thread_ts": masker.mask_text(request.thread_ts),
    }
    if request.incident_id:
        payload["incident_id"] = masker.mask_text(request.incident_id)

    tool_lines = [
        "- get_pod_status, get_pod_spec",
        "- list_pod_events, list_namespace_events, list_cluster_events",
        "- list_pods_in_namespace (use when pod name is missing from alert labels)",
        "- get_previous_pod_logs, get_pod_logs",
        "- get_workload_status, get_daemonset_manifest, get_node_status",
        "- get_pod_metrics, get_node_metrics",
    ]
    if prometheus_enabled:
        tool_lines.append("- discover_prometheus, list_prometheus_metrics")
        tool_lines.append("- query_prometheus, query_prometheus_range")
    tool_block = "\n".join(tool_lines)

    summary_block = _format_session_summaries(
        [masker.mask_text(summary) for summary in recent_summaries]
    )
    prompt = (
        "You are kube-rca-agent. Analyze the alert using the provided Kubernetes context.\n"
        "Return your response in Korean with the following structure:\n"
        "1) 요약 (Summary): 1-2 sentences, <= 300 chars.\n"
        "   Include root cause + impact + next action.\n"
        "2) 상세 분석 (Detail): Use sections for 근본 원인, 확인 근거, 조치 사항, 누락된 데이터.\n"
        "Formatting rules:\n"
        "- Use markdown headers: '### 1) 요약 (Summary)' and '### 2) 상세 분석 (Detail)'.\n"
        "- Use '####' for subsections: 근본 원인, 확인 근거, 조치 사항, 누락된 데이터.\n"
        "- Leave one blank line between sections/subsections.\n"
        "- Use '-' for unordered lists (do not use '*').\n"
        "- Limit each subsection to 3-5 bullets; one sentence per bullet (<= 120 chars).\n"
        "- Use inline code only for literal keys/values/commands; "
        "avoid excessive code formatting.\n"
        "If data is missing, state what is missing.\n"
        "You may call tools if needed:\n"
        f"{tool_block}\n\n"
    )

    if prometheus_enabled:
        prompt += (
            "For Prometheus queries:\n"
            "1. Use list_prometheus_metrics(match='pattern') to discover available metrics.\n"
            "2. Use query_prometheus(query) for current/instant values.\n"
            "3. Use query_prometheus_range(query, start, end, step) for time-series history.\n"
            "   - ALWAYS use range queries to understand metric trends before the alert.\n"
            "   - Use cases: memory/CPU spikes, error rate increase, latency degradation,\n"
            "     request volume changes, network issues, resource exhaustion, etc.\n"
            "   - Use alert's startsAt to calculate start (e.g., 1h before) and end time.\n"
            "   - Example: query_prometheus_range(\n"
            "       query='rate(http_requests_total{pod=\"my-pod\"}[5m])',\n"
            "       start='<startsAt - 1h>', end='<startsAt>', step='1m')\n"
            "Example patterns: 'container_memory.*', 'container_cpu.*', 'http_.*',\n"
            "'istio_request.*', 'kube_pod.*', 'node_.*'\n\n"
        )

    if summary_block:
        prompt += summary_block

    context_dict = _prepare_k8s_context(
        k8s_context,
        max_events=prompt_max_events,
        max_log_lines=prompt_max_log_lines,
    )
    context_dict = cast(dict[str, object], masker.mask_object(context_dict))
    alert_block = f"Alert payload:\n{_to_pretty_json(payload)}\n\n"
    context_block = f"Kubernetes context:\n{_to_pretty_json(context_dict)}\n"
    full_prompt = prompt + alert_block + context_block
    return _apply_prompt_budget(
        full_prompt,
        prompt_prefix=prompt,
        alert_block=alert_block,
        context_dict=context_dict,
        prompt_token_budget=prompt_token_budget,
    )


def _build_runtime_session_id(summary_key: str) -> str:
    suffix = uuid4().hex[:8]
    if summary_key:
        return f"{summary_key}:run:{suffix}"
    return f"run:{suffix}"


def _format_session_summaries(summaries: list[str]) -> str:
    if not summaries:
        return ""
    lines = ["Recent session summaries (latest 3):"]
    for idx, summary in enumerate(summaries, start=1):
        compact = _compact_summary(summary, limit=300)
        if not compact:
            continue
        lines.append(f"{idx}) {compact}")
    if len(lines) == 1:
        return ""
    return "\n".join(lines) + "\n\n"


def _prepare_k8s_context(
    k8s_context: K8sContext, *, max_events: int, max_log_lines: int
) -> dict[str, object]:
    context = k8s_context.to_dict()
    events = context.get("events") or []
    if max_events <= 0:
        context["events"] = []
    else:
        context["events"] = events[:max_events]

    logs = context.get("previous_logs") or []
    if max_log_lines <= 0:
        context["previous_logs"] = []
        return context

    trimmed_logs: list[dict[str, object]] = []
    for snippet in logs:
        lines = snippet.get("logs") or []
        lines = _dedupe_consecutive_lines(lines)
        if max_log_lines > 0:
            lines = lines[-max_log_lines:]
        trimmed = dict(snippet)
        trimmed["logs"] = lines
        trimmed_logs.append(trimmed)
    context["previous_logs"] = trimmed_logs
    return context


def _apply_prompt_budget(
    prompt: str,
    *,
    prompt_prefix: str,
    alert_block: str,
    context_dict: dict[str, object],
    prompt_token_budget: int,
) -> str:
    budget_chars = _prompt_budget_to_chars(prompt_token_budget)
    if budget_chars <= 0 or len(prompt) <= budget_chars:
        return prompt

    def build_context_block(
        context: dict[str, object] | None,
        note: str | None = None,
    ) -> str:
        if context is None:
            reason = note or "omitted due to prompt budget"
            return f"Kubernetes context: {reason}.\n"
        header = "Kubernetes context"
        if note:
            header = f"{header} ({note})"
        return f"{header}:\n{_to_pretty_json(context)}\n"

    # Step 1: remove previous logs
    if context_dict.get("previous_logs"):
        trimmed = dict(context_dict)
        trimmed["previous_logs"] = []
        candidate = prompt_prefix + alert_block + build_context_block(trimmed, "logs omitted")
        if len(candidate) <= budget_chars:
            return candidate

    # Step 2: remove events
    if context_dict.get("events"):
        trimmed = dict(context_dict)
        trimmed["events"] = []
        trimmed["previous_logs"] = []
        candidate = (
            prompt_prefix + alert_block + build_context_block(trimmed, "events/logs omitted")
        )
        if len(candidate) <= budget_chars:
            return candidate

    # Step 3: compact context
    compact = _compact_context_dict(context_dict)
    candidate = prompt_prefix + alert_block + build_context_block(compact, "compact")
    if len(candidate) <= budget_chars:
        return candidate

    # Step 4: omit context entirely
    return prompt_prefix + alert_block + build_context_block(None, "omitted due to prompt budget")


def _compact_context_dict(context: dict[str, object]) -> dict[str, object]:
    pod_status = context.get("pod_status") or {}
    compact_status = None
    if pod_status:
        compact_status = {
            "phase": pod_status.get("phase"),
            "reason": pod_status.get("reason"),
            "message": pod_status.get("message"),
            "node_name": pod_status.get("node_name"),
        }
    return {
        "namespace": context.get("namespace"),
        "pod_name": context.get("pod_name"),
        "workload": context.get("workload"),
        "pod_status": compact_status,
        "warnings": context.get("warnings") or [],
    }


def _dedupe_consecutive_lines(lines: list[str]) -> list[str]:
    deduped: list[str] = []
    last = None
    for line in lines:
        if line == last:
            continue
        deduped.append(line)
        last = line
    return deduped


def _prompt_budget_to_chars(prompt_token_budget: int) -> int:
    if prompt_token_budget <= 0:
        return 0
    return prompt_token_budget * 4


def _resolve_alert_session_id(request: AlertAnalysisRequest) -> str:
    incident_id = _normalize_session_token(request.incident_id)
    fingerprint = _normalize_session_token(request.alert.fingerprint)
    if incident_id and fingerprint:
        return f"{incident_id}:{fingerprint}"
    if fingerprint:
        return f"alert:{fingerprint}"
    fallback = _build_alert_fallback_key(request)
    if incident_id and fallback:
        return f"{incident_id}:{fallback}"
    if fallback:
        return f"alert:{fallback}"
    return incident_id or "default"


def _resolve_summary_session_id(request: IncidentSummaryRequest) -> str:
    incident_id = _normalize_session_token(request.incident_id)
    if incident_id:
        return f"{incident_id}:summary"
    return "summary"


def _normalize_session_token(value: str | None) -> str:
    if not value:
        return ""
    return value.strip()


def _build_alert_fallback_key(request: AlertAnalysisRequest) -> str:
    labels = request.alert.labels
    parts = [
        _normalize_session_token(labels.get("alertname")),
        _normalize_session_token(labels.get("namespace")),
        _normalize_session_token(labels.get("pod")),
        _normalize_session_token(labels.get("workload")),
        _normalize_session_token(labels.get("destination_workload")),
        _normalize_session_token(labels.get("destination_service_name")),
    ]
    compact = "-".join(part for part in parts if part)
    return compact


def _fallback_summary(
    request: AlertAnalysisRequest,
    k8s_context: K8sContext,
    reason: str,
) -> str:
    alert = request.alert
    lines = [
        f"analysis engine unavailable: {reason}",
        f"alert_status={alert.status}",
    ]
    if k8s_context.namespace or k8s_context.pod_name:
        lines.append(f"target: namespace={k8s_context.namespace}, pod={k8s_context.pod_name}")
    if k8s_context.warnings:
        lines.append("warnings: " + ", ".join(k8s_context.warnings))
    if k8s_context.pod_status:
        lines.append(f"pod_phase: {k8s_context.pod_status.phase}")
    return "\n".join(lines)


def _to_pretty_json(payload: dict[str, Any]) -> str:
    return json.dumps(payload, ensure_ascii=True, indent=2, sort_keys=True)


def _fallback_incident_summary(
    request: IncidentSummaryRequest, reason: str
) -> tuple[str, str, str]:
    """Generate fallback summary when analysis engine is unavailable."""
    alert_names = [a.alert_name for a in request.alerts]
    title = request.title  # Keep original title
    summary = f"인시던트 분석 불가: {reason}"
    detail_lines = [
        f"인시던트 ID: {request.incident_id}",
        f"제목: {request.title}",
        f"심각도: {request.severity}",
        f"발생 시각: {request.fired_at}",
        f"해결 시각: {request.resolved_at}",
        f"관련 알림 ({len(request.alerts)}개): {', '.join(alert_names)}",
        "",
        f"분석 엔진 오류: {reason}",
    ]
    return title, summary, "\n".join(detail_lines)


def _parse_incident_summary(result: str, original_title: str) -> tuple[str, str, str]:
    """Parse AI response into title, summary and detail.

    The AI is instructed to return structured content with 제목, 요약 and 상세 분석 sections.
    We extract each section and use the full result as detail.
    """
    lines = result.strip().split("\n")

    title = ""
    summary = ""

    for line in lines:
        stripped = line.strip()
        if not stripped:
            continue

        # Extract title
        if "제목" in stripped or "title" in stripped.lower():
            # Try to get the content after colon or on the next meaningful line
            if ":" in stripped:
                title = stripped.split(":", 1)[1].strip().strip("'\"")[:100]
            continue

        # Extract summary
        if "요약" in stripped or "summary" in stripped.lower():
            if ":" in stripped:
                summary = stripped.split(":", 1)[1].strip()[:200]
            continue

        # If we haven't found title yet and this looks like content, use it
        if not title and not stripped.startswith(("*", "#", "-")):
            title = stripped[:100]
            continue

        # If we have title but no summary and this looks like content
        skip_prefixes = ("*", "#", "-", "상세", "근본", "영향", "해결", "재발")
        if title and not summary and not stripped.startswith(skip_prefixes):
            summary = stripped[:200]
            break

    # Fallbacks
    if not title:
        title = original_title
    if not summary:
        summary = result[:200].replace("\n", " ").strip()
        if len(result) > 200:
            summary += "..."

    # Ensure summary and detail are not identical
    if summary.strip() == result.strip() or len(summary) > 300:
        summary = _compact_summary(result)

    return title, summary, result


def _build_incident_summary_prompt(request: IncidentSummaryRequest, masker: RegexMasker) -> str:
    alerts_info = []
    for alert in request.alerts:
        alert_data = {
            "fingerprint": alert.fingerprint,
            "alert_name": alert.alert_name,
            "severity": alert.severity,
            "status": alert.status,
            "analysis_summary": alert.analysis_summary or "N/A",
            "analysis_detail": alert.analysis_detail or "N/A",
            "artifacts": [artifact.model_dump() for artifact in alert.artifacts or []],
        }
        alerts_info.append(alert_data)

    incident_data = {
        "incident_id": request.incident_id,
        "title": request.title,
        "severity": request.severity,
        "fired_at": request.fired_at,
        "resolved_at": request.resolved_at,
        "alert_count": len(request.alerts),
        "alerts": alerts_info,
    }
    incident_data = cast(dict[str, Any], masker.mask_object(incident_data))

    return (
        "You are kube-rca-agent. An incident has been resolved and you need to "
        "provide a final RCA summary.\n"
        "Analyze all the alerts and their individual analyses to synthesize a "
        "comprehensive incident summary.\n\n"
        "Return your response in Korean with the following structure:\n"
        "1. **제목 (Title)**: A concise incident title (max 100 chars) that includes:\n"
        "   - The specific service/pod/namespace affected\n"
        "   - The root cause or error type\n"
        "   - Examples:\n"
        "     - '[payment-service] OOMKilled로 인한 Pod 재시작'\n"
        "     - '[nginx/prod] ImagePullBackOff - 잘못된 이미지 태그'\n"
        "     - '[redis-cluster] 메모리 부족으로 인한 연결 실패'\n"
        "2. **요약 (Summary)**: 1-2 sentences describing the root cause and resolution\n"
        "3. **상세 분석 (Detail)**:\n"
        "   - 근본 원인 (Root Cause)\n"
        "   - 영향 범위 (Impact)\n"
        "   - 해결 과정 (Resolution)\n"
        "   - 재발 방지 권고 (Prevention Recommendations)\n\n"
        f"Incident data:\n{_to_pretty_json(incident_data)}\n"
    )


def _split_alert_analysis(result: str) -> tuple[str, str]:
    all_keys = ["요약", "summary", "상세", "detail"]
    summary = _extract_section(result, ["요약", "summary"], all_keys)
    detail = _extract_section(result, ["상세 분석", "상세", "detail"], all_keys)

    if not detail:
        detail = result.strip()
    if not summary:
        summary = _compact_summary(detail)

    if summary.strip() == detail.strip():
        summary = _compact_summary(detail)

    return summary, detail


def _extract_section(text: str, keys: list[str], all_keys: list[str]) -> str | None:
    lines = text.splitlines()
    start = None
    for idx, line in enumerate(lines):
        if _is_section_header(line, keys):
            start = idx + 1
            break
    if start is None:
        return None

    end = len(lines)
    for idx in range(start, len(lines)):
        if _is_section_header(lines[idx], all_keys):
            end = idx
            break

    section = "\n".join(lines[start:end]).strip()
    return section or None


def _is_section_header(line: str, keys: list[str]) -> bool:
    stripped = line.strip()
    if not stripped:
        return False
    lowered = stripped.lower()
    has_key = any(key in lowered for key in keys)
    if not has_key:
        return False
    return stripped.startswith("#") or stripped.endswith(":") or stripped.endswith("：")


def _compact_summary(text: str, limit: int = 300) -> str:
    lines = [line.strip() for line in text.splitlines() if line.strip()]
    if not lines:
        return ""
    summary = lines[0]
    if len(lines) > 1 and len(summary) < limit:
        summary = f"{summary} {lines[1]}"
    if len(summary) > limit:
        summary = summary[:limit].rstrip() + "..."
    return summary


def _build_alert_artifacts(k8s_context: K8sContext) -> list[dict[str, object]]:
    artifacts: list[dict[str, object]] = []

    for event in k8s_context.events:
        event_data = event.to_dict()
        summary = " / ".join(
            part
            for part in [
                event_data.get("reason"),
                event_data.get("message"),
            ]
            if part
        )
        artifacts.append(
            {
                "type": "event",
                "summary": summary or "k8s event",
                "result": event_data,
            }
        )

    for snippet in k8s_context.previous_logs:
        logs = snippet.logs[:20]
        artifacts.append(
            {
                "type": "log",
                "summary": f"{snippet.container} logs ({len(logs)} lines)",
                "result": {
                    "container": snippet.container,
                    "previous": snippet.previous,
                    "error": snippet.error,
                    "logs": logs,
                },
            }
        )

    return artifacts
