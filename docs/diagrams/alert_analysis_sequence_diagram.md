```mermaid
sequenceDiagram
  autonumber
  participant AM as Alertmanager
  participant BE as Backend - 구현
  participant SL as Slack
  participant AG as Agent - 구현
  participant LLM as Gemini API - 구현
  participant DB as PostgreSQL - 구현
  participant FE as Frontend - 구현
  participant VDB as Vector DB - 계획

  AM->>BE: Webhook alert
  BE->>SL: 알림 전송 thread 저장
  BE->>AG: POST /analyze
  AG->>LLM: LLM 분석
  LLM-->>AG: 분석 결과
  AG-->>BE: 분석 결과
  BE->>SL: 분석 결과 스레드 전송
  Note over BE,AG: 현재 구현

  opt 인증된 인시던트 조회
    FE->>BE: GET /api/v1/incidents
    BE->>DB: incidents 조회
    DB-->>BE: incident list
    BE-->>FE: incident list
  end

  opt 계획
    BE->>VDB: 임베딩 검색
    VDB-->>BE: 유사 인시던트
  end
```
