"""Tests for the /analyze and /summarize-incident API endpoints.

Builds a minimal FastAPI test app directly from the analysis router (no
lifespan, no real K8s / LLM / DB calls) and injects a fake AnalysisService
via dependency_overrides.
"""

from __future__ import annotations

from typing import Any

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.api import analysis as analysis_router_module
from app.api import health as health_router_module
from app.core.dependencies import get_analysis_service

# ---------------------------------------------------------------------------
# Minimal fake AnalysisService stubs
# ---------------------------------------------------------------------------


class _FakeAnalysisService:
    """Happy-path stub: returns canned strings, never calls LLM or K8s."""

    def analyze_with_i18n(
        self, _request: Any, **_kwargs: Any
    ) -> tuple[str, str, str, dict, dict, dict, list]:
        return (
            "RCA analysis text",
            "Summary text",
            "Detail text",
            {"en": "English summary", "ko": "한국어 요약"},
            {"en": "English detail", "ko": "한국어 상세"},
            {"analysis_quality": "high"},
            [],
        )

    def summarize_incident_with_i18n(
        self, _request: Any, **_kwargs: Any
    ) -> tuple[str, str, str, dict, dict]:
        return (
            "Incident title",
            "Summary",
            "Detail",
            {"en": "English summary"},
            {"en": "English detail"},
        )


# ---------------------------------------------------------------------------
# Test payloads
# ---------------------------------------------------------------------------

_MINIMAL_ALERT_PAYLOAD: dict[str, Any] = {
    "alert": {
        "status": "firing",
        "labels": {"alertname": "TestAlert", "namespace": "default"},
        "annotations": {"summary": "Test alert"},
    },
    "thread_ts": "1234567890.123456",
}

_MINIMAL_INCIDENT_PAYLOAD: dict[str, Any] = {
    "incident_id": "inc-001",
    "title": "Test Incident",
    "severity": "warning",
    "fired_at": "2024-01-01T00:00:00Z",
    "resolved_at": "2024-01-01T01:00:00Z",
    "alerts": [
        {
            "fingerprint": "abc123",
            "alert_name": "TestAlert",
            "severity": "warning",
            "status": "resolved",
        }
    ],
}


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------


def _build_test_app() -> FastAPI:
    """Create a minimal FastAPI app that includes only the analysis + health
    routers with no lifespan, so no real services are initialised."""
    test_app = FastAPI()
    test_app.include_router(analysis_router_module.router)
    test_app.include_router(health_router_module.router)
    return test_app


@pytest.fixture
def client() -> TestClient:
    test_app = _build_test_app()
    test_app.dependency_overrides[get_analysis_service] = _FakeAnalysisService
    with TestClient(test_app, raise_server_exceptions=False) as c:
        yield c


# ---------------------------------------------------------------------------
# /analyze endpoint tests
# ---------------------------------------------------------------------------


