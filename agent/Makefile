SHELL := /bin/bash

UV ?= uv
VENV ?= .venv
APP ?= app.main:app
HOST ?= 0.0.0.0
PORT ?= 8082
ANALYZE_URL ?= http://localhost:8082/analyze
CALLBACK_URL ?= http://kube-rca-backend.kube-rca.svc:8080/callback/agent
THREAD_TS ?= test-thread
ALERT_STATUS ?= firing
ALERT_NAME ?= TestAlert
ALERT_SEVERITY ?= warning
ALERT_NAMESPACE ?= default
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

.PHONY: venv install lint format test run build help curl-analyze test-analysis

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
'{"alert":{"status":"$(ALERT_STATUS)","labels":{"alertname":"$(ALERT_NAME)","severity":"$(ALERT_SEVERITY)","namespace":"$(ALERT_NAMESPACE)","pod":"$(ALERT_POD)"},"annotations":{"summary":"test summary","description":"test description"},"startsAt":"2024-01-01T00:00:00Z","endsAt":"0001-01-01T00:00:00Z","generatorURL":"","fingerprint":"test-fingerprint"},"thread_ts":"$(THREAD_TS)","callback_url":"$(CALLBACK_URL)"}' \
	| curl -sS -X POST "$(ANALYZE_URL)" -H 'Content-Type: application/json' -d @-

test-analysis: ## Create OOMKilled deployment and call analyze endpoint
	@ANALYZE_URL="$(ANALYZE_URL)" \
	THREAD_TS="$(THREAD_TS)" \
	CALLBACK_URL="$(CALLBACK_URL)" \
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

build: ## Build Docker image (IMAGE required)
	@if [ -z "$(IMAGE)" ]; then \
		echo "IMAGE is required. Example: make build IMAGE=<image>"; \
		exit 1; \
	fi
	docker build -t $(IMAGE) .
