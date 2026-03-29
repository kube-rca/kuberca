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
- `KUBE_CONTEXT`: kubectl context (optional)
- `WAIT_SECONDS`: wait timeout (default: `120`)
- `POLL_INTERVAL_SECONDS`: polling interval (default: `3`)
- `FAULT_PERCENTAGE`: fault injection percentage for 4xx/5xx (default: `40`)

Each scenario runs until you press Enter (or Ctrl+C), then cleans up.
