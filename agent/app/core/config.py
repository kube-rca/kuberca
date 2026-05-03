from __future__ import annotations

import json
import os
import re
from dataclasses import dataclass

DEFAULT_GEMINI_MODEL_ID = "gemini-3-flash-preview"
DEFAULT_ANTHROPIC_MAX_TOKENS = 8192
DEFAULT_AI_PROVIDER = "gemini"


def _get_int_env(name: str, default: int) -> int:
    value = os.getenv(name)
    if not value:
        return default
    try:
        return int(value)
    except ValueError:
        return default


def _get_non_negative_int_env(name: str, default: int) -> int:
    return max(0, _get_int_env(name, default))


def _get_positive_int_env(name: str, default: int) -> int:
    value = os.getenv(name)
    if not value:
        return default
    try:
        parsed = int(value)
    except ValueError:
        return default
    if parsed <= 0:
        return default
    return parsed


def _get_float_env(name: str, default: float) -> float:
    value = os.getenv(name)
    if not value:
        return default
    try:
        return float(value)
    except ValueError:
        return default


def _get_string_list_json_env(name: str) -> list[str]:
    value = os.getenv(name, "").strip()
    if not value:
        return []

    try:
        parsed = json.loads(value)
    except json.JSONDecodeError as exc:
        raise ValueError(f"{name} must be a valid JSON array of strings") from exc

    if not isinstance(parsed, list):
        raise ValueError(f"{name} must be a valid JSON array of strings")

    patterns: list[str] = []
    for idx, item in enumerate(parsed):
        if not isinstance(item, str):
            raise ValueError(f"{name}[{idx}] must be a string")
        pattern = item.strip()
        if pattern:
            patterns.append(pattern)
    return patterns


def _validate_regex_list(patterns: list[str], name: str) -> None:
    for idx, pattern in enumerate(patterns):
        try:
            re.compile(pattern)
        except re.error as exc:
            raise ValueError(f"{name}[{idx}] must be a valid regex pattern") from exc


@dataclass(frozen=True)
class Settings:
    port: int
    log_level: str
    # AI Provider settings
    ai_provider: str  # gemini, openai, anthropic
    gemini_api_key: str
    gemini_model_id: str
    openai_model_id: str
    anthropic_model_id: str
    anthropic_max_tokens: int
    openai_api_key: str
    anthropic_api_key: str
    # Session DB settings
    session_db_host: str
    session_db_port: int
    session_db_name: str
    session_db_user: str
    session_db_password: str
    agent_cache_size: int
    agent_cache_ttl_seconds: int
    k8s_api_timeout_seconds: int
    k8s_event_limit: int
    k8s_log_tail_lines: int
    prometheus_url: str
    prometheus_http_timeout_seconds: int
    tempo_url: str
    tempo_http_timeout_seconds: int
    tempo_tenant_id: str
    tempo_trace_limit: int
    tempo_lookback_minutes: int
    tempo_forward_minutes: int
    loki_url: str
    loki_http_timeout_seconds: int
    loki_tenant_id: str
    prompt_token_budget: int
    prompt_max_log_lines: int
    prompt_max_events: int
    prompt_summary_max_items: int
    masking_regex_list: list[str]
    # Builtin redaction
    builtin_redaction_enabled: bool
    builtin_redaction_hash_mode: bool
    # LLM Retry (exponential backoff up to 3 minutes)
    llm_retry_max_attempts: int = 10
    llm_retry_min_wait: float = 1.0
    llm_retry_max_wait: float = 30.0
    llm_retry_total_timeout: float = 180.0
    # Concurrency
    max_concurrent_analyses: int = 5
    # Conversation manager sliding window
    agent_session_window_size: int = 40
    # Istio mesh CRD tools (Helm-controlled). Default off; enable in clusters
    # where Istio is installed so the agent can list VirtualService /
    # DestinationRule / ServiceEntry resources.
    agent_istio_enabled: bool = False

    @property
    def session_store_dsn(self) -> str:
        if not all([self.session_db_host, self.session_db_user, self.session_db_name]):
            return ""
        return (
            f"postgresql://{self.session_db_user}:{self.session_db_password}"
            f"@{self.session_db_host}:{self.session_db_port}/{self.session_db_name}"
        )


