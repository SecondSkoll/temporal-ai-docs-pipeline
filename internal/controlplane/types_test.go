package controlplane_test

import (
	"testing"

	"github.com/canonical/temporal-ai-docs-pipeline/internal/controlplane"
)

func TestStartRequest_Validate_MissingPackageName(t *testing.T) {
	r := controlplane.StartRequest{SourceRepo: "https://example.com/repo.git", GitHash: "abc123"}
	if got := r.Validate(); got == "" {
		t.Fatal("expected validation error for missing packageName, got empty string")
	}
}

func TestStartRequest_Validate_MissingSourceRepo(t *testing.T) {
	r := controlplane.StartRequest{PackageName: "mypkg", GitHash: "abc123"}
	if got := r.Validate(); got == "" {
		t.Fatal("expected validation error for missing sourceRepo, got empty string")
	}
}

func TestStartRequest_Validate_MissingGitHash(t *testing.T) {
	r := controlplane.StartRequest{PackageName: "mypkg", SourceRepo: "https://example.com/repo.git"}
	if got := r.Validate(); got == "" {
		t.Fatal("expected validation error for missing gitHash, got empty string")
	}
}

func TestStartRequest_Validate_AllFieldsPresent(t *testing.T) {
	r := controlplane.StartRequest{
		PackageName: "mypkg",
		SourceRepo:  "https://example.com/repo.git",
		GitHash:     "abc123",
	}
	if got := r.Validate(); got != "" {
		t.Fatalf("expected no validation error, got: %q", got)
	}
}

func TestRunStatus_Constants(t *testing.T) {
	// Ensure the status constants match the string values the ledger and search
	// attributes use — any change is a breaking schema change.
	cases := map[controlplane.RunStatus]string{
		controlplane.RunStatusAccepted:  "accepted",
		controlplane.RunStatusRunning:   "running",
		controlplane.RunStatusCompleted: "completed",
		controlplane.RunStatusFailed:    "failed",
	}
	for status, want := range cases {
		if string(status) == "" {
			t.Errorf("RunStatus %q must not be empty", want)
		}
		if string(status) != want {
			t.Errorf("RunStatus value mismatch: got %q, want %q", string(status), want)
		}
	}
}
