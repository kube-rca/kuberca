from __future__ import annotations

import json
from datetime import datetime, timedelta, timezone

from app.clients.k8s import resolve_alert_target
from app.core.masking import RegexMasker
from app.models.k8s import (
    AnalysisTarget,
    K8sContext,
    PodEventSummary,
    PodLogSnippet,
    PodStatusSnapshot,
)
from app.schemas.alert import Alert
from app.schemas.analysis import (
    AlertAnalysisRequest,
    AlertSummaryInput,
    IncidentSummaryRequest,
    PreviousAnalysisContext,
)
from app.services.analysis import (
    AnalysisService,
    _categorize_analysis_error,
    _collect_missing_data,
    _extract_first_paragraph,
    _parse_incident_summary,
    _resolve_analysis_quality,
)


class FakeKubernetesClient:
    def __init__(self, context: K8sContext) -> None:
        self._context = context

    def collect_context(
        self,
        namespace: str | None,
        pod_name: str | None,
        workload: str | None = None,
        service_name: str | None = None,
        node_name: str | None = None,
    ) -> K8sContext:
        return K8sContext(
            namespace=self._context.namespace,
            pod_name=self._context.pod_name,
            workload=self._context.workload,
            pod_status=self._context.pod_status,
            events=self._context.events,
            previous_logs=self._context.previous_logs,
            warnings=self._context.warnings,
            target=AnalysisTarget(
                namespace=namespace,
                pod_name=pod_name,
                workload=workload,
                service_name=service_name,
                node_name=node_name,
            ),
            current_logs=self._context.current_logs,
            pod_spec=self._context.pod_spec,
            workload_status=self._context.workload_status,
            pod_metrics=self._context.pod_metrics,
            node_status=self._context.node_status,
            service_manifest=self._context.service_manifest,
            endpoints_manifest=self._context.endpoints_manifest,
        )


class FakeAnalysisEngine:
    def __init__(self, result: str) -> None:
        self._result = result

    def analyze(self, prompt: str, incident_id: str | None = None) -> str:
        return self._result


class CapturingAnalysisEngine(FakeAnalysisEngine):
    def __init__(self, result: str) -> None:
        super().__init__(result)
        self.last_prompt = ""

    def analyze(self, prompt: str, incident_id: str | None = None) -> str:
        self.last_prompt = prompt
        return super().analyze(prompt, incident_id)


class FakeSummaryStore:
    def __init__(self, summaries: list[str]) -> None:
        self._summaries = summaries
        self.appended: list[tuple[str, str]] = []
        self.last_session_id: str | None = None

    def list_summaries(self, session_id: str, limit: int) -> list[str]:
        self.last_session_id = session_id
        return self._summaries[:limit]

    def append_summary(self, session_id: str, summary: str, max_items: int) -> None:
        self.appended.append((session_id, summary))


class FakeTempoClient:
    def __init__(self) -> None:
        self.calls: list[dict[str, object]] = []

    def search_traces(
        self,
        *,
        query: str,
        start: str,
        end: str,
        limit: int = 5,
    ) -> dict[str, object]:
        self.calls.append(
            {
                "query": query,
                "start": start,
                "end": end,
                "limit": limit,
            }
        )
        return {
            "trace_count": 1,
            "traces": [
                {
                    "trace_id": "trace-123",
                    "root_service_name": "reviews",
                    "root_trace_name": "GET /reviews",
                    "duration_ms": 245.1,
                }
            ],
        }


class FakeFailingTempoClient:
    def search_traces(
        self,
        *,
        query: str,
        start: str,
        end: str,
        limit: int = 5,
    ) -> dict[str, object]:
        return {
            "error": "failed to search tempo traces",
            "detail": {
                "status_code": 400,
                "reason": "Bad Request",
            },
            "trace_count": 0,
            "traces": [],
        }


def _sample_request() -> AlertAnalysisRequest:
    return AlertAnalysisRequest(
        alert=Alert(
            status="firing",
            labels={"namespace": "default", "pod": "demo-pod"},
            annotations={"summary": "Test"},
            fingerprint="abc123",
        ),
        thread_ts="1234567890.123456",
    )


def _sample_request_with_service() -> AlertAnalysisRequest:
    return AlertAnalysisRequest(
        alert=Alert(
            status="firing",
            labels={
                "namespace": "default",
                "pod": "demo-pod",
                "service": "demo-svc",
            },
            annotations={"summary": "Test"},
            fingerprint="abc123",
        ),
        thread_ts="1234567890.123456",
    )


def _sample_request_with_starts_at(starts_at: str) -> AlertAnalysisRequest:
    return AlertAnalysisRequest(
        alert=Alert(
            status="firing",
            labels={
                "namespace": "bookinfo",
                "destination_service_name": "reviews",
                "destination_service_namespace": "bookinfo",
            },
            annotations={"summary": "Test", "description": "trace test"},
            startsAt=datetime.fromisoformat(starts_at.replace("Z", "+00:00")),
            fingerprint="trace-abc123",
        ),
        thread_ts="1234567890.123456",
    )


