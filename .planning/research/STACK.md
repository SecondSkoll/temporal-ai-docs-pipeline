# 2026 Stack Recommendation: Temporal-Based Ubuntu Docs Pipeline

Project: temporal-ai-docs-pipeline  
Date: 2026-05-15  
Scope: package repository documentation generation from package name + source repo URL + source git hash, publishing per-package docs to a separate output repository.

## Executive Recommendation

Build this as a **Go-first Temporal worker platform** with a **Python analysis/generation sidecar path** where needed, backed by **Temporal Cloud**, **PostgreSQL visibility**, **OpenTelemetry -> Prometheus/Grafana**, and a **GitHub Actions + OIDC** delivery chain.

Why this stack for your exact use case:
- Temporal orchestration is strongest and most operationally mature in Go/TypeScript today, and Go SDK is currently very active and production-focused.
- Python remains best-in-class for repository/content analysis and LLM-adjacent text generation pipelines.
- Ubuntu-focused package docs require dependable extraction from `snapcraft.yaml` + repo artifacts, where strict parsing plus graceful fallbacks is better than "AI-only synthesis".
- Publishing generated docs to a separate output repo is easiest and safest with deterministic git CLI operations and atomic PR-based updates.

---

## Runtime Languages

### 1) Orchestration and long-running workers: Go 1.26.x  
Confidence: [HIGH]

Use Go for:
- Temporal workflow code
- Temporal activity workers for network/IO/git orchestration
- reliability-critical control plane logic (retries, idempotency keys, compensation)

Version guidance:
- Pin `go` to `1.26.x` patch stream.
- Keep workers on N-1 or N Go release line maximum per Go support policy.

Rationale:
- Strong Temporal Go SDK maturity and release velocity.
- Excellent operational profile for long-lived workers.
- Easy static binaries for Ubuntu deployment.

### 2) Repository/content analysis and generation workers: Python 3.13+ (prefer 3.13, evaluate 3.14)  
Confidence: [HIGH]

Use Python for:
- deep repo analysis pass
- markdown drafting/normalization
- confidence scoring and caveat synthesis
- optional model orchestration if you add LLM features later

Version guidance:
- Baseline: `Python 3.13` (bugfix phase, long support runway).
- Evaluate `3.14` only after dependency compatibility checks.

Rationale:
- Strong ecosystem for parsing, NLP/text tooling, and AI integrations.
- Temporal Python SDK is now mature enough for production activities where async/threaded execution model is respected.

---

## Temporal SDK Choice

### Recommended primary SDK: Temporal Go SDK (`go.temporal.io/sdk` v1.43.x line)
Confidence: [HIGH]

Use Go SDK as the primary orchestration language and keep Workflow definitions in Go.

Version guidance:
- Start on latest stable `v1.43.x` patch.
- Use explicit compatibility testing before minor bumps.

### Secondary SDK option (if needed): Temporal Python SDK (`temporalio` 1.27.x)
Confidence: [MEDIUM]

Use Python SDK for:
- specialized analysis workers and activity-heavy pipelines
- not as the first choice for the most central workflow graph unless your team is Python-first

Version guidance:
- Pin `temporalio` to `1.27.x` patch updates.

### TypeScript SDK (`@temporalio/*` 1.17.x)
Confidence: [MEDIUM]

TypeScript is viable and strong, but for this Ubuntu/package-analysis-heavy domain, Go+Python usually gives cleaner operational and analysis ergonomics than TypeScript+Python.

---

## Queue and Workflow Architecture

### Task queue topology
Confidence: [HIGH]

Use dedicated queues by concern:
- `docs.intake` (validation + dedupe)
- `docs.repo-analyze` (clone, extract metadata, parse manifests)
- `docs.generate` (markdown generation, confidence scoring)
- `docs.patch` (section-aware patching)
- `docs.publish` (git branch/commit/pr)
- `docs.verify` (policy + lint + smoke validation)

Design rules:
- Workers on one queue must register identical handlers unless using Worker Versioning rollout semantics.
- Isolate hot paths (`generate`, `publish`) to scale independently.

### Workflow graph (recommended)
Confidence: [HIGH]

1. IntakeWorkflow
- Input contract validation: package name, source URL, source git hash
- idempotency key: `${package}:${source_hash}`
- dedupe check via visibility/search attrs

2. SourceSnapshotWorkflow
- shallow clone/fetch exact hash
- provenance capture (commit, tree hash, timestamp)

3. RepoAnalysisWorkflow (child)
- detect `snapcraft.yaml`
- extract top-level keys and app/part sections
- gather install/config evidence from README, docs, scripts, systemd/service files
- output structured evidence graph + confidence values

4. DocsGenerationWorkflow (child)
- build canonical doc sections:
  - install on Ubuntu
  - initial configuration
  - service/process behavior (if daemonized)
  - caveats/unknowns
- always emit output, even on low-confidence extraction

