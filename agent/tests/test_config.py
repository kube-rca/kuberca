from __future__ import annotations

import pytest

from app.core.config import load_settings


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
