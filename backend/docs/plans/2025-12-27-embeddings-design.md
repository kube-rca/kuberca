# Embeddings API Design

## Goals
- Add a new API endpoint to embed incident summaries using Gemini embeddings.
- Store embeddings in Postgres using pgvector with minimal impact on existing RCA flows.

## Non-Goals
- Automatic embedding on incident create/update.
- Background jobs or async queues.

## API
- **POST** `/api/v1/embeddings`
- **Request**
  ```json
  {
    "incident_id": "INC-123",
    "incident_summary": "..."
  }
  ```
- **Response**
  ```json
  {
    "status": "success",
    "embedding_id": 123,
    "model": "text-embedding-004"
  }
  ```

## Data Model
New table:
```
embeddings (
  id bigserial primary key,
  incident_id text not null,
  incident_summary text not null,
  embedding vector not null,
  model text not null,
  created_at timestamptz not null default now()
)
```

## Components
- `internal/client/genai.go`: Gemini embeddings client using `google.golang.org/genai`.
- `internal/service/embedding.go`: input validation, invoke client, persist embedding.
- `internal/db/embedding.go`: `InsertEmbedding` on `Postgres` repo.
- `internal/handler/embedding.go`: HTTP handler for POST `/api/v1/embeddings`.
- `internal/model/embedding.go`: request/response structs.

## Data Flow
1. Handler validates `incident_id` and `incident_summary`.
2. Service calls Gemini embeddings with model `text-embedding-004` using `GEMINI_API_KEY`.
3. Service stores vector in Postgres via repo.
4. Handler returns success with `embedding_id` and `model`.

## Error Handling
- 400 if required fields are empty.
- 500 for Gemini API failures or DB insert errors.

## Testing
- Optional unit tests for handler/service with stubbed client and repo.
- Ensure code compiles and routes register.
