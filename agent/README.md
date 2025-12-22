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

### OOMKilled 테스트

로컬 Kubernetes 클러스터에서 OOMKilled pod를 생성하고 analyze endpoint를 테스트합니다.

#### 사전 요구사항

- `kubectl` 설치 및 클러스터 연결
- `kube-rca` 네임스페이스 존재 (자동 생성되지 않음)

```bash
kubectl create namespace kube-rca
```

#### 타겟

| 타겟 | 설명 |
|------|------|
| `test-oom-only` | OOM pod 생성 (analyze 호출 없이) |
| `cleanup-oom` | 테스트 deployment 정리 |
| `test-analysis` | OOM pod 생성 + analyze 호출 |
| `test-analysis-local` | 로컬 agent 서버 실행 + 전체 테스트 |

#### 사용 예시

```bash
# OOM pod만 생성 (analyze 호출 없이)
make test-oom-only

# 테스트 후 자동 정리
make test-oom-only CLEANUP=true

# 기존 deployment 정리
make cleanup-oom

# 특정 context 사용
make test-oom-only KUBE_CONTEXT=my-cluster

# 전체 테스트 (로컬 agent 서버 + OOM + analyze)
GEMINI_API_KEY=xxx KUBECONFIG=~/.kube/config make test-analysis-local
```

#### 환경변수

| 변수 | 설명 | 기본값 |
|------|------|--------|
| `KUBE_CONTEXT` | Kubernetes context | current-context |
| `LOCAL_OOM_NAMESPACE` | 네임스페이스 | `kube-rca` |
| `LOCAL_OOM_DEPLOYMENT` | Deployment 이름 | `oomkilled-test` |
| `LOCAL_OOM_IMAGE` | 컨테이너 이미지 | `python:3.11-alpine` |
| `LOCAL_OOM_MEMORY_LIMIT` | 메모리 제한 | `64Mi` |
| `CLEANUP` | 테스트 후 정리 여부 | `false` |
| `WAIT_SECONDS` | OOM 대기 타임아웃 | `120` |

#### 스크립트 직접 실행

```bash
# 도움말 확인
bash scripts/curl-test-oomkilled.sh --help

# 환경변수로 실행
DEPLOYMENT_NAME=my-oom-test \
IMAGE=python:3.11-alpine \
OOM_COMMAND="python -c 'a=bytearray(200000000)'" \
MEMORY_LIMIT=64Mi \
NAMESPACE=kube-rca \
SKIP_ANALYZE=true \
bash scripts/curl-test-oomkilled.sh
```
