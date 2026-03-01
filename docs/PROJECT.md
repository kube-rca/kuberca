# KubeRCA 프로젝트 개요

KubeRCA는 Kubernetes 환경에서 발생하는 알람을 기반으로 인시던트 컨텍스트를 자동 수집하고,
LLM을 활용해 Root Cause Analysis(RCA)와 대응 가이드를 제공하는 도구입니다.

> 참고: Helm 차트/리소스 식별자는 호환성을 위해 `kube-rca`를 사용합니다.

## 주요 기능 (현재 main 기준)

- Backend: Alertmanager Webhook 수신 및 Slack 스레드 알림
  - Auth(JWT + Refresh Cookie + OIDC), Incident/Alert/Embedding API, OpenAPI 제공
  - Hidden Incident 관리 API(`/api/v1/incidents/hidden`, `/api/v1/incidents/:id/unhide`)
  - Feedback API(incident/alert 투표/코멘트)
  - Webhook Settings CRUD API(`/api/v1/settings/webhooks*`)
  - Chat API(`/api/v1/chat`) + SSE 이벤트 스트림(`/api/v1/events`)
- Agent: FastAPI 기반 분석/요약/채팅 API
  - `POST /analyze`, `POST /summarize-incident`, `POST /chat`
  - K8s/Prometheus/Tempo 컨텍스트 + Strands Agents(`gemini|openai|anthropic`) 분석
  - Provider API key 미설정 시 fallback 응답 반환
- Frontend: Incident/Alert 운영 UI
  - 인증/로그인 + OIDC
  - Incident/Alert/Hidden 상세 관리
  - 유사 Incident 검색(`/api/v1/embeddings/search`)
  - Feedback UI, Webhook Settings UI, Floating Chat 패널
  - SSE 실시간 갱신 + polling fallback
- Helm: backend/agent/frontend 배포용 `kube-rca` 차트
- DB: PostgreSQL + pgvector (`incidents`, `alerts`, `alert_analyses`, `embeddings`, `feedback_*`, `webhook_configs`)

상세 런타임 흐름은 `ARCHITECTURE.md`, 다이어그램은 `diagrams/`를 참고합니다.

## 1. 프로젝트 배경

MSA 기반 환경에서는 장애 발생 시 원인 파악과 영향 범위 추적이 어렵고,
초기 대응 품질이 개인 경험에 의존하기 쉽습니다.

대표적인 운영 문제:
- 과거 유사 장애 탐색이 느리고 비체계적임
- 대응 히스토리(원인/조치/결과)가 분산되어 재사용이 어려움
- 알람 시점 로그/메트릭/이벤트 수집이 수동적임

KubeRCA는 알람 발생 즉시 컨텍스트를 수집하고, AI 분석과 유사 Incident 검색을 결합해
사람이 빠르게 의사결정할 수 있는 형태로 제공합니다.

## 2. 프로젝트 주제

KubeRCA - AI 기반 Kubernetes Incident 분석 및 RCA + 대응 가이드 자동화

Prometheus/Alertmanager, Slack, Kubernetes API, LLM을 연동하여
알람 수신부터 분석/공유/회고까지의 운영 루프를 단축하는 것을 목표로 합니다.

## 3. Core Workflow

1. Alertmanager Webhook 수신 및 Incident/Alert 저장
2. Agent 분석 요청(`POST /analyze`)과 Slack thread 결과 전송
3. Incident 종료 시 종합 분석(`POST /summarize-incident`) + 임베딩 저장
4. Frontend에서 유사 Incident 검색 및 대응 참고
5. 운영자가 Feedback(투표/코멘트)와 Chat, Webhook Settings로 운영 품질 개선
6. SSE + polling fallback으로 UI 상태를 실시간 동기화

## 4. 프로젝트 목적

### 4.1 인시던트 발생 즉시 컨텍스트 자동 수집
- Alertmanager 알람 수신과 함께 Incident/Alert 데이터 모델에 연결
- 장애 시점의 로그/메트릭/이벤트/트레이스 컨텍스트를 자동 수집

### 4.2 LLM 기반 RCA 추론 및 설명
- 수집된 원격 측정 데이터와 알람 메타데이터 기반 RCA 생성
- 요약/상세 결과를 Slack thread와 UI에 일관 포맷으로 제공

### 4.3 과거 장애 이력 기반 대응 가이드 제공
- pgvector 기반 유사 Incident 검색 Top-N 제공
- 기존 Incident 분석 이력으로 대응 패턴 재사용

### 4.4 운영 지식 자산화
- Feedback/Comment로 분석 결과에 대한 집단 검증 루프 형성
- Chat 인터페이스로 Incident 컨텍스트 기반 질의응답 지원
- Webhook 설정으로 외부 운영 채널 연동 확장

## 5. 프로젝트 가치

### 5.1 운영 생산성
- 반복 수작업(컨텍스트 수집, 과거 검색, 요약 공유) 자동화
- 실시간 이벤트 동기화로 화면 재조회 비용 감소

### 5.2 대응 품질 표준화
- 일관된 RCA 포맷과 히스토리 저장으로 편차 축소
- Feedback 루프를 통해 분석 품질 개선 가능

### 5.3 확장성
- Helm 배포 구조로 조직별 운영 환경 적용 용이
- LLM provider, 관측 스택, ChatOps 채널 확장 가능

## 6. 기술 스택

### 6.1 인프라 / IaC
- Terraform, Helm, Kubernetes(AWS EKS)

### 6.2 Observability
- Prometheus, Alertmanager, Grafana, Loki, Alloy
- Tempo 연동(선택)

### 6.3 애플리케이션
- Backend: Go + Gin
- Agent: Python + FastAPI + Strands Agents
- Frontend: React + TypeScript + Vite + Tailwind CSS

### 6.4 데이터 / AI
- PostgreSQL + pgvector
- Gemini/OpenAI/Anthropic provider 기반 분석
- Agent session 저장소(PostgreSQL, 선택)

### 6.5 테스트 / 검증
- Chaos Mesh 시나리오
- k6 기반 부하 테스트

## 7. 구현 방향성

### 7.1 데이터 수집 신뢰성 강화
- Alert 수집 파이프라인 안정화
- Incident/Alert/Analysis 저장 정합성 강화

### 7.2 분석 품질 루프 강화
- 분석 근거(artifacts) 저장 일관성 유지
- Feedback 신호를 활용한 운영 튜닝 기반 확보

### 7.3 운영 UX 고도화
- SSE 실시간 반영 + polling fallback 안정화
- Chat/Settings UI를 통한 운영 액션 집중화

### 7.4 배포/문서 일관성 유지
- Helm values/README/아키텍처 문서 동기화
- 컴포넌트 변경 시 `.github`와 `skills` 동시 갱신
