from __future__ import annotations

import logging

from app.clients.k8s import KubernetesClient


class _FakeApiClient:
    def __init__(self, response: object) -> None:
        self._response = response
        self.calls: list[dict[str, object]] = []

    def call_api(self, path: str, method: str, **kwargs: object) -> tuple[object, object, object]:
        self.calls.append({"path": path, "method": method, **kwargs})
        return self._response, {}, {}


class _FakeCoreApi:
    def __init__(self, response: object) -> None:
        self.api_client = _FakeApiClient(response)


class _FakeCustomApi:
    def __init__(self, get_response: object, list_response: object) -> None:
        self.get_response = get_response
        self.list_response = list_response
        self.get_calls: list[dict[str, object]] = []
        self.list_calls: list[dict[str, object]] = []

    def get_namespaced_custom_object(self, **kwargs: object) -> object:
        self.get_calls.append(kwargs)
        return self.get_response

    def list_namespaced_custom_object(self, **kwargs: object) -> object:
        self.list_calls.append(kwargs)
        return self.list_response


def _build_k8s_client(
    *, core_response: object | None = None, custom_response: object | None = None
) -> KubernetesClient:
    client = KubernetesClient.__new__(KubernetesClient)
    client._logger = logging.getLogger(__name__)
    client._timeout_seconds = 5
    client._event_limit = 25
    client._log_tail_lines = 25
    client._core_api = _FakeCoreApi(core_response) if core_response is not None else None
    if custom_response is not None:
        client._custom_api = _FakeCustomApi(
            get_response=custom_response,
            list_response={"items": [custom_response]},
        )
    else:
        client._custom_api = None
    client._apps_api = None
    client._batch_api = None
    client._events_api = None
    return client


def test_get_manifest_core_v1_uses_core_api_path() -> None:
    k8s_client = _build_k8s_client(
        core_response={
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {"name": "reviews", "managedFields": [{"manager": "kubectl"}]},
            "spec": {"selector": {"app": "reviews"}},
            "data": {"preserved": "yes"},
            "status": {"loadBalancer": {}},
        }
    )

    manifest = k8s_client.get_manifest(
        namespace="bookinfo",
        api_version="v1",
        resource="services",
        name="reviews",
    )

    assert manifest is not None
    assert manifest["kind"] == "Service"
    assert manifest["metadata"] == {"name": "reviews"}
    assert manifest["data"] == {"preserved": "yes"}
    assert "status" not in manifest
    calls = k8s_client._core_api.api_client.calls
    assert calls
    assert calls[-1]["path"] == "/api/v1/namespaces/bookinfo/services/reviews"


def test_list_manifests_core_v1_forwards_selector_and_limit() -> None:
    k8s_client = _build_k8s_client(
        core_response={
            "items": [
                {
                    "apiVersion": "v1",
                    "kind": "ConfigMap",
                    "metadata": {"name": "cm-a"},
                    "spec": None,
                }
            ]
        }
    )

    manifests = k8s_client.list_manifests(
        namespace="bookinfo",
        api_version="v1",
        resource="configmaps",
        label_selector="app=reviews",
        limit=10,
    )

    assert len(manifests) == 1
    calls = k8s_client._core_api.api_client.calls
    assert calls
    assert calls[-1]["path"] == "/api/v1/namespaces/bookinfo/configmaps"
    assert calls[-1]["query_params"] == [("labelSelector", "app=reviews"), ("limit", "10")]


def test_get_manifest_crd_uses_custom_object_api() -> None:
    custom_response = {
        "apiVersion": "networking.istio.io/v1",
        "kind": "VirtualService",
        "metadata": {"name": "reviews-route"},
        "spec": {"hosts": ["reviews.bookinfo.svc.cluster.local"]},
    }
    k8s_client = _build_k8s_client(custom_response=custom_response)

    manifest = k8s_client.get_manifest(
        namespace="bookinfo",
        api_version="networking.istio.io/v1",
        resource="virtualservices",
        name="reviews-route",
    )

    assert manifest is not None
    assert manifest["kind"] == "VirtualService"
    custom_api = k8s_client._custom_api
    assert custom_api is not None
    assert custom_api.get_calls
    assert custom_api.get_calls[-1]["group"] == "networking.istio.io"
    assert custom_api.get_calls[-1]["version"] == "v1"
    assert custom_api.get_calls[-1]["plural"] == "virtualservices"


def test_get_manifest_returns_none_on_invalid_api_version() -> None:
    k8s_client = _build_k8s_client()

    manifest = k8s_client.get_manifest(
        namespace="default",
        api_version="",
        resource="pods",
        name="demo-pod",
    )
    manifests = k8s_client.list_manifests(
        namespace="default",
        api_version="",
        resource="pods",
    )

    assert manifest is None
    assert manifests == []


def test_get_manifest_masks_secret_values() -> None:
    secret_response = {
        "apiVersion": "v1",
        "kind": "Secret",
        "metadata": {"name": "demo-secret"},
        "data": {"password": "c2VjcmV0", "token": "dG9rZW4="},
        "stringData": {"raw": "secret"},
        "type": "Opaque",
    }
    k8s_client = _build_k8s_client(core_response=secret_response)

    manifest = k8s_client.get_manifest(
        namespace="default",
        api_version="v1",
        resource="secrets",
        name="demo-secret",
    )

    assert manifest is not None
    assert manifest["kind"] == "Secret"
    assert manifest["data"] == {"password": "[MASKED]", "token": "[MASKED]"}
    assert manifest["stringData"] == {"raw": "[MASKED]"}
