# Research Summary: temporal-ai-docs-pipeline

**Synthesized:** 2026-05-15  
**Sources:** STACK.md, FEATURES.md, ARCHITECTURE.md, PITFALLS.md  
**Purpose:** Roadmap and requirements planning input

---

## Executive Summary

This project is a durable, AI-consumer-focused documentation generation pipeline for Ubuntu packages. The core loop is fixed: receive `(packageName, sourceRepo, gitHash)`, mine the exact source revision for authoritative Ubuntu-relevant signals, generate structured markdown docs with confidence scoring and provenance, and publish section-level patches to a separate output repository. The entire pipeline is orchestrated by Temporal to ensure durability, idempotency, and observable failure handling at production scale.

The recommended approach is **Go-first Temporal orchestration with a Python analysis/generation sidecar**. Go handles workflow code, activity workers for network/git/IO, and the control plane. Python handles deep repository parsing, content analysis, and LLM-adjacent generation steps where the ecosystem is stronger. This is not a pure-LLM pipeline: deterministic parsers and evidence-linked generation are non-negotiable because the primary consumers are AI models that will amplify any hallucinated instructions.

The principal risks are hallucinated documentation, stale source mismatch, brittle parsers, and unsafe command extraction — all of which are preventable with upfront design discipline rather than retroactive fixes. The architecture's natural build order maps well to a 5-phase v1 roadmap that ships a working end-to-end pipeline before adding reliability hardening.

---

## Recommended v1 Focus

Build a correct, safe, end-to-end pipeline for a single well-understood package type before optimising for scale or adding differentiating features.

**v1 outcome:** Given `(packageName, sourceRepo, gitHash)`, the pipeline reliably produces evidence-linked, structured Ubuntu docs with confidence scores, caveat injection, and provenance metadata, published via section-level patch to the output repo, with basic operational observability.

**Defer to v2+:**
- SLSA-style signed provenance attestation
- Claim-evidence graph with line-level links
- Ubuntu release matrix (LTS-targeted output variants)
- Confidence calibration against historical correction rates
- Dual rendering profiles (compact vs. rich) for different AI workloads
- Regression diff intelligence (explains why a section changed)

---

## Must-Have Table Stakes (v1)

| Feature | Why non-negotiable |
|---|---|
| Deterministic Temporal workflow with idempotent activities | Core reliability guarantee; duplicated publishes corrupt output repo history |
| Exact source revision pinning and fail-closed checkout verification | Without this, all provenance claims are untrustworthy |
| Ubuntu-aware multi-pass extraction (packaging metadata → scripts → README) | Extraction ordering is the primary hallucination defence |
| Section-level confidence scoring with structured caveat injection | Enables high throughput without silently degrading quality |
| Run-level and claim-level provenance metadata in output | AI consumers require auditable source traces; not optional for the stated consumer profile |
| Section-level patch publish (no full-document overwrite) | Full overwrite creates noisy diffs and destroys manual edits |
| Failure taxonomy (retryable vs. non-retryable per activity class) | Prevents infinite retries and surfaces permanent failures fast |
| Command safety policy engine (denylist/allowlist + review tiers) | Critical for Ubuntu docs; a single unsafe `curl \| sh` in generated output is a trust-ending event |
| Temporal Search Attributes for run status and package/hash indexing | Required for operational debugging at any non-trivial scale |
| Preflight auth validation before expensive work begins | Repo auth failures are high-likelihood; late detection wastes significant compute |

---

## Key Architecture Choices That Influence Phase Ordering

These decisions are load-bearing for roadmap sequencing — get them wrong early and they require rework across multiple phases.

**1. Workflow ID = `docs/{packageName}/{gitHash}`**  
This is the idempotency key. Start policy must reject duplicate starts by default. Re-run mode uses signals/updates, never a new random ID. This must be in place before any other workflow code.

**2. Activity-first decomposition, not giant composite activities**  
Each stage (fetch, extract, enrich, plan, generate, merge, validate, publish) is a separate activity with its own retry class. This is required for targeted retries and replay safety. Monolithic activities are an explicit anti-pattern.

**3. Four isolated task queues**  
`orchestration-queue`, `repo-io-queue`, `generation-queue`, `validation-queue`. LLM or git slowness must not starve workflow orchestration. Queue split must be architected in Phase 1 even if all queues start with one worker each.

**4. Content-addressed artifact cache (three tiers)**  
- Git snapshot cache keyed by `(repoURL, gitHash)`
- Extraction artifact cache keyed by snapshot key + extractor version
- Generation cache (optional) keyed by plan + context + prompt version  
Establishes the idempotent retry foundation. Implement tiers incrementally but design the key schema in Phase 1.

