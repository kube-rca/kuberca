from __future__ import annotations

import json

from app.core.masking import RegexMasker
from app.models.k8s import K8sContext, PodLogSnippet, PodStatusSnapshot
from app.schemas.alert import Alert
from app.schemas.analysis import AlertAnalysisRequest, AlertSummaryInput, IncidentSummaryRequest
from app.services.analysis import AnalysisService


class FakeKubernetesClient:
    def __init__(self, context: K8sContext) -> None:
        self._context = context

    def collect_context(
        self, namespace: str | None, pod_name: str | None, workload: str | None = None
    ) -> K8sContext:
        return self._context


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
        f"제목: incident {secret}\n"
        f"요약: summary {secret}\n"
        f"상세 분석: detail {secret}\n"
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
