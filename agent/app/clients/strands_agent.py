from __future__ import annotations

from typing import Protocol

from strands import Agent, tool
from strands.models.gemini import GeminiModel

from app.clients.k8s import KubernetesClient
from app.core.config import Settings


class AnalysisEngine(Protocol):
    def analyze(self, prompt: str) -> str:
        raise NotImplementedError


class StrandsAnalysisEngine:
    def __init__(self, settings: Settings, k8s_client: KubernetesClient) -> None:
        model = GeminiModel(
            client_args={"api_key": settings.gemini_api_key},
            model_id=settings.gemini_model_id,
        )
        self._agent = Agent(model=model, tools=_build_tools(k8s_client))

    def analyze(self, prompt: str) -> str:
        return str(self._agent(prompt))


def _build_tools(k8s_client: KubernetesClient) -> list[object]:
    @tool
    def get_pod_status(namespace: str, pod_name: str) -> dict[str, object]:
        """Fetch pod status details."""
        status = k8s_client.get_pod_status(namespace, pod_name)
        if status is None:
            return {"warning": "pod status not found"}
        return status.to_dict()

    @tool
    def list_pod_events(namespace: str, pod_name: str) -> list[dict[str, object]]:
        """List recent events for a pod."""
        return [event.to_dict() for event in k8s_client.list_pod_events(namespace, pod_name)]

    @tool
    def get_previous_pod_logs(
        namespace: str,
        pod_name: str,
    ) -> list[dict[str, object]]:
        """Return previous container logs (tail)."""
        return [
            snippet.to_dict() for snippet in k8s_client.get_previous_logs(namespace, pod_name)
        ]

    return [get_pod_status, list_pod_events, get_previous_pod_logs]
