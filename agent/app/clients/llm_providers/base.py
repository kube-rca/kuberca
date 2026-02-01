"""Base types and protocols for LLM providers."""

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum
from typing import Protocol, runtime_checkable


class LLMProvider(str, Enum):
    """Supported LLM provider types."""

    GEMINI = "gemini"
    OPENAI = "openai"
    ANTHROPIC = "anthropic"

    @classmethod
    def from_string(cls, value: str) -> LLMProvider:
        """Convert string to LLMProvider enum.

        Args:
            value: Provider name (case-insensitive).

        Returns:
            LLMProvider enum value.

        Raises:
            ValueError: If provider is not supported.
        """
        normalized = value.lower().strip()
        for provider in cls:
            if provider.value == normalized:
                return provider
        supported = ", ".join(p.value for p in cls)
        raise ValueError(f"Unsupported provider '{value}'. Supported: {supported}")


@dataclass(frozen=True)
class ModelConfig:
    """Configuration for an LLM model.

    Attributes:
        provider: The LLM provider (gemini, openai, anthropic).
        model_id: The specific model identifier.
        api_key: API key for authentication.
        temperature: Optional sampling temperature (0.0-1.0).
        max_tokens: Optional maximum tokens in response.
    """

    provider: LLMProvider
    model_id: str
    api_key: str
    temperature: float | None = None
    max_tokens: int | None = None


@runtime_checkable
class StrandsModel(Protocol):
    """Protocol for Strands-compatible LLM models.

    This matches the interface expected by strands.Agent for the model parameter.
    """

    def __call__(self, messages: list, **kwargs: object) -> object:
        """Process messages and return response."""
        ...


# Default model IDs per provider
DEFAULT_MODEL_IDS: dict[LLMProvider, str] = {
    LLMProvider.GEMINI: "gemini-3-flash-preview",
    LLMProvider.OPENAI: "gpt-4o",
    LLMProvider.ANTHROPIC: "claude-sonnet-4-20250514",
}
