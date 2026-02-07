from __future__ import annotations

from pydantic import BaseModel

from app.schemas.alert import Alert


class AlertAnalysisRequest(BaseModel):
    alert: Alert
    thread_ts: str
    incident_id: str | None = None


class AlertAnalysisArtifact(BaseModel):
    type: str
    query: str | None = None
    result: dict[str, object] | list[object] | str | None = None
    summary: str | None = None


class AlertAnalysisResponse(BaseModel):
    status: str
    thread_ts: str
    analysis: str
    analysis_summary: str | None = None
    analysis_detail: str | None = None
    analysis_quality: str | None = None
    missing_data: list[str] | None = None
    warnings: list[str] | None = None
    capabilities: dict[str, str] | None = None
    context: dict[str, object] | None = None
    artifacts: list[AlertAnalysisArtifact] | None = None


# Incident Summary schemas (for final RCA when incident is resolved)
class AlertSummaryInput(BaseModel):
    fingerprint: str
    alert_name: str
    severity: str
    status: str
    analysis_summary: str | None = None
    analysis_detail: str | None = None
    artifacts: list[AlertAnalysisArtifact] | None = None


class IncidentSummaryRequest(BaseModel):
    incident_id: str
    title: str
    severity: str
    fired_at: str
    resolved_at: str
    alerts: list[AlertSummaryInput]


class IncidentSummaryResponse(BaseModel):
    status: str
    title: str
    summary: str
    detail: str
