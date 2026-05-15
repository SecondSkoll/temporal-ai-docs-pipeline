# Feature Landscape

**Domain:** Temporal-based Ubuntu package documentation generation pipeline
**Project:** temporal-ai-docs-pipeline
**Researched:** 2026-05-15

## Scope and Assumptions

- Input contract is fixed: package name, source repository URL, source git hash.
- Output must be published to a separate documentation repository in per-package folders.
- Primary consumers are AI models running on Ubuntu systems.
- v1 should optimize for deterministic, reliable generation and high throughput under uncertainty.

## Table Stakes

Features users and downstream platform teams will expect in any serious Temporal-based docs pipeline.

| Feature | Why it matters | Complexity | Dependencies | MVP recommendation |
|---|---|---|---|---|
| Deterministic workflow orchestration with retries and timeouts | Temporal workflows must be replay-safe; extraction and publishing steps will fail intermittently and need durable retries without duplicate side effects. | Med | Temporal workflow/activity design, idempotent activity implementation, retry policy and timeout policy | v1 |
| Idempotent extraction and publish activities | Temporal activities may execute more than once; non-idempotent publish logic can create duplicate commits or inconsistent docs state. | Med | Stable run key (package + repo + hash), idempotency keys for publish operations, content hashing | v1 |
| Ubuntu-aware source extraction (packaging metadata first) | Ubuntu-focused docs are poor without parsing Debian/Ubuntu packaging files such as debian/control and related metadata. | Med | Parsers for debian/control and changelog patterns, repository checkout at exact git hash | v1 |
| Exact source revision pinning and immutable run context | Consumers must know which source snapshot produced a doc; this is required for trust, debugging, and rollback. | Low | Git fetch and checkout by commit hash, run manifest object | v1 |
| Structured output contract for AI consumption | AI consumers need predictable schema and headings, not only prose, for retrieval and tool use. | Med | Markdown template with machine-readable frontmatter/metadata block, stable section taxonomy | v1 |
| Confidence scoring and caveat injection | The project explicitly prioritizes throughput with caveats when confidence is low instead of hard-failing uncertain extraction. | Med | Confidence model (rule-based at minimum), evidence coverage metrics, caveat renderer | v1 |
| Provenance metadata for every generated claim block | AI consumers need to trace statements back to source files/lines and commit hash; otherwise generated guidance is not auditable. | High | Evidence collector, source snippet index, claim-to-evidence mapping, provenance schema | v1 |
| Section-level update/patch mode (minimal diffs) | Replacing full files creates noisy history and makes review/attribution harder; section patching preserves longitudinal clarity. | High | Stable section IDs, semantic or anchor-based patcher, content fingerprinting | v1 |
| Output repository publishing with branch + PR or controlled direct commit policy | Separate repo publishing is core to topology; CI/consumers need a predictable publication path and validation gate. | Med | Git credentials, branch naming strategy, commit message convention, optional PR automation | v1 |
| Run observability and searchable execution metadata | Operators need to find failed or degraded runs quickly; Temporal visibility/search attributes are standard operational expectation. | Med | Temporal Search Attributes, run status model, logging and metrics | v1 |
| Failure taxonomy (retryable vs non-retryable) | Permanent extraction errors should surface quickly; transient errors should auto-retry to protect throughput. | Med | Typed errors from activities, retry policy mapping, dead-letter/report channel | v1 |
| Security and privacy guardrails for metadata | Temporal visibility fields are unencrypted and trace metadata can leak sensitive values unless constrained. | Low | Redaction policy, denylist for tokens/secrets/PII, lint checks on emitted metadata | v1 |

## Differentiators

Features that are not always present in basic pipelines, but materially improve quality, trust, or downstream AI utility.