def test_analysis_service_fallback() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=PodStatusSnapshot(
            phase="CrashLoopBackOff",
            node_name="node-a",
            start_time=None,
            reason=None,
            message=None,
            conditions=[],
            container_statuses=[],
        ),
        events=[],
        previous_logs=[],
        warnings=["namespace/pod_name missing from alert labels"],
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=None,
        prometheus_enabled=False,
    )

    analysis, summary, detail, context, artifacts = service.analyze(_sample_request())

    assert "analysis engine unavailable" in analysis
    assert summary
    assert detail
    assert context
    assert isinstance(artifacts, list)
    assert context.get("analysis_quality") == "low"
    assert "analysis_engine.not_configured" in context.get("missing_data", [])
    capabilities = context.get("capabilities")
    assert isinstance(capabilities, dict)
    assert capabilities.get("mesh_type") == "unknown"
    assert capabilities.get("routing_evidence") == "unavailable"


def test_analysis_service_quality_high_with_only_optional_missing_data() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=PodStatusSnapshot(
            phase="Running",
            node_name="node-a",
            start_time=None,
            reason=None,
            message=None,
            conditions=[],
            container_statuses=[],
        ),
        events=[
            PodEventSummary(
                type="Normal",
                reason="Pulled",
                message="Container image pulled",
                count=1,
                first_timestamp=None,
                last_timestamp=None,
                involved_object=None,
            )
        ],
        previous_logs=[PodLogSnippet(container="app", previous=True, logs=["ok"])],
        warnings=[],
        current_logs=[PodLogSnippet(container="app", previous=False, logs=["live"])],
        pod_spec={
            "containers": [{"name": "app", "image": "demo"}],
            "init_containers": [],
        },
        workload_status={"kind": "Deployment", "name": "demo", "ready_replicas": 1},
        service_manifest={"metadata": {"name": "demo-svc"}, "spec": {"selector": {"app": "demo"}}},
        endpoints_manifest={"metadata": {"name": "demo-svc"}, "subsets": [{"addresses": [{}]}]},
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=FakeAnalysisEngine("ok"),
        prometheus_enabled=False,
    )

    _, _, _, response_context, _ = service.analyze(_sample_request_with_service())

    assert response_context.get("analysis_quality") == "high"
    assert "istio.routing_manifest" not in response_context.get("missing_data", [])


def test_analysis_service_service_target_missing_still_high_quality() -> None:
    """Pod-level alerts without service label can still achieve high quality."""
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=PodStatusSnapshot(
            phase="Running",
            node_name="node-a",
            start_time=None,
            reason=None,
            message=None,
            conditions=[],
            container_statuses=[],
        ),
        events=[],
        previous_logs=[PodLogSnippet(container="app", previous=True, logs=["ok"])],
        warnings=[],
        current_logs=[PodLogSnippet(container="app", previous=False, logs=["live"])],
        pod_spec={"containers": [{"name": "app", "image": "demo"}], "init_containers": []},
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=FakeAnalysisEngine("ok"),
        prometheus_enabled=False,
    )

    _, _, _, response_context, _ = service.analyze(_sample_request())

    assert response_context.get("analysis_quality") == "high"
    assert "analysis.target.service_name" in response_context.get("missing_data", [])


def test_analysis_service_uses_engine() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=FakeAnalysisEngine("ok"),
        prometheus_enabled=False,
    )

    analysis, summary, detail, context, artifacts = service.analyze(_sample_request())

    assert analysis == "ok"
    assert summary
    assert detail
    assert context
    assert isinstance(artifacts, list)


def test_analysis_service_prompt_without_prometheus() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    engine = CapturingAnalysisEngine("ok")
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        prometheus_enabled=False,
    )

    service.analyze(_sample_request())

    assert "query_prometheus" not in engine.last_prompt
    assert "For Prometheus queries" not in engine.last_prompt


def test_analysis_service_prompt_with_prometheus() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    engine = CapturingAnalysisEngine("ok")
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        prometheus_enabled=True,
    )

    service.analyze(_sample_request())

    assert "query_prometheus" in engine.last_prompt


def test_analysis_service_prompt_with_tempo() -> None:
    context = K8sContext(
        namespace="bookinfo",
        pod_name="reviews-v1",
        workload="reviews",
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    engine = CapturingAnalysisEngine("ok")
    tempo = FakeTempoClient()
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        prometheus_enabled=False,
        tempo_client=tempo,
        tempo_enabled=True,
    )

    service.analyze(_sample_request_with_starts_at("2024-01-01T12:00:00Z"))

    assert "search_tempo_traces" in engine.last_prompt
    assert "get_tempo_trace" in engine.last_prompt
    assert "get_service" in engine.last_prompt
    assert "get_endpoints" in engine.last_prompt
    assert "get_manifest" in engine.last_prompt
    assert "list_manifests" in engine.last_prompt
    assert '"tempo"' in engine.last_prompt
    assert '"capabilities"' in engine.last_prompt
    assert '"missing_data"' in engine.last_prompt
    assert "kubectl logs" not in engine.last_prompt
    assert "istioctl proxy-config" not in engine.last_prompt
    assert tempo.calls


