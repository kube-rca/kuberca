from __future__ import annotations

import json
import urllib.error
import urllib.parse

import pytest

import app.clients.prometheus as prometheus_module
from app.clients.prometheus import PrometheusClient
from app.core.config import load_settings

_METRICS_URL = "http://prometheus.monitoring.svc:9090"


class _FakePrometheusResponse:
    def __init__(self, body: str) -> None:
        self._body = body.encode("utf-8")

    def read(self) -> bytes:
        return self._body

    def __enter__(self) -> _FakePrometheusResponse:
        return self

    def __exit__(self, _exc_type, _exc, _tb) -> None:  # type: ignore[no-untyped-def]
        return None


def _install_fake_urlopen(
    monkeypatch: pytest.MonkeyPatch,
    body: str = '{"status":"success","data":[]}',
) -> dict[str, object]:
    captured: dict[str, object] = {}

    def fake_urlopen(request, timeout=0):  # type: ignore[no-untyped-def]
        captured["url"] = request if isinstance(request, str) else request.full_url
        captured["timeout"] = timeout
        return _FakePrometheusResponse(body)

    monkeypatch.setattr(prometheus_module.urllib.request, "urlopen", fake_urlopen)
    return captured


# ---------------------------------------------------------------------------
# enabled / disabled
# ---------------------------------------------------------------------------


