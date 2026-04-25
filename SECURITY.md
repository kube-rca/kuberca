# Security Policy

KubeRCA is an open-source project that ingests Kubernetes alerts, calls third-party LLM
providers, and stores incident data in PostgreSQL with pgvector. Because it runs in
production observability paths, we treat security reports as a first-class concern.

## Supported Versions

We provide security fixes for the latest minor release of each component. Older minor
versions are supported on a best-effort basis only.

| Component  | Tag format         | Supported versions |
| ---------- | ------------------ | ------------------ |
| Backend    | `backend-X.Y.Z`    | `>= 1.0.0`         |
| Agent      | `agent-X.Y.Z`      | `>= 1.0.0`         |
| Frontend   | `frontend-X.Y.Z`   | `>= 1.0.0`         |
| Helm chart | `helm-chart-X.Y.Z` | `>= 1.0.0`         |

Pre-1.0 releases (`< 1.0.0`) are still considered experimental. We will accept reports
against them but cannot guarantee a backport to older lines.

## Reporting a Vulnerability

**Please do not open a public GitHub Issue for security reports.**

Use one of the following private channels:

1. **Preferred — GitHub Private Vulnerability Reporting**
   Go to <https://github.com/kube-rca/kuberca/security/advisories/new> and submit a
   private advisory. This creates a confidential thread visible only to maintainers.
2. **Secondary — Email**
   Send a report to `security@<placeholder-domain>` (the address will be published once
   the project mailbox is provisioned). Until then, please prefer the GitHub channel
   above.

Include as much of the following as possible:

- Affected component (`backend`, `agent`, `frontend`, `helm-chart`) and version/tag.
- Reproduction steps or a minimal proof of concept.
- Impact assessment (information disclosure, RCE, privilege escalation, etc.).
- Any suggested mitigation or patch.

## Triage SLA

We aim to follow this timeline for valid reports:

| Stage          | Target                                                       |
| -------------- | ------------------------------------------------------------ |
| Acknowledgment | Within **7 calendar days** of receipt.                       |
| Initial triage | Severity assessment and ownership assigned within **14 days**. |
| Fix or mitigation | Patch, workaround, or detailed remediation plan within **30 days** for High/Critical issues. Lower-severity issues are scheduled into the regular release train. |
| Public disclosure | Coordinated with the reporter once a fix is shipped, typically via a GitHub Security Advisory and CHANGELOG entry. |

If we cannot meet a deadline, we will communicate the reason and a revised plan to the
reporter through the same channel.

## Coordinated Disclosure

We follow standard coordinated disclosure:

- The reporter and maintainers agree on a disclosure window before any public mention.
- Credit is given to the reporter in the advisory and release notes unless they prefer
  to remain anonymous.
- CVEs may be requested through GitHub for High/Critical findings.

## PGP / Signed Communication

We do not yet publish a PGP key. When the project security mailbox is provisioned, we
will publish a key at `https://github.com/kube-rca/kuberca/security` and update this
section. Until then, please rely on GitHub's encrypted advisory channel.

## Hardening Roadmap

A few security-relevant items are tracked on the public roadmap and are intentionally
called out here so reporters know they are already in scope:

- **HMAC-signed Alertmanager webhooks** — work tracked under the W2 hardening
  workstream. The webhook endpoint will move from "trust the network path" to a
  signed, replay-protected scheme.
- **Dependency scanning + OpenSSF Scorecard** — tracked toward Q2 2026.
- **Supply-chain attestations (SLSA / cosign)** for the published container images
  and Helm OCI artifacts.

If you find a security issue in any of those areas — even before they GA — we still
want to hear about it.

## Out of Scope

- Findings that require physical access to a maintainer's machine.
- Issues only reproducible against unsupported versions (see table above) where the
  fix is already present in a supported version.
- Vendor-side LLM provider (Gemini/OpenAI/Anthropic) issues — please report those to
  the upstream vendor. We will, however, address misuse, leakage, or mishandling of
  responses inside KubeRCA itself.

Thank you for helping keep KubeRCA and its operators safe.
