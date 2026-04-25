# ADR-0002: PostgreSQL + pgvector for Similar-Incident Retrieval

- **Status**: Accepted
- **Date**: 2026-04-25
- **Deciders**: Project Maintainers

## Context

When a new alert fires, KubeRCA wants to surface previous incidents that look
similar so operators can short-circuit the investigation. Similarity is
expressed as a vector embedding distance between the new alert payload and
historical incident summaries.

We needed a similarity backend that:

1. Adds zero new infrastructure components for operators who already run
   KubeRCA's PostgreSQL.
2. Supports approximate-nearest-neighbor (ANN) queries at the scale a typical
   cluster will produce — hundreds to low thousands of incidents, not
   billions.
3. Coexists with the same transactional schema we use for incidents, alerts,
   and analyses (so we can join, filter, and paginate using normal SQL).
4. Has a healthy operational story: backups, replicas, observability —
   things PostgreSQL already nails.

## Decision

We use **PostgreSQL with the `pgvector` extension** as the similarity
backend. Embeddings are stored as `vector(N)` columns alongside the related
incident and alert rows; nearest-neighbor lookups use `<=>` (cosine distance)
with an IVF index.

This means KubeRCA ships exactly **one** database. Operators can use the
chart's bundled PostgreSQL (with `pgvector` enabled) or point the chart at an
external instance via `externalDatabase.*` values, as long as the `vector`
extension is available.

## Alternatives Considered

- **Pinecone / Weaviate / Qdrant** — purpose-built vector databases with
  excellent ergonomics and large-scale ANN performance. Rejected: adding a
  second managed service for a few thousand rows is operationally
  disproportionate and complicates the chart and disaster-recovery story.
- **Elasticsearch / OpenSearch with `dense_vector`** — viable but again
  introduces a heavy second store, and most KubeRCA operators do not already
  run Elasticsearch.
- **In-memory FAISS / HNSW** — fast and dependency-light but loses durability
  and forces us to rebuild indexes on restart. Hard to reason about at
  multi-replica scale.

## Consequences

### Positive

- Single database, single backup/restore story, single observability target.
- Joining vector hits with normal incident metadata is trivial SQL.
- Operators familiar with PostgreSQL get an "it just works" install.

### Negative

- IVF index maintenance is non-trivial as the row count grows. We document
  index reindex guidance and expose a controllable `lists` value via Helm.
- `pgvector` is not enabled in every managed PostgreSQL out of the box. We
  add a wait-for-db init container that fails fast with a helpful error if
  the extension is missing.
- At very large scales (tens of millions of incidents), a dedicated vector
  store would outperform pgvector. We explicitly do not target that scale —
  KubeRCA is a per-cluster operations tool, not a multi-tenant SaaS.

### Operational Notes

- The chart's wait-for-db hook attempts `CREATE EXTENSION IF NOT EXISTS
  vector;`. Operators using restricted PostgreSQL roles must pre-install the
  extension and grant usage to the service user.
- Embedding dimensionality is fixed per provider; switching providers
  requires reindexing. The chart documents this in
  `docs/TROUBLESHOOTING.md`.
