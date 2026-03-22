# Webhook Name Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a required `name` field when creating or editing webhook configs, show that name in the webhook list, and persist it end-to-end through the backend and database.

**Architecture:** Extend the shared webhook contract with a `name` property, then thread it through frontend form state, list rendering, backend request/response models, service mapping, and PostgreSQL persistence. Keep backward compatibility for existing rows by adding a schema migration step that populates empty names from existing channel or URL data.

**Tech Stack:** React, TypeScript, Vite, Go, Gin, PostgreSQL, Go test

---

### Task 1: Add backend failing tests for `name`

**Files:**
- Modify: `../backend/internal/service/webhook_test.go`
- Test: `../backend/internal/service/webhook_test.go`

**Step 1: Write the failing test**

Add assertions that `WebhookConfigRequest.Name` is copied into `model.WebhookConfig.Name` for both create and update flows.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service`
Expected: FAIL because `Name` does not exist on the request/model or is not mapped.

**Step 3: Write minimal implementation**

Add `Name` to webhook model/request structs and service mapping.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service`
Expected: PASS

**Step 5: Commit**

```bash
git add ../backend/internal/model/webhook.go ../backend/internal/service/webhook.go ../backend/internal/service/webhook_test.go
git commit -m "feat: map webhook names in service"
```

### Task 2: Persist `name` in backend storage

**Files:**
- Modify: `../backend/internal/db/webhook.go`
- Modify: `../backend/internal/model/webhook.go`
- Modify: `../backend/internal/handler/webhook.go`

**Step 1: Write the failing test**

Use the service tests from Task 1 as the first red step, then add a DB-level compatibility change by updating SQL selections/inserts/updates to include `name`.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service`
Expected: FAIL until the `name` field is fully threaded through the backend.

**Step 3: Write minimal implementation**

Add `name` column creation/normalization in `EnsureWebhookSchema`, include it in `SELECT`, `INSERT`, and `UPDATE`, and keep old rows readable by filling empty names from `channel`, then `url`, then `'Unnamed Webhook'`.

**Step 4: Run test to verify it passes**

Run: `go test ./...`
Expected: PASS

**Step 5: Commit**

```bash
git add ../backend/internal/db/webhook.go ../backend/internal/handler/webhook.go ../backend/internal/model/webhook.go
git commit -m "feat: persist webhook names"
```

### Task 3: Add frontend failing checks for webhook `name`

**Files:**
- Modify: `src/utils/api.ts`
- Modify: `src/components/WebhookSettings.tsx`
- Modify: `src/components/WebhookList.tsx`

**Step 1: Write the failing test**

Because there is no frontend test runner in this repo, create a compile-time red step by introducing `name` into the TypeScript contracts first and intentionally leaving component usage incomplete so `tsc` fails where required state/props are missing.

**Step 2: Run test to verify it fails**

Run: `npm run build`
Expected: FAIL with TypeScript errors until all form state and list rendering paths handle `name`.

**Step 3: Write minimal implementation**

Add `name` to webhook API types and normalization, add required `Name` input/validation in `WebhookSettings`, load/save it in edit/create mode, and render the configured name as the primary label in `WebhookList`.

**Step 4: Run test to verify it passes**

Run: `npm run build`
Expected: PASS

**Step 5: Commit**

```bash
git add src/utils/api.ts src/components/WebhookSettings.tsx src/components/WebhookList.tsx
git commit -m "feat: add webhook names in frontend"
```

### Task 4: End-to-end verification

**Files:**
- Verify: `src/components/WebhookSettings.tsx`
- Verify: `src/components/WebhookList.tsx`
- Verify: `../backend/internal/model/webhook.go`
- Verify: `../backend/internal/service/webhook.go`
- Verify: `../backend/internal/db/webhook.go`

**Step 1: Run backend verification**

Run: `go test ./...`
Expected: PASS

**Step 2: Run frontend verification**

Run: `npm run build`
Expected: PASS

**Step 3: Sanity-check behavior**

Confirm:
- webhook create/edit requires `name`
- webhook list shows configured `name`
- backend JSON includes `name`
- DB schema stores `name`
- existing rows receive fallback names during schema normalization

**Step 4: Commit**

```bash
git add docs/plans/2026-03-22-webhook-name-support.md
git commit -m "docs: add webhook name support plan"
```
