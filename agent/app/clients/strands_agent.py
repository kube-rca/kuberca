from __future__ import annotations

import time
from collections import OrderedDict
from dataclasses import dataclass
from threading import Lock
from typing import Protocol

from strands import Agent, tool
from strands.models.gemini import GeminiModel
from strands.session import RepositorySessionManager

from app.clients.k8s import KubernetesClient
from app.clients.prometheus import PrometheusClient
from app.clients.session_repository import PostgresSessionRepository
from app.clients.strands_patch import apply_gemini_thought_signature_patch
from app.core.config import Settings


@dataclass
class _AgentCacheEntry:
    agent: Agent
    lock: Lock
    last_access: float


class AnalysisEngine(Protocol):
    def analyze(self, prompt: str, incident_id: str | None = None) -> str:
        raise NotImplementedError


class StrandsAnalysisEngine:
    def __init__(
        self,
        settings: Settings,
        k8s_client: KubernetesClient,
        prometheus_client: PrometheusClient | None = None,
    ) -> None:
        apply_gemini_thought_signature_patch()
        if not settings.session_store_dsn:
            raise ValueError(
                "SESSION_DB_HOST, SESSION_DB_USER, SESSION_DB_NAME are required "
                "for Postgres session storage"
            )
        self._settings = settings
        self._tools = _build_tools(k8s_client, prometheus_client)
        self._cache_lock = Lock()
        self._agent_cache: OrderedDict[str, _AgentCacheEntry] = OrderedDict()
        self._cache_size = max(settings.agent_cache_size, 1)
        self._cache_ttl_seconds = settings.agent_cache_ttl_seconds
        self._session_repo = PostgresSessionRepository(settings.session_store_dsn)

    def analyze(self, prompt: str, incident_id: str | None = None) -> str:
        session_id = self._resolve_session_id(incident_id)
        entry = self._get_cache_entry(session_id)
        with self._session_repo.session_lock(session_id):
            with entry.lock:
                return str(entry.agent(prompt))

    def _resolve_session_id(self, incident_id: str | None) -> str:
        cache_key = incident_id.strip() if incident_id else ""
        return cache_key or "default"

    def _get_cache_entry(self, session_id: str) -> _AgentCacheEntry:
        cache_key = session_id

        with self._cache_lock:
            now = time.time()
            self._evict_expired_entries(now)
            entry = self._agent_cache.get(cache_key)
            if entry is not None:
                entry.last_access = now
                self._agent_cache.move_to_end(cache_key)
                return entry

            entry = _AgentCacheEntry(
                agent=self._build_agent(cache_key),
                lock=Lock(),
                last_access=now,
            )
            self._agent_cache[cache_key] = entry
            if len(self._agent_cache) > self._cache_size:
                self._agent_cache.popitem(last=False)
            return entry

    def _evict_expired_entries(self, now: float) -> None:
        if self._cache_ttl_seconds <= 0:
            return

        expired_keys: list[str] = []
        for key, entry in self._agent_cache.items():
            if now - entry.last_access > self._cache_ttl_seconds:
                expired_keys.append(key)
            else:
                break
        for key in expired_keys:
            self._agent_cache.pop(key, None)

    def _build_agent(self, session_id: str) -> Agent:
        model = GeminiModel(
            client_args={"api_key": self._settings.gemini_api_key},
            model_id=self._settings.gemini_model_id,
        )
        session_manager = RepositorySessionManager(
            session_id=session_id,
            session_repository=self._session_repo,
        )
        return Agent(model=model, tools=self._tools, session_manager=session_manager)


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
        return [snippet.to_dict() for snippet in k8s_client.get_previous_logs(namespace, pod_name)]

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
        """Return the configured Prometheus endpoint (PROMETHEUS_URL)."""
        if prometheus_client is None:
            return {"warning": "prometheus client not configured"}
        return prometheus_client.describe_endpoint()

    @tool
    def list_prometheus_metrics(match: str | None = None) -> dict[str, object]:
        """List available metric names from Prometheus.

        Use this to discover what metrics are available before querying.
        Args:
            match: Optional regex pattern to filter metrics.
                   Examples: 'kube_pod.*', 'istio.*', 'container_.*', 'http_.*'
        """
        if prometheus_client is None:
            return {"warning": "prometheus client not configured"}
        return prometheus_client.list_metrics(match=match)

    @tool
    def query_prometheus(query: str, time: str | None = None) -> dict[str, object]:
        """Run a Prometheus instant query."""
        if prometheus_client is None:
            return {"warning": "prometheus client not configured"}
        return prometheus_client.query(query, time=time)

    @tool
    def list_pods_in_namespace(
        namespace: str, label_selector: str | None = None
    ) -> list[dict[str, object]]:
        """List pods in a namespace with optional label filtering.

        Use this when pod name is not in alert labels to discover affected pods.
        Args:
            namespace: The Kubernetes namespace to list pods from.
            label_selector: Optional label selector (e.g., 'app=nginx', 'version=v1').
        """
        pods = k8s_client.list_pods_in_namespace(namespace, label_selector=label_selector)
        return [pod.to_dict() for pod in pods]

    tools: list[object] = [
        get_pod_status,
        get_pod_spec,
        list_pod_events,
        list_namespace_events,
        list_cluster_events,
        list_pods_in_namespace,
        get_previous_pod_logs,
        get_pod_logs,
        get_workload_status,
        get_node_status,
        get_pod_metrics,
        get_node_metrics,
    ]
    if prometheus_client is not None:
        tools.extend(
            [
                discover_prometheus,
                list_prometheus_metrics,
                query_prometheus,
            ]
        )
    return tools
