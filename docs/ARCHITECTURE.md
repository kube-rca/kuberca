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
4. Frontend는 인증 초기화를 위해 `/api/v1/auth/config` 및 `/api/v1/auth/refresh`를 호출합니다.
5. 인증 후 Frontend는 `GET /api/v1/incidents` 및 `GET /api/v1/incidents/:id`로 데이터를 조회하고,
   `PUT /api/v1/incidents/:id`로 수정합니다.
6. Embedding 요청은 `POST /api/v1/embeddings`로 처리되며 Gemini 임베딩 API를 호출합니다.

## 구성 요소(현재 구현)

- Backend (Gin, 기본 포트 `:8080`)
  - `GET /ping`, `GET /`, `GET /openapi.json`
  - `POST /webhook/alertmanager`
  - `POST /api/v1/auth/register|login|refresh|logout`, `GET /api/v1/auth/config|me`
  - `GET /api/v1/incidents`, `GET /api/v1/incidents/:id`, `PUT /api/v1/incidents/:id`, `POST /api/v1/incidents/mock`
  - `POST /api/v1/embeddings`
  - 환경 변수: `SLACK_BOT_TOKEN`, `SLACK_CHANNEL_ID`, `AGENT_URL`, `AI_API_KEY`, `DATABASE_URL`, `JWT_SECRET`,
    `JWT_ACCESS_TTL`, `JWT_REFRESH_TTL`, `ALLOW_SIGNUP`, `ADMIN_USERNAME`, `ADMIN_PASSWORD`, `AUTH_COOKIE_*`,
    `CORS_ALLOWED_ORIGINS`
- Agent (FastAPI, 기본 포트 `:8000`)
  - `GET /ping`, `GET /healthz`, `GET /`, `POST /analyze`
  - Strands Agents(Gemini) + K8s/Prometheus 컨텍스트 기반 분석
  - `GEMINI_API_KEY` 미설정 시 fallback 요약 반환
- Frontend (React)
  - 로그인/회원가입 UI 및 Incident 목록/상세 UI
  - `Authorization: Bearer` + refresh cookie 사용
- PostgreSQL
  - incidents/auth/embeddings 저장소로 사용 (auth 스키마는 런타임에서 생성)

## 미구현/계획

- Vector DB 연동 및 유사 인시던트 검색
- LLM 기반 대응 가이드 자동 생성 고도화
- 멀티 테넌트/멀티 클러스터 확장 기능
