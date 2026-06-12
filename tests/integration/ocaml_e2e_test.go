//go:build integration
// +build integration

package integration

import (
	"os"
	"testing"

	"github.com/thesouldev/goboxd/internal/runner"
	"github.com/thesouldev/goboxd/internal/types"
)

func TestOcamlE2E(t *testing.T) {
	// Manually inject registry for E2E tests, mirroring languages.yaml
	runner.Registry = map[string]runner.LanguageDef{
		"ocaml": {
			ID:         "ocaml",
			Name:       "OCaml",
			SourceFile: "solution.ml",
			IsCompiled: true,
			BuildCmd:   []string{"ocamlopt", "-o", "solution", "solution.ml"},
			RunCmd:     []string{"./solution"},
		},
	}

	tempDir, err := os.MkdirTemp("", "goboxd-ocaml-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chmod(tempDir, 0755); err != nil {
		t.Fatalf("failed to chmod temp dir: %v", err)
	}

	req := types.RunRequest{
		Language: "ocaml",
		Source:   "let () = print_endline \"Hello from OCaml E2E\"",
		Tests: []types.TestCase{
			{Stdin: "", ExpectedStdout: "Hello from OCaml E2E\n"},
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