# Webhook security: HMAC signature validation

This document describes the HMAC-SHA256 authentication scheme that protects the public
`POST /webhook/alertmanager` endpoint, the rate-limit applied alongside it, and how to
configure the upstream sender (Alertmanager / proxy / sidecar) to produce a valid signature.

For general policy, supported versions, and the security disclosure process see
[SECURITY.md](../SECURITY.md). This file zooms into the webhook layer specifically.

---

## Why this exists

`/webhook/alertmanager` is the only public ingress on the backend. Without a signature scheme
anyone who can reach the endpoint can inject synthetic alerts, trigger the analysis goroutine,
and consume LLM quota. The W2 supply-chain workstream introduces:

1. **HMAC-SHA256 signature validation** on the request body.
2. **Per-IP token-bucket rate limiting** (default 100 req/min) on the same route.
3. **Opt-in rollout** so existing operators are not broken at upgrade time.

---

## Threat model

| Threat                                          | Mitigation                                       |
| ----------------------------------------------- | ------------------------------------------------ |
| Unauthenticated alert injection                 | HMAC-SHA256 over raw body, hex-encoded           |
| Body tampering after signing                    | HMAC covers the full body bytes                  |
| Brute-force / DoS amplification                 | Per-IP rate limit (`WEBHOOK_RATE_LIMIT`)         |
| Timing side-channel on signature compare        | `crypto/hmac.Equal` (constant time)              |
| Operator upgrades and forgets to set the secret | Opt-in mode + single startup warning             |

Out of scope for this document:

- Replay protection (no nonce / timestamp window yet — see "Future work").
- mTLS / network-level authentication.
- Per-tenant signing keys.

---

## On-the-wire protocol

The sender computes:

```
signature = hex( HMAC_SHA256( shared_secret, raw_request_body ) )
```

and delivers it in the `X-Webhook-Signature` header. The header is configurable via
`WEBHOOK_HMAC_HEADER` (default `X-Webhook-Signature`). A `sha256=` prefix is accepted for
GitHub-style payload compatibility — the prefix is stripped before comparison.

The backend recomputes the digest over the raw body and compares the two byte slices with
`crypto/hmac.Equal`. A mismatch yields HTTP `401`. A missing header (when a secret is
configured) also yields `401`. Requests over the rate limit yield HTTP `429` with a
`Retry-After: 60` header.

### Opt-in mode

When `WEBHOOK_HMAC_SECRET` is empty (or unset) the middleware logs:

```
WARNING: WEBHOOK_HMAC_SECRET is not set; /webhook/alertmanager runs in opt-in mode
```

once at startup and forwards requests unchanged. This mirrors how the endpoint behaves in
versions before W2 — operators can upgrade safely and roll the secret out as a follow-up.

> **Strong recommendation**: enable HMAC verification in production. Opt-in mode exists only
> as a transitional hatch. The backend release notes will mark a future minor as the version
> where HMAC becomes the default and the warning becomes an error.

---

## Generating a secret

Pick a high-entropy key (>= 32 bytes / 256 bits). Examples:

```bash
# 32 bytes hex (recommended)
openssl rand -hex 32

# Or with /dev/urandom
head -c 32 /dev/urandom | xxd -p -c 64
```

Store it as a Kubernetes Secret. The chart reads it from the Secret named in
`backend.webhookSecret.existingSecret`:

```bash
kubectl create secret generic kube-rca-webhook \
  --namespace kube-rca \
  --from-literal=hmac-secret="$(openssl rand -hex 32)"
```

```yaml
backend:
  webhookSecret:
    existingSecret: "kube-rca-webhook"
    key: "hmac-secret"
    headerName: "X-Webhook-Signature"
    rateLimitPerMinute: 100
```

When `existingSecret` is empty the deployment template skips the env injection and the
backend boots in opt-in mode.

---

## Configuring the sender

Native Alertmanager **does not sign outgoing webhooks** today (upstream tracking issue:
[prometheus/alertmanager#3252](https://github.com/prometheus/alertmanager/issues/3252)).
Until that lands, operators have three options.

### Option 1 — Sidecar/proxy (recommended)

Run a lightweight signing proxy in front of the backend. Alertmanager points at the proxy;
the proxy adds the `X-Webhook-Signature` header before forwarding to KubeRCA. Examples:

- A small `gin` / `chi` Go service (~30 lines) with the same secret mounted.
- An Envoy / NGINX with a Lua / WASM filter computing HMAC.
- An OpenFunction / Knative function for serverless setups.

Reference signing function (Go):

```go
mac := hmac.New(sha256.New, []byte(secret))
mac.Write(body)
req.Header.Set("X-Webhook-Signature", hex.EncodeToString(mac.Sum(nil)))
```

### Option 2 — Network isolation + shared secret

If Alertmanager and KubeRCA share a private network (same VPC, mesh, NetworkPolicy) operators
can rely on network controls and leave the backend in opt-in mode. This is acceptable for
homelab / single-cluster installs but does **not** satisfy the "OSS hardened" target.

### Option 3 — Fork / patch Alertmanager

Some operators apply local patches to `alertmanager` to add HMAC headers. This is supported
end-to-end but is out of scope for upstream KubeRCA.

---

## Rate limiting

`WEBHOOK_RATE_LIMIT` (default `100`) caps the number of requests per minute per remote IP
using a `golang.org/x/time/rate` token bucket. Set to `0` to disable.

Because the in-memory limiter is per-pod and the backend currently runs as `replicas=1`
(SSE hub constraint), the limit is also a global ceiling. For multi-replica deployments the
limit should be enforced upstream (ingress, mesh, API gateway).

---

## Future work

- **Replay protection**: add a `X-Webhook-Timestamp` header signed alongside the body, with a
  configurable acceptance window (e.g. ±5 min) and an LRU of recently seen `(timestamp, mac)`
  pairs.
- **Per-tenant secrets**: derive the signing key from a tenant ID once multi-tenant support
  lands.
- **Default-on**: flip `WEBHOOK_HMAC_SECRET` from opt-in to required, gated by a chart-level
  feature flag and a deprecation cycle.

If you find a security issue in this area please follow [SECURITY.md](../SECURITY.md) and use
the encrypted GitHub advisory channel.
