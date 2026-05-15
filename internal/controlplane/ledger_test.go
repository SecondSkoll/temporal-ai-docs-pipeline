package controlplane_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/canonical/temporal-ai-docs-pipeline/internal/controlplane"
)

// memLedger is an in-memory Ledger for unit tests — no database required.
type memLedger struct {
	mu   sync.Mutex
	runs map[string]*memRun
}

type memRun struct {
	workflowID string
	runID      string
	repoURL    string
	status     controlplane.RunStatus
	errorClass string
}

func newMemLedger() *memLedger {
	return &memLedger{runs: make(map[string]*memRun)}
}

func runKey(packageName, gitHash string) string {
	return fmt.Sprintf("%s|%s", packageName, gitHash)
}

func (m *memLedger) InsertRun(_ context.Context, packageName, gitHash, workflowID, runID, repoURL string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := runKey(packageName, gitHash)
	if _, exists := m.runs[key]; exists {
		return &controlplane.DuplicateRunError{PackageName: packageName, GitHash: gitHash}
	}
	m.runs[key] = &memRun{
		workflowID: workflowID,
		runID:      runID,
		repoURL:    repoURL,
		status:     controlplane.RunStatusAccepted,
	}
	return nil
}

func (m *memLedger) MarkRunStatus(_ context.Context, packageName, gitHash string, status controlplane.RunStatus, errorClass string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := runKey(packageName, gitHash)
	run, ok := m.runs[key]
	if !ok {
		return fmt.Errorf("run not found: packageName=%s gitHash=%s", packageName, gitHash)
	}
	run.status = status
	run.errorClass = errorClass
	return nil
}

// --- tests ---

func TestLedger_InsertRun_Succeeds(t *testing.T) {
	l := newMemLedger()
	err := l.InsertRun(context.Background(), "mypkg", "abc123", "docs/mypkg/abc123", "run-1", "https://example.com/repo.git")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestLedger_InsertRun_Duplicate(t *testing.T) {
	l := newMemLedger()
	ctx := context.Background()
	_ = l.InsertRun(ctx, "mypkg", "abc123", "docs/mypkg/abc123", "run-1", "https://example.com/repo.git")

	err := l.InsertRun(ctx, "mypkg", "abc123", "docs/mypkg/abc123", "run-2", "https://example.com/repo.git")
	if err == nil {
		t.Fatal("expected DuplicateRunError for second insert, got nil")
	}
	var dup *controlplane.DuplicateRunError
	if !errors.As(err, &dup) {
		t.Fatalf("expected *DuplicateRunError, got %T: %v", err, err)
	}
	if dup.PackageName != "mypkg" || dup.GitHash != "abc123" {
		t.Errorf("DuplicateRunError fields: got package=%q hash=%q, want mypkg/abc123", dup.PackageName, dup.GitHash)
	}
}

func TestLedger_InsertRun_DifferentHashNotDuplicate(t *testing.T) {
	l := newMemLedger()
	ctx := context.Background()
	_ = l.InsertRun(ctx, "mypkg", "abc123", "docs/mypkg/abc123", "run-1", "https://example.com/repo.git")

	// Same package, different hash — must succeed.
	err := l.InsertRun(ctx, "mypkg", "def456", "docs/mypkg/def456", "run-2", "https://example.com/repo.git")
	if err != nil {
		t.Fatalf("expected no error for different hash, got: %v", err)
	}
}

func TestLedger_MarkRunStatus_UpdatesStatus(t *testing.T) {
	l := newMemLedger()
	ctx := context.Background()
	_ = l.InsertRun(ctx, "mypkg", "abc123", "docs/mypkg/abc123", "run-1", "")

	if err := l.MarkRunStatus(ctx, "mypkg", "abc123", controlplane.RunStatusCompleted, ""); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if l.runs[runKey("mypkg", "abc123")].status != controlplane.RunStatusCompleted {
		t.Errorf("expected status %q, got %q", controlplane.RunStatusCompleted, l.runs[runKey("mypkg", "abc123")].status)
	}
}

func TestLedger_MarkRunStatus_NotFound(t *testing.T) {
	l := newMemLedger()
	err := l.MarkRunStatus(context.Background(), "missing", "nohash", controlplane.RunStatusFailed, "not_found")
	if err == nil {
		t.Fatal("expected error for missing run, got nil")
	}
}

func TestDuplicateRunError_Message(t *testing.T) {
	err := &controlplane.DuplicateRunError{PackageName: "pkg", GitHash: "hash1"}
	if err.Error() == "" {
		t.Fatal("DuplicateRunError.Error() must not be empty")
	}
}
