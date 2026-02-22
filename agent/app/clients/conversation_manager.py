"""Safe sliding-window conversation manager that fixes Gemini turn-order bugs.

Workaround for two defects in Strands SDK ``SlidingWindowConversationManager``:

1. ``_truncate_tool_results()`` returns ``False`` as soon as it encounters an
   already-truncated toolResult, instead of scanning for the next truncatable one.
2. After window trimming, the remaining messages may start with an *assistant*
   turn (containing ``toolUse``), which violates Gemini's constraint that a
   function-call turn must be preceded by a user turn.
"""

from __future__ import annotations

import logging
from typing import Any

from strands.agent.agent import Agent
from strands.agent.conversation_manager.sliding_window_conversation_manager import (
    SlidingWindowConversationManager,
)
from strands.types.content import Messages
from strands.types.exceptions import ContextWindowOverflowException

logger = logging.getLogger(__name__)

_TOOL_RESULT_PLACEHOLDER = "The tool result was too large!"


class SafeSlidingWindowConversationManager(SlidingWindowConversationManager):
    """Drop-in replacement that guarantees messages start with a ``user`` turn."""

    # ------------------------------------------------------------------
    # Fix 1: reduce_context post-processing
    # ------------------------------------------------------------------

    def reduce_context(self, agent: Agent, e: Exception | None = None, **kwargs: Any) -> None:
        """Trim context via parent, then ensure first message has ``user`` role."""
        super().reduce_context(agent, e, **kwargs)
        self._ensure_user_turn_first(agent.messages, e)

    def _ensure_user_turn_first(self, messages: Messages, cause: Exception | None = None) -> None:
        """Remove leading non-user messages so the conversation starts with a user turn."""
        while messages and messages[0].get("role") != "user":
            messages.pop(0)
            self.removed_message_count += 1
            logger.debug("Dropped leading non-user message to fix turn order")

        if not messages:
            raise ContextWindowOverflowException(
                "Unable to trim conversation context: no user messages remaining"
            ) from cause

    # ------------------------------------------------------------------
    # Fix 2: _truncate_tool_results scans all messages instead of early return
    # ------------------------------------------------------------------

    def _truncate_tool_results(self, messages: Messages, msg_idx: int) -> bool:
        """Truncate tool results, skipping already-truncated ones.

        The parent implementation returns ``False`` when it encounters a
        toolResult that was already replaced with the placeholder.  This
        override continues scanning backwards through older messages so that
        *another* truncatable toolResult can be found.
        """
        for idx in range(msg_idx, -1, -1):
            if idx >= len(messages) or idx < 0:
                continue

            message = messages[idx]
            changes_made = False

            for i, content in enumerate(message.get("content", [])):
                if not (isinstance(content, dict) and "toolResult" in content):
                    continue

                tool_result_text = next(
                    (item["text"] for item in content["toolResult"]["content"] if "text" in item),
                    "",
                )

                if (
                    content["toolResult"].get("status") == "error"
                    and tool_result_text == _TOOL_RESULT_PLACEHOLDER
                ):
                    # Already truncated — skip and keep scanning
                    continue

                message["content"][i]["toolResult"]["status"] = "error"
                message["content"][i]["toolResult"]["content"] = [
                    {"text": _TOOL_RESULT_PLACEHOLDER}
                ]
                changes_made = True

            if changes_made:
                logger.debug("Truncated tool results in message at index %d", idx)
                return True

        return False
