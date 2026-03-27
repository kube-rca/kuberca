from __future__ import annotations

import pytest

from app.core.masking import (
    MASK_TOKEN,
    BuiltinRedactor,
    ChainedMasker,
    RegexMasker,
    build_masker,
)


# ---------------------------------------------------------------------------
# Existing RegexMasker tests
# ---------------------------------------------------------------------------


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


# ---------------------------------------------------------------------------
# BuiltinRedactor — Key denylist tests
# ---------------------------------------------------------------------------


def test_builtin_redacts_password_key() -> None:
    """Keys containing 'password' have their values masked."""
    redactor = BuiltinRedactor()
    data = {"db_password": "hunter2", "username": "admin"}
    result = redactor.redact_object(data)
    assert result["db_password"] == MASK_TOKEN
    assert result["username"] == "admin"


def test_builtin_redacts_access_key_id() -> None:
    """Keys containing 'access_key' have their values masked."""
    redactor = BuiltinRedactor()
    data = {"access_key_id": "AKIAIOSFODNN7EXAMPLE"}
    result = redactor.redact_object(data)
    assert result["access_key_id"] == MASK_TOKEN


def test_builtin_redacts_api_key_variations() -> None:
    """Various key naming patterns for API keys are caught."""
    redactor = BuiltinRedactor()
    data = {
        "api_key": "sk-abc123",
        "apikey": "sk-def456",
        "x-api-key": "sk-ghi789",
    }
    result = redactor.redact_object(data)
    for key in data:
        assert result[key] == MASK_TOKEN, f"{key} should be masked"


def test_builtin_preserves_secret_name_key() -> None:
    """'secret_name' is in allowlist — value should NOT be masked."""
    redactor = BuiltinRedactor()
    data = {"secret_name": "my-tls-cert", "secret_key": "topsecret"}
    result = redactor.redact_object(data)
    assert result["secret_name"] == "my-tls-cert"
    assert result["secret_key"] == MASK_TOKEN


def test_builtin_preserves_token_budget_key() -> None:
    """'token_budget' is in allowlist — value should NOT be masked."""
    redactor = BuiltinRedactor()
    data = {"token_budget": 32000}
    result = redactor.redact_object(data)
    assert result["token_budget"] == 32000


# ---------------------------------------------------------------------------
# BuiltinRedactor — Value heuristic tests
# ---------------------------------------------------------------------------


def test_builtin_redacts_jwt_in_logs() -> None:
    """JWT tokens in log strings are masked."""
    redactor = BuiltinRedactor()
    jwt = "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"
    text = f"auth token: {jwt}"
    result = redactor.redact_text(text)
    assert "eyJ" not in result
    assert MASK_TOKEN in result


def test_builtin_redacts_bearer_token() -> None:
    """'Bearer eyJ...' patterns are masked."""
    redactor = BuiltinRedactor()
    text = "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIn0.rz8"
    result = redactor.redact_text(text)
    assert "eyJ" not in result
    assert MASK_TOKEN in result


def test_builtin_redacts_url_embedded_token() -> None:
    """'?token=xxx' and '&api_key=xxx' patterns are masked."""
    redactor = BuiltinRedactor()
    text = "https://api.example.com/v1?token=sk-abcdef123456&api_key=mykey123"
    result = redactor.redact_text(text)
    assert "sk-abcdef123456" not in result
    assert "mykey123" not in result


def test_builtin_ignores_short_base64() -> None:
    """Base64 strings shorter than 64 chars should NOT be masked."""
    redactor = BuiltinRedactor()
    short_b64 = "aGVsbG8gd29ybGQ="  # "hello world"
    result = redactor.redact_text(short_b64)
    assert result == short_b64


def test_builtin_preserves_image_field() -> None:
    """'image' field values should NOT be checked by value heuristics."""
    redactor = BuiltinRedactor()
    # SHA256 digest looks like base64 but is not sensitive
    data = {"image": "nginx@sha256:" + "a" * 64}
    result = redactor.redact_object(data)
    assert result["image"] == data["image"]


def test_builtin_preserves_resourceversion() -> None:
    """'resourceVersion' field values should NOT be checked."""
    redactor = BuiltinRedactor()
    data = {"resourceVersion": "123456789012345"}
    result = redactor.redact_object(data)
    assert result["resourceVersion"] == data["resourceVersion"]


# ---------------------------------------------------------------------------
# BuiltinRedactor — K8s-specific tests
# ---------------------------------------------------------------------------


def test_builtin_redacts_env_value() -> None:
    """env entries with {name, value} structure have value masked."""
    redactor = BuiltinRedactor()
    data = {
        "env": [
            {"name": "DB_HOST", "value": "localhost"},
            {"name": "DB_PASSWORD", "value": "secret123"},
        ]
    }
    result = redactor.redact_object(data)
    for entry in result["env"]:
        assert entry["value"] == MASK_TOKEN
        assert "name" in entry