def test_analysis_service_prompt_with_loki() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    engine = CapturingAnalysisEngine("ok")
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        loki_enabled=True,
    )

    service.analyze(_sample_request())

    assert "query_loki_range" in engine.last_prompt
    assert "list_loki_labels" in engine.last_prompt


def test_analysis_service_detects_istio_mesh_and_uses_wrapper_tools() -> None:
    context = K8sContext(
        namespace="bookinfo",
        pod_name="ratings-v1",
        workload="ratings",
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
        pod_spec={
            "containers": [
                {"name": "ratings", "image": "ratings:v1"},
                {"name": "istio-proxy", "image": "proxy:v1"},
            ],
            "init_containers": [],
        },
    )
    engine = CapturingAnalysisEngine("ok")
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        prometheus_enabled=False,
    )

    _, _, _, response_context, _ = service.analyze(
        _sample_request_with_starts_at("2024-01-01T12:00:00Z")
    )

    capabilities = response_context.get("capabilities")
    assert isinstance(capabilities, dict)
    assert capabilities.get("mesh_type") == "istio"
    assert capabilities.get("routing_evidence") == "manifest_only"
    assert "list_virtual_services" in engine.last_prompt
    assert "list_destination_rules" in engine.last_prompt
    assert "list_service_entries" in engine.last_prompt


def test_analysis_service_adds_tempo_trace_artifacts() -> None:
    context = K8sContext(
        namespace="bookinfo",
        pod_name="reviews-v1",
        workload="reviews",
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    tempo = FakeTempoClient()
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=FakeAnalysisEngine("ok"),
        tempo_client=tempo,
        tempo_enabled=True,
    )

    _, _, _, api_context, artifacts = service.analyze(
        _sample_request_with_starts_at("2024-01-01T12:00:00Z")
    )

    artifact_types = {item["type"] for item in artifacts}
    assert "trace_summary" in artifact_types
    assert "trace" in artifact_types
    assert "tempo" in api_context


def test_analysis_service_tempo_window_uses_starts_at() -> None:
    context = K8sContext(
        namespace="bookinfo",
        pod_name="reviews-v1",
        workload="reviews",
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    tempo = FakeTempoClient()
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=FakeAnalysisEngine("ok"),
        tempo_client=tempo,
        tempo_enabled=True,
        tempo_lookback_minutes=15,
        tempo_forward_minutes=5,
    )

    starts_at = datetime(2024, 1, 1, 12, 0, tzinfo=timezone.utc)
    request = _sample_request_with_starts_at("2024-01-01T12:00:00Z")
    service.analyze(request)

    assert tempo.calls
    call = tempo.calls[-1]
    expected_start = (starts_at - timedelta(minutes=15)).isoformat().replace("+00:00", "Z")
    expected_end = (starts_at + timedelta(minutes=5)).isoformat().replace("+00:00", "Z")
    assert call["start"] == expected_start
    assert call["end"] == expected_end


def test_analysis_service_marks_tempo_query_error() -> None:
    context = K8sContext(
        namespace="bookinfo",
        pod_name="ratings-v1",
        workload="ratings",
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    engine = CapturingAnalysisEngine("ok")
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        tempo_client=FakeFailingTempoClient(),
        tempo_enabled=True,
    )

    request = _sample_request_with_starts_at("2024-01-01T12:00:00Z")
    _, _, _, api_context, artifacts = service.analyze(request)

    tempo_context = api_context.get("tempo")
    assert isinstance(tempo_context, dict)
    assert tempo_context.get("query_status") == "error"
    assert "tempo_query_detail: status=400, reason=Bad Request" in tempo_context.get("warnings", [])
    assert any(item.get("type") == "trace_warning" for item in artifacts)
    assert '"query_status": "error"' in engine.last_prompt


def test_analysis_service_includes_recent_summaries() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    store = FakeSummaryStore(["첫 번째 요약", "두 번째 요약", "세 번째 요약"])
    engine = CapturingAnalysisEngine("ok")
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        prometheus_enabled=False,
        summary_store=store,
        summary_history_size=3,
    )

    service.analyze(_sample_request())

    assert "Recent session summaries" in engine.last_prompt
    assert "첫 번째 요약" in engine.last_prompt
    assert store.appended


def test_analysis_service_trims_log_lines() -> None:
    logs = [f"line-{idx}" for idx in range(6)]
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[
            PodLogSnippet(container="app", previous=True, logs=logs),
        ],
        warnings=[],
    )
    engine = CapturingAnalysisEngine("ok")
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        prometheus_enabled=False,
        prompt_max_log_lines=2,
    )

    service.analyze(_sample_request())

    assert "line-0" not in engine.last_prompt
    assert "line-5" in engine.last_prompt


