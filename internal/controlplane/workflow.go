package controlplane

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// WorkflowID returns the deterministic workflow ID for a docs run.
// Pattern: docs/{packageName}/{gitHash}
func WorkflowID(packageName, gitHash string) string {
	return fmt.Sprintf("docs/%s/%s", packageName, gitHash)
}

// LedgerActivities wraps Ledger operations as Temporal activities so the
// workflow can record status transitions without direct DB access.
type LedgerActivities struct {
	Ledger Ledger
}

// MarkRunStatusActivity updates the run ledger from within a workflow.
func (a *LedgerActivities) MarkRunStatusActivity(ctx context.Context, packageName, gitHash string, status RunStatus, errorClass string) error {
	return a.Ledger.MarkRunStatus(ctx, packageName, gitHash, status, errorClass)
}

// RunDocsWorkflow is the Phase 1 walking-skeleton Temporal workflow.
//
// Responsibilities:
//  1. Emit initial typed search attributes (package, hash, status, repo URL).
//  2. Run the source-access preflight activity.
//  3. Emit final search attributes on success or failure.
//  4. Update the run ledger with the final status via activity.
//
// Later phases add repository mining and generation activities in place of
// the stubs below.
func RunDocsWorkflow(ctx workflow.Context, req StartRequest) error {
	info := workflow.GetInfo(ctx)
	logger := workflow.GetLogger(ctx)

	// Emit initial search attributes so the run is immediately discoverable.
	if err := workflow.UpsertTypedSearchAttributes(ctx,
		SAKeyPackageName.ValueSet(req.PackageName),
		SAKeyGitHash.ValueSet(req.GitHash),
		SAKeyFinalStatus.ValueSet(string(RunStatusRunning)),
		SAKeyWorkflowID.ValueSet(info.WorkflowExecution.ID),
		SAKeyRunID.ValueSet(info.WorkflowExecution.RunID),
		SAKeyRepoURL.ValueSet(req.SourceRepo),
		SAKeyErrorClass.ValueSet(""),
		SAKeyPublishState.ValueSet(""),
	); err != nil {
		return fmt.Errorf("emit initial search attributes: %w", err)
	}

	// Activity options for lightweight preflight and ledger operations.
	ao := workflow.ActivityOptions{
		TaskQueue:           TaskQueueOrchestration,
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	actCtx := workflow.WithActivityOptions(ctx, ao)

	// --- Phase 1: Source-access preflight ---
	var authResult AuthResult
	if err := workflow.ExecuteActivity(actCtx, CheckSourceAccessActivity, req.SourceRepo).Get(actCtx, &authResult); err != nil {
		logger.Error("source access preflight failed", "error", err)

		// Best-effort: mark the run as failed in the ledger.
		var la *LedgerActivities
		_ = workflow.ExecuteActivity(actCtx, la.MarkRunStatusActivity,
			req.PackageName, req.GitHash, RunStatusFailed, "source_auth_failure",
		).Get(actCtx, nil)

		// Update search attributes to reflect the failure.
		_ = workflow.UpsertTypedSearchAttributes(ctx,
			SAKeyFinalStatus.ValueSet(string(RunStatusFailed)),
			SAKeyErrorClass.ValueSet("source_auth_failure"),
		)
		return fmt.Errorf("source access preflight: %w", err)
	}

	// --- Phase 2 placeholder: repository mining ---
	// --- Phase 3 placeholder: documentation generation ---
	// --- Phase 4 placeholder: publishing ---

	// Mark the run as completed in the ledger.
	var la *LedgerActivities
	if err := workflow.ExecuteActivity(actCtx, la.MarkRunStatusActivity,
		req.PackageName, req.GitHash, RunStatusCompleted, "",
	).Get(actCtx, nil); err != nil {
		logger.Error("failed to mark run completed", "error", err)
		// Non-fatal for Phase 1 skeleton — log and continue.
	}

	// Emit completion search attributes.
	_ = workflow.UpsertTypedSearchAttributes(ctx,
		SAKeyFinalStatus.ValueSet(string(RunStatusCompleted)),
		SAKeyPublishState.ValueSet("pending"),
	)

	return nil
}
