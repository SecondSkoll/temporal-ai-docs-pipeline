# Architecture Patterns: Temporal-Based Ubuntu Documentation Pipeline

**Domain:** Temporal-orchestrated documentation generation
**Researched:** 2026-05-15
**Project context:** Trigger payload contains package name, source repository URL, and exact git hash. Output is generated per-package docs in a separate repository.

## Recommended Architecture

Use a **single parent Workflow per package revision** with **activity-first decomposition** and selective **Child Workflows** only where failure-domain isolation is valuable.

Primary recommendation:
- Parent Workflow type: `GeneratePackageDocsWorkflow`
- Workflow ID: `docs/{packageName}/{gitHash}`
- Activity-driven pipeline for ingest, fetch, analyze, generate, patch, and publish
- Separate Task Queues for orchestration-heavy work, repository IO, LLM generation, and publishing
- Deterministic orchestration in Workflow code; all external IO in Activities

Why this fits:
- Temporal Workflows must remain deterministic; network, git, and LLM calls belong in Activities.
- Activity retries are default and robust; Workflow retries are not default and usually not recommended.
- Per-package-per-revision Workflow ID gives natural idempotency and traceability.

## Major Components and Boundaries

| Component | Responsibility | Boundary | Communicates With |
|-----------|----------------|----------|-------------------|
| Trigger Ingress API | Validates trigger payload and starts Temporal Workflow | No heavy logic; stateless API layer | Temporal Client, metadata store |
| Temporal Parent Workflow (`GeneratePackageDocsWorkflow`) | Durable orchestration, step ordering, policy decisions, status progression | Deterministic code only; no direct network/FS | Activities, optional Child Workflows |
| Metadata/Validation Activity | Normalizes package metadata, validates repo URL and commit format, computes run keys | Pure validation + canonicalization | Parent Workflow |
| Source Fetch Activity | Fetches exact git revision into isolated workspace, verifies commit exists, prepares working tree | External network and disk IO | git remote, local cache store |
| Content Extraction Activity | Extracts Ubuntu-relevant files (install scripts, packaging manifests, docs, config examples) | Read-only repository analysis | source workspace, heuristics rules |
| Ubuntu Context Enrichment Activity | Augments extracted context with Ubuntu conventions (APT/systemd/snap/paths/service mgmt) | Domain knowledge boundary | ruleset store, optional knowledge base |
| Doc Planning Activity | Produces target doc sections and update map for existing docs | Structural planning only | extraction outputs, existing docs snapshot |
| Draft Generation Activity | Calls LLM to generate or revise sections with citations to source paths | LLM IO and prompt logic | LLM provider, prompt templates |
| Section Merge/Patch Activity | Applies minimal diffs to generated sections and preserves untouched sections | Deterministic patch rules + text merge | existing package docs |
| Validation Activity | Checks required sections, basic correctness gates, and Ubuntu-specific completeness checks | Quality gate boundary | merged docs artifact |
| Publish Activity | Writes commit(s) to output repo under per-package path and pushes | VCS write boundary | output git repo |
| Run Ledger Store | Stores run status, idempotency tokens, artifact fingerprints, and publish references | Durable operational state outside Temporal history | ingress, activities, observability |
| Artifact Cache | Caches repo snapshots and intermediate artifacts by content keys | Performance and dedupe boundary | source fetch, extraction, generation |
| Observability/Telemetry | Metrics, structured logs, trace correlation, workflow search attributes | Operational visibility boundary | all components |

## Component Dependency Order (Roadmap-Ready)

This is the explicit build order to decompose roadmap phases with minimal rework.

1. Ingress Contract + Workflow Skeleton
- Define trigger contract and validation schema.
- Start `GeneratePackageDocsWorkflow` with deterministic status transitions.
- No generation yet; stub activities.

2. Exact Revision Fetch + Workspace Isolation
- Implement Source Fetch Activity with commit verification.
- Introduce isolated per-run work directory strategy.
- Add cache key model for repository snapshots.

