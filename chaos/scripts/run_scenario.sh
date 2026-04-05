#!/usr/bin/env bash
set -euo pipefail
IFS=$'\n\t'

SCENARIO=${1:-}

show_usage() {
  cat <<EOF
Usage: $(basename "$0") <scenario>

Scenarios:
  oomkilled     Apply StressChaos to trigger OOMKilled
  crashloop     Create a crashlooping deployment
  imagepull     Create a deployment with invalid image
  networkdelay  Add network delay to ratings
  404           Istio fault injection: 404 Not Found
  500           Istio fault injection: 500 Internal Server Error
  503           Istio fault injection: 503 Service Unavailable
  ratings-multi Combined ratings faults (404 from load-generator, 503 from reviews)
  504           Istio fault injection: 504 Gateway Timeout

Environment:
  NAMESPACE              Target namespace (default: bookinfo)
  KUBE_CONTEXT           Kubernetes context (optional)
  WAIT_SECONDS           Wait timeout (default: 120)
  POLL_INTERVAL_SECONDS  Poll interval (default: 3)
  FAULT_PERCENTAGE       Fault injection percentage for 4xx/5xx (default: 40)
EOF
}

if [ -z "$SCENARIO" ] || [ "$SCENARIO" = "-h" ] || [ "$SCENARIO" = "--help" ]; then
  show_usage
  exit 0
fi

KUBE_CONTEXT=${KUBE_CONTEXT:-}
WAIT_SECONDS=${WAIT_SECONDS:-120}
POLL_INTERVAL_SECONDS=${POLL_INTERVAL_SECONDS:-3}

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
SCENARIOS_DIR="${SCRIPT_DIR}/../scenarios"

log_info() { printf "[INFO] %s\n" "$*"; }
log_ok() { printf "[OK] %s\n" "$*"; }
log_warn() { printf "[WARN] %s\n" "$*" >&2; }
log_error() { printf "[ERROR] %s\n" "$*" >&2; }

is_valid_percentage() {
  local value=$1
  [[ "$value" =~ ^(100([.]0+)?|[0-9]{1,2}([.][0-9]+)?)$ ]]
}

apply_fault_percentage() {
  local src=$1
  local dest=$2
  local percent=$3

  if ! awk -v pct="$percent" '
    /percentage:/ { in_pct=1 }
    in_pct && /value:/ {
      sub(/value:[[:space:]]*[0-9]+([.][0-9]+)?/, "value: " pct)
      in_pct=0
      replaced=1
    }
    { print }
    END { if (!replaced) exit 2 }
  ' "$src" > "$dest"; then
    log_error "percentage field not found in manifest: $src"
    return 1
  fi
}

DEFAULT_NAMESPACE="bookinfo"

require_cmd() {
  local cmd=$1
  if ! command -v "$cmd" >/dev/null 2>&1; then
    log_error "command not found: $cmd"
    exit 1
  fi
}

require_file() {
  local path=$1
  if [ ! -f "$path" ]; then
    log_error "file not found: $path"
    exit 1
  fi
}

require_cmd kubectl

KUBECTL_ARGS=()
if [ -n "$KUBE_CONTEXT" ]; then
  if ! kubectl config get-contexts "$KUBE_CONTEXT" >/dev/null 2>&1; then
    log_error "kube context not found: $KUBE_CONTEXT"
    exit 1
  fi
  KUBECTL_ARGS+=(--context "$KUBE_CONTEXT")
  log_info "Using context: $KUBE_CONTEXT"
else
  if ! kubectl config current-context >/dev/null 2>&1; then
    log_error "kubectl cannot access cluster. Check KUBECONFIG."
    exit 1
  fi
  log_info "Using current context"
fi

