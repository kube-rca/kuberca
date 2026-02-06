from __future__ import annotations

import pytest

from app.core.masking import MASK_TOKEN, build_masker


def test_regex_masker_masks_nested_values() -> None:
    masker = build_masker([r"token-\d+"])
    payload = {
        "summary": "token-123 found",
        "logs": ["ok", "token-999"],
        "nested": {"detail": "value token-456"},
    }

    masked = masker.mask_object(payload)

    assert masked == {
        "summary": f"{MASK_TOKEN} found",
        "logs": ["ok", MASK_TOKEN],
        "nested": {"detail": f"value {MASK_TOKEN}"},
    }


def test_build_masker_rejects_invalid_regex() -> None:
    with pytest.raises(ValueError, match="invalid masking regex"):
        build_masker([r"(unclosed"])
