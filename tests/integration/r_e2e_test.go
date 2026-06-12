//go:build integration
// +build integration

package integration

import (
	"os"
	"testing"

	"github.com/thesouldev/goboxd/internal/runner"
	"github.com/thesouldev/goboxd/internal/types"
)

func TestRE2E(t *testing.T) {
	runner.Registry = map[string]runner.LanguageDef{
		"r": {
			ID:         "r",
			Name:       "R",
			SourceFile: "solution.R",
			RunCmd:     []string{"Rscript", "--vanilla", "solution.R"},
		},
	}

	tempDir, err := os.MkdirTemp("", "goboxd-r-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chmod(tempDir, 0755); err != nil {
		t.Fatalf("failed to chmod temp dir: %v", err)
	}

	req := types.RunRequest{
		Language: "r",
		Source:   "cat('Hello from R E2E\n')",
		ResourceOverrides: &types.ResourceOverrides{
			AddressSpaceMB: "max",
			Processes:      "max",
			OpenFiles:      "max",
			FileSizeMB:     "max",
		},
		Tests: []types.TestCase{
			{Stdin: "", ExpectedStdout: "Hello from R E2E\n"},
		},
	}

	resp, ok := runner.Run(tempDir, req)
	if !ok {
		t.Fatal("runner.Run returned false")
	}
	if resp.Status != "accepted" {
		t.Fatalf("expected status 'accepted', got '%s'", resp.Status)
	}
}