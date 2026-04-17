from __future__ import annotations

from fastapi import APIRouter, Depends, Request

from app.core.concurrency import run_in_thread_limited
from app.core.dependencies import get_analysis_service
from app.schemas.analysis import (
    AlertAnalysisRequest,
    AlertAnalysisResponse,
    IncidentSummaryRequest,
    IncidentSummaryResponse,
)
from app.services.analysis import AnalysisService

router = APIRouter()


@router.post("/analyze", response_model=AlertAnalysisResponse)
async def analyze_alert(
    http_request: Request,
    request: AlertAnalysisRequest,
    service: AnalysisService = Depends(get_analysis_service),  # noqa: B008
) -> AlertAnalysisResponse:
    (
        analysis,
        summary,
        detail,
        summary_i18n,
        detail_i18n,
        context,
        artifacts,
    ) = await run_in_thread_limited(
        service.analyze, request, request=http_request
    )
    analysis_quality = _extract_optional_str(context, "analysis_quality")
    missing_data = _extract_optional_str_list(context, "missing_data")
    warnings = _extract_optional_str_list(context, "warnings")
    capabilities = _extract_optional_str_dict(context, "capabilities")
    analysis_type = request.analysis_type or request.alert.status
    return AlertAnalysisResponse(
        status="ok",
        thread_ts=request.thread_ts,
        analysis=analysis,
        analysis_summary=summary,
        analysis_detail=detail,
        analysis_summary_i18n=summary_i18n,
        analysis_detail_i18n=detail_i18n,
        analysis_type=analysis_type,
        analysis_quality=analysis_quality,
        missing_data=missing_data,
        warnings=warnings,
        capabilities=capabilities,
        context=context,
        artifacts=artifacts,
    )


@router.post("/summarize-incident", response_model=IncidentSummaryResponse)
async def summarize_incident(
    http_request: Request,
    request: IncidentSummaryRequest,
    service: AnalysisService = Depends(get_analysis_service),  # noqa: B008
) -> IncidentSummaryResponse:
    """Generate final RCA summary for a resolved incident."""
    title, summary, detail, summary_i18n, detail_i18n = await run_in_thread_limited(
        service.summarize_incident, request, request=http_request
    )
    return IncidentSummaryResponse(
        status="ok",
        title=title,
        summary=summary,
        detail=detail,
        summary_i18n=summary_i18n,
        detail_i18n=detail_i18n,
    )


def _extract_optional_str(context: dict[str, object] | None, key: str) -> str | None:
    if not isinstance(context, dict):
        return None
    value = context.get(key)
    if isinstance(value, str):
        return value
    return None


def _extract_optional_str_list(context: dict[str, object] | None, key: str) -> list[str] | None:
    if not isinstance(context, dict):
        return None
    value = context.get(key)
    if not isinstance(value, list):
        return None
    items = [item for item in value if isinstance(item, str)]
    return items or None


def _extract_optional_str_dict(
    context: dict[str, object] | None, key: str
) -> dict[str, str] | None:
    if not isinstance(context, dict):
        return None
    value = context.get(key)
    if not isinstance(value, dict):
        return None
    output: dict[str, str] = {}
    for map_key, map_value in value.items():
        if isinstance(map_key, str) and isinstance(map_value, str):
            output[map_key] = map_value
    return output or None
