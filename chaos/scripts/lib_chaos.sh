#!/usr/bin/env bash
# shellcheck shell=bash
#
# Shared helpers for chaos scenario orchestration.
# Sourced by run_scenario.sh, verify_clean.sh, clean_all.sh.
# Does not enable `set -e` so callers keep control.

if [ -n "${LIB_CHAOS_SOURCED:-}" ]; then
  # shellcheck disable=SC2317  # second clause runs only when executed, not sourced
  return 0 2>/dev/null || exit 0
fi
LIB_CHAOS_SOURCED=1

# Pod-scoped chaos-mesh CRDs that persist iptables/tc/BPF state inside pod netns.
# Leaving these behind (even with empty spec) is the primary residue that caused
# the 2026-04-19 productpage 5xx incident.
CHAOS_POD_KINDS=(
  podnetworkchaos.chaos-mesh.org
  podiochaos.chaos-mesh.org
  podhttpchaos.chaos-mesh.org
)

# Namespace-scoped top-level chaos-mesh resources that may keep reconciling
# pod-scoped objects (Schedule in particular re-creates chaos on a cron).
CHAOS_NS_KINDS=(
  networkchaos.chaos-mesh.org
  stresschaos.chaos-mesh.org
  iochaos.chaos-mesh.org
  httpchaos.chaos-mesh.org
  schedule.chaos-mesh.org
)

# Fault-injection VirtualServices this repo is known to apply.
# Leaving any of these after a demo keeps injecting 4xx/5xx and triggers alerts.
CHAOS_FAULT_VS=(
  ratings-fault-abort
  ratings-combined-faults
  details-fault-abort
  reviews-fault-delay
)

# Workloads the chaos scenarios target. Used by clean_all.sh for a blanket
# rollout restart so any stuck pod-netns state is flushed.
# shellcheck disable=SC2034  # consumed by clean_all.sh via sourcing
CHAOS_BOOKINFO_WORKLOADS=(
  productpage-v1
  ratings-v1
  reviews-v1
  reviews-v2
  reviews-v3
  details-v1
)

# Log helpers — fall through to caller's implementation if already defined.
if ! declare -F log_info >/dev/null 2>&1; then
  log_info() { printf "[INFO] %s\n" "$*"; }
fi
if ! declare -F log_ok >/dev/null 2>&1; then
  log_ok() { printf "[OK] %s\n" "$*"; }
fi
if ! declare -F log_warn >/dev/null 2>&1; then
  log_warn() { printf "[WARN] %s\n" "$*" >&2; }
fi
if ! declare -F log_error >/dev/null 2>&1; then
  log_error() { printf "[ERROR] %s\n" "$*" >&2; }
fi

# kubectl wrapper honoring KUBE_CONTEXT. Run_scenario.sh already defines
# kubectl_local; when not sourced from there we provide our own.
if ! declare -F kubectl_local >/dev/null 2>&1; then
  kubectl_local() {
    if [ -n "${KUBE_CONTEXT:-}" ]; then
      kubectl --context "$KUBE_CONTEXT" "$@"
    else
      kubectl "$@"
    fi
  }
fi

# find_residual_chaos_objects NAMESPACE
#
# Prints one `<kind>/<name>` per line for every chaos-mesh object that exists
# in the namespace. Empty output means clean.
# Exit status is always 0 (informational).
find_residual_chaos_objects() {
  local ns=$1
  local kind
  for kind in "${CHAOS_POD_KINDS[@]}" "${CHAOS_NS_KINDS[@]}"; do
    # --ignore-not-found prevents failure when a CRD is not installed.
    kubectl_local -n "$ns" get "$kind" \
      -o jsonpath='{range .items[*]}{.kind}/{.metadata.name}{"\n"}{end}' \
      --ignore-not-found 2>/dev/null || true
  done
}

# find_residual_istio_faults NAMESPACE
#
# Prints one `VirtualService/<name>` per line for each known fault VS that
# still exists. Empty output means clean.
find_residual_istio_faults() {
  local ns=$1
  local vs
  for vs in "${CHAOS_FAULT_VS[@]}"; do
    if kubectl_local -n "$ns" get virtualservice "$vs" >/dev/null 2>&1; then
      printf "VirtualService/%s\n" "$vs"
    fi
  done
}