| Feature | Why it matters | Complexity | Dependencies | MVP recommendation |
|---|---|---|---|---|
| Claim-evidence graph with line-level provenance links | Moves beyond "generated text" into auditable knowledge artifacts; enables selective revalidation when source changes. | High | Provenance schema, evidence index, stable node IDs, link generator | v2 |
| SLSA-style provenance attestation for doc artifacts | Makes generated docs verifiable supply-chain artifacts (builder, parameters, resolved dependencies, invocation metadata). | High | Attestation format, signer, artifact digesting, verifier tooling | v2 |
| Multi-pass extraction strategy (packaging-first, then code/config, then README/docs) | Reduces hallucinated instructions by prioritizing operationally authoritative files before narrative docs. | Med | Ranked extractor pipeline, source-type classifiers, fallback ordering | v1 |
| Ubuntu release targeting matrix (LTS-focused output variants) | AI consumers on Ubuntu benefit from explicit release-specific install/config differences and compatibility notes. | High | Ubuntu release metadata model, package manager behavior rules, compatibility extractor | v2 |
| Confidence-aware publication gates | High-confidence docs can auto-publish; low-confidence docs can publish with stronger caveats or require review labels. | Med | Confidence thresholds, policy engine, branch protection/approval rules | v1 |
| Regression diff intelligence for docs updates | Explains why a section changed (source file/line deltas), improving trust and reducing manual triage effort. | High | Previous-run provenance store, structural diff engine, change summarizer | v2 |
| Temporal-native remediation workflow (auto re-run by failure class) | Improves reliability by launching targeted child workflows for recoverable failures instead of rerunning the full pipeline blindly. | Med | Child workflows, failure classifier, backoff policy, retry budget controls | v2 |
| AI-consumer optimization bundle (compact + expanded renderings) | Some model workloads need terse retrieval-friendly docs; others need rich context. Dual renderings improve downstream performance. | Med | Dual template system, section compaction rules, publish artifact set | v2 |
| Extraction confidence calibration against historical outcomes | Prevents confidence inflation by comparing previous confidence scores to observed correction rates over time. | High | Outcome feedback capture, calibration model, evaluation dataset | v2 |
| Cross-package dependency context synthesis | Provides package-level operational context (what else must be installed/configured) from dependency metadata. | Med | Dependency parser from packaging metadata, context generation rules | v2 |

## Anti-Features

Capabilities that should be explicitly avoided because they reduce reliability, violate scope, or slow delivery.

| Anti-Feature | Why avoid | Complexity | Dependencies | MVP recommendation |
|---|---|---|---|---|
| End-to-end LLM-only extraction without deterministic parsers | Increases hallucination risk and weakens provenance traceability for Ubuntu/package metadata. | Low | N/A | out |
| Full-file overwrite on every publish | Creates noisy diffs, obscures true changes, and weakens reviewability in output repository history. | Low | N/A | out |
| Hard fail on any low-confidence section | Conflicts with stated throughput goal; should publish with caveats unless confidence drops below safety floor. | Low | N/A | out |
| Storing secrets/PII in Temporal Search Attributes or trace metadata | Visibility and tracing metadata are commonly plaintext/indexed; this creates avoidable data leakage risk. | Low | N/A | out |
| Mixing non-Ubuntu platform guidance in core v1 content | Dilutes focus and increases ambiguity for Ubuntu AI consumers. | Low | N/A | out |
| Non-pinned source extraction (branch-only) | Breaks reproducibility and provenance guarantees. | Low | N/A | out |
| Manual-only publishing flow as default path | Defeats the purpose of Temporal orchestration and limits scale. | Low | N/A | out |

## Behavior Expectations

The following expectations define product behavior for the five areas you called out.

### 1. Extraction Behavior Expectations

- The pipeline MUST check out the exact provided git hash before extraction.
- Extraction MUST prioritize authoritative operational sources in this order:
  1. Packaging metadata (`debian/control`, changelog, package relationship metadata).
  2. Build/install/config scripts and manifests.
  3. Maintainer docs (`README`, `docs/`, examples).
- Each extracted fact SHOULD carry a source pointer (`file path`, optional line span, commit hash).
- Extraction SHOULD classify facts into: installation, configuration, runtime behavior, limitations, and Ubuntu-specific notes.
- If authoritative files are missing, the run SHOULD degrade gracefully to secondary sources and emit explicit caveats.

### 2. Confidence Handling Expectations

- Confidence MUST be computed at section level (not only document level).
- Minimum confidence inputs SHOULD include:
  - Evidence density (number and quality of source-backed facts).
  - Source authority weighting (packaging files > scripts > narrative docs).
  - Conflict detection (contradictory statements lower confidence).
  - Coverage completeness against required section checklist.