def test_analysis_service_masks_prompt_response_and_store() -> None:
    secret = "token-123456"
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[
            PodLogSnippet(container="app", previous=True, logs=[f"error: {secret}"]),
        ],
        warnings=[f"warning: {secret}"],
    )
    store = FakeSummaryStore([f"recent summary {secret}"])
    engine = CapturingAnalysisEngine(
        "### 1) 요약 (Summary)\n"
        f"LLM summary {secret}\n"
        "### 2) 상세 분석 (Detail)\n"
        f"LLM detail {secret}\n"
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        masker=RegexMasker.from_patterns([r"token-\d+"]),
        prometheus_enabled=False,
        summary_store=store,
    )
    request = AlertAnalysisRequest(
        alert=Alert(
            status="firing",
            labels={"namespace": "default", "pod": "demo-pod"},
            annotations={"summary": f"annotation {secret}"},
            fingerprint="abc123",
        ),
        thread_ts="1234567890.123456",
    )

    analysis, summary, detail, response_context, artifacts = service.analyze(request)
    rendered = json.dumps(
        {
            "analysis": analysis,
            "summary": summary,
            "detail": detail,
            "context": response_context,
            "artifacts": artifacts,
        },
        ensure_ascii=False,
    )

    assert secret not in engine.last_prompt
    assert "[MASKED]" in engine.last_prompt
    assert secret not in rendered
    assert "[MASKED]" in rendered
    assert store.appended
    assert secret not in store.appended[0][1]
    assert "[MASKED]" in store.appended[0][1]


def _sample_resolved_request_with_previous() -> AlertAnalysisRequest:
    return AlertAnalysisRequest(
        alert=Alert(
            status="resolved",
            labels={"namespace": "default", "pod": "demo-pod"},
            annotations={"summary": "Test resolved"},
            fingerprint="abc123",
            endsAt=datetime(2024, 1, 1, 12, 30, tzinfo=timezone.utc),
        ),
        thread_ts="1234567890.123456",
        analysis_type="resolved",
        previous_analysis=PreviousAnalysisContext(
            status="firing",
            summary="OOMKilled로 인한 Pod 재시작",
            detail="#### 근본 원인\n- 메모리 부족",
            created_at="2024-01-01T12:00:00Z",
        ),
    )


def test_resolved_prompt_uses_resolved_structure() -> None:
    """Resolved alert with previous_analysis uses resolved prompt structure."""
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    engine = CapturingAnalysisEngine("ok")
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        prometheus_enabled=False,
    )

    service.analyze(_sample_resolved_request_with_previous())

    assert "RESOLVED alert" in engine.last_prompt
    assert "복구 확인" in engine.last_prompt or "Recovery Confirmation" in engine.last_prompt
    assert "DO NOT repeat" in engine.last_prompt
    assert "OOMKilled" in engine.last_prompt


def test_resolved_without_previous_uses_firing_prompt() -> None:
    """Resolved alert without previous_analysis falls back to firing prompt."""
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    engine = CapturingAnalysisEngine("ok")
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        prometheus_enabled=False,
    )

    request = AlertAnalysisRequest(
        alert=Alert(
            status="resolved",
            labels={"namespace": "default", "pod": "demo-pod"},
            annotations={"summary": "Test"},
            fingerprint="abc123",
        ),
        thread_ts="1234567890.123456",
        analysis_type="resolved",
    )
    service.analyze(request)

    assert "RESOLVED alert" not in engine.last_prompt
    assert "Analyze the alert using the provided Kubernetes context" in engine.last_prompt


def test_resolved_reduces_context_limits() -> None:
    """Resolved analysis with previous_analysis halves log/event limits."""
    logs = [f"line-{idx}" for idx in range(20)]
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[
            PodEventSummary(
                type="Normal",
                reason="Pulled",
                message=f"event-{idx}",
                count=1,
                first_timestamp=None,
                last_timestamp=None,
                involved_object=None,
            )
            for idx in range(20)
        ],
        previous_logs=[PodLogSnippet(container="app", previous=True, logs=logs)],
        warnings=[],
    )
    engine = CapturingAnalysisEngine("ok")
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        prometheus_enabled=False,
        prompt_max_log_lines=10,
        prompt_max_events=10,
    )

    service.analyze(_sample_resolved_request_with_previous())

    # 50% 축소: max_log_lines=5, max_events=5
    assert "line-19" in engine.last_prompt
    assert "line-14" not in engine.last_prompt
    assert "event-4" in engine.last_prompt
    assert "event-5" not in engine.last_prompt


