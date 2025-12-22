#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

# =============================================================================
# Configuration
# =============================================================================
CLEANUP=${CLEANUP:-false}
CLEANUP_ONLY=${CLEANUP_ONLY:-false}
SKIP_ANALYZE=${SKIP_ANALYZE:-false}

# Color output (disabled if not tty)
if [ -t 1 ]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[0;33m'
  BLUE='\033[0;34m'
  NC='\033[0m'
else
  RED='' GREEN='' YELLOW='' BLUE='' NC=''
fi

# =============================================================================
# Helper functions
# =============================================================================
log_info()  { printf "${BLUE}[INFO]${NC} %s\n" "$*"; }
log_ok()    { printf "${GREEN}[OK]${NC} %s\n" "$*"; }
log_warn()  { printf "${YELLOW}[WARN]${NC} %s\n" >&2 "$*"; }
log_error() { printf "${RED}[ERROR]${NC} %s\n" >&2 "$*"; }

require_cmd() {
  local cmd=$1
  if ! command -v "$cmd" >/dev/null 2>&1; then
    log_error "command not found: $cmd"
    exit 1
  fi
}

require_env() {
  local name=$1
  if [ -z "${!name:-}" ]; then
    log_error "$name is required"
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

show_usage() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Create OOMKilled deployment and call analyze endpoint.

Environment Variables (required unless CLEANUP_ONLY=true):
  ANALYZE_URL       URL of the analyze endpoint
  THREAD_TS         Thread timestamp for Slack
  ALERT_STATUS      Alert status (e.g., firing)
  ALERT_NAME        Alert name
  ALERT_SEVERITY    Alert severity (e.g., warning, critical)
  DEPLOYMENT_NAME   Name of the deployment to create
  IMAGE             Container image to use
  OOM_COMMAND       Command that will cause OOM
  MEMORY_LIMIT      Memory limit (e.g., 64Mi)

Environment Variables (optional):
  KUBE_CONTEXT      Kubernetes context (default: current-context)
  NAMESPACE         Namespace to deploy (default: kube-rca)
  MEMORY_REQUEST    Memory request (default: same as MEMORY_LIMIT)
  SHELL_PATH        Shell path (default: /bin/sh)
  CONTAINER_NAME    Container name (default: DEPLOYMENT_NAME)
  WAIT_SECONDS      Timeout for waiting (default: 120)
  POLL_INTERVAL_SECONDS  Poll interval (default: 3)
  CLEANUP           Set to 'true' to cleanup after test
  CLEANUP_ONLY      Set to 'true' to only cleanup (skip deployment)
  SKIP_ANALYZE      Set to 'true' to skip analyze call

Examples:
  # Basic usage
  DEPLOYMENT_NAME=oom-test IMAGE=python:3.11-alpine ... ./curl-test-oomkilled.sh

  # With cleanup after test
  CLEANUP=true ... ./curl-test-oomkilled.sh

  # Cleanup only (remove existing deployment)
  CLEANUP_ONLY=true DEPLOYMENT_NAME=oom-test ./curl-test-oomkilled.sh
EOF
}

# =============================================================================
# Argument parsing
# =============================================================================
for arg in "$@"; do
  case "$arg" in
    -h|--help) show_usage; exit 0 ;;
    --cleanup) CLEANUP=true ;;
    --cleanup-only) CLEANUP_ONLY=true ;;
    --skip-analyze) SKIP_ANALYZE=true ;;
  esac
done

# =============================================================================
# Prerequisites
# =============================================================================
require_cmd kubectl
require_cmd curl

# CLEANUP_ONLY mode requires only DEPLOYMENT_NAME
if [ "$CLEANUP_ONLY" = "true" ]; then
  require_env DEPLOYMENT_NAME
else
  # Always required for OOM test
  require_env DEPLOYMENT_NAME
  require_env IMAGE
  require_env OOM_COMMAND
  require_env MEMORY_LIMIT

  # Only required if calling analyze endpoint
  if [ "$SKIP_ANALYZE" != "true" ]; then
    require_env ANALYZE_URL
    require_env THREAD_TS
    require_env ALERT_STATUS
    require_env ALERT_NAME
    require_env ALERT_SEVERITY
  fi
fi

