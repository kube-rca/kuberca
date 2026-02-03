{{/*
Expand the name of the chart.
*/}}
{{- define "kube-rca.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a base name for resources.
Default: Use Release name so that "kube-rca-backend" style names are easy to get by installing with `--name-template kube-rca` or `helm install kube-rca ...`.
*/}}
{{- define "kube-rca.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Chart label.
*/}}
{{- define "kube-rca.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" -}}
{{- end -}}

{{/*
Common labels (base).
*/}}
{{- define "kube-rca.labels.base" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ include "kube-rca.chart" . }}
{{- end -}}

{{/*
Backend labels.
*/}}
{{- define "kube-rca.backend.labels" -}}
app.kubernetes.io/name: {{ include "kube-rca.backend.name" . }}
app.kubernetes.io/component: backend
{{ include "kube-rca.labels.base" . }}
{{- end -}}

{{/*
Backend selector labels.
*/}}
{{- define "kube-rca.backend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kube-rca.backend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Frontend labels.
*/}}
{{- define "kube-rca.frontend.labels" -}}
app.kubernetes.io/name: {{ include "kube-rca.frontend.name" . }}
app.kubernetes.io/component: frontend
{{ include "kube-rca.labels.base" . }}
{{- end -}}

{{/*
Frontend selector labels.
*/}}
{{- define "kube-rca.frontend.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kube-rca.frontend.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Agent labels.
*/}}
{{- define "kube-rca.agent.labels" -}}
app.kubernetes.io/name: {{ include "kube-rca.agent.name" . }}
app.kubernetes.io/component: agent
{{ include "kube-rca.labels.base" . }}
{{- end -}}

{{/*
Agent selector labels.
*/}}
{{- define "kube-rca.agent.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kube-rca.agent.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
OpenAPI labels.
*/}}
{{- define "kube-rca.openapi.labels" -}}
app.kubernetes.io/name: {{ include "kube-rca.openapi.name" . }}
app.kubernetes.io/component: openapi
{{ include "kube-rca.labels.base" . }}
{{- end -}}

