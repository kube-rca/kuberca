from __future__ import annotations

import json
from typing import Any
from unittest.mock import MagicMock, patch

from kubernetes import client as k8s_client

from app.clients.k8s import KubernetesClient

# ---------------------------------------------------------------------------
# Helpers — fake kubernetes API objects
# ---------------------------------------------------------------------------


def _make_pod(
    name: str = "test-pod",
    namespace: str = "bookinfo",
    phase: str = "Running",
    containers: list[str] | None = None,
    owner_kind: str | None = None,
    owner_name: str | None = None,
) -> k8s_client.V1Pod:
    containers = containers or ["app"]
    container_objs = [k8s_client.V1Container(name=c) for c in containers]
    owner_refs = []
    if owner_kind and owner_name:
        owner_refs.append(
            k8s_client.V1OwnerReference(
                api_version="apps/v1",
                kind=owner_kind,
                name=owner_name,
                uid="uid-1",
                controller=True,
            )
        )
    return k8s_client.V1Pod(
        metadata=k8s_client.V1ObjectMeta(
            name=name,
            namespace=namespace,
            labels={"app": name},
            owner_references=owner_refs,
        ),
        spec=k8s_client.V1PodSpec(
            containers=container_objs,
            node_name="node-1",
        ),
        status=k8s_client.V1PodStatus(
            phase=phase,
            container_statuses=[
                k8s_client.V1ContainerStatus(
                    name=c,
                    ready=True,
                    restart_count=0,
                    image="nginx:latest",
                    image_id="",
                )
                for c in containers
            ],
        ),
    )


def _make_event(
    reason: str = "OOMKilling",
    message: str = "OOM killed",
    type: str = "Warning",
    namespace: str = "bookinfo",
    pod_name: str = "test-pod",
) -> k8s_client.CoreV1Event:
    return k8s_client.CoreV1Event(
        metadata=k8s_client.V1ObjectMeta(name="event-1", namespace=namespace),
        reason=reason,
        message=message,
        type=type,
        count=1,
        involved_object=k8s_client.V1ObjectReference(
            kind="Pod",
            name=pod_name,
            namespace=namespace,
        ),
    )


def _client(
    core_api: Any = None,
    apps_api: Any = None,
    batch_api: Any = None,
    custom_api: Any = None,
    events_api: Any = None,
) -> KubernetesClient:
    """Build a KubernetesClient with pre-wired fake APIs, bypassing _build_client."""
    with patch.object(KubernetesClient, "_build_client", return_value=core_api or MagicMock()):
        inst = KubernetesClient(timeout_seconds=10, event_limit=100, log_tail_lines=50)
    inst._core_api = core_api
    inst._apps_api = apps_api
    inst._batch_api = batch_api
    inst._custom_api = custom_api
    inst._events_api = events_api
    return inst


# ---------------------------------------------------------------------------
# collect_context — namespace / pod_name routing
# ---------------------------------------------------------------------------


def test_collect_context_missing_namespace_returns_warning() -> None:
    c = _client()
    ctx = c.collect_context(namespace=None, pod_name=None)
    assert any("namespace missing" in w for w in ctx.warnings)
    assert ctx.pod_status is None


def test_collect_context_no_core_api_returns_warning() -> None:
    c = _client(core_api=None)
    ctx = c.collect_context(namespace="bookinfo", pod_name="test-pod")
    assert any("not configured" in w for w in ctx.warnings)


def test_collect_context_namespace_only_adds_discovery_hint() -> None:
    core = MagicMock()
    core.list_namespaced_event.return_value = k8s_client.CoreV1EventList(items=[])
    c = _client(core_api=core)
    ctx = c.collect_context(namespace="bookinfo", pod_name=None)
    assert any("pod_name missing" in w for w in ctx.warnings)


def test_collect_context_with_pod_reads_status() -> None:
    core = MagicMock()
    pod = _make_pod()
    core.read_namespaced_pod.return_value = pod
    core.list_namespaced_event.return_value = k8s_client.CoreV1EventList(items=[])
    core.read_namespaced_pod_log.return_value = "log line 1\nlog line 2"

    c = _client(core_api=core)
    ctx = c.collect_context(namespace="bookinfo", pod_name="test-pod")

    assert ctx.pod_status is not None
    assert ctx.pod_status.phase == "Running"
    core.read_namespaced_pod.assert_called_once()


# ---------------------------------------------------------------------------
# get_pod_status
# ---------------------------------------------------------------------------


