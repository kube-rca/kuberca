from __future__ import annotations

import urllib.error
import urllib.parse

import pytest

import app.clients.loki as loki_module
from app.clients.loki import LokiClient
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


def _install_fake_urlopen(
    monkeypatch: pytest.MonkeyPatch, body: str = '{"data":{}}'
) -> dict[str, object]:
    captured: dict[str, object] = {}

    def fake_urlopen(request, timeout=0):  # type: ignore[no-untyped-def]
        captured["url"] = request.full_url
        captured["timeout"] = timeout
        captured["headers"] = dict(request.headers)
        return _FakeHTTPResponse(body)

    monkeypatch.setattr(loki_module.urllib.request, "urlopen", fake_urlopen)
    return captured


def test_disabled_when_no_url(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("LOKI_URL", raising=False)
    client = LokiClient(load_settings())
    assert not client.enabled


def test_enabled_when_url_set(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LOKI_URL", "http://loki.monitoring.svc:3100")
    client = LokiClient(load_settings())
    assert client.enabled


def test_query_sends_logql(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LOKI_URL", "http://loki.monitoring.svc:3100")
    captured = _install_fake_urlopen(monkeypatch)

    client = LokiClient(load_settings())
    client.query('{namespace="bookinfo"}', limit=50)

    parsed = urllib.parse.urlparse(str(captured["url"]))
    params = urllib.parse.parse_qs(parsed.query)
    assert parsed.path == "/loki/api/v1/query"
    assert params["query"] == ['{namespace="bookinfo"}']
    assert params["limit"] == ["50"]


def test_query_range_sends_window(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LOKI_URL", "http://loki.monitoring.svc:3100")
    captured = _install_fake_urlopen(monkeypatch)

    client = LokiClient(load_settings())
    client.query_range(
        '{namespace="bookinfo"} |= "error"',
        start="1770402600",
        end="1770403200",
        limit=200,
    )

    parsed = urllib.parse.urlparse(str(captured["url"]))
    params = urllib.parse.parse_qs(parsed.query)
    assert parsed.path == "/loki/api/v1/query_range"
    assert params["start"] == ["1770402600"]
    assert params["end"] == ["1770403200"]
    assert params["limit"] == ["200"]


def test_tenant_id_header(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LOKI_URL", "http://loki.monitoring.svc:3100")
    monkeypatch.setenv("LOKI_TENANT_ID", "my-tenant")
    captured = _install_fake_urlopen(monkeypatch)

    client = LokiClient(load_settings())
    client.list_labels()

    headers = captured.get("headers", {})
    assert headers.get("X-scope-orgid") == "my-tenant"


def test_list_labels_path(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LOKI_URL", "http://loki.monitoring.svc:3100")
    captured = _install_fake_urlopen(monkeypatch)

    client = LokiClient(load_settings())
    client.list_labels()

    parsed = urllib.parse.urlparse(str(captured["url"]))
    assert parsed.path == "/loki/api/v1/labels"


def test_label_values_path(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LOKI_URL", "http://loki.monitoring.svc:3100")
    captured = _install_fake_urlopen(monkeypatch)

    client = LokiClient(load_settings())
    client.label_values("namespace")

    parsed = urllib.parse.urlparse(str(captured["url"]))
    assert parsed.path == "/loki/api/v1/label/namespace/values"


def test_not_configured_returns_warning(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("LOKI_URL", raising=False)
    client = LokiClient(load_settings())
    result = client.query('{namespace="bookinfo"}')
    assert "warning" in result
    assert "not configured" in str(result["warning"])


def test_http_error_returns_error_dict(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LOKI_URL", "http://loki.monitoring.svc:3100")

    def fake_urlopen(request, timeout=0):  # type: ignore[no-untyped-def]
        raise urllib.error.HTTPError(
            url=request.full_url,
            code=503,
            msg="Service Unavailable",
            hdrs=None,  # type: ignore[arg-type]
            fp=None,
        )

    monkeypatch.setattr(loki_module.urllib.request, "urlopen", fake_urlopen)

    client = LokiClient(load_settings())
    result = client.query('{namespace="bookinfo"}')
    assert result["error"] == "failed to query loki"
    assert result["detail"]["status_code"] == 503  # type: ignore[index]
    assert "endpoint" in result


def test_json_decode_error_returns_error_dict(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LOKI_URL", "http://loki.monitoring.svc:3100")
    _install_fake_urlopen(monkeypatch, body="not-json{{{")

    client = LokiClient(load_settings())
    result = client.query('{namespace="bookinfo"}')
    assert result["error"] == "failed to decode loki response"
    assert "endpoint" in result


def test_connection_error_returns_error_dict(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("LOKI_URL", "http://loki.monitoring.svc:3100")

    def fake_urlopen(request, timeout=0):  # type: ignore[no-untyped-def]
        raise ConnectionError("Connection refused")

    monkeypatch.setattr(loki_module.urllib.request, "urlopen", fake_urlopen)

    client = LokiClient(load_settings())
    result = client.query('{namespace="bookinfo"}')
    assert result["error"] == "failed to query loki"
    assert "Connection refused" in str(result["detail"])
