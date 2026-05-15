package controlplane

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DuplicateRunError is returned by InsertRun when a run with the same
// packageName + gitHash already exists in the ledger.
type DuplicateRunError struct {
	PackageName string
	GitHash     string
}

func (e *DuplicateRunError) Error() string {
	return fmt.Sprintf("duplicate run: packageName=%s gitHash=%s", e.PackageName, e.GitHash)
}

// Ledger abstracts run-ledger persistence so the HTTP handler and workflow
// can be tested without a real PostgreSQL instance.
type Ledger interface {
	InsertRun(ctx context.Context, packageName, gitHash, workflowID, runID, repoURL string) error
	MarkRunStatus(ctx context.Context, packageName, gitHash string, status RunStatus, errorClass string) error
}

// pgLedger is the PostgreSQL-backed Ledger implementation.
type pgLedger struct {
	db *pgxpool.Pool
}

// NewLedger returns a Ledger backed by the supplied pgxpool.Pool.
func NewLedger(db *pgxpool.Pool) Ledger {
	return &pgLedger{db: db}
}

// InsertRun reserves a new run in the ledger with status "accepted".
// Returns DuplicateRunError if the (packageName, gitHash) pair already exists.
func (l *pgLedger) InsertRun(ctx context.Context, packageName, gitHash, workflowID, runID, repoURL string) error {
	_, err := l.db.Exec(ctx, `
		INSERT INTO run_ledger (package_name, git_hash, workflow_id, run_id, repo_url, status)
		VALUES ($1, $2, $3, $4, $5, 'accepted')
	`, packageName, gitHash, workflowID, runID, repoURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return &DuplicateRunError{PackageName: packageName, GitHash: gitHash}
		}
		return fmt.Errorf("insert run: %w", err)
	}
	return nil
}

// MarkRunStatus updates the status and error class for a run identified by
// packageName + gitHash.
func (l *pgLedger) MarkRunStatus(ctx context.Context, packageName, gitHash string, status RunStatus, errorClass string) error {
	result, err := l.db.Exec(ctx, `
		UPDATE run_ledger
		SET status = $1, error_class = $2, updated_at = NOW()
		WHERE package_name = $3 AND git_hash = $4
	`, string(status), errorClass, packageName, gitHash)
	if err != nil {
		return fmt.Errorf("mark run status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("run not found: packageName=%s gitHash=%s", packageName, gitHash)
	}
	return nil
}

// InsertRun is a package-level convenience wrapper that delegates to a Ledger.
// Used by code that receives a Ledger by interface and wants a terse call site.
func InsertRun(ctx context.Context, l Ledger, packageName, gitHash, workflowID, runID, repoURL string) error {
	return l.InsertRun(ctx, packageName, gitHash, workflowID, runID, repoURL)
}

// MarkRunStatus is a package-level convenience wrapper that delegates to a Ledger.
func MarkRunStatus(ctx context.Context, l Ledger, packageName, gitHash string, status RunStatus, errorClass string) error {
	return l.MarkRunStatus(ctx, packageName, gitHash, status, errorClass)
}
