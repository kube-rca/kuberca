from __future__ import annotations

from functools import lru_cache

from app.clients.k8s import KubernetesClient
from app.clients.strands_agent import AnalysisEngine, StrandsAnalysisEngine
from app.core.config import Settings, load_settings
from app.services.analysis import AnalysisService


@lru_cache
def get_settings() -> Settings:
    return load_settings()


@lru_cache
def get_k8s_client() -> KubernetesClient:
    settings = get_settings()
    return KubernetesClient(
        timeout_seconds=settings.k8s_api_timeout_seconds,
        event_limit=settings.k8s_event_limit,
        log_tail_lines=settings.k8s_log_tail_lines,
    )


@lru_cache
def get_analysis_engine() -> AnalysisEngine | None:
    settings = get_settings()
    if not settings.gemini_api_key:
        return None
    return StrandsAnalysisEngine(settings, get_k8s_client())


@lru_cache
def get_analysis_service() -> AnalysisService:
    return AnalysisService(get_k8s_client(), get_analysis_engine())