def test_get_pod_status_no_api_returns_none() -> None:
    c = _client(core_api=None)
    assert c.get_pod_status("bookinfo", "test-pod") is None


def test_get_pod_status_pod_not_found_returns_none() -> None:
    core = MagicMock()
    core.read_namespaced_pod.side_effect = Exception("not found")

    c = _client(core_api=core)
    result = c.get_pod_status("bookinfo", "test-pod")
    assert result is None


def test_get_pod_status_returns_snapshot() -> None:
    core = MagicMock()
    pod = _make_pod(phase="CrashLoopBackOff")
    core.read_namespaced_pod.return_value = pod

    c = _client(core_api=core)
    result = c.get_pod_status("bookinfo", "test-pod")
    assert result is not None
    assert result.phase == "CrashLoopBackOff"


# ---------------------------------------------------------------------------
# list_pod_events
# ---------------------------------------------------------------------------


def test_list_pod_events_no_api_returns_empty() -> None:
    c = _client(core_api=None)
    assert c.list_pod_events("bookinfo", "test-pod") == []


def test_list_pod_events_returns_summaries() -> None:
    core = MagicMock()
    core.list_namespaced_event.return_value = k8s_client.CoreV1EventList(
        items=[_make_event(reason="OOMKilling")]
    )

    c = _client(core_api=core)
    events = c.list_pod_events("bookinfo", "test-pod")

    assert len(events) == 1
    assert events[0].reason == "OOMKilling"


def test_list_pod_events_api_error_returns_empty() -> None:
    core = MagicMock()
    core.list_namespaced_event.side_effect = Exception("API error")

    c = _client(core_api=core)
    events = c.list_pod_events("bookinfo", "test-pod")
    assert events == []


# ---------------------------------------------------------------------------
# list_pods_in_namespace
# ---------------------------------------------------------------------------


def test_list_pods_no_api_returns_empty() -> None:
    c = _client(core_api=None)
    assert c.list_pods_in_namespace("bookinfo") == []


def test_list_pods_returns_summaries() -> None:
    core = MagicMock()
    core.list_namespaced_pod.return_value = k8s_client.V1PodList(items=[_make_pod()])

    c = _client(core_api=core)
    pods = c.list_pods_in_namespace("bookinfo")

    assert len(pods) == 1
    assert pods[0].name == "test-pod"
    assert pods[0].namespace == "bookinfo"


def test_list_pods_with_label_selector() -> None:
    core = MagicMock()
    core.list_namespaced_pod.return_value = k8s_client.V1PodList(items=[_make_pod()])

    c = _client(core_api=core)
    c.list_pods_in_namespace("bookinfo", label_selector="app=test-pod")

    call_kwargs = core.list_namespaced_pod.call_args.kwargs
    assert call_kwargs.get("label_selector") == "app=test-pod"


def test_list_pods_api_error_returns_empty() -> None:
    core = MagicMock()
    core.list_namespaced_pod.side_effect = Exception("timeout")

    c = _client(core_api=core)
    result = c.list_pods_in_namespace("bookinfo")
    assert result == []


# ---------------------------------------------------------------------------
# get_previous_logs
# ---------------------------------------------------------------------------


def test_get_previous_logs_no_api_returns_empty() -> None:
    c = _client(core_api=None)
    assert c.get_previous_logs("bookinfo", "test-pod") == []


def test_get_previous_logs_returns_snippets() -> None:
    core = MagicMock()
    pod = _make_pod(containers=["app", "sidecar"])
    core.read_namespaced_pod.return_value = pod
    core.read_namespaced_pod_log.return_value = "OOMKilled\nkilled"

    c = _client(core_api=core)
    snippets = c.get_previous_logs("bookinfo", "test-pod")

    assert len(snippets) == 2
    assert all(s.previous for s in snippets)
    assert snippets[0].container == "app"


def test_get_previous_logs_api_error_returns_snippet_with_error() -> None:
    core = MagicMock()
    pod = _make_pod(containers=["app"])
    core.read_namespaced_pod.return_value = pod
    core.read_namespaced_pod_log.side_effect = Exception("not found")

    c = _client(core_api=core)
    snippets = c.get_previous_logs("bookinfo", "test-pod")

    assert len(snippets) == 1
    assert snippets[0].error is not None


# ---------------------------------------------------------------------------
# get_node_status
# ---------------------------------------------------------------------------


