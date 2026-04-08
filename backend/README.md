<p align="center">
  <img src="https://raw.githubusercontent.com/kube-rca/.github/main/img/Kube-RCA-Logo-NoBG.png" alt="KubeRCA Logo" width="120"/>
</p>

<h1 align="center">KubeRCA Backend</h1>

<p align="center">
  <strong>Go REST API Server for Incident Management & Alert Processing</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24-00ADD8?style=flat-square&logo=go" alt="Go">
  <img src="https://img.shields.io/badge/Gin-Framework-00ADD8?style=flat-square" alt="Gin">
  <img src="https://img.shields.io/badge/PostgreSQL-pgvector-336791?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/JWT-Auth-000000?style=flat-square&logo=jsonwebtokens" alt="JWT">
</p>

---

## Overview

The KubeRCA Backend is a Go-based REST API server that serves as the central hub for the KubeRCA system. It receives Alertmanager webhooks, manages incidents and alerts, handles authentication, and coordinates with the Agent service for AI-powered analysis.

### Key Responsibilities

- Receive and process Alertmanager webhook alerts
- Create and manage incidents with alert associations
- Send Slack notifications with thread tracking
- Coordinate analysis requests with the Agent service
- Store and search incident embeddings via pgvector
- Provide JWT-based authentication and optional OIDC login
- Expose analytics and application settings APIs for operator workflows

---

## Architecture

```mermaid
flowchart LR
  AM[Alertmanager] -->|Webhook| BE[Backend]
  BE -->|Notification| SL[Slack]
  BE -->|Analyze Request| AG[Agent]
  BE -->|Embeddings| LLM[Embedding Provider]
  BE <-->|Data| PG[(PostgreSQL + pgvector)]
  FE[Frontend] <-->|REST + SSE| BE
```

---

## Tech Stack

| Category | Technology |
|----------|------------|
| **Language** | Go 1.24 |
| **Framework** | Gin |
| **Database** | PostgreSQL + pgvector |
| **Auth** | JWT (golang-jwt/jwt/v5) |
| **AI** | Google GenAI (embeddings) |
| **Container** | Docker (multi-stage build) |
| **CI/CD** | GitHub Actions |

---

## Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL with pgvector extension
- (Optional) Slack Bot Token for notifications

### Installation

```bash
# Run in repository root
# (monorepo layout: cd backend)
go mod tidy
```

### Run Development Server

```bash
go run .
```

The server starts at `http://localhost:8080` by default.

### Run Tests

```bash
go test ./...
```

### Build Binary

```bash
go build -o main .
```

---

## Merge Policy (release-please)

- Prefer `Squash and merge` (or `Rebase and merge`) into `main` to keep a single conventional commit per change.
- Avoid merge commits with conventional prefixes (e.g., `feat:`, `fix:`) to prevent duplicate changelog entries.
- Do not edit `CHANGELOG.md` manually; release-please owns it.

---

## API Endpoints

### Health & Status

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Server status |
| GET | `/ping` | Health check (returns "pong") |
| GET | `/readyz` | Readiness probe |
| GET | `/healthz` | Database-backed health probe |
| GET | `/openapi.json` | OpenAPI specification |

### Authentication (`/api/v1/auth`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/register` | User registration (when enabled) |
| POST | `/login` | User login |
| POST | `/refresh` | Refresh access token |
| POST | `/logout` | Logout (revoke refresh token) |
| GET | `/config` | Auth configuration |
| GET | `/me` | Current user info |

