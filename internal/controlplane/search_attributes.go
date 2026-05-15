package controlplane

import "go.temporal.io/sdk/temporal"

// Typed search-attribute keys for the Phase 1 field set.
// Using compile-time constants prevents the "magic string" anti-pattern and makes
// field renames a single-point-of-change.
var (
	SAKeyPackageName  = temporal.NewSearchAttributeKeyString("package_name")
	SAKeyGitHash      = temporal.NewSearchAttributeKeyString("git_hash")
	SAKeyFinalStatus  = temporal.NewSearchAttributeKeyString("final_status")
	SAKeyWorkflowID   = temporal.NewSearchAttributeKeyString("workflow_id")
	SAKeyRunID        = temporal.NewSearchAttributeKeyString("run_id")
	SAKeyRepoURL      = temporal.NewSearchAttributeKeyString("repo_url")
	SAKeyErrorClass   = temporal.NewSearchAttributeKeyString("error_class")
	SAKeyPublishState = temporal.NewSearchAttributeKeyString("publish_state")
)

// SearchAttributes holds all searchable metadata for a docs run.
type SearchAttributes struct {
	PackageName  string
	GitHash      string
	FinalStatus  string
	WorkflowID   string
	RunID        string
	RepoURL      string
	ErrorClass   string
	PublishState string
}

// ToTemporalSearchAttributes converts SearchAttributes to the Temporal typed
// search-attribute format suitable for StartWorkflowOptions.TypedSearchAttributes.
func (sa SearchAttributes) ToTemporalSearchAttributes() temporal.SearchAttributes {
	return temporal.NewSearchAttributes(
		SAKeyPackageName.ValueSet(sa.PackageName),
		SAKeyGitHash.ValueSet(sa.GitHash),
		SAKeyFinalStatus.ValueSet(sa.FinalStatus),
		SAKeyWorkflowID.ValueSet(sa.WorkflowID),
		SAKeyRunID.ValueSet(sa.RunID),
		SAKeyRepoURL.ValueSet(sa.RepoURL),
		SAKeyErrorClass.ValueSet(sa.ErrorClass),
		SAKeyPublishState.ValueSet(sa.PublishState),
	)
}

// InitialSearchAttributes builds the search attributes emitted when a workflow
// is first started — before source preflight or any other work begins.
func InitialSearchAttributes(req StartRequest, workflowID, runID string) SearchAttributes {
	return SearchAttributes{
		PackageName:  req.PackageName,
		GitHash:      req.GitHash,
		FinalStatus:  string(RunStatusRunning),
		WorkflowID:   workflowID,
		RunID:        runID,
		RepoURL:      req.SourceRepo,
		ErrorClass:   "",
		PublishState: "",
	}
}