def test_builtin_redacts_command_password_arg() -> None:
    """'--password=xxx' arguments have the value part masked."""
    redactor = BuiltinRedactor()
    data = {
        "command": ["myapp", "--password=hunter2", "--verbose"],
        "args": ["--token=sk-123", "--debug"],
    }
    result = redactor.redact_object(data)
    assert result["command"][0] == "myapp"
    assert result["command"][1] == f"--password={MASK_TOKEN}"
    assert result["command"][2] == "--verbose"
    assert result["args"][0] == f"--token={MASK_TOKEN}"
    assert result["args"][1] == "--debug"


def test_builtin_redacts_non_standard_annotations() -> None:
    """Custom annotation values are masked; standard K8s annotations are kept."""
    redactor = BuiltinRedactor()
    data = {
        "annotations": {
            "app.kubernetes.io/name": "myapp",
            "helm.sh/chart": "myapp-0.1.0",
            "my-company.com/api-key": "secret-value-123",
            "internal/config": "sensitive-data",
        }
    }
    result = redactor.redact_object(data)
    assert result["annotations"]["app.kubernetes.io/name"] == "myapp"
    assert result["annotations"]["helm.sh/chart"] == "myapp-0.1.0"
    assert result["annotations"]["my-company.com/api-key"] == MASK_TOKEN
    assert result["annotations"]["internal/config"] == MASK_TOKEN


# ---------------------------------------------------------------------------
# ChainedMasker tests
# ---------------------------------------------------------------------------


def test_chained_masker_applies_builtin_then_regex() -> None:
    """BuiltinRedactor runs first, then RegexMasker applies on top."""
    builtin = BuiltinRedactor()
    regex = build_masker([r"custom-secret-\d+"])
    chained = ChainedMasker(builtin=builtin, regex=regex)

    data = {
        "db_password": "hunter2",
        "log": "found custom-secret-42 in output",
    }
    result = chained.mask_object(data)
    assert result["db_password"] == MASK_TOKEN
    assert "custom-secret-42" not in result["log"]
    assert MASK_TOKEN in result["log"]


def test_chained_masker_disabled_builtin() -> None:
    """When builtin is disabled, only RegexMasker runs."""
    builtin = BuiltinRedactor(enabled=False)
    regex = build_masker([r"regex-match"])
    chained = ChainedMasker(builtin=builtin, regex=regex)

    data = {"db_password": "hunter2", "log": "regex-match here"}
    result = chained.mask_object(data)
    # builtin disabled → password NOT masked
    assert result["db_password"] == "hunter2"
    # regex still works
    assert result["log"] == f"{MASK_TOKEN} here"


def test_chained_masker_mask_text() -> None:
    """mask_text chains both layers."""
    builtin = BuiltinRedactor()
    regex = build_masker([r"extra-\d+"])
    chained = ChainedMasker(builtin=builtin, regex=regex)

    text = "Bearer sk-abcdef12345678901234 and extra-99"
    result = chained.mask_text(text)
    assert "sk-abcdef" not in result
    assert "extra-99" not in result


# ---------------------------------------------------------------------------
# Hash mode tests
# ---------------------------------------------------------------------------


def test_hash_mode_produces_deterministic_hash() -> None:
    """Same value always produces the same hash."""
    redactor = BuiltinRedactor(hash_mode=True)
    data1 = {"db_password": "hunter2"}
    data2 = {"db_password": "hunter2"}
    r1 = redactor.redact_object(data1)
    r2 = redactor.redact_object(data2)
    assert r1["db_password"] == r2["db_password"]
    assert r1["db_password"].startswith("[HASH:")
    assert r1["db_password"].endswith("]")


def test_hash_mode_different_values_different_hash() -> None:
    """Different values produce different hashes."""
    redactor = BuiltinRedactor(hash_mode=True)
    data = {"password_a": "value1", "password_b": "value2"}
    result = redactor.redact_object(data)
    assert result["password_a"] != result["password_b"]
    assert result["password_a"].startswith("[HASH:")
    assert result["password_b"].startswith("[HASH:")


# ---------------------------------------------------------------------------
# BuiltinRedactor disabled test
# ---------------------------------------------------------------------------


def test_builtin_disabled_passes_through() -> None:
    """When enabled=False, redact_object returns input unchanged."""
    redactor = BuiltinRedactor(enabled=False)
    data = {"db_password": "hunter2", "token": "secret"}
    result = redactor.redact_object(data)
    assert result == data


# ---------------------------------------------------------------------------
# Nested structure test
# ---------------------------------------------------------------------------


def test_builtin_redacts_nested_structures() -> None:
    """Redaction works recursively through nested dicts and lists."""
    redactor = BuiltinRedactor()
    data = {
        "spec": {
            "containers": [
                {
                    "name": "app",
                    "env": [
                        {"name": "API_KEY", "value": "sk-secret"},
                    ],
                    "command": ["app", "--token=abc123"],
                }
            ],
            "metadata": {
                "annotations": {
                    "custom.io/key": "sensitive",
                    "kubernetes.io/name": "safe",
                },
            },
        }
    }
    result = redactor.redact_object(data)
    container = result["spec"]["containers"][0]
    assert container["env"][0]["value"] == MASK_TOKEN
    assert container["command"][1] == f"--token={MASK_TOKEN}"
    annotations = result["spec"]["metadata"]["annotations"]
    assert annotations["custom.io/key"] == MASK_TOKEN
    assert annotations["kubernetes.io/name"] == "safe"
