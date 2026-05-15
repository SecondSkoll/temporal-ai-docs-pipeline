# State: Temporal AI Docs Pipeline

**Last updated:** 2026-05-15
**Milestone:** v1

---

## Project Reference

**Core value:** Generate reliable Ubuntu-specific package documentation from source metadata with minimal manual effort.

**Current focus:** Phase 1 — Foundation (input contract, workflow skeleton, idempotency, preflight auth)

---

## Current Position

| Field | Value |
|-------|-------|
| Current phase | Phase 1: Foundation |
| Current plan | None yet — awaiting `/gsd-plan-phase 1` |
| Phase status | Not started |
| Overall progress | ░░░░░░░░░░ 0% (0/5 phases complete) |

---

## Phase Pointer

```
→ Phase 1: Foundation
  FLOW-01, FLOW-02, FLOW-03, FLOW-04
  Next action: /gsd-plan-phase 1
```

---

## Performance Metrics

| Metric | Value |
|--------|-------|
| Phases total | 5 |
| Phases complete | 0 |
| Requirements total (v1) | 20 |
| Requirements mapped | 20 |
| Requirements validated | 0 |

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

- [ ] Run `/gsd-plan-phase 1` to plan Phase 1: Foundation

### Blockers

None.

### Notes

- Research SUMMARY.md confirms 5-phase architecture aligned with requirements groupings.
- Top risks per research: hallucinated docs, stale source mismatch, unsafe command extraction, brittle parsers.
- All 20 v1 requirements mapped; v2 deferred (PROV-01, PROV-02, QUAL-01, QUAL-02, DIST-01, DIST-02).

---

## Session Continuity

**To resume:** Read `.planning/ROADMAP.md` and `.planning/STATE.md`, then run `/gsd-plan-phase 1`.

---
*State initialized: 2026-05-15*