def test_resolved_tempo_window_uses_ends_at() -> None:
    """Resolved analysis uses ends_at as tempo window pivot."""
    context = K8sContext(
        namespace="bookinfo",
        pod_name="reviews-v1",
        workload="reviews",
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    tempo = FakeTempoClient()
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=FakeAnalysisEngine("ok"),
        tempo_client=tempo,
        tempo_enabled=True,
        tempo_lookback_minutes=15,
        tempo_forward_minutes=5,
    )

    ends_at = datetime(2024, 1, 1, 12, 30, tzinfo=timezone.utc)
    request = AlertAnalysisRequest(
        alert=Alert(
            status="resolved",
            labels={
                "namespace": "bookinfo",
                "destination_service_name": "reviews",
                "destination_service_namespace": "bookinfo",
            },
            annotations={"summary": "Test"},
            startsAt=datetime(2024, 1, 1, 12, 0, tzinfo=timezone.utc),
            endsAt=ends_at,
            fingerprint="trace-abc123",
        ),
        thread_ts="1234567890.123456",
        analysis_type="resolved",
        previous_analysis=PreviousAnalysisContext(
            status="firing",
            summary="prev summary",
            detail="prev detail",
        ),
    )
    service.analyze(request)

    assert tempo.calls
    call = tempo.calls[-1]
    expected_start = (ends_at - timedelta(minutes=15)).isoformat().replace("+00:00", "Z")
    expected_end = (ends_at + timedelta(minutes=5)).isoformat().replace("+00:00", "Z")
    assert call["start"] == expected_start
    assert call["end"] == expected_end


def test_summarize_incident_masks_prompt_and_response() -> None:
    secret = "secret-value"
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    engine = CapturingAnalysisEngine(
        f"제목: incident {secret}\n요약: summary {secret}\n상세 분석: detail {secret}\n"
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=engine,
        masker=RegexMasker.from_patterns([r"secret-[A-Za-z]+"]),
    )
    request = IncidentSummaryRequest(
        incident_id="INC-1",
        title="원본 타이틀",
        severity="critical",
        fired_at="2026-01-01T00:00:00Z",
        resolved_at="2026-01-01T00:10:00Z",
        alerts=[
            AlertSummaryInput(
                fingerprint="abc",
                alert_name="TestAlert",
                severity="critical",
                status="resolved",
                analysis_summary=f"analysis summary {secret}",
                analysis_detail=f"analysis detail {secret}",
            )
        ],
    )

    title, summary, detail = service.summarize_incident(request)
    rendered = "\n".join([title, summary, detail])

    assert secret not in engine.last_prompt
    assert "[MASKED]" in engine.last_prompt
    assert secret not in rendered
    assert "[MASKED]" in rendered


# ---------------------------------------------------------------------------
# _parse_incident_summary / _extract_first_paragraph unit tests
# ---------------------------------------------------------------------------


class TestExtractFirstParagraph:
    def test_returns_first_paragraph(self):
        text = (
            "### 제목\n\nThis is the first paragraph."
            "\nSecond line of same paragraph.\n\nAnother paragraph."
        )
        assert (
            _extract_first_paragraph(text)
            == "This is the first paragraph. Second line of same paragraph."
        )

    def test_skips_headers_and_labels(self):
        text = "### Some Header\n요약:\nActual content here."
        assert _extract_first_paragraph(text) == "Actual content here."

    def test_empty_text(self):
        assert _extract_first_paragraph("") == ""

    def test_only_headers(self):
        text = "### Header\n## Another\n제목:"
        # Falls back to first line of text
        assert _extract_first_paragraph(text) == "### Header"

    def test_no_blank_lines(self):
        text = "Line one.\nLine two.\nLine three."
        assert _extract_first_paragraph(text) == "Line one. Line two. Line three."