kubectl_local() {
  if [ ${#KUBECTL_ARGS[@]} -eq 0 ]; then
    kubectl "$@"
  else
    kubectl "${KUBECTL_ARGS[@]}" "$@"
  fi
}

get_virtualservice_http_status() {
  local name=$1
  kubectl_local -n "$NAMESPACE" get virtualservice "$name" -o jsonpath='{.spec.http[0].fault.abort.httpStatus}' 2>/dev/null || true
}

ensure_non_conflicting_ratings_faults() {
  local single_vs_name="ratings-fault-abort"
  local combined_vs_name="ratings-combined-faults"
  local existing_status=""

  case "$SCENARIO" in
    404|503)
      if kubectl_local -n "$NAMESPACE" get virtualservice "$combined_vs_name" >/dev/null 2>&1; then
        log_error "combined ratings fault scenario is already active: ${combined_vs_name}"
        log_error "cleanup 'make ratings-multi' first, then run the single-fault scenario"
        exit 1
      fi

      existing_status=$(get_virtualservice_http_status "$single_vs_name")
      if [ -n "$existing_status" ]; then
        if [ "$existing_status" = "$SCENARIO" ]; then
          log_error "ratings ${SCENARIO} scenario is already active: ${single_vs_name}"
          log_error "cleanup the existing run before starting the same scenario again"
        else
          log_error "ratings fault scenario is already active: ${single_vs_name} -> HTTP ${existing_status}"
          log_error "single 404 and 503 scenarios are mutually exclusive; use 'make ratings-multi' to run them together"
        fi
        exit 1
      fi
      ;;
    ratings-multi)
      if kubectl_local -n "$NAMESPACE" get virtualservice "$combined_vs_name" >/dev/null 2>&1; then
        log_error "combined ratings fault scenario is already active: ${combined_vs_name}"
        log_error "cleanup the existing run before starting it again"
        exit 1
      fi

      existing_status=$(get_virtualservice_http_status "$single_vs_name")
      if [ -n "$existing_status" ]; then
        log_error "single ratings fault scenario is already active: ${single_vs_name} -> HTTP ${existing_status}"
        log_error "cleanup 404/503 first, then run 'make ratings-multi'"
        exit 1
      fi
      ;;
  esac
}

TARGET_MANIFEST=""
CHAOS_MANIFEST=""
LABEL_SELECTOR=""
EXPECTED_REASON=""
ALT_REASON=""
REASON_MODE="waiting"
USE_FAULT_PERCENTAGE="false"
TMP_MANIFEST=""
ORIG_CHAOS_MANIFEST=""

case "$SCENARIO" in
  oomkilled)
    # TARGET_MANIFEST="${SCENARIOS_DIR}/oomkilled/target-deployment.yaml"
    TARGET_MANIFEST=""
    CHAOS_MANIFEST="${SCENARIOS_DIR}/oomkilled/stress-chaos.yaml"
    LABEL_SELECTOR="app=details"
    EXPECTED_REASON="OOMKilled"
    REASON_MODE="oom"
    DEFAULT_NAMESPACE="bookinfo"
    ;;
  crashloop)
    TARGET_MANIFEST="${SCENARIOS_DIR}/crashloop/target-deployment.yaml"
    LABEL_SELECTOR="app=chaos-crashloop-target"
    EXPECTED_REASON="CrashLoopBackOff"
    DEFAULT_NAMESPACE="kube-rca"
    ;;
  imagepull)
    TARGET_MANIFEST="${SCENARIOS_DIR}/imagepull/target-deployment.yaml"
    LABEL_SELECTOR="app=chaos-imagepull-target"
    EXPECTED_REASON="ImagePullBackOff"
    ALT_REASON="ErrImagePull"
    DEFAULT_NAMESPACE="kube-rca"
    ;;
  networkdelay)
    CHAOS_MANIFEST="${SCENARIOS_DIR}/networkdelay/network-delay.yaml"
    LABEL_SELECTOR="app=ratings"
    ;;
  404)
    CHAOS_MANIFEST="${SCENARIOS_DIR}/404/fault-abort.yaml"
    LABEL_SELECTOR="app=ratings"
    USE_FAULT_PERCENTAGE="true"
    ;;
  500)
    CHAOS_MANIFEST="${SCENARIOS_DIR}/500/fault-abort.yaml"
    LABEL_SELECTOR="app=details"
    USE_FAULT_PERCENTAGE="true"
    ;;
  503)
    CHAOS_MANIFEST="${SCENARIOS_DIR}/503/fault-abort.yaml"
    LABEL_SELECTOR="app=ratings"
    USE_FAULT_PERCENTAGE="true"
    ;;
  ratings-multi)
    CHAOS_MANIFEST="${SCENARIOS_DIR}/ratings-multi/fault-aborts.yaml"
    LABEL_SELECTOR="app=ratings"
    USE_FAULT_PERCENTAGE="true"
    ;;
  504)
    CHAOS_MANIFEST="${SCENARIOS_DIR}/504/fault-delay.yaml"
    LABEL_SELECTOR="app=reviews"
    USE_FAULT_PERCENTAGE="true"
    ;;
  *)
    log_error "unknown scenario: $SCENARIO"
    show_usage
    exit 1
    ;;
esac

if [ -z "${NAMESPACE:-}" ]; then
  NAMESPACE="$DEFAULT_NAMESPACE"
fi

if ! kubectl_local get namespace "$NAMESPACE" >/dev/null 2>&1; then
  log_error "namespace not found: $NAMESPACE"
  exit 1
fi

ensure_non_conflicting_ratings_faults

if [ -n "$TARGET_MANIFEST" ]; then
  require_file "$TARGET_MANIFEST"
fi
if [ -n "$CHAOS_MANIFEST" ]; then
  require_file "$CHAOS_MANIFEST"
