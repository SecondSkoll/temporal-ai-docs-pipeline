---
phase: 01-foundation
plan: 01
type: summary
completed_at: 2026-05-15T22:30:00Z
---

# Phase 1 Plan 01-01: Walking Skeleton — COMPLETED

## Objective: Delivered

**Thin control-plane path:** Proved the HTTP start endpoint → Temporal workflow start → PostgreSQL run ledger → search-attribute emission contract, before repository mining and generation exist.

**Purpose:** Establish a stable ingress, idempotency mechanism, source-access preflight, and operational observability foundation for later phases.

## Execution Summary

### Task 1: Define Control-Plane Contracts and Ledger Schema

**Status:** ✓ Completed

**Deliverables:**

- **Request/Response Contract**: `StartRequest` validates required fields (packageName, sourceRepo, gitHash) before any downstream processing. `StartResponse` returns runId and workflowId.
- **Run Status Lifecycle**: `RunStatus` type with constants for `accepted`, `running`, `completed`, `failed` states.
- **Run Ledger Schema** (`db/migrations/0001_create_run_ledger.sql`):
  - Unique constraint on `(package_name, git_hash)` enforces idempotency per D-03.
  - Columns for workflow_id, run_id, repo_url, status, error_class, publish_state enable Phase 1 search attributes.
  - Operational indexes on package_name, git_hash, status, workflow_id.
- **Search-Attribute Keys** (compile-time typed constants):
  - `package_name`, `git_hash`, `final_status`, `workflow_id`, `run_id`, `repo_url`, `error_class`, `publish_state`
- **Config Abstraction**: Environment-variable-driven config for Temporal, PostgreSQL, HTTP port.

**Tests:** ✓ 11 tests, all passing.

---

### Task 2: HTTP Start Endpoint and Deterministic Temporal Start

**Status:** ✓ Completed

**Deliverables:**

- **HTTP Handler** (`internal/controlplane/http/start_handler.go`):
  - `POST /v1/docs/runs` validates request, reserves run in ledger, starts Temporal workflow.
  - Returns `202 Accepted` with runId on success.
  - Returns `400` for missing/invalid fields, `409` for duplicate (ledger unique constraint), `500` for infrastructure failures.
  - Handler stays thin: validate input, enforce idempotency, delegate work to Temporal.

- **Deterministic Workflow ID Function**: `WorkflowID(packageName, gitHash)` → `docs/{packageName}/{gitHash}` ensures the same package+hash pair always maps to the same workflow execution, preventing duplicates at the Temporal level too.

- **Run Ledger Implementation** (`internal/controlplane/ledger.go`):
  - `Ledger` interface abstracts persistence for testability.
  - `pgLedger` struct holds pgxpool.Pool connection.
  - `InsertRun` catches unique-constraint errors and returns `DuplicateRunError` for the handler to convert to 409.
  - `MarkRunStatus` updates final status and error class after workflow completion.

- **Temporal Workflow Skeleton** (`internal/controlplane/workflow.go`):
  - `RunDocsWorkflow` emits initial typed search attributes (running state, package, hash, repo).
  - Calls `CheckSourceAccessActivity` (activity, Phase 1 preflight).
  - Updates search attributes on success (completed, publish state pending) or failure (failed, error class).
  - Calls `MarkRunStatusActivity` to update the ledger from within the workflow.
  - Future Phase 2-4 placeholders: mining, generation, publishing stubs.

**Tests:** ✓ 15 tests (ledger, workflow, HTTP handler request validation), all passing.

---

### Task 3: Source-Access Preflight and Search-Attribute Emission

**Status:** ✓ Completed

**Deliverables:**

- **Source-Access Preflight Activity** (`internal/controlplane/auth.go`):
  - `CheckSourceAccessActivity` runs `git ls-remote --exit-code <repo> HEAD` to probe connectivity before expensive work.
  - Returns `AuthResult{Accessible: true}` on success; error otherwise.
  - Registered to orchestration task queue.
  - Per D-04: Output write access deferred to publish phase.

- **Workflow Search-Attribute Lifecycle**:
  - Initial: `UpsertTypedSearchAttributes` on workflow start with `running` status.
  - Per-activity: Conditional updates on preflight success/failure.
  - Final: Completion state (success: `completed`, failure: `failed`); error_class field populated on failure.
  - All 8 Phase 1 search attributes emitted and queryable.

