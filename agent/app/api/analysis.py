from __future__ import annotations

from fastapi import APIRouter, Depends

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
def analyze_alert(
    request: AlertAnalysisRequest,
    service: AnalysisService = Depends(get_analysis_service),  # noqa: B008
) -> AlertAnalysisResponse:
    result = service.analyze(request)
    return AlertAnalysisResponse(status="ok", thread_ts=request.thread_ts, analysis=result)


@router.post("/summarize-incident", response_model=IncidentSummaryResponse)
def summarize_incident(
    request: IncidentSummaryRequest,
    service: AnalysisService = Depends(get_analysis_service),  # noqa: B008
) -> IncidentSummaryResponse:
    """Generate final RCA summary for a resolved incident."""
    title, summary, detail = service.summarize_incident(request)
    return IncidentSummaryResponse(status="ok", title=title, summary=summary, detail=detail)
