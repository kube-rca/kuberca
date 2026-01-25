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
- Provide JWT-based authentication

---

## Architecture

```mermaid
flowchart LR
  AM[Alertmanager] -->|Webhook| BE[Backend]
  BE -->|Notification| SL[Slack]
  BE -->|Analyze Request| AG[Agent]
  BE -->|Embeddings| LLM[Gemini API]
  BE <-->|Data| PG[(PostgreSQL + pgvector)]
  FE[Frontend] -->|REST API| BE
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

- Go 1.22+
- PostgreSQL with pgvector extension
- (Optional) Slack Bot Token for notifications

### Installation

```bash
cd backend
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

## API Endpoints

### Health & Status

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | Server status |
| GET | `/ping` | Health check (returns "pong") |
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
| GET | `/:id/alerts` | List alerts for incident |
| POST | `/mock` | Create mock incident (testing) |

### Alerts (`/api/v1/alerts`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/` | List alerts |
| GET | `/:id` | Get alert details |
| PUT | `/:id/incident` | Reassign alert to different incident |

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
| `AGENT_URL` | Agent service base URL | Yes |
| `AI_API_KEY` | Gemini API key for embeddings | Yes |
| `JWT_SECRET` | JWT signing secret | Yes |
| `JWT_ACCESS_TTL` | Access token TTL (e.g., `15m`) | No |
| `JWT_REFRESH_TTL` | Refresh token TTL (e.g., `168h`) | No |
| `ALLOW_SIGNUP` | Enable user registration (`true`/`false`) | No |
| `ADMIN_USERNAME` | Initial admin username | No |
| `ADMIN_PASSWORD` | Initial admin password | No |
| `CORS_ALLOWED_ORIGINS` | Allowed CORS origins (comma-separated) | No |

### Cookie Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `AUTH_COOKIE_SECURE` | Secure flag | `true` |
| `AUTH_COOKIE_SAMESITE` | SameSite policy | `Lax` |
| `AUTH_COOKIE_DOMAIN` | Cookie domain | - |
| `AUTH_COOKIE_PATH` | Cookie path | `/` |

### Local Development

Create a `.env` file in the backend directory:

```bash
DATABASE_URL=postgres://user:pass@localhost:5432/kubereca?sslmode=disable
SLACK_BOT_TOKEN=xoxb-your-token
SLACK_CHANNEL_ID=C01234567
AGENT_URL=http://localhost:8000
AI_API_KEY=your-gemini-api-key
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
  embedding VECTOR(768) NOT NULL,
  model TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## Project Structure

```
backend/
├── main.go              # Application entrypoint
├── openapi.go           # OpenAPI annotations
├── docs/                # Generated Swagger docs
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── internal/
│   ├── client/          # External service clients (Slack, Agent)
│   ├── config/          # Configuration loading
│   ├── db/              # Database connections & queries
│   ├── handler/         # HTTP handlers (routing)
│   ├── model/           # Data models/DTOs
│   └── service/         # Business logic
├── Dockerfile           # Multi-stage Docker build
└── .github/workflows/
    └── ci.yaml          # CI/CD pipeline
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

- [KubeRCA Agent](../agent/) - Python analysis service
- [KubeRCA Frontend](../frontend/) - React web dashboard
- [Helm Charts](../helm-charts/) - Kubernetes deployment

---

## License

This project is part of KubeRCA, licensed under the Apache License 2.0.
