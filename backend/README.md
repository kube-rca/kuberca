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

- `GEMINI_API_KEY`: Gemini API Key

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