class TestParseIncidentSummary:
    def test_well_structured_response(self):
        result = (
            "### 제목 (Title)\n"
            "[svc] OOMKilled로 인한 Pod 재시작\n\n"
            "### 요약 (Summary)\n"
            "메모리 부족으로 Pod가 재시작되었습니다.\n\n"
            "### 상세 분석 (Detail)\n"
            "근본 원인 분석 내용..."
        )
        title, summary, detail = _parse_incident_summary(result, "fallback-title")
        assert title == "[svc] OOMKilled로 인한 Pod 재시작"
        assert summary == "메모리 부족으로 Pod가 재시작되었습니다."
        assert detail == "근본 원인 분석 내용..."

    def test_fallback_title_used_when_extraction_fails(self):
        result = "Some unstructured response without any section headers."
        title, summary, detail = _parse_incident_summary(result, "my-fallback")
        assert title == "my-fallback"

    def test_summary_fallback_uses_first_paragraph_not_entire_response(self):
        result = (
            "### 제목 (Title)\n"
            "[svc] 장애 제목\n\n"
            "첫 번째 단락입니다.\n\n"
            "두 번째 단락 — 상세 내용이 길게 이어집니다."
        )
        title, summary, detail = _parse_incident_summary(result, "fallback")
        assert title == "[svc] 장애 제목"
        # summary should NOT be the entire response
        assert "두 번째 단락" not in summary
        assert "첫 번째 단락" in summary

    def test_title_truncated_at_max_len(self):
        result = "### 제목 (Title)\n" + "A" * 200 + "\n\n### 요약 (Summary)\nShort summary."
        title, summary, _ = _parse_incident_summary(result, "fallback")
        assert len(title) <= 101  # 100 + ellipsis char
        assert title.endswith("…")

    def test_summary_not_truncated(self):
        result = "### 제목 (Title)\ntitle\n\n### 요약 (Summary)\n" + "B" * 500
        _, summary, _ = _parse_incident_summary(result, "fallback")
        assert len(summary) == 500

    def test_detail_fallback_to_full_result_when_no_detail_section(self):
        result = "Some unstructured response without any section headers."
        _, _, detail = _parse_incident_summary(result, "my-fallback")
        assert detail == result

    def test_bold_markers_stripped_from_title(self):
        result = (
            "### 제목 (Title)\n"
            "**[kube-rca] 복합 장애**\n\n"
            "### 요약 (Summary)\n"
            "요약 내용.\n\n"
            "### 상세 분석 (Detail)\n"
            "근본 원인..."
        )
        title, _, _ = _parse_incident_summary(result, "fallback")
        assert "**" not in title
        assert title == "[kube-rca] 복합 장애"

    def test_italic_markers_stripped_from_title(self):
        result = "### 제목 (Title)\n*긴급 알림*\n\n### 요약 (Summary)\n요약."
        title, _, _ = _parse_incident_summary(result, "fallback")
        assert title == "긴급 알림"

    def test_asterisk_in_middle_of_title_preserved(self):
        result = "### 제목 (Title)\nk8s * wildcard query\n\n### 요약 (Summary)\n요약."
        title, _, _ = _parse_incident_summary(result, "fallback")
        assert title == "k8s * wildcard query"

    def test_detail_excludes_title_and_summary(self):
        result = (
            "### 제목 (Title)\n"
            "장애 제목\n\n"
            "### 요약 (Summary)\n"
            "요약 한 줄.\n\n"
            "### 상세 분석 (Detail)\n"
            "* 근본 원인: OOM\n"
            "* 영향 범위: bookinfo\n"
            "* 해결 과정: 리소스 상향"
        )
        _, _, detail = _parse_incident_summary(result, "fallback")
        assert "장애 제목" not in detail
        assert "요약 한 줄" not in detail
        assert "근본 원인: OOM" in detail
        assert "영향 범위: bookinfo" in detail

    def test_bold_without_colon_extracts_title_and_summary(self):
        """LLM이 colon 없이 **key** value 형식으로 응답한 경우."""
        result = (
            "**제목 (Title)** [bookinfo/ratings] Istio 503 에러 유발\n\n"
            "**요약 (Summary)** ratings 서비스에 Fault Injection이 설정되었습니다.\n\n"
            "**상세 분석 (Detail)**\n"
            "근본 원인 분석 내용..."
        )
        title, summary, detail = _parse_incident_summary(result, "Ongoing")
        assert title == "[bookinfo/ratings] Istio 503 에러 유발"
        assert summary == "ratings 서비스에 Fault Injection이 설정되었습니다."
        assert title != "Ongoing"

    def test_bold_with_colon_still_works(self):
        """기존 **key**: value 형식이 여전히 동작하는지 검증."""
        result = (
            "**제목 (Title)**: [svc] OOMKilled 장애\n\n"
            "**요약 (Summary)**: 메모리 부족으로 재시작.\n\n"
            "**상세 분석 (Detail)**:\n"
            "근본 원인..."
        )
        title, summary, _ = _parse_incident_summary(result, "Ongoing")
        assert title == "[svc] OOMKilled 장애"
        assert summary == "메모리 부족으로 재시작."

    def test_numbered_bold_without_colon(self):
        """1. **key** value 형식 (번호 포함, colon 없음)."""
        result = (
            "1. **제목 (Title)** [ns/pod] 장애\n\n"
            "2. **요약 (Summary)** 요약 내용.\n\n"
            "3. **상세 분석 (Detail)**\n"
            "상세..."
        )
        title, summary, _ = _parse_incident_summary(result, "Ongoing")
        assert title == "[ns/pod] 장애"
        assert summary == "요약 내용."


# ── resolve_alert_target: expanded label keys ──


def test_resolve_target_pod_name_key() -> None:
    labels = {"namespace": "default", "pod_name": "demo-pod"}
    target = resolve_alert_target(labels)
    assert target.pod_name == "demo-pod"


def test_resolve_target_exported_pod_key() -> None:
    labels = {"namespace": "default", "exported_pod": "export-pod"}
    target = resolve_alert_target(labels)
    assert target.pod_name == "export-pod"


def test_resolve_target_pod_priority_over_pod_name() -> None:
    labels = {"namespace": "default", "pod": "primary", "pod_name": "secondary"}
    target = resolve_alert_target(labels)
    assert target.pod_name == "primary"


