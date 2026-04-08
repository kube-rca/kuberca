# KubeRCA 설치 가이드

Kubernetes 알람 기반 Root Cause Analysis(RCA) 도구, KubeRCA를 설치하고 첫 번째 데모를 실행하는 가이드입니다.

---

## 사전 요구 사항

| 항목 | 최소 요구 사항 |
|------|---------------|
| Kubernetes | v1.32 이상 |
| Helm | v3.x |
| kubectl | 클러스터에 연결된 상태 |
| AI API Key | [Google AI Studio API Key](https://aistudio.google.com/apikey) (권장, 분석+임베딩 겸용) |

### 클러스터 리소스

| 컴포넌트 | CPU (request) | Memory (request) | 비고 |
|----------|--------------|-------------------|------|
| Backend | 100m | 128Mi | Go 서버 |
| Agent | 200m | 256Mi | Python FastAPI + LLM 호출 |
| Frontend | 50m | 64Mi | Nginx 정적 서빙 |
| PostgreSQL | 250m | 256Mi | pgvector 포함 |
| **합계** | **~600m** | **~704Mi** | 최소 기준, 운영 시 조정 권장 |

> **참고:** 위 리소스는 최소 동작 기준입니다. values.yaml에서 `resources.requests`/`limits`를 운영 환경에 맞게 조정하세요.

### 사전 확인

```bash
# Kubernetes 버전 확인
kubectl version --short

# Helm 버전 확인
helm version --short
```

---

## 설치 (5분 이내)

### 1단계: values 파일 생성

`my-values.yaml` 파일을 생성합니다:

```yaml
# PostgreSQL — 차트가 Secret을 자동 생성
postgresql:
  auth:
    existingSecret: ""          # bitnami subchart가 Secret 생성
    password: "change-me"       # DB 비밀번호 설정

# Backend
backend:
  slack:
    enabled: false              # Quick Start에서는 Slack 비활성화
  postgresql:
    secret:
      existingSecret: ""        # bitnami에서 자동 참조
  embedding:
    apiKey:
      existingSecret: ""        # Agent의 AI Key Secret 공유

# Agent — AI Provider 및 API Key 설정
agent:
  aiProvider: "gemini"          # gemini | openai | anthropic
  gemini:
    apiKey: "여기에_API_KEY_입력"
    secret:
      existingSecret: ""        # apiKey로 Secret 자동 생성

# Frontend — port-forward로 접속 (Ingress 설정은 프로덕션 시 별도 구성)
```

> **다른 AI Provider 사용 시:**
>
> | Provider | `agent.aiProvider` | apiKey 필드 |
> |----------|-------------------|-------------|
> | Gemini (기본) | `"gemini"` | `agent.gemini.apiKey` |
> | OpenAI | `"openai"` | `agent.openai.apiKey` |
> | Anthropic | `"anthropic"` | `agent.anthropic.apiKey` |
>
> **주의:** Backend의 임베딩 기능은 현재 **Gemini API만 지원**합니다. OpenAI/Anthropic을 Agent에 사용하더라도 임베딩용 Gemini API Key가 별도로 필요합니다.

### 2단계: Helm 설치

```bash
helm upgrade --install kube-rca oci://public.ecr.aws/r5b7j2e4/kube-rca-ecr/charts/kube-rca \
  -n kube-rca --create-namespace \
  -f my-values.yaml
```

### 3단계: 설치 확인

```bash
kubectl get pods -n kube-rca -w
```

**아래와 같이 모든 Pod이 `Running` 상태이면 성공입니다:**

```
NAME                                READY   STATUS    RESTARTS   AGE
kube-rca-backend-xxxxxxxxxx-xxxxx   1/1     Running   0          2m
kube-rca-agent-xxxxxxxxxx-xxxxx     1/1     Running   0          2m
kube-rca-frontend-xxxxxxxxxx-xxxxx  1/1     Running   0          2m
kube-rca-postgresql-0               1/1     Running   0          2m
```

Backend 헬스체크:

```bash
kubectl port-forward svc/kube-rca-backend 8080:8080 -n kube-rca &
curl http://localhost:8080/ping
# 응답: {"message":"pong"}
```

### 4단계: UI 접속

```bash
# 터미널 1 — Frontend
kubectl port-forward svc/kube-rca-frontend 3000:80 -n kube-rca

# 터미널 2 — Backend (Frontend의 API 호출에 필요)
kubectl port-forward svc/kube-rca-backend 8080:8080 -n kube-rca
```

브라우저에서 `http://localhost:3000`을 열면 로그인 화면이 표시됩니다.

| 항목 | 기본값 |
|------|--------|
| ID | `kube-rca` |
| Password | `kube-rca` |

> 운영 환경에서는 반드시 `backend.auth.admin.username`과 `backend.auth.admin.password`를 변경하세요.

---

## 삭제

```bash
# KubeRCA 삭제
helm uninstall kube-rca -n kube-rca

# PostgreSQL PVC 및 네임스페이스 삭제 (데이터 완전 삭제)
kubectl delete pvc -n kube-rca -l app.kubernetes.io/name=postgresql
kubectl delete namespace kube-rca
```

---

## 첫 번째 시나리오: OOMKilled 장애 분석 (10분 이내)

설치 직후 바로 실행할 수 있는 데모입니다. 메모리 부족으로 Pod이 OOMKilled되는 상황을 만들고, KubeRCA가 자동으로 Root Cause Analysis를 수행하는 과정을 확인합니다.

### 사전 조건

- KubeRCA 설치 완료 (위 단계 완료)
- Alertmanager가 KubeRCA webhook과 연결된 상태

### Step 1: Alertmanager 연동

KubeRCA가 알림을 수신하려면 Alertmanager에 webhook receiver를 추가해야 합니다.

**kube-prometheus-stack 사용 시** (Helm values에 추가):

```yaml
alertmanager:
  config:
    receivers:
      - name: "kube-rca"
        webhook_configs:
          - url: "http://kube-rca-backend.kube-rca.svc.cluster.local:8080/webhook/alertmanager"
            send_resolved: true
    route:
      receiver: "kube-rca"
```

**독립 Alertmanager 사용 시** (`alertmanager.yml`에 추가):

```yaml
receivers:
  - name: "kube-rca"
    webhook_configs:
      - url: "http://kube-rca-backend.kube-rca.svc.cluster.local:8080/webhook/alertmanager"
        send_resolved: true

route:
  receiver: "kube-rca"
```

### Step 2: OOMKilled 장애 주입

메모리 제한을 극단적으로 낮춘 Pod을 배포합니다:

```bash
kubectl create namespace demo-oom 2>/dev/null || true

kubectl apply -n demo-oom -f - <<'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: memory-hog
  labels:
    app: memory-hog
spec:
  replicas: 1
  selector:
    matchLabels:
      app: memory-hog
  template:
    metadata:
      labels:
        app: memory-hog
    spec:
      containers:
        - name: stress
          image: polinux/stress:1.0.4
          command: ["stress"]
          args: ["--vm", "1", "--vm-bytes", "128M", "--vm-hang", "1"]
          resources:
            requests:
              memory: "32Mi"
            limits:
              memory: "64Mi"    # 128M 할당 시도 → OOMKilled 발생
EOF
```

### Step 3: 장애 확인

```bash
# Pod 상태 확인 (OOMKilled 또는 CrashLoopBackOff 표시)
kubectl get pods -n demo-oom -w
```

약 30초~1분 내에 Pod이 `OOMKilled` → `CrashLoopBackOff` 상태가 됩니다.

### Step 4: KubeRCA 분석 결과 확인

Alertmanager가 `KubeOOMKilled` 또는 `KubePodCrashLooping` 알림을 발생시키면, KubeRCA가 자동으로:

1. Kubernetes 이벤트, Pod 로그, 리소스 상태를 수집
2. LLM을 통해 Root Cause Analysis를 수행
3. 분석 결과를 저장

브라우저에서 `http://kube-rca.local`에 접속하여:
- **Incidents** 목록에서 새로운 알림을 확인
- 알림을 클릭하면 **RCA 분석 결과**와 **대응 가이드**가 표시됩니다

### Step 5: 정리

```bash
kubectl delete namespace demo-oom
```

---

## 다음 단계

KubeRCA를 프로덕션 환경에 적용하려면:

- AI API Key를 ExternalSecrets/SealedSecrets로 관리
- Slack 연동 활성화 (`backend.slack.enabled: true`)
- OIDC 인증 설정 (Google, GitHub 등)
- Prometheus/Loki/Tempo 연동으로 풍부한 컨텍스트 수집
- 리소스 요청/제한 조정

전체 Helm values 옵션과 상세 설정은 [Helm Chart README](https://github.com/kube-rca/kuberca/blob/main/charts/kube-rca/README.md)를 참고하세요.

**프로덕션 적용 및 기술 지원이 필요하시면 [contact@cloudbro.ai](mailto:contact@cloudbro.ai)로 문의해 주세요.**
