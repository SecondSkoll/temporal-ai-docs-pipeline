# Phase 1: Foundation - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-05-15
**Phase:** 1-Foundation
**Areas discussed:** Trigger entrypoint, Trigger response, Duplicate runs, Auth preflight order, Run metadata

---

## Trigger entrypoint

| Option | Description | Selected |
|--------|-------------|----------|
| HTTP start endpoint | A small API validates the payload and starts the Temporal workflow | ✓ |
| Queue consumer | An event/topic consumer receives package events and starts the workflow | |
| Direct CLI start | Automation or operators start the workflow directly | |

**User's choice:** HTTP start endpoint
**Notes:** The endpoint validates the payload and starts the workflow.

---

## Trigger response

| Option | Description | Selected |
|--------|-------------|----------|
| Accepted + run ID | Validate, start the workflow, and return an immediate workflow/run identifier | ✓ |
| Accepted + status URL | Return a location to poll for progress as well as the run ID | |
| Wait for completion | Only return when the run finishes | |

**User's choice:** Accepted + run ID
**Notes:** The caller should get an immediate handle to the run rather than waiting for the workflow to finish.

---

## Duplicate runs

| Option | Description | Selected |
|--------|-------------|----------|
| Return existing run | Treat it as idempotent and hand back the original run ID/status | |
| Reject as duplicate | Fail fast with an already-submitted error | ✓ |
| Restart explicitly | Allow a forced restart path for operators | |

**User's choice:** Reject as duplicate
**Notes:** The same package and git hash should not be reused or auto-restarted through a second start request.

---

## Auth preflight order

| Option | Description | Selected |
|--------|-------------|----------|
| Check both up front | Validate source read and output write before expensive work starts | |
| Source now, output later | Only verify source access early; defer output write until publish time | ✓ |
| Source first, output if needed | Validate output write before extraction so the workflow fails early if publishing will be impossible | |

**User's choice:** Source now, output later
**Notes:** Source access is checked immediately; output write checks are deferred until the workflow reaches the publishing path.

---

## Run metadata

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal lookup | Only package, git hash, and final status — keep search attrs small | |
| Add run context | Also include workflow/run ID and phase name for debugging | |
| Richer ops view | Add repo URL, error class, and publish state too | ✓ |

**User's choice:** Richer ops view
**Notes:** Keep the required lookup fields, but add extra operational metadata so search stays useful for debugging and later ops workflows.

---

## the agent's Discretion

None.

## Deferred Ideas

- Forced restart / operator override for duplicate package-plus-hash submissions — noted for a later ops workflow if manual replay becomes necessary.
