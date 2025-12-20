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

## Example Request

```bash
curl -X POST http://localhost:8082/analyze/alertmanager \
  -H 'Content-Type: application/json' \
  -d '{"receiver":"backend","status":"firing","alerts":[]}'
```

## Response Schema

```json
{
  "status": "ok",
  "receiver": "kube-rca-backend",
  "analysis": "Analysis Complete!"
}
```

## Notes

- `analysis`는 분석 기능 구현 전까지 고정 메시지를 반환합니다.
