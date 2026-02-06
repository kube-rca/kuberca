from __future__ import annotations

import urllib.parse

import pytest

import app.clients.tempo as tempo_module
from app.clients.tempo import TempoClient
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