**5. Section ownership via HTML comment markers**  
`<!-- GENERATED:INSTALL START/END -->` etc. enables safe partial updates. Must be in the output schema from day one — retrofitting markers into existing published docs is painful.

**6. Run Ledger outside Temporal history**  
Idempotency tokens, artifact fingerprints, and publish references live in a separate store (PostgreSQL recommended). Temporal history is not the right place for cross-run deduplication state.

---

## Top Risks and Mitigations

| Risk | Severity | Likelihood | Mitigation |
|---|---|---|---|
| Hallucinated documentation | Critical | Medium | Enforce evidence-linked generation: every actionable claim must reference a mined source file/line or be explicitly caveated. Block publication of critical sections with zero evidence. |
| Stale source mismatch | Critical | Medium | Fail closed if checkout does not verify to the exact requested hash. Record and compare hash in both workflow input and provenance manifest. |
| Unsafe command extraction | Critical | Medium | Implement command policy tiers (allowed / needs-review / blocked) before generation is wired up. Static denylist at minimum; do not ship without it. |
| Brittle parsers | High | High | Build layered extraction: strict schema parse → tolerant parse → heuristic fallback. Maintain a fixture corpus of representative repo structures from day one. Treat extraction coverage as a first-class SLO metric. |
| Repo authentication failures | High | High | Add preflight auth validation (read scope for source, write scope for output) before the expensive pipeline begins. Classify 401/403 as non-retryable, 429 as retryable with backoff. Use token rotation monitors. |
| Markdown drift | High | Medium | Use AST-aware section patching (not line-based diffing). Pin markdown formatter version. Add round-trip idempotence regression tests from Phase 4 onward. |
| Scale bottlenecks | High | Medium | Partition queues by stage and repo size class. Cache immutable mining artifacts. Apply concurrency budgets. (Address in Phase 5; design for it in Phase 2.) |

---

## Implementation Guardrails

These are hard constraints, not preferences. Violating them creates issues that are difficult to fix post-facto.

1. **No git/LLM calls inside Workflow code.** All external IO lives in Activities. Violation breaks replay safety and causes non-deterministic workflow failures.

2. **No full-document overwrite on publish.** Section-level patch is mandatory. Full replacement only as explicit emergency fallback with a flag set in run metadata.

3. **No non-pinned source extraction.** Branch-only checkout breaks reproducibility. Only `gitHash`-pinned checkout is acceptable.

4. **No secrets or PII in Temporal Search Attributes or trace metadata.** These fields are commonly plaintext and indexed. Redact before emitting.

5. **No hard-failing on low-confidence sections.** Low confidence publishes with machine-readable caveats. Hard failure is reserved for the safety floor (zero-evidence critical commands) and input contract violations.

6. **No unbounded retries.** Every activity class has an explicit `scheduleToClose` budget. Infinite retry loops are an operational incident waiting to happen.

7. **No LLM-only extraction path.** LLM generation is always downstream of deterministic parser output. The parser produces the evidence; the LLM drafts text grounded in that evidence.

8. **Command safety policy must be active before any generated content reaches the output repo.** Not a Phase 3 nice-to-have — it is a precondition for the publish path being usable at all.

---

## Suggested Phase Decomposition (High-Level)

Phase order is derived from the architecture's component dependency chain and the pitfall phase-warning map.

### Phase 1 — Foundation: Input Contract, Provenance, and Auth
*Build before everything else — later phases depend on these primitives.*

Deliverables:
- Temporal namespace, task queues (all four), and worker bootstrapping
- Trigger ingress API with schema validation
- `GeneratePackageDocsWorkflow` skeleton with deterministic workflow ID (`docs/{pkg}/{hash}`) and stubbed activities
- Exact-hash source fetch with fail-closed checkout verification
- Run ledger schema and idempotency token design
- Preflight auth validation for source read and output write scopes
- Search Attributes schema (package, hash, confidence bucket, publish result)
- Section ownership marker schema defined in output contract

Pitfalls addressed: stale source mismatch, repo auth failures

---

### Phase 2 — Repository Mining and Extraction
*Provides the evidence foundation for all generation quality.*

Deliverables:
- Layered extraction: `snapcraft.yaml` / `debian/control` → build/install scripts → README/docs
- Ubuntu context enrichment (APT, systemd, snap paths, service management conventions)
- Structured context artifact with source pointers (file path, line span, commit hash)
- Per-file extraction coverage metrics and extraction SLO baseline
- Fixture corpus for representative repo shapes (monorepo, polyglot, minimal)
- Git snapshot cache (tier 1) and extraction artifact cache (tier 2)

Pitfalls addressed: brittle parsers, early scale pressure from expensive parsing

---

### Phase 3 — Generation, Safety, and Confidence
*Wire extraction output to generation; enforce safety gates before any publish path opens.*

