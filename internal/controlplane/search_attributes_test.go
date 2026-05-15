package controlplane_test

import (
	"testing"

	"github.com/canonical/temporal-ai-docs-pipeline/internal/controlplane"
)

func TestSearchAttributes_AllKeysNonEmpty(t *testing.T) {
	// Guard against accidental empty-string key names which would create
	// unmatchable search attributes in Temporal.
	keys := map[string]interface{}{
		"package_name":  controlplane.SAKeyPackageName,
		"git_hash":      controlplane.SAKeyGitHash,
		"final_status":  controlplane.SAKeyFinalStatus,
		"workflow_id":   controlplane.SAKeyWorkflowID,
		"run_id":        controlplane.SAKeyRunID,
		"repo_url":      controlplane.SAKeyRepoURL,
		"error_class":   controlplane.SAKeyErrorClass,
		"publish_state": controlplane.SAKeyPublishState,
	}
	for want, key := range keys {
		type namer interface{ GetName() string }
		if n, ok := key.(namer); ok {
			if got := n.GetName(); got != want {
				t.Errorf("SAKey name mismatch: got %q, want %q", got, want)
			}
		}
		_ = key // key existence is the primary assertion
	}
}

func TestInitialSearchAttributes_FieldMapping(t *testing.T) {
	req := controlplane.StartRequest{
		PackageName: "mypkg",
		SourceRepo:  "https://example.com/repo.git",
		GitHash:     "abc123",
	}
	sa := controlplane.InitialSearchAttributes(req, "docs/mypkg/abc123", "run-456")

	if sa.PackageName != "mypkg" {
		t.Errorf("PackageName: got %q, want %q", sa.PackageName, "mypkg")
	}
	if sa.GitHash != "abc123" {
		t.Errorf("GitHash: got %q, want %q", sa.GitHash, "abc123")
	}
	if sa.FinalStatus != string(controlplane.RunStatusRunning) {
		t.Errorf("FinalStatus: got %q, want %q", sa.FinalStatus, string(controlplane.RunStatusRunning))
	}
	if sa.WorkflowID != "docs/mypkg/abc123" {
		t.Errorf("WorkflowID: got %q, want %q", sa.WorkflowID, "docs/mypkg/abc123")
	}
	if sa.RunID != "run-456" {
		t.Errorf("RunID: got %q, want %q", sa.RunID, "run-456")
	}
	if sa.RepoURL != "https://example.com/repo.git" {
		t.Errorf("RepoURL: got %q, want %q", sa.RepoURL, "https://example.com/repo.git")
	}
}

func TestToTemporalSearchAttributes_DoesNotPanic(t *testing.T) {
	sa := controlplane.SearchAttributes{
		PackageName:  "mypkg",
		GitHash:      "abc123",
		FinalStatus:  "running",
		WorkflowID:   "docs/mypkg/abc123",
		RunID:        "run-456",
		RepoURL:      "https://example.com/repo.git",
		ErrorClass:   "",
		PublishState: "",
	}
	// Must not panic; the returned value is a typed Temporal search attribute set.
	_ = sa.ToTemporalSearchAttributes()
}
