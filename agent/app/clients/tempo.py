from __future__ import annotations

import json
import logging
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from datetime import datetime, timezone

from app.core.config import Settings


@dataclass(frozen=True)
class TempoEndpoint:
    base_url: str
    tenant_id: str | None

    def to_dict(self) -> dict[str, object]:
        return {
            "base_url": self.base_url,
            "tenant_id": self.tenant_id or "",
        }


class TempoClient:
    def __init__(self, settings: Settings) -> None:
        self._logger = logging.getLogger(__name__)
        self._base_url = _normalize_base_url(settings.tempo_url)
        self._timeout_seconds = settings.tempo_http_timeout_seconds
        self._tenant_id = settings.tempo_tenant_id.strip()
        if settings.tempo_url and not self._base_url:
            self._logger.warning("Invalid TEMPO_URL: %s", settings.tempo_url)

    @property
    def enabled(self) -> bool:
        return bool(self._base_url)

    def describe_endpoint(self) -> dict[str, object]:
        endpoint, detail = self._resolve_endpoint()
        if endpoint:
            return {"endpoint": endpoint.to_dict()}
        return detail

    def search_traces(
        self,
        *,
        query: str,
        start: str,
        end: str,
        limit: int = 5,
    ) -> dict[str, object]:
        endpoint, detail = self._resolve_endpoint()
        if endpoint is None:
            return detail

        params = {
            "q": query,
            "start": _normalize_search_time(start),
            "end": _normalize_search_time(end),
            "limit": str(max(1, limit)),
        }
        payload, error = self._request_with_fallback(
            endpoint,
            paths=["/api/search", "/tempo/api/search"],
            params=params,
        )
        if error is not None:
            return {
                "error": "failed to search tempo traces",
                "detail": error,
                "endpoint": endpoint.to_dict(),
                "query": query,
                "window": {"start": start, "end": end},
            }

        traces = _extract_trace_summaries(payload)
        return {
            "endpoint": endpoint.to_dict(),
            "query": query,
            "window": {"start": start, "end": end},
            "trace_count": len(traces),
            "traces": traces,
            "data": payload,
        }

    def get_trace(self, trace_id: str) -> dict[str, object]:
        endpoint, detail = self._resolve_endpoint()
        if endpoint is None:
            return detail

        trace_path = urllib.parse.quote(trace_id, safe="")
        payload, error = self._request_with_fallback(
            endpoint,
            paths=[
                f"/api/traces/{trace_path}",
                f"/tempo/api/traces/{trace_path}",
            ],
            params=None,
        )
        if error is not None:
            return {
                "error": "failed to get tempo trace",
                "detail": error,
                "endpoint": endpoint.to_dict(),
                "trace_id": trace_id,
            }

        return {
            "endpoint": endpoint.to_dict(),
            "trace_id": trace_id,
            "data": payload,
        }

    def _request_with_fallback(
        self,
        endpoint: TempoEndpoint,
        *,
        paths: list[str],
        params: dict[str, str] | None,
    ) -> tuple[dict[str, object] | list[object] | None, dict[str, object] | None]:
        last_error: dict[str, object] | None = None
        for path in paths:
            payload, error, status_code = self._request_json(
                endpoint,
                path=path,
                params=params,
            )
            if error is None:
                return payload, None
            last_error = error
            if status_code == 404:
                continue
            return None, error
        return None, last_error

    def _request_json(
        self,
        endpoint: TempoEndpoint,
        *,
        path: str,
        params: dict[str, str] | None,
    ) -> tuple[
        dict[str, object] | list[object] | None,
        dict[str, object] | None,
        int | None,
    ]:
        url = f"{endpoint.base_url}{path}"
        if params:
            url = f"{url}?{urllib.parse.urlencode(params)}"

        headers = {"Accept": "application/json"}
        if endpoint.tenant_id:
            headers["X-Scope-OrgID"] = endpoint.tenant_id

        request = urllib.request.Request(url, headers=headers)
        try:
            with urllib.request.urlopen(request, timeout=self._timeout_seconds) as response:
                payload = response.read()
        except urllib.error.HTTPError as exc:
            body = exc.read().decode("utf-8", errors="replace")
            self._logger.warning("Tempo HTTP error %s for %s", exc.code, url)
            return None, {
                "status_code": exc.code,
                "reason": str(exc.reason),
                "url": url,
                "body": body[:300],
            }, exc.code
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to query Tempo: %s", exc)
            return None, {"reason": str(exc), "url": url}, None

        try:
            data = json.loads(payload.decode("utf-8"))
        except json.JSONDecodeError as exc:
            self._logger.warning("Failed to decode Tempo response: %s", exc)
            return None, {
                "reason": "failed to decode tempo response",
                "detail": str(exc),
                "url": url,
            }, None

        if not isinstance(data, (dict, list)):
            return None, {"reason": "unexpected tempo payload type", "url": url}, None
        return data, None, 200

    def _resolve_endpoint(self) -> tuple[TempoEndpoint | None, dict[str, object]]:
        if not self._base_url:
            return None, {"warning": "tempo url not configured"}
        endpoint = TempoEndpoint(
            base_url=self._base_url,
            tenant_id=self._tenant_id or None,
        )
        return endpoint, {"endpoint": endpoint.to_dict()}


