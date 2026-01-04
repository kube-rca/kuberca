아래 ERD는 현재 구현과 계획을 함께 표현합니다.
- users/refresh_tokens: 구현 (AuthService에서 스키마 생성)
- incidents/embeddings: 구현 (알림 저장/분석 결과/임베딩)
- strands_sessions/strands_agents/strands_messages: 구현 (Agent 세션 저장)
- RCA_DOCUMENT: 계획

```mermaid
erDiagram
  USER ||--o{ REFRESH_TOKEN : issues
  INCIDENT ||--o{ EMBEDDING : references
  INCIDENT ||--|| RCA_DOCUMENT : has
  STRANDS_SESSION ||--o{ STRANDS_AGENT : owns
  STRANDS_AGENT ||--o{ STRANDS_MESSAGE : stores

  USER {
    bigint id PK
    string login_id
    string password_hash
    datetime created_at
    datetime updated_at
  }

  REFRESH_TOKEN {
    bigint id PK
    bigint user_id
    string token_hash
    datetime expires_at
    datetime revoked_at
    datetime created_at
  }

  INCIDENT {
    string incident_id PK
    string alarm_title
    string severity
    string status
    datetime fired_at
    datetime resolved_at
    text analysis_summary
    text analysis_detail
    jsonb similar_incidents
    string fingerprint
    string thread_ts
    jsonb labels
    jsonb annotations
    boolean is_enabled
    datetime created_at
    datetime updated_at
  }

  EMBEDDING {
    bigint id PK
    string incident_id
    text incident_summary
    vector embedding
    string model
    datetime created_at
  }

  STRANDS_SESSION {
    string session_id PK
    jsonb data
  }

  STRANDS_AGENT {
    string session_id PK
    string agent_id PK
    jsonb data
  }

  STRANDS_MESSAGE {
    string session_id PK
    string agent_id PK
    int message_id PK
    jsonb data
  }

  RCA_DOCUMENT {
    string incident_id PK
    text content_md
    text additional_context_md
    datetime created_at
    datetime updated_at
  }
```
