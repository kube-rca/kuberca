# Chaos

This directory contains chaos experiments and helper scripts for injecting failures
and generating alerts.

## Prerequisites

- Chaos Mesh installed in the cluster (CRDs available).
- Istio installed in the cluster (VirtualService CRDs for 503/504 scenarios).
- `kubectl` access to the target namespace (default: `bookinfo`).

## Layout

- `scenarios/oomkilled/`: OOMKilled scenario (StressChaos targeting reviews).
- `scenarios/crashloop/`: CrashLoopBackOff scenario (target deployment).
- `scenarios/imagepull/`: ImagePullBackOff scenario (target deployment).
- `scenarios/networkdelay/`: Network delay scenario (NetworkChaos targeting ratings).
- `scenarios/500/`: 500 fault abort (details).
- `scenarios/503/`: 503 fault abort (productpage).
- `scenarios/504/`: 504 fault delay (ratings).
- `scripts/run_scenario.sh`: Shared scenario runner.
- `scripts/run-*.sh`: Scenario entrypoints.
- `scripts/agent/`: Agent analyze test helpers.

All scenarios run in `bookinfo` by default. OOMKilled uses Chaos Mesh and
targets the `reviews` workload. Network delay uses Chaos Mesh and targets the
`ratings` workload. CrashLoopBackOff and ImagePullBackOff use plain
deployments for deterministic failure states. 500/503/504 scenarios use Istio
VirtualService fault injection. 500 aborts return 500 for details. 503 aborts
return 503 for productpage. 504 delay injects 15s latency for ratings (actual
504 depends on client/gateway timeout settings).

## Quick Start

```bash
cd chaos
make oomkilled
make crashloop
make imagepull
make networkdelay
make 500
make 503
make 504
make incident   # alias for oomkilled
```

## Environment

- `NAMESPACE`: target namespace (default: `bookinfo`)
- `KUBE_CONTEXT`: kubectl context (optional)
- `WAIT_SECONDS`: wait timeout (default: `120`)
- `POLL_INTERVAL_SECONDS`: polling interval (default: `3`)

Each scenario runs until you press Enter (or Ctrl+C), then cleans up.
