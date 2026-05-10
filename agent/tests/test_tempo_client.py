from __future__ import annotations

import urllib.parse

import pytest

import app.clients.tempo as tempo_module
from app.clients.tempo import TempoClient, build_traceql_query
from app.core.config import load_settings


class _FakeHTTPResponse:
    def __init__(self, body: str) -> None:
        self._body = body.encode("utf-8")

    def read(self) -> bytes:
        return self._body

    def __enter__(self) -> _FakeHTTPResponse:
        return self

    def __exit__(self, exc_type, exc, tb) -> None:  # type: ignore[no-untyped-def]
        return None


def _install_fake_urlopen(monkeypatch: pytest.MonkeyPatch) -> dict[str, object]:
    captured: dict[str, object] = {}

    def fake_urlopen(request, timeout=0):  # type: ignore[no-untyped-def]
        captured["url"] = request.full_url
        captured["timeout"] = timeout
        return _FakeHTTPResponse('{"traces":[]}')

    monkeypatch.setattr(tempo_module.urllib.request, "urlopen", fake_urlopen)
    return captured


def test_search_traces_normalizes_rfc3339_window_to_unix(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.setenv("TEMPO_URL", "http://tempo.monitoring.svc.cluster.local:3200")
    captured = _install_fake_urlopen(monkeypatch)

    client = TempoClient(load_settings())
    client.search_traces(
        query='{ resource.service.name = "ratings" }',
        start="2026-02-06T18:30:00Z",
        end="2026-02-06T18:40:00Z",
    )

    parsed = urllib.parse.urlparse(str(captured["url"]))
    params = urllib.parse.parse_qs(parsed.query)
    assert params["start"] == ["1770402600"]
    assert params["end"] == ["1770403200"]


def test_build_traceql_query_uses_fqdn_service_name() -> None:
    query = build_traceql_query(service_name="ratings", namespace="bookinfo")
    assert query == '{ resource.service.name = "ratings.bookinfo" }'


def test_build_traceql_query_service_only() -> None:
    query = build_traceql_query(service_name="ratings", namespace=None)
    assert query == '{ resource.service.name = "ratings" }'


def test_build_traceql_query_namespace_only() -> None:
    # Use a regex char class `[.]` for the literal dot. TraceQL string literals
    # only accept \" \\ \n \t escapes; `\.` triggers a parse error
    # ("invalid char escape" at col 28) and Tempo returns HTTP 400.
    query = build_traceql_query(service_name=None, namespace="bookinfo")
    assert query == '{ resource.service.name =~ ".*[.]bookinfo" }'


def test_build_traceql_query_namespace_with_dot() -> None:
    # Namespaces may contain dots in some setups; every dot must remain literal
    # via char class, never as `\.` (which TraceQL rejects).
    query = build_traceql_query(service_name=None, namespace="team.bookinfo")
    assert query == '{ resource.service.name =~ ".*[.]team[.]bookinfo" }'


def test_build_traceql_query_namespace_only_avoids_invalid_traceql_escape() -> None:
    # Regression guard: the bytes `\.` MUST NOT appear in the namespace-only
    # fallback output. Previous implementation produced `.*\.<ns>` which Tempo
    # rejected with HTTP 400 "invalid char escape" (alert_analyses 422-426).
    query = build_traceql_query(service_name=None, namespace="bookinfo")
    assert "\\." not in query, f"output contains invalid TraceQL escape: {query!r}"


def test_build_traceql_query_empty() -> None:
    query = build_traceql_query(service_name=None, namespace=None)
    assert query == "{}"


def test_extract_trace_summaries_returns_empty_for_none_payload() -> None:
    # Caller (`search_traces`) returns early on error, but the function's type
    # signature now accepts None to mirror the union from `_request_with_fallback`.
    from app.clients.tempo import _extract_trace_summaries

    assert _extract_trace_summaries(None) == []


def test_extract_trace_summaries_falls_back_to_data_then_results() -> None:
    # When `traces` key is missing, the function should walk `data` then
    # `results` before giving up — preserves prior behavior across Tempo
    # response shape variants.
    from app.clients.tempo import _extract_trace_summaries

    payload_data: dict[str, object] = {"data": [{"traceID": "abc"}]}
    payload_results: dict[str, object] = {"results": [{"traceID": "def"}]}

    assert [s["trace_id"] for s in _extract_trace_summaries(payload_data)] == ["abc"]
    assert [s["trace_id"] for s in _extract_trace_summaries(payload_results)] == ["def"]


def test_search_traces_preserves_unix_window(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("TEMPO_URL", "http://tempo.monitoring.svc.cluster.local:3200")
    captured = _install_fake_urlopen(monkeypatch)

    client = TempoClient(load_settings())
    client.search_traces(
        query='{ resource.service.name = "ratings" }',
        start="1770402600",
        end="1770403200",
    )

    parsed = urllib.parse.urlparse(str(captured["url"]))
    params = urllib.parse.parse_qs(parsed.query)
    assert params["start"] == ["1770402600"]
    assert params["end"] == ["1770403200"]