### Webhook

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/webhook/alertmanager` | Receive Alertmanager alerts |

### Incidents (`/api/v1/incidents`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | List incidents |
| GET | `/:id` | Get incident details |
| PUT | `/:id` | Update incident |
| PATCH | `/:id` | Hide incident |
| GET | `/hidden` | List hidden incidents |
| PATCH | `/:id/unhide` | Unhide incident |
| POST | `/:id/resolve` | Resolve incident & trigger final analysis |
| POST | `/:id/analyze` | Trigger incident analysis manually |
| GET | `/:id/alerts` | List alerts for incident |
| POST | `/mock` | Create mock incident (testing) |

### Alerts (`/api/v1/alerts`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | List alerts |
| GET | `/:id` | Get alert details |
| PUT | `/:id/incident` | Reassign alert to different incident |
| POST | `/:id/analyze` | Trigger alert analysis |
| POST | `/:id/resolve` | Manually resolve alert (Slack + Agent analysis) |
| POST | `/bulk-resolve` | Bulk resolve alerts (up to 50, Slack only) |

### Feedback (`/api/v1/incidents/:id` & `/api/v1/alerts/:id`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/incidents/:id/feedback` | Get incident feedback (votes + comments) |
| POST | `/incidents/:id/comments` | Add incident comment |
| PUT | `/incidents/:id/comments/:commentId` | Update incident comment |
| DELETE | `/incidents/:id/comments/:commentId` | Delete incident comment |
| POST | `/incidents/:id/vote` | Vote on incident analysis |
| GET | `/alerts/:id/feedback` | Get alert feedback (votes + comments) |
| POST | `/alerts/:id/comments` | Add alert comment |
| PUT | `/alerts/:id/comments/:commentId` | Update alert comment |
| DELETE | `/alerts/:id/comments/:commentId` | Delete alert comment |
| POST | `/alerts/:id/vote` | Vote on alert analysis |

### Chat (`/api/v1/chat`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/` | Send chat message (context-aware AI chat) |

### Webhook Settings (`/api/v1/settings/webhooks`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | List webhook configurations |
| POST | `/` | Create webhook configuration |
| GET | `/:id` | Get webhook configuration |
| PUT | `/:id` | Update webhook configuration |
| DELETE | `/:id` | Delete webhook configuration |

### Realtime Events (`/api/v1/events`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | SSE event stream (incident/alert updates) |

### Analytics (`/api/v1/analytics`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/dashboard` | Dashboard metrics, trends, severity distribution, and namespace breakdown |

### App Settings (`/api/v1/settings/app`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | List all app settings |
| GET | `/:key` | Get a single app setting with ENV fallback |
| PUT | `/:key` | Update a single app setting |

### Embeddings (`/api/v1/embeddings`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/` | Create embedding |
| POST | `/search` | Search similar incidents |

---

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `SLACK_BOT_TOKEN` | Slack Bot OAuth token | No |
| `SLACK_CHANNEL_ID` | Slack channel for notifications | No |
| `AGENT_URL` | Agent service base URL | No (default: `http://kube-rca-agent.kube-rca.svc:8000`) |
| `AGENT_HTTP_TIMEOUT_SECONDS` | Agent request timeout in seconds | No |
| `AGENT_RETRY_MAX_ATTEMPTS` | Agent retry attempts | No |
| `AGENT_RETRY_BASE_BACKOFF_SECONDS` | Agent retry base backoff | No |
| `AGENT_RETRY_MAX_BACKOFF_SECONDS` | Agent retry max backoff | No |
| `EMBEDDING_PROVIDER` | Embedding provider name | No (default: `google`) |
| `AI_API_KEY` | Embedding provider API key | Yes |
| `EMBEDDING_MODEL` | Embedding model ID | No (default: `text-embedding-004`) |
| `JWT_SECRET` | JWT signing secret | Yes |
| `JWT_ACCESS_TTL` | Access token TTL (e.g., `15m`) | No |
| `JWT_REFRESH_TTL` | Refresh token TTL (e.g., `168h`) | No |
| `ALLOW_SIGNUP` | Enable user registration (`true`/`false`) | No |
| `ADMIN_USERNAME` | Initial admin username | No |
| `ADMIN_PASSWORD` | Initial admin password | No |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins (comma-separated) | No |
| `AI_PROVIDER` | App setting fallback AI provider | No |
| `AI_MODEL_ID` | App setting fallback AI model | No |
| `MANUAL_ANALYZE_SEVERITIES` | Comma-separated severities requiring manual analysis | No |
| `FLAP_ENABLED` | Enable flapping detection | No |
| `FLAP_DETECTION_WINDOW_MINUTES` | Flapping detection window | No |
| `FLAP_CYCLE_THRESHOLD` | Flapping cycle threshold | No |
| `FLAP_CLEARANCE_WINDOW_MINUTES` | Flapping clearance window | No |

