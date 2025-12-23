# kube-rca

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

Deploy kube-rca backend and frontend

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| agent.affinity | object | `{}` | Affinity for agent pods assignment. |
| agent.containerPort | int | `8000` | Agent container port. |
| agent.gemini.modelId | string | `"gemini-3-flash-preview"` | Gemini model ID for Strands Agents. |
| agent.gemini.secret.create | bool | `false` | Create a Secret for the Gemini API key. |
| agent.gemini.secret.existingSecret | string | `"kube-rca-agent-secret"` | Existing Secret name for the Gemini API key. |
| agent.gemini.secret.key | string | `"ai-studio-api-key"` | Secret key name for the Gemini API key. |
| agent.image.pullPolicy | string | `"IfNotPresent"` | Agent image pull policy. |
| agent.image.repository | string | `""` | Agent image repository. |
| agent.image.tag | string | `""` | Agent image tag. |
| agent.ingress.annotations | object | `{}` | Annotations for agent ingress. |
| agent.ingress.enabled | bool | `false` | Enable agent ingress. |
| agent.ingress.hosts | list | `[]` | Hostnames for agent ingress. |
| agent.ingress.ingressClassName | string | `""` | IngressClass name for agent ingress. |
| agent.ingress.pathType | string | `"Prefix"` | PathType for agent ingress. |
| agent.ingress.paths | list | `["/"]` | Paths for agent ingress. |
| agent.ingress.tls | list | `[]` | TLS configuration for agent ingress. |
| agent.nodeSelector | object | `{}` | Node labels for agent pods assignment. |
| agent.replicaCount | int | `1` | Number of agent replicas. |
| agent.resources | object | `{}` | Agent resource requests/limits. |
| agent.service.port | int | `8000` | Agent service port. |
| agent.service.type | string | `"ClusterIP"` | Agent service type. |
| agent.tolerations | list | `[]` | Tolerations for agent pods assignment. |
| backend.affinity | object | `{}` | Affinity for backend pods assignment. |
| backend.containerPort | int | `8080` | Backend container port. |
| backend.image.pullPolicy | string | `"IfNotPresent"` | Backend image pull policy. |
| backend.image.repository | string | `""` | Backend image repository. |
| backend.image.tag | string | `""` | Backend image tag. |
| backend.ingress.annotations | object | `{}` | Annotations for backend ingress. |
| backend.ingress.enabled | bool | `false` | Enable backend ingress. |
| backend.ingress.hosts | list | `[]` | Hostnames for backend ingress. |
| backend.ingress.ingressClassName | string | `""` | IngressClass name for backend ingress. |
| backend.ingress.pathType | string | `"Prefix"` | PathType for backend ingress. |
| backend.ingress.paths | list | `["/"]` | Paths for backend ingress. |
| backend.ingress.tls | list | `[]` | TLS configuration for backend ingress. |
| backend.nodeSelector | object | `{}` | Node labels for backend pods assignment. |
| backend.postgresql.database | string | `"kube-rca"` | PostgreSQL database. |
| backend.postgresql.host | string | `"postgresql.kube-rca.svc.cluster.local"` | PostgreSQL host. |
| backend.postgresql.port | int | `5432` | PostgreSQL port. |
| backend.postgresql.user | string | `"kube-rca"` | PostgreSQL user. |
| backend.replicaCount | int | `1` | Number of backend replicas. |
| backend.resources | object | `{}` | Backend resource requests/limits. |
| backend.service.port | int | `8080` | Backend service port. |
| backend.service.type | string | `"ClusterIP"` | Backend service type. |
| backend.slack.channelId | string | `""` | Slack channel ID (used when backend.slack.source=values). |
| backend.slack.enabled | bool | `true` | Enable Slack notifications. |
| backend.slack.secret.channelIdKey | string | `"kube-rca-slack-channel-id"` | Secret key for Slack channel ID. |
| backend.slack.secret.existingSecret | string | `"kube-rca-slack"` | Existing Secret name for Slack credentials. |
| backend.slack.secret.tokenKey | string | `"kube-rca-slack-token"` | Secret key for Slack bot token. |
| backend.slack.token | string | `""` | Slack bot token (used when backend.slack.source=values). |
| backend.tolerations | list | `[]` | Tolerations for backend pods assignment. |
| frontend.affinity | object | `{}` | Affinity for frontend pods assignment. |
| frontend.containerPort | int | `80` | Frontend container port. |
| frontend.image.pullPolicy | string | `"IfNotPresent"` | Frontend image pull policy. |
| frontend.image.repository | string | `""` | Frontend image repository. |
| frontend.image.tag | string | `""` | Frontend image tag. |
| frontend.ingress.annotations | object | `{}` | Annotations for frontend ingress. |
| frontend.ingress.enabled | bool | `false` | Enable frontend ingress. |
| frontend.ingress.hosts | list | `[]` | Hostnames for frontend ingress. |
| frontend.ingress.ingressClassName | string | `""` | IngressClass name for frontend ingress. |
| frontend.ingress.pathType | string | `"Prefix"` | PathType for frontend ingress. |
| frontend.ingress.paths | list | `["/"]` | Paths for frontend ingress. |
| frontend.ingress.tls | list | `[]` | TLS configuration for frontend ingress. |
| frontend.nodeSelector | object | `{}` | Node labels for frontend pods assignment. |
| frontend.replicaCount | int | `1` | Number of frontend replicas. |
| frontend.resources | object | `{}` | Frontend resource requests/limits. |
| frontend.service.port | int | `80` | Frontend service port. |
| frontend.service.type | string | `"ClusterIP"` | Frontend service type. |
| frontend.tolerations | list | `[]` | Tolerations for frontend pods assignment. |
| fullnameOverride | string | `""` | Override the full name of the release. |
| nameOverride | string | `""` | Override the name of the chart. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
