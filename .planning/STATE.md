# State: Temporal AI Docs Pipeline

**Last updated:** 2026-05-15T22:32:00Z
**Milestone:** v1
**Phase 1 Status:** ✓ COMPLETE (walked skeleton delivered, 33 tests passing, 3 commits)

---

## Project Reference

**Core value:** Generate reliable Ubuntu-specific package documentation from source metadata with minimal manual effort.

**Current focus:** Phase 1 COMPLETE — Foundation (control-plane contracts, HTTP start endpoint, Temporal skeleton, idempotency, preflight auth, search attributes, run ledger)

---

## Current Position

| Field | Value |
|-------|-------|
| Current phase | Phase 1: Foundation |
| Current plan | 01-01-PLAN.md (executed) |
| Phase status | ✓ COMPLETE — 3 tasks (contracts, HTTP+workflow, auth+search), 33 tests, 3 commits |
| Overall progress | ████░░░░░░ 20% (1/5 phases complete) |

---

## Phase Pointer

```
✓ Phase 1: Foundation (COMPLETE)
  FLOW-01, FLOW-02, FLOW-03, FLOW-04 delivered
  Commits: d3b85fc (Task 1 RED), 068c56b (Task 2), 90ad7a9 (Task 3), b6c246f (SUMMARY)
  Next action: /gsd-discuss-phase 2 or /gsd-plan-phase 2
```

---

## Performance Metrics

| Metric | Value |
|--------|-------|
| Phases total | 5 |
| Phases complete | 1 |
| Requirements total (v1) | 20 |
| Requirements mapped | 20 |
| Requirements delivered (Phase 1) | 4 (FLOW-01–04) |
| Tests passing | 33/33 |
| Code commits | 4 |

---

## Accumulated Context

### Key Decisions Logged

| Decision | Rationale |
|----------|-----------|
| Workflow ID = `docs/{packageName}/{gitHash}` | Idempotency key — duplicate starts rejected by default |
| Four isolated task queues | `orchestration`, `repo-io`, `generation`, `validation` — LLM/git slowness must not starve orchestration |
| Section ownership via HTML comment markers | `<!-- GENERATED:INSTALL START/END -->` — must be in output schema from day one |
| Go orchestration + Python sidecar | Go for workflow/activity workers; Python for deep parsing and LLM-adjacent generation |
| Run ledger outside Temporal history | PostgreSQL for idempotency tokens and artifact fingerprints; Temporal history is not cross-run state |
| Evidence-linked generation only | Every actionable claim references a mined source file/line or is explicitly caveated — no LLM-only path |
| Command safety policy precedes publish | Policy engine must be active before any generated content reaches the output repo |

### Active Todos

- [x] Phase 1: Foundation (walking skeleton) — COMPLETE
  - [x] Task 1: Define control-plane contracts and ledger schema
  - [x] Task 2: Implement HTTP start endpoint and Temporal workflow
  - [x] Task 3: Add source-access preflight and search metadata
  - [x] Tests: 33 passing ✓
- [ ] Phase 2: Repository I/O (extraction and mining)
- [ ] Phase 3: Generation (content synthesis)
- [ ] Phase 4: Publishing (patching and provenance)
- [ ] Phase 5: Operations (monitoring, alerting, feedback loops)

### Blockers

None. Phase 1 complete; ready to discuss/plan Phase 2.

### Notes

**Phase 1 Delivered:**
- HTTP start endpoint: `POST /v1/docs/runs` returns 202 Accepted with runId+workflowId
- Run idempotency: Duplicate (packageName, gitHash) rejected by ledger unique constraint
- Temporal workflow skeleton: Accepts start request, runs source preflight, emits search attributes
- Run ledger: PostgreSQL schema with status, error tracking, search-attribute indexes
- Source-access preflight: `git ls-remote` activity probe before expensive work begins
- Search attributes: 8 typed fields (package, hash, status, workflow/run IDs, repo, error, publish state) — queryable
- Tests: All acceptance criteria from plan verified (33 tests, 0 failures)

**Architecture Snapshot:**
```
HTTP 202 Accepted → Temporal workflow start (docs/{pkg}/{hash})
  → CheckSourceAccessActivity (preflight)
  → MarkRunStatusActivity (ledger update)
  → [Phase 2: mining] [Phase 3: generation] [Phase 4: publish]
  → SearchAttributes emitted at each state transition
```

**Next Phase Context:**
- Phase 2 will add repository mining tasks and extend the workflow to call extraction activities
- The walking skeleton (Phase 1) is stable and backward-compatible for Phase 2+ additions
- All locks from Phase 1 CONTEXT.md and SKELETON.md are enforced in the implementation

---

## Session Continuity

**To resume Phase 2:** Read `.planning/phases/02-extraction/` then run `/gsd-discuss-phase 2` or `/gsd-plan-phase 2`.

**To review Phase 1:** See `.planning/phases/01-foundation/01-01-SUMMARY.md` and git commits:
- d3b85fc (Task 1 contracts+schema)
- 068c56b (Task 2 HTTP+workflow+ledger)
- 90ad7a9 (Task 3 auth+main)
- b6c246f (SUMMARY.md)

---
*State updated: 2026-05-15T22:32:00Z — Phase 1 execution complete*