def test_get_node_status_no_api_returns_none() -> None:
    c = _client(core_api=None)
    assert c.get_node_status("node-1") is None


def test_get_node_status_returns_dict() -> None:
    core = MagicMock()
    node = k8s_client.V1Node(
        metadata=k8s_client.V1ObjectMeta(name="node-1", labels={"role": "worker"}),
        spec=k8s_client.V1NodeSpec(unschedulable=False, taints=[]),
        status=k8s_client.V1NodeStatus(
            conditions=[
                k8s_client.V1NodeCondition(
                    type="Ready",
                    status="True",
                    reason="KubeletReady",
                    message="kubelet is posting ready status",
                )
            ],
            capacity={"cpu": "4", "memory": "8Gi"},
            allocatable={"cpu": "3900m", "memory": "7Gi"},
        ),
    )
    core.read_node.return_value = node

    c = _client(core_api=core)
    result = c.get_node_status("node-1")

    assert result is not None
    assert result["name"] == "node-1"
    assert result["unschedulable"] is False
    conditions = result["conditions"]
    assert isinstance(conditions, list)
    assert len(conditions) == 1
    assert conditions[0]["type"] == "Ready"  # type: ignore[index]


def test_get_node_status_api_error_returns_none() -> None:
    core = MagicMock()
    core.read_node.side_effect = Exception("unauthorized")

    c = _client(core_api=core)
    assert c.get_node_status("node-1") is None


# ---------------------------------------------------------------------------
# namespace events — v1 path (nosec B112 line coverage)
# ---------------------------------------------------------------------------


def test_list_namespace_events_v1_parses_raw_json() -> None:
    core = MagicMock()
    events_api = MagicMock()

    items_payload = [
        {
            "reason": "OOMKilling",
            "note": "OOM killed",
            "type": "Warning",
            "metadata": {"name": "e1", "namespace": "bookinfo"},
            "regarding": {"kind": "Pod", "name": "test-pod", "namespace": "bookinfo"},
            "series": None,
            "reportingComponent": "",
        }
    ]
    raw_response = MagicMock()
    raw_response.data = json.dumps({"items": items_payload}).encode()
    events_api.list_namespaced_event.return_value = raw_response

    c = _client(core_api=core, events_api=events_api)
    result = c.list_namespace_events("bookinfo")

    assert len(result) >= 1


def test_list_namespace_events_v1_drops_malformed_items() -> None:
    """Covers the nosec B112 branch: malformed items are silently dropped."""
    core = MagicMock()
    events_api = MagicMock()

    items_payload = [
        None,  # malformed — will raise during _to_event_summary_v1_raw
        {
            "reason": "Pulled",
            "note": "Successfully pulled image",
            "type": "Normal",
            "metadata": {"name": "e2", "namespace": "bookinfo"},
            "regarding": {"kind": "Pod", "name": "test-pod", "namespace": "bookinfo"},
        },
    ]
    raw_response = MagicMock()
    raw_response.data = json.dumps({"items": items_payload}).encode()
    events_api.list_namespaced_event.return_value = raw_response

    c = _client(core_api=core, events_api=events_api)
    # Should not raise; malformed item is skipped
    result = c.list_namespace_events("bookinfo")
    assert isinstance(result, list)


def test_list_namespace_events_v1_api_error_falls_back_to_core() -> None:
    core = MagicMock()
    core.list_namespaced_event.return_value = k8s_client.CoreV1EventList(items=[])
    events_api = MagicMock()
    events_api.list_namespaced_event.side_effect = Exception("API unavailable")

    c = _client(core_api=core, events_api=events_api)
    result = c.list_namespace_events("bookinfo")
    assert isinstance(result, list)


# ---------------------------------------------------------------------------
# _sanitize_manifest_for_read — masking
# ---------------------------------------------------------------------------


def test_sanitize_manifest_masks_secret_data() -> None:
    payload: dict[str, Any] = {
        "apiVersion": "v1",
        "kind": "Secret",
        "metadata": {"name": "my-secret", "namespace": "default", "managedFields": ["noise"]},
        "data": {"token": "c2VjcmV0", "password": "cGFzc3dvcmQ="},
        "status": {"some": "field"},
    }
    result = KubernetesClient._sanitize_manifest_for_read(payload, resource="secrets")

    assert result["data"] == {"token": "[MASKED]", "password": "[MASKED]"}  # type: ignore[comparison-overlap]
    assert "status" not in result
    assert "managedFields" not in result["metadata"]  # type: ignore[operator]


