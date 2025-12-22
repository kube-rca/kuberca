from __future__ import annotations

import base64
import binascii
from collections.abc import Callable
from typing import Any

from strands.event_loop import streaming as streaming_module
from strands.models import gemini as gemini_module

_PATCH_APPLIED = False
_SIGNATURE_B64_PREFIX = "base64:"


def _encode_signature(signature: Any) -> str:
    if isinstance(signature, (bytes, bytearray)):
        try:
            return signature.decode("utf-8")
        except UnicodeDecodeError:
            encoded = base64.b64encode(signature).decode("ascii")
            return f"{_SIGNATURE_B64_PREFIX}{encoded}"
    return str(signature)


def _decode_signature(signature: Any) -> bytes:
    if isinstance(signature, (bytes, bytearray)):
        return bytes(signature)
    signature_text = str(signature)
    if signature_text.startswith(_SIGNATURE_B64_PREFIX):
        encoded = signature_text[len(_SIGNATURE_B64_PREFIX) :]
        try:
            return base64.b64decode(encoded)
        except (binascii.Error, ValueError):
            return signature_text.encode("utf-8")
    return signature_text.encode("utf-8")


def apply_gemini_thought_signature_patch() -> None:
    """Propagate Gemini thought signatures for tool calls.

    This keeps toolUse blocks carrying thought signatures so Gemini accepts
    follow-up requests that include function_call parts.
    """
    global _PATCH_APPLIED
    if _PATCH_APPLIED:
        return
    _PATCH_APPLIED = True

    _patch_gemini_format_chunk()
    _patch_gemini_request_formatter()
    _patch_streaming_tool_signature()


def _patch_gemini_format_chunk() -> None:
    GeminiModel = gemini_module.GeminiModel
    original_format_chunk: Callable[
        [Any, dict[str, Any]],
        dict[str, Any],
    ] = GeminiModel._format_chunk

    def patched_format_chunk(self: Any, event: dict[str, Any]) -> dict[str, Any]:
        chunk = original_format_chunk(self, event)
        if event.get("chunk_type") == "content_start" and event.get("data_type") == "tool":
            part = event.get("data")
            signature = getattr(part, "thought_signature", None)
            if signature:
                signature_text = _encode_signature(signature)
                tool_use = (
                    chunk.get("contentBlockStart", {})
                    .get("start", {})
                    .get("toolUse")
                )
                if tool_use is not None:
                    tool_use["thoughtSignature"] = signature_text
        return chunk

    GeminiModel._format_chunk = patched_format_chunk  # type: ignore[assignment]


def _patch_gemini_request_formatter() -> None:
    GeminiModel = gemini_module.GeminiModel
    original_formatter: Callable[
        [Any, dict[str, Any]],
        Any,
    ] = GeminiModel._format_request_content_part

    def patched_formatter(self: Any, content: dict[str, Any]) -> Any:
        if "toolUse" in content:
            tool_use = content["toolUse"]
            signature = tool_use.get("thoughtSignature") or tool_use.get("thought_signature")
            if signature:
                signature_bytes = _decode_signature(signature)
                return gemini_module.genai.types.Part(
                    function_call=gemini_module.genai.types.FunctionCall(
                        args=tool_use["input"],
                        id=tool_use["toolUseId"],
                        name=tool_use["name"],
                    ),
                    thought_signature=signature_bytes,
                )
        return original_formatter(self, content)

    GeminiModel._format_request_content_part = patched_formatter  # type: ignore[assignment]


def _patch_streaming_tool_signature() -> None:
    original_handle_content_block_start = streaming_module.handle_content_block_start
    original_handle_content_block_stop = streaming_module.handle_content_block_stop

    def patched_handle_content_block_start(event: dict[str, Any]) -> dict[str, Any]:
        current_tool_use = original_handle_content_block_start(event)
        start = event.get("start", {})
        tool_use = start.get("toolUse")
        if tool_use and tool_use.get("thoughtSignature"):
            signature_text = _encode_signature(tool_use["thoughtSignature"])
            current_tool_use["thoughtSignature"] = signature_text
        return current_tool_use

    def patched_handle_content_block_stop(state: dict[str, Any]) -> dict[str, Any]:
        signature = state.get("current_tool_use", {}).get("thoughtSignature")
        new_state = original_handle_content_block_stop(state)
        if signature and new_state.get("content"):
            last_block = new_state["content"][-1]
            if "toolUse" in last_block:
                last_block["toolUse"]["thoughtSignature"] = signature
        return new_state

    streaming_module.handle_content_block_start = patched_handle_content_block_start
    streaming_module.handle_content_block_stop = patched_handle_content_block_stop
