package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "github.com/canonical/temporal-ai-docs-pipeline/internal/controlplane/http"

	"github.com/canonical/temporal-ai-docs-pipeline/internal/controlplane"
	"go.temporal.io/sdk/mocks"
)

func startBody(pkg, repo, hash string) *bytes.Buffer {
	b, _ := json.Marshal(controlplane.StartRequest{
		PackageName: pkg,
		SourceRepo:  repo,
		GitHash:     hash,
	})
	return bytes.NewBuffer(b)
}

// TestHandleStart_WrongMethod_Returns405 tests HTTP method validation.
func TestHandleStart_WrongMethod_Returns405(t *testing.T) {
	h := &handler.StartHandler{Temporal: &mocks.Client{}, Ledger: nil}
	req := httptest.NewRequest(http.MethodGet, "/v1/docs/runs", nil)
	w := httptest.NewRecorder()
	h.HandleStart(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

// TestHandleStart_MissingPackageName_Returns400 tests request validation.
func TestHandleStart_MissingPackageName_Returns400(t *testing.T) {
	h := &handler.StartHandler{Temporal: &mocks.Client{}, Ledger: nil}
	req := httptest.NewRequest(http.MethodPost, "/v1/docs/runs",
		startBody("", "https://example.com/repo.git", "abc123"))
	w := httptest.NewRecorder()
	h.HandleStart(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestHandleStart_MissingSourceRepo_Returns400 tests request validation.
func TestHandleStart_MissingSourceRepo_Returns400(t *testing.T) {
	h := &handler.StartHandler{Temporal: &mocks.Client{}, Ledger: nil}
	req := httptest.NewRequest(http.MethodPost, "/v1/docs/runs",
		startBody("mypkg", "", "abc123"))
	w := httptest.NewRecorder()
	h.HandleStart(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestHandleStart_MissingGitHash_Returns400 tests request validation.
func TestHandleStart_MissingGitHash_Returns400(t *testing.T) {
	h := &handler.StartHandler{Temporal: &mocks.Client{}, Ledger: nil}
	req := httptest.NewRequest(http.MethodPost, "/v1/docs/runs",
		startBody("mypkg", "https://example.com/repo.git", ""))
	w := httptest.NewRecorder()
	h.HandleStart(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

