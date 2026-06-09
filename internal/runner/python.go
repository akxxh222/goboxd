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

func runPython(tempDir string, req types.RunRequest) types.RunResponse {
	sourcePath := filepath.Join(tempDir, "solution.py")
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		return types.RunResponse{
			Status: "internal_error",
		}
	}

	results := make([]types.TestResult, 0, len(req.Tests))
	overallStatus := "accepted"

	for _, test := range req.Tests {
		result := runPythonTest(tempDir, test)
		overallStatus = firstNonAccepted(overallStatus, result.Status)
		results = append(results, result)
	}

	return types.RunResponse{
		Status: overallStatus,
		Tests:  results,
	}
}

func runPythonTest(tempDir string, test types.TestCase) types.TestResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.RunTimeout)
	defer cancel()

	cmd := sandboxedCommand(ctx, tempDir, PythonCommand(), "solution.py")
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
	stderr.Write([]byte(processStderr(rawStderr, "python")))
	logNsjailFile(tempDir, "python")

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
