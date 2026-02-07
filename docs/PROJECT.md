# KubeRCA 프로젝트 개요

KubeRCA는 Kubernetes 환경에서 발생하는 알람을 기반으로 인시던트 컨텍스트를 자동 수집하고,
LLM을 활용해 Root Cause Analysis(RCA)와 대응 가이드를 제공하는 것을 목표로 하는 도구입니다.

> 참고: 현재 리포지토리의 Helm 차트/리소스 식별자는 호환성을 위해 `kube-rca`를 사용합니다.

## 주요 기능

- Backend: Alertmanager Webhook 수신 및 Slack 스레드 알림 전송
  - Auth(JWT + Refresh Cookie), Incident/Alert/Embedding API, OpenAPI 제공
  - Incident 숨김/복원 API(`/api/v1/incidents/hidden`, `/api/v1/incidents/:id/unhide`) 포함
- Agent: FastAPI 기반 분석 API
  - K8s/Prometheus 컨텍스트 + Strands Agents(`gemini|openai|anthropic`) 분석
  - Provider API key 미설정 시 fallback 요약 반환
  - `SESSION_DB_*` 설정 시 세션 저장 사용
- Frontend: 로그인/회원가입 + Incident/Alert 목록/상세 + 숨김(뮤트) 인시던트 UI
  - `/api/v1/auth/*`, `/api/v1/incidents*`, `/api/v1/alerts*`, `/api/v1/embeddings/search` 사용
- Helm: backend/agent/frontend + OpenAPI UI 배포용 `kube-rca` 차트 포함
- DB: PostgreSQL 연동(incident/auth/embeddings/alert_analyses/artifacts)
- Slack Slash Command 기반 조회/요약 UX
- 관측 스택 통합 고도화(Tempo/Loki/Grafana/Alloy)
- 고급 RAG/추천 파이프라인 및 대응 플레이북 자동화

상세 아키텍처와 런타임 흐름은 `ARCHITECTURE.md` 및 `diagrams/`를 참고합니다.

## 1. 프로젝트 배경

Microservice Architecture(MSA) 기반 시스템은 서비스 간 의존성이 높고 구조가 복잡합니다.
이로 인해 장애 발생 시 원인을 파악하는 과정이 어렵고, 다양한 알람 중 무엇이 중요한지
식별하는 데에도 시간이 소요됩니다.

온보딩 과정에서 신규 엔지니어는 “어디를 확인해야 하는지, 어떤 순서로 추적해야 하는지”가
명확하지 않아 혼란을 겪기 쉽습니다. 기존 운영 인력 또한 장애 분석 방식이 개인 경험에
의존하면, 동일한 장애에서도 대응 방식이 일관되지 못해 운영 비효율이 반복됩니다.

대표적인 문제는 다음과 같습니다.

- 과거에 동일/유사 장애가 있었는지 확인하기 어려움
- 과거 장애 당시 어떤 대응을 수행했는지 파악하기 어려움
- 장애 시점의 로그·메트릭·이벤트 수집이 수동적이며 비효율적임

따라서 장애 발생 즉시 관련 정보를 자동 수집하고, AI 분석을 통해 원인 후보와 근거를 제시하며,
과거 유사 인시던트 기반의 대응 가이드를 제공하는 자동화 도구가 필요해졌습니다.

## 2. 프로젝트 주제

KubeRCA – AI 기반 Kubernetes 인시던트 알람 분석 및 RCA + 대응 가이드 자동화 도구

Prometheus/Alertmanager, Slack, 로그, 메트릭, 트레이스 데이터를 연동하여 장애 발생 시
관련 정보를 수집하고, LLM을 활용해 원인 분석과 대응 가이드를 제공하는 것을 목표로 합니다.

## 3. Core Workflow

1. 알람 수신 시 관련 로그, 메트릭, 이벤트를 자동 수집
2. LLM 분석을 통해 장애 원인 후보와 추론 근거 생성
3. 과거 인시던트 유사도 비교 및 대응 가이드/체크리스트 추천
4. 결과를 Slack 및 웹 UI(대시보드)로 제공

## 4. 프로젝트 목적

### 4.1 인시던트 발생 즉시 컨텍스트 자동 수집

- Alertmanager Webhook을 통해 알람 내용을 수신
- 알람 발생 시점의 로그, 메트릭, 이벤트를 자동 수집 및 정규화

### 4.2 LLM 기반 RCA 추론 및 설명