# wait_for_chaos_finalizer NAMESPACE KIND TIMEOUT_SECONDS
#
# Chaos Mesh removes pod-scoped objects (podnetworkchaos etc.) asynchronously
# after the parent CRD is deleted. Poll until no object of KIND remains in the
# namespace, or TIMEOUT_SECONDS elapses. Returns 0 on clean, 1 on timeout.
wait_for_chaos_finalizer() {
  local ns=$1
  local kind=$2
  local timeout=${3:-30}
  local deadline=$(( $(date +%s) + timeout ))

  while [ "$(date +%s)" -le "$deadline" ]; do
    local remaining
    remaining=$(kubectl_local -n "$ns" get "$kind" \
      -o jsonpath='{range .items[*]}{.metadata.name}{" "}{end}' \
      --ignore-not-found 2>/dev/null || true)
    if [ -z "${remaining// /}" ]; then
      return 0
    fi
    sleep 2
  done
  return 1
}

# find_deployments_by_label NAMESPACE LABEL_SELECTOR
#
# Returns deployment names (one per line) that match the given label selector.
# If the selector is in the form "app=foo", this translates Pod-level labels
# to Deployments by using `kubectl get deployment -l`.
find_deployments_by_label() {
  local ns=$1
  local selector=$2
  if [ -z "$selector" ]; then
    return 0
  fi
  kubectl_local -n "$ns" get deployment -l "$selector" \
    -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' \
    --ignore-not-found 2>/dev/null || true
}

# restart_workload NAMESPACE LABEL_SELECTOR TIMEOUT
#
# Roll-restarts every Deployment matching LABEL_SELECTOR and waits for
# `rollout status` with the given TIMEOUT (kubectl duration, e.g. 120s).
# Returns 0 when all rollouts complete, 1 if any restart or status wait fails.
restart_workload() {
  local ns=$1
  local selector=$2
  local timeout=${3:-120s}

  if [ -z "$selector" ]; then
    log_warn "restart_workload: empty label selector, skipping"
    return 0
  fi

  local deployments
  deployments=$(find_deployments_by_label "$ns" "$selector")
  if [ -z "$deployments" ]; then
    log_warn "restart_workload: no deployments matched '${selector}' in ${ns}"
    return 0
  fi

  local rc=0
  local deploy
  while IFS= read -r deploy; do
    [ -z "$deploy" ] && continue
    log_info "rollout restart deployment/${deploy} (namespace ${ns})"
    if ! kubectl_local -n "$ns" rollout restart "deployment/${deploy}"; then
      log_error "rollout restart failed: ${deploy}"
      rc=1
      continue
    fi
    if ! kubectl_local -n "$ns" rollout status "deployment/${deploy}" --timeout="$timeout"; then
      log_error "rollout status timed out: ${deploy}"
      rc=1
    fi
  done <<<"$deployments"
  return $rc
}

# verify_workload_healthy NAMESPACE LABEL_SELECTOR
#
# Returns 0 when every pod matching the selector reports `Ready=True` with no
# container in a waiting state. Returns 1 otherwise. Prints a short reason
# on failure.
verify_workload_healthy() {
  local ns=$1
  local selector=$2

  if [ -z "$selector" ]; then
    return 0
  fi

  local pods
  pods=$(kubectl_local -n "$ns" get pod -l "$selector" \
    -o jsonpath='{range .items[*]}{.metadata.name}{"|"}{.status.conditions[?(@.type=="Ready")].status}{"|"}{.status.containerStatuses[*].state.waiting.reason}{"\n"}{end}' \
    --ignore-not-found 2>/dev/null || true)

  if [ -z "$pods" ]; then
    log_warn "verify_workload_healthy: no pods matched '${selector}' in ${ns}"
    return 0
  fi

  local bad=0
  local line
  while IFS= read -r line; do
    [ -z "$line" ] && continue
    local name ready waiting
    name=${line%%|*}
    local rest=${line#*|}
    ready=${rest%%|*}
    waiting=${rest#*|}
    if [ "$ready" != "True" ] || [ -n "${waiting// /}" ]; then
      log_warn "unhealthy pod: ${name} (ready=${ready:-?}, waiting=${waiting:-none})"
      bad=1
    fi
  done <<<"$pods"
  return $bad
}
