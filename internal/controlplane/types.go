package controlplane

// StartRequest is the payload for POST /v1/docs/runs.
// All three fields are required; the handler rejects requests missing any of them.
type StartRequest struct {
	PackageName string `json:"packageName"`
	SourceRepo  string `json:"sourceRepo"`
	GitHash     string `json:"gitHash"`
}

// Validate returns an error string describing the first missing required field,
// or an empty string when the request is valid.
func (r StartRequest) Validate() string {
	switch {
	case r.PackageName == "":
		return "packageName is required"
	case r.SourceRepo == "":
		return "sourceRepo is required"
	case r.GitHash == "":
		return "gitHash is required"
	}
	return ""
}

// StartResponse is the body returned with HTTP 202 Accepted.
type StartResponse struct {
	RunID      string `json:"runId"`
	WorkflowID string `json:"workflowId"`
	Status     string `json:"status"`
}

// RunStatus represents the lifecycle state of a docs run stored in the ledger.
type RunStatus string

const (
	RunStatusAccepted  RunStatus = "accepted"
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
)

// AuthResult is returned by CheckSourceAccessActivity.
type AuthResult struct {
	Accessible bool
}

// Task-queue names shared between the HTTP entrypoint, worker, and workflow.
const (
	TaskQueueOrchestration = "orchestration"
	TaskQueueRepoIO        = "repo-io"
	TaskQueueGeneration    = "generation"
	TaskQueueValidation    = "validation"
)
