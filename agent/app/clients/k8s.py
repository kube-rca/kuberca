from __future__ import annotations

import logging
from collections.abc import Iterable

from kubernetes import client, config
from kubernetes.config.config_exception import ConfigException

from app.models.k8s import K8sContext, PodEventSummary, PodLogSnippet, PodStatusSnapshot, PodSummary


class KubernetesClient:
    def __init__(self, timeout_seconds: int, event_limit: int, log_tail_lines: int) -> None:
        self._logger = logging.getLogger(__name__)
        self._timeout_seconds = timeout_seconds
        self._event_limit = event_limit
        self._log_tail_lines = log_tail_lines
        self._core_api = self._build_client()
        self._apps_api = client.AppsV1Api() if self._core_api else None
        self._batch_api = client.BatchV1Api() if self._core_api else None
        self._custom_api = client.CustomObjectsApi() if self._core_api else None
        self._events_api = client.EventsV1Api() if self._core_api else None

    def collect_context(
        self, namespace: str | None, pod_name: str | None, workload: str | None = None
    ) -> K8sContext:
        warnings: list[str] = []
        if not namespace or not pod_name:
            hint = ""
            if namespace and workload:
                hint = (
                    f" Use list_pods_in_namespace('{namespace}', 'app={workload}') "
                    "to discover pods."
                )
            warnings.append(f"namespace/pod_name missing from alert labels.{hint}")
            return K8sContext(
                namespace=namespace,
                pod_name=pod_name,
                workload=workload,
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
                workload=workload,
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
            workload=workload,
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

    def get_previous_logs(self, namespace: str, pod_name: str) -> list[PodLogSnippet]:
        if self._core_api is None:
            return []
        pod = self._read_pod(namespace, pod_name, [])
        return self._get_previous_logs(namespace, pod, [])

    def get_pod_spec_summary(self, namespace: str, pod_name: str) -> dict[str, object] | None:
        if self._core_api is None:
            return None
        pod = self._read_pod(namespace, pod_name, [])
        if pod is None:
            return None
        return self._summarize_pod_spec(pod)

    def list_namespace_events(self, namespace: str) -> list[PodEventSummary]:
        if self._core_api is None and self._events_api is None:
            return []
        v1_events, v1_error = self._list_namespace_events_v1(namespace)
        if v1_events:
            return v1_events
        core_events, core_error = self._list_namespace_events_core(namespace)
        if core_events:
            return core_events
        return self._event_warnings([v1_error, core_error])

    def list_cluster_events(self) -> list[PodEventSummary]:
        if self._core_api is None and self._events_api is None:
            return []
        v1_events, v1_error = self._list_cluster_events_v1()
        if v1_events:
            return v1_events
        core_events, core_error = self._list_cluster_events_core()
        if core_events:
            return core_events
        return self._event_warnings([v1_error, core_error])

    def get_pod_logs(
        self,
        namespace: str,
        pod_name: str,
        *,
        container: str | None = None,
        tail_lines: int | None = None,
        since_seconds: int | None = None,
    ) -> list[PodLogSnippet]:
        if self._core_api is None:
            return []
        pod = self._read_pod(namespace, pod_name, [])
        return self._get_current_logs(
            namespace,
            pod,
            container=container,
            tail_lines=tail_lines,
            since_seconds=since_seconds,
        )

    def get_workload_summary(self, namespace: str, pod_name: str) -> dict[str, object] | None:
        if self._apps_api is None and self._batch_api is None:
            return None
        pod = self._read_pod(namespace, pod_name, [])
        if pod is None or pod.metadata is None:
            return None

        owner_ref = self._select_owner_reference(pod.metadata.owner_references or [])
        if owner_ref is None:
            return None

        owner_ref = self._resolve_owner_reference(namespace, owner_ref)
        if owner_ref.kind == "Deployment" and self._apps_api:
            try:
                deployment = self._apps_api.read_namespaced_deployment(
                    name=owner_ref.name,
                    namespace=namespace,
                    _request_timeout=self._timeout_seconds,
                )
            except Exception as exc:  # noqa: BLE001
                self._logger.warning(
                    "Failed to read Deployment %s/%s: %s", namespace, owner_ref.name, exc
                )
                return None
            return self._summarize_deployment(deployment)
        if owner_ref.kind == "ReplicaSet" and self._apps_api:
            try:
                replica_set = self._apps_api.read_namespaced_replica_set(
                    name=owner_ref.name,
                    namespace=namespace,
                    _request_timeout=self._timeout_seconds,
                )
            except Exception as exc:  # noqa: BLE001
                self._logger.warning(
                    "Failed to read ReplicaSet %s/%s: %s", namespace, owner_ref.name, exc
                )
                return None
            return self._summarize_replica_set(replica_set)
        if owner_ref.kind == "StatefulSet" and self._apps_api:
            try:
                stateful_set = self._apps_api.read_namespaced_stateful_set(
                    name=owner_ref.name,
                    namespace=namespace,
                    _request_timeout=self._timeout_seconds,
                )
            except Exception as exc:  # noqa: BLE001
                self._logger.warning(
                    "Failed to read StatefulSet %s/%s: %s", namespace, owner_ref.name, exc
                )
                return None
            return self._summarize_stateful_set(stateful_set)
        if owner_ref.kind == "DaemonSet" and self._apps_api:
            try:
                daemon_set = self._apps_api.read_namespaced_daemon_set(
                    name=owner_ref.name,
                    namespace=namespace,
                    _request_timeout=self._timeout_seconds,
                )
            except Exception as exc:  # noqa: BLE001
                self._logger.warning(
                    "Failed to read DaemonSet %s/%s: %s", namespace, owner_ref.name, exc
                )
                return None
            return self._summarize_daemon_set(daemon_set)
        if owner_ref.kind == "Job" and self._batch_api:
            try:
                job = self._batch_api.read_namespaced_job(
                    name=owner_ref.name,
                    namespace=namespace,
                    _request_timeout=self._timeout_seconds,
                )
            except Exception as exc:  # noqa: BLE001
                self._logger.warning("Failed to read Job %s/%s: %s", namespace, owner_ref.name, exc)
                return None
            return self._summarize_job(job)
        if owner_ref.kind == "CronJob" and self._batch_api:
            try:
                cron_job = self._batch_api.read_namespaced_cron_job(
                    name=owner_ref.name,
                    namespace=namespace,
                    _request_timeout=self._timeout_seconds,
                )
            except Exception as exc:  # noqa: BLE001
                self._logger.warning(
                    "Failed to read CronJob %s/%s: %s", namespace, owner_ref.name, exc
                )
                return None
            return self._summarize_cron_job(cron_job)

        return None

    def get_daemon_set_manifest(self, namespace: str, name: str) -> dict[str, object] | None:
        if self._apps_api is None:
            return None
        try:
            daemon_set = self._apps_api.read_namespaced_daemon_set(
                name=name,
                namespace=namespace,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to read DaemonSet %s/%s: %s", namespace, name, exc)
            return None
        payload = daemon_set.to_dict()
        return {
            "apiVersion": payload.get("api_version"),
            "kind": payload.get("kind") or "DaemonSet",
            "metadata": payload.get("metadata"),
            "spec": payload.get("spec"),
        }

    def get_node_status(self, node_name: str) -> dict[str, object] | None:
        if self._core_api is None:
            return None
        try:
            node = self._core_api.read_node(
                name=node_name,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to read node %s: %s", node_name, exc)
            return None

        conditions = [
            {
                "type": condition.type,
                "status": condition.status,
                "reason": condition.reason,
                "message": condition.message,
                "last_transition_time": self._to_iso(condition.last_transition_time),
            }
            for condition in node.status.conditions or []
        ]

        return {
            "name": node.metadata.name if node.metadata else node_name,
            "unschedulable": node.spec.unschedulable if node.spec else None,
            "labels": node.metadata.labels if node.metadata else {},
            "taints": [taint.to_dict() for taint in node.spec.taints or []] if node.spec else [],
            "capacity": node.status.capacity if node.status else None,
            "allocatable": node.status.allocatable if node.status else None,
            "node_info": (
                node.status.node_info.to_dict() if node.status and node.status.node_info else None
            ),
            "conditions": conditions,
        }

    def get_pod_metrics(self, namespace: str, pod_name: str) -> dict[str, object] | None:
        if self._custom_api is None:
            return None
        try:
            response = self._custom_api.get_namespaced_custom_object(
                group="metrics.k8s.io",
                version="v1beta1",
                namespace=namespace,
                plural="pods",
                name=pod_name,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning(
                "Failed to read pod metrics for %s/%s: %s",
                namespace,
                pod_name,
                exc,
            )
            return None
        return self._summarize_pod_metrics(response)

    def get_node_metrics(self, node_name: str) -> dict[str, object] | None:
        if self._custom_api is None:
            return None
        try:
            response = self._custom_api.get_cluster_custom_object(
                group="metrics.k8s.io",
                version="v1beta1",
                plural="nodes",
                name=node_name,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to read node metrics for %s: %s", node_name, exc)
            return None
        return self._summarize_node_metrics(response)

    def get_manifest(
        self,
        namespace: str,
        api_version: str,
        resource: str,
        name: str,
    ) -> dict[str, object] | None:
        """Fetch a namespaced manifest for either core (v1) or custom resources."""
        parsed = self._parse_api_version(api_version)
        normalized_resource = resource.strip().lower()
        if parsed is None or not normalized_resource:
            return None
        group, version = parsed
        if group is None:
            return self._get_core_manifest(namespace, version, normalized_resource, name)
        return self._get_custom_manifest(
            group=group,
            version=version,
            namespace=namespace,
            plural=normalized_resource,
            name=name,
        )

    def list_manifests(
        self,
        namespace: str,
        api_version: str,
        resource: str,
        label_selector: str | None = None,
        limit: int = 20,
    ) -> list[dict[str, object]]:
        """List namespaced manifests for either core (v1) or custom resources."""
        parsed = self._parse_api_version(api_version)
        normalized_resource = resource.strip().lower()
        if parsed is None or not normalized_resource:
            return []
        group, version = parsed
        if group is None:
            return self._list_core_manifests(
                namespace,
                version,
                normalized_resource,
                label_selector=label_selector,
                limit=limit,
            )
        return self._list_custom_manifests(
            group=group,
            version=version,
            namespace=namespace,
            plural=normalized_resource,
            label_selector=label_selector,
            limit=limit,
        )

    def list_services_by_label(
        self, label_selector: str, namespaces: list[str] | None = None
    ) -> list[client.V1Service]:
        if self._core_api is None or not label_selector:
            return []
        try:
            if namespaces:
                services: list[client.V1Service] = []
                for namespace in namespaces:
                    response = self._core_api.list_namespaced_service(
                        namespace=namespace,
                        label_selector=label_selector,
                        _request_timeout=self._timeout_seconds,
                    )
                    services.extend(response.items)
                return services
            response = self._core_api.list_service_for_all_namespaces(
                label_selector=label_selector,
                _request_timeout=self._timeout_seconds,
            )
            return response.items
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to list services for selector %s: %s", label_selector, exc)
            return []

    def list_pods_in_namespace(
        self, namespace: str, label_selector: str | None = None
    ) -> list[PodSummary]:
        """List pods in a namespace with optional label filtering."""
        if self._core_api is None:
            return []
        try:
            if label_selector:
                response = self._core_api.list_namespaced_pod(
                    namespace=namespace,
                    label_selector=label_selector,
                    _request_timeout=self._timeout_seconds,
                )
            else:
                response = self._core_api.list_namespaced_pod(
                    namespace=namespace,
                    _request_timeout=self._timeout_seconds,
                )
            return [self._to_pod_summary(pod) for pod in response.items]
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to list pods in namespace %s: %s", namespace, exc)
            return []

    def _to_pod_summary(self, pod: client.V1Pod) -> PodSummary:
        """Convert V1Pod to PodSummary."""
        metadata = pod.metadata
        status = pod.status
        spec = pod.spec

        restart_count = 0
        ready = False
        if status and status.container_statuses:
            restart_count = sum(cs.restart_count or 0 for cs in status.container_statuses)
            ready = all(cs.ready for cs in status.container_statuses)

        return PodSummary(
            name=metadata.name if metadata else "",
            namespace=metadata.namespace if metadata else "",
            phase=status.phase if status else "Unknown",
            node_name=spec.node_name if spec else None,
            labels=metadata.labels or {} if metadata else {},
            restart_count=restart_count,
            ready=ready,
            start_time=self._to_iso(status.start_time) if status else None,
        )

    def _list_custom_manifests(
        self,
        *,
        group: str,
        version: str,
        namespace: str,
        plural: str,
        label_selector: str | None = None,
        limit: int = 20,
    ) -> list[dict[str, object]]:
        if self._custom_api is None:
            return []
        try:
            kwargs: dict[str, object] = {
                "group": group,
                "version": version,
                "namespace": namespace,
                "plural": plural,
                "_request_timeout": self._timeout_seconds,
            }
            if label_selector:
                kwargs["label_selector"] = label_selector
            if limit > 0:
                kwargs["limit"] = limit
            response = self._custom_api.list_namespaced_custom_object(**kwargs)
        except Exception as exc:  # noqa: BLE001
            self._logger.warning(
                "Failed to list custom resources %s/%s in namespace %s: %s",
                group,
                plural,
                namespace,
                exc,
            )
            return []

        if not isinstance(response, dict):
            return []
        items = response.get("items")
        if not isinstance(items, list):
            return []
        manifests: list[dict[str, object]] = []
        for item in items:
            if isinstance(item, dict):
                manifests.append(self._to_manifest(item, resource=plural))
        return manifests

    def _get_custom_manifest(
        self,
        *,
        group: str,
        version: str,
        namespace: str,
        plural: str,
        name: str,
    ) -> dict[str, object] | None:
        if self._custom_api is None:
            return None
        try:
            response = self._custom_api.get_namespaced_custom_object(
                group=group,
                version=version,
                namespace=namespace,
                plural=plural,
                name=name,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning(
                "Failed to get custom resource %s/%s %s/%s: %s",
                group,
                plural,
                namespace,
                name,
                exc,
            )
            return None
        if not isinstance(response, dict):
            return None
        return self._sanitize_manifest_for_read(response, resource=plural)

    @staticmethod
    def _to_manifest(payload: dict[str, object], *, resource: str) -> dict[str, object]:
        return {
            "apiVersion": payload.get("apiVersion"),
            "kind": payload.get("kind") or resource,
            "metadata": payload.get("metadata"),
            "spec": payload.get("spec"),
        }

    def _get_core_manifest(
        self,
        namespace: str,
        version: str,
        resource: str,
        name: str,
    ) -> dict[str, object] | None:
        if self._core_api is None:
            return None
        path = f"/api/{version}/namespaces/{namespace}/{resource}/{name}"
        try:
            response, _, _ = self._core_api.api_client.call_api(
                path,
                "GET",
                auth_settings=["BearerToken"],
                response_type="object",
                _return_http_data_only=False,
                _preload_content=True,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning(
                "Failed to get core resource %s/%s %s/%s: %s",
                version,
                resource,
                namespace,
                name,
                exc,
            )
            return None
        if not isinstance(response, dict):
            return None
        return self._sanitize_manifest_for_read(response, resource=resource)

    def _list_core_manifests(
        self,
        namespace: str,
        version: str,
        resource: str,
        *,
        label_selector: str | None = None,
        limit: int = 20,
    ) -> list[dict[str, object]]:
        if self._core_api is None:
            return []
        path = f"/api/{version}/namespaces/{namespace}/{resource}"
        query_params: list[tuple[str, str]] = []
        if label_selector:
            query_params.append(("labelSelector", label_selector))
        if limit > 0:
            query_params.append(("limit", str(limit)))
        try:
            response, _, _ = self._core_api.api_client.call_api(
                path,
                "GET",
                auth_settings=["BearerToken"],
                response_type="object",
                _return_http_data_only=False,
                _preload_content=True,
                _request_timeout=self._timeout_seconds,
                query_params=query_params,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning(
                "Failed to list core resources %s/%s in namespace %s: %s",
                version,
                resource,
                namespace,
                exc,
            )
            return []
        if not isinstance(response, dict):
            return []
        items = response.get("items")
        if not isinstance(items, list):
            return []
        manifests: list[dict[str, object]] = []
        for item in items:
            if isinstance(item, dict):
                manifests.append(self._to_manifest(item, resource=resource))
        return manifests

    @staticmethod
    def _parse_api_version(api_version: str) -> tuple[str | None, str] | None:
        value = api_version.strip()
        if not value:
            return None
        if "/" not in value:
            return None, value
        group, version = value.split("/", 1)
        if not group or not version:
            return None
        return group, version

    @staticmethod
    def _sanitize_manifest_for_read(
        payload: dict[str, object], *, resource: str
    ) -> dict[str, object]:
        # Return near-raw manifest for diagnostics while stripping noisy/sensitive fields.
        sanitized = dict(payload)

        metadata = sanitized.get("metadata")
        if isinstance(metadata, dict):
            metadata_copy = dict(metadata)
            metadata_copy.pop("managedFields", None)
            sanitized["metadata"] = metadata_copy

        sanitized.pop("status", None)

        kind = str(sanitized.get("kind") or "").lower()
        if resource == "secrets" or kind == "secret":
            data = sanitized.get("data")
            if isinstance(data, dict):
                sanitized["data"] = {str(key): "[MASKED]" for key in data}
            string_data = sanitized.get("stringData")
            if isinstance(string_data, dict):
                sanitized["stringData"] = {str(key): "[MASKED]" for key in string_data}

        return sanitized

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

    def _list_namespace_events_core(
        self, namespace: str
    ) -> tuple[list[PodEventSummary], str | None]:
        if self._core_api is None:
            return [], None
        try:
            response = self._core_api.list_namespaced_event(
                namespace=namespace,
                limit=self._event_limit,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to list core events for %s: %s", namespace, exc)
            return [], f"core events query failed: {exc}"
        return [self._to_event_summary(item) for item in response.items], None

    def _list_namespace_events_v1(
        self, namespace: str
    ) -> tuple[list[PodEventSummary], str | None]:
        if self._events_api is None:
            return [], None
        try:
            response = self._events_api.list_namespaced_event(
                namespace=namespace,
                limit=self._event_limit,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to list events.k8s.io for %s: %s", namespace, exc)
            return [], f"events.k8s.io query failed: {exc}"
        return [self._to_event_summary_v1(item) for item in response.items], None

    def _list_cluster_events_core(self) -> tuple[list[PodEventSummary], str | None]:
        if self._core_api is None:
            return [], None
        try:
            response = self._core_api.list_event_for_all_namespaces(
                limit=self._event_limit,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to list core cluster events: %s", exc)
            return [], f"core cluster events query failed: {exc}"
        return [self._to_event_summary(item) for item in response.items], None

    def _list_cluster_events_v1(self) -> tuple[list[PodEventSummary], str | None]:
        if self._events_api is None:
            return [], None
        try:
            response = self._events_api.list_event_for_all_namespaces(
                limit=self._event_limit,
                _request_timeout=self._timeout_seconds,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to list events.k8s.io cluster events: %s", exc)
            return [], f"events.k8s.io cluster events query failed: {exc}"
        return [self._to_event_summary_v1(item) for item in response.items], None

    @staticmethod
    def _event_warnings(messages: list[str | None]) -> list[PodEventSummary]:
        combined = "; ".join(message for message in messages if message)
        if not combined:
            return []
        return [
            PodEventSummary(
                type="Warning",
                reason="EventQueryFailed",
                message=combined,
                count=None,
                first_timestamp=None,
                last_timestamp=None,
                involved_object=None,
            )
        ]

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

    def _get_current_logs(
        self,
        namespace: str,
        pod: client.V1Pod | None,
        *,
        container: str | None,
        tail_lines: int | None,
        since_seconds: int | None,
    ) -> list[PodLogSnippet]:
        if pod is None:
            return []

        if container:
            container_names = [container]
        else:
            container_names = [container.name for container in pod.spec.containers or []]

        snippets: list[PodLogSnippet] = []
        for container_name in container_names:
            try:
                effective_tail_lines = self._log_tail_lines
                if tail_lines is not None:
                    effective_tail_lines = min(tail_lines, self._log_tail_lines)
                logs = self._core_api.read_namespaced_pod_log(
                    name=pod.metadata.name,
                    namespace=namespace,
                    container=container_name,
                    previous=False,
                    tail_lines=effective_tail_lines,
                    since_seconds=since_seconds,
                    timestamps=True,
                    _request_timeout=self._timeout_seconds,
                )
                snippets.append(
                    PodLogSnippet(
                        container=container_name,
                        previous=False,
                        logs=logs.splitlines() if logs else [],
                    )
                )
            except Exception as exc:  # noqa: BLE001
                self._logger.warning(
                    "Failed to read logs for %s/%s (%s): %s",
                    namespace,
                    pod.metadata.name,
                    container_name,
                    exc,
                )
                snippets.append(
                    PodLogSnippet(
                        container=container_name,
                        previous=False,
                        logs=[],
                        error="failed to read logs",
                    )
                )
        return snippets

    def _summarize_pod_spec(self, pod: client.V1Pod) -> dict[str, object]:
        spec = pod.spec
        return {
            "service_account": spec.service_account_name,
            "node_name": spec.node_name,
            "restart_policy": spec.restart_policy,
            "priority_class_name": spec.priority_class_name,
            "image_pull_secrets": [item.name for item in spec.image_pull_secrets or []],
            "node_selector": spec.node_selector,
            "tolerations": [item.to_dict() for item in spec.tolerations or []],
            "affinity": spec.affinity.to_dict() if spec.affinity else None,
            "volumes": self._summarize_volumes(spec.volumes or []),
            "init_containers": [
                self._summarize_container(item) for item in spec.init_containers or []
            ],
            "containers": [self._summarize_container(item) for item in spec.containers or []],
        }

    @staticmethod
    def _summarize_volumes(volumes: list[client.V1Volume]) -> list[dict[str, object]]:
        summary: list[dict[str, object]] = []
        for volume in volumes:
            item: dict[str, object] = {"name": volume.name}
            if volume.config_map:
                item["config_map"] = {"name": volume.config_map.name}
            elif volume.secret:
                item["secret"] = {"secret_name": volume.secret.secret_name}
            elif volume.persistent_volume_claim:
                item["persistent_volume_claim"] = {
                    "claim_name": volume.persistent_volume_claim.claim_name
                }
            elif volume.empty_dir:
                item["empty_dir"] = volume.empty_dir.to_dict()
            elif volume.host_path:
                item["host_path"] = volume.host_path.to_dict()
            else:
                item["raw"] = volume.to_dict()
            summary.append(item)
        return summary

    def _summarize_container(self, container: client.V1Container) -> dict[str, object]:
        return {
            "name": container.name,
            "image": container.image,
            "image_pull_policy": container.image_pull_policy,
            "command": container.command,
            "args": container.args,
            "resources": container.resources.to_dict() if container.resources else None,
            "ports": [
                {
                    "name": port.name,
                    "container_port": port.container_port,
                    "protocol": port.protocol,
                }
                for port in container.ports or []
            ],
            "liveness_probe": self._probe_to_dict(container.liveness_probe),
            "readiness_probe": self._probe_to_dict(container.readiness_probe),
            "startup_probe": self._probe_to_dict(container.startup_probe),
            "env": [
                {
                    "name": env.name,
                    "value_from": self._env_value_from_to_dict(env.value_from),
                }
                for env in container.env or []
            ],
        }

    @staticmethod
    def _probe_to_dict(probe: client.V1Probe | None) -> dict[str, object] | None:
        if probe is None:
            return None
        return probe.to_dict()

    @staticmethod
    def _env_value_from_to_dict(
        value_from: client.V1EnvVarSource | None,
    ) -> dict[str, object] | None:
        if value_from is None:
            return None
        return value_from.to_dict()

    @staticmethod
    def _select_owner_reference(
        owner_references: list[client.V1OwnerReference],
    ) -> client.V1OwnerReference | None:
        if not owner_references:
            return None
        for owner_reference in owner_references:
            if owner_reference.controller:
                return owner_reference
        return owner_references[0]

    def _resolve_owner_reference(
        self,
        namespace: str,
        owner_reference: client.V1OwnerReference,
    ) -> client.V1OwnerReference:
        if owner_reference.kind == "ReplicaSet" and self._apps_api:
            try:
                replica_set = self._apps_api.read_namespaced_replica_set(
                    name=owner_reference.name,
                    namespace=namespace,
                    _request_timeout=self._timeout_seconds,
                )
            except Exception as exc:  # noqa: BLE001
                self._logger.warning(
                    "Failed to read ReplicaSet %s/%s: %s",
                    namespace,
                    owner_reference.name,
                    exc,
                )
                return owner_reference
            deployment_owner = self._select_owner_reference(
                replica_set.metadata.owner_references or []
            )
            if deployment_owner and deployment_owner.kind == "Deployment":
                return deployment_owner
        if owner_reference.kind == "Job" and self._batch_api:
            try:
                job = self._batch_api.read_namespaced_job(
                    name=owner_reference.name,
                    namespace=namespace,
                    _request_timeout=self._timeout_seconds,
                )
            except Exception as exc:  # noqa: BLE001
                self._logger.warning(
                    "Failed to read Job %s/%s: %s",
                    namespace,
                    owner_reference.name,
                    exc,
                )
                return owner_reference
            cron_job_owner = self._select_owner_reference(job.metadata.owner_references or [])
            if cron_job_owner and cron_job_owner.kind == "CronJob":
                return cron_job_owner
        return owner_reference

    def _summarize_deployment(self, deployment: client.V1Deployment) -> dict[str, object]:
        status = deployment.status
        metadata = deployment.metadata
        return {
            "kind": "Deployment",
            "name": metadata.name if metadata else "",
            "namespace": metadata.namespace if metadata else "",
            "generation": metadata.generation if metadata else None,
            "observed_generation": status.observed_generation if status else None,
            "replicas": status.replicas if status else None,
            "ready_replicas": status.ready_replicas if status else None,
            "updated_replicas": status.updated_replicas if status else None,
            "available_replicas": status.available_replicas if status else None,
            "unavailable_replicas": status.unavailable_replicas if status else None,
            "conditions": self._summarize_conditions(status.conditions if status else None),
        }

    def _summarize_replica_set(self, replica_set: client.V1ReplicaSet) -> dict[str, object]:
        status = replica_set.status
        metadata = replica_set.metadata
        return {
            "kind": "ReplicaSet",
            "name": metadata.name if metadata else "",
            "namespace": metadata.namespace if metadata else "",
            "generation": metadata.generation if metadata else None,
            "observed_generation": status.observed_generation if status else None,
            "replicas": status.replicas if status else None,
            "ready_replicas": status.ready_replicas if status else None,
            "available_replicas": status.available_replicas if status else None,
            "conditions": self._summarize_conditions(status.conditions if status else None),
        }

    def _summarize_stateful_set(self, stateful_set: client.V1StatefulSet) -> dict[str, object]:
        status = stateful_set.status
        metadata = stateful_set.metadata
        return {
            "kind": "StatefulSet",
            "name": metadata.name if metadata else "",
            "namespace": metadata.namespace if metadata else "",
            "generation": metadata.generation if metadata else None,
            "observed_generation": status.observed_generation if status else None,
            "replicas": status.replicas if status else None,
            "ready_replicas": status.ready_replicas if status else None,
            "current_replicas": status.current_replicas if status else None,
            "updated_replicas": status.updated_replicas if status else None,
            "available_replicas": status.available_replicas if status else None,
            "conditions": self._summarize_conditions(status.conditions if status else None),
        }

    def _summarize_daemon_set(self, daemon_set: client.V1DaemonSet) -> dict[str, object]:
        status = daemon_set.status
        metadata = daemon_set.metadata
        spec = daemon_set.spec
        template_metadata = spec.template.metadata if spec and spec.template else None
        template_spec = spec.template.spec if spec and spec.template else None
        return {
            "kind": "DaemonSet",
            "name": metadata.name if metadata else "",
            "namespace": metadata.namespace if metadata else "",
            "generation": metadata.generation if metadata else None,
            "observed_generation": status.observed_generation if status else None,
            "desired_number_scheduled": status.desired_number_scheduled if status else None,
            "current_number_scheduled": status.current_number_scheduled if status else None,
            "updated_number_scheduled": status.updated_number_scheduled if status else None,
            "number_ready": status.number_ready if status else None,
            "number_available": status.number_available if status else None,
            "number_unavailable": status.number_unavailable if status else None,
            "number_misscheduled": status.number_misscheduled if status else None,
            "conditions": self._summarize_conditions(status.conditions if status else None),
            "update_strategy": (
                spec.update_strategy.to_dict() if spec and spec.update_strategy else None
            ),
            "selector": spec.selector.to_dict() if spec and spec.selector else None,
            "template_labels": template_metadata.labels if template_metadata else {},
            "node_selector": template_spec.node_selector if template_spec else None,
            "tolerations": [
                toleration.to_dict() for toleration in (template_spec.tolerations or [])
            ]
            if template_spec
            else [],
            "affinity": template_spec.affinity.to_dict()
            if template_spec and template_spec.affinity
            else None,
        }

    def _summarize_job(self, job: client.V1Job) -> dict[str, object]:
        status = job.status
        metadata = job.metadata
        return {
            "kind": "Job",
            "name": metadata.name if metadata else "",
            "namespace": metadata.namespace if metadata else "",
            "generation": metadata.generation if metadata else None,
            "start_time": self._to_iso(status.start_time) if status else None,
            "completion_time": self._to_iso(status.completion_time) if status else None,
            "active": status.active if status else None,
            "succeeded": status.succeeded if status else None,
            "failed": status.failed if status else None,
            "conditions": self._summarize_conditions(status.conditions if status else None),
        }

    def _summarize_cron_job(self, cron_job: client.V1CronJob) -> dict[str, object]:
        status = cron_job.status
        metadata = cron_job.metadata
        return {
            "kind": "CronJob",
            "name": metadata.name if metadata else "",
            "namespace": metadata.namespace if metadata else "",
            "generation": metadata.generation if metadata else None,
            "active": [job.name for job in status.active or []] if status else [],
            "last_schedule_time": self._to_iso(status.last_schedule_time) if status else None,
            "last_successful_time": self._to_iso(status.last_successful_time) if status else None,
        }

    def _summarize_conditions(
        self, conditions: list[client.V1DeploymentCondition] | None
    ) -> list[dict[str, object]]:
        return [
            {
                "type": condition.type,
                "status": condition.status,
                "reason": condition.reason,
                "message": condition.message,
                "last_transition_time": self._to_iso(condition.last_transition_time),
            }
            for condition in conditions or []
        ]

    @staticmethod
    def _summarize_pod_metrics(response: dict[str, object]) -> dict[str, object]:
        metadata = response.get("metadata", {}) if isinstance(response, dict) else {}
        containers = []
        for container in response.get("containers", []) if isinstance(response, dict) else []:
            containers.append({"name": container.get("name"), "usage": container.get("usage")})
        return {
            "name": metadata.get("name"),
            "namespace": metadata.get("namespace"),
            "timestamp": response.get("timestamp"),
            "window": response.get("window"),
            "containers": containers,
        }

    @staticmethod
    def _summarize_node_metrics(response: dict[str, object]) -> dict[str, object]:
        metadata = response.get("metadata", {}) if isinstance(response, dict) else {}
        return {
            "name": metadata.get("name"),
            "timestamp": response.get("timestamp"),
            "window": response.get("window"),
            "usage": response.get("usage"),
        }

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
            involved_object=self._summarize_object_ref(event.involved_object),
        )

    def _to_event_summary_v1(self, event: client.EventsV1Event) -> PodEventSummary:
        metadata = getattr(event, "metadata", None)
        series = getattr(event, "series", None)
        regarding = getattr(event, "regarding", None) or getattr(event, "related", None)
        message = getattr(event, "note", None) or getattr(event, "deprecated_message", None)
        first_timestamp = (
            getattr(event, "deprecated_first_timestamp", None)
            or getattr(event, "event_time", None)
            or getattr(metadata, "creation_timestamp", None)
        )
        last_timestamp = (
            getattr(event, "deprecated_last_timestamp", None)
            or getattr(series, "last_observed_time", None)
            or getattr(event, "event_time", None)
            or getattr(metadata, "creation_timestamp", None)
        )
        count = getattr(series, "count", None) or getattr(event, "deprecated_count", None)
        return PodEventSummary(
            type=getattr(event, "type", None),
            reason=getattr(event, "reason", None),
            message=message,
            count=count,
            first_timestamp=self._to_iso(first_timestamp),
            last_timestamp=self._to_iso(last_timestamp),
            involved_object=self._summarize_object_ref(regarding),
        )

    @staticmethod
    def _summarize_object_ref(obj: object | None) -> dict[str, str | None] | None:
        if obj is None:
            return None
        return {
            "kind": getattr(obj, "kind", None),
            "name": getattr(obj, "name", None),
            "namespace": getattr(obj, "namespace", None),
            "uid": getattr(obj, "uid", None),
        }

    @staticmethod
    def _to_iso(value: object | None) -> str | None:
        if value is None:
            return None
        if hasattr(value, "isoformat"):
            return value.isoformat()
        return str(value)


def extract_pod_target(labels: dict[str, str]) -> tuple[str | None, str | None, str | None]:
    """Extract namespace, pod name, and workload from alert labels.

    Returns:
        tuple: (namespace, pod_name, workload)
        - namespace: from 'namespace' or 'destination_service_namespace' labels
        - pod_name: from 'pod' label
        - workload: from 'workload' or 'destination_workload' labels (for pod discovery)
    """
    namespace_keys = ["namespace", "destination_service_namespace"]
    pod_keys = ["pod"]
    workload_keys = ["workload", "destination_workload"]

    namespace = _first_label_value(labels, namespace_keys)
    pod_name = _first_label_value(labels, pod_keys)
    workload = _first_label_value(labels, workload_keys)

    return namespace, pod_name, workload


def _first_label_value(labels: dict[str, str], keys: Iterable[str]) -> str | None:
    for key in keys:
        value = labels.get(key)
        if value:
            return value
    return None
