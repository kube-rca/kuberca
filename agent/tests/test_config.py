from __future__ import annotations

import pytest

from app.core.config import (
    DEFAULT_ANTHROPIC_MAX_TOKENS,
    DEFAULT_LANGUAGE,
    load_settings,
    normalize_language,
)


def test_load_settings_parses_masking_regex_list_json(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv(
        "MASKING_REGEX_LIST_JSON",
        '["token-\\\\d+", "secret-[A-Za-z]+", "  "]',
    )

    settings = load_settings()

    assert settings.masking_regex_list == [r"token-\d+", "secret-[A-Za-z]+"]


def test_load_settings_rejects_non_array_masking_regex_json(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.setenv("MASKING_REGEX_LIST_JSON", '{"regex": "token-\\\\d+"}')

    with pytest.raises(ValueError, match="MASKING_REGEX_LIST_JSON"):
        load_settings()


def test_load_settings_rejects_non_string_masking_regex_item(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.setenv("MASKING_REGEX_LIST_JSON", '["token-\\\\d+", 1]')

    with pytest.raises(ValueError, match="MASKING_REGEX_LIST_JSON\\[1\\]"):
        load_settings()


def test_load_settings_rejects_invalid_regex_pattern(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("MASKING_REGEX_LIST_JSON", r'["(unclosed"]')

    with pytest.raises(ValueError, match="MASKING_REGEX_LIST_JSON\\[0\\]"):
        load_settings()


def test_load_settings_uses_exponential_llm_retry_defaults(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.delenv("LLM_RETRY_MAX_ATTEMPTS", raising=False)
    monkeypatch.delenv("LLM_RETRY_MAX_WAIT", raising=False)
    monkeypatch.delenv("LLM_RETRY_TOTAL_TIMEOUT", raising=False)

    settings = load_settings()

    assert settings.llm_retry_max_attempts == 10
    assert settings.llm_retry_max_wait == 30.0
    assert settings.llm_retry_total_timeout == 180.0


def test_load_settings_uses_anthropic_max_tokens_default(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    monkeypatch.delenv("ANTHROPIC_MAX_TOKENS", raising=False)

    settings = load_settings()

    assert settings.anthropic_max_tokens == DEFAULT_ANTHROPIC_MAX_TOKENS


@pytest.mark.parametrize("value", ["abc", "0", "-1"])
def test_load_settings_falls_back_for_invalid_anthropic_max_tokens(
    monkeypatch: pytest.MonkeyPatch, value: str
) -> None:
    monkeypatch.setenv("ANTHROPIC_MAX_TOKENS", value)

    settings = load_settings()

    assert settings.anthropic_max_tokens == DEFAULT_ANTHROPIC_MAX_TOKENS


def test_load_settings_language_defaults_to_en(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.delenv("LANGUAGE", raising=False)

    settings = load_settings()

    assert settings.language == DEFAULT_LANGUAGE == "en"


@pytest.mark.parametrize(
    "raw, expected",
    [("ko", "ko"), ("en", "en"), ("KO", "ko"), ("  En  ", "en")],
)
def test_load_settings_language_accepts_supported_values(
    monkeypatch: pytest.MonkeyPatch, raw: str, expected: str
) -> None:
    monkeypatch.setenv("LANGUAGE", raw)

    settings = load_settings()

    assert settings.language == expected


@pytest.mark.parametrize("value", ["fr", "ja", "", "english"])
def test_load_settings_language_invalid_falls_back_to_en(
    monkeypatch: pytest.MonkeyPatch, value: str
) -> None:
    monkeypatch.setenv("LANGUAGE", value)

    settings = load_settings()

    assert settings.language == "en"


def test_normalize_language_uses_default_for_invalid_value() -> None:
    assert normalize_language("fr", default="ko") == "ko"
    assert normalize_language(None, default="ko") == "ko"
    # Invalid default itself falls back to "en"
    assert normalize_language("fr", default="zz") == "en"
