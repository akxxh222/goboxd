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

func runVerilog(tempDir string, req types.RunRequest) types.RunResponse {
	sourcePath := filepath.Join(tempDir, "solution.v")
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		return types.RunResponse{
			Status: "internal_error",
		}
	}

	binaryName := "solution.vvp"
	build := buildVerilog(tempDir, binaryName)
	if build.Status != "ok" {
		return types.RunResponse{
			Status: "build_failed",
			Build:  &build,
			Tests:  notExecutedResults(len(req.Tests)),
		}
	}

	results := make([]types.TestResult, 0, len(req.Tests))
	overallStatus := "accepted"

	for _, test := range req.Tests {
		result := runVerilogTest(tempDir, binaryName, test)
		overallStatus = firstNonAccepted(overallStatus, result.Status)
		results = append(results, result)
	}

	return types.RunResponse{
		Status: overallStatus,
		Build:  &build,
		Tests:  results,
	}
}

func buildVerilog(tempDir string, binaryName string) types.BuildResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.BuildTimeout)
	defer cancel()

	cmd := sandboxedCommandWithOptions(ctx, tempDir, SandboxOptions{
		TimeLimitSeconds: "10",
		AddressSpaceMB:   "max",
		FileSizeMB:       "10",
		OpenFiles:        "128",
		Processes:        "32",
		ReadWriteDirs:    []string{tempDir},
	}, "env", "TMPDIR=.", "iverilog", "-o", binaryName, "solution.v")
	cmd.Dir = tempDir

	stdout := newCappedBuffer(config.MaxCapturedOutputLen)
	stderr := newCappedBuffer(config.MaxCapturedOutputLen)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	status := "ok"
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			status = "time_limit_exceeded"
		} else {
			status = "failed"
		}
	}

	rawStderr := stderr.String()
	stderr = newCappedBuffer(config.MaxCapturedOutputLen)
	stderr.Write([]byte(processStderr(rawStderr, "iverilog build")))
	logNsjailFile(tempDir, "iverilog build")

	return types.BuildResult{
		Status:          status,
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		StdoutTruncated: stdout.Truncated(),
		StderrTruncated: stderr.Truncated(),
		DurationMS:      time.Since(start).Milliseconds(),
	}
}

func runVerilogTest(tempDir string, executable string, test types.TestCase) types.TestResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.RunTimeout)
	defer cancel()

	cmd := sandboxedCommand(ctx, tempDir, "vvp", executable)
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
	stderr.Write([]byte(processStderr(rawStderr, "vvp run")))
	logNsjailFile(tempDir, "vvp run")

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