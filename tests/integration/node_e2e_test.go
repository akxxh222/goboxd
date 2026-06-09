//go:build integration
// +build integration

package integration

import (
	"os"
	"testing"

	"github.com/thesouldev/goboxd/internal/runner"
	"github.com/thesouldev/goboxd/internal/types"
)

func TestNodeE2E(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "goboxd-node-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chmod(tempDir, 0755); err != nil {
		t.Fatalf("failed to chmod temp dir: %v", err)
	}

	req := types.RunRequest{
		Language: "node",
		Source:   "console.log('Hello from Node E2E');",
		Tests: []types.TestCase{
			{
				Stdin:          "",
				ExpectedStdout: "Hello from Node E2E\n",
			},
		},
	}

	resp, ok := runner.Run(tempDir, req)
	if !ok {
		t.Fatal("runner.Run returned false")
	}

	if resp.Status != "accepted" {
		t.Fatalf("expected status 'accepted', got '%s'", resp.Status)
	}

	if len(resp.Tests) != 1 {
		t.Fatalf("expected 1 test result, got %d", len(resp.Tests))
	}

	if resp.Tests[0].Stdout != "Hello from Node E2E\n" {
		t.Errorf("unexpected stdout: %q", resp.Tests[0].Stdout)
	}
}