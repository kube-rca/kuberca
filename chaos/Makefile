.PHONY: help incident oomkilled crashloop imagepull

NAMESPACE ?= kube-rca
KUBE_CONTEXT ?=
WAIT_SECONDS ?= 120
POLL_INTERVAL_SECONDS ?= 3

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

incident: oomkilled ## Alias for OOMKilled scenario

oomkilled: ## Run OOMKilled scenario (Chaos Mesh)
	@NAMESPACE="$(NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	./scripts/run-oomkilled.sh

crashloop: ## Run CrashLoopBackOff scenario
	@NAMESPACE="$(NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	./scripts/run-crashloop.sh

imagepull: ## Run ImagePullBackOff scenario
	@NAMESPACE="$(NAMESPACE)" \
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	WAIT_SECONDS="$(WAIT_SECONDS)" \
	POLL_INTERVAL_SECONDS="$(POLL_INTERVAL_SECONDS)" \
	./scripts/run-imagepull.sh
