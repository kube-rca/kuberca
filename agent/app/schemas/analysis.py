from __future__ import annotations

from pydantic import BaseModel

from app.schemas.alert import Alert


class AlertAnalysisRequest(BaseModel):
    alert: Alert
    thread_ts: str
    incident_id: str | None = None


class AlertAnalysisResponse(BaseModel):
    status: str
    thread_ts: str
    analysis: str