def test_resolve_target_app_as_workload() -> None:
    labels = {"namespace": "default", "app": "my-app"}
    target = resolve_alert_target(labels)
    assert target.workload == "my-app"


def test_resolve_target_k8s_app_as_workload() -> None:
    labels = {"namespace": "default", "k8s_app": "my-svc"}
    target = resolve_alert_target(labels)
    assert target.workload == "my-svc"


def test_resolve_target_workload_priority_over_app() -> None:
    labels = {"namespace": "default", "workload": "deploy-a", "app": "app-b"}
    target = resolve_alert_target(labels)
    assert target.workload == "deploy-a"


# ── resolve_alert_target: sentinel value filtering ──


def test_resolve_target_sentinel_unknown_filtered() -> None:
    labels = {"namespace": "bookinfo", "destination_workload": "unknown"}
    target = resolve_alert_target(labels)
    assert target.workload is None


def test_resolve_target_sentinel_none_filtered() -> None:
    labels = {"namespace": "bookinfo", "pod": "none"}
    target = resolve_alert_target(labels)
    assert target.pod_name is None


def test_resolve_target_sentinel_case_insensitive() -> None:
    labels = {"namespace": "bookinfo", "destination_workload": "Unknown"}
    target = resolve_alert_target(labels)
    assert target.workload is None


def test_resolve_target_sentinel_skips_to_next_key() -> None:
    labels = {
        "namespace": "bookinfo",
        "destination_workload": "unknown",
        "app": "reviews",
    }
    target = resolve_alert_target(labels)
    assert target.workload == "reviews"


# ── resolve_alert_target: node label extraction ──


def test_resolve_target_node_label() -> None:
    labels = {"alertname": "NodeFilesystemSpaceFillingUp", "node": "ip-10-0-1-179.ec2.internal"}
    target = resolve_alert_target(labels)
    assert target.node_name == "ip-10-0-1-179.ec2.internal"


def test_resolve_target_nodename_label() -> None:
    labels = {"alertname": "NodeFilesystemSpaceFillingUp", "nodename": "worker-01"}
    target = resolve_alert_target(labels)
    assert target.node_name == "worker-01"


def test_resolve_target_node_priority_over_nodename() -> None:
    labels = {"node": "primary-node", "nodename": "fallback-node"}
    target = resolve_alert_target(labels)
    assert target.node_name == "primary-node"


def test_resolve_target_node_sentinel_filtered() -> None:
    labels = {"node": "unknown"}
    target = resolve_alert_target(labels)
    assert target.node_name is None


def test_resolve_target_node_with_namespace() -> None:
    labels = {
        "namespace": "monitoring",
        "node": "worker-02",
        "alertname": "NodeMemoryHighUtilization",
    }
    target = resolve_alert_target(labels)
    assert target.node_name == "worker-02"
    assert target.namespace == "monitoring"


# ── _resolve_analysis_quality: node alerts ──


def _make_node_k8s_context(
    *,
    node_name: str | None = "worker-01",
    node_status: dict[str, object] | None = None,
    namespace: str | None = None,
) -> K8sContext:
    return K8sContext(
        namespace=namespace,
        pod_name=None,
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
        target=AnalysisTarget(
            namespace=namespace,
            pod_name=None,
            workload=None,
            service_name=None,
            node_name=node_name,
        ),
        node_status=node_status,
    )


def test_quality_node_alert_with_status_is_high() -> None:
    ctx = _make_node_k8s_context(node_status={"name": "worker-01", "conditions": []})
    quality = _resolve_analysis_quality(
        k8s_context=ctx,
        missing_data=[],
        capabilities={"k8s_core": "ok"},
        engine_issue=None,
    )
    assert quality == "high"


def test_quality_node_alert_without_status_is_medium() -> None:
    ctx = _make_node_k8s_context(node_status=None)
    quality = _resolve_analysis_quality(
        k8s_context=ctx,
        missing_data=[],
        capabilities={"k8s_core": "ok"},
        engine_issue=None,
    )
    assert quality == "medium"


def test_quality_node_alert_no_namespace_not_penalized() -> None:
    ctx = _make_node_k8s_context(
        node_status={"name": "worker-01", "conditions": []},
        namespace=None,
    )
    quality = _resolve_analysis_quality(
        k8s_context=ctx,
        missing_data=[],
        capabilities={"k8s_core": "ok"},
        engine_issue=None,
    )
    assert quality == "high"


# ── _collect_missing_data: node alerts ──


def test_missing_data_node_alert_no_pod_not_reported() -> None:
    ctx = _make_node_k8s_context(node_status={"name": "worker-01", "conditions": []})
    missing = _collect_missing_data(
        k8s_context=ctx,
        tempo_context=None,
        capabilities={"k8s_core": "ok", "manifest_read": "ok", "prometheus": "ok", "loki": "ok"},
    )
    assert "alert.labels.namespace" not in missing
    assert "alert.labels.pod" not in missing
    assert "k8s.pod_status" not in missing


