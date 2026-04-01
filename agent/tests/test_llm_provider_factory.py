from __future__ import annotations

from typing import Any, cast

import pytest

from app.clients.llm_providers.base import LLMProvider, ModelConfig
from app.clients.llm_providers.factory import create_model, get_provider_config
from app.core.config import DEFAULT_ANTHROPIC_MAX_TOKENS, load_settings


def test_get_provider_config_sets_anthropic_max_tokens(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.setenv("AI_PROVIDER", "anthropic")
    monkeypatch.setenv("ANTHROPIC_API_KEY", "test-key")
    monkeypatch.setenv("ANTHROPIC_MODEL_ID", "claude-sonnet-4-6")
    monkeypatch.setenv("ANTHROPIC_MAX_TOKENS", "2048")

    config = get_provider_config(load_settings())

    assert config is not None
    assert config.provider == LLMProvider.ANTHROPIC
    assert config.model_id == "claude-sonnet-4-6"
    assert config.max_tokens == 2048


def test_create_model_formats_anthropic_request_with_default_max_tokens() -> None:
    model = cast(
        Any,
        create_model(
            ModelConfig(
                provider=LLMProvider.ANTHROPIC,
                model_id="claude-sonnet-4-6",
                api_key="test-key",
            )
        ),
    )

    request = model.format_request(
        messages=[{"role": "user", "content": [{"text": "hello"}]}],
    )

    assert request["model"] == "claude-sonnet-4-6"
    assert request["max_tokens"] == DEFAULT_ANTHROPIC_MAX_TOKENS
