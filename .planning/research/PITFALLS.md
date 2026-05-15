# Domain Pitfalls: Automated Documentation Generation with Temporal + Repository Mining

**Domain:** Ubuntu-focused automated docs from source repos
**Researched:** 2026-05-15
**Overall confidence:** MEDIUM

## Severity and Likelihood Scale

- **Severity:** Critical, High, Medium, Low
- **Likelihood:** High, Medium, Low

## Risk Matrix (At a Glance)

| Pitfall | Typical Failure Mode | Severity | Likelihood | Early Warning Signs | Suggested Phase to Address |
|---|---|---|---|---|---|
| Hallucinated docs | Generated steps or flags not present in source, or Ubuntu instructions fabricated from prior model bias | Critical | Medium | Increasing reviewer comments like "cannot reproduce"; generated commands not found in repo; high caveat rate without evidence links | Phase 3: Generation Guardrails + Validation |
| Stale source mismatch | Pipeline analyzes a different repo state than provided git hash, causing traceability break | Critical | Medium | Output metadata hash differs from requested hash; non-deterministic reruns; diff churn with identical input | Phase 1: Input Contract + Provenance Locking |
| Brittle parsers | Parser breaks on monorepos, uncommon build systems, or non-standard docs layout | High | High | Sudden extraction drop to near-zero; spikes in "fallback to generic" docs; parser error retries dominate Temporal metrics | Phase 2: Ingestion and Extraction Resilience |
| Markdown drift | Repeated update cycles degrade structure, links, and section intent over time | High | Medium | Section IDs duplicate; broken anchors increase; generated docs reorder unrelated content each run | Phase 4: Render, Diff, and Publish Controls |
| Repo auth failures | Token expiry, permission scope mismatch, or rate-limit failures block clone/fetch/push | High | High | Temporal activities retrying auth/network steps; increased 401/403/429 rates; backlog growth of pending packages | Phase 1: Access Reliability Baseline |
| Scale bottlenecks | Throughput collapse due to large repos, expensive parsing, or unbounded activity fanout | High | Medium | Workflow latency climbs with queue depth; worker CPU saturation; frequent activity heartbeats timing out | Phase 5: Capacity and Performance Hardening |
| Unsafe command extraction patterns | Model emits unsafe shell commands from README/scripts that could damage systems or leak data | Critical | Medium | Commands include curl|sh, sudo without scope, rm/chown on broad paths, dynamic eval patterns | Phase 3: Safety Policy + Command Sanitization |

## Critical Pitfalls

### 1) Hallucinated Documentation

**What goes wrong:**
The generation layer produces plausible Ubuntu install/config instructions that are unsupported by repository evidence (wrong package names, non-existent CLI flags, fabricated config paths).

**Why it happens:**
- Retrieval context is thin, noisy, or missing key files.
- Prompting favors completion over evidence citation.
- No hard requirement to attach claims to mined artifacts.

**Consequences:**
- Users run invalid instructions and lose trust in generated docs.
- Downstream AI consumers amplify incorrect guidance.
- Review cost rises because every command requires manual verification.

**Early warning signs:**
- Rising "cannot verify" labels in QA or post-run checks.
- Commands in output not found via grep in source tree.
- Caveats are present but not tied to concrete uncertainty reasons.

**Prevention / mitigation:**
- Enforce evidence-linked generation: each actionable step must reference a mined file/line or explicit external Ubuntu source.
- Introduce claim classes: `verified`, `inferred`, `speculative`; require caveats for non-verified classes.
- Add executable lint checks for command existence and option validity where feasible.
- Block publication when critical command claims have zero evidence.

**Suggested phase:**
- Primary: Phase 3 (Generation Guardrails + Validation)
- Supporting: Phase 2 (better retrieval quality)

**Ratings:** Severity = Critical, Likelihood = Medium

---

### 2) Stale Source Mismatch

**What goes wrong:**
The pipeline receives an exact git hash but mines a different state (branch tip, shallow fetch mismatch, or wrong submodule revision), so produced docs are not traceable to declared input.

**Why it happens:**
- Clone/fetch logic defaults to branch refs.
- Submodules/LFS artifacts are unresolved or inconsistent.
- Hash propagated in workflow metadata but not enforced in activities.

**Consequences:**
- Non-reproducible docs and audit failures.
- Conflicting docs for same package/hash pair.
- Hard-to-debug diffs when rerunning same input.

