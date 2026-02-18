from __future__ import annotations

from pydantic import BaseModel


class ChatRequest(BaseModel):
    """Request for chat Q&A. Matches AgentChatRequest from backend."""

    message: str
    conversation_id: str | None = None
    context: dict[str, object] | None = None
    metadata: dict[str, object] | None = None


class ChatResponse(BaseModel):
    """Response from chat. Matches AgentChatResponse from backend (multiple formats)."""

    status: str = "ok"
    answer: str | None = None
    message: str | None = None
    response: str | None = None
    output_text: str | None = None
    analysis: str | None = None
    conversation_id: str | None = None