{{/*
OpenAPI selector labels.
*/}}
{{- define "kube-rca.openapi.selectorLabels" -}}
app.kubernetes.io/name: {{ include "kube-rca.openapi.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Component names.
*/}}
{{- define "kube-rca.backend.name" -}}
{{- printf "%s-backend" (include "kube-rca.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kube-rca.frontend.name" -}}
{{- printf "%s-frontend" (include "kube-rca.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "kube-rca.agent.name" -}}
{{- printf "%s-agent" (include "kube-rca.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Gemini Secret name for agent.
*/}}
{{- define "kube-rca.agent.geminiSecretName" -}}
{{- if .Values.agent.gemini.secret.existingSecret -}}
{{- .Values.agent.gemini.secret.existingSecret -}}
{{- else -}}
{{- printf "%s-secret" (include "kube-rca.agent.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
OpenAI Secret name for agent.
*/}}
{{- define "kube-rca.agent.openaiSecretName" -}}
{{- if .Values.agent.openai.secret.existingSecret -}}
{{- .Values.agent.openai.secret.existingSecret -}}
{{- else -}}
{{- printf "%s-openai" (include "kube-rca.agent.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Anthropic Secret name for agent.
*/}}
{{- define "kube-rca.agent.anthropicSecretName" -}}
{{- if .Values.agent.anthropic.secret.existingSecret -}}
{{- .Values.agent.anthropic.secret.existingSecret -}}
{{- else -}}
{{- printf "%s-anthropic" (include "kube-rca.agent.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Slack Secret name for backend (chart-managed Secret).
*/}}
{{- define "kube-rca.backend.slackSecretName" -}}
{{- if .Values.backend.slack.secret.name -}}
{{- .Values.backend.slack.secret.name -}}
{{- else -}}
{{- printf "%s-slack" (include "kube-rca.backend.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Auth Secret name for backend (chart-managed Secret).
*/}}
{{- define "kube-rca.backend.authSecretName" -}}
{{- if .Values.backend.auth.secret.name -}}
{{- .Values.backend.auth.secret.name -}}
{{- else -}}
{{- printf "%s-auth" (include "kube-rca.backend.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
OpenAPI (Swagger UI) component name.
*/}}
{{- define "kube-rca.openapi.name" -}}
{{- printf "%s-openapi" (include "kube-rca.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Hook job labels.
*/}}
{{- define "kube-rca.hook.labels" -}}
app.kubernetes.io/component: hook
{{ include "kube-rca.labels.base" . }}
{{- end -}}

{{/*
Wait job image.
*/}}
{{- define "kube-rca.hook.waitJob.image" -}}
{{- $hooks := default (dict) .Values.hooks -}}
{{- $waitJob := default (dict) $hooks.waitJob -}}
{{- $image := default (dict) $waitJob.image -}}
{{- $repository := default "busybox" $image.repository -}}
{{- $tag := default "1.36" $image.tag -}}
{{- printf "%s:%s" $repository $tag -}}
{{- end -}}

{{/*
Wait job resources.
*/}}
{{- define "kube-rca.hook.waitJob.resources" -}}
{{- $hooks := default (dict) .Values.hooks -}}
{{- $waitJob := default (dict) $hooks.waitJob -}}
{{- $resources := default (dict) $waitJob.resources -}}
{{- if $resources }}
{{- toYaml $resources }}
{{- else }}
requests:
  cpu: 10m
  memory: 16Mi
limits:
  cpu: 50m
  memory: 32Mi
{{- end }}
{{- end -}}

{{/*
PostgreSQL service name for the embedded dependency.
*/}}
{{- define "kube-rca.postgresql.primary.fullname" -}}
{{- $pgValues := default (dict) .Values.postgresql -}}
{{- $fullname := include "common.names.dependency.fullname" (dict "chartName" "postgresql" "chartValues" $pgValues "context" $) -}}
{{- $architecture := default "standalone" $pgValues.architecture -}}
{{- $primary := default (dict) $pgValues.primary -}}
{{- $primaryName := default "primary" $primary.name -}}
{{- if eq $architecture "replication" -}}
{{- printf "%s-%s" $fullname $primaryName | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $fullname -}}
{{- end -}}
{{- end -}}

{{/*
PostgreSQL service DNS for the embedded dependency.
*/}}
{{- define "kube-rca.postgresql.service.host" -}}
{{- $pgValues := default (dict) .Values.postgresql -}}
{{- $namespace := default .Release.Namespace $pgValues.namespaceOverride -}}
{{- printf "%s.%s.svc.cluster.local" (include "kube-rca.postgresql.primary.fullname" .) $namespace -}}
{{- end -}}

{{/*
PostgreSQL connection host/port for backend and hooks.
If backend.postgresql.host is set, it is used (external DB).
Otherwise it falls back to the embedded postgresql service DNS.
*/}}
{{- define "kube-rca.postgresql.host" -}}
{{- $pgHost := default "" .Values.backend.postgresql.host -}}
{{- if $pgHost -}}
{{- $pgHost -}}
{{- else -}}
{{- include "kube-rca.postgresql.service.host" . -}}
{{- end -}}
{{- end -}}

{{- define "kube-rca.postgresql.port" -}}
{{- default 5432 .Values.backend.postgresql.port -}}
{{- end -}}

{{/*
PostgreSQL connection string for wait-for-db.
*/}}
{{- define "kube-rca.hook.postgresql.host" -}}
{{- include "kube-rca.postgresql.host" . -}}
{{- end -}}

{{- define "kube-rca.hook.postgresql.port" -}}
{{- include "kube-rca.postgresql.port" . -}}
{{- end -}}

{{/*
Backend service endpoint for wait-for-backend.
*/}}
{{- define "kube-rca.hook.backend.host" -}}
{{- include "kube-rca.backend.name" . -}}
{{- end -}}

{{- define "kube-rca.hook.backend.port" -}}
{{- default 8080 .Values.backend.service.port -}}
{{- end -}}

{{/*
Agent service endpoint for wait-for-agent.
*/}}
{{- define "kube-rca.hook.agent.host" -}}
{{- include "kube-rca.agent.name" . -}}
{{- end -}}

{{- define "kube-rca.hook.agent.port" -}}
{{- default 8000 .Values.agent.service.port -}}
{{- end -}}