**Early warning signs:**
- Re-run with same input yields different output checksum.
- Commit hash in output metadata differs from workflow request.
- Unexpected file content relative to requested revision.

**Prevention / mitigation:**
- Make `repo_url + commit_hash` an immutable workflow identity key.
- Fail closed if checkout does not resolve exactly to requested hash.
- Record provenance manifest in output (repo URL, resolved hash, timestamp, tool versions).
- Pin submodule and dependency resolution strategy.

**Suggested phase:**
- Primary: Phase 1 (Input Contract + Provenance Locking)
- Supporting: Phase 4 (publish provenance alongside docs)

**Ratings:** Severity = Critical, Likelihood = Medium

---

### 3) Unsafe Command Extraction Patterns

**What goes wrong:**
Generated docs include dangerous commands copied or inferred from source materials (for example, broad file deletion, remote script execution, privilege escalation without constraints).

**Why it happens:**
- Extraction logic treats all command snippets as equally safe.
- No policy engine for deny/allow command patterns.
- Ubuntu context is interpreted too loosely (system-wide commands without least privilege guidance).

**Consequences:**
- Potential user system damage or security exposure.
- Trust and adoption loss for generated docs.
- Compliance and governance risk.

**Early warning signs:**
- Commands matching dangerous regexes: `curl .*\|\s*sh`, `rm -rf /`, wildcard ownership changes, direct `eval`.
- Sudden increase in `sudo` usage in generated outputs.
- Security review rejects outputs frequently.

**Prevention / mitigation:**
- Implement command policy tiers: `allowed`, `needs-review`, `blocked`.
- Require contextual guards for privileged commands (path scope, backup notes, rollback guidance).
- Sandbox command validation with static checks and denylist/allowlist matching.
- Emit safer alternatives by default (apt packages over piping remote scripts).

**Suggested phase:**
- Primary: Phase 3 (Safety Policy + Sanitization)
- Supporting: Phase 5 (security hardening and continuous policy tuning)

**Ratings:** Severity = Critical, Likelihood = Medium

## High-Impact Pitfalls

### 4) Brittle Parsers

**What goes wrong:**
Repository mining fails or partially fails on monorepos, polyglot builds, generated files, unusual docs layouts, or very large trees.

**Why it happens:**
- Parsers assume fixed directory conventions.
- Too much logic in regex-only extraction.
- Lack of layered fallbacks (AST, manifest, heuristic, semantic).

**Consequences:**
- Missing install/config signals.
- Generic output with low utility.
- Frequent retries and higher workflow cost.

**Early warning signs:**
- Extraction coverage metrics trend downward.
- Frequent parser exceptions clustered by repo type.
- Increased share of docs tagged "low confidence".

**Prevention / mitigation:**
- Build parser stack with progressive fallback strategy.
- Add per-repo capability profiles (monorepo, language ecosystem, manifest-first).
- Maintain fixture corpus of representative repo structures.
- Make extraction quality a first-class metric with SLO thresholds.

**Suggested phase:**
- Primary: Phase 2 (Ingestion and Extraction Resilience)
- Supporting: Phase 5 (performance optimization under large repos)

**Ratings:** Severity = High, Likelihood = High

---

### 5) Markdown Drift

**What goes wrong:**
Update runs gradually degrade markdown quality and semantic structure: duplicated sections, broken anchors, unstable ordering, and noisy diffs.

**Why it happens:**
- Regeneration does not preserve stable section identity.
- Patch strategy operates on brittle textual anchors.
- Renderer and formatter versions differ across runs.

**Consequences:**
- Review fatigue due to noisy diffs.
- Broken downstream consumption by AI tooling.
- Eventually forces full document rewrites.

**Early warning signs:**
- Anchor/link break rate increases each run.
- Minimal source changes yield large markdown diffs.
- Section ordering flips between runs.

**Prevention / mitigation:**
- Use semantic section IDs and stable template contracts.
- Apply AST-aware markdown diff/patch instead of plain line-based edits.
- Enforce deterministic rendering options and pinned formatter versions.
- Add regression tests for anchor stability and round-trip idempotence.

**Suggested phase:**
- Primary: Phase 4 (Render, Diff, and Publish Controls)
- Supporting: Phase 3 (structured generation schema)

