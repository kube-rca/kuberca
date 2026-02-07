```mermaid
flowchart LR
  %% External
  AM[Alertmanager]
  SL[Slack Bot]
  SC[Slack Slash Command 계획]
  LLM[LLM API Gemini OpenAI Anthropic]
  PR[Prometheus]
  K8S[Kubernetes API]
  GK[Grafana 계획]
  LO[Loki 계획]
  TP[Tempo 계획]
  AL[Alloy 계획]

  %% Internal
  subgraph Core[" "]
    direction TB
    FE[Frontend UI: auth + incidents + alerts + embedding search + muted]
    BE[Backend API: webhook + auth + incidents + alerts + embedding + hidden]
  end
  style Core fill:transparent,stroke:transparent

  AG[Agent API: Strands + K8s + summarize]
  PG[(PostgreSQL + pgvector: incidents alerts auth embeddings alert_analyses artifacts)]
  SDB[(Session DB optional)]

  AM -->|Webhook alert| BE
  BE -->|Slack message| SL
  SC -.->|Slash query 계획| BE

  FE -->|Auth incidents alerts embedding API| BE
  BE -->|POST /analyze, /summarize-incident| AG
  AG -->|Analysis result| BE

  AG -->|K8s query| K8S
  AG -->|PromQL query| PR
  AG -.->|Trace query 계획| TP
  AG -->|LLM inference| LLM
  BE -->|Embedding create| LLM

  BE -->|Incidents alerts auth embeddings store| PG
  BE -->|Cosine similarity search| PG
  AG -->|Session store optional| SDB

  AL -.->|Collector 계획| PR
  AL -.->|Collector 계획| LO
  AL -.->|Collector 계획| TP
  GK -.->|Dashboard 계획| PR
  GK -.->|Dashboard 계획| LO
  GK -.->|Dashboard 계획| TP
```