def load_settings() -> Settings:
    # Resolve AI provider and model
    ai_provider = os.getenv("AI_PROVIDER", DEFAULT_AI_PROVIDER).lower()
    masking_regex_list = _get_string_list_json_env("MASKING_REGEX_LIST_JSON")
    _validate_regex_list(masking_regex_list, "MASKING_REGEX_LIST_JSON")

    return Settings(
        port=_get_int_env("PORT", 8000),
        log_level=os.getenv("LOG_LEVEL", "info"),
        # AI Provider settings
        ai_provider=ai_provider,
        gemini_api_key=os.getenv("GEMINI_API_KEY", ""),
        gemini_model_id=os.getenv("GEMINI_MODEL_ID", DEFAULT_GEMINI_MODEL_ID),
        openai_model_id=os.getenv("OPENAI_MODEL_ID", "").strip(),
        anthropic_model_id=os.getenv("ANTHROPIC_MODEL_ID", "").strip(),
        anthropic_max_tokens=_get_positive_int_env(
            "ANTHROPIC_MAX_TOKENS", DEFAULT_ANTHROPIC_MAX_TOKENS
        ),
        openai_api_key=os.getenv("OPENAI_API_KEY", ""),
        anthropic_api_key=os.getenv("ANTHROPIC_API_KEY", ""),
        # Session DB settings
        session_db_host=os.getenv("SESSION_DB_HOST", ""),
        session_db_port=_get_int_env("SESSION_DB_PORT", 5432),
        session_db_name=os.getenv("SESSION_DB_NAME", ""),
        session_db_user=os.getenv("SESSION_DB_USER", ""),
        session_db_password=os.getenv("SESSION_DB_PASSWORD", ""),
        agent_cache_size=_get_non_negative_int_env("AGENT_CACHE_SIZE", 128),
        agent_cache_ttl_seconds=_get_non_negative_int_env("AGENT_CACHE_TTL_SECONDS", 0),
        k8s_api_timeout_seconds=_get_int_env("K8S_API_TIMEOUT_SECONDS", 5),
        k8s_event_limit=_get_int_env("K8S_EVENT_LIMIT", 25),
        k8s_log_tail_lines=_get_int_env("K8S_LOG_TAIL_LINES", 25),
        prometheus_url=os.getenv("PROMETHEUS_URL", "").strip(),
        prometheus_http_timeout_seconds=_get_int_env("PROMETHEUS_HTTP_TIMEOUT_SECONDS", 5),
        tempo_url=os.getenv("TEMPO_URL", "").strip(),
        tempo_http_timeout_seconds=_get_int_env("TEMPO_HTTP_TIMEOUT_SECONDS", 10),
        tempo_tenant_id=os.getenv("TEMPO_TENANT_ID", "").strip(),
        tempo_trace_limit=max(1, _get_int_env("TEMPO_TRACE_LIMIT", 5)),
        tempo_lookback_minutes=_get_non_negative_int_env("TEMPO_LOOKBACK_MINUTES", 15),
        tempo_forward_minutes=_get_non_negative_int_env("TEMPO_FORWARD_MINUTES", 5),
        loki_url=os.getenv("LOKI_URL", "").strip(),
        loki_http_timeout_seconds=_get_int_env("LOKI_HTTP_TIMEOUT_SECONDS", 10),
        loki_tenant_id=os.getenv("LOKI_TENANT_ID", "").strip(),
        prompt_token_budget=_get_non_negative_int_env("PROMPT_TOKEN_BUDGET", 32000),
        prompt_max_log_lines=_get_non_negative_int_env("PROMPT_MAX_LOG_LINES", 25),
        prompt_max_events=_get_non_negative_int_env("PROMPT_MAX_EVENTS", 25),
        prompt_summary_max_items=max(1, _get_int_env("PROMPT_SUMMARY_MAX_ITEMS", 3)),
        masking_regex_list=masking_regex_list,
        # Builtin redaction
        builtin_redaction_enabled=(
            os.getenv("BUILTIN_REDACTION_ENABLED", "true").lower() != "false"
        ),
        builtin_redaction_hash_mode=(
            os.getenv("BUILTIN_REDACTION_HASH_MODE", "false").lower() == "true"
        ),
        # LLM Retry
        llm_retry_max_attempts=_get_int_env("LLM_RETRY_MAX_ATTEMPTS", 10),
        llm_retry_min_wait=_get_float_env("LLM_RETRY_MIN_WAIT", 1.0),
        llm_retry_max_wait=_get_float_env("LLM_RETRY_MAX_WAIT", 30.0),
        llm_retry_total_timeout=_get_float_env("LLM_RETRY_TOTAL_TIMEOUT", 180.0),
        # Concurrency
        max_concurrent_analyses=_get_int_env("MAX_CONCURRENT_ANALYSES", 5),
        # Conversation manager sliding window
        agent_session_window_size=max(1, _get_int_env("AGENT_SESSION_WINDOW_SIZE", 40)),
        # Istio mesh CRD tools (Helm-controlled)
        agent_istio_enabled=os.getenv("AGENT_ISTIO_ENABLED", "false").lower() == "true",
    )
