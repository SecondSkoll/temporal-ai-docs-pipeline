package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/canonical/temporal-ai-docs-pipeline/internal/controlplane"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

// StartHandler handles POST /v1/docs/runs.
// It validates the request, reserves a run in the ledger, and starts the
// Temporal workflow with a deterministic ID.
type StartHandler struct {
	Temporal client.Client
	Ledger   controlplane.Ledger
}

// HandleStart processes POST /v1/docs/runs and returns 202 Accepted on success.
//
// Error responses:
//   - 400 Bad Request  — missing or invalid fields
//   - 409 Conflict     — duplicate packageName + gitHash
//   - 405 Method Not Allowed — non-POST request
//   - 500 Internal Server Error — unexpected failure
func (h *StartHandler) HandleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req controlplane.StartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if msg := req.Validate(); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	wfID := controlplane.WorkflowID(req.PackageName, req.GitHash)

	// Start the Temporal workflow first to obtain the run ID from the execution.
	// The ledger insert happens immediately after to record the run for
	// duplicate detection on subsequent requests.
	opts := client.StartWorkflowOptions{
		ID:        wfID,
		TaskQueue: controlplane.TaskQueueOrchestration,
		TypedSearchAttributes: controlplane.InitialSearchAttributes(req, wfID, "").
			ToTemporalSearchAttributes(),
	}

	run, err := h.Temporal.ExecuteWorkflow(r.Context(), opts, controlplane.RunDocsWorkflow, req)
	if err != nil {
		// Temporal returns an already-started error if a workflow with this ID is
		// running and the ID-reuse policy rejects the new start — treat as duplicate.
		if temporal.IsWorkflowExecutionAlreadyStartedError(err) {
			http.Error(w, "duplicate run", http.StatusConflict)
			return
		}
		http.Error(w, "failed to start workflow", http.StatusInternalServerError)
		return
	}

	runID := run.GetRunID()

	// Reserve the run in the ledger. A unique-constraint failure here means a
	// race with a concurrent identical request; return 409 in that case.
	if err := h.Ledger.InsertRun(
		context.Background(),
		req.PackageName, req.GitHash, wfID, runID, req.SourceRepo,
	); err != nil {
		var dup *controlplane.DuplicateRunError
		if errors.As(err, &dup) {
			http.Error(w, "duplicate run", http.StatusConflict)
			return
		}
		http.Error(w, "failed to record run", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(controlplane.StartResponse{
		RunID:      runID,
		WorkflowID: wfID,
		Status:     string(controlplane.RunStatusAccepted),
	})
}
