# agent

## Overview

This service receives Alertmanager webhook payloads from the backend, performs basic analysis,
and returns the results to the backend.

## Requirements

- Go 1.22+

## Setup

```bash
cd agent
go mod tidy
```

## Run

```bash
cd agent
go run .
```

The server listens on `:8082` by default. Set `PORT` to change it.

## Endpoints

- `GET /ping`
- `GET /healthz`
- `GET /`
- `POST /analyze/alertmanager`

## Example Request Payload

```json
{
  "alert": {
    "status": "firing",
    "labels": {
      "alertname": "HighMemoryUsage",
      "severity": "critical",
      "namespace": "default",
      ...
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

`...`는 생략 표기입니다.

## Response Schema

```json
{
  "status": "ok",
  "thread_ts": "1234567890.123456",
  "analysis": "Analysis Complete!"
}
```

## Notes

- `analysis`는 분석 기능 구현 전까지 고정 메시지를 반환합니다.
