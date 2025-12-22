from __future__ import annotations

from typing import Protocol

from strands import Agent, tool
from strands.models.gemini import GeminiModel

from app.clients.k8s import KubernetesClient
from app.clients.prometheus import PrometheusClient
from app.clients.strands_patch import apply_gemini_thought_signature_patch
from app.core.config import Settings


class AnalysisEngine(Protocol):
    def analyze(self, prompt: str) -> str:
        raise NotImplementedError


class StrandsAnalysisEngine:
    def __init__(
        self,
        settings: Settings,
        k8s_client: KubernetesClient,
        prometheus_client: PrometheusClient | None = None,
    ) -> None:
        apply_gemini_thought_signature_patch()
        model = GeminiModel(
            client_args={"api_key": settings.gemini_api_key},
            model_id=settings.gemini_model_id,
        )
        self._agent = Agent(model=model, tools=_build_tools(k8s_client, prometheus_client))

    def analyze(self, prompt: str) -> str:
        return str(self._agent(prompt))


def _build_tools(
    k8s_client: KubernetesClient,
    prometheus_client: PrometheusClient | None,
) -> list[object]:
    @tool
    def get_pod_status(namespace: str, pod_name: str) -> dict[str, object]:
        """Fetch pod status details."""
        status = k8s_client.get_pod_status(namespace, pod_name)
        if status is None:
            return {"warning": "pod status not found"}
        return status.to_dict()

    @tool
    def get_pod_spec(namespace: str, pod_name: str) -> dict[str, object]:
        """Fetch a summarized pod spec (containers/resources/probes)."""
        spec = k8s_client.get_pod_spec_summary(namespace, pod_name)
        if spec is None:
            return {"warning": "pod spec not found"}
        return spec

    @tool
    def list_pod_events(namespace: str, pod_name: str) -> list[dict[str, object]]:
        """List recent events for a pod."""
        return [event.to_dict() for event in k8s_client.list_pod_events(namespace, pod_name)]

    @tool
    def list_namespace_events(namespace: str) -> list[dict[str, object]]:
        """List recent events for a namespace."""
        return [event.to_dict() for event in k8s_client.list_namespace_events(namespace)]

    @tool
    def list_cluster_events() -> list[dict[str, object]]:
        """List recent events across all namespaces."""
        return [event.to_dict() for event in k8s_client.list_cluster_events()]

    @tool
    def get_previous_pod_logs(
        namespace: str,
        pod_name: str,
    ) -> list[dict[str, object]]:
        """Return previous container logs (tail)."""
        return [
            snippet.to_dict() for snippet in k8s_client.get_previous_logs(namespace, pod_name)
        ]

    @tool
    def get_pod_logs(
        namespace: str,
        pod_name: str,
        container: str | None = None,
        tail_lines: int | None = None,
        since_seconds: int | None = None,
    ) -> list[dict[str, object]]:
        """Return current container logs (tail)."""
        return [
            snippet.to_dict()
            for snippet in k8s_client.get_pod_logs(
                namespace,
                pod_name,
                container=container,
                tail_lines=tail_lines,
                since_seconds=since_seconds,
            )
        ]

    @tool
    def get_workload_status(namespace: str, pod_name: str) -> dict[str, object]:
        """Fetch top-level workload status (Deployment/StatefulSet/Job)."""
        summary = k8s_client.get_workload_summary(namespace, pod_name)
        if summary is None:
            return {"warning": "workload not found"}
        return summary

    @tool
    def get_node_status(node_name: str) -> dict[str, object]:
        """Fetch node status details."""
        status = k8s_client.get_node_status(node_name)
        if status is None:
            return {"warning": "node not found"}
        return status

    @tool
    def get_pod_metrics(namespace: str, pod_name: str) -> dict[str, object]:
        """Fetch pod CPU/memory usage from metrics-server."""
        metrics = k8s_client.get_pod_metrics(namespace, pod_name)
        if metrics is None:
            return {"warning": "pod metrics not available"}
        return metrics

    @tool
    def get_node_metrics(node_name: str) -> dict[str, object]:
        """Fetch node CPU/memory usage from metrics-server."""
        metrics = k8s_client.get_node_metrics(node_name)
        if metrics is None:
            return {"warning": "node metrics not available"}
        return metrics

    @tool
    def discover_prometheus() -> dict[str, object]:
        """Discover the in-cluster Prometheus service endpoint."""
        if prometheus_client is None:
            return {"warning": "prometheus client not configured"}
        return prometheus_client.describe_endpoint()

    @tool
    def query_prometheus(query: str, time: str | None = None) -> dict[str, object]:
        """Run a Prometheus instant query."""
        if prometheus_client is None:
            return {"warning": "prometheus client not configured"}
        return prometheus_client.query(query, time=time)

    return [
        get_pod_status,
        get_pod_spec,
        list_pod_events,
        list_namespace_events,
        list_cluster_events,
        get_previous_pod_logs,
        get_pod_logs,
        get_workload_status,
        get_node_status,
        get_pod_metrics,
        get_node_metrics,
        discover_prometheus,
        query_prometheus,
    ]
