from __future__ import annotations

import os
from dataclasses import dataclass

DEFAULT_GEMINI_MODEL_ID = "gemini-3-flash-preview"


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


@dataclass(frozen=True)
class Settings:
    port: int
    log_level: str
    gemini_api_key: str
    gemini_model_id: str
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
    prompt_token_budget: int
    prompt_max_log_lines: int
    prompt_max_events: int
    prompt_summary_max_items: int

    @property
    def session_store_dsn(self) -> str:
        if not all([self.session_db_host, self.session_db_user, self.session_db_name]):
            return ""
        return (
            f"postgresql://{self.session_db_user}:{self.session_db_password}"
            f"@{self.session_db_host}:{self.session_db_port}/{self.session_db_name}"
        )


def load_settings() -> Settings:
    return Settings(
        port=_get_int_env("PORT", 8000),
        log_level=os.getenv("LOG_LEVEL", "info"),
        gemini_api_key=os.getenv("GEMINI_API_KEY", ""),
        gemini_model_id=os.getenv("GEMINI_MODEL_ID", DEFAULT_GEMINI_MODEL_ID),
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
        prompt_token_budget=_get_non_negative_int_env("PROMPT_TOKEN_BUDGET", 32000),
        prompt_max_log_lines=_get_non_negative_int_env("PROMPT_MAX_LOG_LINES", 25),
        prompt_max_events=_get_non_negative_int_env("PROMPT_MAX_EVENTS", 25),
        prompt_summary_max_items=max(1, _get_int_env("PROMPT_SUMMARY_MAX_ITEMS", 3)),
    )
