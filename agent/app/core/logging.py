from __future__ import annotations

import logging
from collections.abc import Iterable


class _HealthCheckFilter(logging.Filter):
    def __init__(self, paths: Iterable[str]) -> None:
        super().__init__()
        self._paths = {path.rstrip("/") or "/" for path in paths}

    def filter(self, record: logging.LogRecord) -> bool:
        path = _extract_path(record)
        if path and path in self._paths:
            return False
        return True


def _extract_path(record: logging.LogRecord) -> str | None:
    args = record.args
    if isinstance(args, tuple) and len(args) >= 3:
        path = str(args[2])
    elif isinstance(args, dict) and "path" in args:
        path = str(args["path"])
    else:
        path = ""
    if path:
        return path.split("?", 1)[0]
    message = record.getMessage()
    for candidate in ("/healthz", "/ping", "/openapi.json", "/"):
        if f" {candidate} " in message or f'"{candidate} ' in message:
            return candidate
    return None


def configure_logging(level: str) -> None:
    logging.basicConfig(
        level=level.upper(),
        format="%(asctime)s %(levelname)s %(name)s %(message)s",
    )
    access_logger = logging.getLogger("uvicorn.access")
    access_logger.addFilter(_HealthCheckFilter({"/healthz", "/ping", "/openapi.json", "/"}))
