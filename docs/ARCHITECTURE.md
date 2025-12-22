# KubeRCA 현재 아키텍처(as-is)

이 문서는 현재 구현된 런타임 흐름을 요약합니다. 목표(to-be) 아키텍처는
`diagrams/`를 참고합니다.

## 런타임 흐름(현재)

1. Alertmanager가 `POST /webhook/alertmanager`로 Backend에 알림을 전달합니다.
2. Backend는 Slack API로 알림 메시지를 전송합니다.
   - `firing` 알림은 새 메시지 전송 후 fingerprint 기준으로 thread_ts를 메모리에 저장합니다.
   - `resolved` 알림은 저장된 thread_ts로 스레드에 답글을 전송한 뒤 삭제합니다.
3. Backend는 Agent에 `POST /analyze`로 분석을 요청합니다 (동기 호출, 120s timeout).
   - Agent 응답을 받아 Slack 스레드에 분석 결과를 전송합니다.
4. Frontend는 `GET /api/rca` 호출을 시도하며, 개발 환경에서는 mock 데이터로 fallback 합니다.

## 구성 요소(현재 구현)

- Backend (Gin, 기본 포트 `:8080`)
  - `GET /ping`, `GET /`, `POST /webhook/alertmanager`
  - 환경 변수: `SLACK_BOT_TOKEN`, `SLACK_CHANNEL_ID`, `AGENT_URL`, `DATABASE_URL`
  - Agent 연동: `POST /analyze` 동기 호출 → Slack 스레드에 분석 결과 전송
- Agent (FastAPI, 기본 포트 `:8000`)
  - `GET /ping`, `GET /healthz`, `GET /`, `POST /analyze`
  - AI 기반 Root Cause Analysis 수행
- Frontend (React)
  - `frontend/src/utils/api.ts` 기준으로 `GET /api/rca`를 호출합니다.

## 미구현/계획

- RCA 결과 저장/조회 API (PostgreSQL 연결은 구현됨)
- Incident Store, Vector DB 연동
- Frontend의 실제 RCA API 연동
- K8s client를 통한 클러스터 상태 조회
