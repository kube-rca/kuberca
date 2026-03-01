# KubeRCA 아키텍처

이 문서는 현재 `main` 워크트리 기준 런타임 흐름을 요약합니다. 상세 도식은 `.github/diagrams/`를 참고합니다.

## 런타임 흐름

### 1. Alert 수신 및 Incident 연결
1. Alertmanager가 `POST /webhook/alertmanager`로 Backend에 알림을 전달합니다.
2. Backend는 firing 상태 Incident를 조회하거나 새로 생성합니다.
3. Alert를 `alerts`에 저장하고 Incident와 연결합니다 (Incident:Alert = 1:N).
4. Backend는 Slack 메시지를 전송하고 thread_ts를 저장합니다.

### 2. 개별 Alert 분석 및 아티팩트 저장
5. Backend는 Agent에 `POST /analyze`를 비동기 요청합니다 (goroutine).
6. Agent는 K8s/Prometheus/Tempo 컨텍스트를 수집하고 Strands Agents로 분석합니다.
7. Backend는 `alerts.analysis_summary/detail`, `alert_analyses`, `alert_analysis_artifacts`를 저장합니다.
8. Backend는 Slack thread에 분석 결과를 전송합니다.

### 3. 실시간 UI 반영
9. Backend는 분석/상태 변경 이벤트를 SSE로 발행합니다 (`GET /api/v1/events`).
10. Frontend는 SSE 이벤트를 수신해 목록을 갱신하고, SSE 단절 시 polling fallback으로 동기화합니다.

### 4. Incident 종료 및 유사 Incident 검색
11. Frontend에서 `POST /api/v1/incidents/:id/resolve`로 Incident 종료를 요청합니다.
12. Backend는 Agent에 `POST /summarize-incident`를 비동기 요청합니다.
13. Agent는 Incident의 Alert 분석들을 종합해 `title/summary/detail`을 반환합니다.
14. Backend는 Incident 분석 결과를 저장하고 임베딩을 생성해 `embeddings`에 저장합니다.
15. Frontend는 `POST /api/v1/embeddings/search`로 유사 Incident를 조회합니다.

### 5. 인증 및 OIDC
16. Frontend는 초기화 시 `/api/v1/auth/config`, `/api/v1/auth/refresh`를 호출합니다.
17. OIDC 활성화 시 `/api/v1/auth/oidc/login` → Provider 인증 → `/api/v1/auth/oidc/callback` 흐름으로 로그인합니다.
18. Backend는 JWT Access Token + Refresh Cookie를 발급하고 보호 API를 제공합니다.

### 6. 운영 기능 (Chat/Feedback/Webhook Settings)
19. Frontend의 AI Chat 패널은 Backend `POST /api/v1/chat`을 호출합니다.
20. Backend는 Agent `POST /chat`으로 컨텍스트 기반 응답을 요청합니다.
21. Frontend는 Incident/Alert별 피드백 API로 코멘트/투표를 처리합니다.
22. Frontend Settings에서 Webhook 설정 CRUD를 수행합니다 (`/api/v1/settings/webhooks*`).

## 구성 요소

### Backend (Gin, 기본 포트 `:8080`)
- Public
  - `GET /ping`, `GET /`, `GET /openapi.json`
  - `POST /webhook/alertmanager`
- Auth
  - `POST /api/v1/auth/register|login|refresh|logout`
  - `GET /api/v1/auth/config|me`
  - `GET /api/v1/auth/oidc/login`, `GET /api/v1/auth/oidc/callback`
- Incident
  - `GET /api/v1/incidents`, `GET /api/v1/incidents/:id`, `PUT /api/v1/incidents/:id`
  - `PATCH /api/v1/incidents/:id`, `GET /api/v1/incidents/hidden`, `PATCH /api/v1/incidents/:id/unhide`
  - `POST /api/v1/incidents/:id/resolve`, `GET /api/v1/incidents/:id/alerts`, `POST /api/v1/incidents/mock`
- Alert
  - `GET /api/v1/alerts`, `GET /api/v1/alerts/:id`, `PUT /api/v1/alerts/:id/incident`
- Feedback
  - Incident: `GET /api/v1/incidents/:id/feedback`, `POST /api/v1/incidents/:id/comments`, `PUT /api/v1/incidents/:id/comments/:commentId`, `DELETE /api/v1/incidents/:id/comments/:commentId`, `POST /api/v1/incidents/:id/vote`
  - Alert: `GET /api/v1/alerts/:id/feedback`, `POST /api/v1/alerts/:id/comments`, `PUT /api/v1/alerts/:id/comments/:commentId`, `DELETE /api/v1/alerts/:id/comments/:commentId`, `POST /api/v1/alerts/:id/vote`
- Embedding/Chat/Settings
  - `POST /api/v1/embeddings`, `POST /api/v1/embeddings/search`
  - `POST /api/v1/chat`
  - `GET|POST /api/v1/settings/webhooks`, `GET|PUT|DELETE /api/v1/settings/webhooks/:id`
- Realtime
  - `GET /api/v1/events` (SSE)

### Agent (FastAPI, 기본 포트 `:8000`)
- `GET /ping`, `GET /healthz`, `GET /`
- `POST /analyze` - 개별 Alert 분석
- `POST /summarize-incident` - Incident 종합 분석
- `POST /chat` - 컨텍스트 기반 Q&A
- Strands Agents(`gemini|openai|anthropic`) + K8s/Prometheus/Tempo 컨텍스트 분석
- Provider API key 미설정 시 fallback 요약/안내 응답

### Frontend (React + TypeScript)
- 인증 초기화 + OIDC 로그인 UI
- Incident/Alert/Hidden(muted) 목록 및 상세
- Feedback(투표/코멘트) UI
- AI Chat 패널
- Webhook 설정 관리 페이지
- SSE 실시간 반영 + polling fallback

### PostgreSQL + pgvector
- 핵심 테이블: `incidents`, `alerts`, `alert_analyses`, `alert_analysis_artifacts`, `embeddings`, `users`, `refresh_tokens`
- 운영 테이블: `feedback_votes`, `feedback_comments`, `webhook_configs`
- 세션 저장 활성화 시 Agent session 테이블(`strands_*`, `kube_rca_session_summaries`) 사용

## 확장 통합 항목

- Slack Slash Command 기반 Incident 조회/요약
- Tempo/Loki/Grafana/Alloy 기반 관측 연동 고도화
