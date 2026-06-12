package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/types"
)

// runNode is a Stage 1 legacy executor for Node.js code.
// It serves as a fallback if node is removed from languages.yaml.
func runNode(tempDir string, req types.RunRequest) types.RunResponse {
	sourcePath := filepath.Join(tempDir, "solution.js")
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		return types.RunResponse{
			Status: "internal_error",
		}
	}

	results := make([]types.TestResult, 0, len(req.Tests))
	overallStatus := "accepted"

	for _, test := range req.Tests {
		result := runNodeTest(tempDir, test)
		overallStatus = firstNonAccepted(overallStatus, result.Status)
		results = append(results, result)
	}

	return types.RunResponse{
		Status: overallStatus,
		Tests:  results,
	}
}

func runNodeTest(tempDir string, test types.TestCase) types.TestResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.RunTimeout)
	defer cancel()

	cmd := sandboxedCommandWithOptions(ctx, tempDir, SandboxOptions{
		TimeLimitSeconds: config.SandboxCPUSeconds,
		AddressSpaceMB:   "max", // V8 engine requires a large address space
		FileSizeMB:       config.SandboxFileSizeMB,
		OpenFiles:        config.SandboxOpenFiles,
		Processes:        config.SandboxProcesses,
	}, "node", "solution.js")
	cmd.Dir = tempDir

	stdout := newCappedBuffer(config.MaxCapturedOutputLen)
	stderr := newCappedBuffer(config.MaxCapturedOutputLen)
	cmd.Stdin = strings.NewReader(test.Stdin)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	status := "accepted"
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			status = "time_limit_exceeded"
		} else {
			status = "runtime_error"
		}
	}

	rawStderr := stderr.String()
	stderr = newCappedBuffer(config.MaxCapturedOutputLen)
	stderr.Write([]byte(processStderr(rawStderr, "node")))
	logNsjailFile(tempDir, "node")

	if status == "accepted" && strings.TrimSpace(stdout.String()) != strings.TrimSpace(test.ExpectedStdout) {
		status = "wrong_output"
	}

	return types.TestResult{
		Status:          status,
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		StdoutTruncated: stdout.Truncated(),
		StderrTruncated: stderr.Truncated(),
		DurationMS:      time.Since(start).Milliseconds(),
	}
}