**Ratings:** Severity = High, Likelihood = Medium

---

### 6) Repository Authentication Failures

**What goes wrong:**
Clone/fetch/push operations intermittently fail due to token scope issues, expiration, organization policy changes, or API rate limits.

**Why it happens:**
- Credential lifecycle and scope are not centrally managed.
- Temporal retries are configured without auth-aware backoff and categorization.
- No preflight permission checks before expensive work.

**Consequences:**
- Backlogs and SLA misses.
- Partial workflow completion with orphaned artifacts.
- On-call load increases from repeated transient failures.

**Early warning signs:**
- Rising 401/403/429 error rates.
- Retries concentrated in SCM activities.
- Push failures only for specific org/repo combinations.

**Prevention / mitigation:**
- Add preflight auth validation for read/write scopes before mining.
- Classify retryable vs non-retryable auth errors in Temporal activities.
- Use token rotation monitors and expiry alerts.
- Introduce circuit breakers and dead-letter routing for persistent auth failures.

**Suggested phase:**
- Primary: Phase 1 (Access Reliability Baseline)
- Supporting: Phase 5 (operational hardening)

**Ratings:** Severity = High, Likelihood = High

---

### 7) Scale Bottlenecks in Temporal + Mining Pipeline

**What goes wrong:**
Throughput degrades as package volume and repo size increase: long queue times, worker saturation, memory pressure, and timeouts.

**Why it happens:**
- Unbounded activity fanout per workflow.
- Heavy parsing tasks executed synchronously without caching.
- Missing queue partitioning and priority controls.

**Consequences:**
- Missed processing windows.
- Higher cloud/compute cost.
- Increased failure cascades during peak ingestion.

**Early warning signs:**
- Queue depth grows faster than completion rate.
- p95/p99 workflow latency rises steadily.
- Worker heartbeat timeouts and OOM events increase.

**Prevention / mitigation:**
- Partition task queues by workload class (small/medium/large repo).
- Add concurrency budgets and backpressure controls.
- Cache immutable mining artifacts by `repo+hash`.
- Use benchmark-driven autoscaling and load-shedding for non-critical work.

**Suggested phase:**
- Primary: Phase 5 (Capacity and Performance Hardening)
- Supporting: Phase 2 (efficient extractors), Phase 4 (incremental publish)

**Ratings:** Severity = High, Likelihood = Medium

## Phase-Specific Warning Map

| Suggested Phase Topic | Likely Pitfall | Mitigation Focus |
|---|---|---|
| Phase 1: Input contracts, provenance, auth | Stale source mismatch, repo auth failures | Immutable input identity, fail-closed checkout, preflight permissions, retry taxonomy |
| Phase 2: Repository mining and parsing | Brittle parsers, early scale pressure | Layered extraction strategy, fixtures, coverage metrics, artifact caching |
| Phase 3: Generation and safety | Hallucinated docs, unsafe command extraction | Evidence-linked claims, confidence classes, command policy engine |
| Phase 4: Output rendering and publishing | Markdown drift, provenance visibility gaps | AST-aware patching, deterministic formatting, publish provenance manifests |
| Phase 5: Reliability and hardening | Scale bottlenecks, recurring auth/security regressions | Capacity modeling, queue partitioning, policy tuning, operational SLOs |

## Additional Cross-Cutting Controls

- Define quality gates before publish:
  - Provenance gate: output must include exact source hash metadata.
  - Safety gate: blocked command patterns must be absent.
  - Evidence gate: all actionable instructions must be evidence-linked or explicitly caveated.
  - Drift gate: markdown idempotence and link integrity checks must pass.
- Instrument Temporal with pipeline-level SLOs:
  - success rate, p95 latency, retry depth, low-confidence ratio, blocked-command ratio.
- Keep low-confidence behavior, but make caveats machine-readable:
  - include structured reason codes (`missing_source_signal`, `parser_uncertain`, `ambiguous_install_path`).

## Confidence Notes

- **High confidence:** Provenance mismatch risk, parser brittleness, auth failure operational impact, and scaling bottlenecks are common in orchestration-heavy repository pipelines.
- **Medium confidence:** Exact mitigation ROI depends on repository diversity, target package ecosystems, and your Temporal deployment profile.
- **Known unknowns:** Real failure distribution will vary until baseline telemetry is collected from production-like workloads.