def test_sanitize_manifest_strips_managed_fields_non_secret() -> None:
    payload: dict[str, Any] = {
        "apiVersion": "apps/v1",
        "kind": "Deployment",
        "metadata": {"name": "my-deploy", "managedFields": ["noisy"]},
        "spec": {"replicas": 3},
        "status": {"readyReplicas": 3},
    }
    result = KubernetesClient._sanitize_manifest_for_read(payload, resource="deployments")

    assert "status" not in result
    assert "managedFields" not in result["metadata"]  # type: ignore[operator]
    assert result["spec"] == {"replicas": 3}  # type: ignore[comparison-overlap]


# ---------------------------------------------------------------------------
# _parse_api_version
# ---------------------------------------------------------------------------


def test_parse_api_version_core() -> None:
    assert KubernetesClient._parse_api_version("v1") == (None, "v1")


def test_parse_api_version_group() -> None:
    assert KubernetesClient._parse_api_version("apps/v1") == ("apps", "v1")


def test_parse_api_version_empty_returns_none() -> None:
    assert KubernetesClient._parse_api_version("") is None


def test_parse_api_version_invalid_group_returns_none() -> None:
    assert KubernetesClient._parse_api_version("/v1") is None


# ---------------------------------------------------------------------------
# get_workload_summary — deployment path
# ---------------------------------------------------------------------------


def test_get_workload_summary_deployment() -> None:
    core = MagicMock()
    apps = MagicMock()

    pod = _make_pod(owner_kind="ReplicaSet", owner_name="test-rs")
    core.read_namespaced_pod.return_value = pod

    rs = k8s_client.V1ReplicaSet(
        metadata=k8s_client.V1ObjectMeta(
            name="test-rs",
            namespace="bookinfo",
            owner_references=[
                k8s_client.V1OwnerReference(
                    api_version="apps/v1",
                    kind="Deployment",
                    name="test-deploy",
                    uid="uid-2",
                    controller=True,
                )
            ],
        ),
        status=k8s_client.V1ReplicaSetStatus(replicas=1, ready_replicas=1),
    )
    apps.read_namespaced_replica_set.return_value = rs

    deployment = k8s_client.V1Deployment(
        metadata=k8s_client.V1ObjectMeta(
            name="test-deploy",
            namespace="bookinfo",
            generation=1,
        ),
        status=k8s_client.V1DeploymentStatus(
            replicas=1,
            ready_replicas=1,
            updated_replicas=1,
            available_replicas=1,
        ),
    )
    apps.read_namespaced_deployment.return_value = deployment

    c = _client(core_api=core, apps_api=apps)
    result = c.get_workload_summary("bookinfo", "test-pod")

    assert result is not None
    assert result["kind"] == "Deployment"
    assert result["name"] == "test-deploy"


def test_get_workload_summary_no_apis_returns_none() -> None:
    c = _client(core_api=None, apps_api=None, batch_api=None)
    assert c.get_workload_summary("bookinfo", "test-pod") is None


# ---------------------------------------------------------------------------
# collect_context with node_name
# ---------------------------------------------------------------------------


def test_collect_context_node_not_found_adds_warning() -> None:
    core = MagicMock()
    core.read_node.side_effect = Exception("not found")

    c = _client(core_api=core)
    ctx = c.collect_context(namespace=None, pod_name=None, node_name="missing-node")

    # node_name provided, so no "namespace missing" early return
    assert any("not found" in w or "inaccessible" in w for w in ctx.warnings)


# ---------------------------------------------------------------------------
# get_pod_logs
# ---------------------------------------------------------------------------


def test_get_pod_logs_no_api_returns_empty() -> None:
    c = _client(core_api=None)
    assert c.get_pod_logs("bookinfo", "test-pod") == []


def test_get_pod_logs_returns_snippets() -> None:
    core = MagicMock()
    pod = _make_pod(containers=["app"])
    core.read_namespaced_pod.return_value = pod
    core.read_namespaced_pod_log.return_value = "line1\nline2"

    c = _client(core_api=core)
    snippets = c.get_pod_logs("bookinfo", "test-pod")

    assert len(snippets) == 1
    assert snippets[0].logs == ["line1", "line2"]
    assert not snippets[0].previous


