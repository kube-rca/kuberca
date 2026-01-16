from __future__ import annotations

import json
import logging
from typing import Any

from app.clients.k8s import KubernetesClient, extract_pod_target
from app.clients.strands_agent import AnalysisEngine
from app.models.k8s import K8sContext
from app.schemas.analysis import AlertAnalysisRequest, IncidentSummaryRequest


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
            return self._analysis_engine.analyze(prompt, request.incident_id)
        except Exception as exc:  # noqa: BLE001
            self._logger.exception("Strands analysis failed")
            return _fallback_summary(request, k8s_context, f"analysis failed: {exc}")

    def summarize_incident(self, request: IncidentSummaryRequest) -> tuple[str, str]:
        """Synthesize final RCA summary for a resolved incident.

        Returns:
            tuple[str, str]: (summary, detail) - summary is 1-2 sentences, detail is full analysis
        """
        if self._analysis_engine is None:
            return _fallback_incident_summary(request, "analysis engine not configured")

        prompt = _build_incident_summary_prompt(request)
        try:
            result = self._analysis_engine.analyze(prompt, request.incident_id)
            return _parse_incident_summary(result)
        except Exception as exc:  # noqa: BLE001
            self._logger.exception("Incident summary analysis failed")
            return _fallback_incident_summary(request, f"analysis failed: {exc}")


def _build_prompt(request: AlertAnalysisRequest, k8s_context: K8sContext) -> str:
    alert_payload = request.alert.model_dump(by_alias=True, mode="json")
    payload = {
        "alert": alert_payload,
        "thread_ts": request.thread_ts,
    }
    if request.incident_id:
        payload["incident_id"] = request.incident_id

    return (
        "You are kube-rca-agent. Analyze the alert using the provided Kubernetes context.\n"
        "Return a concise RCA summary in Korean with: root cause, evidence, and next action.\n"
        "If data is missing, state what is missing.\n"
        "You may call tools if needed:\n"
        "- get_pod_status, get_pod_spec\n"
        "- list_pod_events, list_namespace_events, list_cluster_events\n"
        "- get_previous_pod_logs, get_pod_logs\n"
        "- get_workload_status, get_node_status\n"
        "- get_pod_metrics, get_node_metrics\n"
        "- discover_prometheus, query_prometheus\n\n"
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


def _fallback_incident_summary(request: IncidentSummaryRequest, reason: str) -> tuple[str, str]:
    """Generate fallback summary when analysis engine is unavailable."""
    alert_names = [a.alert_name for a in request.alerts]
    summary = f"인시던트 분석 불가: {reason}"
    detail_lines = [
        f"인시던트 ID: {request.incident_id}",
        f"제목: {request.title}",
        f"심각도: {request.severity}",
        f"발생 시각: {request.fired_at}",
        f"해결 시각: {request.resolved_at}",
        f"관련 알림 ({len(request.alerts)}개): {', '.join(alert_names)}",
        "",
        f"분석 엔진 오류: {reason}",
    ]
    return summary, "\n".join(detail_lines)


def _parse_incident_summary(result: str) -> tuple[str, str]:
    """Parse AI response into summary and detail.

    The AI is instructed to return structured content with 요약 and 상세 분석 sections.
    We extract the summary (first meaningful paragraph) and use the full result as detail.
    """
    lines = result.strip().split("\n")

    # Try to extract summary from the response
    summary = ""
    for line in lines:
        line = line.strip()
        # Skip empty lines and headers
        if not line or line.startswith("#") or line.startswith("**"):
            continue
        # First non-empty, non-header line is likely the summary
        if "요약" in line.lower() or "summary" in line.lower():
            continue
        summary = line[:200]  # Limit summary length
        break

    if not summary:
        # Fallback: use first 200 chars of result
        summary = result[:200].replace("\n", " ").strip()
        if len(result) > 200:
            summary += "..."

    return summary, result


def _build_incident_summary_prompt(request: IncidentSummaryRequest) -> str:
    alerts_info = []
    for alert in request.alerts:
        alert_data = {
            "fingerprint": alert.fingerprint,
            "alert_name": alert.alert_name,
            "severity": alert.severity,
            "status": alert.status,
            "analysis_summary": alert.analysis_summary or "N/A",
            "analysis_detail": alert.analysis_detail or "N/A",
        }
        alerts_info.append(alert_data)

    incident_data = {
        "incident_id": request.incident_id,
        "title": request.title,
        "severity": request.severity,
        "fired_at": request.fired_at,
        "resolved_at": request.resolved_at,
        "alert_count": len(request.alerts),
        "alerts": alerts_info,
    }

    return (
        "You are kube-rca-agent. An incident has been resolved and you need to provide a final RCA summary.\n"
        "Analyze all the alerts and their individual analyses to synthesize a comprehensive incident summary.\n\n"
        "Return your response in Korean with the following structure:\n"
        "1. **요약 (Summary)**: 1-2 sentences describing the root cause and resolution\n"
        "2. **상세 분석 (Detail)**:\n"
        "   - 근본 원인 (Root Cause)\n"
        "   - 영향 범위 (Impact)\n"
        "   - 해결 과정 (Resolution)\n"
        "   - 재발 방지 권고 (Prevention Recommendations)\n\n"
        f"Incident data:\n{_to_pretty_json(incident_data)}\n"
    )
