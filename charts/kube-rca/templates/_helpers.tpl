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
