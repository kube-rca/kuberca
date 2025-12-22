#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

require_cmd() {
  local cmd=$1
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "ERROR: command not found: $cmd" >&2
    exit 1
  fi
}

require_env() {
  local name=$1
  if [ -z "${!name:-}" ]; then
    echo "ERROR: $name is required" >&2
    exit 1
  fi
}

json_escape() {
  local s=${1-}
  s=${s//\\/\\\\}
  s=${s//\"/\\\"}
  s=${s//$'\n'/\\n}
  s=${s//$'\r'/\\r}
  s=${s//$'\t'/\\t}
  printf '%s' "$s"
}

require_cmd kubectl
require_cmd curl

require_env ANALYZE_URL
require_env THREAD_TS
require_env CALLBACK_URL
require_env ALERT_STATUS
require_env ALERT_NAME
require_env ALERT_SEVERITY
require_env NAMESPACE
require_env DEPLOYMENT_NAME
require_env IMAGE
require_env OOM_COMMAND
require_env MEMORY_LIMIT

SHELL_PATH=${SHELL_PATH:-/bin/sh}
CONTAINER_NAME=${CONTAINER_NAME:-$DEPLOYMENT_NAME}
MEMORY_REQUEST=${MEMORY_REQUEST:-$MEMORY_LIMIT}
WAIT_SECONDS=${WAIT_SECONDS:-120}
POLL_INTERVAL_SECONDS=${POLL_INTERVAL_SECONDS:-3}

case "$OOM_COMMAND" in
  *$'\n'*)
    echo "ERROR: OOM_COMMAND must be a single line" >&2
    exit 1
    ;;
esac

if ! kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
  echo "ERROR: namespace not found: $NAMESPACE" >&2
  exit 1
fi

oom_command_escaped=$(printf '%s' "$OOM_COMMAND" | sed "s/'/''/g")

echo "Applying deployment: $DEPLOYMENT_NAME"
cat <<EOF | kubectl -n "$NAMESPACE" apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${DEPLOYMENT_NAME}
  labels:
    app: ${DEPLOYMENT_NAME}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ${DEPLOYMENT_NAME}
  template:
    metadata:
      labels:
        app: ${DEPLOYMENT_NAME}
    spec:
      containers:
      - name: ${CONTAINER_NAME}
        image: ${IMAGE}
        command:
        - ${SHELL_PATH}
        - -c
        - '${oom_command_escaped}'
        resources:
          requests:
            memory: ${MEMORY_REQUEST}
          limits:
            memory: ${MEMORY_LIMIT}
EOF

pod_deadline=$((SECONDS + WAIT_SECONDS))
pod_name=""
while [ "$SECONDS" -lt "$pod_deadline" ]; do
  pod_name=$(
    kubectl -n "$NAMESPACE" get pods -l "app=${DEPLOYMENT_NAME}" \
      -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true
  )
  if [ -n "$pod_name" ]; then
    break
  fi
  sleep "$POLL_INTERVAL_SECONDS"
done

if [ -z "$pod_name" ]; then
  echo "ERROR: timed out waiting for pod creation: $DEPLOYMENT_NAME" >&2
  exit 1
fi

echo "Waiting for OOMKilled: pod=$pod_name"
oom_deadline=$((SECONDS + WAIT_SECONDS))
oom_reason=""
while [ "$SECONDS" -lt "$oom_deadline" ]; do
  reason=$(
    kubectl -n "$NAMESPACE" get pod "$pod_name" \
      -o jsonpath='{.status.containerStatuses[0].lastState.terminated.reason}' 2>/dev/null \
      || true
  )
  if [ -z "$reason" ]; then
    reason=$(
      kubectl -n "$NAMESPACE" get pod "$pod_name" \
        -o jsonpath='{.status.containerStatuses[0].state.terminated.reason}' 2>/dev/null \
        || true
    )
  fi
  if [ "$reason" = "OOMKilled" ]; then
    oom_reason="$reason"
    break
  fi
  sleep "$POLL_INTERVAL_SECONDS"
done

if [ -z "$oom_reason" ]; then
  echo "ERROR: timed out waiting for OOMKilled: pod=$pod_name" >&2
  exit 1
fi

restart_count=$(
  kubectl -n "$NAMESPACE" get pod "$pod_name" \
    -o jsonpath='{.status.containerStatuses[0].restartCount}' 2>/dev/null || true
)
exit_code=$(
  kubectl -n "$NAMESPACE" get pod "$pod_name" \
    -o jsonpath='{.status.containerStatuses[0].lastState.terminated.exitCode}' 2>/dev/null \
    || true
)
if [ -z "$exit_code" ]; then
  exit_code=$(
    kubectl -n "$NAMESPACE" get pod "$pod_name" \
      -o jsonpath='{.status.containerStatuses[0].state.terminated.exitCode}' 2>/dev/null \
      || true
  )
fi

pod_node=$(
  kubectl -n "$NAMESPACE" get pod "$pod_name" -o jsonpath='{.spec.nodeName}' 2>/dev/null \
    || true
)
deployment_replicas=$(
  kubectl -n "$NAMESPACE" get deployment "$DEPLOYMENT_NAME" \
    -o jsonpath='{.spec.replicas}' 2>/dev/null || true
)
deployment_available=$(
  kubectl -n "$NAMESPACE" get deployment "$DEPLOYMENT_NAME" \
    -o jsonpath='{.status.availableReplicas}' 2>/dev/null || true
)
deployment_unavailable=$(
  kubectl -n "$NAMESPACE" get deployment "$DEPLOYMENT_NAME" \
    -o jsonpath='{.status.unavailableReplicas}' 2>/dev/null || true
)

summary="OOMKilled detected"
description="Deployment ${DEPLOYMENT_NAME} pod ${pod_name} was OOMKilled"
now=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

payload=$(
  cat <<JSON
{
  "alert": {
    "status": "$(json_escape "$ALERT_STATUS")",
    "labels": {
      "alertname": "$(json_escape "$ALERT_NAME")",
      "severity": "$(json_escape "$ALERT_SEVERITY")",
      "namespace": "$(json_escape "$NAMESPACE")",
      "pod": "$(json_escape "$pod_name")",
      "deployment": "$(json_escape "$DEPLOYMENT_NAME")"
    },
    "annotations": {
      "summary": "$(json_escape "$summary")",
      "description": "$(json_escape "$description")",
      "deployment_replicas": "$(json_escape "$deployment_replicas")",
      "deployment_available_replicas": "$(json_escape "$deployment_available")",
      "deployment_unavailable_replicas": "$(json_escape "$deployment_unavailable")",
      "pod_node": "$(json_escape "$pod_node")",
      "container_name": "$(json_escape "$CONTAINER_NAME")",
      "container_restart_count": "$(json_escape "$restart_count")",
      "oom_exit_code": "$(json_escape "$exit_code")"
    },
    "startsAt": "$now",
    "endsAt": "0001-01-01T00:00:00Z",
    "generatorURL": ""
  },
  "thread_ts": "$(json_escape "$THREAD_TS")",
  "callback_url": "$(json_escape "$CALLBACK_URL")"
}
JSON
)

printf '%s' "$payload" | curl -sS -X POST "$ANALYZE_URL" \
  -H 'Content-Type: application/json' -d @-
