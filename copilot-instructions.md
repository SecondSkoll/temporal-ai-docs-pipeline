# Copilot Instructions

This repository uses the GSD planning workflow. Prefer planning-first execution and keep planning artifacts in `.planning/` current.

## Current Project Context

- Project: Temporal AI Docs Pipeline
- Milestone: v1
- Core value: Generate reliable Ubuntu-specific package documentation from source metadata with minimal manual effort.
- Current phase: Phase 1 - Foundation

## Workflow Rules

- Read `.planning/PROJECT.md`, `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, and `.planning/STATE.md` before major implementation work.
- Keep requirements traceability intact when roadmap or scope changes.
- Use deterministic, testable acceptance criteria tied to roadmap phases.
- Do not overwrite generated planning artifacts without preserving intent and history.

## Next Command

Run:

`/gsd-discuss-phase 1`

If discussion is already complete, run:

`/gsd-plan-phase 1`