def test_get_pod_logs_specific_container() -> None:
    core = MagicMock()
    pod = _make_pod(containers=["app", "sidecar"])
    core.read_namespaced_pod.return_value = pod
    core.read_namespaced_pod_log.return_value = "output"

    c = _client(core_api=core)
    snippets = c.get_pod_logs("bookinfo", "test-pod", container="sidecar")

    assert len(snippets) == 1
    assert snippets[0].container == "sidecar"


def test_get_pod_logs_api_error_returns_snippet_with_error() -> None:
    core = MagicMock()
    pod = _make_pod(containers=["app"])
    core.read_namespaced_pod.return_value = pod
    core.read_namespaced_pod_log.side_effect = Exception("forbidden")

    c = _client(core_api=core)
    snippets = c.get_pod_logs("bookinfo", "test-pod")

    assert len(snippets) == 1
    assert snippets[0].error is not None


# ---------------------------------------------------------------------------
# list_services_by_label
# ---------------------------------------------------------------------------


def test_list_services_no_api_returns_empty() -> None:
    c = _client(core_api=None)
    assert c.list_services_by_label("app=nginx") == []


def test_list_services_empty_selector_returns_empty() -> None:
    c = _client(core_api=MagicMock())
    assert c.list_services_by_label("") == []


def test_list_services_all_namespaces() -> None:
    core = MagicMock()
    svc = k8s_client.V1Service(
        metadata=k8s_client.V1ObjectMeta(name="my-svc", namespace="bookinfo")
    )
    core.list_service_for_all_namespaces.return_value = k8s_client.V1ServiceList(items=[svc])

    c = _client(core_api=core)
    result = c.list_services_by_label("app=nginx")

    assert len(result) == 1
    core.list_service_for_all_namespaces.assert_called_once()


def test_list_services_specific_namespaces() -> None:
    core = MagicMock()
    svc = k8s_client.V1Service(
        metadata=k8s_client.V1ObjectMeta(name="my-svc", namespace="bookinfo")
    )
    core.list_namespaced_service.return_value = k8s_client.V1ServiceList(items=[svc])

    c = _client(core_api=core)
    result = c.list_services_by_label("app=nginx", namespaces=["bookinfo"])

    assert len(result) == 1
    core.list_namespaced_service.assert_called_once()


def test_list_services_api_error_returns_empty() -> None:
    core = MagicMock()
    core.list_service_for_all_namespaces.side_effect = Exception("timeout")

    c = _client(core_api=core)
    result = c.list_services_by_label("app=nginx")
    assert result == []


# ---------------------------------------------------------------------------
# get_pod_spec_summary
# ---------------------------------------------------------------------------


def test_get_pod_spec_summary_no_api_returns_none() -> None:
    c = _client(core_api=None)
    assert c.get_pod_spec_summary("bookinfo", "test-pod") is None


def test_get_pod_spec_summary_pod_not_found_returns_none() -> None:
    core = MagicMock()
    core.read_namespaced_pod.side_effect = Exception("not found")

    c = _client(core_api=core)
    assert c.get_pod_spec_summary("bookinfo", "test-pod") is None


def test_get_pod_spec_summary_returns_dict() -> None:
    core = MagicMock()
    pod = _make_pod(containers=["app"])
    core.read_namespaced_pod.return_value = pod

    c = _client(core_api=core)
    result = c.get_pod_spec_summary("bookinfo", "test-pod")

    assert result is not None
    assert "containers" in result
    assert result["node_name"] == "node-1"


# ---------------------------------------------------------------------------
# get_workload_summary — StatefulSet / DaemonSet / error paths
# ---------------------------------------------------------------------------


def test_get_workload_summary_stateful_set() -> None:
    core = MagicMock()
    apps = MagicMock()

    pod = _make_pod(owner_kind="StatefulSet", owner_name="test-ss")
    core.read_namespaced_pod.return_value = pod

    ss = k8s_client.V1StatefulSet(
        metadata=k8s_client.V1ObjectMeta(
            name="test-ss",
            namespace="bookinfo",
            owner_references=[],
        ),
        status=k8s_client.V1StatefulSetStatus(
            replicas=1,
            ready_replicas=1,
            current_replicas=1,
            observed_generation=1,
        ),
    )
    apps.read_namespaced_stateful_set.return_value = ss

    c = _client(core_api=core, apps_api=apps)
    result = c.get_workload_summary("bookinfo", "test-pod")

    assert result is not None
    assert result["kind"] == "StatefulSet"


