# backend

## Gin API Server

이 프로젝트는 Go 언어와 Gin 프레임워크를 사용한 기본 REST API 서버입니다.

### 1. 환경 준비

- **Go 설치**: Go 1.22 이상이 설치되어 있어야 합니다.
- **모듈 의존성 설치** (처음 한 번만 실행):

```bash
cd backend
go mod tidy
```

### 2. 환경 변수(선택)

Alertmanager 웹훅을 Slack으로 전송하려면 아래 환경 변수를 설정합니다.

- `SLACK_BOT_TOKEN`: Slack Bot Token (xoxb-...)
- `SLACK_CHANNEL_ID`: Slack 채널 ID (C...)

Embeddings API를 사용하려면 아래 환경 변수를 설정합니다.

- `AI_API_KEY`: Gemini API Key

Postgres에는 pgvector 확장이 필요하며, 아래 예시처럼 테이블을 생성합니다.

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE embeddings (
  id bigserial primary key,
  incident_id text not null,
  incident_summary text not null,
  embedding vector(<dim>) not null,
  model text not null,
  created_at timestamptz not null default now()
);
```

`<dim>`은 사용 중인 모델의 임베딩 차원으로 교체합니다.

인증 기능을 사용하려면 아래 환경 변수를 설정합니다.

- `JWT_SECRET`: JWT 서명 시크릿
- `JWT_ACCESS_TTL`: 액세스 토큰 TTL (예: `15m`)
- `JWT_REFRESH_TTL`: 리프레시 토큰 TTL (예: `168h`)
- `ALLOW_SIGNUP`: 회원가입 허용 여부 (`true`/`false`)
- `ADMIN_USERNAME`: 사전 생성되는 admin ID
- `ADMIN_PASSWORD`: 사전 생성되는 admin 비밀번호
- `AUTH_COOKIE_SECURE`: 리프레시 쿠키 Secure 플래그 (`true`/`false`)
- `AUTH_COOKIE_SAMESITE`: 리프레시 쿠키 SameSite (`Lax`/`Strict`/`None`)
- `AUTH_COOKIE_DOMAIN`: 리프레시 쿠키 Domain
- `AUTH_COOKIE_PATH`: 리프레시 쿠키 Path
- `CORS_ALLOWED_ORIGINS`: CORS 허용 Origin (콤마로 구분)

로컬 테스트는 `backend/.env`로 환경 변수를 로드하며, `.env`가 없으면 무시됩니다.

### 3. 서버 실행

```bash
cd backend
go run .
```

기본적으로 `http://localhost:8080` 에서 서버가 실행됩니다.

### 4. API 테스트

- **루트 엔드포인트**

```bash
curl http://localhost:8080/
```

예상 응답:

```json
{
  "status": "ok",
  "message": "Gin basic API server is running"
}
```

- **테스트 엔드포인트 (`/ping`)**

```bash
curl http://localhost:8080/ping
```

예상 응답:

```json
{
  "message": "pong"
}
```

이 정도 구성이면 Gin을 이용해 기본적인 API 요청/응답 흐름을 테스트하기에 충분합니다.

### 5. 인증 API

모든 인증 API는 `/api/v1/auth` prefix를 사용합니다.

- `POST /api/v1/auth/register`: 회원가입 (ALLOW_SIGNUP=true일 때만)
- `POST /api/v1/auth/login`: 로그인 (access token 반환 + refresh 쿠키 설정)
- `POST /api/v1/auth/refresh`: refresh 쿠키로 access token 재발급
- `POST /api/v1/auth/logout`: refresh 토큰 폐기
- `GET /api/v1/auth/config`: `allowSignup` 반환
- `GET /api/v1/auth/me`: 액세스 토큰으로 사용자 정보 조회

### 6. OpenAPI(Swagger)

OpenAPI 스펙은 `backend/docs/`에 생성되며 Git에 포함됩니다.

```bash
cd backend
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g openapi.go --parseInternal --output docs
```

런타임 스펙 엔드포인트:

- `GET /openapi.json`

#### Git hook (선택)

커밋 시 OpenAPI를 자동 갱신하려면 hooksPath를 설정합니다.

```bash
cd backend
git config core.hooksPath .githooks
```
