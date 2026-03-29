from __future__ import annotations

import json
import logging
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass

from app.core.config import Settings


@dataclass(frozen=True)
class LokiEndpoint:
    base_url: str
    tenant_id: str | None

    def to_dict(self) -> dict[str, object]:
        return {
            "base_url": self.base_url,
            "tenant_id": self.tenant_id or "",
        }


class LokiClient:
    def __init__(self, settings: Settings) -> None:
        self._logger = logging.getLogger(__name__)
        self._base_url = _normalize_base_url(settings.loki_url)
        self._timeout_seconds = settings.loki_http_timeout_seconds
        self._tenant_id = settings.loki_tenant_id.strip()
        if settings.loki_url and not self._base_url:
            self._logger.warning("Invalid LOKI_URL: %s", settings.loki_url)

    @property
    def enabled(self) -> bool:
        return bool(self._base_url)

    def describe_endpoint(self) -> dict[str, object]:
        endpoint, detail = self._resolve_endpoint()
        if endpoint:
            return {"endpoint": endpoint.to_dict()}
        return detail

    def list_labels(self) -> dict[str, object]:
        """List available label names from Loki."""
        endpoint, detail = self._resolve_endpoint()
        if endpoint is None:
            return detail

        return self._get_json(endpoint, "/loki/api/v1/labels")

    def label_values(self, label: str) -> dict[str, object]:
        """List values for a specific label.

        Args:
            label: Label name (e.g., 'namespace', 'pod', 'container').
        """
        endpoint, detail = self._resolve_endpoint()
        if endpoint is None:
            return detail

        safe_label = urllib.parse.quote(label, safe="")
        return self._get_json(endpoint, f"/loki/api/v1/label/{safe_label}/values")

    def query(self, query: str, *, limit: int = 100, time: str | None = None) -> dict[str, object]:
        """Run a Loki instant query (LogQL).

        Args:
            query: LogQL query string.
            limit: Max log entries to return.
            time: Evaluation time (RFC3339 or Unix timestamp, optional).
        """
        endpoint, detail = self._resolve_endpoint()
        if endpoint is None:
            return detail

        params: dict[str, str] = {"query": query, "limit": str(max(1, limit))}
        if time:
            params["time"] = time

        return self._get_json(endpoint, "/loki/api/v1/query", params=params)

    def query_range(
        self,
        query: str,
        *,
        start: str,
        end: str,
        limit: int = 100,
        step: str | None = None,
    ) -> dict[str, object]:
        """Run a Loki range query to get log entries over a time window.

        Args:
            query: LogQL query string.
            start: Start time (RFC3339 or Unix timestamp).
            end: End time (RFC3339 or Unix timestamp).
            limit: Max log entries to return.
            step: Query resolution step (e.g., '1m', '5m'). Optional.
        """
        endpoint, detail = self._resolve_endpoint()
        if endpoint is None:
            return detail

        params: dict[str, str] = {
            "query": query,
            "start": start,
            "end": end,
            "limit": str(max(1, limit)),
        }
        if step:
            params["step"] = step

        return self._get_json(endpoint, "/loki/api/v1/query_range", params=params)

    def _get_json(
        self,
        endpoint: LokiEndpoint,
        path: str,
        params: dict[str, str] | None = None,
    ) -> dict[str, object]:
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
            self._logger.warning("Loki HTTP error %s for %s", exc.code, url)
            return {
                "error": "failed to query loki",
                "detail": {"status_code": exc.code, "reason": str(exc.reason), "body": body[:300]},
                "endpoint": endpoint.to_dict(),
            }
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to query Loki: %s", exc)
            return {
                "error": "failed to query loki",
                "detail": str(exc),
                "endpoint": endpoint.to_dict(),
            }

        try:
            data = json.loads(payload.decode("utf-8"))
        except json.JSONDecodeError as exc:
            self._logger.warning("Failed to decode Loki response: %s", exc)
            return {
                "error": "failed to decode loki response",
                "detail": str(exc),
                "endpoint": endpoint.to_dict(),
            }

        return {"endpoint": endpoint.to_dict(), "data": data}

    def _resolve_endpoint(self) -> tuple[LokiEndpoint | None, dict[str, object]]:
        if not self._base_url:
            return None, {"warning": "loki url not configured"}
        endpoint = LokiEndpoint(
            base_url=self._base_url,
            tenant_id=self._tenant_id or None,
        )
        return endpoint, {"endpoint": endpoint.to_dict()}


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