Deliverables:
- Claim taxonomy: `verified` / `inferred` / `speculative`
- Evidence-linked generation (each actionable claim references mined source or is explicitly caveated)
- Section-level confidence scoring (evidence density + source authority weighting + conflict detection + coverage completeness)
- Structured caveat renderer (machine-parsable, section-scoped, reason-coded)
- Command safety policy engine: denylist/allowlist, `blocked` / `needs-review` / `allowed` tiers
- Publication gate: block critical sections with zero evidence or blocked commands
- Run-level and claim-level provenance metadata schema in output

Pitfalls addressed: hallucinated documentation, unsafe command extraction

---

### Phase 4 — Output Rendering, Patching, and Publishing
*Section-level patch, minimal diffs, and clean publish to the output repo.*

Deliverables:
- AST-aware section patch engine (marker-based, full-rewrite fallback when markers absent)
- Deterministic markdown formatting with pinned formatter version
- Validation activity: required sections, anchor integrity, provenance presence, blocked-command gate
- Publish activity: branch → commit → push to `packages/{packageName}/` in output repo
- Publish idempotency: no-op if target tree is identical; content-addressed commit keys
- Commit message convention carrying `(package, gitHash, workflowRunId, confidence bucket)`
- Round-trip idempotence regression tests for markdown stability

Pitfalls addressed: markdown drift, provenance visibility gaps

---

### Phase 5 — Reliability Hardening and Observability
*Production-grade operational posture.*

Deliverables:
- Per-activity retry class tuning (Class A: infra IO, Class B: deterministic transforms, Class C: LLM generation)
- Non-retryable error classification finalized across all activities
- `scheduleToClose` budgets set and tested for each activity
- Queue partitioning by repo size class (small/medium/large)
- Concurrency budgets and backpressure controls
- OpenTelemetry → Prometheus/Grafana dashboards with pipeline SLOs: success rate, p95 latency, retry depth, low-confidence ratio, blocked-command ratio
- Token rotation monitoring and expiry alerts
- Continue-As-New patterns for workflows at history budget risk
- Load and performance baseline established

Pitfalls addressed: scale bottlenecks, recurring auth/security regressions

---

## Stack Recommendations

| Layer | Choice | Version | Notes |
|---|---|---|---|
| Orchestration language | Go | 1.26.x | Temporal Go SDK maturity; static binaries; N or N-1 release line |
| Analysis/generation language | Python | 3.13+ | Best ecosystem for parsing, NLP, LLM integration |
| Temporal SDK (primary) | `go.temporal.io/sdk` | v1.43.x | Go SDK as workflow and activity orchestration layer |
| Temporal SDK (secondary) | `temporalio` | 1.27.x | Python activities for deep analysis workers only |
| Temporal deployment | Temporal Cloud | — | Operational maturity; use PostgreSQL visibility backend |
| YAML parsing (Go) | `gopkg.in/yaml.v3` | latest stable | Strict snapcraft.yaml / debian/control extraction |
| YAML parsing (Python) | `ruamel.yaml` | latest stable | Comment-preserving fallback when layout matters |
| Fast text scan | `ripgrep` | 14.x | Evidence extraction across repo trees |
| Markdown parsing/patching | `markdown-it-py` or `mistune` | latest stable | Python-side AST-aware section patching |
| Observability | OpenTelemetry → Prometheus + Grafana | — | Trace correlation, workflow metrics, SLO dashboards |
| Delivery | GitHub Actions + OIDC | — | No long-lived credentials; workload identity for repo access |
| Visibility / run ledger | PostgreSQL | 16.x | Temporal visibility backend + run ledger store |

---

## Confidence Assessment

| Area | Confidence | Basis |
|---|---|---|
| Stack choices | HIGH | Official Temporal, Go, and Python SDK docs; well-established production patterns |
| Feature scope and v1 boundaries | HIGH | Grounded in Temporal behavior specs and Ubuntu packaging semantics |
| Architecture and component ordering | HIGH | Official Temporal determinism/retry docs; confirmed by pitfall alignment |
| Pitfall identification | HIGH for provenance, auth, parsers, scale; MEDIUM for exact mitigation ROI | Common in orchestration-heavy mining pipelines; ROI depends on real repo diversity |
| Ubuntu-specific extraction ordering | MEDIUM-HIGH | Domain-driven design; verify against actual snapcraft/debian corpus early |

**Gaps to address in requirements:**
- Confirm output repository ownership model (automated commits only vs. PR-gated; who reviews?)
- Clarify LLM provider constraints (self-hosted vs. cloud; latency and cost budget per doc)
- Define the confidence floor threshold below which a critical section blocks publication
- Confirm whether submodule and LFS resolution is required for target packages
- Establish the extraction fixture corpus sources (which Ubuntu packages are the baseline test set?)
