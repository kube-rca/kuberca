from __future__ import annotations

from app.models.k8s import K8sContext, PodStatusSnapshot
from app.schemas.alert import Alert
from app.schemas.analysis import AlertAnalysisRequest
from app.services.analysis import AnalysisService


class FakeKubernetesClient:
    def __init__(self, context: K8sContext) -> None:
        self._context = context

    def collect_context(self, namespace: str | None, pod_name: str | None) -> K8sContext:
        return self._context


class FakeAnalysisEngine:
    def __init__(self, result: str) -> None:
        self._result = result

    def analyze(self, prompt: str, incident_id: str | None = None) -> str:
        return self._result


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


def test_analysis_service_fallback() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
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
    service = AnalysisService(FakeKubernetesClient(context), analysis_engine=None)

    analysis, summary, detail, context, artifacts = service.analyze(_sample_request())

    assert "analysis engine unavailable" in analysis
    assert summary
    assert detail
    assert context
    assert isinstance(artifacts, list)


def test_analysis_service_uses_engine() -> None:
    context = K8sContext(
        namespace="default",
        pod_name="demo-pod",
        pod_status=None,
        events=[],
        previous_logs=[],
        warnings=[],
    )
    service = AnalysisService(
        FakeKubernetesClient(context),
        analysis_engine=FakeAnalysisEngine("ok"),
    )

    analysis, summary, detail, context, artifacts = service.analyze(_sample_request())

    assert analysis == "ok"
    assert summary
    assert detail
    assert context
    assert isinstance(artifacts, list)
