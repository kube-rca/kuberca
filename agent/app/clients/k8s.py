from __future__ import annotations

import logging
from collections.abc import Iterable

from kubernetes import client, config
from kubernetes.config.config_exception import ConfigException

from app.models.k8s import K8sContext, PodEventSummary, PodLogSnippet, PodStatusSnapshot


class KubernetesClient:
    def __init__(self, timeout_seconds: int, event_limit: int, log_tail_lines: int) -> None:
        self._logger = logging.getLogger(__name__)
        self._timeout_seconds = timeout_seconds
        self._event_limit = event_limit
        self._log_tail_lines = log_tail_lines
        self._core_api = self._build_client()

    def collect_context(self, namespace: str | None, pod_name: str | None) -> K8sContext:
        warnings: list[str] = []
        if not namespace or not pod_name:
            warnings.append("namespace/pod_name missing from alert labels")
            return K8sContext(
                namespace=namespace,
                pod_name=pod_name,
                pod_status=None,
                events=[],
                previous_logs=[],
                warnings=warnings,
            )

        if self._core_api is None:
            warnings.append("kubernetes client is not configured")
            return K8sContext(
                namespace=namespace,
                pod_name=pod_name,
                pod_status=None,
                events=[],
                previous_logs=[],
                warnings=warnings,
            )

        pod = self._read_pod(namespace, pod_name, warnings)
        pod_status = self._extract_pod_status(pod) if pod else None
        events = self._list_pod_events(namespace, pod_name, warnings)
        previous_logs = self._get_previous_logs(namespace, pod, warnings)

        return K8sContext(
            namespace=namespace,
            pod_name=pod_name,
            pod_status=pod_status,
            events=events,
            previous_logs=previous_logs,
            warnings=warnings,
        )

    def get_pod_status(self, namespace: str, pod_name: str) -> PodStatusSnapshot | None:
        if self._core_api is None:
            return None
        pod = self._read_pod(namespace, pod_name, [])
        if pod is None:
            return None
        return self._extract_pod_status(pod)

    def list_pod_events(self, namespace: str, pod_name: str) -> list[PodEventSummary]:
        if self._core_api is None:
            return []
        return self._list_pod_events(namespace, pod_name, [])

    def get_previous_logs(
        self, namespace: str, pod_name: str
    ) -> list[PodLogSnippet]:
        if self._core_api is None:
            return []
        pod = self._read_pod(namespace, pod_name, [])
        return self._get_previous_logs(namespace, pod, [])

    def _build_client(self) -> client.CoreV1Api | None:
        try:
            config.load_incluster_config()
            self._logger.info("Loaded in-cluster Kubernetes config")
        except ConfigException:
            try:
                config.load_kube_config()
                self._logger.info("Loaded kubeconfig for local development")
            except ConfigException as exc:
                self._logger.warning("Failed to configure Kubernetes client: %s", exc)
                return None
        return client.CoreV1Api()

    def _read_pod(
        self,
        namespace: str,
        pod_name: str,
        warnings: list[str],
    ) -> client.V1Pod | None:
        try:
            return self._core_api.read_namespaced_pod(
                name=pod_name,
                namespace=namespace,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001 - Kubernetes client raises many exception types.
            self._logger.warning("Failed to read pod %s/%s: %s", namespace, pod_name, exc)
            warnings.append("failed to read pod status")
            return None

    def _list_pod_events(
        self,
        namespace: str,
        pod_name: str,
        warnings: list[str],
    ) -> list[PodEventSummary]:
        field_selector = f"involvedObject.kind=Pod,involvedObject.name={pod_name}"
        try:
            response = self._core_api.list_namespaced_event(
                namespace=namespace,
                field_selector=field_selector,
                limit=self._event_limit,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to list events for %s/%s: %s", namespace, pod_name, exc)
            warnings.append("failed to list events")
            return []

        return [self._to_event_summary(item) for item in response.items]

    def _get_previous_logs(
        self,
        namespace: str,
        pod: client.V1Pod | None,
        warnings: list[str],
    ) -> list[PodLogSnippet]:
        if pod is None:
            return []

        container_names = [container.name for container in pod.spec.containers or []]
        snippets: list[PodLogSnippet] = []
        for container_name in container_names:
            try:
                logs = self._core_api.read_namespaced_pod_log(
                    name=pod.metadata.name,
                    namespace=namespace,
                    container=container_name,
                    previous=True,
                    tail_lines=self._log_tail_lines,
                    timestamps=True,
                    _request_timeout=self._timeout_seconds,
                )
                snippets.append(
                    PodLogSnippet(
                        container=container_name,
                        previous=True,
                        logs=logs.splitlines() if logs else [],
                    )
                )
            except Exception as exc:  # noqa: BLE001
                self._logger.warning(
                    "Failed to read previous logs for %s/%s (%s): %s",
                    namespace,
                    pod.metadata.name,
                    container_name,
                    exc,
                )
                warnings.append(f"failed to read previous logs for container {container_name}")
                snippets.append(
                    PodLogSnippet(
                        container=container_name,
                        previous=True,
                        logs=[],
                        error="failed to read previous logs",
                    )
                )
        return snippets

    def _extract_pod_status(self, pod: client.V1Pod) -> PodStatusSnapshot:
        status = pod.status
        conditions = [
            {
                "type": condition.type,
                "status": condition.status,
                "reason": condition.reason,
                "message": condition.message,
                "last_transition_time": self._to_iso(condition.last_transition_time),
            }
            for condition in status.conditions or []
        ]
        container_statuses = [
            {
                "name": container_status.name,
                "ready": container_status.ready,
                "restart_count": container_status.restart_count,
                "state": self._extract_container_state(container_status.state),
                "last_state": self._extract_container_state(container_status.last_state),
            }
            for container_status in status.container_statuses or []
        ]

        return PodStatusSnapshot(
            phase=status.phase or "",
            node_name=pod.spec.node_name,
            start_time=self._to_iso(status.start_time),
            reason=status.reason,
            message=status.message,
            conditions=conditions,
            container_statuses=container_statuses,
        )

    def _extract_container_state(
        self,
        state: client.V1ContainerState | None,
    ) -> dict[str, str] | None:
        if state is None:
            return None
        if state.waiting:
            return {
                "type": "waiting",
                "reason": state.waiting.reason,
                "message": state.waiting.message,
            }
        if state.terminated:
            return {
                "type": "terminated",
                "reason": state.terminated.reason,
                "message": state.terminated.message,
                "exit_code": str(state.terminated.exit_code),
            }
        if state.running:
            return {
                "type": "running",
                "started_at": self._to_iso(state.running.started_at),
            }
        return None

    def _to_event_summary(self, event: client.V1Event) -> PodEventSummary:
        return PodEventSummary(
            type=event.type,
            reason=event.reason,
            message=event.message,
            count=event.count,
            first_timestamp=self._to_iso(event.first_timestamp),
            last_timestamp=self._to_iso(event.last_timestamp or event.event_time),
        )

    @staticmethod
    def _to_iso(value: object | None) -> str | None:
        if value is None:
            return None
        if hasattr(value, "isoformat"):
            return value.isoformat()
        return str(value)


def extract_pod_target(labels: dict[str, str]) -> tuple[str | None, str | None]:
    namespace_keys = ["namespace"]
    pod_keys = ["pod"]

    namespace = _first_label_value(labels, namespace_keys)
    pod_name = _first_label_value(labels, pod_keys)

    return namespace, pod_name


def _first_label_value(labels: dict[str, str], keys: Iterable[str]) -> str | None:
    for key in keys:
        value = labels.get(key)
        if value:
            return value
    return None
