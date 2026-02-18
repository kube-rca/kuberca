from __future__ import annotations

from fastapi import APIRouter, Depends

from app.core.dependencies import get_chat_service
from app.schemas.chat import ChatRequest, ChatResponse
from app.services.chat import ChatService

router = APIRouter()


@router.post("/chat", response_model=ChatResponse)
def chat(
    request: ChatRequest,
    service: ChatService = Depends(get_chat_service),  # noqa: B008
) -> ChatResponse:
    """Answer user questions about an incident (name, id, content, metrics, etc.)."""
    reply, conversation_id = service.chat(request)
    return ChatResponse(
        status="ok",
        answer=reply,
        message=reply,
        response=reply,
        conversation_id=conversation_id,
    )
