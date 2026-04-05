<p align="center">
  <img src="docs/img/Kube-RCA-Logo-NoBG.png" alt="KubeRCA Logo" width="200"/>
</p>

<h1 align="center">KubeRCA</h1>

<p align="center">
  <b>Kubernetes Alert Root Cause Analysis</b><br/>
  Automated incident context collection and LLM-powered RCA for Kubernetes alerts
</p>

<p align="center">
  <a href="https://github.com/kube-rca/kube-rca/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"/></a>
</p>

---

## Overview

KubeRCA is an open-source tool that automatically collects incident context from Kubernetes clusters and provides Root Cause Analysis (RCA) using LLM agents.

**Key Features:**
- Alertmanager webhook integration for automatic alert ingestion
- Kubernetes/Prometheus/Loki/Tempo context collection
- Multi-provider LLM analysis (Gemini, OpenAI, Anthropic) via Strands Agents
- Slack notification with threaded RCA results
- Web dashboard for incident management
- Helm chart for easy deployment

## Architecture

```
Alertmanager → Backend (Go/Gin) → Agent (Python/FastAPI) → LLM Analysis
                    ↓                                           ↓
               PostgreSQL                              K8s/Prometheus/Loki
                    ↓
            Frontend (React) + Slack
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed runtime flows.

## Quick Start

### Prerequisites
- Kubernetes cluster
- Helm 3.x
- Alertmanager (for alert ingestion)

### Install via Helm

```bash
helm install kube-rca oci://public.ecr.aws/r5b7j2e4/kube-rca-ecr/charts/kube-rca \
  --namespace kube-rca --create-namespace \
  --values your-values.yaml
```

See [Installation Guide (KR)](docs/installation-guide-ko.md) for detailed setup instructions.

## Repository Structure

```
├── backend/       Go REST API (Alertmanager webhook, Auth, Incident API)
├── frontend/      React web dashboard
├── agent/         Python FastAPI analysis agent (Strands Agents)
├── charts/        Helm chart for Kubernetes deployment
├── chaos/         Chaos Mesh test scenarios
└── docs/          Project documentation and diagrams
```

## Development

| Component | Commands |
|-----------|----------|
| Backend | `cd backend && go test ./...` |
| Frontend | `cd frontend && npm install && npm run dev` |
| Agent | `cd agent && make install && make test` |
| Helm | `helm lint charts/kube-rca` |

## Tech Stack

- **Backend**: Go 1.24, Gin, PostgreSQL + pgvector
- **Agent**: Python 3.10+, FastAPI, Strands Agents
- **Frontend**: React 18, TypeScript, Vite, Tailwind CSS
- **Infrastructure**: Helm, Prometheus, Alertmanager, Loki, Grafana Alloy

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
