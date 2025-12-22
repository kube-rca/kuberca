SHELL := /bin/bash

UV ?= uv
VENV ?= .venv
APP ?= app.main:app
HOST ?= 0.0.0.0
PORT ?= 8082
GEMINI_MODEL_ID ?= gemini-3-flash-preview
ANALYZE_URL ?= http://localhost:8082/analyze
THREAD_TS ?= test-thread
ALERT_STATUS ?= firing
ALERT_NAME ?= TestAlert
ALERT_SEVERITY ?= warning
ALERT_NAMESPACE ?= kube-rca
ALERT_POD ?= example-pod
IMAGE ?=
DEPLOYMENT_NAME ?=
OOM_COMMAND ?=
MEMORY_LIMIT ?=
MEMORY_REQUEST ?=
SHELL_PATH ?= /bin/sh
CONTAINER_NAME ?=
WAIT_SECONDS ?= 120
POLL_INTERVAL_SECONDS ?= 3
LOCAL_ANALYZE_URL ?= http://localhost:$(PORT)/analyze
LOCAL_OOM_DEPLOYMENT ?= oomkilled-test
# Version pin for local OOM test image: python:3.11-alpine
LOCAL_OOM_IMAGE ?= python:3.11-alpine
LOCAL_OOM_COMMAND ?= python -c 'import os,time; a=os.urandom(200000000); time.sleep(1000)'
LOCAL_OOM_MEMORY_LIMIT ?= 64Mi
LOCAL_OOM_MEMORY_REQUEST ?= 64Mi

.PHONY: venv install lint format test run build help curl-analyze test-analysis test-analysis-local

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\\n  make <target>\\n\\nTargets:\\n"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-16s %s\\n", $$1, $$2}' $(MAKEFILE_LIST)

venv: ## Create virtual environment if missing
	@if [ ! -d "$(VENV)" ]; then \
		$(UV) venv; \
	else \
		echo "$(VENV) already exists. Skipping venv creation."; \
	fi

install: venv ## Install dev dependencies
	. $(VENV)/bin/activate && $(UV) pip install -e ".[dev]"

lint: install ## Run Ruff lint
	. $(VENV)/bin/activate && ruff check .

format: install ## Run Ruff formatter
	. $(VENV)/bin/activate && ruff format .

test: install ## Run tests
	. $(VENV)/bin/activate && pytest

run: install ## Run API server
	. $(VENV)/bin/activate && uvicorn $(APP) --host $(HOST) --port $(PORT)

curl-analyze: ## Call analyze endpoint with sample payload
	@printf '%s' \
'{"alert":{"status":"$(ALERT_STATUS)","labels":{"alertname":"$(ALERT_NAME)","severity":"$(ALERT_SEVERITY)","namespace":"$(ALERT_NAMESPACE)","pod":"$(ALERT_POD)"},"annotations":{"summary":"test summary","description":"test description"},"startsAt":"2024-01-01T00:00:00Z","endsAt":"0001-01-01T00:00:00Z","generatorURL":"","fingerprint":"test-fingerprint"},"thread_ts":"$(THREAD_TS)"}' \
	| curl -sS -X POST "$(ANALYZE_URL)" -H 'Content-Type: application/json' -d @-

test-analysis: ## Create OOMKilled deployment and call analyze endpoint
	@ANALYZE_URL="$(ANALYZE_URL)" \
	THREAD_TS="$(THREAD_TS)" \
	ALERT_STATUS="$(ALERT_STATUS)" \
	ALERT_NAME="$(ALERT_NAME)" \
	ALERT_SEVERITY="$(ALERT_SEVERITY)" \
	NAMESPACE="$(ALERT_NAMESPACE)" \
	DEPLOYMENT_NAME="$(DEPLOYMENT_NAME)" \
	IMAGE="$(IMAGE)" \
	OOM_COMMAND="$(OOM_COMMAND)" \
	MEMORY_LIMIT="$(MEMORY_LIMIT)" \
	MEMORY_REQUEST="$(MEMORY_REQUEST)" \
	SHELL_PATH="$(SHELL_PATH)" \
	CONTAINER_NAME="$(CONTAINER_NAME)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	bash scripts/curl-test-oomkilled.sh

test-analysis-local: install ## Run local agent w/ Gemini and test OOMKilled analyze
	@set -euo pipefail; \
	if [ -z "$${GEMINI_API_KEY:-}" ]; then \
		echo "ERROR: GEMINI_API_KEY is required"; \
		exit 1; \
	fi; \
	if [ -z "$${KUBECONFIG:-}" ]; then \
		echo "ERROR: KUBECONFIG is required"; \
		exit 1; \
	fi; \
	. $(VENV)/bin/activate; \
	GEMINI_API_KEY="$$GEMINI_API_KEY" \
	GEMINI_MODEL_ID="$(GEMINI_MODEL_ID)" \
	KUBECONFIG="$$KUBECONFIG" \
	uvicorn $(APP) --host $(HOST) --port $(PORT) & \
	pid=$$!; \
	trap 'kill $$pid 2>/dev/null || true' EXIT; \
	code=""; \
	for _ in {1..10}; do \
		code=$$(curl -sS -o /dev/null -w "%{http_code}" http://localhost:$(PORT)/healthz || true); \
		if [ "$$code" = "200" ]; then \
			break; \
		fi; \
		sleep 1; \
	done; \
	if [ "$$code" != "200" ]; then \
		echo "ERROR: local server not healthy"; \
		exit 1; \
	fi; \
	ANALYZE_URL="$(LOCAL_ANALYZE_URL)" \
	DEPLOYMENT_NAME="$(LOCAL_OOM_DEPLOYMENT)" \
	IMAGE="$(LOCAL_OOM_IMAGE)" \
	OOM_COMMAND="$(LOCAL_OOM_COMMAND)" \
	MEMORY_LIMIT="$(LOCAL_OOM_MEMORY_LIMIT)" \
	MEMORY_REQUEST="$(LOCAL_OOM_MEMORY_REQUEST)" \
	$(MAKE) --no-print-directory test-analysis

build: ## Build Docker image (IMAGE required)
	@if [ -z "$(IMAGE)" ]; then \
		echo "IMAGE is required. Example: make build IMAGE=<image>"; \
		exit 1; \
	fi
	docker build -t $(IMAGE) .
