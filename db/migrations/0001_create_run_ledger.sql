-- Run ledger: tracks every docs-pipeline workflow start for duplicate detection,
-- status visibility, and operational lookup.
--
-- Keyed by (package_name, git_hash) — the deterministic idempotency unit.
-- The unique constraint is the enforcement mechanism for D-03 (reject duplicates).

CREATE TABLE IF NOT EXISTS run_ledger (
    id           UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    package_name TEXT        NOT NULL,
    git_hash     TEXT        NOT NULL,
    workflow_id  TEXT        NOT NULL,
    run_id       TEXT        NOT NULL,
    repo_url     TEXT        NOT NULL DEFAULT '',
    status       TEXT        NOT NULL DEFAULT 'accepted',
    error_class  TEXT        NOT NULL DEFAULT '',
    publish_state TEXT       NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_run UNIQUE (package_name, git_hash)
);

-- Operational indexes: support the Phase 1 search-attribute query patterns.
CREATE INDEX IF NOT EXISTS idx_run_ledger_package_name  ON run_ledger (package_name);
CREATE INDEX IF NOT EXISTS idx_run_ledger_git_hash      ON run_ledger (git_hash);
CREATE INDEX IF NOT EXISTS idx_run_ledger_status        ON run_ledger (status);
CREATE INDEX IF NOT EXISTS idx_run_ledger_workflow_id   ON run_ledger (workflow_id);
