# KubeRCA 아키텍처

이 문서는 `code-kube-rca` 워크스페이스의 런타임 아키텍처와 주요 데이터 흐름을 요약합니다.
근거는 리포지토리 내 코드/매니페스트(예: `backend/main.go`)를 기준으로 합니다.

> 참고: 현재 리포지토리의 Helm 차트/리소스 식별자는 호환성을 위해 `kube-rca`를 사용합니다.

## 전체 흐름

```text
[Prometheus Alertmanager]
  └─ HTTP POST /webhook/alertmanager
       v
[KubeRCA backend (Go + Gin)]
  ├─ Parse: AlertmanagerWebhook(JSON)
  ├─ Filter/Process alerts
  └─ Send: Slack Web API(chat.postMessage)
       v
[Slack Channel]
```

현재 `frontend/`는 서버 연동 없이 mock 데이터로 화면을 구성합니다(`frontend/src/App.tsx`).

## 알림 파이프라인(Alertmanager → Backend → Slack)

1. Alertmanager가 backend로 `POST /webhook/alertmanager`를 호출합니다(`backend/main.go`).
2. backend가 요청 본문(JSON)을 `AlertmanagerWebhook`로 파싱합니다(`backend/internal/model/alert.go`).
3. backend가 웹훅 메타데이터 및 개별 알림을 로깅합니다(`backend/internal/handler/alert.go`).
4. `AlertService`가 알림별로 전송 여부를 판단하고(Slack 필터), 전송을 시도합니다
   (`backend/internal/service/alert.go`).
5. `SlackClient`가 Slack Web API `chat.postMessage`로 메시지를 전송합니다
   (`backend/internal/client/slack.go`, `backend/internal/client/slack_alert.go`).

## Backend 내부 구조(레이어링)

코드 상 의존성 방향은 `handler → service → client` 입니다(`backend/main.go`).

- `handler`: HTTP 요청 수신 및 JSON 파싱/응답
  - `POST /webhook/alertmanager` (`backend/internal/handler/alert.go`)
  - `GET /ping`, `GET /` (`backend/internal/handler/health.go`)
- `service`: 비즈니스 로직(필터링/전송 정책)
  - `AlertService.ProcessWebhook` (`backend/internal/service/alert.go`)
- `client`: 외부 시스템 통신(Slack)
  - `SlackClient.send` (`backend/internal/client/slack.go`)
  - `SlackClient.SendAlert` (`backend/internal/client/slack_alert.go`)

## Slack 연동

- 설정 값(환경변수)
  - `SLACK_BOT_TOKEN`
  - `SLACK_CHANNEL_ID`
  - 근거: `backend/internal/client/slack.go`
- API 호출
  - endpoint: `https://slack.com/api/chat.postMessage`
  - timeout: 10s(`http.Client{Timeout: 10 * time.Second}`)
  - 근거: `backend/internal/client/slack.go`
- 스레드 처리(메모리 기반)
  - `firing` 알림 전송 시 `fingerprint -> thread_ts`를 저장합니다.
  - `resolved` 알림은 기존 `thread_ts`로 reply 시도 후, 매핑을 삭제합니다.
  - 근거: `backend/internal/client/slack.go`, `backend/internal/client/slack_alert.go`

## Frontend 현황

- `frontend/src/App.tsx`는 `generateMockAlerts()`로 생성한 mock 데이터를 기준으로 화면을 구성합니다.
- 현재 코드 기준으로 backend API 호출 로직은 없습니다.

## 배포 관점(Kubernetes/Helm)

### Helm Chart: `helm-charts/charts/kube-rca`

- 리소스 네이밍 규칙
  - backend: `<release>-backend` (`helm-charts/charts/kube-rca/templates/_helpers.tpl`)
  - frontend: `<release>-frontend` (`helm-charts/charts/kube-rca/templates/_helpers.tpl`)
- 기본 포트(values 기준)
  - backend: `8080` (`helm-charts/charts/kube-rca/values.yaml`)
  - frontend: `80` (`helm-charts/charts/kube-rca/values.yaml`)
- 프로브
  - backend는 `GET /ping` 기반 liveness/readiness probe를 사용합니다
    (`helm-charts/charts/kube-rca/templates/backend-deployment.yaml`).

### Slack Secret 주입

- `helm-charts/charts/kube-rca/values.yaml` 기본값은 기존 Secret `kube-rca-slack`을 참조합니다.
- 기본 키
  - token: `kube-rca-slack-token`
  - channel id: `kube-rca-slack-channel-id`
  - 근거: `helm-charts/charts/kube-rca/values.yaml`
- External Secrets Operator 리소스
  - `k8s-resources/external-secrets/cluster-external-secret/kube-rca-slack-ces.yaml`가
    `kube-rca` 네임스페이스에 `kube-rca-slack` Secret을 생성/동기화합니다.

## 운영 고려사항(현재 코드 기준)

- backend는 인증/인가 로직이 없으므로(`backend/internal/handler/*.go`) 네트워크 레벨 보호가 필요합니다.
- Slack 전송은 클러스터 egress(외부 통신)가 필요합니다.
- `fingerprint -> thread_ts` 매핑은 프로세스 메모리에만 저장되어 Pod 재시작 시 유실됩니다.