def test_missing_data_node_alert_node_status_missing() -> None:
    ctx = _make_node_k8s_context(node_status=None)
    missing = _collect_missing_data(
        k8s_context=ctx,
        tempo_context=None,
        capabilities={"k8s_core": "ok", "manifest_read": "ok", "prometheus": "ok", "loki": "ok"},
    )
    assert "k8s.node_status" in missing
    assert "alert.labels.pod" not in missing


# ── _categorize_analysis_error ──


def test_categorize_error_auth_401() -> None:
    exc = Exception("HTTP 401 Unauthorized")
    cat = _categorize_analysis_error(exc)
    assert cat.name == "llm_auth"


def test_categorize_error_rate_limit() -> None:
    exc = Exception("HTTP 429 rate limit exceeded")
    cat = _categorize_analysis_error(exc)
    assert cat.name == "llm_rate_limit"


def test_categorize_error_timeout() -> None:
    exc = Exception("request timed out after 180s")
    cat = _categorize_analysis_error(exc)
    assert cat.name == "llm_timeout"


def test_categorize_error_session_db() -> None:
    exc = Exception("psycopg2.OperationalError: connection refused")
    cat = _categorize_analysis_error(exc)
    assert cat.name == "session_db"


def test_categorize_error_bad_request() -> None:
    exc = Exception("HTTP 400 bad request")
    cat = _categorize_analysis_error(exc)
    assert cat.name == "llm_bad_request"


def test_categorize_error_empty_str() -> None:
    exc = RuntimeError()
    cat = _categorize_analysis_error(exc)
    assert cat.name == "unknown"
    assert "RuntimeError" in cat.user_message


def test_categorize_error_chained_cause() -> None:
    cause = ConnectionError("connection refused to session db")
    outer = RuntimeError()
    outer.__cause__ = cause
    cat = _categorize_analysis_error(outer)
    assert cat.name == "session_db"


# ── fallback_summary: events and annotations ──


def test_fallback_with_events() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[
            PodEventSummary(
                type="Warning",
                reason="BackOff",
                message="Back-off restarting failed container",
                count=3,
                first_timestamp=None,
                last_timestamp=None,
                involved_object=None,
            )
        ],
        previous_logs=[],
        warnings=[],
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=None,
    )
    analysis, _, _, _, _ = service.analyze(_sample_request())
    assert "recent_events (1)" in analysis
    assert "[Warning] BackOff" in analysis


def test_fallback_with_alert_annotation() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=None,
    )
    request = AlertAnalysisRequest(
        alert=Alert(
            status="firing",
            labels={"namespace": "default", "pod": "demo-pod"},
            annotations={"summary": "Pod is crash looping"},
            fingerprint="ann-test",
        ),
        thread_ts="1234567890.123456",
    )
    analysis, _, _, _, _ = service.analyze(request)
    assert "alert_summary: Pod is crash looping" in analysis


# ── _resolve_analysis_quality: workload medium path ──


def test_quality_medium_with_workload_no_pod() -> None:
    """k8s_app resolves workload only (not in service_keys) -> medium quality."""
    context = K8sContext(
        namespace="bookinfo",
        pod_name=None,
        workload="reviews",
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=FakeAnalysisEngine("## 요약\ntest\n## 상세 분석\ndetail"),
    )
    request = AlertAnalysisRequest(
        alert=Alert(
            status="firing",
            labels={"namespace": "bookinfo", "k8s_app": "reviews"},
            annotations={},
            fingerprint="quality-test",
        ),
        thread_ts="1234567890.123456",
    )
    _, _, _, ctx, _ = service.analyze(request)
    assert ctx.get("analysis_quality") == "medium"


def test_quality_low_without_workload_or_pod() -> None:
    context = K8sContext(
        namespace="bookinfo",
        pod_name=None,
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=FakeAnalysisEngine("## 요약\ntest\n## 상세 분석\ndetail"),
    )
    request = AlertAnalysisRequest(
        alert=Alert(
            status="firing",
            labels={"namespace": "bookinfo"},
            annotations={},
            fingerprint="quality-low-test",
        ),
        thread_ts="1234567890.123456",
    )
    _, _, _, ctx, _ = service.analyze(request)
    assert ctx.get("analysis_quality") == "low"


# ── error categorization in service.analyze() integration ──


class FailingAnalysisEngine:
    def analyze(self, prompt: str, incident_id: str | None = None) -> str:
        raise RuntimeError()


def test_analysis_service_error_categorization_integration() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        workload=None,
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=FailingAnalysisEngine(),
    )
    analysis, _, _, ctx, _ = service.analyze(_sample_request())
    assert "analysis engine unavailable" in analysis
    assert "RuntimeError" in analysis
    assert ctx.get("analysis_quality") == "low"
