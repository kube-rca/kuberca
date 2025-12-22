from __future__ import annotations

import json
import logging
from typing import Any

from app.clients.k8s import KubernetesClient, extract_pod_target
from app.clients.strands_agent import AnalysisEngine
from app.models.k8s import K8sContext
from app.schemas.analysis import AlertAnalysisRequest


class AnalysisService:
    def __init__(
        self,
        k8s_client: KubernetesClient,
        analysis_engine: AnalysisEngine | None,
    ) -> None:
        self._logger = logging.getLogger(__name__)
        self._k8s_client = k8s_client
        self._analysis_engine = analysis_engine

    def analyze(self, request: AlertAnalysisRequest) -> str:
        namespace, pod_name = extract_pod_target(request.alert.labels)
        k8s_context = self._k8s_client.collect_context(namespace, pod_name)

        if self._analysis_engine is None:
            return _fallback_summary(request, k8s_context, "analysis engine not configured")

        prompt = _build_prompt(request, k8s_context)
        try:
            return self._analysis_engine.analyze(prompt)
        except Exception as exc:  # noqa: BLE001
            self._logger.exception("Strands analysis failed")
            return _fallback_summary(request, k8s_context, f"analysis failed: {exc}")


def _build_prompt(request: AlertAnalysisRequest, k8s_context: K8sContext) -> str:
    alert_payload = request.alert.model_dump(by_alias=True, mode="json")
    payload = {
        "alert": alert_payload,
        "thread_ts": request.thread_ts,
        "callback_url": request.callback_url,
    }

    return (
        "You are kube-rca-agent. Analyze the alert using the provided Kubernetes context.\n"
        "Return a concise RCA summary in Korean with: root cause, evidence, and next action.\n"
        "If data is missing, state what is missing.\n"
        "You may call tools (get_pod_status, list_pod_events, get_previous_pod_logs) if needed.\n\n"
        f"Alert payload:\n{_to_pretty_json(payload)}\n\n"
        f"Kubernetes context:\n{_to_pretty_json(k8s_context.to_dict())}\n"
    )


def _fallback_summary(
    request: AlertAnalysisRequest,
    k8s_context: K8sContext,
    reason: str,
) -> str:
    alert = request.alert
    lines = [
        f"analysis engine unavailable: {reason}",
        f"alert_status={alert.status}",
    ]
    if k8s_context.namespace or k8s_context.pod_name:
        lines.append(f"target: namespace={k8s_context.namespace}, pod={k8s_context.pod_name}")
    if k8s_context.warnings:
        lines.append("warnings: " + ", ".join(k8s_context.warnings))
    if k8s_context.pod_status:
        lines.append(f"pod_phase: {k8s_context.pod_status.phase}")
    return "\n".join(lines)


def _to_pretty_json(payload: dict[str, Any]) -> str:
    return json.dumps(payload, ensure_ascii=True, indent=2, sort_keys=True)