- Publication policy:
  - High/medium confidence sections publish normally.
  - Low confidence sections publish with caveat blocks and provenance detail.
  - Critical confidence floor breaches (for safety-impacting instructions) should fail that section and mark run degraded.
- Confidence scores SHOULD be stable across re-runs for identical inputs.

### 3. Caveat Behavior Expectations

- Caveats MUST be explicit, machine-parsable, and tied to specific section IDs.
- Caveat categories SHOULD include: missing evidence, conflicting evidence, inferred behavior, and Ubuntu-version ambiguity.
- Caveats MUST include a "what would improve confidence" hint (for example, missing file type or unresolved dependency context).
- Caveats SHOULD never be generic boilerplate; each caveat must name concrete missing or conflicting evidence.

### 4. Provenance Traceability Expectations

- Every published document MUST include run-level provenance:
  - package name, source repo URL, source git hash
  - workflow ID/run ID
  - generation timestamp
  - generator version
- Every major claim block SHOULD include claim-level provenance:
  - source file path(s)
  - line span(s) or symbol reference where possible
  - extraction method identifier
  - confidence score
- Provenance metadata SHOULD be represented in a structured block (YAML frontmatter or adjacent JSON) so AI consumers can parse it reliably.
- Provenance chain SHOULD remain stable across section patch updates; changed sections must update their claim provenance only.

### 5. Output Publishing Expectations

- Publishing MUST target the separate docs repository and package-scoped folder path.
- The publisher MUST enforce deterministic file naming and section IDs to enable minimal-diff patching.
- Publish workflow SHOULD include:
  1. Generate candidate docs.
  2. Diff against existing docs.
  3. Apply section-level patch where possible.
  4. Commit with machine-readable commit metadata (package/hash/run ID).
  5. Push via configured policy (direct commit or PR).
- A failed publish step SHOULD be retryable without creating duplicate commits.
- Publish result MUST return artifact references (commit SHA, changed files, and run status summary).

## Feature Dependencies (High-Level)

- Deterministic orchestration + idempotent activities -> reliable extraction and publish.
- Source pinning + provenance schema -> claim traceability and reproducibility.
- Confidence model + caveat renderer -> throughput under uncertainty without silent quality regressions.
- Stable section IDs + patch engine -> minimal diffs and maintainable docs history.
- Search attributes + logs/metrics -> operational debugging and SLA management.

## MVP Recommendation

Prioritize for v1:

1. Deterministic Temporal workflow with idempotent activities.
2. Ubuntu-aware extraction with source-hash pinning.
3. Section-level confidence scoring with caveat injection.
4. Claim/run provenance metadata in output.
5. Section-level patch publish to separate docs repo.
6. Basic observability and failure taxonomy.

Defer to v2:

- SLSA-style attestations and signed provenance.
- Claim-evidence graph and regression diff intelligence.
- Ubuntu release-targeted variants.
- Calibration feedback loop for confidence.
- Dual rendering profiles for different AI workloads.

Keep out of scope:

- LLM-only extraction without deterministic evidence pipeline.
- Default non-Ubuntu breadth.
- Full-document replacement on each update.

## Source Notes

Primary references used:

- Temporal workflow execution, replay, and durability model: https://docs.temporal.io/workflow-execution
- Temporal activity idempotency and retry behavior: https://docs.temporal.io/activity-definition
- Temporal retry policy defaults and non-retryable failures: https://docs.temporal.io/encyclopedia/retry-policies
- Temporal search attributes and visibility constraints: https://docs.temporal.io/search-attribute
- Debian control fields and package metadata semantics (Ubuntu packaging baseline): https://www.debian.org/doc/debian-policy/ch-controlfields.html
- SLSA provenance schema and build attestation model: https://slsa.dev/spec/v1.0/provenance
- W3C trace context for interoperable request-level traceability metadata: https://www.w3.org/TR/trace-context/

Confidence assessment for this report:

- Temporal behavior expectations: HIGH
- Ubuntu packaging extraction expectations: HIGH
- Provenance/traceability recommendations: MEDIUM-HIGH
- AI-consumer specific differentiators: MEDIUM (opinionated product guidance based on verified primitives)
