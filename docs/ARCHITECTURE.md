# KubeRCA 아키텍처

이 문서는 런타임 흐름을 요약합니다. 상세 아키텍처는 `diagrams/`를 참고합니다.

## 런타임 흐름

### Alert 수신 및 Incident 연결
1. Alertmanager가 `POST /webhook/alertmanager`로 Backend에 알림을 전달합니다.
2. Backend는 firing 상태의 Incident를 조회하거나 새로 생성합니다 (`getOrCreateIncident`).
3. Alert를 DB에 저장하고 Incident와 연결합니다 (Incident:Alert = 1:N).
4. Backend는 Slack API로 알림 메시지를 전송합니다.
   - `firing` 알림은 새 메시지 전송 후 fingerprint 기준으로 thread_ts를 DB에 저장합니다.
   - `resolved` 알림은 저장된 thread_ts로 스레드에 답글을 전송합니다.

### 개별 Alert 실시간 분석
5. Backend는 Agent에 `POST /analyze`로 분석을 요청합니다 (goroutine 비동기).
   - Agent 응답을 받아 alerts 테이블에 analysis_summary/detail 저장.
   - 상세 히스토리는 alert_analyses/alert_analysis_artifacts에 저장합니다.
   - Slack 스레드에 분석 결과를 전송합니다.

### Incident 종료 및 최종 분석
6. Frontend에서 `POST /api/v1/incidents/:id/resolve`로 Incident 종료 요청.
7. Backend는 Agent에 `POST /summarize-incident`로 최종 분석 요청 (goroutine 비동기).
   - Agent는 연결된 모든 Alert의 분석 내용을 종합하여 title/summary/detail 반환.
8. Backend는 incidents 테이블에 analysis_summary/detail 저장.
9. Backend는 Gemini로 임베딩을 생성하고 embeddings 테이블에 저장.

### 유사 인시던트 검색
10. Frontend에서 `POST /api/v1/embeddings/search`로 유사 인시던트 검색.
11. Backend는 쿼리를 임베딩하고 pgvector cosine similarity로 검색.
12. 유사도 점수와 함께 결과를 반환합니다.

### 인증 및 UI
13. Frontend는 인증 초기화를 위해 `/api/v1/auth/config` 및 `/api/v1/auth/refresh`를 호출합니다.
14. 인증 후 Frontend는 Incident/Alert 목록 및 상세를 조회하고 수정합니다.

## 구성 요소

- Backend (Gin, 기본 포트 `:8080`)
  - `GET /ping`, `GET /`, `GET /openapi.json`
  - `POST /webhook/alertmanager`
  - `POST /api/v1/auth/register|login|refresh|logout`, `GET /api/v1/auth/config|me`
  - `GET /api/v1/incidents`, `GET /api/v1/incidents/:id`, `PUT /api/v1/incidents/:id`
  - `PATCH /api/v1/incidents/:id`, `POST /api/v1/incidents/:id/resolve`
  - `GET /api/v1/incidents/:id/alerts`, `POST /api/v1/incidents/mock`
  - `GET /api/v1/alerts`, `GET /api/v1/alerts/:id`, `PUT /api/v1/alerts/:id/incident`
  - `POST /api/v1/embeddings`, `POST /api/v1/embeddings/search`
  - 환경 변수: `SLACK_BOT_TOKEN`, `SLACK_CHANNEL_ID`, `AGENT_URL`, `AI_API_KEY`, `DATABASE_URL`, `JWT_SECRET`,
    `JWT_ACCESS_TTL`, `JWT_REFRESH_TTL`, `ALLOW_SIGNUP`, `ADMIN_USERNAME`, `ADMIN_PASSWORD`, `AUTH_COOKIE_*`,
    `CORS_ALLOWED_ORIGINS`
- Agent (FastAPI, 기본 포트 `:8000`)
  - `GET /ping`, `GET /healthz`, `GET /`
  - `POST /analyze` - 개별 Alert 실시간 분석
  - `POST /summarize-incident` - Incident 종료 시 최종 분석
  - Strands Agents(`gemini|openai|anthropic`) + K8s/Prometheus 컨텍스트 기반 분석
  - provider API key 미설정 시 fallback 요약 반환
- Frontend (React)
  - 로그인/회원가입 UI
  - Incident 목록/상세 UI (유사 인시던트 검색 포함)
  - Alert 목록/상세 UI (Incident 재할당 기능)
  - `Authorization: Bearer` + refresh cookie 사용
- PostgreSQL + pgvector
  - incidents/alerts/auth/embeddings/alert_analyses/artifacts 저장소
  - pgvector로 cosine similarity 기반 유사 인시던트 검색
  - session DB 활성화 시 STRANDS_* 및 KUBE_RCA_SESSION_SUMMARY 사용

## 확장 통합 항목

- Slack Slash Command 기반 인시던트 조회/요약
- Tempo/Loki/Grafana/Alloy 연동 확장
