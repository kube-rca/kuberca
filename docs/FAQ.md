# KubeRCA FAQ

Frequently asked questions about deploying and operating KubeRCA.

---

## Which AI provider should I choose?

KubeRCA supports three providers via `agent.aiProvider`:

| Provider   | `agent.aiProvider` | Default model              | Cost (rough) | Latency  | Quality | Prompt budget |
|------------|--------------------|----------------------------|--------------|----------|---------|----------------|
| Gemini     | `gemini`           | `gemini-3.1-flash-lite-preview` | Lowest       | Fastest  | Good     | Large (~1M)    |
| OpenAI     | `openai`           | `gpt-5-mini`               | Mid          | Mid      | High     | Mid (~128k+)   |
| Anthropic  | `anthropic`        | `claude-haiku-4-5`         | Mid          | Mid-low  | Highest tool reasoning | Large (~200k)  |

Rules of thumb:

- **Default / cost-sensitive** -> Gemini (`gemini-3.1-flash-lite-preview` is the recommended runtime baseline; `flash-lite` alone has been observed underperforming on KubeRCA prompts).
- **Highest quality on long, multi-step traces** -> Anthropic. Bump `agent.anthropic.maxTokens` if you see truncated responses.
- **Balanced default if your org already standardized on OpenAI** -> OpenAI.

> Note: backend embeddings (similarity search for past incidents) only use Gemini. Even if you switch the agent to OpenAI / Anthropic, you still need a Gemini API key for `backend.embedding.apiKey`.

The provider name must be exactly one of `gemini`, `openai`, `anthropic`. The chart `values.schema.json` rejects other strings, including the common typo `antrophic`.

---

## Is Slack integration required?

**No.** KubeRCA works end-to-end without Slack:

- Alertmanager -> backend webhook still fires
- Agent still runs analysis
- Frontend still shows incidents and analyses

Slack is purely an outbound notification + thread-context channel. Disable it with:

```yaml
backend:
  slack:
    enabled: false
```

If you later want it, set `backend.slack.enabled=true`, fill in `backend.slack.token` and `backend.slack.channelId` (or point to an existing Secret), and `helm upgrade`.

---

## Can I use an external PostgreSQL?

**Yes.** Disable the bundled subchart and point the backend at your own database:

```yaml
postgresql:
  enabled: false

externalDatabase:
  host: "my-rds.cluster-xxxx.us-east-1.rds.amazonaws.com"
  port: 5432
  user: "kube-rca"
  database: "kube-rca"
  existingSecret: "kube-rca-external-db"   # must contain key 'password'

backend:
  postgresql:
    host: "my-rds.cluster-xxxx.us-east-1.rds.amazonaws.com"
    port: 5432
    user: "kube-rca"
    database: "kube-rca"
    secret:
      existingSecret: "kube-rca-external-db"
      key: "password"

agent:
  sessionDB:
    host: "my-rds.cluster-xxxx.us-east-1.rds.amazonaws.com"
    port: 5432
    name: "kube-rca"
    user: "kube-rca"
    secret:
      existingSecret: "kube-rca-external-db"
      key: "password"
```

Make sure the `vector` extension is available on your PostgreSQL — see `docs/TROUBLESHOOTING.md` ("pgvector extension not found").

---

## Does it support multiple clusters?

**Not yet.** Multi-cluster fan-in is on the roadmap (target Q3 2026). The current architecture is single-cluster:

- One Alertmanager -> one KubeRCA backend
- The backend collects context from the cluster it runs in (via the in-cluster Kubernetes API)

For now, the recommended pattern is one KubeRCA install per cluster, with each install posting to its own Slack channel / OIDC tenant. We track this work under <https://github.com/kube-rca/kuberca/issues> with the `multi-cluster` label.

---

## What happens if Loki / Tempo / Istio are not installed?

KubeRCA treats Loki, Tempo, and Istio as **optional context enrichers**. If their endpoints are not configured, the agent skips them silently and still produces an RCA from Kubernetes events + pod logs + Prometheus metrics:

```yaml
agent:
  prometheus:
    url: ""    # disabled
  loki:
    url: ""    # disabled
  tempo:
    url: ""    # disabled
```

When you install one of them later, just set the URL in your values file and `helm upgrade`. No code change is needed in the backend or agent.

Tempo URLs must use the `service.namespace` form (e.g. `tempo.monitoring`) — Tempo does not emit a `k8s.namespace.name` tag, so the agent uses the FQDN to resolve traces.

---

## How are LLM responses masked?

KubeRCA runs a **two-stage masking pipeline** on every prompt and response:

1. **Built-in redaction** (`agent.masking.builtinRedaction: true`, default) — applies a curated denylist of secret-shaped patterns (JWTs, AWS access keys, Bearer tokens, K8s `data.*` blobs, ConfigMap values that look base64-encoded, etc.).
2. **User regex list** (`agent.masking.regexList`) — your own JSON array of regex patterns to redact.

Optionally, set `agent.masking.builtinRedactionHashMode: true` to replace matches with a deterministic `[HASH:xxx]` token so the same secret produces the same hash across the conversation (useful for correlation without leaking the value).

For the source of truth, see `agent/tests/test_masking.py` in the repo — it doubles as a regression suite for the masking rules.

---

## How do I upgrade?

`helm upgrade` is generally safe between minor versions:

```bash
helm upgrade kube-rca oci://public.ecr.aws/r5b7j2e4/kube-rca-ecr/charts/kube-rca \
  -n kube-rca \
  -f my-values.yaml
```

Before upgrading, check `CHANGELOG.md` (auto-generated by release-please) for any entries marked **BREAKING CHANGE**. Examples of changes that need extra care:

- Bitnami `postgresql` major version bump (review their migration notes — usually a PVC deletion + restore is *not* required, but read the chart notes)
- New required values key (the chart's `values.schema.json` will fail-fast and tell you which key is missing)
- Backend or agent image tag with a database migration — the chart's `wait-for-db` init container handles ordering; you should not need manual SQL

If the upgrade fails partway, `helm rollback kube-rca <previous-revision> -n kube-rca` returns the cluster to the prior state. Backend / agent are stateless except for PostgreSQL; data is preserved across rollbacks.