3. Extraction + Ubuntu Context Enrichment
- Implement repository extraction heuristics.
- Add Ubuntu-specific rule pack and canonical section model.
- Produce normalized context artifact used by downstream steps.

4. Planning + Generation + Merge
- Implement doc planning artifact.
- Implement generation activity and section-level patching strategy.
- Preserve minimal diffs and generated-section ownership markers.

5. Validation + Publish
- Implement quality gates and publish-safe checks.
- Commit and push to output repository in per-package structure.
- Persist publish references in run ledger.

6. Reliability Hardening
- Fine-tune retries/timeouts per activity class.
- Add non-retryable failure classification for permanent errors.
- Add Child Workflow split only if needed for stronger fault isolation.

7. Throughput and Cost Optimization
- Add aggressive caching for fetch/extraction/context artifacts.
- Add queue partitioning and worker scaling by activity type.
- Add metrics-driven tuning (timeouts, backoff, concurrency).

Dependency notes:
- 2 depends on 1.
- 3 depends on 2.
- 4 depends on 3.
- 5 depends on 4.
- 6 depends on 1-5.
- 7 depends on 1-6.

## Data Flow

Input to output flow:
1. Trigger payload arrives: `{ packageName, sourceRepository, gitHash }`.
2. Ingress validates schema and starts Workflow with deterministic `workflowId = docs/{package}/{gitHash}`.
3. Workflow runs metadata validation and canonicalization.
4. Source Fetch Activity retrieves the exact commit and verifies checkout hash match.
5. Extraction Activity creates structured repository context artifact.
6. Ubuntu Enrichment Activity overlays Ubuntu-specific operational context.
7. Planning Activity computes target sections and update/insert map.
8. Generation Activity drafts new or revised generated sections.
9. Merge/Patch Activity updates only generated sections and preserves manual/non-generated areas.
10. Validation Activity enforces gates.
11. Publish Activity commits to output repo path `packages/{packageName}/` and pushes.
12. Workflow records completion metadata and emitted artifact fingerprints.

## Sequence-Style Flow (One End-to-End Execution)

```text
Trigger -> Ingress API: POST package payload
Ingress API -> Temporal: start GeneratePackageDocsWorkflow(workflowId=docs/pkg/hash)
Temporal -> Workflow Worker: schedule workflow task
Workflow -> Metadata Activity: validate/normalize payload
Metadata Activity -> Workflow: canonical metadata + run keys
Workflow -> Source Fetch Activity: fetch repo at exact hash
Source Fetch Activity -> Git Remote: fetch/clone objects
Source Fetch Activity -> Workflow: workspace path + verified hash + snapshot key
Workflow -> Extraction Activity: extract doc-relevant repo context
Extraction Activity -> Workflow: extracted context artifact key
Workflow -> Ubuntu Enrichment Activity: apply Ubuntu rules/context
Ubuntu Enrichment Activity -> Workflow: ubuntu-enriched context key
Workflow -> Planning Activity: build section/update plan
Planning Activity -> Workflow: doc plan
Workflow -> Generation Activity: generate/revise sections
Generation Activity -> Workflow: generated sections artifact
Workflow -> Merge/Patch Activity: apply minimal updates
Merge/Patch Activity -> Workflow: merged docs artifact
Workflow -> Validation Activity: run quality/coverage gates
Validation Activity -> Workflow: pass/fail + diagnostics
Workflow -> Publish Activity: commit/push to output repo package folder
Publish Activity -> Output Repo: create commit and push
Publish Activity -> Workflow: commit sha + published paths
Workflow -> Run Ledger: persist completion record
Workflow -> Temporal: complete
```

## Idempotency Strategy

### Workflow-level idempotency
- Use deterministic Workflow IDs: `docs/{packageName}/{gitHash}`.
- Start policy should reject duplicate starts unless explicit re-run mode is requested.
- Re-run mode should use suffix or dedicated signal/update path, not random IDs.

