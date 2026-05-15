package controlplane_test

import (
	"context"
	"testing"

	"github.com/canonical/temporal-ai-docs-pipeline/internal/controlplane"
)

func TestCheckSourceAccessActivity_EmptyRepo_ReturnsError(t *testing.T) {
	_, err := controlplane.CheckSourceAccessActivity(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty sourceRepo, got nil")
	}
}

func TestCheckSourceAccessActivity_InvalidURL_ReturnsError(t *testing.T) {
	// A clearly unreachable URL must produce an error — the activity must not
	// swallow access failures.
	_, err := controlplane.CheckSourceAccessActivity(
		context.Background(),
		"https://this.host.does.not.exist.example.invalid/repo.git",
	)
	if err == nil {
		t.Fatal("expected error for unreachable repo, got nil (AuthResult{Accessible: true})")
	}
}

func TestCheckSourceAccessActivity_LocalRepo_Succeeds(t *testing.T) {
	// Use the current git repository as a known-accessible source.
	// This is a lightweight integration check that runs without network access.
	result, err := controlplane.CheckSourceAccessActivity(context.Background(), ".")
	if err != nil {
		t.Fatalf("expected success for local git repo, got: %v", err)
	}
	if !result.Accessible {
		t.Error("expected Accessible=true for local git repo")
	}
}

func TestRunStatus_SearchAttributeEmission(t *testing.T) {
	// Verify that InitialSearchAttributes correctly populates FinalStatus as
	// RunStatusRunning — this is the value emitted at workflow start.
	req := controlplane.StartRequest{
		PackageName: "test-pkg",
		SourceRepo:  "https://example.com/repo.git",
		GitHash:     "deadbeef",
	}
	sa := controlplane.InitialSearchAttributes(req, "docs/test-pkg/deadbeef", "run-789")
	if sa.FinalStatus != string(controlplane.RunStatusRunning) {
		t.Errorf("initial FinalStatus: got %q, want %q", sa.FinalStatus, controlplane.RunStatusRunning)
	}
	if sa.ErrorClass != "" {
		t.Errorf("initial ErrorClass should be empty, got %q", sa.ErrorClass)
	}
	if sa.PublishState != "" {
		t.Errorf("initial PublishState should be empty, got %q", sa.PublishState)
	}
}
