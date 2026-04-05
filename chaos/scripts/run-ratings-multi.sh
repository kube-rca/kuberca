#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
exec "${SCRIPT_DIR}/run_scenario.sh" ratings-multi "$@"
