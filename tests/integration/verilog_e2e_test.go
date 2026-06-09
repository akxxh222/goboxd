//go:build integration
// +build integration

package integration

import (
	"os"
	"testing"

	"github.com/thesouldev/goboxd/internal/runner"
	"github.com/thesouldev/goboxd/internal/types"
)

func TestVerilogE2E(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "goboxd-verilog-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chmod(tempDir, 0755); err != nil {
		t.Fatalf("failed to chmod temp dir: %v", err)
	}

	req := types.RunRequest{
		Language: "verilog",
		Source:   "module main;\ninitial begin\n$display(\"Hello from Verilog E2E\");\n$finish;\nend\nendmodule",
		Tests: []types.TestCase{
			{
				Stdin:          "",
				ExpectedStdout: "Hello from Verilog E2E\n",
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

	if resp.Build == nil || resp.Build.Status != "ok" {
		t.Fatalf("expected build status 'ok', got '%v'", resp.Build)
	}

	if len(resp.Tests) != 1 {
		t.Fatalf("expected 1 test result, got %d", len(resp.Tests))
	}

	if resp.Tests[0].Stdout != "Hello from Verilog E2E\n" {
		t.Errorf("unexpected stdout: %q", resp.Tests[0].Stdout)
	}
}