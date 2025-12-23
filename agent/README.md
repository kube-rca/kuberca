# agent

## Overview

This service receives Alertmanager webhook payloads from the backend, performs
RCA (Root Cause Analysis) using Strands Agents and in-cluster Kubernetes APIs,
and returns the results to the backend.

## Requirements

- Python 3.10+
- uv

## Setup

```bash
cd agent
uv venv
source .venv/bin/activate
uv pip install -e ".[dev]"
```

## Run

```bash
cd agent
uvicorn app.main:app --host 0.0.0.0 --port 8000
```

The server listens on `:8000` by default. Set `PORT` to change it.

## Environment Variables

- `GEMINI_API_KEY`: Gemini API key (Secret-based in Helm deployment).
- `GEMINI_MODEL_ID`: Gemini model ID (default: `gemini-3-flash-preview`).
- `K8S_API_TIMEOUT_SECONDS`: Kubernetes API timeout in seconds (default: `5`).
- `K8S_EVENT_LIMIT`: Maximum number of events to fetch (default: `20`).
- `K8S_LOG_TAIL_LINES`: Number of previous log lines to fetch (default: `50`).
- `PROMETHEUS_LABEL_SELECTOR`: Label selector for Prometheus Service (default:
  `app=kube-prometheus-stack-prometheus`).
- `PROMETHEUS_NAMESPACE_ALLOWLIST`: Comma-separated namespaces to search (default: empty = all).
- `PROMETHEUS_PORT_NAME`: Service port name to use when multiple ports exist (default: empty).
- `PROMETHEUS_SCHEME`: Prometheus scheme (default: `http`).
- `PROMETHEUS_HTTP_TIMEOUT_SECONDS`: Prometheus HTTP timeout in seconds (default: `5`).

## Endpoints

- `GET /ping`
- `GET /healthz`
- `GET /`
- `POST /analyze`

## Curl Test

```bash
curl -X POST http://localhost:8000/analyze \
  -H 'Content-Type: application/json' \
  -d @- <<'JSON'
{
  "alert": {
    "status": "firing",
    "labels": {
      "alertname": "TestAlert",
      "severity": "warning",
      "namespace": "default",
      "pod": "example-pod"
    },
    "annotations": {
      "summary": "test summary",
      "description": "test description"
    },
    "startsAt": "2024-01-01T00:00:00Z",
    "endsAt": "0001-01-01T00:00:00Z",
    "generatorURL": "",
    "fingerprint": "test-fingerprint"
  },
  "thread_ts": "test-thread"
}
JSON
```

Or use Makefile:

```bash
cd agent
make curl-analyze
```

Override values if needed:

```bash
make curl-analyze ANALYZE_URL=http://localhost:8000/analyze \
  THREAD_TS=test-thread ALERT_NAMESPACE=default ALERT_POD=example-pod
```

## Example Request Payload

```json
{
  "alert": {
    "status": "firing",
    "labels": {
      "alertname": "HighMemoryUsage",
      "severity": "critical",
      "namespace": "default",
      "pod": "example-pod"
    },
    "annotations": {
      "summary": "...",
      "description": "..."
    },
    "startsAt": "2024-01-01T00:00:00Z",
    "endsAt": "...",
    "fingerprint": "abc123..."
  },
  "thread_ts": "1234567890.123456"
}
```

`...` is omitted.

## Response Schema

```json
{
  "status": "ok",
  "thread_ts": "1234567890.123456",
  "analysis": "..."
}
```

## Notes

- For Kubernetes queries, alerts should include `namespace` and `pod` labels.
- If `GEMINI_API_KEY` is missing, the service returns a fallback summary.
- Use `pytest` for tests:

```bash
cd agent
pytest
```

## Makefile

```bash
cd agent
make help
make lint
make curl-analyze
make test
make run
```

### OOMKilled Test

Create an OOMKilled pod in the local Kubernetes cluster and test the analyze endpoint.

#### Prerequisites

- `kubectl` installed and connected to a cluster
- `kube-rca` namespace exists (not created automatically)

```bash
kubectl create namespace kube-rca
```

#### Targets

| Target | Description |
|---|---|
| `test-oom-only` | Create OOM pod (without calling analyze) |
| `cleanup-oom` | Cleanup test deployment |
| `test-analysis` | Create OOM pod + call analyze |
| `test-analysis-local` | Run local agent server + full test |

#### Usage Examples

```bash
# Create OOM pod only (without calling analyze)
make test-oom-only

# Auto cleanup after test
make test-oom-only CLEANUP=true

# Cleanup existing deployment
make cleanup-oom

# Use specific context
make test-oom-only KUBE_CONTEXT=my-cluster

# Full test (local agent server + OOM + analyze)
GEMINI_API_KEY=xxx KUBECONFIG=~/.kube/config make test-analysis-local
```

#### Environment Variables

| Variable | Description | Default |
|---|---|---|
| `KUBE_CONTEXT` | Kubernetes context | current-context |
| `LOCAL_OOM_NAMESPACE` | Namespace | `kube-rca` |
| `LOCAL_OOM_DEPLOYMENT` | Deployment name | `oomkilled-test` |
| `LOCAL_OOM_IMAGE` | Container image | `python:3.11-alpine` |
| `LOCAL_OOM_MEMORY_LIMIT` | Memory limit | `64Mi` |
| `CLEANUP` | Cleanup after test | `false` |
| `WAIT_SECONDS` | OOM wait timeout | `120` |

#### Direct Script Execution

```bash
# Check help
bash scripts/curl-test-oomkilled.sh --help

# Run with environment variables
DEPLOYMENT_NAME=my-oom-test \
IMAGE=python:3.11-alpine \
OOM_COMMAND="python -c 'a=bytearray(200000000)'" \
MEMORY_LIMIT=64Mi \
NAMESPACE=kube-rca \
SKIP_ANALYZE=true \
bash scripts/curl-test-oomkilled.sh
```
