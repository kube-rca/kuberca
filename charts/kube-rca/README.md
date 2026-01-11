# kube-rca

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

Deploy kube-rca backend and frontend

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| agent.affinity | object | `{}` | Affinity for agent pods assignment. |
| agent.cache.size | int | `128` | Max number of cached agents (AGENT_CACHE_SIZE). |
| agent.cache.ttlSeconds | int | `0` | Cache TTL in seconds (AGENT_CACHE_TTL_SECONDS, 0 = disable). |
| agent.containerPort | int | `8000` | Agent container port. |
| agent.gemini.modelId | string | `"gemini-3-flash-preview"` | Gemini model ID for Strands Agents. |
| agent.gemini.secret.create | bool | `false` | Create a Secret for the Gemini API key. |
| agent.gemini.secret.existingSecret | string | `"kube-rca-ai"` | Existing Secret name for the Gemini API key. |
| agent.gemini.secret.key | string | `"ai-studio-api-key"` | Secret key name for the Gemini API key. |
| agent.image.pullPolicy | string | `"IfNotPresent"` | Agent image pull policy. |
| agent.image.repository | string | `"public.ecr.aws/r5b7j2e4/kube-rca-ecr"` | Agent image repository. |
| agent.image.tag | string | `""` | Agent image tag. |
| agent.ingress.annotations | object | `{}` | Annotations for agent ingress. |
| agent.ingress.enabled | bool | `false` | Enable agent ingress. |
| agent.ingress.hosts | list | `[]` | Hostnames for agent ingress. |
| agent.ingress.ingressClassName | string | `""` | IngressClass name for agent ingress. |
| agent.ingress.pathType | string | `"Prefix"` | PathType for agent ingress. |
| agent.ingress.paths | list | `["/"]` | Paths for agent ingress. |
| agent.ingress.tls | list | `[]` | TLS configuration for agent ingress. |
| agent.k8s.apiTimeoutSeconds | int | `5` | Kubernetes API timeout in seconds. |
| agent.k8s.eventLimit | int | `20` | Kubernetes event limit. |
| agent.k8s.logTailLines | int | `50` | Kubernetes log tail lines. |
| agent.logLevel | string | `"info"` | Agent log level (LOG_LEVEL). |
| agent.nodeSelector | object | `{}` | Node labels for agent pods assignment. |
| agent.prometheus.httpTimeoutSeconds | int | `5` | Prometheus HTTP timeout in seconds. |
| agent.prometheus.labelSelector | string | `"app=kube-prometheus-stack-prometheus"` | Prometheus service label selector. |
| agent.prometheus.namespaceAllowlist | list | `[]` | Prometheus namespace allowlist (empty = all). |
| agent.prometheus.portName | string | `""` | Prometheus service port name (empty = auto). |
| agent.prometheus.scheme | string | `"http"` | Prometheus scheme. |
| agent.replicaCount | int | `1` | Number of agent replicas. |
| agent.resources | object | `{}` | Agent resource requests/limits. |
| agent.service.port | int | `8000` | Agent service port. |
| agent.service.type | string | `"ClusterIP"` | Agent service type. |
| agent.sessionDB.host | string | `""` | PostgreSQL host for Strands session persistence. |
| agent.sessionDB.name | string | `""` | PostgreSQL database name. |
| agent.sessionDB.port | int | `5432` | PostgreSQL port. |
| agent.sessionDB.secret.existingSecret | string | `"postgresql"` | Existing Secret name for session DB password. |
| agent.sessionDB.secret.key | string | `"password"` | Secret key for session DB password. |
| agent.sessionDB.user | string | `""` | PostgreSQL user. |
| agent.tolerations | list | `[]` | Tolerations for agent pods assignment. |
| agent.workers | int | `1` | Uvicorn worker count (WEB_CONCURRENCY). |
| backend.affinity | object | `{}` | Affinity for backend pods assignment. |
| backend.auth.admin.password | string | `"kube-rca"` | Admin password (default: kube-rca). |
| backend.auth.admin.username | string | `"kube-rca"` | Admin login ID (default: kube-rca). |
| backend.auth.allowSignup | bool | `false` | Allow user signup (ALLOW_SIGNUP). |
| backend.auth.cookie.domain | string | `""` | Cookie domain (AUTH_COOKIE_DOMAIN). |
| backend.auth.cookie.path | string | `"/"` | Cookie path (AUTH_COOKIE_PATH). |
| backend.auth.cookie.sameSite | string | `"Lax"` | Cookie SameSite (AUTH_COOKIE_SAMESITE). |
| backend.auth.cookie.secure | bool | `true` | Cookie secure flag (AUTH_COOKIE_SECURE). |
| backend.auth.cors.allowedOrigins | list | `[]` | Allowed origins (CORS_ALLOWED_ORIGINS), comma-separated when rendered. |
| backend.auth.enabled | bool | `true` | Enable backend auth. |
| backend.auth.jwt.accessTtl | string | `"15m"` | Access token TTL (e.g. 15m). |
| backend.auth.jwt.refreshTtl | string | `"168h"` | Refresh token TTL (e.g. 168h). |
| backend.auth.jwt.secret | string | `""` | JWT secret (auto-generated when empty and no existingSecret). |
| backend.auth.secret.existingSecret | string | `""` | Existing Secret name for auth credentials (ExternalSecret 연계 시 사용, default keys: admin-username/admin-password/kube-rca-jwt-secret). |
| backend.auth.secret.keys.adminPassword | string | `"admin-password"` | Secret key for admin password. |
| backend.auth.secret.keys.adminUsername | string | `"admin-username"` | Secret key for admin login ID. |
| backend.auth.secret.keys.jwtSecret | string | `"kube-rca-jwt-secret"` | Secret key for JWT secret. |
| backend.auth.secret.name | string | `""` | Custom Secret name for chart-managed auth Secret. |
| backend.containerPort | int | `8080` | Backend container port. |
| backend.embedding.apiKey.existingSecret | string | `"kube-rca-ai"` |  |
| backend.embedding.apiKey.key | string | `"ai-studio-api-key"` |  |
| backend.embedding.model | string | `"gemini-embedding-001"` |  |
| backend.embedding.provider | string | `"gemini"` |  |
| backend.image.pullPolicy | string | `"IfNotPresent"` | Backend image pull policy. |
| backend.image.repository | string | `"public.ecr.aws/r5b7j2e4/kube-rca-ecr"` | Backend image repository. |
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
| backend.postgresql.secret.existingSecret | string | `"postgresql"` | Existing Secret name for PostgreSQL password. |
| backend.postgresql.secret.key | string | `"password"` | Secret key for PostgreSQL password. |
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
| frontend.image.repository | string | `"public.ecr.aws/r5b7j2e4/kube-rca-ecr"` | Frontend image repository. |
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
| openapi.affinity | object | `{}` | Affinity for OpenAPI pods assignment. |
| openapi.baseUrl | string | `"/"` | Base URL for Swagger UI. |
| openapi.containerPort | int | `8080` | OpenAPI container port. |
| openapi.enabled | bool | `false` | Enable OpenAPI (Swagger UI) deployment. |
| openapi.image.pullPolicy | string | `"IfNotPresent"` | OpenAPI UI image pull policy. |
| openapi.image.repository | string | `"swaggerapi/swagger-ui"` | OpenAPI UI image repository. |
| openapi.image.tag | string | `"v5.31.0"` | OpenAPI UI image tag. |
| openapi.ingress.annotations | object | `{}` | Annotations for OpenAPI ingress. |
| openapi.ingress.enabled | bool | `false` | Enable OpenAPI ingress. |
| openapi.ingress.hosts | list | `[]` | Hostnames for OpenAPI ingress. |
| openapi.ingress.ingressClassName | string | `""` | IngressClass name for OpenAPI ingress. |
| openapi.ingress.pathType | string | `"Prefix"` | PathType for OpenAPI ingress. |
| openapi.ingress.paths | list | `["/"]` | Paths for OpenAPI ingress. |
| openapi.ingress.tls | list | `[]` | TLS configuration for OpenAPI ingress. |
| openapi.nodeSelector | object | `{}` | Node labels for OpenAPI pods assignment. |
| openapi.replicaCount | int | `1` | Number of OpenAPI replicas. |
| openapi.resources | object | `{}` | OpenAPI resource requests/limits. |
| openapi.service.port | int | `8080` | OpenAPI service port. |
| openapi.service.type | string | `"ClusterIP"` | OpenAPI service type. |
| openapi.specs.agent.name | string | `"agent"` | Display name for agent spec. |
| openapi.specs.agent.path | string | `"/openapi.json"` | Agent OpenAPI path. |
| openapi.specs.agent.service.name | string | `""` | Agent service name override (empty = chart default). |
| openapi.specs.agent.service.port | int | `8000` | Agent service port. |
| openapi.specs.backend.name | string | `"backend"` | Display name for backend spec. |
| openapi.specs.backend.path | string | `"/openapi.json"` | Backend OpenAPI path. |
| openapi.specs.backend.service.name | string | `""` | Backend service name override (empty = chart default). |
| openapi.specs.backend.service.port | int | `8080` | Backend service port. |
| openapi.tolerations | list | `[]` | Tolerations for OpenAPI pods assignment. |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
