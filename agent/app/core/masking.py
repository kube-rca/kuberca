from __future__ import annotations

import base64
import hashlib
import re
from dataclasses import dataclass, field
from re import Pattern
from typing import Any, Protocol, runtime_checkable

MASK_TOKEN = "[MASKED]"  # nosec B105 - placeholder token for masked output, not a credential

# ---------------------------------------------------------------------------
# Key denylist / allowlist for dict-key-based masking
# ---------------------------------------------------------------------------
_KEY_DENYLIST = frozenset(
    [
        "password",
        "token",
        "secret",
        "key",
        "credential",
        "authorization",
        "api_key",
        "apikey",
        "private_key",
        "access_key",
        "client_secret",
    ]
)

_KEY_ALLOWLIST = frozenset(
    [
        "secret_name",
        "token_budget",
        "service_account_token",
        "secret_ref",
        "secret_key_ref",
        "key_ref",
    ]
)

# ---------------------------------------------------------------------------
# Value heuristic allowlist — fields whose values look like base64/tokens
# but are not sensitive (e.g. K8s metadata fields)
# ---------------------------------------------------------------------------
_VALUE_HEURISTIC_SKIP_KEYS = frozenset(
    [
        "image",
        "imageid",
        "containerid",
        "resourceversion",
        "uid",
        "selflink",
        "generation",
    ]
)

# ---------------------------------------------------------------------------
# K8s annotation prefixes that are safe to keep unmasked
# ---------------------------------------------------------------------------
_SAFE_ANNOTATION_PREFIXES = (
    "app.kubernetes.io/",
    "kubernetes.io/",
    "helm.sh/",
    "meta.helm.sh/",
    "kubectl.kubernetes.io/",
)

# ---------------------------------------------------------------------------
# Value heuristic patterns — compiled once at module load
# ---------------------------------------------------------------------------
_VALUE_PATTERNS: list[Pattern[str]] = [
    # JWT tokens (3-segment base64url)
    re.compile(r"eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+"),
    # Bearer / Basic / Token / Digest auth values
    re.compile(r"(Basic|Bearer|Token|Digest)\s+[A-Za-z0-9+/=_.~-]{8,}"),
    # URL-embedded tokens
    re.compile(r"[?&](token|access_token|api_key|apikey|key)=[^\s&]+"),
    # Long base64 blobs (64+ chars) — handled separately with validation
]

_LONG_BASE64_RE = re.compile(r"[A-Za-z0-9+/=]{64,}")

# ---------------------------------------------------------------------------
# K8s-specific: sensitive command-line argument patterns
# ---------------------------------------------------------------------------
_SENSITIVE_ARG_RE = re.compile(
    r"(--(password|token|key|secret|credential|api[_-]?key|access[_-]?key|client[_-]?secret))=(\S+)"
)


def _is_valid_base64(text: str) -> bool:
    """Check if a string is plausibly base64-encoded content."""
    try:
        decoded = base64.b64decode(text, validate=True)
        # Must produce some non-trivial output
        return len(decoded) >= 16
    except Exception:
        return False


def _mask_replacement(hash_mode: bool, original: str) -> str:
    """Return the replacement string — either [MASKED] or [HASH:xxxxxxxx]."""
    if not hash_mode:
        return MASK_TOKEN
    digest = hashlib.sha256(original.encode("utf-8", errors="replace")).hexdigest()[:8]
    return f"[HASH:{digest}]"


# ---------------------------------------------------------------------------
# Masker Protocol — duck-typed interface for all masker implementations
# ---------------------------------------------------------------------------
@runtime_checkable
class Masker(Protocol):
    def mask_object(self, value: Any) -> Any: ...
    def mask_text(self, text: str) -> str: ...


# ---------------------------------------------------------------------------
# RegexMasker — existing user-defined pattern masker (unchanged)
# ---------------------------------------------------------------------------
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