### Activity-level idempotency
- Every external-side-effect Activity receives `idempotencyKey = {workflowRunId}:{activityId}` or equivalent stable key.
- Publish Activity must be idempotent:
  - If target commit for same content already exists, treat as success.
  - If branch already contains identical tree for target package path, no-op.
- LLM generation can be retried safely because merge/publish steps are content-addressed and deduped.

### Artifact idempotency
- Cache keys are content-based:
  - Repo snapshot key: hash of `{repoURL, gitHash}`
  - Extraction key: hash of snapshot key + extractor version
  - Generation key: hash of plan + enriched context + prompt/template version
- Repeated attempts reuse artifacts rather than recomputing.

## Retry Semantics

Use per-activity retry classes, not one global policy.

### Class A: Transient infrastructure IO (network/git/remote APIs)
- Applies to: Source Fetch, push transport, optional external metadata calls.
- Retry policy:
  - initial interval: 1s-5s
  - backoff coefficient: 2.0
  - max interval: 60s-120s
  - maximum attempts: bounded (for example 8-12)
- Timeouts:
  - set `startToClose` for single attempt
  - set `scheduleToClose` for total retry window

### Class B: Compute-heavy but deterministic transformation
- Applies to: extraction, planning, merge, validation.
- Retry policy:
  - lower attempts (for example 3-5)
  - tighter max interval
- Mark schema/contract violations as non-retryable.

### Class C: LLM generation
- Retry policy:
  - moderate attempts with bounded total window
  - classify provider quota/rate failures as retryable
  - classify prompt contract failures as non-retryable after one corrective pass
- Keep generation in activity boundaries; never in Workflow code.

### Workflow retry guidance
- Do not rely on full Workflow retries for regular failure handling.
- Handle failures within workflow logic and activity retries.
- Deliberately fail Workflow only for terminal conditions.

## Failure Isolation Strategy

### Isolation by task queues and worker pools
- `orchestration-queue`: workflow tasks only
- `repo-io-queue`: source fetch and output-repo write activities
- `generation-queue`: LLM generation activities
- `validation-queue`: static validation activities

This prevents LLM or git slowness from starving orchestration progress.

### Isolation by activity boundaries
- Keep fetch, generate, merge, and publish as separate Activities.
- Avoid giant composite Activities to reduce replay risk and improve targeted retries.

### Optional Child Workflow boundary
Use Child Workflows only if either condition becomes true:
- Package run becomes too event-heavy for parent history budget.
- You need independent lifecycle control per major stage.

Recommended child split when needed:
- `PrepareSourceChildWorkflow` (metadata + fetch + extract)
- `GenerateDocsChildWorkflow` (plan + generate + merge + validate)
- `PublishDocsChildWorkflow` (publish + post-publish verification)

Parent orchestrates and aggregates outcomes.

## Caching Strategy

### Cache tiers
1. Git object/snapshot cache
- Store bare or partial clone objects keyed by repository URL.
- Materialize exact commit worktree on demand.
- Use `--filter=blob:none`/partial clone where compatible to reduce transfer.

2. Extraction artifact cache
- Store parsed repo context keyed by snapshot key + extractor version.
- Reuse across repeated runs for same commit.

3. Enrichment and planning cache
- Cache Ubuntu-enriched context and section plans by content/version keys.

4. Generation cache
- Optional: cache section outputs by `(prompt version, model version, section input hash)`.
- Use cautiously if model nondeterminism is acceptable; still validate output.

### Cache invalidation
- Revision-bound artifacts are immutable by design.
- Invalidate only on tool version change or explicit purge policy.
- Keep a retention policy based on storage budget and access frequency.

## Build Order for Runtime Components

This is the implementation order for engineering execution (not just roadmap narrative):

