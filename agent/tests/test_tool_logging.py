from __future__ import annotations

from app.clients.strands_agent import (
    _prometheus_range_result_summary,
    _prometheus_range_summary,
)


def test_prometheus_range_summary_includes_safe_window_metadata() -> None:
    summary = _prometheus_range_summary(
        {
            "query": 'sum(rate(istio_requests_total{destination_service_name="ratings"}[5m]))',
            "start": "2026-03-09T14:00:00Z",
            "end": "2026-03-09T15:00:00Z",
            "step": "30s",
        }
    )

    assert summary["query_length"] == 71
    assert summary["window_seconds"] == 3600.0
    assert summary["step_seconds"] == 30.0
    assert summary["query_hash"]
    assert "istio_requests_total" in str(summary["query_preview"])


def test_prometheus_range_result_summary_counts_series_and_samples() -> None:
    summary = _prometheus_range_result_summary(
        {
            "endpoint": {"base_url": "http://prometheus"},
            "data": {
                "status": "success",
                "data": {
                    "resultType": "matrix",
                    "result": [
                        {"metric": {"pod": "a"}, "values": [[1, "2"], [2, "3"]]},
                        {"metric": {"pod": "b"}, "values": [[1, "4"]]},
                    ],
                },
            },
        }
    )

    assert summary["prom_result_type"] == "matrix"
    assert summary["series_count"] == 2
    assert summary["sample_count"] == 3