# ---------------------------------------------------------------------------
# BuiltinRedactor — zero-config K8s-aware redaction rules
# ---------------------------------------------------------------------------
@dataclass(frozen=True)
class BuiltinRedactor:
    """Built-in redaction rules that work without any user configuration."""

    enabled: bool = True
    hash_mode: bool = False

    # -- Key denylist helpers ------------------------------------------------

    def _is_denied_key(self, key: str) -> bool:
        """Check if a dict key matches the denylist (substring match)."""
        lower = key.lower().replace("-", "_")
        # Check allowlist first — exact match
        if lower in _KEY_ALLOWLIST:
            return False
        return any(denied in lower for denied in _KEY_DENYLIST)

    # -- Value heuristic helpers ---------------------------------------------

    def _redact_text_value(self, text: str, *, parent_key: str = "") -> str:
        """Apply value heuristic patterns to a plain string."""
        if not text or len(text) < 8:
            return text

        # Skip heuristic for known-safe fields
        if parent_key.lower().replace("-", "_") in _VALUE_HEURISTIC_SKIP_KEYS:
            return text

        result = text

        # Apply regex-based patterns
        for pattern in _VALUE_PATTERNS:
            result = pattern.sub(lambda m: _mask_replacement(self.hash_mode, m.group(0)), result)

        # Long base64 — only mask if it's valid base64
        def _base64_replacer(m: re.Match[str]) -> str:
            candidate = m.group(0)
            if _is_valid_base64(candidate):
                return _mask_replacement(self.hash_mode, candidate)
            return candidate

        result = _LONG_BASE64_RE.sub(_base64_replacer, result)

        return result

    # -- K8s-specific helpers ------------------------------------------------

    def _redact_env_list(self, items: list[Any]) -> list[Any]:
        """Redact value fields in env-entry-shaped list items."""
        out: list[Any] = []
        for item in items:
            if (
                isinstance(item, dict)
                and "name" in item
                and "value" in item
                and isinstance(item["value"], str)
            ):
                redacted = dict(item)
                redacted["value"] = _mask_replacement(self.hash_mode, item["value"])
                out.append(redacted)
            else:
                out.append(self._redact_value(item))
        return out

    def _redact_command_args(self, items: list[Any]) -> list[Any]:
        """Redact sensitive flags like --password=xxx in command/args lists."""
        out: list[Any] = []
        for item in items:
            if isinstance(item, str):
                redacted = _SENSITIVE_ARG_RE.sub(
                    lambda m: f"{m.group(1)}={_mask_replacement(self.hash_mode, m.group(3))}",
                    item,
                )
                out.append(redacted)
            else:
                out.append(self._redact_value(item))
        return out

    def _redact_annotations(self, annotations: dict[str, Any]) -> dict[str, Any]:
        """Mask annotation values unless they have a known-safe prefix."""
        out: dict[str, Any] = {}
        for key, val in annotations.items():
            if isinstance(val, str) and not any(
                key.startswith(prefix) for prefix in _SAFE_ANNOTATION_PREFIXES
            ):
                out[key] = _mask_replacement(self.hash_mode, val)
            else:
                out[key] = val
        return out

    # -- Core traversal ------------------------------------------------------

    def _redact_value(self, value: Any, *, parent_key: str = "") -> Any:
        """Recursively redact a value."""
        if isinstance(value, str):
            return self._redact_text_value(value, parent_key=parent_key)
        if isinstance(value, list):
            # Check for K8s-specific list patterns
            if parent_key.lower() in ("env", "envfrom"):
                return self._redact_env_list(value)
            if parent_key.lower() in ("command", "args"):
                return self._redact_command_args(value)
            return [self._redact_value(item) for item in value]
        if isinstance(value, tuple):
            return tuple(self._redact_value(item) for item in value)
        if isinstance(value, dict):
            return self._redact_dict(value, parent_key=parent_key)
        return value

    def _redact_dict(self, d: dict[str, Any], *, parent_key: str = "") -> dict[str, Any]:
        """Redact a dictionary applying key denylist and recursion."""
        out: dict[str, Any] = {}
        for key, val in d.items():
            # Special handling for annotations
            if key.lower() in ("annotations",) and isinstance(val, dict):
                out[key] = self._redact_annotations(val)
                continue

            # Key denylist — mask the entire value
            if self._is_denied_key(key) and isinstance(val, (str, int, float)):
                out[key] = _mask_replacement(self.hash_mode, str(val))
                continue

            # Recurse
            out[key] = self._redact_value(val, parent_key=key)
        return out

    # -- Public interface (matches Masker protocol) --------------------------

    def redact_object(self, value: Any) -> Any:
        """Redact sensitive data from an arbitrary object tree."""
        if not self.enabled:
            return value
        return self._redact_value(value)

    def redact_text(self, text: str) -> str:
        """Redact sensitive patterns from a plain string."""
        if not self.enabled:
            return text
        return self._redact_text_value(text)

    # Masker protocol aliases
    def mask_object(self, value: Any) -> Any:
        return self.redact_object(value)

    def mask_text(self, text: str) -> str:
        return self.redact_text(text)


# ---------------------------------------------------------------------------
# ChainedMasker — BuiltinRedactor -> RegexMasker pipeline
# ---------------------------------------------------------------------------
@dataclass(frozen=True)
class ChainedMasker:
    """Chains BuiltinRedactor (first) and RegexMasker (second)."""

    builtin: BuiltinRedactor
    regex: RegexMasker

    def mask_object(self, value: Any) -> Any:
        return self.regex.mask_object(self.builtin.redact_object(value))

    def mask_text(self, text: str) -> str:
        return self.regex.mask_text(self.builtin.redact_text(text))


# ---------------------------------------------------------------------------
# Factory
# ---------------------------------------------------------------------------
def build_masker(patterns: list[str]) -> RegexMasker:
    return RegexMasker.from_patterns(patterns)