class TestAnalyzeEndpoint:
    def test_analyze_returns_200_with_ok_status(self, client: TestClient) -> None:
        resp = client.post("/analyze", json=_MINIMAL_ALERT_PAYLOAD)
        assert resp.status_code == 200
        body = resp.json()
        assert body["status"] == "ok"

    def test_analyze_response_contains_analysis_text(self, client: TestClient) -> None:
        resp = client.post("/analyze", json=_MINIMAL_ALERT_PAYLOAD)
        assert resp.status_code == 200
        body = resp.json()
        assert isinstance(body.get("analysis"), str)
        assert len(body["analysis"]) > 0

    def test_analyze_response_echoes_thread_ts(self, client: TestClient) -> None:
        resp = client.post("/analyze", json=_MINIMAL_ALERT_PAYLOAD)
        assert resp.status_code == 200
        assert resp.json()["thread_ts"] == "1234567890.123456"

    def test_analyze_with_resolved_status(self, client: TestClient) -> None:
        payload = {
            **_MINIMAL_ALERT_PAYLOAD,
            "alert": {**_MINIMAL_ALERT_PAYLOAD["alert"], "status": "resolved"},
        }
        resp = client.post("/analyze", json=payload)
        assert resp.status_code == 200
        assert resp.json()["status"] == "ok"

    def test_analyze_with_incident_id(self, client: TestClient) -> None:
        payload = {**_MINIMAL_ALERT_PAYLOAD, "incident_id": "inc-99"}
        resp = client.post("/analyze", json=payload)
        assert resp.status_code == 200

    def test_analyze_with_explicit_analysis_type(self, client: TestClient) -> None:
        payload = {**_MINIMAL_ALERT_PAYLOAD, "analysis_type": "firing"}
        resp = client.post("/analyze", json=payload)
        assert resp.status_code == 200
        body = resp.json()
        assert body["analysis_type"] == "firing"

    def test_analyze_missing_alert_field_returns_422(self, client: TestClient) -> None:
        """Pydantic validation rejects payload without required 'alert' field."""
        resp = client.post("/analyze", json={"thread_ts": "ts"})
        assert resp.status_code == 422

    def test_analyze_missing_thread_ts_returns_422(self, client: TestClient) -> None:
        """Pydantic validation rejects payload without required 'thread_ts' field."""
        resp = client.post(
            "/analyze",
            json={"alert": {"status": "firing", "labels": {}}},
        )
        assert resp.status_code == 422

    def test_analyze_empty_body_returns_422(self, client: TestClient) -> None:
        resp = client.post("/analyze", json={})
        assert resp.status_code == 422

    def test_analyze_invalid_json_returns_422(self, client: TestClient) -> None:
        resp = client.post(
            "/analyze",
            content=b"not-json",
            headers={"Content-Type": "application/json"},
        )
        assert resp.status_code == 422

    def test_analyze_i18n_summary_included(self, client: TestClient) -> None:
        resp = client.post("/analyze", json=_MINIMAL_ALERT_PAYLOAD)
        body = resp.json()
        assert body.get("analysis_summary_i18n") is not None

    def test_analyze_firing_sets_analysis_type_to_firing(self, client: TestClient) -> None:
        """When no analysis_type is given, it defaults to alert.status."""
        payload = {
            **_MINIMAL_ALERT_PAYLOAD,
            "alert": {**_MINIMAL_ALERT_PAYLOAD["alert"], "status": "firing"},
        }
        resp = client.post("/analyze", json=payload)
        body = resp.json()
        assert body["analysis_type"] == "firing"


# ---------------------------------------------------------------------------
# /summarize-incident endpoint tests
# ---------------------------------------------------------------------------


class TestSummarizeIncidentEndpoint:
    def test_summarize_returns_200_with_ok_status(self, client: TestClient) -> None:
        resp = client.post("/summarize-incident", json=_MINIMAL_INCIDENT_PAYLOAD)
        assert resp.status_code == 200
        assert resp.json()["status"] == "ok"

    def test_summarize_response_has_required_fields(self, client: TestClient) -> None:
        resp = client.post("/summarize-incident", json=_MINIMAL_INCIDENT_PAYLOAD)
        body = resp.json()
        for field in ("title", "summary", "detail"):
            assert field in body, f"missing field: {field}"

    def test_summarize_missing_alerts_returns_422(self, client: TestClient) -> None:
        payload = {k: v for k, v in _MINIMAL_INCIDENT_PAYLOAD.items() if k != "alerts"}
        resp = client.post("/summarize-incident", json=payload)
        assert resp.status_code == 422

    def test_summarize_empty_body_returns_422(self, client: TestClient) -> None:
        resp = client.post("/summarize-incident", json={})
        assert resp.status_code == 422

    def test_summarize_with_enriched_alert(self, client: TestClient) -> None:
        payload = {
            **_MINIMAL_INCIDENT_PAYLOAD,
            "alerts": [
                {
                    **_MINIMAL_INCIDENT_PAYLOAD["alerts"][0],
                    "analysis_summary": "Prior summary",
                    "analysis_detail": "Prior detail",
                }
            ],
        }
        resp = client.post("/summarize-incident", json=payload)
        assert resp.status_code == 200


# ---------------------------------------------------------------------------
# /health endpoint sanity
# ---------------------------------------------------------------------------


class TestHealthEndpoint:
    def test_healthz_returns_200(self, client: TestClient) -> None:
        resp = client.get("/healthz")
        assert resp.status_code == 200

    def test_ping_returns_200(self, client: TestClient) -> None:
        resp = client.get("/ping")
        assert resp.status_code == 200
