"""LLM Provider abstraction for multi-provider support.

This module provides a unified interface for different LLM providers:
- Gemini (Google)
- OpenAI (GPT-4)
- Anthropic (Claude)

Usage:
    from app.clients.llm_providers import create_model, LLMProvider

    model = create_model(
        provider="gemini",
        model_id="gemini-3-flash-preview",
        api_key="your-api-key"
    )
"""

from app.clients.llm_providers.base import LLMProvider, ModelConfig
from app.clients.llm_providers.factory import create_model, get_provider_config

__all__ = [
    "LLMProvider",
    "ModelConfig",
    "create_model",
    "get_provider_config",
]
