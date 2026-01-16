```mermaid
sequenceDiagram
  autonumber
  participant AM as Alertmanager
  participant BE as Backend - 구현
  participant SL as Slack
  participant AG as Agent - 구현
  participant LLM as Gemini API - 구현
  participant DB as PostgreSQL + pgvector - 구현
  participant SDB as Session DB - 구현
  participant FE as Frontend - 구현

  Note over AM,BE: Alert 수신 및 Incident 연결
  AM->>BE: Webhook alert
  BE->>DB: getOrCreateIncident - firing Incident 조회 또는 생성
  BE->>DB: alerts 저장 - incident_id FK 연결
  opt resolved
    BE->>DB: resolved_at 업데이트
  end
  BE->>SL: 알림 전송 및 thread_ts 생성
  opt firing
    BE->>DB: thread_ts 저장
  end

  Note over BE,AG: 개별 Alert 실시간 분석
  BE->>AG: POST /analyze - goroutine 비동기
  AG->>SDB: 세션 조회 및 저장
  AG->>LLM: LLM 분석
  LLM-->>AG: 분석 결과
  AG-->>BE: 분석 결과
  BE->>DB: alerts.analysis_summary/detail 업데이트
  BE->>SL: 분석 결과 스레드 전송

  Note over FE,BE: Incident 종료 및 최종 분석
  FE->>BE: POST /api/v1/incidents/:id/resolve
  BE->>DB: incidents.status = resolved
  BE->>AG: POST /summarize-incident - goroutine 비동기
  AG->>LLM: 연결된 Alert 분석 종합
  LLM-->>AG: title + summary + detail
  AG-->>BE: 최종 분석 결과
  BE->>DB: incidents.analysis_summary/detail 저장
  BE->>LLM: 임베딩 생성 - text-embedding-004
  BE->>DB: embeddings 테이블 저장

  Note over FE,DB: 유사 인시던트 검색 - 구현
  FE->>BE: POST /api/v1/embeddings/search
  BE->>LLM: 쿼리 임베딩 생성
  BE->>DB: pgvector cosine similarity 검색
  DB-->>BE: 유사 인시던트 목록
  BE-->>FE: similarity 점수와 함께 반환
```
