# Roadmap: Temporal AI Docs Pipeline

**Milestone:** v1
**Mode:** mvp
**Granularity:** standard
**Created:** 2026-05-15

## Phases

- [ ] **Phase 1: Foundation** — Input validation, workflow skeleton, idempotency, and preflight auth
- [ ] **Phase 2: Repository Mining** — Exact-hash checkout and layered Ubuntu-focused extraction with evidence records
- [ ] **Phase 3: Generation, Safety, and Confidence** — Evidence-linked doc generation with scoring, caveats, and command safety gates
- [ ] **Phase 4: Publishing and Patching** — Section-level patch publish to the output repository with provenance metadata
- [ ] **Phase 5: Reliability and Observability** — Failure taxonomy, metrics, timeout budgets, and queue isolation

## Phase Details

### Phase 1: Foundation
**Goal**: Pipeline can be triggered, validates required inputs, enforces deterministic idempotency, and gates on auth before any expensive work begins.
**Mode:** mvp
**Depends on**: Nothing
**Requirements**: FLOW-01, FLOW-02, FLOW-03, FLOW-04
**Success Criteria** (what must be TRUE):
  1. A trigger payload missing `packageName`, `sourceRepo`, or `gitHash` is rejected before the workflow starts.
  2. Two concurrent start requests for the same `(packageName, gitHash)` pair result in exactly one workflow execution.
  3. A run against a source repository with insufficient credentials fails at preflight, not mid-execution.
  4. Completed runs are searchable by package name, git hash, and final status via Temporal Search Attributes.
**Plans**: 1

Plans:
- [ ] 01-01-PLAN.md — HTTP start endpoint, deterministic workflow ID, preflight auth, and searchable run metadata

### Phase 2: Repository Mining
**Goal**: Pipeline extracts authoritative Ubuntu-relevant signals from the exact source revision and emits structured evidence records.
**Mode:** mvp
**Depends on**: Phase 1
**Requirements**: EXTR-01, EXTR-02, EXTR-03, EXTR-04
**Success Criteria** (what must be TRUE):
  1. A checkout that does not match the requested git hash causes the run to fail immediately with a clear error.
  2. Extraction produces structured signals from packaging metadata (snapcraft.yaml, debian/control), install scripts, and repository documentation — in that priority order.
  3. Every extracted claim carries a source file path and line-range pointer traceable to the checked-out revision.
  4. Each run emits extraction coverage metrics (files examined, signals found, coverage ratio) that are persisted with the run record.
**Plans**: TBD

### Phase 3: Generation, Safety, and Confidence
**Goal**: Pipeline generates Ubuntu-focused documentation sections grounded in extracted evidence, with per-section confidence scoring, structured caveat injection, and command safety gates blocking unsafe content from reaching the output.
**Mode:** mvp
**Depends on**: Phase 2
**Requirements**: GEN-01, GEN-02, GEN-03, GEN-04
**Success Criteria** (what must be TRUE):
  1. Each generated output includes Ubuntu-focused install and configuration sections derived from extraction evidence.
  2. Every generated section carries a machine-readable confidence score computed from evidence density and source authority.
  3. A section whose confidence falls below the threshold is published with a structured, reason-coded caveat instead of causing the run to fail.
  4. A generated command matching the blocked tier of the safety policy is withheld from the output and logged; the run is not aborted unless all critical sections are unsafe.
**Plans**: TBD

### Phase 4: Publishing and Patching
**Goal**: Pipeline writes section-level patches to the output repository, creates new docs when none exist, and attaches provenance metadata to every published output.
**Mode:** mvp
**Depends on**: Phase 3
**Requirements**: PUB-01, PUB-02, PUB-03, PUB-04
**Success Criteria** (what must be TRUE):
  1. Generated docs appear in `packages/{packageName}/` in the output repository after a successful run.
  2. Re-running the pipeline for a package with existing docs updates only the sections whose content changed; unchanged sections are not touched.
  3. Running the pipeline for a package with no prior docs creates a complete documentation file in the correct output location.
  4. Every published commit includes provenance metadata identifying the source git hash, workflow run ID, and confidence bucket.
**Plans**: TBD

### Phase 5: Reliability and Observability
**Goal**: Pipeline operates at production-grade reliability with explicit failure taxonomy, baseline operational metrics, enforced activity timeout budgets, and stage-isolated task queues.
**Mode:** mvp
**Depends on**: Phase 4
**Requirements**: OPS-01, OPS-02, OPS-03, OPS-04
**Success Criteria** (what must be TRUE):
  1. A network timeout on a git fetch is retried automatically; a 403 auth rejection terminates the activity immediately as non-retryable.
  2. An operational dashboard shows success rate, failure rate, p95 latency, and low-confidence ratio across pipeline runs.
  3. No activity can execute for longer than its configured `scheduleToClose` budget; breach causes a classified failure, not an indefinite hang.
  4. A slowdown in the generation stage does not delay workflow orchestration heartbeats because orchestration and generation run on separate queues.
**Plans**: TBD

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 0/0 | Not started | - |
| 2. Repository Mining | 0/0 | Not started | - |
| 3. Generation, Safety, and Confidence | 0/0 | Not started | - |
| 4. Publishing and Patching | 0/0 | Not started | - |
| 5. Reliability and Observability | 0/0 | Not started | - |

---
*Roadmap created: 2026-05-15*
*Coverage: 20/20 v1 requirements mapped*
