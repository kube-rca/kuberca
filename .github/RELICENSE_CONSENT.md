# Relicense Consent: MIT to Apache License 2.0

KubeRCA is being relicensed from the MIT License to the Apache License,
Version 2.0. This document tracks contributor consent for that change.

## Why we are relicensing

- **CNCF ecosystem alignment** — Apache 2.0 is the de facto license across the
  CNCF Kubernetes ecosystem (Kubernetes, Prometheus, Istio, Argo CD, etc.).
  Aligning with our peers reduces friction for downstream integrators.
- **Explicit patent grant + retaliation clause** — Section 3 of Apache 2.0
  grants an explicit patent license to all contributors and includes a patent
  retaliation clause. MIT is silent on patents, which leaves users exposed.
- **Enterprise legal acceptance** — Apache 2.0 is on the allowlist of nearly
  every corporate open source policy. MIT is too, but Apache 2.0 ships with
  the patent and trademark protections enterprise legal teams expect.
- **Project maturity** — KubeRCA has 4 human contributors today, which makes
  consent collection tractable. Doing it now is far cheaper than later.

## What this changes

- The repository's [`LICENSE`](../LICENSE) is replaced with the verbatim
  Apache License, Version 2.0 text.
- A [`NOTICE`](../NOTICE) file is added at the repository root, as required
  by Apache 2.0 Section 4(d).
- All future contributions to KubeRCA are made under Apache License 2.0,
  per Section 5 (Submission of Contributions) of the new license.
- Existing forks must update their `LICENSE` references when picking up
  changes from `main` after this relicense lands.

## How to give consent

You can record your consent in any of the following ways:

1. **Comment on the relicense PR** — A simple "I consent to relicensing my
   prior contributions to KubeRCA under Apache License 2.0." reply on the
   PR is sufficient.
2. **Sign-off on a follow-up commit** — Include
   `Signed-off-by: Your Name <email>` on a commit that explicitly references
   this relicense. The commit body should state your consent.
3. **DCO trailer on a future PR** — A `Signed-off-by:` trailer on any PR
   merged after this document lands implies acceptance of the current
   project license (Apache 2.0). New contributors are covered automatically.
4. **GitHub Discussion** — Open or reply on the relicense Discussion thread
   in the [`kube-rca/kuberca`](https://github.com/kube-rca/kuberca) repo.

If none of those work for you, DM the project lead (see contact below) and
we will record your consent manually with a link back to the message.

## Contact

- **Project lead**: KKamJi (Ethan)
- **Preferred channel**: Open a GitHub Discussion in
  [`kube-rca/kuberca`](https://github.com/kube-rca/kuberca/discussions) and
  tag the relicense topic.
- **Alternate**: PR comment on the relicense PR.

## Contributor checklist

The following identities appear in `git log` history on `main` as of the
relicense PR. Each unique `Name <email>` pair gets its own checkbox so we
can track consent precisely. Bot accounts (`dependabot[bot]`,
`github-actions[bot]`) do not require consent because their commits are
either trivial automation or carry no human authorship.

Some humans authored commits under multiple email addresses; those
identities will be consolidated and consent recorded once per person, but
they are listed separately here for traceability.

### Humans (consent required)

- [ ] binhao &lt;binhao@naver.com&gt;
- [ ] binhao &lt;binhao@neowiz.com&gt;
- [ ] binhao22 &lt;binhao@naver.com&gt;
- [ ] Binoo &lt;73528043+binhao22@users.noreply.github.com&gt;
- [ ] Brilly-Bohyun &lt;102973953+Brilly-Bohyun@users.noreply.github.com&gt;
- [ ] hjk1996 &lt;dunhill741@naver.com&gt;
- [ ] KKamJi &lt;xowl5460@naver.com&gt;
- [ ] KKamJi98 &lt;72260110+KKamJi98@users.noreply.github.com&gt;

### Bots (no consent required)

- [x] dependabot[bot] &lt;49699333+dependabot[bot]@users.noreply.github.com&gt; — automated dependency bumps, no human authorship
- [x] github-actions[bot] &lt;41898282+github-actions[bot]@users.noreply.github.com&gt; — release-please / CI commits, no human authorship

## Reference

- [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)
- [Repository `LICENSE`](../LICENSE)
- [Repository `NOTICE`](../NOTICE)
- [Apache Software Foundation FAQ on relicensing](https://www.apache.org/legal/resolved.html)
