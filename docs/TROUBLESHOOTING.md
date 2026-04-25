# KubeRCA Troubleshooting

Common installation and runtime issues, ordered roughly from "first install" to "deeper integration."

> Need a fresh install walkthrough first? See `docs/installation-guide-en.md`.

---

## 1. AI provider key not configured

**Symptom**

The agent pod is `Running`, but `/analyze` returns errors like `AI_PROVIDER api key missing`, or analyses never complete and you see `401`/`403` from the LLM provider in agent logs.

**Cause**

`agent.aiProvider` is set, but the matching `agent.<provider>.apiKey` (or `agent.<provider>.secret.existingSecret`) is empty.

**Fix**

Pick one of:

1. **Inline value** (Quick Start only — not recommended for production):

   ```yaml
   agent:
     aiProvider: "gemini"   # or "openai" / "anthropic"
     gemini:
       apiKey: "AIza..."
   ```

2. **Existing Secret** (recommended):

   ```yaml
   agent:
     aiProvider: "openai"
     openai:
       secret:
         create: false
         existingSecret: "kube-rca-ai"
         key: "openai-api-key"
   ```

   Then create the Secret out-of-band:

   ```bash
   kubectl -n kube-rca create secret generic kube-rca-ai \
     --from-literal=openai-api-key="sk-..."
   ```

The allowed `aiProvider` values are exactly `gemini`, `openai`, `anthropic`. Note the spelling — `anthropic`, not `antrophic`.

---

## 2. `pgvector` extension not found

**Symptom**

Backend pod CrashLoops with logs like `ERROR: type "vector" does not exist` or `extension "vector" is not available`.

**Cause**

The PostgreSQL instance is missing the `vector` extension. By default the bundled Bitnami `postgresql` sub-chart runs an `initdb` script (`enable-pgvector.sh`) that runs `CREATE EXTENSION IF NOT EXISTS vector;` on first boot. If you bring your own database, this never runs.

**Fix**

- **Bundled PostgreSQL**: confirm the chart was installed with default values; you should see `kube-rca-postgresql-0` running. If the volume already existed from a previous install without `pgvector`, delete the PVC and reinstall:

  ```bash
  helm uninstall kube-rca -n kube-rca
  kubectl delete pvc -n kube-rca -l app.kubernetes.io/name=postgresql
  helm upgrade --install kube-rca ... # reinstall
  ```

- **External PostgreSQL** (`postgresql.enabled=false`): connect as a superuser and run:

  ```sql
  CREATE EXTENSION IF NOT EXISTS vector;
  ```

  Make sure the PostgreSQL build has the `pgvector` extension installed at the OS level (some managed databases need to enable it via the cloud console first — for example, RDS requires `rds.allowed_extensions` to include `vector`).

---

## 3. Slack messages not posting

**Symptom**

Alerts arrive at the backend (`/webhook/alertmanager` returns 200) but no message appears in Slack. Backend logs show `slack: not_authed`, `invalid_auth`, or `channel_not_found`.

**Cause**

One of:

- `backend.slack.enabled` is `false`
- `backend.slack.token` (or the Secret it references) is empty / wrong
- Bot lacks `chat:write` and/or `channels:read` scopes
- Channel ID is invalid or the bot has not been invited to the channel

**Fix**

1. Confirm the Slack token is set:

   ```bash
   kubectl -n kube-rca get secret kube-rca-slack -o jsonpath='{.data.kube-rca-slack-token}' | base64 -d
   ```

2. In Slack, open your app config and verify the OAuth scopes:
   - `chat:write` (post messages)
   - `channels:read` (look up the channel)

3. Invite the bot to the target channel:

   ```
   /invite @your-kube-rca-bot
   ```

4. Confirm `channelId` is the **channel ID** (e.g. `C12345ABCDE`) — not the human-readable name.

---

## 4. OIDC `redirect_uri` mismatch

**Symptom**

Google/Okta login fails with `Error 400: redirect_uri_mismatch`, or the user lands back at the login page after authenticating.

**Cause**

`backend.auth.oidc.redirectUri` does not match any redirect URI registered in your OIDC provider console.

**Fix**

The backend always uses `<frontend-host>/api/v1/auth/oidc/callback` as the callback path. Both sides must agree:

- **Helm values**:

  ```yaml
  backend:
    auth:
      oidc:
        enabled: true
        issuer: "https://accounts.google.com"
        redirectUri: "https://kube-rca.example.com/api/v1/auth/oidc/callback"
  ```

- **Google Cloud Console** -> APIs & Services -> Credentials -> Authorized redirect URIs:
  - `https://kube-rca.example.com/api/v1/auth/oidc/callback`

The host, scheme, and path must match exactly. After editing, `helm upgrade` and re-trigger login.

---

