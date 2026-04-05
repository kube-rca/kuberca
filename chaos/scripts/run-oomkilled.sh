#!/usr/bin/env bash
set -euo pipefail

show_usage() {
  cat <<EOF
Usage: $(basename "$0")

Options:
  -h, --help  Show this help message
EOF
}

for arg in "$@"; do
  case "$arg" in
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

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
exec "${SCRIPT_DIR}/run_scenario.sh" oomkilled
