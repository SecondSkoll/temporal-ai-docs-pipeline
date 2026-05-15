# Temporal AI Docs Pipeline

## What This Is

Temporal AI Docs Pipeline is an automation workflow that takes package metadata (package name, source repository, and source git hash), analyzes repository content, and produces Ubuntu-focused usage documentation. It updates existing generated docs where possible and falls back to creating new docs when none exist. The generated docs are published to a separate output repository in a per-package folder structure for downstream AI model consumption.

## Core Value

Generate reliable Ubuntu-specific package documentation from source metadata with minimal manual effort.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Triggered pipeline accepts package name, source repository, and source git hash.
- [ ] Pipeline extracts repository context needed for Ubuntu usage documentation.
- [ ] Pipeline produces install and configuration documentation for each package.
- [ ] Pipeline updates existing generated sections with minimal diffs.
- [ ] Pipeline creates documentation when no prior docs exist.
- [ ] Pipeline writes outputs to a separate repository using per-package folders.
- [ ] Pipeline preserves throughput by generating docs with caveats when confidence is low.

### Out of Scope

- Full human-style editorial quality review loop in v1 — focus is reliable automated generation first.
- Broad non-Ubuntu platform documentation in v1 — core objective is Ubuntu-targeted guidance.

## Context

The workflow is temporal-based and receives package metadata tied to an exact source revision. Documentation consumers are AI models deployed on Ubuntu systems, so generated docs must prioritize Ubuntu-specific usage and operational relevance. The initial release must favor deterministic, repeatable generation over broad feature scope, while still handling uncertain extraction gracefully through caveats.

## Constraints

- **Orchestration**: Temporal-driven pipeline execution — workflow reliability and retry behavior must align with Temporal patterns.
- **Input Contract**: Package name, source repository, and source git hash are required — output must be traceable to an exact source revision.
- **Output Topology**: Separate output repository with per-package folders — avoids direct mutation of source repository docs.
- **Content Scope**: Ubuntu-focused install and configuration only for v1 — keeps MVP bounded and testable.
- **Update Strategy**: Prefer patching changed generated sections — preserve readable history and minimize unnecessary diffs.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Publish docs to a separate output repository | Decouples generation from source repository mutation and supports central consumption | — Pending |
| Use per-package folder output structure | Makes artifacts discoverable and scalable for many packages | — Pending |
| Include install + configuration in v1 docs | Matches immediate usefulness for Ubuntu deployments without overextending scope | — Pending |
| Use lenient confidence behavior with caveats | Maintains pipeline throughput while signaling uncertainty | — Pending |
| Prefer patching changed sections over full replacement | Keeps diffs small and makes output evolution easier to review | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? -> Move to Out of Scope with reason
2. Requirements validated? -> Move to Validated with phase reference
3. New requirements emerged? -> Add to Active
4. Decisions to log? -> Add to Key Decisions
5. "What This Is" still accurate? -> Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check - still the right priority?
3. Audit Out of Scope - reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-05-15 after initialization*