fi

if [ "$USE_FAULT_PERCENTAGE" = "true" ]; then
  FAULT_PERCENTAGE=${FAULT_PERCENTAGE:-40}
  if ! is_valid_percentage "$FAULT_PERCENTAGE"; then
    log_error "invalid FAULT_PERCENTAGE: $FAULT_PERCENTAGE (expected 0-100)"
    exit 1
  fi
  if [ -n "$CHAOS_MANIFEST" ]; then
    TMP_MANIFEST=$(mktemp)
    apply_fault_percentage "$CHAOS_MANIFEST" "$TMP_MANIFEST" "$FAULT_PERCENTAGE"
    ORIG_CHAOS_MANIFEST="$CHAOS_MANIFEST"
    CHAOS_MANIFEST="$TMP_MANIFEST"
    log_info "Fault injection percentage set to ${FAULT_PERCENTAGE}%"
  fi
fi

cleanup() {
  trap '' INT TERM
  log_info "Cleaning up..."
  local cleanup_manifest="${ORIG_CHAOS_MANIFEST:-$CHAOS_MANIFEST}"
  if [ -n "$cleanup_manifest" ]; then
    kubectl_local -n "$NAMESPACE" delete -f "$cleanup_manifest" --ignore-not-found=true || true
  fi
  if [ -n "$TARGET_MANIFEST" ]; then
    kubectl_local -n "$NAMESPACE" delete -f "$TARGET_MANIFEST" --ignore-not-found=true || true
  fi
  if [ -n "$TMP_MANIFEST" ] && [ -f "$TMP_MANIFEST" ]; then
    rm -f "$TMP_MANIFEST"
  fi
  log_ok "Cleanup complete"
  exit 0
}

trap cleanup EXIT
trap 'log_warn "Interrupted"; exit 130' INT TERM

if [ -n "$TARGET_MANIFEST" ]; then
  log_info "Applying target deployment..."
  kubectl_local -n "$NAMESPACE" apply -f "$TARGET_MANIFEST"
fi

if [ -n "$CHAOS_MANIFEST" ]; then
  log_info "Applying chaos manifest..."
  kubectl_local -n "$NAMESPACE" apply -f "$CHAOS_MANIFEST"
fi

log_info "Watch pods with:"
if [ -n "$KUBE_CONTEXT" ]; then
  log_info "  kubectl --context $KUBE_CONTEXT -n $NAMESPACE get pods -l '$LABEL_SELECTOR' -w"
else
  log_info "  kubectl -n $NAMESPACE get pods -l '$LABEL_SELECTOR' -w"
fi

wait_for_reason() {
  local label_selector=$1
  local expected_reason=$2
  local alt_reason=$3
  local mode=$4
  local deadline=$(( $(date +%s) + WAIT_SECONDS ))

  while [ "$(date +%s)" -le "$deadline" ]; do
    local pod_name=""
    pod_name=$(kubectl_local -n "$NAMESPACE" get pods -l "$label_selector" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
    if [ -z "$pod_name" ]; then
      log_info "Waiting for pod..."
      sleep "$POLL_INTERVAL_SECONDS"
      continue
    fi

    local reason=""
    if [ "$mode" = "oom" ]; then
      reason=$(kubectl_local -n "$NAMESPACE" get pod "$pod_name" -o jsonpath='{.status.containerStatuses[0].lastState.terminated.reason}' 2>/dev/null || true)
      if [ -z "$reason" ]; then
        reason=$(kubectl_local -n "$NAMESPACE" get pod "$pod_name" -o jsonpath='{.status.containerStatuses[0].state.terminated.reason}' 2>/dev/null || true)
      fi
    else
      reason=$(kubectl_local -n "$NAMESPACE" get pod "$pod_name" -o jsonpath='{.status.containerStatuses[0].state.waiting.reason}' 2>/dev/null || true)
    fi

    if [ "$reason" = "$expected_reason" ] || { [ -n "$alt_reason" ] && [ "$reason" = "$alt_reason" ]; }; then
      log_ok "Detected expected reason: ${reason} (pod: ${pod_name})"
      return 0
    fi

    log_info "Waiting for reason (${expected_reason}${alt_reason:+|$alt_reason})... current: ${reason:-none}"
    sleep "$POLL_INTERVAL_SECONDS"
  done

  log_warn "Timeout waiting for expected reason. Check pod status manually."
  return 1
}

if [ -n "$EXPECTED_REASON" ]; then
  wait_for_reason "$LABEL_SELECTOR" "$EXPECTED_REASON" "$ALT_REASON" "$REASON_MODE" || true
else
  log_info "Skipping reason check for scenario."
fi

log_info "Press Enter to cleanup and exit (Ctrl+C also cleans up)."
read -r
