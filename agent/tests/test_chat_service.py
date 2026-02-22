from __future__ import annotations

from app.schemas.chat import ChatRequest
from app.services.chat import ChatService


class RaisingAnalysisEngine:
    def __init__(self, exc: Exception) -> None:
        self._exc = exc

    def analyze(self, prompt: str, incident_id: str | None = None) -> str:
        raise self._exc


def test_chat_service_masks_invalid_session_restore_error() -> None:
    service = ChatService(
        analysis_engine=RaisingAnalysisEngine(
            ValueError("Invalid conversation manager state.")
        )
    )
    request = ChatRequest(message="What happened?", conversation_id="conv-1")

    reply, conversation_id = service.chat(request)

    assert conversation_id == "conv-1"
    assert "Failed to restore the conversation session." in reply
    assert "Invalid conversation manager state." not in reply


def test_chat_service_returns_generic_message_for_other_errors() -> None:
    service = ChatService(analysis_engine=RaisingAnalysisEngine(RuntimeError("db password leaked")))
    request = ChatRequest(message="Retry please", conversation_id="conv-2")

    reply, conversation_id = service.chat(request)

    assert conversation_id == "conv-2"
    assert reply == "An error occurred while processing chat. Please try again shortly."
