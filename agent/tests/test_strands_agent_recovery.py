from __future__ import annotations

from collections import OrderedDict
from contextlib import contextmanager
from threading import Lock

import pytest

from app.clients.strands_agent import (
    StrandsAnalysisEngine,
    _has_user_text,
    _sanitize_message_order,
    _strip_tool_blocks,
)


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


# ---------------------------------------------------------------------------
# Gemini turn-order recovery (session reset via analyze)
# ---------------------------------------------------------------------------


def _gemini_turn_order_error() -> Exception:
    """Simulate a Gemini 400 turn-order error."""

    class _FakeGeminiClientError(Exception):
        __module__ = "google.genai.errors"

        def __init__(self) -> None:
            self.status_code = 400
            super().__init__(
                "400 INVALID_ARGUMENT. Please ensure that function response "
                "turn comes immediately after a function call turn."
            )

    return _FakeGeminiClientError()


def test_gemini_turn_order_error_triggers_session_reset_and_retry() -> None:
    engine, repo, calls = _build_engine(
        [
            _gemini_turn_order_error(),
            "recovered-after-reset",
        ]
    )

    result = engine.analyze("prompt", "incident-4")

    assert result == "recovered-after-reset"
    assert calls["count"] == 2
    assert repo.deleted_sessions == ["conv-1:chat"]


def test_gemini_turn_order_recovery_fails_then_raises() -> None:
    engine, repo, calls = _build_engine(
        [
            _gemini_turn_order_error(),
            RuntimeError("still broken after reset"),
        ]
    )

    with pytest.raises(RuntimeError, match="still broken after reset"):
        engine.analyze("prompt", "incident-5")

    assert calls["count"] == 2
    assert repo.deleted_sessions == ["conv-1:chat"]


# ---------------------------------------------------------------------------
# _sanitize_message_order unit tests
# ---------------------------------------------------------------------------


def _user_text(text: str) -> dict:
    return {"role": "user", "content": [{"text": text}]}


def _assistant_text(text: str) -> dict:
    return {"role": "assistant", "content": [{"text": text}]}


def _assistant_tool_use(tool_use_id: str, name: str = "get_pod") -> dict:
    return {
        "role": "assistant",
        "content": [
            {"toolUse": {"toolUseId": tool_use_id, "name": name, "input": {}}},
        ],
    }


def _user_tool_result(tool_use_id: str, text: str = "ok") -> dict:
    return {
        "role": "user",
        "content": [
            {
                "toolResult": {
                    "toolUseId": tool_use_id,
                    "content": [{"text": text}],
                    "status": "success",
                }
            },
        ],
    }


def test_sanitize_removes_orphaned_tool_result_at_start() -> None:
    """toolResult without preceding toolUse should be removed.

    Phase 2 strips the orphan → empty message dropped → [assistant, user_text].
    Phase 3 strips leading non-user-text → [user_text].
    """
    messages = [
        _user_tool_result("orphan-1"),
        _assistant_text("response"),
        _user_text("follow-up"),
    ]
    _sanitize_message_order(messages)

    assert len(messages) == 1
    assert messages[0]["content"][0]["text"] == "follow-up"


def test_sanitize_removes_orphaned_tool_use_at_end() -> None:
    """toolUse without matching toolResult should be removed."""
    messages = [
        _user_text("hello"),
        _assistant_tool_use("orphan-2"),
    ]
    _sanitize_message_order(messages)

    # orphan toolUse removed → assistant message becomes empty → dropped
    assert len(messages) == 1
    assert messages[0]["content"][0]["text"] == "hello"


def test_sanitize_keeps_paired_tool_messages() -> None:
    """Properly paired toolUse/toolResult should remain."""
    messages = [
        _user_text("check pod"),
        _assistant_tool_use("paired-1"),
        _user_tool_result("paired-1"),
        _assistant_text("Pod is healthy"),
    ]
    _sanitize_message_order(messages)

    assert len(messages) == 4


def test_sanitize_mixed_orphans_and_pairs() -> None:
    """Only orphans are removed; paired messages stay."""
    messages = [
        _user_tool_result("orphan-a"),  # orphan result (no matching use)
        _user_text("question"),
        _assistant_tool_use("paired-b"),
        _user_tool_result("paired-b"),
        _assistant_text("answer"),
        _assistant_tool_use("orphan-c"),  # orphan use (no matching result)
    ]
    _sanitize_message_order(messages)

    # orphan-a result removed (empty msg dropped), orphan-c use removed (empty msg dropped)
    assert len(messages) == 4
    assert messages[0]["content"][0]["text"] == "question"


def test_sanitize_strips_leading_user_tool_result_only_message() -> None:
    """A user message with only toolResult (no text) at start should be stripped."""
    messages = [
        _user_tool_result("paired-1"),
        _assistant_text("response"),
        _user_text("next question"),
    ]
    # paired-1 has no matching toolUse → orphan → removed
    # Then Phase 3: first is assistant → stripped
    _sanitize_message_order(messages)

    assert len(messages) == 1
    assert messages[0]["content"][0]["text"] == "next question"


def test_sanitize_noop_on_clean_history() -> None:
    """Clean conversation should not be modified."""
    messages = [
        _user_text("hello"),
        _assistant_text("hi"),
        _user_text("bye"),
    ]
    _sanitize_message_order(messages)
    assert len(messages) == 3


def test_has_user_text_true() -> None:
    assert _has_user_text(_user_text("hello")) is True


def test_has_user_text_false_for_tool_result_only() -> None:
    assert _has_user_text(_user_tool_result("id-1")) is False


def test_has_user_text_false_for_assistant() -> None:
    assert _has_user_text(_assistant_text("hi")) is False


def test_strip_tool_blocks_removes_specified_ids() -> None:
    messages = [
        {
            "role": "assistant",
            "content": [
                {"text": "Let me check"},
                {"toolUse": {"toolUseId": "keep-1", "name": "x", "input": {}}},
                {"toolUse": {"toolUseId": "remove-1", "name": "y", "input": {}}},
            ],
        },
    ]
    _strip_tool_blocks(messages, orphan_use_ids={"remove-1"}, orphan_result_ids=set())

    assert len(messages) == 1
    assert len(messages[0]["content"]) == 2
    assert messages[0]["content"][1]["toolUse"]["toolUseId"] == "keep-1"
