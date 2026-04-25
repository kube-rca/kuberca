# ADR-0001: Strands Agents for LLM Orchestration

- **Status**: Accepted
- **Date**: 2026-04-25
- **Deciders**: Project Maintainers

## Context

KubeRCA's Python `agent` service must call out to a Large Language Model to
produce root-cause analyses from collected Kubernetes / Prometheus / Tempo
context. Operators run KubeRCA on many different cloud accounts, and "the
right LLM" varies — some teams have an existing OpenAI contract, some run on
GCP and prefer Gemini, some are Anthropic Claude shops.

We needed an orchestration layer that:

1. Abstracts the provider behind one interface (Gemini / OpenAI / Anthropic).
2. Supports tool/function calling so the agent can fetch additional context on
   demand (e.g., "give me logs from the past 5 minutes").
3. Stays lightweight — KubeRCA is meant to be self-hosted; we cannot impose
   heavy frameworks on a small operations team.
4. Supports server-side use without bringing along a CLI-style callback handler
   that prints to stdout (which we discovered breaks structured logs in pods).

## Decision

We adopt **Strands Agents** as the orchestration layer for the `agent` service.

Strands provides:

- A single `Agent` abstraction that swaps providers via configuration.
- Native tool-use semantics with a clean Python API.
- A `null_callback_handler` we can plug in for headless server runs.
- An ecosystem aligned with Anthropic's recommended Claude usage patterns,
  with adapters for Gemini and OpenAI.

Provider selection is driven by the Helm value `agent.aiProvider` (one of
`gemini | openai | anthropic` — strict enum) and per-provider `modelId`.

## Alternatives Considered

- **LangChain** — mature but heavy. Frequent breaking changes and a
  large dependency footprint we did not want to push onto operators.
- **LlamaIndex** — strong on retrieval pipelines, weaker on tool-use
  ergonomics for our agent style.
- **Hand-rolled provider clients** — minimal dependencies, but we would have
  reimplemented retries, tool-calling, and prompt budget tracking. Not worth
  the maintenance burden.

## Consequences

### Positive

- Operators can switch LLM providers via Helm values without touching code.
- Tool-calling support keeps the agent focused on small, composable context
  collectors.
- Lightweight enough to fit comfortably in a single-pod deployment.

### Negative

- Dependency on a relatively young framework. We mitigate by pinning the
  `strands-agents` version and tracking upstream releases through Dependabot.
- Some advanced features (e.g., distributed sessions) are still upstream-side
  TODOs; we use simple session IDs + a `pgvector`-backed session repository
  instead.

### Operational Notes

- `null_callback_handler` is **mandatory** in the server context — the default
  CLI handler writes directly to stdout and corrupts structured log output.
- Prompt budget enforcement is implemented in our service layer rather than
  relying on Strands' built-in budget; this gives us tighter control over which
  context sections to drop first when nearing the limit.
