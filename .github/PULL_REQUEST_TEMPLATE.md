<!--
Thanks for contributing to KubeRCA!

Before opening this PR, please confirm:
- Your commits follow Conventional Commits.
- You have signed off your commits with `git commit -s` (DCO required).
- Documentation and tests are updated as needed.
-->

## Summary

<!-- One short paragraph describing what this PR does and why. -->

## Type of change

<!-- Check all that apply -->

- [ ] `feat` — New feature
- [ ] `fix` — Bug fix
- [ ] `refac` — Refactor (no behavior change)
- [ ] `perf` — Performance improvement
- [ ] `docs` — Documentation only
- [ ] `chore` — Build / tooling / dependency
- [ ] `test` — Adding or improving tests
- [ ] `BREAKING CHANGE` — Marked with `!` and a `BREAKING CHANGE:` footer

## Components touched

- [ ] backend
- [ ] agent
- [ ] frontend
- [ ] helm-chart
- [ ] chaos
- [ ] docs
- [ ] ci

## Test plan

<!--
Describe how you verified this change end-to-end. Include commands run,
chaos scenarios exercised, or manual UI flows checked.
-->

- [ ] `go test ./...` (backend)
- [ ] `uv run pytest` (agent)
- [ ] `npm run lint && npm run build` (frontend)
- [ ] `helm lint charts/kube-rca` (helm)
- [ ] Manual UI / chaos verification (describe below)

## Checklist

- [ ] Conventional Commits used in all commit messages.
- [ ] Tests added or updated where applicable.
- [ ] Documentation updated (`README.md`, `AGENTS.md`, component README, or `docs/`).
- [ ] `CHANGELOG.md` is left untouched — release-please will manage it.
- [ ] DCO sign-off present on every commit (`git commit -s`).

## Related issues

<!-- e.g. Closes #123, Refs #456 -->
