# KubeRCA Installation Guide

A guide to installing KubeRCA, an alert-driven Root Cause Analysis (RCA) tool for Kubernetes, and running your first demo.

---

## Prerequisites

| Item | Minimum requirement |
|------|---------------------|
| Kubernetes | v1.32 or later |
| Helm | v3.x |
| kubectl | Connected to the target cluster |
| AI API Key | [Google AI Studio API Key](https://aistudio.google.com/apikey) (recommended; usable for both analysis and embeddings) |

### Cluster resources

| Component | CPU (request) | Memory (request) | Notes |
|-----------|--------------|-------------------|-------|
| Backend | 100m | 128Mi | Go server |
| Agent | 200m | 256Mi | Python FastAPI + LLM calls |
| Frontend | 50m | 64Mi | Nginx static serving |
| PostgreSQL | 250m | 256Mi | Includes pgvector |
| **Total** | **~600m** | **~704Mi** | Minimum baseline; tune for production |

> **Note:** The values above are the minimum to make the chart run. Tune `resources.requests` / `resources.limits` in your values file for production workloads.

### Pre-flight checks

```bash
# Check Kubernetes version
kubectl version --short

# Check Helm version
helm version --short
```

---

## Install (under 5 minutes)

### Step 1: Create a values file

Create a `my-values.yaml` file:

```yaml
# PostgreSQL — the chart auto-creates the Secret
postgresql:
  auth:
    existingSecret: ""          # let the bitnami subchart manage the Secret
    password: "change-me"       # set the DB password

# Backend
backend:
  slack:
    enabled: false              # disable Slack for the Quick Start
  postgresql:
    secret:
      existingSecret: ""        # bitnami auto-reference
  embedding:
    apiKey:
      existingSecret: ""        # share the agent's AI key Secret

# Agent — AI provider and API key
agent:
  aiProvider: "gemini"          # gemini | openai | anthropic
  gemini:
    apiKey: "PASTE_YOUR_API_KEY_HERE"
    secret:
      existingSecret: ""        # apiKey value triggers Secret creation

# Frontend — accessed via port-forward (configure Ingress separately for production)
```

> **Using a different AI provider:**
>
> | Provider | `agent.aiProvider` | apiKey field |
> |----------|--------------------|--------------|
> | Gemini (default) | `"gemini"` | `agent.gemini.apiKey` |
> | OpenAI | `"openai"` | `agent.openai.apiKey` |
> | Anthropic | `"anthropic"` | `agent.anthropic.apiKey` |
>
> **Heads up:** The backend embedding feature currently **only supports Gemini**. Even if you run the agent on OpenAI or Anthropic, you still need a separate Gemini API key for embeddings.

### Step 2: Install with Helm

```bash
helm upgrade --install kube-rca oci://public.ecr.aws/r5b7j2e4/kube-rca-ecr/charts/kube-rca \
  -n kube-rca --create-namespace \
  -f my-values.yaml
```

### Step 3: Verify the install

```bash
kubectl get pods -n kube-rca -w
```

**The install is successful when every pod is `Running`:**

```
NAME                                READY   STATUS    RESTARTS   AGE
kube-rca-backend-xxxxxxxxxx-xxxxx   1/1     Running   0          2m
kube-rca-agent-xxxxxxxxxx-xxxxx     1/1     Running   0          2m
kube-rca-frontend-xxxxxxxxxx-xxxxx  1/1     Running   0          2m
kube-rca-postgresql-0               1/1     Running   0          2m
```

Backend health check:

```bash
kubectl port-forward svc/kube-rca-backend 8080:8080 -n kube-rca &
curl http://localhost:8080/ping
# Response: {"message":"pong"}
```

### Step 4: Access the UI

```bash
# Terminal 1 — Frontend
kubectl port-forward svc/kube-rca-frontend 3000:80 -n kube-rca

# Terminal 2 — Backend (needed for the frontend's API calls)
kubectl port-forward svc/kube-rca-backend 8080:8080 -n kube-rca
```

Open `http://localhost:3000` in your browser to see the login screen.

| Field | Default |
|-------|---------|
| ID | `kube-rca` |
| Password | `kube-rca` |

> Always change `backend.auth.admin.username` and `backend.auth.admin.password` for production.

---

## Uninstall

```bash
# Remove KubeRCA
helm uninstall kube-rca -n kube-rca

# Remove the PostgreSQL PVC and namespace (deletes all data)
kubectl delete pvc -n kube-rca -l app.kubernetes.io/name=postgresql
kubectl delete namespace kube-rca
```

---

## First scenario: OOMKilled root-cause analysis (under 10 minutes)

A demo you can run right after install. It triggers an OOMKilled pod and watches KubeRCA perform Root Cause Analysis automatically.

### Preconditions

- KubeRCA install complete (steps above)
- Alertmanager wired to the KubeRCA webhook

### Step 1: Wire up Alertmanager

KubeRCA only receives alerts when Alertmanager has a webhook receiver configured.

**With kube-prometheus-stack** (add to your Helm values):

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

**With a standalone Alertmanager** (add to `alertmanager.yml`):

```yaml
receivers:
  - name: "kube-rca"
    webhook_configs:
      - url: "http://kube-rca-backend.kube-rca.svc.cluster.local:8080/webhook/alertmanager"
        send_resolved: true

route:
  receiver: "kube-rca"
```

### Step 2: Inject an OOMKilled failure

Deploy a pod with an aggressively low memory limit:

```bash
kubectl create namespace demo-oom 2>/dev/null || true

kubectl apply -n demo-oom -f - <<'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: memory-hog
  labels:
    app: memory-hog
spec:
  replicas: 1
  selector:
    matchLabels:
      app: memory-hog
  template:
    metadata:
      labels:
        app: memory-hog
    spec:
      containers:
        - name: stress
          image: polinux/stress:1.0.4
          command: ["stress"]
          args: ["--vm", "1", "--vm-bytes", "128M", "--vm-hang", "1"]
          resources:
            requests:
              memory: "32Mi"
            limits:
              memory: "64Mi"    # allocates 128M -> OOMKilled
EOF
```

### Step 3: Observe the failure

```bash
# Pod state (will show OOMKilled or CrashLoopBackOff)
kubectl get pods -n demo-oom -w
```

Within roughly 30 seconds to 1 minute, the pod transitions to `OOMKilled` -> `CrashLoopBackOff`.

### Step 4: Inspect the KubeRCA analysis

When Alertmanager fires `KubeOOMKilled` or `KubePodCrashLooping`, KubeRCA automatically:

1. Collects Kubernetes events, pod logs, and resource state
2. Runs Root Cause Analysis through the LLM
3. Persists the analysis result

In the browser, open `http://kube-rca.local` and:

- Find the new alert in the **Incidents** list
- Click the alert to view the **RCA analysis** and **remediation guide**

### Step 5: Clean up

```bash
kubectl delete namespace demo-oom
```

---

## Next steps

To take KubeRCA to production:

- Manage AI API keys via ExternalSecrets / SealedSecrets
- Enable Slack integration (`backend.slack.enabled: true`)
- Configure OIDC (Google, Okta, GitHub, etc.)
- Wire in Prometheus / Loki / Tempo for richer context collection
- Tune resource requests / limits

For the full set of Helm values and detailed configuration, see the [Helm Chart README](https://github.com/kube-rca/kuberca/blob/main/charts/kube-rca/README.md).

**For production rollouts and technical support, reach out at [contact@cloudbro.ai](mailto:contact@cloudbro.ai).**
