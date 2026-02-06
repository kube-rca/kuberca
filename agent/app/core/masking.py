from __future__ import annotations

import re
from dataclasses import dataclass, field
from re import Pattern
from typing import Any

MASK_TOKEN = "[MASKED]"


@dataclass(frozen=True)
class RegexMasker:
    patterns: tuple[Pattern[str], ...] = field(default_factory=tuple)

    @classmethod
    def from_patterns(cls, patterns: list[str]) -> RegexMasker:
        compiled: list[Pattern[str]] = []
        for idx, pattern in enumerate(patterns):
            try:
                compiled.append(re.compile(pattern))
            except re.error as exc:
                raise ValueError(f"invalid masking regex at index {idx}: {pattern}") from exc
        return cls(patterns=tuple(compiled))

    def mask_text(self, text: str) -> str:
        if not text:
            return text
        masked = text
        for pattern in self.patterns:
            masked = pattern.sub(MASK_TOKEN, masked)
        return masked

    def mask_object(self, value: Any) -> Any:
        if isinstance(value, str):
            return self.mask_text(value)
        if isinstance(value, list):
            return [self.mask_object(item) for item in value]
        if isinstance(value, tuple):
            return tuple(self.mask_object(item) for item in value)
        if isinstance(value, dict):
            return {key: self.mask_object(item) for key, item in value.items()}
        return value


def build_masker(patterns: list[str]) -> RegexMasker:
    return RegexMasker.from_patterns(patterns)
