"""Factory for creating LLM model instances."""

from __future__ import annotations

import logging
from typing import Any

from app.clients.llm_providers.base import (
    DEFAULT_MODEL_IDS,
    LLMProvider,
    ModelConfig,
    StrandsModel,
)
from app.core.config import Settings

logger = logging.getLogger(__name__)


def create_model(config: ModelConfig) -> StrandsModel:
    """Create a Strands-compatible model instance.

    Args:
        config: Model configuration including provider, model_id, and api_key.

    Returns:
        A Strands-compatible model instance.

    Raises:
        ValueError: If provider is not supported or API key is missing.
        ImportError: If required provider package is not installed.
    """
    if not config.api_key:
        raise ValueError(f"API key is required for provider '{config.provider.value}'")

    if config.provider == LLMProvider.GEMINI:
        return _create_gemini_model(config)
    elif config.provider == LLMProvider.OPENAI:
        return _create_openai_model(config)
    elif config.provider == LLMProvider.ANTHROPIC:
        return _create_anthropic_model(config)
    else:
        raise ValueError(f"Unsupported provider: {config.provider}")


def _create_gemini_model(config: ModelConfig) -> StrandsModel:
    """Create a Gemini model instance."""
    try:
        from strands.models.gemini import GeminiModel
    except ImportError as e:
        raise ImportError(
            "Gemini support requires 'strands-agents[gemini]'. "
            "Install with: pip install 'strands-agents[gemini]'"
        ) from e

    model_kwargs: dict[str, Any] = {
        "client_args": {"api_key": config.api_key},
        "model_id": config.model_id,
    }
    return GeminiModel(**model_kwargs)


def _create_openai_model(config: ModelConfig) -> StrandsModel:
    """Create an OpenAI model instance."""
    try:
        from strands.models.openai import OpenAIModel
    except ImportError as e:
        raise ImportError(
            "OpenAI support requires 'strands-agents[openai]'. "
            "Install with: pip install 'strands-agents[openai]'"
        ) from e

    model_kwargs: dict[str, Any] = {
        "client_args": {"api_key": config.api_key},
        "model_id": config.model_id,
    }
    if config.temperature is not None:
        model_kwargs["temperature"] = config.temperature
    if config.max_tokens is not None:
        model_kwargs["max_tokens"] = config.max_tokens

    return OpenAIModel(**model_kwargs)


def _create_anthropic_model(config: ModelConfig) -> StrandsModel:
    """Create an Anthropic (Claude) model instance."""
    try:
        from strands.models.anthropic import AnthropicModel
    except ImportError as e:
        raise ImportError(
            "Anthropic support requires 'strands-agents[anthropic]'. "
            "Install with: pip install 'strands-agents[anthropic]'"
        ) from e

    model_kwargs: dict[str, Any] = {
        "client_args": {"api_key": config.api_key},
        "model_id": config.model_id,
    }
    if config.temperature is not None:
        model_kwargs["temperature"] = config.temperature
    if config.max_tokens is not None:
        model_kwargs["max_tokens"] = config.max_tokens

    return AnthropicModel(**model_kwargs)


def get_provider_config(settings: Settings) -> ModelConfig | None:
    """Build ModelConfig from application settings.

    Determines the appropriate provider and API key based on settings.
    Supports fallback to Gemini for backward compatibility.

    Args:
        settings: Application settings.

    Returns:
        ModelConfig if a valid provider is configured, None otherwise.
    """
    provider_str = settings.ai_provider.lower()

    # Resolve provider and API key
    if provider_str == "gemini":
        api_key = settings.gemini_api_key
        default_model = DEFAULT_MODEL_IDS[LLMProvider.GEMINI]
    elif provider_str == "openai":
        api_key = settings.openai_api_key
        default_model = DEFAULT_MODEL_IDS[LLMProvider.OPENAI]
    elif provider_str == "anthropic":
        api_key = settings.anthropic_api_key
        default_model = DEFAULT_MODEL_IDS[LLMProvider.ANTHROPIC]
    else:
        logger.warning("Unknown AI provider '%s', falling back to gemini", provider_str)
        api_key = settings.gemini_api_key
        provider_str = "gemini"
        default_model = DEFAULT_MODEL_IDS[LLMProvider.GEMINI]

    if not api_key:
        logger.warning(
            "No API key configured for provider '%s'. Analysis engine will be disabled.",
            provider_str,
        )
        return None

    # Use configured model_id or default
    model_id = settings.ai_model_id or default_model

    try:
        provider = LLMProvider.from_string(provider_str)
    except ValueError:
        logger.error("Invalid provider: %s", provider_str)
        return None

    return ModelConfig(
        provider=provider,
        model_id=model_id,
        api_key=api_key,
    )
