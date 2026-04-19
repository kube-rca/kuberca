.PHONY: help incident oomkilled crashloop imagepull networkdelay 404 500 503 504 ratings-multi verify-clean clean-all lint

NAMESPACE ?= bookinfo
NAMESPACES ?= bookinfo kube-rca
KUBE_CONTEXT ?=
WAIT_SECONDS ?= 120
POLL_INTERVAL_SECONDS ?= 3
FAULT_PERCENTAGE ?= 40
RESTART_TIMEOUT ?= 120s
SKIP_RESTART ?= false

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

incident: oomkilled ## Alias for OOMKilled scenario

oomkilled: ## Run OOMKilled scenario (Chaos Mesh)
	@KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	./scripts/run-oomkilled.sh

crashloop: ## Run CrashLoopBackOff scenario
	@KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	./scripts/run-crashloop.sh

imagepull: ## Run ImagePullBackOff scenario
	@KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	./scripts/run-imagepull.sh

networkdelay: ## Run Network delay scenario (Chaos Mesh)
	@NAMESPACE="$(NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	./scripts/run-networkdelay.sh

404: ## Run 404 Not Found scenario (Istio fault injection)
	@NAMESPACE="$(NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	FAULT_PERCENTAGE="$(FAULT_PERCENTAGE)" \
	./scripts/run-404.sh

500: ## Run 500 Internal Server Error scenario (Istio fault injection)
	@NAMESPACE="$(NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	FAULT_PERCENTAGE="$(FAULT_PERCENTAGE)" \
	./scripts/run-500.sh

503: ## Run 503 Service Unavailable scenario (Istio fault injection)
	@NAMESPACE="$(NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	FAULT_PERCENTAGE="$(FAULT_PERCENTAGE)" \
	./scripts/run-503.sh

ratings-multi: ## Run combined ratings faults (404 from load-generator, 503 from reviews)
	@NAMESPACE="$(NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	FAULT_PERCENTAGE="$(FAULT_PERCENTAGE)" \
	./scripts/run-ratings-multi.sh

504: ## Run 504 Gateway Timeout scenario (Istio fault injection)
	@NAMESPACE="$(NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	FAULT_PERCENTAGE="$(FAULT_PERCENTAGE)" \
	./scripts/run-504.sh

verify-clean: ## Read-only check for residual chaos state (exit 1 if found)
	@NAMESPACES="$(NAMESPACES)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	./scripts/verify_clean.sh

clean-all: ## Force-remove residual chaos state and restart known workloads
	@NAMESPACES="$(NAMESPACES)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	RESTART_TIMEOUT="$(RESTART_TIMEOUT)" \
	SKIP_RESTART="$(SKIP_RESTART)" \
	./scripts/clean_all.sh

lint: ## Run shellcheck on all scripts/*.sh
	@shellcheck -x scripts/*.sh
