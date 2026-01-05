#!/usr/bin/env bash
set -euo pipefail

show_usage() {
  cat <<EOF
Usage: $(basename "$0") [--no-chaos-mesh]

Options:
  --no-chaos-mesh, -no-chaos-mesh  Run without Chaos Mesh (self-oom target)
EOF
}

NO_CHAOS_MESH="${NO_CHAOS_MESH:-}"
for arg in "$@"; do
  case "$arg" in
    --no-chaos-mesh|-no-chaos-mesh)
      NO_CHAOS_MESH=1
      ;;
    -h|--help)
      show_usage
      exit 0
      ;;
    *)
      printf "[ERROR] unknown argument: %s\n" "$arg" >&2
      show_usage
      exit 1
      ;;
  esac
done

export NO_CHAOS_MESH
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
exec "${SCRIPT_DIR}/run_scenario.sh" oomkilled
