from __future__ import annotations

import pytest

from app.core.log_sanitize import sanitize_log


class TestSanitizeLog:
    def test_empty_string(self) -> None:
        assert sanitize_log("") == ""

    def test_plain_string_is_unchanged(self) -> None:
        assert sanitize_log("provider=gemini") == "provider=gemini"

    def test_strips_lf(self) -> None:
        forged = "gemini\nINFO: forged log entry"
        result = sanitize_log(forged)
        assert "\n" not in result
        assert result == "gemini INFO: forged log entry"

    def test_strips_cr(self) -> None:
        forged = "gemini\rINFO: forged log entry"
        result = sanitize_log(forged)
        assert "\r" not in result
        assert result == "gemini INFO: forged log entry"

    def test_strips_nul(self) -> None:
        forged = "gemini\x00truncated"
        result = sanitize_log(forged)
        assert "\x00" not in result
        assert result == "gemini truncated"

    def test_strips_mixed_control_chars(self) -> None:
        forged = "a\r\nb\x00c\nd"
        result = sanitize_log(forged)
        assert "\r" not in result
        assert "\n" not in result
        assert "\x00" not in result
        # Each control char becomes a single space, so 4 control chars -> 4 spaces
        assert result == "a  b c d"

    def test_preserves_other_whitespace(self) -> None:
        # Tabs and ordinary spaces must not be stripped
        assert sanitize_log("a\tb c") == "a\tb c"

    def test_non_string_int(self) -> None:
        assert sanitize_log(42) == "42"

    def test_non_string_dict(self) -> None:
        # dict.__str__ output should be coerced and contain no control bytes
        result = sanitize_log({"key": "value"})
        assert "\n" not in result
        assert "\r" not in result
        assert "\x00" not in result
        assert "key" in result
        assert "value" in result

    def test_non_string_none(self) -> None:
        assert sanitize_log(None) == "None"

    def test_non_string_with_embedded_lf_via_str(self) -> None:
        class Custom:
            def __str__(self) -> str:
                return "hello\nworld"

        assert sanitize_log(Custom()) == "hello world"


@pytest.mark.parametrize(
    "value,expected",
    [
        ("\n", " "),
        ("\r", " "),
        ("\x00", " "),
        ("\r\n", "  "),
        ("\n\n\n", "   "),
    ],
)
def test_sanitize_log_parametrized_control_only(value: str, expected: str) -> None:
    assert sanitize_log(value) == expected