## 5. `ImagePullBackOff` on KubeRCA images

**Symptom**

Pods stuck in `ImagePullBackOff`. `kubectl describe pod` shows errors pulling from `public.ecr.aws/r5b7j2e4/kube-rca-ecr/...`.

**Cause**

ECR Public *does* allow anonymous pulls — no credentials needed. So this is almost always a node-side networking or DNS issue:

- The cluster cannot reach `public.ecr.aws` (egress firewall, NAT gateway, no IPv4 route)
- Region-restricted egress policy blocks `*.amazonaws.com`
- Stale image cache after an image tag was force-pushed (rare)

**Fix**

1. From a debug pod inside the cluster:

   ```bash
   kubectl run --rm -it --image=alpine:3 net-debug -- sh
   apk add --no-cache curl
   curl -sI https://public.ecr.aws/v2/
   ```

   You should get a 401 (challenge) — that proves DNS + TLS work.

2. If your nodes are in a private subnet, confirm the NAT gateway / VPC endpoint route is healthy.

3. Verify the tag actually exists:

   ```bash
   docker pull public.ecr.aws/r5b7j2e4/kube-rca-ecr/backend:backend-0.5.1
   ```

---

## 6. `CrashLoopBackOff` on backend startup

**Symptom**

`kube-rca-backend` keeps restarting. Logs show `dial tcp: connection refused` or `pq: SSL is not enabled on the server` immediately after pod start.

**Cause**

Backend is starting before PostgreSQL is reachable, or the password Secret is wrong.

**Fix**

1. Confirm the `wait-for-db` init container is enabled (default is `true`):

   ```yaml
   backend:
     waitForDb:
       enabled: true
   ```

2. Confirm the password key matches:

   ```yaml
   backend:
     postgresql:
       secret:
         existingSecret: "postgresql"
         key: "password"
   ```

3. Inspect the postgresql Secret:

   ```bash
   kubectl -n kube-rca get secret postgresql -o jsonpath='{.data}' | jq 'keys'
   ```

4. Tune retries if your DB is slow to become ready (e.g. cold-started RDS Aurora):

   ```yaml
   backend:
     postgresql:
       retry:
         maxAttempts: 30
         maxBackoffSeconds: 60
   ```

---

## 7. Agent `/analyze` times out

**Symptom**

The backend marks the alert as analyzed late (or as failed). Agent logs show requests still running past 120 seconds.

**Cause**

The default backend->agent timeout is 120 seconds. Long LLM calls (especially `claude-haiku-4-5` on a long context, or a slow Tempo trace search) can exceed it.

**Fix**

- **Backend side** — bump the agent timeout via env on the backend Deployment (set the `AGENT_TIMEOUT` env in your values override or via a custom patch).
- **Agent side** — give the agent more headroom:

  ```yaml
  agent:
    resources:
      requests:
        cpu: 500m
        memory: 512Mi
      limits:
        cpu: 2
        memory: 1Gi
    prompt:
      tokenBudget: 16000     # cut prompt size
      maxLogLines: 15
      maxEvents: 15
  ```

- **Provider side** — lower `agent.anthropic.maxTokens`, or switch to a faster model (e.g. `gemini-3.1-flash-lite-preview`).

If you have **Loki** or **Tempo** integrated, also confirm those endpoints are reachable; the agent waits on each enricher up to its `httpTimeoutSeconds` setting. Optional enrichers degrade gracefully when their `url` is empty.

---

## 8. Webhook returns `401 Unauthorized`

**Symptom**

After upgrading to a future release with HMAC webhook validation enabled, Alertmanager logs show `401 Unauthorized` from `POST /webhook/alertmanager`.

**Cause**

Inbound HMAC validation is gated on a future release (W2). When that lands, the backend will reject webhook calls whose `X-KubeRCA-Signature` header is missing or wrong.

**Fix**

When the feature ships:

1. Set a shared secret on the backend:

   ```yaml
   webhookSecret: "$(openssl rand -hex 32)"
   ```

2. Configure Alertmanager to sign webhook payloads with the same secret (Alertmanager doesn't do HMAC natively today; pair with a small relay such as the planned `kube-rca-webhook-proxy`, or temporarily disable validation by leaving `webhookSecret` unset).

Until that release, the webhook accepts any request that matches the Alertmanager payload schema. Restrict access at the network layer (e.g. NetworkPolicy or service mesh policy) in the meantime.

---

## Still stuck?

- Pull pod logs: `kubectl -n kube-rca logs deploy/kube-rca-backend --tail=200`
- Tail agent logs while triggering an alert: `kubectl -n kube-rca logs -f deploy/kube-rca-agent`
- Open an issue with logs and your `helm get values kube-rca -n kube-rca`: <https://github.com/kube-rca/kuberca/issues>
