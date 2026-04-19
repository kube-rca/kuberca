#!/usr/bin/env bash
# verify_clean.sh — read-only residual chaos detector.
#
# Exit 0  — no residual chaos objects or fault VirtualServices detected.
# Exit 1  — one or more residues found. Names are printed to stderr.
# Exit >1 — CLI / cluster access error.
#
# Environment:
#   NAMESPACES      Space-separated namespace list (default: "bookinfo kube-rca")
#   KUBE_CONTEXT    kubectl context (optional)

set -euo pipefail
IFS=$'\n\t'

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
# shellcheck source-path=SCRIPTDIR
# shellcheck source=./lib_chaos.sh
. "${SCRIPT_DIR}/lib_chaos.sh"

NAMESPACES=${NAMESPACES:-bookinfo kube-rca}

if ! command -v kubectl >/dev/null 2>&1; then
  log_error "kubectl not found"
  exit 2
fi
if ! kubectl_local config current-context >/dev/null 2>&1; then
  log_error "kubectl cannot access cluster. Check KUBECONFIG / KUBE_CONTEXT."
  exit 2
fi

found=0
for ns in $NAMESPACES; do
  if ! kubectl_local get namespace "$ns" >/dev/null 2>&1; then
    log_info "namespace not present, skipping: ${ns}"
    continue
  fi

  chaos=$(find_residual_chaos_objects "$ns")
  faults=$(find_residual_istio_faults "$ns")

  if [ -z "${chaos}${faults}" ]; then
    log_ok "${ns}: clean"
    continue
  fi

  found=1
  log_warn "${ns}: residual state detected"
  if [ -n "$chaos" ]; then
    printf '%s\n' "$chaos" | sed "s|^|  chaos: |" >&2
  fi
  if [ -n "$faults" ]; then
    printf '%s\n' "$faults" | sed "s|^|  fault: |" >&2
  fi
done

if [ "$found" -ne 0 ]; then
  log_error "residual chaos state — run 'make clean-all' or delete listed objects manually"
  exit 1
fi
log_ok "all namespaces clean"
