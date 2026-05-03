from __future__ import annotations

from app.clients.strands_agent import _build_tools
from app.core.masking import RegexMasker


def _tool_names(
    *,
    prometheus: object | None,
    tempo: object | None,
    loki: object | None,
    istio_enabled: bool = False,
) -> set[str]:
    tools = _build_tools(
        k8s_client=object(),
        prometheus_client=prometheus,
        tempo_client=tempo,
        loki_client=loki,
        masker=RegexMasker(),
        istio_enabled=istio_enabled,
    )
    return {tool.tool_name for tool in tools}


def test_build_tools_registers_portable_wrapper_tools() -> None:
    names = _tool_names(prometheus=object(), tempo=object(), loki=object(), istio_enabled=True)

    assert "get_service" in names
    assert "get_endpoints" in names
    assert "list_virtual_services" in names
    assert "list_destination_rules" in names
    assert "list_service_entries" in names
    assert "query_prometheus_range" in names
    assert "search_tempo_traces" in names
    assert "query_loki_range" in names


def test_build_tools_omits_optional_observability_tools_when_clients_are_missing() -> None:
    names = _tool_names(prometheus=None, tempo=None, loki=None, istio_enabled=True)

    assert "get_service" in names
    assert "get_endpoints" in names
    assert "list_virtual_services" in names
    assert "query_prometheus" not in names
    assert "search_tempo_traces" not in names
    assert "query_loki" not in names


def test_build_tools_omits_istio_tools_when_istio_disabled() -> None:
    names = _tool_names(
        prometheus=object(),
        tempo=object(),
        loki=object(),
        istio_enabled=False,
    )

    assert "get_service" in names
    assert "list_virtual_services" not in names
    assert "list_destination_rules" not in names
    assert "list_service_entries" not in names
