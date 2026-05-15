# Phase 1 Walking Skeleton

## Purpose

This skeleton locks the thin end-to-end control plane for the first release slice: an HTTP start request enters the system, Temporal starts a deterministic workflow, duplicates are rejected by packageName + gitHash, source-read access is checked up front, and the run is searchable by operational metadata.

## Locked Decisions

- Public entrypoint: `POST /v1/docs/runs`
- Immediate response: HTTP `202 Accepted` with `runId` and `workflowId`
- Deterministic workflow ID: `docs/{packageName}/{gitHash}`
- Duplicate policy: reject the second start for the same packageName + gitHash pair
- Auth order: source read now, output write later in the publish phase
- Run metadata: package, git hash, final status, workflow/run ID, repo URL, error class, publish state

## Runtime Shape

- Language: Go for the control plane and Temporal workers
- HTTP library: standard `net/http` unless a later phase introduces a stronger need
- Workflow engine: Temporal Go SDK
- Persistence: PostgreSQL run ledger outside Temporal history
- Worker queues: orchestration, repo-io, generation, validation

## Directory Layout

- `cmd/docs-pipeline/` for the HTTP server and worker bootstrap
- `internal/controlplane/` for request validation, workflow start, idempotency, auth preflight, and search-attribute helpers
- `db/migrations/` for the run ledger schema

## Request and Response Contract

Request body:

```json
{
  "packageName": "example-package",
  "sourceRepo": "https://example.com/repo.git",
  "gitHash": "abcdef1234567890"
}
```

Accepted response:

```json
{
  "runId": "...",
  "workflowId": "docs/example-package/abcdef1234567890",
  "status": "accepted"
}
```

## Persistence Contract

- Run ledger row keyed by `packageName` + `gitHash`
- Final state recorded outside Temporal history for search and duplicate detection
- Duplicate start requests fail closed instead of returning an existing run

## Search Attributes

- `package_name`
- `git_hash`
- `final_status`
- `workflow_id`
- `run_id`
- `repo_url`
- `error_class`
- `publish_state`

## Deferred Until Later Phases

- Output repository write preflight
- Repository mining and extraction
- Generation and confidence scoring
- Publish patching and provenance commit creation

## Why This Is Thin Enough

The skeleton proves the ingress, idempotency, preflight, and searchability contract without pretending the mining or publish stages exist yet. Later phases can add behavior behind the same workflow ID and ledger without renegotiating the control-plane shape.