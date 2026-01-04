SHELL := /bin/bash

UV ?= uv
VENV ?= .venv
APP ?= app.main:app
HOST ?= 0.0.0.0
PORT ?= 8000
GEMINI_MODEL_ID ?= gemini-3-flash-preview
ANALYZE_URL ?= http://localhost:8000/analyze
CHAOS_SCRIPTS_DIR ?= ../chaos/scripts/agent
THREAD_TS ?= test-thread
ALERT_STATUS ?= firing
ALERT_NAME ?= TestAlert
ALERT_SEVERITY ?= warning
ALERT_NAMESPACE ?= kube-rca
ALERT_POD ?= oomkilled-test-bf86845c6-4fr6w
IMAGE ?=
DEPLOYMENT_NAME ?=
OOM_COMMAND ?=
MEMORY_LIMIT ?=
MEMORY_REQUEST ?=
SHELL_PATH ?= /bin/sh
CONTAINER_NAME ?=
WAIT_SECONDS ?= 120
POLL_INTERVAL_SECONDS ?= 3
# Kubernetes context (empty = use current-context)
KUBE_CONTEXT ?=
# Cleanup after test (true/false)
CLEANUP ?= false

# Local OOM test defaults
LOCAL_ANALYZE_URL ?= http://localhost:$(PORT)/analyze
LOCAL_OOM_DEPLOYMENT ?= oomkilled-test
LOCAL_OOM_IMAGE ?= python:3.11-alpine
LOCAL_OOM_COMMAND ?= python -c 'import os,time; a=os.urandom(200000000); time.sleep(1000)'
LOCAL_OOM_MEMORY_LIMIT ?= 64Mi
LOCAL_OOM_MEMORY_REQUEST ?= 64Mi
LOCAL_OOM_NAMESPACE ?= kube-rca

# Local CrashLoopBackOff test defaults
LOCAL_CRASH_DEPLOYMENT ?= crashloop-test
LOCAL_CRASH_IMAGE ?= python:3.11-alpine
LOCAL_CRASH_COMMAND ?= /bin/sh -c 'exit 1'
LOCAL_CRASH_NAMESPACE ?= kube-rca

# Local ImagePullBackOff test defaults
LOCAL_IMAGEPULL_DEPLOYMENT ?= imagepull-test
LOCAL_IMAGEPULL_IMAGE ?= nginx:non-existent-tag-12345
LOCAL_IMAGEPULL_NAMESPACE ?= kube-rca

# Curl pod settings for in-cluster testing
CURL_POD_IMAGE ?= curlimages/curl:8.11.1
AGENT_SERVICE_NAME ?= kube-rca-agent
AGENT_SERVICE_NAMESPACE ?= kube-rca
AGENT_SERVICE_PORT ?= 8000

.PHONY: venv install lint format test run build help curl-analyze curl-analyze-local test-analysis test-analysis-local test-oom-only cleanup-oom test-crash-only test-analysis-crash cleanup-crash test-imagepull-only test-analysis-imagepull cleanup-imagepull

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

curl-analyze: ## Call analyze endpoint for OOMKilled test pod via curlpod
	@set -euo pipefail; \
	POD_NAME=$$(kubectl get pods -n "$(LOCAL_OOM_NAMESPACE)" -l "app=$(LOCAL_OOM_DEPLOYMENT)" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true); \
	if [ -z "$$POD_NAME" ]; then \
		echo "[ERROR] OOMKilled test pod not found. Run 'make test-oom-only' first."; \
		exit 1; \
	fi; \
	echo "[INFO] Found OOMKilled pod: $$POD_NAME"; \
	echo "[INFO] Calling analyze endpoint via curlpod..."; \
	kubectl run curlpod-$$$$ --rm -i --restart=Never \
		--namespace="$(AGENT_SERVICE_NAMESPACE)" \
		--image="$(CURL_POD_IMAGE)" \
		-- curl -sS -X POST \
		"http://$(AGENT_SERVICE_NAME).$(AGENT_SERVICE_NAMESPACE).svc.cluster.local:$(AGENT_SERVICE_PORT)/analyze" \
		-H 'Content-Type: application/json' \
		-d "{\"alert\":{\"status\":\"$(ALERT_STATUS)\",\"labels\":{\"alertname\":\"OOMKilled\",\"severity\":\"$(ALERT_SEVERITY)\",\"namespace\":\"$(LOCAL_OOM_NAMESPACE)\",\"pod\":\"$$POD_NAME\",\"deployment\":\"$(LOCAL_OOM_DEPLOYMENT)\"},\"annotations\":{\"summary\":\"OOMKilled detected\",\"description\":\"Test OOMKilled pod\"},\"startsAt\":\"2024-01-01T00:00:00Z\",\"endsAt\":\"0001-01-01T00:00:00Z\",\"generatorURL\":\"\",\"fingerprint\":\"test-fingerprint\"},\"thread_ts\":\"$(THREAD_TS)\"}"

