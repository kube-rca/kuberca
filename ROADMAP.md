# KubeRCA Roadmap

This roadmap captures the direction of the project at a high level. It is **not
a binding commitment** — priorities shift as the community grows, alerts surface
real-world needs, and contributors invest their time. Items here will move
between quarters; the roadmap is updated by Maintainers in coordination with
the community.

If you want to influence the roadmap, the best paths are:

- Open a thread in [GitHub Discussions](https://github.com/kube-rca/kuberca/discussions).
- Comment on (or open) an issue tagged with the relevant area label.
- Contribute a proof-of-concept PR — code generally accelerates a discussion.

## Q2 2026 — Foundations & Discoverability

- **Tempo enricher (GA)** — promote the Tempo trace enricher from optional
  beta to a documented, supported context source.
- **English documentation parity** — full English translation for installation
  and operations guides; bilingual default landing page.
- **OpenSSF Scorecard ≥ 7** — publish supply-chain controls (cosign signing,
  SBOM via syft, SLSA build provenance, SHA-pinned GitHub Actions).
- **Webhook hardening** — HMAC signature validation and rate limiting on
  `/webhook/alertmanager` (W2 hardening workstream).
- **Test coverage floor** — 40% Go coverage in `backend/internal/handler`,
  bootstrap Vitest in frontend, weekly Chaos Mesh CI run (W3 workstream).

## Q3 2026 — Scale & Provider Breadth

- **Multi-cluster federation** — single KubeRCA instance correlating
  alerts/incidents across multiple Kubernetes clusters.
- **Additional LLM providers** — AWS Bedrock and Azure OpenAI as first-class
  providers alongside Gemini / OpenAI / Anthropic, with a provider-agnostic
  cost & latency dashboard.
- **Pluggable enrichers** — well-defined extension point so the community can
  add custom context collectors (e.g., AWS CloudWatch, Datadog, Splunk).
- **Helm chart on Artifact Hub** — listed, signed, with `values.schema.json`.

## Q4 2026 — Community & Long-term Stewardship

- **CNCF Sandbox application** — apply once governance, ADRs, and security
  documentation are in steady state.
- **Stable plugin / agent SDK** — `v1` interface for community-built enrichers
  and tools.
- **Meaningful similar-incident retrieval** — improve the pgvector + embedding
  pipeline with quality benchmarks and tunable thresholds.
- **Public incident postmortem template** — reusable templates for shipping
  KubeRCA-driven postmortems.

## Always-On Backlog

These are continuously-tracked threads that may surface in any quarter:

- Reduce LLM token usage and latency (prompt budget, caching).
- Improve masking pipeline quality (PII / secret coverage).
- Expand chaos scenario library, especially Istio / network failure modes.
- Lower the operator burden — better defaults, clearer error messages,
  one-command installs.

## How Roadmap Changes Land

Updates to this document follow the
[GOVERNANCE.md](GOVERNANCE.md) rules for governance changes (2/3 vote of
active Maintainers, 7-day review window).
