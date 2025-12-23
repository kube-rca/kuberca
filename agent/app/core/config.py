from __future__ import annotations

import os
from dataclasses import dataclass

DEFAULT_GEMINI_MODEL_ID = "gemini-3-flash-preview"
DEFAULT_PROMETHEUS_LABEL_SELECTOR = "app=kube-prometheus-stack-prometheus"


def _get_int_env(name: str, default: int) -> int:
    value = os.getenv(name)
    if not value:
        return default
    try:
        return int(value)
    except ValueError:
        return default


def _get_list_env(name: str) -> list[str]:
    value = os.getenv(name, "")
    if not value:
        return []
    return [item.strip() for item in value.split(",") if item.strip()]


@dataclass(frozen=True)
class Settings:
    port: int
    log_level: str
    gemini_api_key: str
    gemini_model_id: str
    k8s_api_timeout_seconds: int
    k8s_event_limit: int
    k8s_log_tail_lines: int
    prometheus_label_selector: str
    prometheus_namespace_allowlist: list[str]
    prometheus_port_name: str
    prometheus_scheme: str
    prometheus_http_timeout_seconds: int


def load_settings() -> Settings:
    return Settings(
        port=_get_int_env("PORT", 8000),
        log_level=os.getenv("LOG_LEVEL", "info"),
        gemini_api_key=os.getenv("GEMINI_API_KEY", ""),
        gemini_model_id=os.getenv("GEMINI_MODEL_ID", DEFAULT_GEMINI_MODEL_ID),
        k8s_api_timeout_seconds=_get_int_env("K8S_API_TIMEOUT_SECONDS", 5),
        k8s_event_limit=_get_int_env("K8S_EVENT_LIMIT", 20),
        k8s_log_tail_lines=_get_int_env("K8S_LOG_TAIL_LINES", 50),
        prometheus_label_selector=os.getenv(
            "PROMETHEUS_LABEL_SELECTOR", DEFAULT_PROMETHEUS_LABEL_SELECTOR
        ),
        prometheus_namespace_allowlist=_get_list_env("PROMETHEUS_NAMESPACE_ALLOWLIST"),
        prometheus_port_name=os.getenv("PROMETHEUS_PORT_NAME", ""),
        prometheus_scheme=os.getenv("PROMETHEUS_SCHEME", "http"),
        prometheus_http_timeout_seconds=_get_int_env("PROMETHEUS_HTTP_TIMEOUT_SECONDS", 5),
    )
