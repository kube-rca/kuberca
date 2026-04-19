# Chaos

This directory contains chaos experiments and helper scripts for injecting failures
and generating alerts.

## Prerequisites

- Chaos Mesh installed in the cluster (CRDs available).
- Istio installed in the cluster (VirtualService CRDs for 404/500/503/504 scenarios).
- `kubectl` access to the target namespace (default: `bookinfo`).

## Layout

- `scenarios/oomkilled/`: OOMKilled scenario (Chaos Mesh Schedule targeting a dedicated OOM deployment).
- `scenarios/crashloop/`: CrashLoopBackOff scenario (target deployment).
- `scenarios/imagepull/`: ImagePullBackOff scenario (target deployment).
- `scenarios/networkdelay/`: Network delay scenario (NetworkChaos targeting ratings).
- `scenarios/404/`: 404 fault abort (ratings).
- `scenarios/500/`: 500 fault abort (details).
- `scenarios/503/`: 503 fault abort (ratings).
- `scenarios/ratings-multi/`: Combined ratings faults (404 for load-generator, 503 for reviews).
- `scenarios/504/`: 504 fault delay (reviews).
- `scripts/run_scenario.sh`: Shared scenario runner.
- `scripts/run-*.sh`: Scenario entrypoints.
- `scripts/agent/`: Agent analyze test helpers.

`oomkilled`, `crashloop`, and `imagepull` run in `kube-rca` namespace and use dedicated deployments.
Other scenarios run in `bookinfo` namespace by default.

OOMKilled uses Chaos Mesh Schedule to trigger memory stress every minute, causing the `chaos-oom-target` container to be OOMKilled and restarted repeatedly.
Network delay uses Chaos Mesh and targets the `ratings` workload.
CrashLoopBackOff and ImagePullBackOff use plain deployments for deterministic failure states.
404/500/503/504 scenarios use Istio VirtualService fault injection.
`ratings-multi` is the supported way to run concurrent ratings faults in one shared `VirtualService`.
It applies 404 to requests from `app=load-generator` and 503 to requests from `app=reviews`.
500 aborts return 500 for details.
503 aborts return 503 for ratings.
504 aborts return 504 for reviews.

## Quick Start

```bash
cd chaos
make oomkilled
make crashloop
make imagepull
make networkdelay
make 404
make 500
make 503
make ratings-multi
make 504
make 500 FAULT_PERCENTAGE=30
make incident   # alias for oomkilled
```

## Environment

- `NAMESPACE`: target namespace (default: `bookinfo`)
- `NAMESPACES`: space-separated namespace list for `verify-clean` / `clean-all` (default: `"bookinfo kube-rca"`)
- `KUBE_CONTEXT`: kubectl context (optional)
- `WAIT_SECONDS`: wait timeout (default: `120`)
- `POLL_INTERVAL_SECONDS`: polling interval (default: `3`)
- `FAULT_PERCENTAGE`: fault injection percentage for 4xx/5xx (default: `40`)
- `FORCE=true`: bypass the residual-chaos preflight (use only when intentionally stacking experiments)
- `FORCE_RESTART=true`: rollout-restart target workloads on cleanup even for Istio-only scenarios
- `RESTART_TIMEOUT`: `kubectl rollout status` timeout (default: `120s`)
- `SKIP_RESTART=true`: on `clean-all`, skip the workload rollout restart (teardown only)

Each scenario runs until you press Enter (or Ctrl+C), then cleans up.

## Teardown guarantees

Cleanup runs on normal exit **and** on Ctrl+C (via `trap cleanup EXIT`). It:

1. `kubectl delete` the scenario's chaos manifest and any helper Deployment.
2. (Chaos Mesh scenarios only) Wait up to 30s for `podnetworkchaos` / `podiochaos` / `podhttpchaos` finalizers to revert pod-netns rules.
3. (Chaos Mesh scenarios, or any scenario with `FORCE_RESTART=true`) `kubectl rollout restart` every Deployment matching the scenario's label selector and wait for `rollout status`.
4. Report any residual chaos objects that remain (e.g. stuck finalizer) and any pods that are still unhealthy.

Scenarios that require pod restart on teardown (they inject pod-scoped state that `kubectl delete` does not revert):

- `oomkilled` → `app=details`
- `networkdelay` → `app=ratings`

Scenarios that do **not** restart by default (pure Istio VirtualService fault injection):

- `404`, `500`, `503`, `504`, `ratings-multi`

## Pre-demo checklist

Run before every demo to catch residual state from aborted runs:

```bash
cd chaos
make verify-clean
# Exit 0 → safe to start
# Exit 1 → residue listed on stderr; run 'make clean-all' then re-verify
```

## Recovery

If a scenario was aborted without running its cleanup (crashed terminal, missed Ctrl+C, manual CRD delete outside the script), the cluster can end up with orphaned chaos objects that keep injecting failures. Recover with:

```bash
cd chaos
make clean-all                 # deletes residual CRDs + fault VS, restarts bookinfo workloads
make verify-clean              # confirm exit 0

# Scoped recovery (one namespace, no workload restart)
NAMESPACES=bookinfo SKIP_RESTART=true make clean-all
```

See `presentation/docs/rca/2026-04-19-bookinfo-productpage-5xx.md` for a concrete incident where `clean-all`-equivalent action was required.

## Lint

```bash
make lint   # shellcheck -x scripts/*.sh
```
