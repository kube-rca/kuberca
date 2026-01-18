from __future__ import annotations

import json
import logging
import urllib.parse
import urllib.request
from dataclasses import dataclass

from kubernetes import client

from app.clients.k8s import KubernetesClient
from app.core.config import Settings


@dataclass(frozen=True)
class PrometheusEndpoint:
    name: str
    namespace: str
    port: int
    scheme: str

    @property
    def base_url(self) -> str:
        return f"{self.scheme}://{self.name}.{self.namespace}.svc:{self.port}"

    def to_dict(self) -> dict[str, object]:
        return {
            "name": self.name,
            "namespace": self.namespace,
            "port": self.port,
            "scheme": self.scheme,
            "base_url": self.base_url,
        }


class PrometheusClient:
    def __init__(self, settings: Settings, k8s_client: KubernetesClient) -> None:
        self._logger = logging.getLogger(__name__)
        self._k8s_client = k8s_client
        self._label_selector = settings.prometheus_label_selector
        self._namespace_allowlist = settings.prometheus_namespace_allowlist
        self._port_name = settings.prometheus_port_name
        self._scheme = settings.prometheus_scheme
        self._timeout_seconds = settings.prometheus_http_timeout_seconds
        self._endpoint: PrometheusEndpoint | None = None

    def describe_endpoint(self) -> dict[str, object]:
        endpoint, detail = self._discover_endpoint()
        if endpoint:
            return {"endpoint": endpoint.to_dict()}
        return detail

    def list_metrics(self, match: str | None = None) -> dict[str, object]:
        """List available metric names from Prometheus.

        Args:
            match: Optional regex pattern to filter metric names.
                   Examples: 'kube_pod.*', 'istio_requests.*', 'container_.*'
        """
        endpoint, detail = self._discover_endpoint()
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
        endpoint, detail = self._discover_endpoint()
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

    def _discover_endpoint(self) -> tuple[PrometheusEndpoint | None, dict[str, object]]:
        if self._endpoint:
            return self._endpoint, {"endpoint": self._endpoint.to_dict()}
        if not self._label_selector:
            return None, {"error": "prometheus label selector not configured"}

        services = self._k8s_client.list_services_by_label(
            self._label_selector,
            namespaces=self._namespace_allowlist or None,
        )
        if not services:
            return None, {"error": "prometheus service not found"}
        if len(services) > 1:
            return None, {
                "error": "multiple prometheus services found",
                "candidates": [self._describe_service(item) for item in services],
            }

        service = services[0]
        port = self._select_port(service)
        if port is None:
            return None, {
                "error": "prometheus service port not resolved",
                "service": self._describe_service(service),
            }

        metadata = service.metadata
        name = metadata.name if metadata else ""
        namespace = metadata.namespace if metadata else ""
        endpoint = PrometheusEndpoint(
            name=name,
            namespace=namespace,
            port=port,
            scheme=self._scheme,
        )
        self._endpoint = endpoint
        return endpoint, {"endpoint": endpoint.to_dict()}

    def _select_port(self, service: client.V1Service) -> int | None:
        ports = service.spec.ports if service.spec else None
        if not ports:
            return None
        if self._port_name:
            for port in ports:
                if port.name == self._port_name:
                    return port.port
            return None
        if len(ports) == 1:
            return ports[0].port
        return None

    @staticmethod
    def _describe_service(service: client.V1Service) -> dict[str, object]:
        metadata = service.metadata
        return {
            "name": metadata.name if metadata else "",
            "namespace": metadata.namespace if metadata else "",
            "labels": metadata.labels if metadata else {},
            "ports": [
                {"name": port.name, "port": port.port, "protocol": port.protocol}
                for port in (service.spec.ports if service.spec and service.spec.ports else [])
            ],
        }
