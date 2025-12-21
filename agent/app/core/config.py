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


@dataclass(frozen=True)
class Settings:
    port: int
    log_level: str
    gemini_api_key: str
    gemini_model_id: str
    k8s_api_timeout_seconds: int
    k8s_event_limit: int
    k8s_log_tail_lines: int


def load_settings() -> Settings:
    return Settings(
        port=_get_int_env("PORT", 8082),
        log_level=os.getenv("LOG_LEVEL", "info"),
        gemini_api_key=os.getenv("GEMINI_API_KEY", ""),
        gemini_model_id=os.getenv("GEMINI_MODEL_ID", DEFAULT_GEMINI_MODEL_ID),
        k8s_api_timeout_seconds=_get_int_env("K8S_API_TIMEOUT_SECONDS", 5),
        k8s_event_limit=_get_int_env("K8S_EVENT_LIMIT", 20),
        k8s_log_tail_lines=_get_int_env("K8S_LOG_TAIL_LINES", 50),
    )
