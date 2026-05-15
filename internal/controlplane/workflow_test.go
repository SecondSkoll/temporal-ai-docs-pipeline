package controlplane_test

import (
	"fmt"
	"testing"

	"github.com/canonical/temporal-ai-docs-pipeline/internal/controlplane"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

// WorkflowTestSuite embeds Temporal's test suite helpers.
type WorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
}

func TestWorkflowSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuite))
}

func TestWorkflowID_Format(t *testing.T) {
	got := controlplane.WorkflowID("mypkg", "abc123")
	want := "docs/mypkg/abc123"
	if got != want {
		t.Errorf("WorkflowID: got %q, want %q", got, want)
	}
}

func TestWorkflowID_SlashSeparated(t *testing.T) {
	// Ensure the deterministic ID never changes format — breaking this is a
	// migration-level change to the duplicate-detection key.
	id := controlplane.WorkflowID("example-package", "abcdef1234567890")
	if id != "docs/example-package/abcdef1234567890" {
		t.Errorf("unexpected WorkflowID format: %q", id)
	}
}

func (s *WorkflowTestSuite) TestRunDocsWorkflow_SourceAuthSuccess() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(controlplane.RunDocsWorkflow)
	env.RegisterActivity(controlplane.CheckSourceAccessActivity)

	la := &controlplane.LedgerActivities{Ledger: newMemLedger()}
	env.RegisterActivity(la)

	env.OnActivity(controlplane.CheckSourceAccessActivity, mock.Anything, "https://example.com/repo.git").
		Return(controlplane.AuthResult{Accessible: true}, nil)
	env.OnActivity(la.MarkRunStatusActivity, mock.Anything,
		"mypkg", "abc123", controlplane.RunStatusCompleted, "").
		Return(nil)

	env.ExecuteWorkflow(controlplane.RunDocsWorkflow, controlplane.StartRequest{
		PackageName: "mypkg",
		SourceRepo:  "https://example.com/repo.git",
		GitHash:     "abc123",
	})

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())
}

func (s *WorkflowTestSuite) TestRunDocsWorkflow_SourceAuthFailure() {
	env := s.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(controlplane.RunDocsWorkflow)
	env.RegisterActivity(controlplane.CheckSourceAccessActivity)

	la := &controlplane.LedgerActivities{Ledger: newMemLedger()}
	env.RegisterActivity(la)

	env.OnActivity(controlplane.CheckSourceAccessActivity, mock.Anything, "https://bad.example.com/repo.git").
		Return(controlplane.AuthResult{}, fmt.Errorf("access denied"))
	env.OnActivity(la.MarkRunStatusActivity, mock.Anything,
		"mypkg", "abc123", controlplane.RunStatusFailed, "source_auth_failure").
		Return(nil)

	env.ExecuteWorkflow(controlplane.RunDocsWorkflow, controlplane.StartRequest{
		PackageName: "mypkg",
		SourceRepo:  "https://bad.example.com/repo.git",
		GitHash:     "abc123",
	})

	s.True(env.IsWorkflowCompleted())
	s.Error(env.GetWorkflowError(), "workflow should fail when source auth fails")
}
