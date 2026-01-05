# Chaos

This directory contains chaos experiments and helper scripts for injecting failures
and generating alerts.

## Prerequisites

- Chaos Mesh installed in the cluster (CRDs available).
- `kubectl` access to the target namespace (default: `kube-rca`).

## Layout

- `scenarios/oomkilled/`: OOMKilled scenario (target + StressChaos).
- `scenarios/crashloop/`: CrashLoopBackOff scenario (target deployment).
- `scenarios/imagepull/`: ImagePullBackOff scenario (target deployment).
- `scripts/run_scenario.sh`: Shared scenario runner.
- `scripts/run-*.sh`: Scenario entrypoints.
- `scripts/agent/`: Agent analyze test helpers.

OOMKilled uses Chaos Mesh. CrashLoopBackOff and ImagePullBackOff use plain deployments
for deterministic failure states.

## Quick Start

```bash
cd chaos
make oomkilled
make crashloop
make imagepull
make incident   # alias for oomkilled
make incident NO_CHAOS_MESH=1   # run without Chaos Mesh
./scripts/run-oomkilled.sh --no-chaos-mesh
```

## Environment

- `NAMESPACE`: target namespace (default: `kube-rca`)
- `KUBE_CONTEXT`: kubectl context (optional)
- `WAIT_SECONDS`: wait timeout (default: `120`)
- `POLL_INTERVAL_SECONDS`: polling interval (default: `3`)
- `NO_CHAOS_MESH`: set to `1/true` to run oomkilled without Chaos Mesh

Each scenario runs until you press Enter (or Ctrl+C), then cleans up.
