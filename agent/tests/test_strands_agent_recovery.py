from __future__ import annotations

from collections import OrderedDict
from contextlib import contextmanager
from threading import Lock

import pytest

from app.clients.strands_agent import StrandsAnalysisEngine


class FakeSessionRepository:
    def __init__(self, manager_name: str | None = None) -> None:
        self.deleted_sessions: list[str] = []
        self.manager_name = manager_name

    @contextmanager
    def session_lock(self, session_id: str):
        yield

    def read_conversation_manager_name(self, session_id: str) -> str | None:
        return self.manager_name

    def delete_session(self, session_id: str) -> None:
        self.deleted_sessions.append(session_id)


def _build_engine(
    sequence: list[object],
) -> tuple[StrandsAnalysisEngine, FakeSessionRepository, dict[str, int]]:
    repo = FakeSessionRepository(manager_name="SlidingWindowConversationManager")
    engine = StrandsAnalysisEngine.__new__(StrandsAnalysisEngine)
    engine._session_repo = repo  # type: ignore[attr-defined]
    engine._cache_lock = Lock()  # type: ignore[attr-defined]
    engine._agent_cache = OrderedDict({"conv-1:chat": object()})  # type: ignore[attr-defined]

    calls = {"count": 0}

    def _resolve_session_id(_: str | None) -> str:
        return "conv-1:chat"

    def _analyze_once(prompt: str, session_id: str) -> str:
        idx = calls["count"]
        calls["count"] += 1
        value = sequence[idx]
        if isinstance(value, Exception):
            raise value
        return str(value)

    engine._resolve_session_id = _resolve_session_id  # type: ignore[attr-defined]
    engine._analyze_once = _analyze_once  # type: ignore[attr-defined]
    return engine, repo, calls


def test_recover_invalid_conversation_state_and_retry_once() -> None:
    engine, repo, calls = _build_engine(
        [
            ValueError("Invalid conversation manager state."),
            "recovered-response",
        ]
    )

    result = engine.analyze("prompt", "incident-1")

    assert result == "recovered-response"
    assert calls["count"] == 2
    assert repo.deleted_sessions == ["conv-1:chat"]
    assert "conv-1:chat" not in engine._agent_cache


def test_recovery_fails_then_raises_after_single_retry() -> None:
    engine, repo, calls = _build_engine(
        [
            ValueError("Invalid conversation manager state."),
            RuntimeError("still broken"),
        ]
    )

    with pytest.raises(RuntimeError, match="still broken"):
        engine.analyze("prompt", "incident-2")

    assert calls["count"] == 2
    assert repo.deleted_sessions == ["conv-1:chat"]


def test_non_session_error_does_not_reset_state() -> None:
    engine, repo, calls = _build_engine(
        [
            RuntimeError("temporary failure"),
        ]
    )

    with pytest.raises(RuntimeError, match="temporary failure"):
        engine.analyze("prompt", "incident-3")

    assert calls["count"] == 1
    assert repo.deleted_sessions == []
