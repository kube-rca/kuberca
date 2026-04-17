from __future__ import annotations

import json
import logging
from typing import Any, cast

from app.clients.strands_agent import AnalysisEngine
from app.core.masking import Masker, RegexMasker
from app.schemas.chat import ChatRequest

logger = logging.getLogger(__name__)


def _is_invalid_session_restore_error(exc: BaseException) -> bool:
    """Detect Strands session restore errors caused by incompatible state."""
    stack: list[BaseException | None] = [exc]
    visited: set[int] = set()
    while stack:
        current = stack.pop()
        if current is None:
            continue
        marker = id(current)
        if marker in visited:
            continue
        visited.add(marker)
        if isinstance(current, ValueError) and "invalid conversation manager state" in str(
            current
        ).lower():
            return True
        stack.append(getattr(current, "__cause__", None))
        stack.append(getattr(current, "__context__", None))
        stack.append(getattr(current, "original_exception", None))
    return False


def _to_pretty_json(payload: dict[str, Any]) -> str:
    return json.dumps(payload, ensure_ascii=True, indent=2, sort_keys=True)


def _build_chat_prompt(request: ChatRequest, masker: Masker) -> str:
    """Build prompt for chat Q&A about an incident."""
    user_msg = masker.mask_text(request.message).strip() or "Tell me about this incident."
    language = "en" if (request.language or "").strip().lower() == "en" else "ko"
    response_language = "English" if language == "en" else "Korean"
    base = (
        "You are kube-rca-agent. A user is asking a question about an incident. "
        "Answer based on the incident/alert context provided and any prior analysis. "
        "You have tools for K8s, Prometheus, and Tempo—"
        "use them if needed to look up metrics, logs, or traces. "
        "For questions like 'What metric triggered this alert?', "
        "check context/artifacts for Prometheus queries (query field). "
        f"Respond concisely in {response_language}. "
        "Follow the selected UI language even if the user's question is in a different language.\n\n"
    )
    ctx = request.context
    if ctx:
        masked = cast(dict[str, Any], masker.mask_object(ctx))
        base += f"Incident/alert context:\n{_to_pretty_json(masked)}\n\n"
    base += f"User question: {user_msg}"
    return base


class ChatService:
    def __init__(
        self,
        analysis_engine: AnalysisEngine | None,
        masker: Masker | None = None,
    ) -> None:
        self._logger = logger
        self._analysis_engine = analysis_engine
        self._masker = masker or RegexMasker()

    def chat(self, request: ChatRequest) -> tuple[str, str | None]:
        """Answer user questions about an incident (name, id, content, metrics, etc.).

        Returns:
            tuple[str, str | None]: (reply, conversation_id)
        """
        if self._analysis_engine is None:
            return (
                self._masker.mask_text(
                    "Chat is unavailable because the analysis engine is not configured."
                ),
                request.conversation_id,
            )
        prompt = _build_chat_prompt(request, self._masker)
        session_id = (request.conversation_id or "default").strip() or "default"
        session_id = f"{session_id}:chat"
        try:
            reply = self._analysis_engine.analyze(prompt, session_id)
            if not isinstance(reply, str):
                reply = ""
            reply = self._masker.mask_text(reply).strip() or (
                "I couldn't generate a response. Please try rephrasing your question."
            )
            return reply, request.conversation_id
        except Exception as exc:  # noqa: BLE001
            self._logger.exception("Chat failed")
            if _is_invalid_session_restore_error(exc):
                return (
                    self._masker.mask_text(
                        "Failed to restore the conversation session. "
                        "Please start a new chat and try again."
                    ),
                    request.conversation_id,
                )
            return (
                self._masker.mask_text(
                    "An error occurred while processing chat. Please try again shortly."
                ),
                request.conversation_id,
            )
