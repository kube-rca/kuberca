# KubeRCA Governance

KubeRCA is an open-source project under the Apache License 2.0. This document
describes how the project is organized, how decisions are made, and how
contributors can advance into formal project roles.

## Roles

### Contributors

Anyone who interacts with the project — opening an issue, asking a question,
proposing a change, reviewing a PR, or improving the docs. There is no formal
on-ramp; a single accepted contribution makes you a Contributor.

### Reviewers

Contributors who have demonstrated familiarity with one or more components and
provide consistent, high-quality reviews. Reviewers can `LGTM` PRs in their area
of expertise but cannot merge alone.

A Maintainer may nominate a Contributor to become a Reviewer; lazy consensus of
active Maintainers (objection within 7 days) confirms the nomination.

### Maintainers

Maintainers are responsible for the long-term health of the project. They:

- Review and merge pull requests across components.
- Triage issues and apply labels.
- Cut releases via the `release-please` automation.
- Vote on changes to governance, license, or roadmap.
- Make final decisions on conflicts that lazy consensus cannot resolve.

The current Maintainers are listed in [MAINTAINERS.md](MAINTAINERS.md).

## Decision Making

We default to **lazy consensus**:

- A change is considered approved if no Maintainer objects within a reasonable
  review window (typically 3 working days for non-trivial PRs, 7 days for
  governance / roadmap / breaking changes).
- Anyone may raise an objection. An objection from a Maintainer blocks the
  change until resolved through discussion.

For changes that cannot reach consensus, Maintainers vote. A simple majority of
active Maintainers carries the decision. Ties are broken by the project lead.

### Breaking Changes

Breaking changes (Conventional Commits `!` or `BREAKING CHANGE:` footer) require
**explicit approval** from at least one Maintainer beyond the author. They are
called out in release notes by `release-please`.

## Adding and Removing Maintainers

### Adding

A Reviewer may be nominated for Maintainer by an existing Maintainer. The
nomination requires support from **2/3 of the current active Maintainers**.
Active is defined as having merged or reviewed at least one PR in the past
90 days.

### Stepping Down

Maintainers may step down at any time by opening a PR removing themselves from
[MAINTAINERS.md](MAINTAINERS.md).

### Removing

A Maintainer who is inactive for 180 days (no merges, reviews, or
participation) may be moved to "Emeritus" status by 2/3 vote of active
Maintainers. They retain credit and are welcome to return.

## Releases

Releases are managed by [release-please](https://github.com/googleapis/release-please)
based on Conventional Commits. Each component is versioned independently:

| Component  | Tag format         |
| ---------- | ------------------ |
| Backend    | `backend-X.Y.Z`    |
| Agent      | `agent-X.Y.Z`      |
| Frontend   | `frontend-X.Y.Z`   |
| Helm chart | `helm-chart-X.Y.Z` |

Maintainers approve and merge release PRs; `release-please` handles tagging,
CHANGELOG generation, and image/chart publishing.

## Governance Changes

Updates to this document, the Code of Conduct, the project license, or the
roadmap require **2/3 vote of active Maintainers** and a 7-day review window.

## Security

See [SECURITY.md](SECURITY.md) for vulnerability reporting and disclosure
policy.