SHELL_PATH=${SHELL_PATH:-/bin/sh}
CONTAINER_NAME=${CONTAINER_NAME:-$DEPLOYMENT_NAME}
MEMORY_REQUEST=${MEMORY_REQUEST:-${MEMORY_LIMIT:-64Mi}}
WAIT_SECONDS=${WAIT_SECONDS:-120}
POLL_INTERVAL_SECONDS=${POLL_INTERVAL_SECONDS:-3}
NAMESPACE=${NAMESPACE:-kube-rca}
RUN_TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# =============================================================================
# Kubeconfig setup
# =============================================================================
# KUBECONFIG can contain multiple paths separated by ':' (e.g., file1:file2)
# Let kubectl handle the validation instead of checking file existence directly
if [ -z "${KUBECONFIG:-}" ]; then
  if [ -n "${HOME:-}" ] && [ -f "${HOME}/.kube/config" ]; then
    export KUBECONFIG="${HOME}/.kube/config"
  fi
fi

# Verify kubectl can access cluster
if ! kubectl config current-context >/dev/null 2>&1; then
  log_error "kubectl cannot access cluster. Check KUBECONFIG or cluster connection."
  exit 1
fi

# Use provided KUBE_CONTEXT or current-context
if [ -z "${KUBE_CONTEXT:-}" ]; then
  KUBE_CONTEXT=$(kubectl config current-context 2>/dev/null || true)
  if [ -z "$KUBE_CONTEXT" ]; then
    log_error "no current kube context found. Set KUBE_CONTEXT env var."
    exit 1
  fi
  log_info "Using current context: $KUBE_CONTEXT"
else
  if ! kubectl config get-contexts "$KUBE_CONTEXT" >/dev/null 2>&1; then
    log_error "kube context not found: $KUBE_CONTEXT"
    exit 1
  fi
  log_info "Using context: $KUBE_CONTEXT"
fi

kubectl_local() {
  kubectl --context "$KUBE_CONTEXT" "$@"
}

# =============================================================================
# Cleanup function
# =============================================================================
cleanup_deployment() {
  log_info "Cleaning up deployment: $DEPLOYMENT_NAME in namespace: $NAMESPACE"
  if kubectl_local -n "$NAMESPACE" get deployment "$DEPLOYMENT_NAME" >/dev/null 2>&1; then
    kubectl_local -n "$NAMESPACE" delete deployment "$DEPLOYMENT_NAME" --wait=false
    log_ok "Deployment deleted: $DEPLOYMENT_NAME"
  else
    log_warn "Deployment not found: $DEPLOYMENT_NAME (already cleaned up?)"
  fi
}

# Handle CLEANUP_ONLY mode
if [ "$CLEANUP_ONLY" = "true" ]; then
  cleanup_deployment
  exit 0
fi

# Setup trap for cleanup on error/exit if CLEANUP=true
if [ "$CLEANUP" = "true" ]; then
  trap 'cleanup_deployment' EXIT
fi

# =============================================================================
# Validation
# =============================================================================
case "$OOM_COMMAND" in
  *$'\n'*)
    log_error "OOM_COMMAND must be a single line"
    exit 1
    ;;
esac

if ! kubectl_local get namespace "$NAMESPACE" >/dev/null 2>&1; then
  log_error "Namespace not found: $NAMESPACE"
  log_error "Please create the namespace first: kubectl create namespace $NAMESPACE"
  exit 1
fi
log_ok "Namespace exists: $NAMESPACE"

oom_command_escaped=$(printf '%s' "$OOM_COMMAND" | sed "s/'/''/g")

# =============================================================================
# Deploy OOMKilled test pod
# =============================================================================
log_info "Applying deployment: $DEPLOYMENT_NAME"
log_info "  Namespace: $NAMESPACE"
log_info "  Image: $IMAGE"
log_info "  Memory Limit: $MEMORY_LIMIT"
cat <<EOF | kubectl_local -n "$NAMESPACE" apply -f -
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
log_ok "Deployment applied"

# =============================================================================
# Wait for pod creation
# =============================================================================
log_info "Waiting for pod to be created... (timeout: ${WAIT_SECONDS}s)"
pod_deadline=$((SECONDS + WAIT_SECONDS))
pod_name=""
while [ "$SECONDS" -lt "$pod_deadline" ]; do
  pod_name=$(
    kubectl_local -n "$NAMESPACE" get pods -l "app=${DEPLOYMENT_NAME}" \
      -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true
  )
  if [ -n "$pod_name" ]; then
    break
  fi
  printf "."
  sleep "$POLL_INTERVAL_SECONDS"
done
echo ""