curl-analyze-local: ## Call analyze endpoint on localhost (for local dev)
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
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	CLEANUP="$(CLEANUP)" \
	bash $(CHAOS_SCRIPTS_DIR)/curl-test-oomkilled.sh

test-oom-only: ## Create OOMKilled pod without calling analyze (for testing)
	@ANALYZE_URL="http://dummy" \
	THREAD_TS="test" \
	ALERT_STATUS="firing" \
	ALERT_NAME="OOMKilled" \
	ALERT_SEVERITY="warning" \
	NAMESPACE="$(LOCAL_OOM_NAMESPACE)" \
	DEPLOYMENT_NAME="$(LOCAL_OOM_DEPLOYMENT)" \
	IMAGE="$(LOCAL_OOM_IMAGE)" \
	OOM_COMMAND="$(LOCAL_OOM_COMMAND)" \
	MEMORY_LIMIT="$(LOCAL_OOM_MEMORY_LIMIT)" \
	MEMORY_REQUEST="$(LOCAL_OOM_MEMORY_REQUEST)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	CLEANUP="$(CLEANUP)" \
	SKIP_ANALYZE=true \
	bash $(CHAOS_SCRIPTS_DIR)/curl-test-oomkilled.sh

cleanup-oom: ## Cleanup OOMKilled test deployment
	@DEPLOYMENT_NAME="$(LOCAL_OOM_DEPLOYMENT)" \
	NAMESPACE="$(LOCAL_OOM_NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	CLEANUP_ONLY=true \
	bash $(CHAOS_SCRIPTS_DIR)/curl-test-oomkilled.sh

test-crash-only: ## Create CrashLoopBackOff pod without calling analyze (for testing)
	@ANALYZE_URL="http://dummy" \
	THREAD_TS="test" \
	ALERT_STATUS="firing" \
	ALERT_NAME="KubePodCrashLooping" \
	ALERT_SEVERITY="warning" \
	NAMESPACE="$(LOCAL_CRASH_NAMESPACE)" \
	DEPLOYMENT_NAME="$(LOCAL_CRASH_DEPLOYMENT)" \
	IMAGE="$(LOCAL_CRASH_IMAGE)" \
	CRASH_COMMAND="$(LOCAL_CRASH_COMMAND)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	CLEANUP="$(CLEANUP)" \
	SKIP_ANALYZE=true \
	bash $(CHAOS_SCRIPTS_DIR)/curl-test-crashloop.sh

test-analysis-crash: ## Create CrashLoopBackOff deployment and call analyze endpoint
	@ANALYZE_URL="$(ANALYZE_URL)" \
	THREAD_TS="$(THREAD_TS)" \
	ALERT_STATUS="$(ALERT_STATUS)" \
	ALERT_NAME="KubePodCrashLooping" \
	ALERT_SEVERITY="$(ALERT_SEVERITY)" \
	NAMESPACE="$(ALERT_NAMESPACE)" \
	DEPLOYMENT_NAME="$(LOCAL_CRASH_DEPLOYMENT)" \
	IMAGE="$(LOCAL_CRASH_IMAGE)" \
	CRASH_COMMAND="$(LOCAL_CRASH_COMMAND)" \
	SHELL_PATH="$(SHELL_PATH)" \
	CONTAINER_NAME="$(CONTAINER_NAME)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	CLEANUP="$(CLEANUP)" \
	bash $(CHAOS_SCRIPTS_DIR)/curl-test-crashloop.sh

cleanup-crash: ## Cleanup CrashLoopBackOff test deployment
	@DEPLOYMENT_NAME="$(LOCAL_CRASH_DEPLOYMENT)" \
	NAMESPACE="$(LOCAL_CRASH_NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	CLEANUP_ONLY=true \
	bash $(CHAOS_SCRIPTS_DIR)/curl-test-crashloop.sh

