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
uvicorn app.main:app --host 0.0.0.0 --port 8082
```

The server listens on `:8082` by default. Set `PORT` to change it.

## Environment Variables

- `GEMINI_API_KEY`: Gemini API key (Secret-based in Helm deployment).
- `GEMINI_MODEL_ID`: Gemini model ID (default: `gemini-3-flash-preview`).
- `K8S_API_TIMEOUT_SECONDS`: Kubernetes API timeout in seconds (default: `5`).
- `K8S_EVENT_LIMIT`: Maximum number of events to fetch (default: `20`).
- `K8S_LOG_TAIL_LINES`: Number of previous log lines to fetch (default: `50`).

## Endpoints

- `GET /ping`
- `GET /healthz`
- `GET /`
- `POST /analyze`

## Curl Test

```bash
curl -X POST http://localhost:8082/analyze \
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
  "thread_ts": "test-thread",
  "callback_url": "http://kube-rca-backend.kube-rca.svc:8080/callback/agent"
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
make curl-analyze ANALYZE_URL=http://localhost:8082/analyze \
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
  "thread_ts": "1234567890.123456",
  "callback_url": "http://kube-rca-backend.kube-rca.svc:8080/callback/agent"
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