if [ -z "$pod_name" ]; then
  log_error "timed out waiting for pod creation: $DEPLOYMENT_NAME"
  exit 1
fi
log_ok "Pod created: $pod_name"

# =============================================================================
# Wait for OOMKilled
# =============================================================================
log_info "Waiting for OOMKilled event... (timeout: ${WAIT_SECONDS}s)"
oom_deadline=$((SECONDS + WAIT_SECONDS))
oom_reason=""
while [ "$SECONDS" -lt "$oom_deadline" ]; do
  reason=$(
    kubectl_local -n "$NAMESPACE" get pod "$pod_name" \
      -o jsonpath='{.status.containerStatuses[0].lastState.terminated.reason}' 2>/dev/null \
      || true
  )
  if [ -z "$reason" ]; then
    reason=$(
      kubectl_local -n "$NAMESPACE" get pod "$pod_name" \
        -o jsonpath='{.status.containerStatuses[0].state.terminated.reason}' 2>/dev/null \
        || true
    )
  fi
  if [ "$reason" = "OOMKilled" ]; then
    oom_reason="$reason"
    break
  fi
  printf "."
  sleep "$POLL_INTERVAL_SECONDS"
done
echo ""

if [ -z "$oom_reason" ]; then
  log_error "timed out waiting for OOMKilled: pod=$pod_name"
  kubectl_local -n "$NAMESPACE" describe pod "$pod_name" || true
  exit 1
fi
log_ok "OOMKilled detected!"

restart_count=$(
  kubectl_local -n "$NAMESPACE" get pod "$pod_name" \
    -o jsonpath='{.status.containerStatuses[0].restartCount}' 2>/dev/null || true
)
exit_code=$(
  kubectl_local -n "$NAMESPACE" get pod "$pod_name" \
    -o jsonpath='{.status.containerStatuses[0].lastState.terminated.exitCode}' 2>/dev/null \
    || true
)
if [ -z "$exit_code" ]; then
  exit_code=$(
    kubectl_local -n "$NAMESPACE" get pod "$pod_name" \
      -o jsonpath='{.status.containerStatuses[0].state.terminated.exitCode}' 2>/dev/null \
      || true
  )
fi

pod_node=$(
  kubectl_local -n "$NAMESPACE" get pod "$pod_name" \
    -o jsonpath='{.spec.nodeName}' 2>/dev/null \
    || true
)
deployment_replicas=$(
  kubectl_local -n "$NAMESPACE" get deployment "$DEPLOYMENT_NAME" \
    -o jsonpath='{.spec.replicas}' 2>/dev/null || true
)
deployment_available=$(
  kubectl_local -n "$NAMESPACE" get deployment "$DEPLOYMENT_NAME" \
    -o jsonpath='{.status.availableReplicas}' 2>/dev/null || true
)
deployment_unavailable=$(
  kubectl_local -n "$NAMESPACE" get deployment "$DEPLOYMENT_NAME" \
    -o jsonpath='{.status.unavailableReplicas}' 2>/dev/null || true
)

# =============================================================================
# Collect pod info
# =============================================================================
log_info "Collecting pod information..."
summary="OOMKilled detected"
description="Deployment ${DEPLOYMENT_NAME} pod ${pod_name} was OOMKilled"
now="$RUN_TIMESTAMP"

log_info "  Pod: $pod_name"
log_info "  Node: ${pod_node:-unknown}"
log_info "  Restart Count: ${restart_count:-0}"
log_info "  Exit Code: ${exit_code:-unknown}"

# =============================================================================
# Call analyze endpoint
# =============================================================================
if [ "$SKIP_ANALYZE" = "true" ]; then
  log_warn "Skipping analyze call (SKIP_ANALYZE=true)"
  log_ok "Test completed successfully (analyze skipped)"
  exit 0
fi

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
    "endsAt": "$now",
    "generatorURL": ""
  },
  "thread_ts": "$(json_escape "$THREAD_TS")"
}
JSON
)

log_info "Calling analyze endpoint: $ANALYZE_URL"
response=$(printf '%s' "$payload" | curl -sS -w "\n%{http_code}" -X POST "$ANALYZE_URL" \
  -H 'Content-Type: application/json' -d @-)

http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
  log_ok "Analyze call successful (HTTP $http_code)"
  echo "$body"
else
  log_error "Analyze call failed (HTTP $http_code)"
  echo "$body"
  exit 1
fi

log_ok "Test completed successfully!"