1. Temporal namespace, task queues, and worker bootstrapping.
2. Trigger ingress and workflow start contract.
3. Workflow skeleton with deterministic state machine and activity stubs.
4. Source fetch implementation with exact-hash verification.
5. Extraction and Ubuntu enrichment artifacts.
6. Planning and merge logic with generated-section ownership model.
7. Generation activity and provider integration.
8. Validation gates and result taxonomy.
9. Publish activity and output-repo commit semantics.
10. Run ledger, observability, and SLA dashboards.
11. Retry/timeouts tuning per class.
12. Performance and cache tuning.

## Patterns to Follow

### Pattern 1: Deterministic orchestrator, non-deterministic activities
**What:** Keep Workflow code as pure orchestration and move all external interactions to Activities.
**When:** Always.
**Why:** Maintains replay safety and aligns with Temporal determinism constraints.

### Pattern 2: Content-addressed artifacts
**What:** Key intermediate outputs by content + version hash.
**When:** Fetch/extract/enrich/plan/generate stages.
**Why:** Enables idempotent retries, dedupe, and reproducibility.

### Pattern 3: Section ownership for minimal diffs
**What:** Mark generated sections and only patch those sections.
**When:** Updating existing docs.
**Why:** Preserves stable history and reduces noisy commits.

### Pattern 4: Error taxonomy with non-retryable classification
**What:** Distinguish permanent input/contract errors from transient infra errors.
**When:** In each activity exception path.
**Why:** Prevents wasteful retries and shortens failure feedback loop.

## Anti-Patterns to Avoid

### Anti-Pattern 1: Git/LLM calls inside Workflow code
- Breaks determinism and causes replay hazards.

### Anti-Pattern 2: One huge activity for end-to-end generation
- Enlarges blast radius and makes retries expensive/non-targeted.

### Anti-Pattern 3: Unbounded retries without schedule-to-close budget
- Risks infinite burn and queue congestion.

### Anti-Pattern 4: Full-document overwrite on every run
- Creates large noisy diffs and loses manual edits.

## Scalability Considerations

| Concern | 100 packages/day | 10K packages/day | 1M packages/day |
|---------|------------------|------------------|-----------------|
| Workflow throughput | Single queue + modest workers | Split queues by stage, autoscale workers | Multi-namespace or sharded queues, strict quotas |
| Repo fetch cost | Basic cache sufficient | Strong snapshot/object cache required | Distributed object cache and aggressive partial clone |
| LLM cost/latency | Direct generation | Batch/concurrency control and section caching | Tiered generation strategies and strict budgets |
| Event history growth | Parent-only workflow fine | Monitor event counts, selective child split | Mandatory workflow partitioning and continue-as-new patterns |
| Publish contention | Low | Per-package path locking in output repo | Partitioned output branches and merge orchestration |

## Confidence and Assumptions

- **Temporal reliability semantics:** HIGH (official Temporal docs)
- **Git exact-revision and partial-clone strategy:** HIGH (official git docs)
- **Ubuntu-specific enrichment design choices:** MEDIUM (domain-driven architecture choice; verify against stakeholder doc style expectations)

Assumptions made:
- Output repository accepts automated commits from pipeline service identity.
- Package-level documentation ownership is isolated to per-package paths.
- Trigger payload provides a reachable repository and valid commit hash.

## Sources

- Temporal Workflows: https://docs.temporal.io/workflows
- Temporal Workflow Definition (determinism/versioning): https://docs.temporal.io/workflow-definition
- Temporal Activities and idempotency guidance: https://docs.temporal.io/activity-definition
- Temporal Retry Policies: https://docs.temporal.io/encyclopedia/retry-policies
- Temporal Task Queues: https://docs.temporal.io/task-queue
- Temporal Workers: https://docs.temporal.io/workers
- Temporal Child Workflows: https://docs.temporal.io/child-workflows
- Temporal TypeScript Activity Timeouts and retries: https://docs.temporal.io/develop/typescript/activities/timeouts
- Temporal Events and Event History limits: https://docs.temporal.io/workflow-execution/event
- git clone: https://git-scm.com/docs/git-clone
- git fetch: https://git-scm.com/docs/git-fetch
- git worktree: https://git-scm.com/docs/git-worktree
