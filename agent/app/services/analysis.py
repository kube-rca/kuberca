from __future__ import annotations

import json
import logging
import re
import time
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from typing import Any, cast
from uuid import uuid4

from app.clients.k8s import KubernetesClient, resolve_alert_target
from app.clients.strands_agent import AnalysisEngine
from app.clients.summary_store import SummaryStore
from app.clients.tempo import TempoClient, build_traceql_query
from app.core.masking import Masker, RegexMasker
from app.models.k8s import AnalysisTarget, K8sContext
from app.schemas.analysis import AlertAnalysisRequest, IncidentSummaryRequest


class AnalysisService:
    def __init__(
        self,
        k8s_client: KubernetesClient,
        analysis_engine: AnalysisEngine | None,
        masker: Masker | None = None,
        prometheus_enabled: bool = False,
        loki_enabled: bool = False,
        tempo_client: TempoClient | None = None,
        tempo_enabled: bool = False,
        tempo_trace_limit: int = 5,
        tempo_lookback_minutes: int = 15,
        tempo_forward_minutes: int = 5,
        summary_store: SummaryStore | None = None,
        summary_history_size: int = 3,
        prompt_token_budget: int = 32000,
        prompt_max_log_lines: int = 25,
        prompt_max_events: int = 25,
    ) -> None:
        self._logger = logging.getLogger(__name__)
        self._k8s_client = k8s_client
        self._analysis_engine = analysis_engine
        self._masker = masker or RegexMasker()
        self._prometheus_enabled = prometheus_enabled
        self._loki_enabled = loki_enabled
        self._tempo_client = tempo_client
        self._tempo_enabled = tempo_enabled and tempo_client is not None
        self._tempo_trace_limit = max(1, tempo_trace_limit)
        self._tempo_lookback_minutes = max(0, tempo_lookback_minutes)
        self._tempo_forward_minutes = max(0, tempo_forward_minutes)
        self._summary_store = summary_store
        self._summary_history_size = max(1, summary_history_size)
        self._prompt_token_budget = max(0, prompt_token_budget)
        self._prompt_max_log_lines = max(0, prompt_max_log_lines)
        self._prompt_max_events = max(0, prompt_max_events)

    def analyze(
        self, request: AlertAnalysisRequest
    ) -> tuple[str, str, str, dict[str, object], list[dict[str, object]]]:
        t_start = time.perf_counter()

        target = resolve_alert_target(request.alert.labels)
        t_resolve = time.perf_counter()

        k8s_context = self._k8s_client.collect_context(
            target.namespace,
            target.pod_name,
            target.workload,
            service_name=target.service_name,
        )
        t_k8s = time.perf_counter()

        tempo_context = self._collect_tempo_context(request, target)
        t_tempo = time.perf_counter()

        artifacts = _build_alert_artifacts(k8s_context, tempo_context)
        masked_artifacts = cast(list[dict[str, object]], self._masker.mask_object(artifacts))
        capabilities, capability_warnings = self._collect_capabilities(
            k8s_context=k8s_context,
            tempo_context=tempo_context,
        )
        base_missing_data = _collect_missing_data(
            k8s_context=k8s_context,
            tempo_context=tempo_context,
            capabilities=capabilities,
        )
        base_warnings = _collect_analysis_warnings(
            k8s_warnings=k8s_context.warnings,
            tempo_context=tempo_context,
            capability_warnings=capability_warnings,
        )

        def build_masked_context(engine_issue: str | None = None) -> dict[str, object]:
            missing_data = list(base_missing_data)
            if engine_issue:
                missing_data.append(f"analysis_engine.{engine_issue}")
            missing_data = _dedupe_strings(missing_data)

            warnings = list(base_warnings)
            if engine_issue:
                warnings.append(f"analysis engine issue: {engine_issue}")
            warnings = _dedupe_strings(warnings)

            analysis_quality = _resolve_analysis_quality(
                k8s_context=k8s_context,
                missing_data=missing_data,
                capabilities=capabilities,
                engine_issue=engine_issue,
            )
            context = k8s_context.to_dict()
            if tempo_context:
                context["tempo"] = tempo_context
            context["analysis_quality"] = analysis_quality
            context["missing_data"] = missing_data
            context["warnings"] = warnings
            context["capabilities"] = capabilities
            return cast(dict[str, object], self._masker.mask_object(context))

        if self._analysis_engine is None:
            analysis = self._masker.mask_text(
                _fallback_summary(request, k8s_context, "analysis engine not configured")
            )
            summary, detail = _split_alert_analysis(analysis)
            masked_context = build_masked_context(engine_issue="not_configured")
            return analysis, summary, detail, masked_context, masked_artifacts

        summary_key = _resolve_alert_session_id(request)
        recent_summaries = self._load_recent_summaries(summary_key)

        # Resolved 분석 시 컨텍스트 축소 (이전 분석이 이미 상세 분석을 수행)
        analysis_type = request.analysis_type or request.alert.status
        effective_max_log_lines = self._prompt_max_log_lines
        effective_max_events = self._prompt_max_events
        if analysis_type == "resolved" and request.previous_analysis is not None:
            effective_max_log_lines = max(1, self._prompt_max_log_lines // 2)
            effective_max_events = max(1, self._prompt_max_events // 2)

        prompt = _build_prompt(
            request,
            k8s_context,
            self._prometheus_enabled,
            self._loki_enabled,
            self._tempo_enabled,
            tempo_context,
            capabilities,
            base_missing_data,
            base_warnings,
            recent_summaries,
            self._prompt_token_budget,
            effective_max_log_lines,
            effective_max_events,
            self._masker,
        )
        t_prompt = time.perf_counter()

        try:
            session_id = _build_runtime_session_id(summary_key)
            analysis = self._analysis_engine.analyze(prompt, session_id)
            t_llm = time.perf_counter()
            if not isinstance(analysis, str):
                analysis = ""
            analysis = self._masker.mask_text(analysis)
            if not analysis.strip():
                self._logger.warning("Strands analysis returned empty response")
                analysis = self._masker.mask_text(
                    _fallback_summary(
                        request,
                        k8s_context,
                        "analysis engine returned empty response",
                    )
                )
                summary, detail = _split_alert_analysis(analysis)
                masked_context = build_masked_context(engine_issue="empty_response")
                self._log_analysis_timing(
                    t_start,
                    t_resolve,
                    t_k8s,
                    t_tempo,
                    t_prompt,
                    t_llm,
                )
                return analysis, summary, detail, masked_context, masked_artifacts
            summary, detail = _split_alert_analysis(analysis)
            self._store_summary(summary_key, summary)
            masked_context = build_masked_context()
            self._log_analysis_timing(
                t_start,
                t_resolve,
                t_k8s,
                t_tempo,
                t_prompt,
                t_llm,
            )
            return analysis, summary, detail, masked_context, masked_artifacts
        except Exception as exc:  # noqa: BLE001
            t_llm = time.perf_counter()
            error_cat = _categorize_analysis_error(exc)
            self._logger.exception(
                "Strands analysis failed: category=%s exc_type=%s exc_repr=%r",
                error_cat.name,
                type(exc).__name__,
                exc,
            )
            analysis = self._masker.mask_text(
                _fallback_summary(request, k8s_context, error_cat.user_message)
            )
            summary, detail = _split_alert_analysis(analysis)
            masked_context = build_masked_context(engine_issue=error_cat.name)
            self._log_analysis_timing(
                t_start,
                t_resolve,
                t_k8s,
                t_tempo,
                t_prompt,
                t_llm,
            )
            return analysis, summary, detail, masked_context, masked_artifacts

    def _log_analysis_timing(
        self,
        t_start: float,
        t_resolve: float,
        t_k8s: float,
        t_tempo: float,
        t_prompt: float,
        t_llm: float,
    ) -> None:
        pre_llm_ms = (t_prompt - t_start) * 1000
        total_ms = (t_llm - t_start) * 1000
        self._logger.info(
            "analysis_timing resolve_ms=%.1f k8s_ms=%.1f tempo_ms=%.1f "
            "prompt_build_ms=%.1f llm_ms=%.1f pre_llm_ms=%.1f total_ms=%.1f",
            (t_resolve - t_start) * 1000,
            (t_k8s - t_resolve) * 1000,
            (t_tempo - t_k8s) * 1000,
            (t_prompt - t_tempo) * 1000,
            (t_llm - t_prompt) * 1000,
            pre_llm_ms,
            total_ms,
        )

    def summarize_incident(self, request: IncidentSummaryRequest) -> tuple[str, str, str]:
        """Synthesize final RCA summary for a resolved incident.

        Returns:
            tuple[str, str, str]: (title, summary, detail)
        """
        if self._analysis_engine is None:
            return self._mask_incident_result(
                _fallback_incident_summary(request, "analysis engine not configured")
            )

        prompt = _build_incident_summary_prompt(request, self._masker)
        try:
            session_id = _resolve_summary_session_id(request)
            result = self._analysis_engine.analyze(prompt, session_id)
            if not isinstance(result, str):
                result = ""
            masked_result = self._masker.mask_text(result)
            return self._mask_incident_result(_parse_incident_summary(masked_result, request.title))
        except Exception as exc:  # noqa: BLE001
            self._logger.exception("Incident summary analysis failed")
            return self._mask_incident_result(
                _fallback_incident_summary(request, f"analysis failed: {exc}")
            )

    def _mask_incident_result(self, result: tuple[str, str, str]) -> tuple[str, str, str]:
        title, summary, detail = result
        return (
            self._masker.mask_text(title),
            self._masker.mask_text(summary),
            self._masker.mask_text(detail),
        )

    def _load_recent_summaries(self, session_id: str) -> list[str]:
        if self._summary_store is None or self._summary_history_size <= 0:
            return []
        try:
            summaries = self._summary_store.list_summaries(session_id, self._summary_history_size)
            return [self._masker.mask_text(summary) for summary in summaries]
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to load session summaries: %s", exc)
            return []

    def _store_summary(self, session_id: str, summary: str) -> None:
        if self._summary_store is None or self._summary_history_size <= 0:
            return
        masked_summary = self._masker.mask_text(summary)
        compact = _compact_summary(masked_summary, limit=300) or masked_summary.strip()
        if not compact:
            return
        try:
            self._summary_store.append_summary(
                session_id,
                compact,
                max_items=self._summary_history_size,
            )
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to store session summary: %s", exc)

    def _collect_tempo_context(
        self,
        request: AlertAnalysisRequest,
        target: AnalysisTarget,
    ) -> dict[str, object] | None:
        if not self._tempo_enabled or self._tempo_client is None:
            return None

        namespace = target.namespace
        service_name = target.service_name
        analysis_type = request.analysis_type or request.alert.status
        start, end = _resolve_tempo_window(
            request.alert.starts_at,
            ends_at=request.alert.ends_at,
            analysis_type=analysis_type,
            lookback_minutes=self._tempo_lookback_minutes,
            forward_minutes=self._tempo_forward_minutes,
        )
        query = build_traceql_query(service_name=service_name, namespace=namespace)
        response = self._tempo_client.search_traces(
            query=query,
            start=start,
            end=end,
            limit=self._tempo_trace_limit,
        )
        traces = _extract_tempo_traces(response)
        trace_count = response.get("trace_count")
        if not isinstance(trace_count, int):
            trace_count = len(traces)

        warnings: list[str] = []
        if isinstance(response.get("warning"), str):
            warnings.append(response["warning"])
        if isinstance(response.get("error"), str):
            warnings.append(response["error"])
        detail = response.get("detail")
        if isinstance(detail, dict):
            status_code = detail.get("status_code")
            reason = detail.get("reason")
            if status_code is not None or reason is not None:
                warnings.append(f"tempo_query_detail: status={status_code}, reason={reason}")

        query_status = "error" if isinstance(response.get("error"), str) else "ok"

        return {
            "endpoint": response.get("endpoint"),
            "query": query,
            "window": {"start": start, "end": end},
            "service_name": service_name,
            "namespace": namespace,
            "query_status": query_status,
            "trace_count": trace_count,
            "traces": traces[: self._tempo_trace_limit],
            "warnings": warnings,
        }

    def _collect_capabilities(
        self,
        *,
        k8s_context: K8sContext,
        tempo_context: dict[str, object] | None,
    ) -> tuple[dict[str, str], list[str]]:
        capabilities: dict[str, str] = {
            "k8s_core": "ok",
            "manifest_read": "ok",
            "prometheus": "ok" if self._prometheus_enabled else "unavailable",
            "loki": "ok" if self._loki_enabled else "unavailable",
            "tempo": "ok" if self._tempo_enabled else "unavailable",
            "mesh_type": _resolve_mesh_type(k8s_context),
            "routing_evidence": "unavailable",
        }
        warnings: list[str] = []
        if any(
            "kubernetes client is not configured" in warning for warning in k8s_context.warnings
        ):
            capabilities["k8s_core"] = "unavailable"
            capabilities["manifest_read"] = "unavailable"
            capabilities["mesh_type"] = "unknown"
            capabilities["routing_evidence"] = "unavailable"
            warnings.append("kubernetes core api unavailable")
        elif capabilities["mesh_type"] == "istio":
            capabilities["routing_evidence"] = "manifest_only"
        elif capabilities["mesh_type"] == "none":
            capabilities["routing_evidence"] = "not_applicable"
        if tempo_context and tempo_context.get("query_status") == "error":
            capabilities["tempo"] = "degraded"

        return capabilities, warnings


def _build_prompt(
    request: AlertAnalysisRequest,
    k8s_context: K8sContext,
    prometheus_enabled: bool,
    loki_enabled: bool,
    tempo_enabled: bool,
    tempo_context: dict[str, object] | None,
    capabilities: dict[str, str],
    missing_data: list[str],
    diagnostic_warnings: list[str],
    recent_summaries: list[str],
    prompt_token_budget: int,
    prompt_max_log_lines: int,
    prompt_max_events: int,
    masker: Masker,
) -> str:
    alert_payload = cast(
        dict[str, Any],
        masker.mask_object(request.alert.model_dump(by_alias=True, mode="json")),
    )
    payload = {
        "alert": alert_payload,
        "thread_ts": masker.mask_text(request.thread_ts),
    }
    if request.incident_id:
        payload["incident_id"] = masker.mask_text(request.incident_id)

    analysis_type = request.analysis_type or request.alert.status
    mesh_type = capabilities.get("mesh_type", "unknown")

    tool_lines = [
        "- get_pod_status, get_pod_spec",
        "- list_pod_events, list_namespace_events, list_cluster_events",
        "- list_pods_in_namespace (use when pod name is missing from alert labels)",
        "- get_previous_pod_logs, get_pod_logs",
        "- get_workload_status, get_daemonset_manifest, get_node_status",
        "- get_pod_metrics, get_node_metrics",
        "- get_service, get_endpoints",
        "- get_manifest, list_manifests",
    ]
    if prometheus_enabled:
        tool_lines.append("- discover_prometheus, list_prometheus_metrics")
        tool_lines.append("- query_prometheus, query_prometheus_range")
    if loki_enabled:
        tool_lines.append("- discover_loki, list_loki_labels, get_loki_label_values")
        tool_lines.append("- query_loki, query_loki_range")
    if tempo_enabled:
        tool_lines.append("- discover_tempo, search_tempo_traces, get_tempo_trace")
    if mesh_type == "istio":
        tool_lines.append("- list_virtual_services, list_destination_rules, list_service_entries")
    tool_block = "\n".join(tool_lines)
    policy_block = (
        "Analysis policy:\n"
        "- Prefer built-in tools over manual kubectl/istioctl instructions.\n"
        "- Do not ask the user to run kubectl, istioctl, or equivalent commands "
        "when available tools can fetch the same evidence.\n"
        "- Do not mention Prometheus, Loki, Tempo, or Istio as action items "
        "when the corresponding capability is unavailable.\n"
        "- If direct evidence is incomplete, label the conclusion as a "
        "hypothesis and explain the confidence gap.\n"
        "- If mesh_type is 'none', do not mention Istio resources or mesh routing.\n"
        "- If mesh_type is 'istio', treat routing evidence as manifest-only "
        "and do not claim live proxy state.\n\n"
    )

    summary_block = _format_session_summaries(
        [masker.mask_text(summary) for summary in recent_summaries]
    )
    if analysis_type == "resolved" and request.previous_analysis is not None:
        prev = request.previous_analysis
        prompt = (
            "You are kube-rca-agent. This is a RESOLVED alert.\n"
            "Your goal is NOT to repeat the root cause analysis from the firing phase.\n"
            "Focus on recovery confirmation and post-incident insights.\n\n"
            "Return your response in Korean with the following structure:\n"
            "1) 요약 (Summary): Recovery confirmation + key metric changes.\n"
            "2) 상세 분석 (Detail):\n"
            "   #### 복구 확인 (Recovery Confirmation)\n"
            "   #### 장애 영향 (Impact Assessment) - 장애 지속 시간, 영향 범위\n"
            "   #### 이전 분석 대비 변화 (Delta Analysis)\n"
            "   #### 재발 방지 권고 (Prevention Recommendations)\n"
            "Formatting rules:\n"
            "- Use markdown headers: '### 1) 요약 (Summary)' and '### 2) 상세 분석 (Detail)'.\n"
            "- Use '####' for subsections.\n"
            "- Leave one blank line between sections/subsections.\n"
            "- Use '-' for unordered lists (do not use '*').\n"
            "- Limit each subsection to 3-5 bullets; one sentence per bullet (<= 120 chars).\n"
            "- Use inline code only for literal keys/values/commands.\n"
            f"{policy_block}"
            "You may call tools to verify recovery:\n"
            f"{tool_block}\n\n"
            "Previous firing analysis (DO NOT repeat this content):\n"
            f"Summary: {masker.mask_text(prev.summary)}\n"
            f"Detail: {masker.mask_text(prev.detail)}\n\n"
        )
    else:
        prompt = (
            "You are kube-rca-agent. Analyze the alert using the provided Kubernetes context.\n"
            "Your audience is a human operator reading this on Slack.\n"
            "Return your response in Korean with the following structure:\n"
            "1) 요약 (Summary): 3-5 sentences. Include root cause + impact + next action.\n"
            "2) 상세 분석 (Detail): Use sections for "
            "근본 원인, 확인 근거, 조치 사항, 누락된 데이터.\n"
            "\n"
            "Analysis behavior:\n"
            "- ACTIVELY use tools to discover information. "
            "If alert labels are missing (namespace, pod, service), "
            "use available tools (e.g. list_pods_in_namespace, list_namespace_events, "
            "get_pod_status) to find the affected resources yourself.\n"
            "- Your analysis MUST contain completed findings based on evidence "
            "you gathered using tools.\n"
            "- 조치 사항 must be actionable operator recommendations "
            "(e.g., '메모리 limit을 512Mi로 상향 조정하십시오').\n"
            "- 누락된 데이터 lists ONLY data you could NOT obtain even after using tools. "
            "Do NOT list data that is expectedly absent "
            "(e.g., previous_logs when restart_count is 0, "
            "or service info when no service label exists).\n"
            "- Prefer direct evidence from logs, events, workload state, "
            "Service, and Endpoints before adding manual follow-up actions.\n"
            "- Do not infer routing behavior or external dependency failures "
            "without direct evidence.\n"
            "\n"
            "Formatting rules:\n"
            "- Use markdown headers: '### 1) 요약 (Summary)' and '### 2) 상세 분석 (Detail)'.\n"
            "- Each subsection MUST start with a bold '####' markdown header exactly as shown:\n"
            "  #### **근본 원인**\n"
            "  #### **확인 근거**\n"
            "  #### **조치 사항**\n"
            "  #### **누락된 데이터**\n"
            "- Leave one blank line between sections/subsections.\n"
            "- Use '-' for unordered lists (do not use '*').\n"
            "- Limit each subsection to 3-5 bullets; one sentence per bullet (<= 120 chars).\n"
            "- Use inline code only for literal keys/values/commands; "
            "avoid excessive code formatting.\n"
            f"{policy_block}"
            "Use these tools to gather evidence before writing your analysis:\n"
            f"{tool_block}\n\n"
        )

    if prometheus_enabled:
        prompt += (
            "For Prometheus queries:\n"
            "1. Use list_prometheus_metrics(match='pattern') to discover available metrics.\n"
            "2. Use query_prometheus(query) for current/instant values.\n"
            "3. Use query_prometheus_range(query, start, end, step) for time-series history.\n"
            "   - ALWAYS use range queries to understand metric trends before the alert.\n"
            "   - Use cases: memory/CPU spikes, error rate increase, latency degradation,\n"
            "     request volume changes, network issues, resource exhaustion, etc.\n"
            "   - Use alert's startsAt to calculate start (e.g., 1h before) and end time.\n"
            "   - Example: query_prometheus_range(\n"
            "       query='rate(http_requests_total{pod=\"my-pod\"}[5m])',\n"
            "       start='<startsAt - 1h>', end='<startsAt>', step='1m')\n"
            "Example patterns: 'container_memory.*', 'container_cpu.*', 'http_.*',\n"
            "'istio_request.*', 'kube_pod.*', 'node_.*'\n\n"
        )

    if loki_enabled:
        prompt += (
            "For Loki queries:\n"
            "1. Use list_loki_labels() and get_loki_label_values(label) to "
            "discover labels before building LogQL.\n"
            "2. Use query_loki_range(query, start, end, limit, step) for "
            "incident-time historical logs.\n"
            "3. Use query_loki(query, limit, time) for point-in-time log checks.\n"
            "4. Before claiming that detailed logs are missing, check whether "
            "Loki is available and query the incident window.\n\n"
        )

    if tempo_enabled:
        prompt += (
            "For Tempo trace queries:\n"
            "1. Use search_tempo_traces(start, end, service_name, namespace, query, limit).\n"
            "2. Use get_tempo_trace(trace_id) to inspect spans for a selected trace.\n"
            "3. Use alert's startsAt to search around the incident time window.\n"
            "4. Prioritize failed spans and high-latency path evidence.\n"
            "5. If tempo query has warnings/errors, treat as query failure, not no-data.\n\n"
        )

    if mesh_type == "istio":
        prompt += (
            "For Istio routing evidence:\n"
            "1. Use get_service(namespace, name) and get_endpoints(namespace, name) first.\n"
            "2. Use list_virtual_services(), list_destination_rules(), and "
            "list_service_entries() for manifest evidence.\n"
            "3. Treat routing evidence as desired-state configuration only; "
            "do not claim live Envoy behavior.\n\n"
        )

    if summary_block:
        prompt += summary_block

    context_dict = _prepare_k8s_context(
        k8s_context,
        max_events=prompt_max_events,
        max_log_lines=prompt_max_log_lines,
    )
    if tempo_context:
        context_dict["tempo"] = _compact_tempo_context(tempo_context)
    context_dict["capabilities"] = capabilities
    context_dict["missing_data"] = missing_data
    context_dict["diagnostic_warnings"] = diagnostic_warnings
    context_dict = cast(dict[str, object], masker.mask_object(context_dict))
    alert_block = f"Alert payload:\n{_to_pretty_json(payload)}\n\n"
    context_block = f"Kubernetes/APM context:\n{_to_pretty_json(context_dict)}\n"
    full_prompt = prompt + alert_block + context_block
    return _apply_prompt_budget(
        full_prompt,
        prompt_prefix=prompt,
        alert_block=alert_block,
        context_dict=context_dict,
        prompt_token_budget=prompt_token_budget,
    )


def _build_runtime_session_id(summary_key: str) -> str:
    suffix = uuid4().hex[:8]
    if summary_key:
        return f"{summary_key}:run:{suffix}"
    return f"run:{suffix}"


def _format_session_summaries(summaries: list[str]) -> str:
    if not summaries:
        return ""
    lines = ["Recent session summaries (latest 3):"]
    for idx, summary in enumerate(summaries, start=1):
        compact = _compact_summary(summary, limit=300)
        if not compact:
            continue
        lines.append(f"{idx}) {compact}")
    if len(lines) == 1:
        return ""
    return "\n".join(lines) + "\n\n"


def _prepare_k8s_context(
    k8s_context: K8sContext, *, max_events: int, max_log_lines: int
) -> dict[str, object]:
    context = k8s_context.to_dict()
    events = context.get("events") or []
    if max_events <= 0:
        context["events"] = []
    else:
        context["events"] = events[:max_events]

    if max_log_lines <= 0:
        context["current_logs"] = []
        context["previous_logs"] = []
        return context

    context["current_logs"] = _trim_log_context(
        context.get("current_logs") or [],
        max_log_lines=max_log_lines,
    )
    context["previous_logs"] = _trim_log_context(
        context.get("previous_logs") or [],
        max_log_lines=max_log_lines,
    )
    return context


def _trim_log_context(raw_snippets: list[object], *, max_log_lines: int) -> list[dict[str, object]]:
    trimmed_logs: list[dict[str, object]] = []
    for snippet in raw_snippets:
        if not isinstance(snippet, dict):
            continue
        lines = snippet.get("logs") or []
        if not isinstance(lines, list):
            lines = []
        normalized = [line for line in lines if isinstance(line, str)]
        normalized = _dedupe_consecutive_lines(normalized)
        normalized = normalized[-max_log_lines:]
        trimmed = dict(snippet)
        trimmed["logs"] = normalized
        trimmed_logs.append(trimmed)
    return trimmed_logs


def _apply_prompt_budget(
    prompt: str,
    *,
    prompt_prefix: str,
    alert_block: str,
    context_dict: dict[str, object],
    prompt_token_budget: int,
) -> str:
    budget_chars = _prompt_budget_to_chars(prompt_token_budget)
    if budget_chars <= 0 or len(prompt) <= budget_chars:
        return prompt

    def build_context_block(
        context: dict[str, object] | None,
        note: str | None = None,
    ) -> str:
        if context is None:
            reason = note or "omitted due to prompt budget"
            return f"Kubernetes/APM context: {reason}.\n"
        header = "Kubernetes/APM context"
        if note:
            header = f"{header} ({note})"
        return f"{header}:\n{_to_pretty_json(context)}\n"

    # Step 1: remove optional trace context first
    if context_dict.get("tempo"):
        trimmed = dict(context_dict)
        trimmed.pop("tempo", None)
        candidate = prompt_prefix + alert_block + build_context_block(trimmed, "tempo omitted")
        if len(candidate) <= budget_chars:
            return candidate

    # Step 2: remove previous logs
    if context_dict.get("previous_logs"):
        trimmed = dict(context_dict)
        trimmed["previous_logs"] = []
        candidate = prompt_prefix + alert_block + build_context_block(trimmed, "logs omitted")
        if len(candidate) <= budget_chars:
            return candidate

    # Step 2b: remove all logs (current + previous)
    if context_dict.get("current_logs") or context_dict.get("previous_logs"):
        trimmed = dict(context_dict)
        trimmed["current_logs"] = []
        trimmed["previous_logs"] = []
        candidate = prompt_prefix + alert_block + build_context_block(trimmed, "all logs omitted")
        if len(candidate) <= budget_chars:
            return candidate

    # Step 3: remove events + all logs
    if context_dict.get("events"):
        trimmed = dict(context_dict)
        trimmed["events"] = []
        trimmed["current_logs"] = []
        trimmed["previous_logs"] = []
        candidate = (
            prompt_prefix + alert_block + build_context_block(trimmed, "events/logs omitted")
        )
        if len(candidate) <= budget_chars:
            return candidate

    # Step 4: compact context
    compact = _compact_context_dict(context_dict)
    candidate = prompt_prefix + alert_block + build_context_block(compact, "compact")
    if len(candidate) <= budget_chars:
        return candidate

    # Step 5: omit context entirely
    return prompt_prefix + alert_block + build_context_block(None, "omitted due to prompt budget")


def _compact_context_dict(context: dict[str, object]) -> dict[str, object]:
    pod_status = context.get("pod_status") or {}
    compact_status = None
    if pod_status:
        compact_status = {
            "phase": pod_status.get("phase"),
            "reason": pod_status.get("reason"),
            "message": pod_status.get("message"),
            "node_name": pod_status.get("node_name"),
        }
    compact_tempo = None
    tempo = context.get("tempo")
    if isinstance(tempo, dict):
        compact_tempo = _compact_tempo_context(tempo)
    return {
        "target": context.get("target"),
        "namespace": context.get("namespace"),
        "pod_name": context.get("pod_name"),
        "workload": context.get("workload"),
        "service_name": context.get("service_name"),
        "pod_status": compact_status,
        "current_logs": _compact_log_snippets(context.get("current_logs")),
        "tempo": compact_tempo,
        "warnings": context.get("warnings") or [],
        "capabilities": context.get("capabilities") or {},
        "missing_data": context.get("missing_data") or [],
        "diagnostic_warnings": context.get("diagnostic_warnings") or [],
    }


def _dedupe_consecutive_lines(lines: list[str]) -> list[str]:
    deduped: list[str] = []
    last = None
    for line in lines:
        if line == last:
            continue
        deduped.append(line)
        last = line
    return deduped


def _compact_log_snippets(
    raw_snippets: object, *, limit_snippets: int = 1
) -> list[dict[str, object]]:
    if not isinstance(raw_snippets, list):
        return []
    compact: list[dict[str, object]] = []
    for snippet in raw_snippets[:limit_snippets]:
        if not isinstance(snippet, dict):
            continue
        logs = snippet.get("logs")
        if not isinstance(logs, list):
            logs = []
        lines = [line for line in logs if isinstance(line, str)]
        lines = _dedupe_consecutive_lines(lines)[-5:]
        compact.append(
            {
                "container": snippet.get("container"),
                "previous": snippet.get("previous"),
                "error": snippet.get("error"),
                "logs": lines,
            }
        )
    return compact


def _compact_pod_spec(raw_pod_spec: object) -> dict[str, object] | None:
    if not isinstance(raw_pod_spec, dict):
        return None
    return {
        "service_account": raw_pod_spec.get("service_account"),
        "restart_policy": raw_pod_spec.get("restart_policy"),
        "containers": _compact_container_specs(raw_pod_spec.get("containers")),
        "init_containers": _compact_container_specs(raw_pod_spec.get("init_containers")),
    }


def _compact_container_specs(raw_containers: object) -> list[dict[str, object]]:
    if not isinstance(raw_containers, list):
        return []
    compact: list[dict[str, object]] = []
    for container in raw_containers[:4]:
        if not isinstance(container, dict):
            continue
        compact.append(
            {
                "name": container.get("name"),
                "image": container.get("image"),
                "liveness_probe": container.get("liveness_probe") is not None,
                "readiness_probe": container.get("readiness_probe") is not None,
                "startup_probe": container.get("startup_probe") is not None,
            }
        )
    return compact


def _compact_manifest_metadata(raw_metadata: object) -> dict[str, object] | None:
    if not isinstance(raw_metadata, dict):
        return None
    return {
        "name": raw_metadata.get("name"),
        "namespace": raw_metadata.get("namespace"),
        "labels": raw_metadata.get("labels"),
    }


def _prompt_budget_to_chars(prompt_token_budget: int) -> int:
    if prompt_token_budget <= 0:
        return 0
    return prompt_token_budget * 4


def _resolve_alert_session_id(request: AlertAnalysisRequest) -> str:
    incident_id = _normalize_session_token(request.incident_id)
    fingerprint = _normalize_session_token(request.alert.fingerprint)
    if incident_id and fingerprint:
        return f"{incident_id}:{fingerprint}"
    if fingerprint:
        return f"alert:{fingerprint}"
    fallback = _build_alert_fallback_key(request)
    if incident_id and fallback:
        return f"{incident_id}:{fallback}"
    if fallback:
        return f"alert:{fallback}"
    return incident_id or "default"


def _resolve_summary_session_id(request: IncidentSummaryRequest) -> str:
    incident_id = _normalize_session_token(request.incident_id)
    if incident_id:
        return f"{incident_id}:summary"
    return "summary"


def _normalize_session_token(value: str | None) -> str:
    if not value:
        return ""
    return value.strip()


def _build_alert_fallback_key(request: AlertAnalysisRequest) -> str:
    labels = request.alert.labels
    parts = [
        _normalize_session_token(labels.get("alertname")),
        _normalize_session_token(labels.get("namespace")),
        _normalize_session_token(labels.get("pod")),
        _normalize_session_token(labels.get("workload")),
        _normalize_session_token(labels.get("destination_workload")),
        _normalize_session_token(labels.get("destination_service_name")),
    ]
    compact = "-".join(part for part in parts if part)
    return compact


def _resolve_tempo_window(
    starts_at: datetime | None,
    *,
    ends_at: datetime | None = None,
    analysis_type: str | None = None,
    lookback_minutes: int,
    forward_minutes: int,
) -> tuple[str, str]:
    if analysis_type == "resolved" and ends_at is not None:
        pivot = ends_at
    else:
        pivot = starts_at or datetime.now(timezone.utc)
    if pivot.tzinfo is None:
        pivot = pivot.replace(tzinfo=timezone.utc)
    else:
        pivot = pivot.astimezone(timezone.utc)

    start = pivot - timedelta(minutes=max(0, lookback_minutes))
    end = pivot + timedelta(minutes=max(0, forward_minutes))
    return _to_iso_z(start), _to_iso_z(end)


def _to_iso_z(value: datetime) -> str:
    return value.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")


def _collect_missing_data(
    *,
    k8s_context: K8sContext,
    tempo_context: dict[str, object] | None,
    capabilities: dict[str, str],
) -> list[str]:
    missing_data: list[str] = []
    target = _resolve_context_target(k8s_context)
    if not target.namespace:
        missing_data.append("alert.labels.namespace")
    if not target.pod_name:
        missing_data.append("alert.labels.pod")
    if not target.service_name:
        missing_data.append("analysis.target.service_name")
    if k8s_context.pod_status is None:
        missing_data.append("k8s.pod_status")
    if not k8s_context.events:
        missing_data.append("k8s.events")
    if not _has_log_lines(k8s_context.current_logs):
        missing_data.append("k8s.current_logs")
    if not _has_log_lines(k8s_context.previous_logs) and _total_restart_count(k8s_context) > 0:
        missing_data.append("k8s.previous_logs")
    if capabilities.get("k8s_core") != "ok":
        missing_data.append("k8s.core_api")
    if capabilities.get("manifest_read") != "ok":
        missing_data.append("k8s.manifest_read")
    if capabilities.get("prometheus") != "ok":
        missing_data.append("prometheus.metrics")
    if capabilities.get("loki") != "ok":
        missing_data.append("loki.logs")
    if capabilities.get("tempo") == "unavailable":
        missing_data.append("tempo.traces")
    if (
        capabilities.get("mesh_type") == "istio"
        and capabilities.get("routing_evidence") == "unavailable"
    ):
        missing_data.append("istio.routing_manifest")
    if tempo_context:
        if tempo_context.get("query_status") == "error":
            missing_data.append("tempo.query_result")
        trace_count = tempo_context.get("trace_count")
        if isinstance(trace_count, int) and trace_count <= 0:
            missing_data.append("tempo.traces")
    return _dedupe_strings(missing_data)


def _collect_analysis_warnings(
    *,
    k8s_warnings: list[str],
    tempo_context: dict[str, object] | None,
    capability_warnings: list[str],
) -> list[str]:
    warnings = [warning for warning in k8s_warnings if warning]
    warnings.extend(warning for warning in capability_warnings if warning)
    if tempo_context:
        raw_warnings = tempo_context.get("warnings")
        if isinstance(raw_warnings, list):
            for warning in raw_warnings:
                if isinstance(warning, str) and warning:
                    warnings.append(f"tempo: {warning}")
    return _dedupe_strings(warnings)


def _resolve_analysis_quality(
    *,
    k8s_context: K8sContext,
    missing_data: list[str],
    capabilities: dict[str, str],
    engine_issue: str | None,
) -> str:
    if engine_issue:
        return "low"
    if capabilities.get("k8s_core") != "ok":
        return "low"
    target = _resolve_context_target(k8s_context)
    if not target.namespace:
        return "low"
    if not target.pod_name and not target.service_name:
        if k8s_context.workload:
            return "medium"
        return "low"

    critical_missing = {
        "alert.labels.namespace",
        "k8s.core_api",
    }
    if any(item in critical_missing for item in missing_data):
        return "medium"

    return "high"


def _resolve_context_target(k8s_context: K8sContext) -> AnalysisTarget:
    if k8s_context.target is not None:
        return k8s_context.target
    return AnalysisTarget(
        namespace=k8s_context.namespace,
        pod_name=k8s_context.pod_name,
        workload=k8s_context.workload,
        service_name=None,
    )


def _total_restart_count(k8s_context: K8sContext) -> int:
    if k8s_context.pod_status is None:
        return -1
    total = 0
    for cs in k8s_context.pod_status.container_statuses:
        rc = cs.get("restart_count") if isinstance(cs, dict) else getattr(cs, "restart_count", 0)
        if isinstance(rc, int):
            total += rc
    return total


def _has_log_lines(snippets: list[object]) -> bool:
    for snippet in snippets:
        logs = getattr(snippet, "logs", None)
        if logs is None and isinstance(snippet, dict):
            logs = snippet.get("logs")
        if isinstance(logs, list) and any(isinstance(line, str) and line for line in logs):
            return True
    return False


def _resolve_mesh_type(k8s_context: K8sContext) -> str:
    if k8s_context.pod_spec is None:
        return "unknown"

    containers = k8s_context.pod_spec.get("containers")
    if not isinstance(containers, list):
        return "unknown"
    for container in containers:
        if not isinstance(container, dict):
            continue
        if container.get("name") == "istio-proxy":
            return "istio"
    return "none"


@dataclass(frozen=True)
class _AnalysisErrorCategory:
    name: str
    user_message: str


def _iter_exc_chain(exc: BaseException) -> list[BaseException]:
    """Collect all exceptions in __cause__/__context__ chain."""
    chain: list[BaseException] = []
    seen: set[int] = set()
    current: BaseException | None = exc
    while current and id(current) not in seen:
        chain.append(current)
        seen.add(id(current))
        current = getattr(current, "__cause__", None) or getattr(current, "__context__", None)
    return chain


def _categorize_analysis_error(exc: Exception) -> _AnalysisErrorCategory:
    """Categorize analysis exception for structured logging and user messaging."""
    exc_chain_str = " ".join(str(e) for e in _iter_exc_chain(exc) if str(e)).lower()
    exc_type = type(exc).__name__

    if any(
        kw in exc_chain_str for kw in ("401", "403", "api_key", "unauthorized", "authentication")
    ):
        return _AnalysisErrorCategory("llm_auth", "LLM API authentication failed")
    if any(kw in exc_chain_str for kw in ("429", "rate limit", "resource_exhausted")):
        return _AnalysisErrorCategory("llm_rate_limit", "LLM API rate limit exceeded")
    if any(kw in exc_chain_str for kw in ("timeout", "timed out", "deadline")):
        return _AnalysisErrorCategory("llm_timeout", "LLM request timed out")
    if any(kw in exc_chain_str for kw in ("connection refused", "psycopg", "operationalerror")):
        return _AnalysisErrorCategory("session_db", "session database unavailable")
    if any(kw in exc_chain_str for kw in ("400", "bad request")):
        return _AnalysisErrorCategory("llm_bad_request", "LLM API rejected the request")
    return _AnalysisErrorCategory("unknown", f"analysis failed ({exc_type})")


def _dedupe_strings(values: list[str]) -> list[str]:
    deduped: list[str] = []
    seen: set[str] = set()
    for value in values:
        if not value or value in seen:
            continue
        deduped.append(value)
        seen.add(value)
    return deduped


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
    if k8s_context.events:
        lines.append(f"recent_events ({len(k8s_context.events)}):")
        for event in k8s_context.events[:5]:
            lines.append(f"  - [{event.type}] {event.reason}: {event.message}")
    if k8s_context.warnings:
        lines.append("warnings: " + ", ".join(k8s_context.warnings))
    if k8s_context.pod_status:
        lines.append(f"pod_phase: {k8s_context.pod_status.phase}")
    if alert.annotations:
        ann_summary = alert.annotations.get("summary") or alert.annotations.get("description")
        if ann_summary:
            lines.append(f"alert_summary: {ann_summary}")
    return "\n".join(lines)


def _to_pretty_json(payload: dict[str, Any]) -> str:
    return json.dumps(payload, ensure_ascii=True, indent=2, sort_keys=True)


def _fallback_incident_summary(
    request: IncidentSummaryRequest, reason: str
) -> tuple[str, str, str]:
    """Generate fallback summary when analysis engine is unavailable."""
    alert_names = [a.alert_name for a in request.alerts]
    title = request.title  # Keep original title
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
    return title, summary, "\n".join(detail_lines)


_TITLE_MAX_LEN = 100
_SUMMARY_MAX_LEN = 300


_BOLD_VALUE_RE = re.compile(r"^(?:\d+\.\s*)?\*{1,2}[^*]+\*{1,2}\s*(.*)", re.DOTALL)


def _extract_inline_value(text: str, keys: list[str]) -> str:
    """Extract value from inline formats.

    Supports both colon-separated (``**제목 (Title)**: content``) and
    colon-less bold (``**제목 (Title)** content``) formats that LLMs
    sometimes produce.
    """
    for line in text.strip().splitlines():
        stripped = line.strip()
        if not stripped:
            continue
        lowered = stripped.lower()
        if not any(key in lowered for key in keys):
            continue
        # 1) Colon-separated: **key**: value  /  key: value
        if ":" in stripped:
            return stripped.split(":", 1)[1].strip().strip("'\"*")
        # 2) Bold-marker without colon: **key** value
        m = _BOLD_VALUE_RE.match(stripped)
        if m and m.group(1).strip():
            return m.group(1).strip().strip("'\"*")
    return ""


def _parse_incident_summary(result: str, original_title: str) -> tuple[str, str, str]:
    """Parse AI response into title, summary and detail.

    The AI is instructed to return structured content with 제목, 요약 and 상세 분석 sections.
    Tries markdown-header extraction first, then inline colon-based extraction.
    """
    all_keys = ["제목", "title", "요약", "summary", "상세", "detail"]

    # 1) Markdown header format (### 제목\ncontent)
    title = _extract_section(result, ["제목", "title"], all_keys) or ""
    summary = _extract_section(result, ["요약", "summary"], all_keys) or ""

    # 2) Inline format (**제목 (Title)**: content)
    if not title:
        title = _extract_inline_value(result, ["제목", "title"])
    if not summary:
        summary = _extract_inline_value(result, ["요약", "summary"])

    # Title should be first line only
    if title:
        title = title.splitlines()[0].strip().strip("'\"*")

    # Fallbacks
    if not title:
        title = original_title
    if not summary:
        # Extract first non-empty paragraph instead of the entire response.
        # When title was already extracted, skip past it to avoid duplication.
        remainder = result
        if title and title in remainder:
            idx = remainder.index(title) + len(title)
            remainder = remainder[idx:]
        summary = _extract_first_paragraph(remainder) or _extract_first_paragraph(result)

    # Enforce title length limit
    if len(title) > _TITLE_MAX_LEN:
        title = title[:_TITLE_MAX_LEN].rstrip() + "…"

    # Extract only the detail section; fall back to full result if not found.
    detail = _extract_section(result, ["상세 분석", "상세", "detail"], all_keys) or result
    return title, summary, detail


def _build_incident_summary_prompt(request: IncidentSummaryRequest, masker: Masker) -> str:
    alerts_info = []
    for alert in request.alerts:
        alert_data = {
            "fingerprint": alert.fingerprint,
            "alert_name": alert.alert_name,
            "severity": alert.severity,
            "status": alert.status,
            "analysis_summary": alert.analysis_summary or "N/A",
            "analysis_detail": alert.analysis_detail or "N/A",
            "artifacts": [artifact.model_dump() for artifact in alert.artifacts or []],
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
    incident_data = cast(dict[str, Any], masker.mask_object(incident_data))

    return (
        "You are kube-rca-agent. An incident has been resolved and you need to "
        "provide a final RCA summary.\n"
        "Analyze ALL alerts and their individual analyses to synthesize a "
        "comprehensive incident summary. Every distinct alert type "
        "(e.g. 5xx errors AND 4xx errors) must be addressed in the summary "
        "and detail sections.\n\n"
        "IMPORTANT: Use EXACTLY this format with colon separators:\n"
        "**제목 (Title)**: A concise incident title (max 100 chars) that includes:\n"
        "  - The specific service/pod/namespace affected\n"
        "  - The root cause or error type\n"
        "  - Examples:\n"
        "    - '[payment-service] OOMKilled로 인한 Pod 재시작'\n"
        "    - '[nginx/prod] ImagePullBackOff - 잘못된 이미지 태그'\n"
        "    - '[redis-cluster] 메모리 부족으로 인한 연결 실패'\n"
        "**요약 (Summary)**: 1-2 sentences describing the root cause and resolution\n"
        "**상세 분석 (Detail)**:\n"
        "  - 근본 원인 (Root Cause)\n"
        "  - 영향 범위 (Impact)\n"
        "  - 해결 과정 (Resolution)\n"
        "  - 재발 방지 권고 (Prevention Recommendations)\n\n"
        f"Incident data:\n{_to_pretty_json(incident_data)}\n"
    )


def _split_alert_analysis(result: str) -> tuple[str, str]:
    all_keys = ["요약", "summary", "상세", "detail"]
    summary = _extract_section(result, ["요약", "summary"], all_keys)
    detail = _extract_section(result, ["상세 분석", "상세", "detail"], all_keys)

    if not detail:
        detail = result.strip()
    if not summary:
        summary = _extract_first_paragraph(result)
        if len(summary) > _SUMMARY_MAX_LEN:
            summary = summary[:_SUMMARY_MAX_LEN].rstrip() + "…"

    return summary, detail


def _extract_first_paragraph(text: str) -> str:
    """Extract the first non-empty paragraph from text.

    Splits on blank lines and returns the first paragraph that is not a
    markdown header (``#``-prefixed) or a section label ending with ``:``.
    """
    paragraph: list[str] = []
    for line in text.strip().splitlines():
        stripped = line.strip()
        if not stripped:
            if paragraph:
                break
            continue
        # Skip markdown headers and section labels
        if stripped.startswith("#") or stripped.endswith(":") or stripped.endswith("："):
            if paragraph:
                break
            continue
        paragraph.append(stripped)
    return (
        " ".join(paragraph) if paragraph else text.strip().splitlines()[0] if text.strip() else ""
    )


def _extract_section(text: str, keys: list[str], all_keys: list[str]) -> str | None:
    lines = text.splitlines()
    start = None
    for idx, line in enumerate(lines):
        if _is_section_header(line, keys):
            start = idx + 1
            break
    if start is None:
        return None

    end = len(lines)
    for idx in range(start, len(lines)):
        if _is_section_header(lines[idx], all_keys):
            end = idx
            break

    section = "\n".join(lines[start:end]).strip()
    return section or None


def _is_section_header(line: str, keys: list[str]) -> bool:
    stripped = line.strip()
    if not stripped:
        return False
    lowered = stripped.lower()
    has_key = any(key in lowered for key in keys)
    if not has_key:
        return False
    return stripped.startswith("#") or stripped.endswith(":") or stripped.endswith("：")


def _compact_summary(text: str, limit: int = 300) -> str:
    lines = [line.strip() for line in text.splitlines() if line.strip()]
    if not lines:
        return ""
    summary = lines[0]
    if len(lines) > 1 and len(summary) < limit:
        summary = f"{summary} {lines[1]}"
    if len(summary) > limit:
        summary = summary[:limit].rstrip() + "..."
    return summary


def _build_alert_artifacts(
    k8s_context: K8sContext,
    tempo_context: dict[str, object] | None = None,
) -> list[dict[str, object]]:
    artifacts: list[dict[str, object]] = []

    for event in k8s_context.events:
        event_data = event.to_dict()
        summary = " / ".join(
            part
            for part in [
                event_data.get("reason"),
                event_data.get("message"),
            ]
            if part
        )
        artifacts.append(
            {
                "type": "event",
                "summary": summary or "k8s event",
                "result": event_data,
            }
        )

    for snippet in [*k8s_context.current_logs, *k8s_context.previous_logs]:
        logs = snippet.logs[:20]
        log_kind = "previous" if snippet.previous else "current"
        artifacts.append(
            {
                "type": "log",
                "summary": f"{snippet.container} {log_kind} logs ({len(logs)} lines)",
                "result": {
                    "container": snippet.container,
                    "previous": snippet.previous,
                    "error": snippet.error,
                    "logs": logs,
                },
            }
        )

    artifacts.extend(_build_tempo_artifacts(tempo_context))
    return artifacts


def _extract_tempo_traces(response: dict[str, object]) -> list[dict[str, object]]:
    traces = response.get("traces")
    if not isinstance(traces, list):
        return []
    return [trace for trace in traces if isinstance(trace, dict)]


def _compact_tempo_context(tempo_context: dict[str, object]) -> dict[str, object]:
    traces = tempo_context.get("traces")
    compact_traces: list[dict[str, object]] = []
    if isinstance(traces, list):
        for trace in traces[:3]:
            if not isinstance(trace, dict):
                continue
            compact_traces.append(
                {
                    "trace_id": trace.get("trace_id"),
                    "root_service_name": trace.get("root_service_name"),
                    "root_trace_name": trace.get("root_trace_name"),
                    "duration_ms": trace.get("duration_ms"),
                }
            )
    return {
        "query": tempo_context.get("query"),
        "window": tempo_context.get("window"),
        "query_status": tempo_context.get("query_status"),
        "trace_count": tempo_context.get("trace_count"),
        "traces": compact_traces,
        "warnings": tempo_context.get("warnings") or [],
    }


def _build_tempo_artifacts(tempo_context: dict[str, object] | None) -> list[dict[str, object]]:
    if not tempo_context:
        return []

    query = tempo_context.get("query")
    window = tempo_context.get("window")
    warnings = tempo_context.get("warnings") or []
    traces = tempo_context.get("traces") if isinstance(tempo_context.get("traces"), list) else []
    raw_query_status = tempo_context.get("query_status")
    query_status = raw_query_status if isinstance(raw_query_status, str) else "ok"
    trace_count = tempo_context.get("trace_count")
    if not isinstance(trace_count, int):
        trace_count = len(traces)

    summary = (
        "tempo trace query failed" if query_status == "error" else f"tempo traces ({trace_count})"
    )
    artifacts: list[dict[str, object]] = [
        {
            "type": "trace_summary",
            "query": query if isinstance(query, str) else None,
            "summary": summary,
            "result": {
                "window": window,
                "query_status": query_status,
                "trace_count": trace_count,
                "warnings": warnings,
            },
        }
    ]
    if query_status == "error":
        artifacts.append(
            {
                "type": "trace_warning",
                "query": query if isinstance(query, str) else None,
                "summary": "tempo trace query failed",
                "result": {
                    "window": window,
                    "warnings": warnings,
                },
            }
        )

    for trace in traces[:3]:
        if not isinstance(trace, dict):
            continue
        trace_id = trace.get("trace_id")
        summary_parts = [
            str(trace.get("root_service_name") or "").strip(),
            str(trace.get("root_trace_name") or "").strip(),
        ]
        summary = " / ".join(part for part in summary_parts if part)
        if trace_id:
            summary = f"{summary} ({trace_id})" if summary else str(trace_id)
        artifacts.append(
            {
                "type": "trace",
                "query": query if isinstance(query, str) else None,
                "summary": summary or "tempo trace",
                "result": trace,
            }
        )
    return artifacts