def build_traceql_query(service_name: str | None, namespace: str | None) -> str:
    filters: list[str] = []
    if service_name:
        escaped = service_name.replace('"', '\\"')
        filters.append(f'resource.service.name = "{escaped}"')
    if namespace:
        escaped = namespace.replace('"', '\\"')
        filters.append(f'resource.k8s.namespace.name = "{escaped}"')
    if not filters:
        return "{}"
    return "{ " + " && ".join(filters) + " }"


def _extract_trace_summaries(
    payload: dict[str, object] | list[object],
) -> list[dict[str, object]]:
    entries: list[object]
    if isinstance(payload, dict):
        if isinstance(payload.get("traces"), list):
            entries = payload.get("traces", [])
        elif isinstance(payload.get("data"), list):
            entries = payload.get("data", [])
        elif isinstance(payload.get("results"), list):
            entries = payload.get("results", [])
        else:
            entries = []
    elif isinstance(payload, list):
        entries = payload
    else:
        entries = []

    normalized: list[dict[str, object]] = []
    for entry in entries:
        if not isinstance(entry, dict):
            continue
        trace_id = _first_non_empty(entry, ["traceID", "traceId", "trace_id", "id"])
        if not trace_id:
            continue
        summary: dict[str, object] = {
            "trace_id": trace_id,
            "root_service_name": _first_non_empty(
                entry, ["rootServiceName", "root_service_name", "serviceName", "service_name"]
            ),
            "root_trace_name": _first_non_empty(
                entry, ["rootTraceName", "root_trace_name", "traceName", "trace_name"]
            ),
            "start_time_unix_nano": _first_non_empty(
                entry, ["startTimeUnixNano", "start_time_unix_nano"]
            ),
            "duration_ms": _first_non_empty(entry, ["durationMs", "duration_ms"]),
        }
        if entry.get("spanSet"):
            summary["span_set"] = entry.get("spanSet")
        normalized.append(summary)
    return normalized


def _first_non_empty(entry: dict[str, object], keys: list[str]) -> object | None:
    for key in keys:
        value = entry.get(key)
        if value not in ("", None):
            return value
    return None


def _normalize_base_url(raw: str) -> str:
    value = raw.strip()
    if not value:
        return ""
    if "://" not in value:
        value = f"http://{value}"
    parsed = urllib.parse.urlparse(value)
    if not parsed.scheme or not parsed.netloc:
        return ""
    return value.rstrip("/")


def _normalize_search_time(raw: str) -> str:
    value = raw.strip()
    if not value:
        return raw

    # Tempo /api/search expects unix timestamp values for start/end.
    try:
        return str(int(float(value)))
    except ValueError:
        pass

    iso_value = value.replace("Z", "+00:00") if value.endswith("Z") else value
    try:
        parsed = datetime.fromisoformat(iso_value)
    except ValueError:
        return raw

    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    else:
        parsed = parsed.astimezone(timezone.utc)
    return str(int(parsed.timestamp()))
