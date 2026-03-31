from __future__ import annotations

from dataclasses import asdict, dataclass, field


@dataclass(frozen=True)
class PodStatusSnapshot:
    phase: str
    node_name: str | None
    start_time: str | None
    reason: str | None
    message: str | None
    conditions: list[dict[str, str | None]]
    container_statuses: list[dict[str, str | int | bool | None]]

    def to_dict(self) -> dict[str, object]:
        return asdict(self)


@dataclass(frozen=True)
class PodEventSummary:
    type: str | None
    reason: str | None
    message: str | None
    count: int | None
    first_timestamp: str | None
    last_timestamp: str | None
    involved_object: dict[str, str | None] | None

    def to_dict(self) -> dict[str, object]:
        return asdict(self)


@dataclass(frozen=True)
class PodLogSnippet:
    container: str
    previous: bool
    logs: list[str]
    error: str | None = None

    def to_dict(self) -> dict[str, object]:
        return asdict(self)


@dataclass(frozen=True)
class PodSummary:
    """Lightweight pod summary for listing pods in a namespace."""

    name: str
    namespace: str
    phase: str
    node_name: str | None
    labels: dict[str, str]
    restart_count: int
    ready: bool
    start_time: str | None

    def to_dict(self) -> dict[str, object]:
        return asdict(self)


@dataclass(frozen=True)
class AnalysisTarget:
    namespace: str | None
    pod_name: str | None
    workload: str | None
    service_name: str | None

    def to_dict(self) -> dict[str, object]:
        return asdict(self)


@dataclass(frozen=True)
class K8sContext:
    namespace: str | None
    pod_name: str | None
    workload: str | None
    pod_status: PodStatusSnapshot | None
    events: list[PodEventSummary]
    previous_logs: list[PodLogSnippet]
    warnings: list[str]
    target: AnalysisTarget | None = None
    current_logs: list[PodLogSnippet] = field(default_factory=list)
    pod_spec: dict[str, object] | None = None
    workload_status: dict[str, object] | None = None
    pod_metrics: dict[str, object] | None = None
    node_status: dict[str, object] | None = None
    service_manifest: dict[str, object] | None = None
    endpoints_manifest: dict[str, object] | None = None

    def to_dict(self) -> dict[str, object]:
        return {
            "namespace": self.namespace,
            "pod_name": self.pod_name,
            "workload": self.workload,
            "service_name": self.target.service_name if self.target else None,
            "target": None if self.target is None else self.target.to_dict(),
            "pod_status": None if self.pod_status is None else self.pod_status.to_dict(),
            "events": [event.to_dict() for event in self.events],
            "current_logs": [snippet.to_dict() for snippet in self.current_logs],
            "previous_logs": [snippet.to_dict() for snippet in self.previous_logs],
            "pod_spec": self.pod_spec,
            "workload_status": self.workload_status,
            "pod_metrics": self.pod_metrics,
            "node_status": self.node_status,
            "service_manifest": self.service_manifest,
            "endpoints_manifest": self.endpoints_manifest,
            "warnings": self.warnings,
        }
