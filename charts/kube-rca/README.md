# kube-rca

![Version: 1.0.0](https://img.shields.io/badge/Version-1.0.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

Deploy kube-rca backend and frontend

## Quick Start

### Prerequisites

- Kubernetes cluster (v1.32+)
- Helm 3.x
- kubectl configured for your cluster
- An Ingress controller (e.g., [ingress-nginx](https://kubernetes.github.io/ingress-nginx/deploy/)) — required for UI access
- A [Google AI Studio API key](https://aistudio.google.com/apikey) (recommended for Quick Start, covers both analysis and embedding with a single key)

> **Other AI providers:** OpenAI and Anthropic are also supported. See [AI Provider Configuration](#ai-provider-configuration) for details.

### Step 1: Create Values File

Create a `my-values.yaml` with your configuration:

```yaml
# PostgreSQL — chart auto-creates the Secret
postgresql:
  auth:
    existingSecret: ""          # Let bitnami subchart generate the Secret
    password: "change-me"       # Set your DB password

# Backend
backend:
  slack:
    enabled: false              # Disable Slack for Quick Start
  postgresql:
    secret:
      existingSecret: ""        # Auto-resolve from bitnami
  embedding:
    apiKey:
      existingSecret: ""        # Share the agent's AI key Secret

# Agent — set your AI provider and API key
agent:
  aiProvider: "gemini"          # Options: gemini, openai, anthropic
  gemini:
    apiKey: "YOUR_GEMINI_API_KEY"
    secret:
      existingSecret: ""        # Chart creates the Secret from apiKey above

# Frontend
frontend:
  ingress:
    enabled: true
    ingressClassName: "nginx"   # Match your ingress controller (kubectl get ingressclass)
    hosts:
      - "kube-rca.local"       # Or your domain
```

> **Other AI providers:** See [AI Provider Configuration](#ai-provider-configuration) below.
>
> **Production use:** For production, use `existingSecret` with [ExternalSecrets](https://external-secrets.io/) or [Sealed Secrets](https://sealed-secrets.netlify.app/) instead of inline values.

### Step 2: Install

**Option A — From OCI Registry (recommended):**

```bash
helm upgrade --install kube-rca oci://public.ecr.aws/r5b7j2e4/kube-rca-ecr/charts/kube-rca \
  -n kube-rca --create-namespace \
  -f my-values.yaml
```

**Option B — From source (after cloning the repo root):**

```bash
git clone https://github.com/kube-rca/kuberca.git
cd kuberca

helm repo add bitnami https://charts.bitnami.com/bitnami --force-update
helm dependency build charts/kube-rca
helm upgrade --install kube-rca ./charts/kube-rca \
  -n kube-rca --create-namespace \
  -f my-values.yaml
```

### Step 3: Verify

```bash
# Wait for all pods to be ready
kubectl get pods -n kube-rca -w

# Expected (all Running):
#   kube-rca-backend-xxx     1/1  Running
#   kube-rca-agent-xxx       1/1  Running
#   kube-rca-frontend-xxx    1/1  Running
#   kube-rca-postgresql-0    1/1  Running

# Test backend health
kubectl port-forward svc/kube-rca-backend 8080:8080 -n kube-rca
# In another terminal:
curl http://localhost:8080/ping
# Expected: {"message":"pong"}
```

### Step 4: Access the UI

For local testing, add to `/etc/hosts`:

```
# Replace with your Ingress controller's external IP (kubectl get svc -n ingress-nginx)
127.0.0.1 kube-rca.local
```

Open `http://kube-rca.local` in your browser.

> **Default login:** username `kube-rca` / password `kube-rca`
>
> These are set via `backend.auth.admin.username` and `backend.auth.admin.password`. **Change them before production use.**

### Step 5: Connect Alertmanager (Optional)

Add KubeRCA as a webhook receiver in your Alertmanager configuration:

```yaml
# alertmanager.yml
receivers:
  - name: "kube-rca"
    webhook_configs:
      - url: "http://kube-rca-backend.kube-rca.svc.cluster.local:8080/webhook/alertmanager"
        send_resolved: true

route:
  receiver: "kube-rca"
```

If using `kube-prometheus-stack`, add to your Helm values:

```yaml
alertmanager:
  config:
    receivers:
      - name: "kube-rca"
        webhook_configs:
          - url: "http://kube-rca-backend.kube-rca.svc.cluster.local:8080/webhook/alertmanager"
            send_resolved: true
    route:
      receiver: "kube-rca"
```

### Uninstall

```bash
helm uninstall kube-rca -n kube-rca

# PostgreSQL PVCs persist after uninstall. To delete all data:
kubectl delete pvc -n kube-rca -l app.kubernetes.io/name=postgresql
kubectl delete namespace kube-rca
```

### AI Provider Configuration

The default setup uses **Gemini** for both the agent (RCA analysis) and backend (embedding). To use a different provider for the agent:

| Provider | `agent.aiProvider` | Secret key required in `kube-rca-ai` |
|----------|-------------------|--------------------------------------|
| Gemini (default) | `"gemini"` | `ai-studio-api-key` |
| OpenAI | `"openai"` | `openai-api-key` |
| Anthropic | `"anthropic"` | `anthropic-api-key` |

> **Note:** Backend embedding uses Gemini by default (`backend.embedding.provider`). If you choose OpenAI or Anthropic for the agent, you still need a Gemini API key for embedding — or override `backend.embedding` to match your provider.

### Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| Pod `CrashLoopBackOff` (backend) | PostgreSQL Secret missing or wrong key | Verify: `kubectl get secret postgresql -n kube-rca -o yaml` |
| Pod `CrashLoopBackOff` (agent) | AI API key Secret missing | Verify: `kubectl get secret kube-rca-ai -n kube-rca -o yaml` |
| Ingress has no IP/hostname | Ingress controller not installed | Install one: `kubectl get ingressclass` should list your controller |
| UI loads but API calls fail (401/502) | Ingress not routing `/api/*` to backend | Check Ingress rules: `kubectl describe ingress -n kube-rca` |
| `403 Forbidden` on login | Default credentials changed or OIDC misconfigured | Check `backend.auth.admin.*` values or OIDC setup |
| Analysis returns empty results | Agent has no cluster access | Verify agent ServiceAccount RBAC: `kubectl auth can-i list pods --as=system:serviceaccount:kube-rca:kube-rca-agent -n default` |

## Architecture

Frontend ingress automatically routes `/api/*` to the backend service, so a single domain serves both UI and API.

```
your-domain.com/         → frontend (React)
your-domain.com/api/*    → backend  (Go/Gin)
```

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | postgresql | 18.1.13 |

## OIDC Setup Guide (Google)

KubeRCA supports Google OIDC login. Follow these steps to enable it.

### Step 1: Create Google OAuth 2.0 Credentials

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Select or create a project
3. Navigate to **APIs & Services > OAuth consent screen**
   - User Type: **External**
   - App name: `kube-rca` (or any name)
   - Scopes: add `email`, `profile`, `openid`
   - If app status is "Testing", add allowed users under **Test users**
4. Navigate to **APIs & Services > Credentials > + CREATE CREDENTIALS > OAuth client ID**
   - Application type: **Web application**
   - Authorized redirect URIs: add your callback URL
     ```
     https://<YOUR_DOMAIN>/api/v1/auth/oidc/callback
     ```
   - Click **CREATE** and copy the **Client ID** and **Client Secret**

### Step 2: Create a Kubernetes Secret

Store the OAuth credentials in a Kubernetes Secret:

```bash
kubectl create secret generic kube-rca-oidc \
  -n kube-rca \
  --from-literal=oidc-client-id="YOUR_CLIENT_ID" \
  --from-literal=oidc-client-secret="YOUR_CLIENT_SECRET"
```

> **Tip:** You can also add these keys to an existing Secret (e.g., `kube-rca-auth`) instead of creating a new one.

### Step 3: Configure Helm Values

Create or update your override values file:

```yaml
backend:
  auth:
    oidc:
      enabled: true
      redirectUri: "https://<YOUR_DOMAIN>/api/v1/auth/oidc/callback"
      # Allow specific emails (recommended for small teams)
      allowedEmails:
        - user@gmail.com
      # Or allow entire domains
      # allowedDomains:
      #   - your-company.com
      secret:
        existingSecret: kube-rca-oidc  # Secret name from Step 2
```

Then install/upgrade:

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami --force-update
helm dependency build charts/kube-rca
helm upgrade --install kube-rca ./charts/kube-rca \
  -n kube-rca \
  -f my-values.yaml
```

### Step 4: Verify

After deployment, visit `https://<YOUR_DOMAIN>` and click the login button.

### OIDC Configuration Reference

| Value | Description | Default |
|-------|-------------|---------|
| `backend.auth.oidc.enabled` | Enable OIDC authentication | `false` |
| `backend.auth.oidc.issuer` | OIDC issuer URL | `https://accounts.google.com` |
| `backend.auth.oidc.redirectUri` | Callback URL (must match provider console) | `""` |
| `backend.auth.oidc.allowedDomains` | Allowed email domains (e.g., `your-company.com`) | `[]` |
| `backend.auth.oidc.allowedEmails` | Allowed individual emails | `[]` |
| `backend.auth.oidc.secret.existingSecret` | K8s Secret with `oidc-client-id` and `oidc-client-secret` keys | `""` |

> **Note:** If both `allowedDomains` and `allowedEmails` are empty, all authenticated users will be denied.

### Supported OIDC Providers

The login button automatically adapts based on the `issuer` URL. No additional configuration needed.

| Provider | Issuer URL | Button |
|----------|-----------|--------|
| Google | `https://accounts.google.com` | Google로 로그인 |
| Keycloak | `https://keycloak.example.com/realms/my-realm` | Keycloak으로 로그인 |
| Okta | `https://your-org.okta.com` | Okta로 로그인 |
| Azure AD | `https://login.microsoftonline.com/{tenant-id}/v2.0` | Microsoft로 로그인 |
| GitLab | `https://gitlab.com` or self-hosted | GitLab으로 로그인 |
| Other | Any OIDC-compliant issuer | SSO로 로그인 |

#### Example: Keycloak

```yaml
backend:
  auth:
    oidc:
      enabled: true
      issuer: "https://keycloak.example.com/realms/my-realm"
      redirectUri: "https://<YOUR_DOMAIN>/api/v1/auth/oidc/callback"
      allowedDomains:
        - your-company.com
      secret:
        existingSecret: kube-rca-oidc
```

#### Example: Azure AD

```yaml
backend:
  auth:
    oidc:
      enabled: true
      issuer: "https://login.microsoftonline.com/<TENANT_ID>/v2.0"
      redirectUri: "https://<YOUR_DOMAIN>/api/v1/auth/oidc/callback"
      allowedDomains:
        - your-company.com
      secret:
        existingSecret: kube-rca-oidc
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| agent.affinity | object | `{}` | Affinity for agent pods assignment. |
| agent.aiProvider | string | `"gemini"` | AI provider for agent (AI_PROVIDER). Allowed: gemini, openai, anthropic. |
| agent.anthropic.apiKey | string | `""` | Anthropic API key value. When set (and no existingSecret), chart creates the Secret. |
| agent.anthropic.maxTokens | int | `8192` | Anthropic maximum output tokens. |
| agent.anthropic.modelId | string | `"claude-haiku-4-5"` | Anthropic model ID for Strands Agents. |
| agent.anthropic.secret.create | bool | `false` | Create a Secret for the Anthropic API key. |
| agent.anthropic.secret.existingSecret | string | `"kube-rca-ai"` | Existing Secret name for the Anthropic API key. |
| agent.anthropic.secret.key | string | `"anthropic-api-key"` | Secret key name for the Anthropic API key. |
| agent.cache.size | int | `128` | Max number of cached agents (AGENT_CACHE_SIZE). |
| agent.cache.ttlSeconds | int | `0` | Cache TTL in seconds (AGENT_CACHE_TTL_SECONDS, 0 = disable). |
| agent.containerPort | int | `8000` | Agent container port. |
| agent.gemini.apiKey | string | `""` | Gemini API key value. When set (and no existingSecret), chart creates the Secret. |
| agent.gemini.modelId | string | `"gemini-3.1-flash-lite-preview"` | Gemini model ID for Strands Agents. |
| agent.gemini.secret.create | bool | `false` | Create a Secret for the Gemini API key. |
| agent.gemini.secret.existingSecret | string | `"kube-rca-ai"` | Existing Secret name for the Gemini API key. |
| agent.gemini.secret.key | string | `"ai-studio-api-key"` | Secret key name for the Gemini API key. |
| agent.image.pullPolicy | string | `"IfNotPresent"` | Agent image pull policy. |
| agent.image.repository | string | `"public.ecr.aws/r5b7j2e4/kube-rca-ecr/agent"` | Agent image repository. |
| agent.image.tag | string | `"agent-1.2.1"` | Agent image tag. |
| agent.ingress.annotations | object | `{}` | Annotations for agent ingress. |
| agent.ingress.enabled | bool | `false` | Enable agent ingress. |
| agent.ingress.hosts | list | `[]` | Hostnames for agent ingress. |
| agent.ingress.ingressClassName | string | `""` | IngressClass name for agent ingress. |
| agent.ingress.pathType | string | `"Prefix"` | PathType for agent ingress. |
| agent.ingress.paths | list | `["/"]` | Paths for agent ingress. |
| agent.ingress.tls | list | `[]` | TLS configuration for agent ingress. |
| agent.k8s.apiTimeoutSeconds | int | `5` | Kubernetes API timeout in seconds. |
| agent.k8s.eventLimit | int | `25` | Kubernetes event limit. |
| agent.k8s.logTailLines | int | `25` | Kubernetes log tail lines. |
| agent.logLevel | string | `"info"` | Agent log level (LOG_LEVEL). |
| agent.loki.httpTimeoutSeconds | int | `10` | Loki HTTP timeout in seconds. |
| agent.loki.tenantId | string | `""` | Optional Loki tenant header value (LOKI_TENANT_ID / X-Scope-OrgID). |
| agent.loki.url | string | `""` | Loki base URL (LOKI_URL). If empty, Loki log queries are disabled. |
| agent.masking.builtinRedaction | bool | `true` | Enable built-in redaction rules (key denylist, value heuristics, K8s-specific patterns). |
| agent.masking.builtinRedactionHashMode | bool | `false` | Use deterministic hash replacement [HASH:xxx] instead of [MASKED] for correlation. |
| agent.masking.regexList | list | `[]` | Regex list (JSON array) for masking sensitive values before LLM requests and DB persistence. |
| agent.nodeSelector | object | `{}` | Node labels for agent pods assignment. |
| agent.openai.apiKey | string | `""` | OpenAI API key value. When set (and no existingSecret), chart creates the Secret. |
| agent.openai.modelId | string | `"gpt-5-mini"` | OpenAI model ID for Strands Agents. |
| agent.openai.secret.create | bool | `false` | Create a Secret for the OpenAI API key. |
| agent.openai.secret.existingSecret | string | `"kube-rca-ai"` | Existing Secret name for the OpenAI API key. |
| agent.openai.secret.key | string | `"openai-api-key"` | Secret key name for the OpenAI API key. |
| agent.podSecurityContext | object | `{"seccompProfile":{"type":"RuntimeDefault"}}` | Pod-level security context for agent pods. |
| agent.prometheus.httpTimeoutSeconds | int | `5` | Prometheus HTTP timeout in seconds. |
| agent.prometheus.url | string | `""` | Prometheus base URL (PROMETHEUS_URL). If empty, Prometheus queries are disabled. |
| agent.prompt.maxEvents | int | `25` | Max events included in prompt. |
| agent.prompt.maxLogLines | int | `25` | Max log lines included in prompt. |
| agent.prompt.summaryMaxItems | int | `3` | Max session summaries included in prompt. |
| agent.prompt.tokenBudget | int | `32000` | Prompt token budget (approx). |
| agent.replicaCount | int | `2` | Number of agent replicas. Use 2+ for concurrent analysis (Python GIL limits per-process parallelism). |
| agent.resources | object | `{}` | Agent resource requests/limits. |
| agent.retry.maxAttempts | int | `5` | Max retry attempts for transient LLM API errors (5xx, 429). |
| agent.retry.maxWait | float | `60` | Maximum exponential backoff wait time in seconds. |
| agent.retry.minWait | float | `1` | Minimum exponential backoff wait time in seconds. |
| agent.securityContext | object | `{}` | Container-level security context for the agent container. |
| agent.service.port | int | `8000` | Agent service port. |
| agent.service.type | string | `"ClusterIP"` | Agent service type. |
| agent.sessionDB.host | string | `""` | PostgreSQL host for Strands session persistence. |
| agent.sessionDB.name | string | `""` | PostgreSQL database name. |
| agent.sessionDB.port | int | `5432` | PostgreSQL port. |
| agent.sessionDB.secret.existingSecret | string | `"postgresql"` | Existing Secret name for session DB password. |
| agent.sessionDB.secret.key | string | `"password"` | Secret key for session DB password. |
| agent.sessionDB.user | string | `""` | PostgreSQL user. |
| agent.tempo.forwardMinutes | int | `5` | Minutes after alert startsAt for Tempo query window (TEMPO_FORWARD_MINUTES). |
| agent.tempo.httpTimeoutSeconds | int | `10` | Tempo HTTP timeout in seconds. Trace search can be slower than K8s/Prometheus queries under load. |
| agent.tempo.lookbackMinutes | int | `15` | Minutes before alert startsAt for Tempo query window (TEMPO_LOOKBACK_MINUTES). |
| agent.tempo.tenantId | string | `""` | Optional Tempo tenant header value (TEMPO_TENANT_ID / X-Scope-OrgID). |
| agent.tempo.traceLimit | int | `5` | Max traces fetched per alert analysis (TEMPO_TRACE_LIMIT). |
| agent.tempo.url | string | `""` | Tempo base URL (TEMPO_URL). If empty, Tempo trace queries are disabled. |
| agent.tolerations | list | `[]` | Tolerations for agent pods assignment. |
| agent.workers | int | `1` | Uvicorn worker count (WEB_CONCURRENCY). |
| backend.affinity | object | `{}` | Affinity for backend pods assignment. |
| backend.analysis.manualAnalyzeSeverities | string | `""` | Comma-separated severities that require manual analysis (MANUAL_ANALYZE_SEVERITIES). Default (empty): all severities are auto-analyzed. Example: "info" -> info requires manual trigger, warning and critical stay auto. Example: "warning,critical" -> warning and critical require manual trigger, info stays auto. Valid values: info, warning, critical (comma-separated) |
| backend.auth.admin.password | string | `"kube-rca"` | Admin password (default: kube-rca). |
| backend.auth.admin.username | string | `"kube-rca"` | Admin login ID (default: kube-rca). |
| backend.auth.allowSignup | bool | `false` | Allow user signup (ALLOW_SIGNUP). |
| backend.auth.cookie.domain | string | `""` | Cookie domain (AUTH_COOKIE_DOMAIN). |
| backend.auth.cookie.path | string | `"/"` | Cookie path (AUTH_COOKIE_PATH). |
| backend.auth.cookie.sameSite | string | `"Lax"` | Cookie SameSite (AUTH_COOKIE_SAMESITE). |
| backend.auth.cookie.secure | bool | `true` | Cookie secure flag (AUTH_COOKIE_SECURE). |
| backend.auth.cors.allowedOrigins | list | `[]` | Allowed origins (CORS_ALLOWED_ORIGINS), comma-separated when rendered. |
| backend.auth.enabled | bool | `true` | Enable backend auth. |
| backend.auth.jwt.accessTtl | string | `"15m"` | Access token TTL (e.g. 15m). |
| backend.auth.jwt.refreshTtl | string | `"168h"` | Refresh token TTL (e.g. 168h). |
| backend.auth.jwt.secret | string | `""` | JWT secret (auto-generated when empty and no existingSecret). |
| backend.auth.oidc.allowedDomains | list | `[]` | Allowed email domains for OIDC login (OIDC_ALLOWED_DOMAINS). Users with matching email domains are allowed. |
| backend.auth.oidc.allowedEmails | list | `[]` | Allowed individual emails for OIDC login (OIDC_ALLOWED_EMAILS). Specific emails allowed regardless of domain. |
| backend.auth.oidc.enabled | bool | `false` | Enable OIDC authentication (OIDC_ENABLED). |
| backend.auth.oidc.issuer | string | `"https://accounts.google.com"` | OIDC issuer URL (OIDC_ISSUER). |
| backend.auth.oidc.redirectUri | string | `""` | OIDC redirect URI (OIDC_REDIRECT_URI). Must match the callback URL registered with the provider. |
| backend.auth.oidc.secret.existingSecret | string | `""` | Existing Secret name for OIDC credentials. If empty, chart-managed Secret is used. |
| backend.auth.oidc.secret.keys.clientId | string | `"oidc-client-id"` | Secret key for OIDC Client ID. |
| backend.auth.oidc.secret.keys.clientSecret | string | `"oidc-client-secret"` | Secret key for OIDC Client Secret. |
| backend.auth.secret.existingSecret | string | `""` | Existing Secret name for auth credentials (ExternalSecret 연계 시 사용, default keys: admin-username/admin-password/kube-rca-jwt-secret). |
| backend.auth.secret.keys.adminPassword | string | `"admin-password"` | Secret key for admin password. |
| backend.auth.secret.keys.adminUsername | string | `"admin-username"` | Secret key for admin login ID. |
| backend.auth.secret.keys.jwtSecret | string | `"kube-rca-jwt-secret"` | Secret key for JWT secret. |
| backend.auth.secret.name | string | `""` | Custom Secret name for chart-managed auth Secret. |
| backend.containerPort | int | `8080` | Backend container port. |
| backend.embedding.apiKey.existingSecret | string | `"kube-rca-ai"` |  |
| backend.embedding.apiKey.key | string | `"ai-studio-api-key"` |  |
| backend.embedding.apiKey.value | string | `""` | Embedding API key value. When set (and no existingSecret), uses the agent's generated Secret. |
| backend.embedding.model | string | `"gemini-embedding-001"` |  |
| backend.embedding.provider | string | `"gemini"` |  |
| backend.flapping.clearanceWindowMinutes | int | `30` | Time (minutes) after last resolved to clear flapping status (FLAP_CLEARANCE_WINDOW_MINUTES). |
| backend.flapping.cycleThreshold | int | `3` | Number of firing→resolved cycles to trigger flapping detection (FLAP_CYCLE_THRESHOLD). |
| backend.flapping.detectionWindowMinutes | int | `30` | Time window (minutes) for detecting flapping cycles (FLAP_DETECTION_WINDOW_MINUTES). |
| backend.flapping.enabled | bool | `true` | Enable alert flapping detection (FLAP_ENABLED). |
| backend.image.pullPolicy | string | `"IfNotPresent"` | Backend image pull policy. |
| backend.image.repository | string | `"public.ecr.aws/r5b7j2e4/kube-rca-ecr/backend"` | Backend image repository. |
| backend.image.tag | string | `"backend-0.5.1"` | Backend image tag. |
| backend.ingress.annotations | object | `{}` | Annotations for backend ingress. |
| backend.ingress.enabled | bool | `false` | Enable backend ingress. |
| backend.ingress.hosts | list | `[]` | Hostnames for backend ingress. |
| backend.ingress.ingressClassName | string | `""` | IngressClass name for backend ingress. |
| backend.ingress.pathType | string | `"Prefix"` | PathType for backend ingress. |
| backend.ingress.paths | list | `["/"]` | Paths for backend ingress. |
| backend.ingress.tls | list | `[]` | TLS configuration for backend ingress. |
| backend.nodeSelector | object | `{}` | Node labels for backend pods assignment. |
| backend.podSecurityContext | object | `{"seccompProfile":{"type":"RuntimeDefault"}}` | Pod-level security context for backend pods. |
| backend.postgresql.database | string | `"kube-rca"` | PostgreSQL database. |
| backend.postgresql.host | string | `""` | PostgreSQL host. Leave empty to auto-resolve from the embedded postgresql dependency. |
| backend.postgresql.port | int | `5432` | PostgreSQL port. |
| backend.postgresql.retry.initialBackoffSeconds | int | `1` | Initial backoff interval (seconds). |
| backend.postgresql.retry.maxAttempts | int | `10` | Maximum number of DB connection attempts at startup. |
| backend.postgresql.retry.maxBackoffSeconds | int | `30` | Maximum backoff interval (seconds). |
| backend.postgresql.secret.existingSecret | string | `"postgresql"` | Existing Secret name for PostgreSQL password. |
| backend.postgresql.secret.key | string | `"password"` | Secret key for PostgreSQL password. |
| backend.postgresql.user | string | `"kube-rca"` | PostgreSQL user. |
| backend.replicaCount | int | `1` | Number of backend replicas. |
| backend.resources | object | `{}` | Backend resource requests/limits. |
| backend.securityContext | object | `{}` | Container-level security context for the backend container. |
| backend.service.port | int | `8080` | Backend service port. |
| backend.service.type | string | `"ClusterIP"` | Backend service type. |
| backend.slack.channelId | string | `""` | Slack channel ID (used when backend.slack.source=values). |
| backend.slack.enabled | bool | `true` | Enable Slack notifications. |
| backend.slack.secret.channelIdKey | string | `"kube-rca-slack-channel-id"` | Secret key for Slack channel ID. |
| backend.slack.secret.existingSecret | string | `"kube-rca-slack"` | Existing Secret name for Slack credentials. |
| backend.slack.secret.tokenKey | string | `"kube-rca-slack-token"` | Secret key for Slack bot token. |
| backend.slack.token | string | `""` | Slack bot token (used when backend.slack.source=values). |
| backend.tolerations | list | `[]` | Tolerations for backend pods assignment. |
| backend.waitForDb.checkInterval | int | `2` | Interval between checks (seconds). |
| backend.waitForDb.checkTimeout | int | `5` | Timeout for each TCP check (seconds). |
| backend.waitForDb.enabled | bool | `true` | Enable init container to wait for DB before backend starts. |
| frontend.affinity | object | `{}` | Affinity for frontend pods assignment. |
| frontend.containerPort | int | `80` | Frontend container port. |
| frontend.image.pullPolicy | string | `"IfNotPresent"` | Frontend image pull policy. |
| frontend.image.repository | string | `"public.ecr.aws/r5b7j2e4/kube-rca-ecr/frontend"` | Frontend image repository. |
| frontend.image.tag | string | `"frontend-0.4.1"` | Frontend image tag. |
| frontend.ingress.annotations | object | `{}` | Annotations for frontend ingress. |
| frontend.ingress.enabled | bool | `false` | Enable frontend ingress. |
| frontend.ingress.hosts | list | `[]` | Hostnames for frontend ingress. |
| frontend.ingress.ingressClassName | string | `""` | IngressClass name for frontend ingress. |
| frontend.ingress.pathType | string | `"Prefix"` | PathType for frontend ingress. |
| frontend.ingress.paths | list | `["/"]` | Paths for frontend ingress. |
| frontend.ingress.tls | list | `[]` | TLS configuration for frontend ingress. |
| frontend.nodeSelector | object | `{}` | Node labels for frontend pods assignment. |
| frontend.podSecurityContext | object | `{"seccompProfile":{"type":"RuntimeDefault"}}` | Pod-level security context for frontend pods. |
| frontend.replicaCount | int | `1` | Number of frontend replicas. |
| frontend.resources | object | `{}` | Frontend resource requests/limits. |
| frontend.securityContext | object | `{}` | Container-level security context for the frontend container. |
| frontend.service.port | int | `80` | Frontend service port. |
| frontend.service.type | string | `"ClusterIP"` | Frontend service type. |
| frontend.tolerations | list | `[]` | Tolerations for frontend pods assignment. |
| fullnameOverride | string | `""` | Override the full name of the release. |
| hooks.enabled | bool | `true` | Enable Helm hooks for readiness checks. |
| hooks.waitForAgent.enabled | bool | `false` | Enable wait-for-agent post-install hook (optional). |
| hooks.waitForAgent.healthPath | string | `"/healthz"` | Agent health check path. |
| hooks.waitForBackend.enabled | bool | `false` | Enable wait-for-backend post-install hook (optional). |
| hooks.waitForBackend.healthPath | string | `"/ping"` | Backend health check path. |
| hooks.waitForDb.enabled | Deprecated | `false` | Use backend.waitForDb instead. Legacy post-install hook. |
| hooks.waitJob.activeDeadlineSeconds | int | `300` | Maximum time for the job to complete. |
| hooks.waitJob.backoffLimit | int | `6` | Number of retries before giving up. |
| hooks.waitJob.checkInterval | int | `5` | Interval between connection checks (seconds). |
| hooks.waitJob.checkTimeout | int | `10` | Timeout for each check (seconds). |
| hooks.waitJob.image.pullPolicy | string | `"IfNotPresent"` | Wait job image pull policy. |
| hooks.waitJob.image.repository | string | `"busybox"` | Wait job image repository. |
| hooks.waitJob.image.tag | string | `"1.36"` | Wait job image tag. |
| hooks.waitJob.resources.limits.cpu | string | `"50m"` |  |
| hooks.waitJob.resources.limits.memory | string | `"32Mi"` |  |
| hooks.waitJob.resources.requests.cpu | string | `"10m"` |  |
| hooks.waitJob.resources.requests.memory | string | `"16Mi"` |  |
| nameOverride | string | `""` | Override the name of the chart. |
| openapi.affinity | object | `{}` | Affinity for OpenAPI pods assignment. |
| openapi.baseUrl | string | `"/"` | Base URL for Swagger UI. |
| openapi.containerPort | int | `8080` | OpenAPI container port. |
| openapi.enabled | bool | `false` | Enable OpenAPI (Swagger UI) deployment. |
| openapi.image.pullPolicy | string | `"IfNotPresent"` | OpenAPI UI image pull policy. |
| openapi.image.repository | string | `"swaggerapi/swagger-ui"` | OpenAPI UI image repository. |
| openapi.image.tag | string | `"v5.31.0"` | OpenAPI UI image tag. |
| openapi.ingress.annotations | object | `{}` | Annotations for OpenAPI ingress. |
| openapi.ingress.enabled | bool | `false` | Enable OpenAPI ingress. |
| openapi.ingress.hosts | list | `[]` | Hostnames for OpenAPI ingress. |
| openapi.ingress.ingressClassName | string | `""` | IngressClass name for OpenAPI ingress. |
| openapi.ingress.pathType | string | `"Prefix"` | PathType for OpenAPI ingress. |
| openapi.ingress.paths | list | `["/"]` | Paths for OpenAPI ingress. |
| openapi.ingress.tls | list | `[]` | TLS configuration for OpenAPI ingress. |
| openapi.nodeSelector | object | `{}` | Node labels for OpenAPI pods assignment. |
| openapi.podSecurityContext | object | `{"seccompProfile":{"type":"RuntimeDefault"}}` | Pod-level security context for OpenAPI pods. |
| openapi.replicaCount | int | `1` | Number of OpenAPI replicas. |
| openapi.resources | object | `{}` | OpenAPI resource requests/limits. |
| openapi.securityContext | object | `{}` | Container-level security context for the OpenAPI container. |
| openapi.service.port | int | `8080` | OpenAPI service port. |
| openapi.service.type | string | `"ClusterIP"` | OpenAPI service type. |
| openapi.specs.agent.name | string | `"agent"` | Display name for agent spec. |
| openapi.specs.agent.path | string | `"/openapi.json"` | Agent OpenAPI path. |
| openapi.specs.agent.service.name | string | `""` | Agent service name override (empty = chart default). |
| openapi.specs.agent.service.port | int | `8000` | Agent service port. |
| openapi.specs.backend.name | string | `"backend"` | Display name for backend spec. |
| openapi.specs.backend.path | string | `"/openapi.json"` | Backend OpenAPI path. |
| openapi.specs.backend.service.name | string | `""` | Backend service name override (empty = chart default). |
| openapi.specs.backend.service.port | int | `8080` | Backend service port. |
| openapi.tolerations | list | `[]` | Tolerations for OpenAPI pods assignment. |
| postgresql.auth.database | string | `"kube-rca"` |  |
| postgresql.auth.existingSecret | string | `"postgresql"` |  |
| postgresql.auth.secretKeys.adminPasswordKey | string | `"postgres-password"` |  |
| postgresql.auth.secretKeys.userPasswordKey | string | `"password"` |  |
| postgresql.auth.username | string | `"kube-rca"` |  |
| postgresql.primary.initdb.scripts."enable-pgvector.sh" | string | `"#!/bin/sh\nset -e\nDB_USER=\"postgres\"\nif [ -n \"${POSTGRES_POSTGRES_PASSWORD_FILE:-}\" ] && [ -f \"$POSTGRES_POSTGRES_PASSWORD_FILE\" ]; then\n  PGPASSWORD=\"$(cat \"$POSTGRES_POSTGRES_PASSWORD_FILE\")\"\nelif [ -n \"${POSTGRES_POSTGRES_PASSWORD:-}\" ]; then\n  PGPASSWORD=\"$POSTGRES_POSTGRES_PASSWORD\"\nelse\n  DB_USER=\"${POSTGRES_USER:-postgres}\"\n  if [ -n \"${POSTGRES_PASSWORD_FILE:-}\" ] && [ -f \"$POSTGRES_PASSWORD_FILE\" ]; then\n    PGPASSWORD=\"$(cat \"$POSTGRES_PASSWORD_FILE\")\"\n  elif [ -n \"${POSTGRES_PASSWORD:-}\" ]; then\n    PGPASSWORD=\"$POSTGRES_PASSWORD\"\n  fi\nfi\nexport PGPASSWORD\nDB_NAME=\"${POSTGRES_DATABASE:-kube-rca}\"\npsql -U \"$DB_USER\" -d \"$DB_NAME\" -c \"CREATE EXTENSION IF NOT EXISTS vector;\"\n"` |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs](https://github.com/norwoodj/helm-docs)