- 알람 메시지와 수집된 원격 측정 데이터를 기반으로 LLM 분석 수행
- 장애 원인 후보와 추론 근거를 제공하여 사람이 빠르게 판단할 수 있도록 지원

### 4.3 과거 장애 이력 기반 대응 가이드 제공

- Vector DB를 활용해 유사 인시던트 Top 5 검색
- 당시 실행했던 대응 절차와 최종 RCA 자료를 기반으로 대응 가이드 제공

### 4.4 장애 지식의 자산화 및 온보딩 단축

- 인시던트 대응을 시스템화하여 개인 경험 의존도를 낮춤
- 표준 분석 플로우 제공으로 온보딩 시 “어디부터/어떻게”를 빠르게 학습 가능

## 5. 프로젝트 가치

### 5.1 생태계 기여

- Kubernetes + Observability + LLM 통합을 위한 참조 아키텍처 형태로 공개 가능
- Helm 기반 배포로 설치 장벽을 낮추고, 조직별로 확장 가능한 구조 지향
- 인시던트 데이터 스키마/프롬프트 템플릿을 기반으로 조직별 튜닝 여지 제공

### 5.2 상용화 가능성

- MTTR 단축 및 장애 분석 시간 감소에 기여
- ChatOps(Slack) 기반 워크플로우로 운영 프로세스에 자연스럽게 통합 가능

### 5.3 운영 효율/안정성

- 반복적인 수동 로그 검색/대조 작업을 자동화하여 운영 부담 감소
- 일관된 RCA 포맷 및 대응 체크리스트로 대응 품질의 편차를 줄임

## 6. 기술 스택

### 6.1 인프라 / IaC

- IaC: Terraform(리포지토리 포함)
- Packaging/Deploy: Helm(리포지토리 포함)
- Kubernetes: AWS EKS

### 6.2 Observability(연동/확장)

- Alerting: Prometheus/Alertmanager(kube-prometheus-stack 차트 포함)
- Visualization: Grafana(차트 포함)
- Collector: Grafana Alloy(차트 포함)
- Logs: Loki(차트 포함)
- Metrics/Traces 확장: Grafana Mimir, Tempo

### 6.3 알람 및 인터페이스

- Alertmanager Webhook → Backend
- Slack App/Bot(Bot Token 기반 메시지 전송)
- Slack Slash Command

### 6.4 애플리케이션

- Backend: Go + Gin(Auth/Incident/Embedding API)
- Agent: Python + FastAPI(Strands Agents 기반 분석)
- Frontend: React + TypeScript + Vite + Tailwind CSS(인증 UI 포함)

### 6.5 데이터베이스 / AI

- PostgreSQL + pgvector(incident/auth/embeddings)
- Agent 세션 저장소(PostgreSQL)
- LLM API(Gemini, Strands Agents/Embeddings)
- Vector DB: 유사 인시던트 검색

### 6.6 테스트 및 검증

- Chaos Engineering: Chaos Mesh 시나리오, AWS Fault Injection Service(FIS)
- Load Testing: k6

## 7. 구현 방향성

### 7.1 기본 인프라 및 모니터링 환경 구축

- IaC로 클러스터/기반 인프라 구성(Terraform)
- Observability 스택 설치 및 알람 파이프라인 구성

### 7.2 알람 수집 및 데이터 모델링

- Alertmanager Webhook 수신 인터페이스 고도화
- 알람 발생 시점의 로그/메트릭/이벤트를 조회해 Incident Context로 저장

### 7.3 LLM 기반 RCA 및 유사 인시던트 추천

- Incident Context + 과거 데이터 기반 RAG 파이프라인 구성
- 원인 후보, 추론 근거, 대응 가이드/체크리스트 생성
- 유사 인시던트 Top 5 검색 및 정렬 로직 구현

### 7.4 Frontend 및 Slack 통합 UX

- 인시던트 목록/상세, RCA 결과, 유사 이력, 재발 방지 액션 화면 제공
- Slack 스레드에 RCA 요약/가이드 자동 포스팅
- Slack Slash Command 기반 조회/액션

### 7.5 Chaos/Load 테스트로 데이터셋 확보 및 검증

- 장애 시나리오를 생성해 데이터셋을 확보하고, 분석 품질을 반복 개선

### 7.6 패키징 및 배포

- Helm 차트로 구성 요소를 패키징하고 설치 가이드/레퍼런스 문서화
- 멀티 클러스터/멀티 테넌트 확장성 고려