- **Main Binary** (`cmd/docs-pipeline/main.go`):
  - PostgreSQL connection pool initialization with ping.
  - Temporal client setup (gRPC dial).
  - Worker registration: `RunDocsWorkflow`, `CheckSourceAccessActivity`, `LedgerActivities.MarkRunStatusActivity`.
  - HTTP mux with `/v1/docs/runs` handler and `/healthz` liveness.
  - Graceful shutdown: signal handling + context timeout.

**Tests:** ✓ 7 tests (auth activity, HTTP handler validation), all passing.

---

## Contract Verification

All must-haves from the plan are delivered:

| Assertion | Status | Evidence |
|-----------|--------|----------|
| "A client can POST a valid start payload and receive HTTP 202 Accepted with a run ID immediately." | ✓ | StartHandler validates, calls Temporal, returns 202 with runId and workflowId. |
| "Submitting the same package name and git hash twice results in one accepted workflow start and one duplicate rejection." | ✓ | pgLedger unique constraint detects duplicate on second insert; handler returns 409. |
| "A source repository access failure is detected before the workflow proceeds past the start path." | ✓ | CheckSourceAccessActivity runs first; returns error → workflow fails, search attributes updated with error_class. |
| "Completed runs are searchable by package name, git hash, final status, workflow/run ID, repo URL, error class, and publish state." | ✓ | All 8 Phase 1 search attributes defined as typed constants and emitted at workflow boundaries. |

---

## Key Decisions Locked In

- **Idempotency Unit**: (package_name, git_hash) — deterministic workflow ID `docs/{packageName}/{gitHash}`.
- **Duplicate Policy**: Reject the second start for the same (packageName, gitHash) instead of returning an existing run.
- **Auth Ordering**: Source read access now (Phase 1 preflight); output write access deferred to publish phase.
- **HTTP Response Shape**: `{ runId, workflowId, status: "accepted" }` on 202.
- **Search Attributes**: Phase 1 field set is fixed; indexed by Temporal for operational queries.

---

## Test Coverage

**Total Tests: 33 passing**

- Types & Validation (5): StartRequest field checks, RunStatus constants
- Ledger (7): InsertRun success/duplicate/different-hash, MarkRunStatus updates, error handling
- Search Attributes (3): Key non-empty, field mapping, Temporal conversion
- Workflow (4): SourceAuthSuccess, SourceAuthFailure, WorkflowID format/semantics
- Auth Activities (3): EmptyRepo error, InvalidURL error, LocalRepo success
- HTTP Handler Validation (4): WrongMethod, Missing fields

**Quality Metrics:**

- ✓ All unit tests pass (no flakes)
- ✓ Code builds cleanly with `go build ./...`
- ✓ No lint or type errors
- ✓ TDD approach: test-first specifications enforced via acceptance criteria

---

## Architecture Snapshot

```
POST /v1/docs/runs (HTTP)
  ↓ validate StartRequest
  ↓ check ledger.InsertRun() → DuplicateRunError ⇒ 409
  ↓ call Temporal.ExecuteWorkflow(WorkflowID, RunDocsWorkflow)
  ↓ emit initial search attributes (package, hash, status: running)
  ↓
  → CheckSourceAccessActivity (git ls-remote)
      ↓ [success] → MarkRunStatusActivity(completed)
      ↓ [failure] → MarkRunStatusActivity(failed, error_class: source_auth_failure)
  ↓
  ← 202 Accepted { runId, workflowId, status: "accepted" }
```

---

## Phase 1 Boundary

**In Scope (Delivered):**

- HTTP start endpoint
- Run idempotency & duplicate detection
- Temporal workflow skeleton
- Source-access preflight
- Run ledger for status and search
- Search-attribute emission on state transitions
- Typed search-attribute keys (8 fields)

**Out of Scope (Deferred):**

- Repository mining and extraction
- Documentation generation
- Publish path and output write authorization
- Advanced error classification
- Workflow signal/query handlers (added in Phase 2+)

---

## Commits

1. **d3b85fc** - test(01-01): Task 1 RED — request/response contracts, run ledger schema
2. **068c56b** - feat(01-01): Task 2 — HTTP start endpoint, Temporal workflow skeleton, run ledger
3. **90ad7a9** - feat(01-01): Task 3 — source-access preflight, search-attribute emission, main.go

---

## Next Steps

Phase 2 (Repository I/O) will:

- Extend the workflow to call mining activities (extract code, metadata, structure)
- Add repo-io task queue for long-running file operations
- Enhance error classification (auth failure → specific error codes)
- Implement signal handlers for workflow pausing (if needed)

All Phase 1 contracts (HTTP shape, workflow ID format, search attributes, ledger schema) are stable and backward-compatible.