def test_get_workload_summary_daemon_set() -> None:
    core = MagicMock()
    apps = MagicMock()

    pod = _make_pod(owner_kind="DaemonSet", owner_name="test-ds")
    core.read_namespaced_pod.return_value = pod

    ds = k8s_client.V1DaemonSet(
        metadata=k8s_client.V1ObjectMeta(
            name="test-ds",
            namespace="bookinfo",
            owner_references=[],
        ),
        status=k8s_client.V1DaemonSetStatus(
            number_ready=1,
            desired_number_scheduled=1,
            current_number_scheduled=1,
            number_misscheduled=0,
        ),
    )
    apps.read_namespaced_daemon_set.return_value = ds

    c = _client(core_api=core, apps_api=apps)
    result = c.get_workload_summary("bookinfo", "test-pod")

    assert result is not None
    assert result["kind"] == "DaemonSet"


def test_get_workload_summary_pod_not_found_returns_none() -> None:
    core = MagicMock()
    apps = MagicMock()
    core.read_namespaced_pod.side_effect = Exception("not found")

    c = _client(core_api=core, apps_api=apps)
    assert c.get_workload_summary("bookinfo", "test-pod") is None


def test_get_workload_summary_no_owner_returns_none() -> None:
    core = MagicMock()
    apps = MagicMock()
    pod = _make_pod()  # no owner references
    core.read_namespaced_pod.return_value = pod

    c = _client(core_api=core, apps_api=apps)
    assert c.get_workload_summary("bookinfo", "test-pod") is None


# ---------------------------------------------------------------------------
# get_manifest / list_manifests — custom resource path
# ---------------------------------------------------------------------------


def test_get_manifest_custom_resource() -> None:
    custom = MagicMock()
    payload = {
        "apiVersion": "chaos-mesh.org/v1alpha1",
        "kind": "PodChaos",
        "metadata": {"name": "my-chaos", "namespace": "bookinfo"},
        "spec": {"action": "pod-kill"},
    }
    custom.get_namespaced_custom_object.return_value = payload

    c = _client(custom_api=custom)
    result = c.get_manifest(
        namespace="bookinfo",
        api_version="chaos-mesh.org/v1alpha1",
        resource="podchaos",
        name="my-chaos",
    )

    assert result is not None
    custom.get_namespaced_custom_object.assert_called_once()


def test_get_manifest_invalid_api_version_returns_none() -> None:
    c = _client()
    result = c.get_manifest(
        namespace="bookinfo",
        api_version="",
        resource="pods",
        name="test-pod",
    )
    assert result is None


def test_list_manifests_custom_resource() -> None:
    custom = MagicMock()
    custom.list_namespaced_custom_object.return_value = {
        "items": [
            {
                "apiVersion": "chaos-mesh.org/v1alpha1",
                "kind": "PodChaos",
                "metadata": {"name": "chaos-1"},
                "spec": {"action": "pod-kill"},
            }
        ]
    }

    c = _client(custom_api=custom)
    results = c.list_manifests(
        namespace="bookinfo",
        api_version="chaos-mesh.org/v1alpha1",
        resource="podchaos",
    )

    assert len(results) == 1


def test_list_manifests_empty_resource_returns_empty() -> None:
    c = _client()
    assert c.list_manifests(namespace="bookinfo", api_version="v1", resource="") == []


# ---------------------------------------------------------------------------
# cluster events
# ---------------------------------------------------------------------------


def test_list_cluster_events_v1_fallback_to_core() -> None:
    core = MagicMock()
    core.list_event_for_all_namespaces.return_value = k8s_client.CoreV1EventList(
        items=[_make_event()]
    )
    events_api = MagicMock()
    events_api.list_event_for_all_namespaces.side_effect = Exception("unavailable")

    c = _client(core_api=core, events_api=events_api)
    result = c.list_cluster_events()

    assert isinstance(result, list)
    assert len(result) >= 1


def test_list_cluster_events_no_apis_returns_empty() -> None:
    c = _client(core_api=None, events_api=None)
    assert c.list_cluster_events() == []


# ---------------------------------------------------------------------------
# get_pod_metrics
# ---------------------------------------------------------------------------


def test_get_pod_metrics_no_api_returns_none() -> None:
    c = _client(custom_api=None)
    assert c.get_pod_metrics("bookinfo", "test-pod") is None


def test_get_pod_metrics_api_error_returns_none() -> None:
    custom = MagicMock()
    custom.get_namespaced_custom_object.side_effect = Exception("not found")

    c = _client(custom_api=custom)
    assert c.get_pod_metrics("bookinfo", "test-pod") is None