def test_disabled_when_no_url(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("PROMETHEUS_URL", raising=False)
    client = PrometheusClient(load_settings())
    assert not client.enabled


def test_enabled_when_url_set(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    client = PrometheusClient(load_settings())
    assert client.enabled


def test_describe_endpoint_returns_base_url(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    client = PrometheusClient(load_settings())
    result = client.describe_endpoint()
    assert result["endpoint"]["base_url"] == _METRICS_URL  # type: ignore[index]


def test_describe_endpoint_disabled_returns_warning(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("PROMETHEUS_URL", raising=False)
    client = PrometheusClient(load_settings())
    result = client.describe_endpoint()
    assert "warning" in result


# ---------------------------------------------------------------------------
# list_metrics
# ---------------------------------------------------------------------------


def test_list_metrics_not_configured(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("PROMETHEUS_URL", raising=False)
    client = PrometheusClient(load_settings())
    result = client.list_metrics()
    assert "warning" in result


def test_list_metrics_no_match(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    body = json.dumps({"status": "success", "data": ["up", "kube_pod_status_phase"]})
    captured = _install_fake_urlopen(monkeypatch, body=body)

    client = PrometheusClient(load_settings())
    result = client.list_metrics()

    parsed = urllib.parse.urlparse(str(captured["url"]))
    assert parsed.path == "/api/v1/label/__name__/values"
    assert result["count"] == 2
    assert "up" in result["metrics"]  # type: ignore[operator]


def test_list_metrics_with_regex_match(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    body = json.dumps(
        {
            "status": "success",
            "data": ["up", "kube_pod_status_phase", "kube_node_status_condition"],
        }
    )
    _install_fake_urlopen(monkeypatch, body=body)

    client = PrometheusClient(load_settings())
    result = client.list_metrics(match="kube_pod.*")

    assert result["count"] == 1
    assert result["metrics"] == ["kube_pod_status_phase"]  # type: ignore[comparison-overlap]


def test_list_metrics_invalid_regex(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    body = json.dumps({"status": "success", "data": ["up"]})
    _install_fake_urlopen(monkeypatch, body=body)

    client = PrometheusClient(load_settings())
    result = client.list_metrics(match="[invalid(")

    assert result["error"] == "invalid regex pattern"
    assert "pattern" in result


def test_list_metrics_http_error(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)

    def fake_urlopen(request, timeout=0):  # type: ignore[no-untyped-def]  # noqa: ARG001
        raise urllib.error.HTTPError(
            url=str(request.full_url),
            code=503,
            msg="Service Unavailable",
            hdrs=None,  # type: ignore[arg-type]
            fp=None,
        )

    monkeypatch.setattr(prometheus_module.urllib.request, "urlopen", fake_urlopen)

    client = PrometheusClient(load_settings())
    result = client.list_metrics()

    assert result["error"] == "failed to list Prometheus metrics"
    assert "endpoint" in result


def test_list_metrics_json_decode_error(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    _install_fake_urlopen(monkeypatch, body="not-json{{")

    client = PrometheusClient(load_settings())
    result = client.list_metrics()

    assert result["error"] == "failed to decode Prometheus response"
    assert "endpoint" in result


# ---------------------------------------------------------------------------
# query
# ---------------------------------------------------------------------------


def test_query_not_configured(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("PROMETHEUS_URL", raising=False)
    client = PrometheusClient(load_settings())
    result = client.query("up")
    assert "warning" in result


def test_query_sends_correct_path_and_param(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    body = json.dumps({"status": "success", "data": {"resultType": "vector", "result": []}})
    captured = _install_fake_urlopen(monkeypatch, body=body)

    client = PrometheusClient(load_settings())
    client.query("up")

    parsed = urllib.parse.urlparse(str(captured["url"]))
    params = urllib.parse.parse_qs(parsed.query)
    assert parsed.path == "/api/v1/query"
    assert params["query"] == ["up"]


def test_query_with_time_param(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    body = json.dumps({"status": "success", "data": {"resultType": "vector", "result": []}})
    captured = _install_fake_urlopen(monkeypatch, body=body)

    client = PrometheusClient(load_settings())
    client.query("up", time="1770402600")

    params = urllib.parse.parse_qs(urllib.parse.urlparse(str(captured["url"])).query)
    assert params["time"] == ["1770402600"]


def test_query_http_error(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)

    def fake_urlopen(request, timeout=0):  # type: ignore[no-untyped-def]  # noqa: ARG001
        raise urllib.error.HTTPError(
            url=str(request.full_url),
            code=429,
            msg="Too Many Requests",
            hdrs=None,  # type: ignore[arg-type]
            fp=None,
        )

    monkeypatch.setattr(prometheus_module.urllib.request, "urlopen", fake_urlopen)

    client = PrometheusClient(load_settings())
    result = client.query("up")

    assert result["error"] == "failed to query Prometheus"
    assert "endpoint" in result


def test_query_json_decode_error(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    _install_fake_urlopen(monkeypatch, body="not-json{{{")

    client = PrometheusClient(load_settings())
    result = client.query("up")

    assert result["error"] == "failed to decode Prometheus response"
    assert "raw" in result
    assert "endpoint" in result


def test_query_connection_error(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)

    def fake_urlopen(_request, timeout=0):  # type: ignore[no-untyped-def]  # noqa: ARG001
        raise ConnectionError("connection refused")

    monkeypatch.setattr(prometheus_module.urllib.request, "urlopen", fake_urlopen)

    client = PrometheusClient(load_settings())
    result = client.query("up")

    assert result["error"] == "failed to query Prometheus"
    assert "connection refused" in str(result["detail"])


# ---------------------------------------------------------------------------
# query_range
# ---------------------------------------------------------------------------


def test_query_range_not_configured(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("PROMETHEUS_URL", raising=False)
    client = PrometheusClient(load_settings())
    result = client.query_range("up", start="1770402600", end="1770403200")
    assert "warning" in result


def test_query_range_sends_correct_params(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    body = json.dumps({"status": "success", "data": {"resultType": "matrix", "result": []}})
    captured = _install_fake_urlopen(monkeypatch, body=body)

    client = PrometheusClient(load_settings())
    client.query_range(
        'rate(http_requests_total{namespace="bookinfo"}[5m])',
        start="1770402600",
        end="1770403200",
        step="30s",
    )

    parsed = urllib.parse.urlparse(str(captured["url"]))
    params = urllib.parse.parse_qs(parsed.query)
    assert parsed.path == "/api/v1/query_range"
    assert params["start"] == ["1770402600"]
    assert params["end"] == ["1770403200"]
    assert params["step"] == ["30s"]


def test_query_range_default_step(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    body = json.dumps({"status": "success", "data": {"resultType": "matrix", "result": []}})
    captured = _install_fake_urlopen(monkeypatch, body=body)

    client = PrometheusClient(load_settings())
    client.query_range("up", start="1770402600", end="1770403200")

    params = urllib.parse.parse_qs(urllib.parse.urlparse(str(captured["url"])).query)
    assert params["step"] == ["1m"]


def test_query_range_http_error(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)

    def fake_urlopen(request, timeout=0):  # type: ignore[no-untyped-def]  # noqa: ARG001
        raise urllib.error.HTTPError(
            url=str(request.full_url),
            code=500,
            msg="Internal Server Error",
            hdrs=None,  # type: ignore[arg-type]
            fp=None,
        )

    monkeypatch.setattr(prometheus_module.urllib.request, "urlopen", fake_urlopen)

    client = PrometheusClient(load_settings())
    result = client.query_range("up", start="1770402600", end="1770403200")

    assert result["error"] == "failed to query_range Prometheus"
    assert "endpoint" in result


def test_query_range_json_decode_error(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL)
    _install_fake_urlopen(monkeypatch, body="bad-json")

    client = PrometheusClient(load_settings())
    result = client.query_range("up", start="1770402600", end="1770403200")

    assert result["error"] == "failed to decode Prometheus response"
    assert "raw" in result


# ---------------------------------------------------------------------------
# URL normalisation edge cases
# ---------------------------------------------------------------------------


def test_trailing_slash_stripped(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", _METRICS_URL + "/")
    client = PrometheusClient(load_settings())
    assert client.enabled
    result = client.describe_endpoint()
    assert not str(result["endpoint"]["base_url"]).endswith("/")  # type: ignore[index]


def test_scheme_added_when_missing(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", "prometheus.monitoring.svc:9090")
    client = PrometheusClient(load_settings())
    assert client.enabled


def test_invalid_scheme_disables_client(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("PROMETHEUS_URL", "ftp://prometheus.monitoring.svc:9090")
    client = PrometheusClient(load_settings())
    assert not client.enabled