test-imagepull-only: ## Create ImagePullBackOff pod without calling analyze (for testing)
	@ANALYZE_URL="http://dummy" \
	THREAD_TS="test" \
	ALERT_STATUS="firing" \
	ALERT_NAME="KubePodImagePullBackOff" \
	ALERT_SEVERITY="warning" \
	NAMESPACE="$(LOCAL_IMAGEPULL_NAMESPACE)" \
	DEPLOYMENT_NAME="$(LOCAL_IMAGEPULL_DEPLOYMENT)" \
	IMAGE="$(LOCAL_IMAGEPULL_IMAGE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	CLEANUP="$(CLEANUP)" \
	SKIP_ANALYZE=true \
	bash $(CHAOS_SCRIPTS_DIR)/curl-test-imagepull.sh

test-analysis-imagepull: ## Create ImagePullBackOff deployment and call analyze endpoint
	@ANALYZE_URL="$(ANALYZE_URL)" \
	THREAD_TS="$(THREAD_TS)" \
	ALERT_STATUS="$(ALERT_STATUS)" \
	ALERT_NAME="KubePodImagePullBackOff" \
	ALERT_SEVERITY="$(ALERT_SEVERITY)" \
	NAMESPACE="$(ALERT_NAMESPACE)" \
	DEPLOYMENT_NAME="$(LOCAL_IMAGEPULL_DEPLOYMENT)" \
	IMAGE="$(LOCAL_IMAGEPULL_IMAGE)" \
	SHELL_PATH="$(SHELL_PATH)" \
	CONTAINER_NAME="$(CONTAINER_NAME)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	CLEANUP="$(CLEANUP)" \
	bash $(CHAOS_SCRIPTS_DIR)/curl-test-imagepull.sh

cleanup-imagepull: ## Cleanup ImagePullBackOff test deployment
	@DEPLOYMENT_NAME="$(LOCAL_IMAGEPULL_DEPLOYMENT)" \
	NAMESPACE="$(LOCAL_IMAGEPULL_NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	CLEANUP_ONLY=true \
	bash $(CHAOS_SCRIPTS_DIR)/curl-test-imagepull.sh

test-analysis-local: install ## Run local agent w/ Gemini and test OOMKilled analyze
	@set -euo pipefail; \
	if [ -z "$${GEMINI_API_KEY:-}" ]; then \
		echo "[ERROR] GEMINI_API_KEY is required"; \
		exit 1; \
	fi; \
	if [ -z "$${KUBECONFIG:-}" ]; then \
		echo "[ERROR] KUBECONFIG is required"; \
		exit 1; \
	fi; \
	echo "[INFO] Starting local agent server..."; \
	. $(VENV)/bin/activate; \
	GEMINI_API_KEY="$$GEMINI_API_KEY" \
	GEMINI_MODEL_ID="$(GEMINI_MODEL_ID)" \
	KUBECONFIG="$$KUBECONFIG" \
	uvicorn $(APP) --host $(HOST) --port $(PORT) & \
	pid=$$!; \
	trap 'echo "[INFO] Stopping server..."; kill $$pid 2>/dev/null || true' EXIT; \
	echo "[INFO] Waiting for server health check..."; \
	code=""; \
	for i in $$(seq 1 15); do \
		code=$$(curl -sS -o /dev/null -w "%{http_code}" http://localhost:$(PORT)/healthz 2>/dev/null || true); \
		if [ "$$code" = "200" ]; then \
			echo "[OK] Server is healthy"; \
			break; \
		fi; \
		printf "."; \
		sleep 1; \
	done; \
	echo ""; \
	if [ "$$code" != "200" ]; then \
		echo "[ERROR] Local server not healthy after 15s"; \
		exit 1; \
	fi; \
	echo "[INFO] Running OOMKilled test..."; \
	ANALYZE_URL="$(LOCAL_ANALYZE_URL)" \
	DEPLOYMENT_NAME="$(LOCAL_OOM_DEPLOYMENT)" \
	IMAGE="$(LOCAL_OOM_IMAGE)" \
	OOM_COMMAND="$(LOCAL_OOM_COMMAND)" \
	MEMORY_LIMIT="$(LOCAL_OOM_MEMORY_LIMIT)" \
	MEMORY_REQUEST="$(LOCAL_OOM_MEMORY_REQUEST)" \
	NAMESPACE="$(LOCAL_OOM_NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	CLEANUP="$(CLEANUP)" \
	$(MAKE) --no-print-directory test-analysis

build: ## Build Docker image (IMAGE required)
	@if [ -z "$(IMAGE)" ]; then \
		echo "IMAGE is required. Example: make build IMAGE=<image>"; \
		exit 1; \
	fi
	docker build -t $(IMAGE) .
