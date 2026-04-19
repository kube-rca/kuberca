#!/usr/bin/env bash
# clean_all.sh — force-remove all residual chaos state and flush pod netns
# on known bookinfo workloads.
#
# Use when a scenario teardown didn't run (crashed, Ctrl+C missed, demo
# aborted) and 'make verify-clean' reports residue.
#
# Environment:
#   NAMESPACES          Space-separated namespace list (default: "bookinfo kube-rca")
#   KUBE_CONTEXT        kubectl context (optional)
#   RESTART_TIMEOUT     kubectl rollout status timeout (default: 120s)
#   SKIP_RESTART=true   Skip rollout restart (cleanup only)
#
# Exits non-zero if any delete or rollout fails.

set -uo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck source-path=SCRIPTDIR
# shellcheck source=./lib_chaos.sh
. "${SCRIPT_DIR}/lib_chaos.sh"

NAMESPACES=${NAMESPACES:-bookinfo kube-rca}
RESTART_TIMEOUT=${RESTART_TIMEOUT:-120s}
SKIP_RESTART=${SKIP_RESTART:-false}

rc=0

for ns in $NAMESPACES; do
  if ! kubectl_local get namespace "$ns" >/dev/null 2>&1; then
    log_info "namespace not present, skipping: ${ns}"
    continue
  fi

  log_info "=== ${ns}: deleting residual chaos objects ==="

  # Delete top-level chaos-mesh objects first so their pod-scoped children
  # stop being re-created, then sweep the pod-scoped CRDs.
  for kind in "${CHAOS_NS_KINDS[@]}" "${CHAOS_POD_KINDS[@]}"; do
    names=$(kubectl_local -n "$ns" get "$kind" \
      -o jsonpath='{range .items[*]}{.metadata.name}{"\n"}{end}' \
      --ignore-not-found 2>/dev/null || true)
    [ -z "$names" ] && continue
    while IFS= read -r name; do
      [ -z "$name" ] && continue
      log_info "deleting ${kind}/${name}"
      if ! kubectl_local -n "$ns" delete "$kind" "$name" \
           --ignore-not-found --wait=false >/dev/null 2>&1; then
        log_warn "delete failed: ${kind}/${name}"
        rc=1
      fi
    done <<<"$names"
  done

  log_info "=== ${ns}: deleting fault VirtualServices ==="
  for vs in "${CHAOS_FAULT_VS[@]}"; do
    if kubectl_local -n "$ns" get virtualservice "$vs" >/dev/null 2>&1; then
      log_info "deleting VirtualService/${vs}"
      if ! kubectl_local -n "$ns" delete virtualservice "$vs" \
           --ignore-not-found >/dev/null 2>&1; then
        log_warn "delete failed: VirtualService/${vs}"
        rc=1
      fi
    fi
  done

  # Give chaos-mesh finalizers a moment to revert pod-netns rules before we
  # restart workloads. Residual podnetworkchaos after 30s gets reported below.
  for kind in "${CHAOS_POD_KINDS[@]}"; do
    if ! wait_for_chaos_finalizer "$ns" "$kind" 30; then
      log_warn "${kind} still present in ${ns} after 30s — proceeding anyway"
      rc=1
    fi
  done

  if [ "$SKIP_RESTART" = "true" ]; then
    log_info "SKIP_RESTART=true, not restarting workloads in ${ns}"
    continue
  fi

  if [ "$ns" = "bookinfo" ]; then
    log_info "=== ${ns}: rollout restart known workloads ==="
    for deploy in "${CHAOS_BOOKINFO_WORKLOADS[@]}"; do
      if ! kubectl_local -n "$ns" get "deployment/${deploy}" >/dev/null 2>&1; then
        continue
      fi
      log_info "rollout restart deployment/${deploy}"
      if ! kubectl_local -n "$ns" rollout restart "deployment/${deploy}" >/dev/null 2>&1; then
        log_warn "rollout restart failed: ${deploy}"
        rc=1
        continue
      fi
    done
    for deploy in "${CHAOS_BOOKINFO_WORKLOADS[@]}"; do
      if ! kubectl_local -n "$ns" get "deployment/${deploy}" >/dev/null 2>&1; then
        continue
      fi
      if ! kubectl_local -n "$ns" rollout status "deployment/${deploy}" \
           --timeout="$RESTART_TIMEOUT" >/dev/null 2>&1; then
        log_warn "rollout status timed out: ${deploy}"
        rc=1
      fi
    done
  fi
done

log_info "=== final verification ==="
if ! "${SCRIPT_DIR}/verify_clean.sh"; then
  log_error "verify_clean still reports residue — manual intervention needed"
  exit 1
fi

if [ "$rc" -ne 0 ]; then
  log_warn "clean_all finished with soft errors (see warnings above)"
  exit "$rc"
fi
log_ok "clean_all complete"
