# Contributing to KubeRCA

Thanks for your interest in contributing. KubeRCA is an open-source Kubernetes RCA
tool maintained by the [`kube-rca`](https://github.com/kube-rca) GitHub organization.
Contributions of all kinds are welcome — bug reports, feature ideas, documentation
fixes, design reviews, and code.

This document explains the practical mechanics: how to file issues, how to set up each
component locally, what we expect in commits and pull requests, and the licensing /
sign-off model.

## Code of Conduct

By participating in this project, you agree to abide by the
[Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). Please read it before
opening issues, PRs, or discussion threads.

## Ways to Contribute

### Reporting Bugs

Use the **Bug report** issue template at
<https://github.com/kube-rca/kuberca/issues/new/choose>. Please include:

- KubeRCA component and version (`backend-X.Y.Z`, `agent-X.Y.Z`, `frontend-X.Y.Z`,
  `helm-chart-X.Y.Z`).
- Kubernetes version and deployment method (Helm CLI or ArgoCD).
- AI provider in use (`gemini`, `openai`, or `anthropic`).
- Reproduction steps, expected behavior, actual behavior, and any relevant logs.

### Proposing Features

Use the **Feature request** issue template. Describe the operational problem first,
then a proposed solution and any alternatives you considered. Larger proposals
(cross-component, breaking changes, new dependencies) should be discussed in
[GitHub Discussions](https://github.com/kube-rca/kuberca/discussions) first.

### Asking Questions

Open a thread in [GitHub Discussions](https://github.com/kube-rca/kuberca/discussions)
rather than the issue tracker. The **Question** issue template will redirect you there.

### Reporting Security Vulnerabilities

**Do not open a public issue.** Follow [SECURITY.md](SECURITY.md) and use GitHub
Private Vulnerability Reporting at
<https://github.com/kube-rca/kuberca/security/advisories/new>.

## Development Setup

KubeRCA is a monorepo with four primary components. Run commands from each
component's directory.

### Backend (Go 1.24)

```bash
cd backend
go fmt ./...
go vet ./...
go test ./...
```

Build the binary locally:

```bash
go build -o bin/kuberca-backend .
```

### Agent (Python 3.10+, uv)

```bash
cd agent
make install   # uv-managed virtualenv + dependency install
make lint      # ruff check
make test      # pytest
```

We use [uv](https://docs.astral.sh/uv/) — please do not introduce `pip` or `pip3`
calls. For one-off scripts use `uv run --with <package>`.

### Frontend (React 18 + TypeScript, Vite)

```bash
cd frontend
npm install
npm run lint
npm run build
```

For interactive development:

```bash
npm run dev
```

### Helm Chart

The chart depends on Bitnami `postgresql`, so the dependency build step is required:

```bash
helm repo add bitnami https://charts.bitnami.com/bitnami --force-update
helm dependency build charts/kube-rca
helm lint charts/kube-rca
```

### End-to-End Verification

For changes that touch the runtime path (Alertmanager → backend → agent → Slack/UI),
exercise a chaos scenario locally where possible:

```bash
cd chaos
make oomkilled    # OOMKilled scenario
make crashloop    # CrashLoopBackOff scenario
make 500          # 500 fault injection
```

## Commit Conventions

We use [Conventional Commits](https://www.conventionalcommits.org/). The first line
of each commit must follow:

```
<type>(<scope>): <short summary>
```

| Type     | When to use                                                     |
| -------- | --------------------------------------------------------------- |
| `feat`   | A new feature visible to users or operators.                    |
| `fix`    | A bug fix.                                                      |
| `refac`  | A refactor that does not change behavior.                       |
| `docs`   | Documentation only.                                             |
| `chore`  | Build, tooling, dependency bumps, repo housekeeping.            |
| `test`   | Adding or improving tests, no behavior change.                  |
| `perf`   | A performance improvement.                                      |

Common scopes: `backend`, `agent`, `frontend`, `helm`, `chaos`, `docs`, `ci`,
`release`, `governance`.

For breaking changes, append `!` to the type/scope and include a `BREAKING CHANGE:`
footer:

```
feat(backend)!: rename /webhook/alertmanager to /v1/webhook/alertmanager

BREAKING CHANGE: Alertmanager receivers must update the webhook URL.
```

If the change relates to a tracked ticket, append the ticket key in parentheses:

```
feat(agent): tune Strands tool budget (KAN-123)
```

Releases are automated by [release-please](https://github.com/googleapis/release-please)
based on these commit messages — please do not edit `CHANGELOG.md` by hand.

## Developer Certificate of Origin (DCO)

All commits must be signed off under the
[Developer Certificate of Origin 1.1](https://developercertificate.org/). This is a
lightweight, standard mechanism (used by the Linux kernel and the CNCF) confirming
that you have the right to contribute the code under the project's Apache 2.0 license.

Sign off your commits with `-s`:

```bash
git commit -s -m "feat(agent): add Tempo trace enricher"
```

This appends a trailer to the commit message:

```
Signed-off-by: Your Name <you@example.com>
```

PRs without DCO sign-off cannot be merged.

## Pull Request Workflow

1. Fork the repo and create a topic branch — `<type>/<short-description>` (kebab-case),
   e.g. `feat/tempo-enricher` or `fix/slack-thread-race`.
2. Make your changes and run the relevant local checks (lint, test, build).
3. Commit using Conventional Commits with `git commit -s`.
4. Open a PR against `main`. The PR template
   ([`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md)) will
   guide you through the checklist:
   - Conventional Commits used.
   - Tests added or updated where applicable.
   - Documentation updated (`README.md`, `AGENTS.md`, component README).
   - CHANGELOG entry will be auto-generated by release-please — no manual edits.
   - DCO sign-off present (`git commit -s`).
5. CI must be green. A maintainer will review per the
   [governance model](GOVERNANCE.md). Lazy-consensus applies; breaking changes
   require explicit maintainer approval.

## Documentation Expectations

If your change affects observable behavior, update the relevant docs in the same PR:

- Top-level `README.md` for user-facing capability changes.
- Component `README.md` files (`backend/README.md`, `agent/README.md`,
  `frontend/README.md`).
- `AGENTS.md` files when commands or directory layout change.
- Helm `values.yaml` documentation comments.

Diagrams under `docs/diagrams/` and architecture notes under `docs/ARCHITECTURE.md`
should also be updated when the runtime flow changes.

## Licensing

KubeRCA is licensed under the [Apache License 2.0](LICENSE). By submitting a pull
request, you agree that your contribution is licensed under the same terms. The DCO
sign-off (above) is the mechanism by which we record this assertion — we do not
require a separate CLA.

## Getting Help

- General questions: [GitHub Discussions](https://github.com/kube-rca/kuberca/discussions)
- Bugs: [GitHub Issues](https://github.com/kube-rca/kuberca/issues)
- Security: [SECURITY.md](SECURITY.md)

Welcome aboard, and thanks for helping make Kubernetes incident response a little
less painful.
