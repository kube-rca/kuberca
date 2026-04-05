from __future__ import annotations

import time
from collections import OrderedDict
from contextlib import contextmanager
from threading import Lock
from unittest.mock import MagicMock, patch

import httpx
import pytest
from tenacity import RetryError

from app.clients.strands_agent import StrandsAnalysisEngine, _has_transport_error, _is_retryable


class _FakeServerError(Exception):
    """Simulate a Google Gemini 503 ServerError."""


# Patch module/class so _is_retryable recognises it as a Google ServerError
_FakeServerError.__module__ = "google.genai.errors"
_FakeServerError.__qualname__ = "ServerError"
_FakeServerError.__name__ = "ServerError"


# ---------------------------------------------------------------------------
# httpx transport error tests
# ---------------------------------------------------------------------------


def test_httpx_read_error_is_retryable() -> None:
    assert _is_retryable(httpx.ReadError("connection reset"))


def test_httpx_connect_error_is_retryable() -> None:
    assert _is_retryable(httpx.ConnectError("connection refused"))


def test_httpx_remote_protocol_error_is_retryable() -> None:
    assert _is_retryable(httpx.RemoteProtocolError("protocol error"))


def test_wrapped_httpx_error_is_retryable() -> None:
    """Transport error wrapped via __cause__ should be retryable."""
    wrapper = RuntimeError("event loop failed")
    wrapper.__cause__ = httpx.ReadError("connection reset")
    assert _is_retryable(wrapper)


def test_has_transport_error_direct() -> None:
    assert _has_transport_error(httpx.ReadError("read error"))


def test_has_transport_error_chained() -> None:
    wrapper = RuntimeError("strands error")
    wrapper.__cause__ = httpx.ReadError("read error")
    assert _has_transport_error(wrapper)


def test_has_transport_error_false_for_server_error() -> None:
    assert not _has_transport_error(_FakeServerError("503"))


def test_retries_and_sanitizes_on_transport_error() -> None:
    """On httpx.ReadError, retry should sanitize messages before next attempt."""
    engine = _build_engine(max_attempts=5, total_timeout=10)
    call_count = 0

    def _fake_agent(prompt: str) -> str:
        nonlocal call_count
        call_count += 1
        if call_count == 1:
            raise httpx.ReadError("connection reset")
        return "success"

    agent = MagicMock()
    agent.side_effect = _fake_agent
    agent.messages = []

    with patch("app.clients.strands_agent._sanitize_message_order") as mock_sanitize:
        result = engine._invoke_with_retry(agent, "test")

    assert result == "success"
    assert call_count == 2
    mock_sanitize.assert_called_once_with([])


class _FakeSessionRepository:
    @contextmanager
    def session_lock(self, session_id: str):
        yield

    def read_conversation_manager_name(self, session_id: str) -> str | None:
        return None

    def delete_session(self, session_id: str) -> None:
        pass


def _build_engine(
    *,
    max_attempts: int = 8,
    min_wait: float = 0.0,
    max_wait: float = 0.0,
    total_timeout: float = 120.0,
) -> StrandsAnalysisEngine:
    engine = StrandsAnalysisEngine.__new__(StrandsAnalysisEngine)
    engine._session_repo = _FakeSessionRepository()  # type: ignore[attr-defined]
    engine._cache_lock = Lock()
    engine._agent_cache = OrderedDict()
    engine._retry_max_attempts = max_attempts
    engine._retry_min_wait = min_wait
    engine._retry_max_wait = max_wait
    engine._retry_total_timeout = total_timeout
    return engine


def test_fake_server_error_is_retryable() -> None:
    assert _is_retryable(_FakeServerError("503"))


def test_stops_after_max_attempts() -> None:
    engine = _build_engine(max_attempts=3, total_timeout=600)
    call_count = 0

    def _fake_agent(prompt: str) -> str:
        nonlocal call_count
        call_count += 1
        raise _FakeServerError("503 UNAVAILABLE")

    agent = MagicMock()
    agent.side_effect = _fake_agent

    with pytest.raises(RetryError):
        engine._invoke_with_retry(agent, "test")

    assert call_count == 3


def test_stops_after_total_timeout() -> None:
    engine = _build_engine(max_attempts=100, total_timeout=0.5, min_wait=0.2, max_wait=0.2)
    call_count = 0

    def _fake_agent(prompt: str) -> str:
        nonlocal call_count
        call_count += 1
        raise _FakeServerError("503 UNAVAILABLE")

    agent = MagicMock()
    agent.side_effect = _fake_agent

    start = time.monotonic()
    with pytest.raises(RetryError):
        engine._invoke_with_retry(agent, "test")
    elapsed = time.monotonic() - start

    # Should stop due to total_timeout (0.5s), not max_attempts (100)
    assert call_count < 100
    assert elapsed < 5.0


def test_succeeds_on_retry() -> None:
    engine = _build_engine(max_attempts=5, total_timeout=10)
    call_count = 0

    def _fake_agent(prompt: str) -> str:
        nonlocal call_count
        call_count += 1
        if call_count < 3:
            raise _FakeServerError("503 UNAVAILABLE")
        return "success"

    agent = MagicMock()
    agent.side_effect = _fake_agent

    result = engine._invoke_with_retry(agent, "test")

    assert result == "success"
    assert call_count == 3
