from __future__ import annotations

import logging
from functools import lru_cache

from app.clients.k8s import KubernetesClient
from app.clients.llm_providers import get_provider_config
from app.clients.prometheus import PrometheusClient
from app.clients.strands_agent import AnalysisEngine, StrandsAnalysisEngine
from app.clients.summary_store import PostgresSummaryStore, SummaryStore
from app.core.config import Settings, load_settings
from app.core.masking import RegexMasker, build_masker
from app.services.analysis import AnalysisService

logger = logging.getLogger(__name__)


@lru_cache
def get_settings() -> Settings:
    return load_settings()


@lru_cache
def get_k8s_client() -> KubernetesClient:
    settings = get_settings()
    event_limit = settings.k8s_event_limit
    if settings.prompt_max_events > 0:
        event_limit = min(event_limit, settings.prompt_max_events)
    log_tail_lines = settings.k8s_log_tail_lines
    if settings.prompt_max_log_lines > 0:
        log_tail_lines = min(log_tail_lines, settings.prompt_max_log_lines)
    return KubernetesClient(
        timeout_seconds=settings.k8s_api_timeout_seconds,
        event_limit=event_limit,
        log_tail_lines=log_tail_lines,
    )


@lru_cache
def get_prometheus_client() -> PrometheusClient | None:
    settings = get_settings()
    client = PrometheusClient(settings)
    if not client.enabled:
        return None
    return client


@lru_cache
def get_masker() -> RegexMasker:
    settings = get_settings()
    return build_masker(settings.masking_regex_list)


@lru_cache
def get_analysis_engine() -> AnalysisEngine | None:
    settings = get_settings()

    # Use multi-provider factory to get model configuration
    model_config = get_provider_config(settings)
    if model_config is None:
        logger.warning("No valid AI provider configured. Analysis engine disabled.")
        return None

    return StrandsAnalysisEngine(
        settings,
        get_k8s_client(),
        get_prometheus_client(),
        masker=get_masker(),
        model_config=model_config,
    )


@lru_cache
def get_summary_store() -> SummaryStore | None:
    settings = get_settings()
    if not settings.session_store_dsn:
        return None
    return PostgresSummaryStore(settings.session_store_dsn)


@lru_cache
def get_analysis_service() -> AnalysisService:
    prometheus_client = get_prometheus_client()
    settings = get_settings()
    return AnalysisService(
        get_k8s_client(),
        get_analysis_engine(),
        masker=get_masker(),
        prometheus_enabled=prometheus_client is not None,
        summary_store=get_summary_store(),
        summary_history_size=settings.prompt_summary_max_items,
        prompt_token_budget=settings.prompt_token_budget,
        prompt_max_log_lines=settings.prompt_max_log_lines,
        prompt_max_events=settings.prompt_max_events,
    )
