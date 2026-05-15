package controlplane

import (
	"context"
	"fmt"
	"os/exec"
)

// CheckSourceAccessActivity is a Temporal activity that verifies the source
// repository is reachable before expensive work begins.
//
// It runs `git ls-remote --exit-code <repo> HEAD` to probe connectivity and
// basic read access without cloning. A non-zero exit code is treated as an
// access failure.
//
// Registered to the orchestration task queue. Output auth (write access to the
// output repository) is deferred to the publish phase per D-04.
func CheckSourceAccessActivity(ctx context.Context, sourceRepo string) (AuthResult, error) {
	if sourceRepo == "" {
		return AuthResult{}, fmt.Errorf("sourceRepo must not be empty")
	}

	cmd := exec.CommandContext(ctx, "git", "ls-remote", "--exit-code", sourceRepo, "HEAD")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return AuthResult{}, fmt.Errorf("source repository not accessible (%s): %w", string(out), err)
	}

	return AuthResult{Accessible: true}, nil
}