5. PatchOrReplaceWorkflow
- parse existing generated doc in output repo
- replace only generated section blocks (marker-based + AST safety checks)
- full rewrite fallback when markers/AST mismatch

6. PublishWorkflow
- branch, commit, push, PR (or direct push for bot-managed branch)
- attach provenance and confidence metadata

7. VerifyWorkflow
- markdown lint, link check, schema check, policy check
- finalize status + metrics

### Temporal safety features to use by default
Confidence: [HIGH]

- Worker Versioning for safe workflow code evolution.
- Activity retries with bounded backoff and non-retryable classification.
- Continue-As-New for very large histories.
- Signals/Updates for human override (e.g., force publish, skip patch mode).
- Search Attributes: package, source hash, confidence bucket, publish result.

---

## Repository Analysis Tooling

### Core toolchain
Confidence: [HIGH]

- `git` CLI (2.4x+): clone/fetch/rev-parse/diff/apply.
- YAML parsing: `gopkg.in/yaml.v3` (Go) and `ruamel.yaml` (Python fallback when preserving comments/layout matters).
- Fast text scan: `ripgrep` for evidence extraction.
- Structural code parsing where beneficial: `tree-sitter` (optional, selective use).

### Snapcraft metadata extraction strategy
Confidence: [HIGH]

Primary extraction sources:
- `snapcraft.yaml` top-level keys and sections (e.g. name/version/summary/description/base/grade/confinement/apps/parts/plugs/slots/platforms or architectures depending on base generation).
- Snapcraft model semantics from Canonical tooling/docs references.

Behavior:
- strict parse first (schema-constrained)
- tolerant parse second (best effort)
- unresolved fields -> explicit caveat section, never silent omission

---

## Markdown Generation and Patch Strategy

### Generation format
Confidence: [HIGH]

Use deterministic, sectioned markdown with immutable markers:

- `<!-- GENERATED:INSTALL START -->`
- `<!-- GENERATED:INSTALL END -->`
- `<!-- GENERATED:CONFIG START -->`
- `<!-- GENERATED:CONFIG END -->`
- `<!-- GENERATED:CAVEATS START -->`
- `<!-- GENERATED:CAVEATS END -->`

This enables safe partial updates and human-authored content preservation.

### Markdown tooling
Confidence: [MEDIUM]

Primary recommendation:
- Python pipeline with `markdown-it-py` or `mistune` + custom section rewriter.

Alternative:
- TypeScript `unified/remark` if you later converge more generation in Node.

### Low-confidence handling policy
Confidence: [HIGH]

Do not block generation on low confidence.
Always produce docs with:
- confidence score per section
- source trace references (file/path/line where available)
- caveat callouts: "could not verify", "inferred from pattern", "manual validation recommended"

---

## Git Interaction Strategy (Output Repository)

### Recommended approach
Confidence: [HIGH]

Use **git CLI**, not API-only content writes.

Flow:
- clone output repo once per worker execution context (or ephemeral clone per run)
- checkout/update target branch
- write package path: `packages/<package-name>/...`
- run deterministic formatter/lint
- commit with structured message and metadata trailer
- push branch + open PR via GitHub API

Commit message convention:
- `docs(<package>): update generated ubuntu install/config for <source-hash-short>`

Idempotency and conflict strategy:
- rebase/merge retry loop bounded (e.g. 3 attempts)
- if conflict in generated sections: regenerate against latest base and retry
- if still conflicting: raise review-required status

---

## CI/CD and Deployment Recommendations

### CI platform
Confidence: [HIGH]

Use GitHub Actions for:
- build/test/lint
- container build/sign/scan
- deployment automation

Pin major actions versions:
- `actions/checkout@v6`
- `actions/setup-go@v5` (or latest stable major)
- `actions/setup-python@v5` (or latest stable major)

### Auth and secrets
Confidence: [HIGH]

Use GitHub Actions OIDC federation for cloud credentials.
Avoid long-lived cloud secrets.
Use repo/environment protection rules for deploy workflows.

### Deployment model
Confidence: [MEDIUM]

Preferred:
- Temporal Cloud + Kubernetes worker deployments (Ubuntu-based images)

Alternative:
- Temporal self-hosted only if compliance/data residency requires it and SRE capacity exists.

Release strategy:
- blue/green or canary worker rollout with Worker Versioning
- automatic rollback on SLO breach

---

## Observability Stack

### Recommended stack
Confidence: [HIGH]

- OpenTelemetry SDK instrumentation (Go + Python workers)
- OpenTelemetry Collector
- Prometheus for metrics storage/scrape
- Grafana dashboards + alerting
- Optional: Datadog if organization standard requires it

Temporal-specific telemetry:
- SDK + server metrics enabled
- queue depth / poller health / activity retry rates
- workflow latency percentiles by workflow type
- publish success/failure ratio

Golden signals to alert on:
- publish failure rate > threshold
- repeated patch conflicts
- low-confidence ratio spikes
- stuck workflows / growing backlog / poller starvation

---

## Security Controls

