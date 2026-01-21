from __future__ import annotations

import json
import logging
import urllib.parse
import urllib.request
from dataclasses import dataclass

from app.core.config import Settings


@dataclass(frozen=True)
class PrometheusEndpoint:
    base_url: str

    def to_dict(self) -> dict[str, object]:
        return {
            "base_url": self.base_url,
        }


class PrometheusClient:
    def __init__(self, settings: Settings) -> None:
        self._logger = logging.getLogger(__name__)
        self._base_url = _normalize_base_url(settings.prometheus_url)
        self._timeout_seconds = settings.prometheus_http_timeout_seconds
        if settings.prometheus_url and not self._base_url:
            self._logger.warning("Invalid PROMETHEUS_URL: %s", settings.prometheus_url)

    @property
    def enabled(self) -> bool:
        return bool(self._base_url)

    def describe_endpoint(self) -> dict[str, object]:
        endpoint, detail = self._resolve_endpoint()
        if endpoint:
            return {"endpoint": endpoint.to_dict()}
        return detail

    def list_metrics(self, match: str | None = None) -> dict[str, object]:
        """List available metric names from Prometheus.

        Args:
            match: Optional regex pattern to filter metric names.
                   Examples: 'kube_pod.*', 'istio_requests.*', 'container_.*'
        """
        endpoint, detail = self._resolve_endpoint()
        if endpoint is None:
            return detail

        url = f"{endpoint.base_url}/api/v1/label/__name__/values"

        try:
            with urllib.request.urlopen(url, timeout=self._timeout_seconds) as response:
                payload = response.read()
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to list Prometheus metrics: %s", exc)
            return {
                "error": "failed to list Prometheus metrics",
                "detail": str(exc),
                "endpoint": endpoint.to_dict(),
            }

        try:
            data = json.loads(payload.decode("utf-8"))
        except json.JSONDecodeError as exc:
            self._logger.warning("Failed to decode Prometheus response: %s", exc)
            return {
                "error": "failed to decode Prometheus response",
                "detail": str(exc),
                "endpoint": endpoint.to_dict(),
            }

        metrics = data.get("data", [])
        if match:
            import re

            try:
                pattern = re.compile(match)
                metrics = [m for m in metrics if pattern.search(m)]
            except re.error as exc:
                return {
                    "error": "invalid regex pattern",
                    "detail": str(exc),
                    "pattern": match,
                }

        return {"endpoint": endpoint.to_dict(), "metrics": metrics, "count": len(metrics)}

    def query(self, query: str, *, time: str | None = None) -> dict[str, object]:
        endpoint, detail = self._resolve_endpoint()
        if endpoint is None:
            return detail

        params: dict[str, str] = {"query": query}
        if time:
            params["time"] = time
        url = f"{endpoint.base_url}/api/v1/query?{urllib.parse.urlencode(params)}"

        try:
            with urllib.request.urlopen(url, timeout=self._timeout_seconds) as response:
                payload = response.read()
        except Exception as exc:  # noqa: BLE001
            self._logger.warning("Failed to query Prometheus: %s", exc)
            return {
                "error": "failed to query Prometheus",
                "detail": str(exc),
                "endpoint": endpoint.to_dict(),
            }

        try:
            data = json.loads(payload.decode("utf-8"))
        except json.JSONDecodeError as exc:
            self._logger.warning("Failed to decode Prometheus response: %s", exc)
            return {
                "error": "failed to decode Prometheus response",
                "detail": str(exc),
                "endpoint": endpoint.to_dict(),
                "raw": payload.decode("utf-8", errors="replace"),
            }

        return {"endpoint": endpoint.to_dict(), "data": data}

    def _resolve_endpoint(self) -> tuple[PrometheusEndpoint | None, dict[str, object]]:
        if not self._base_url:
            return None, {"warning": "prometheus url not configured"}
        endpoint = PrometheusEndpoint(base_url=self._base_url)
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