### Cookie Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `AUTH_COOKIE_SECURE` | Secure flag | `true` |
| `AUTH_COOKIE_SAMESITE` | SameSite policy | `Lax` |
| `AUTH_COOKIE_DOMAIN` | Cookie domain | - |
| `AUTH_COOKIE_PATH` | Cookie path | `/` |

### OIDC Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `OIDC_ENABLED` | Enable OIDC authentication | `false` |
| `OIDC_ISSUER` | OIDC issuer URL | `https://accounts.google.com` |
| `OIDC_CLIENT_ID` | OIDC client ID | - |
| `OIDC_CLIENT_SECRET` | OIDC client secret | - |
| `OIDC_REDIRECT_URI` | OIDC callback URL | - |
| `OIDC_ALLOWED_DOMAINS` | Allowed email domains (comma-separated) | - |
| `OIDC_ALLOWED_EMAILS` | Allowed email addresses (comma-separated) | - |

### Local Development

Create a `.env` file in the backend directory:

```bash
DATABASE_URL=postgres://user:pass@localhost:5432/kubereca?sslmode=disable
SLACK_BOT_TOKEN=xoxb-your-token
SLACK_CHANNEL_ID=C01234567
AGENT_URL=http://localhost:8000
EMBEDDING_PROVIDER=google
AI_API_KEY=your-embedding-api-key
EMBEDDING_MODEL=text-embedding-004
JWT_SECRET=your-secret-key
ALLOW_SIGNUP=true
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin
```

---

## Database Schema

PostgreSQL with pgvector extension is required. Key tables:

```sql
-- Enable pgvector
CREATE EXTENSION IF NOT EXISTS vector;

-- Embeddings table for similarity search
CREATE TABLE embeddings (
  id BIGSERIAL PRIMARY KEY,
  incident_id TEXT NOT NULL,
  incident_summary TEXT NOT NULL,
  embedding VECTOR(3072) NOT NULL,
  model TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## Project Structure

```text
backend/
├── main.go            # Application entrypoint and route wiring
├── openapi.go         # OpenAPI annotations
├── docs/              # Generated Swagger docs
├── internal/client/   # Slack, Agent, embedding, notifier integrations
├── internal/config/   # Environment-based configuration loading
├── internal/db/       # Schema management and database access
├── internal/handler/  # HTTP handlers
├── internal/model/    # DTOs and response models
└── internal/service/  # Incident, alert, auth, analytics, settings logic
```

---

## Development

### Generate OpenAPI Spec

```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g openapi.go --parseInternal --output docs
```

### Code Formatting

```bash
go fmt ./...
```

### Linting & Testing

```bash
go fmt ./... && go test ./...
```

---

## Docker

### Build Image

```bash
docker build -t kube-rca-backend .
```

### Run Container

```bash
docker run -d -p 8080:8080 \
  -e DATABASE_URL=postgres://... \
  -e JWT_SECRET=your-secret \
  kube-rca-backend
```

---

## API Testing

### Health Check

```bash
curl http://localhost:8080/ping
# Response: {"message": "pong"}
```

### Server Status

```bash
curl http://localhost:8080/
# Response: {"status": "ok", "message": "Gin basic API server is running"}
```

---

## Related Components

- [KubeRCA Agent](https://github.com/kube-rca/kuberca/blob/main/agent/README.md) - Python analysis service
- [KubeRCA Frontend](https://github.com/kube-rca/kuberca/blob/main/frontend/README.md) - React web dashboard
- [Helm Chart](https://github.com/kube-rca/kuberca/blob/main/charts/kube-rca/README.md) - Kubernetes deployment

---

## License

This project is part of KubeRCA, licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
