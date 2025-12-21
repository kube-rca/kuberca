from __future__ import annotations

from dataclasses import asdict, dataclass


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
class K8sContext:
    namespace: str | None
    pod_name: str | None
    pod_status: PodStatusSnapshot | None
    events: list[PodEventSummary]
    previous_logs: list[PodLogSnippet]
    warnings: list[str]

    def to_dict(self) -> dict[str, object]:
        return {
            "namespace": self.namespace,
            "pod_name": self.pod_name,
            "pod_status": None if self.pod_status is None else self.pod_status.to_dict(),
            "events": [event.to_dict() for event in self.events],
            "previous_logs": [snippet.to_dict() for snippet in self.previous_logs],
            "warnings": self.warnings,
        }
