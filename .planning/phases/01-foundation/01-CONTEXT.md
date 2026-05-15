# Phase 1: Foundation - Context

**Gathered:** 2026-05-15
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase delivers the control-plane foundation for the Temporal pipeline: the HTTP trigger surface, deterministic workflow identity and duplicate handling, early source-repo authorization checks, and searchable run metadata needed for later debugging and operations.

</domain>

<decisions>
## Implementation Decisions

### Trigger Surface
- **D-01:** Use an HTTP start endpoint as the public entrypoint. It validates the package payload and starts the Temporal workflow.
- **D-02:** The endpoint returns `Accepted + run ID` immediately instead of waiting for completion.

### Duplicate Runs
- **D-03:** Reject duplicate submissions for the same package and git hash instead of returning an existing run or auto-restarting it.

### Auth Ordering
- **D-04:** Check source repository access up front, but defer output repository write checks until the workflow reaches the later publishing path.

### Run Metadata
- **D-05:** Keep the required search fields for package, git hash, and final status, and add richer operational fields: workflow/run ID, repo URL, error class, and publish state.

### the agent's Discretion
None — all discussed gray areas were decided explicitly.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Context
- [.planning/PROJECT.md](../../../PROJECT.md) — project vision, core value, constraints, and key decisions
- [.planning/REQUIREMENTS.md](../../../REQUIREMENTS.md) — v1 requirement definitions and out-of-scope boundaries
- [.planning/ROADMAP.md](../../../ROADMAP.md) — Phase 1 scope and requirement mapping
- [.planning/STATE.md](../../../STATE.md) — current phase pointer and accumulated project state

No external specs — requirements and implementation decisions are captured in the planning documents above.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- None yet — this workspace currently contains only a README and planning artifacts.

### Established Patterns
- Temporal orchestration is already locked in the project decisions, but no implementation pattern exists yet in source files.

### Integration Points
- Future Phase 1 implementation will need a Temporal workflow entrypoint, repository authorization checks, and search-attribute wiring.

</code_context>

<specifics>
## Specific Ideas

- The public entrypoint should be an HTTP start endpoint.
- The endpoint should return an accepted response and run ID immediately.
- Re-submitting the same package and git hash should be rejected as a duplicate.
- Source access is checked now; output write access is checked later in the workflow.
- Search metadata should remain useful for ops, not just satisfy the minimum requirement.

</specifics>

<deferred>
## Deferred Ideas

- Forced restart / operator override for duplicate package-plus-hash submissions — not requested for v1 and can be revisited if operations later need a manual replay path.

</deferred>

---
*Phase: 1-Foundation*
*Context gathered: 2026-05-15*