### Supply chain and artifact integrity
Confidence: [HIGH]

- Generate SBOM (CycloneDX or SPDX) per build artifact.
- Sign artifacts with Sigstore Cosign (keyless preferred via OIDC).
- Target SLSA level progression (start with provenance + hardened build, then raise level).

### Code and dependency scanning
Confidence: [HIGH]

- Trivy in CI for vuln/misconfig/SBOM scan.
- Gitleaks for secret scanning on PR + main.
- Dependabot/Renovate for automated updates.

### Runtime and access controls
Confidence: [HIGH]

- Least-privilege IAM for repo write bot and cloud infra.
- Separate credentials for source read vs output repo write.
- Enforce `safe.directory` and no untrusted hook execution in automation clones.
- Branch protection + required checks on output repo.

### Data handling
Confidence: [MEDIUM]

- Encrypt at rest and in transit.
- Redact secrets/tokens from logs/traces.
- Store only required provenance and confidence metadata.

---

## Do Not Use

### 1) Single monolithic queue for all tasks
Why not:
- difficult scaling and noisy-neighbor effects
- less control over retries and worker sizing
Confidence: [HIGH]

### 2) Full-file overwrite-only doc updates
Why not:
- destroys manual improvements and causes noisy diffs
- increases merge conflicts in output repo
Confidence: [HIGH]

### 3) LLM-only extraction with no deterministic parser path
Why not:
- hallucination risk for install/config instructions
- poor traceability to source files
Confidence: [HIGH]

### 4) Long-lived cloud secrets in GitHub Actions
Why not:
- unnecessary credential exposure window
- OIDC short-lived tokens are stronger baseline
Confidence: [HIGH]

### 5) Running Temporal workflow logic with nondeterministic calls
Why not:
- replay failures and production instability
- hard-to-debug non-determinism regressions
Confidence: [HIGH]

### 6) Self-hosted Temporal by default for greenfield without SRE budget
Why not:
- operational burden (DB tuning, upgrades, scaling, backup/restore)
- delays product delivery
Confidence: [MEDIUM]

### 7) Pushing direct-to-main in output docs repo without review gates
Why not:
- accidental bad docs at scale
- no policy checkpoint for low-confidence content spikes
Confidence: [HIGH]

---

## Concrete Version Baseline (Starting Point)

- Go: `1.26.x`  [HIGH]
- Python: `3.13.x` (evaluate 3.14 after dependency validation)  [HIGH]
- Temporal Go SDK: `v1.43.x`  [HIGH]
- Temporal Python SDK: `1.27.x`  [HIGH]
- Temporal TypeScript SDK (if used): `1.17.x`  [MEDIUM]
- Node.js for auxiliary tooling: `v24 LTS`  [HIGH]
- actions/checkout: `v6`  [HIGH]
- Trivy: `v0.70.x`  [HIGH]
- Gitleaks: `v8.30.x`  [HIGH]
- Cosign: `v3.0.x`  [HIGH]

---

## Suggested Initial Build Plan (Pragmatic)

1. Stand up Temporal Cloud namespace + queues + base worker in Go. [HIGH]
2. Implement intake, source snapshot, and publish workflows first. [HIGH]
3. Add deterministic snapcraft/repo extraction pipeline and structured evidence schema. [HIGH]
4. Add markdown section marker patcher before any LLM features. [HIGH]
5. Add low-confidence caveat policy and confidence telemetry. [HIGH]
6. Add OTel + Prometheus + Grafana dashboards before broad rollout. [HIGH]
7. Add supply-chain controls (SBOM + cosign + trivy + gitleaks) as required checks. [HIGH]

---

## Sources

Temporal:
- https://docs.temporal.io/develop/go
- https://docs.temporal.io/develop/typescript
- https://docs.temporal.io/task-queue
- https://docs.temporal.io/workflow-execution
- https://docs.temporal.io/production-deployment/worker-deployments
- https://docs.temporal.io/self-hosted-guide/monitoring
- https://github.com/temporalio/sdk-go/releases
- https://github.com/temporalio/sdk-typescript/releases
- https://pypi.org/project/temporalio/

Ubuntu/Snapcraft:
- https://documentation.ubuntu.com/snapcraft/stable/reference/project-file/snapcraft-yaml/
- https://github.com/canonical/snapcraft/tree/main/docs/reference/snapcraft-yaml.rst
- https://github.com/canonical/snapcraft/tree/main/snapcraft/models/project.py

Runtime/version policy:
- https://go.dev/doc/devel/release
- https://nodejs.org/en/about/previous-releases
- https://devguide.python.org/versions/

CI/CD and security:
- https://docs.github.com/en/actions/concepts/security/openid-connect
- https://github.com/actions/checkout
- https://slsa.dev/
- https://github.com/sigstore/cosign
- https://github.com/aquasecurity/trivy
- https://github.com/gitleaks/gitleaks
- https://opentelemetry.io/docs/
- https://prometheus.io/docs/introduction/overview/
