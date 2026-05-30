package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/types"
)

func runCpp(tempDir string, req types.RunRequest) types.RunResponse {
	sourcePath := filepath.Join(tempDir, "solution.cpp")
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		return types.RunResponse{
			Status: "internal_error",
		}
	}

	binaryName := cppBinaryName()
	build := buildCpp(tempDir, binaryName)
	if build.Status != "ok" {
		return types.RunResponse{
			Status: "build_failed",
			Build:  &build,
			Tests:  notExecutedResults(len(req.Tests)),
		}
	}

	results := make([]types.TestResult, 0, len(req.Tests))
	overallStatus := "accepted"
	executable := "." + string(os.PathSeparator) + binaryName

	for _, test := range req.Tests {
		result := runCppTest(tempDir, executable, test)
		overallStatus = firstNonAccepted(overallStatus, result.Status)
		results = append(results, result)
	}

	return types.RunResponse{
		Status: overallStatus,
		Build:  &build,
		Tests:  results,
	}
}

func cppBinaryName() string {
	if runtime.GOOS == "windows" {
		return "solution.exe"
	}

	return "solution"
}

func buildCpp(tempDir string, binaryName string) types.BuildResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.BuildTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "g++", "solution.cpp", "-o", binaryName)
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

	return types.BuildResult{
		Status:          status,
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		StdoutTruncated: stdout.Truncated(),
		StderrTruncated: stderr.Truncated(),
		DurationMS:      time.Since(start).Milliseconds(),
	}
}

func runCppTest(tempDir string, executable string, test types.TestCase) types.TestResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.RunTimeout)
	defer cancel()

	cmd := sandboxedCommand(ctx, tempDir, executable)
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
	} else if strings.TrimSpace(stdout.String()) != strings.TrimSpace(test.ExpectedStdout) {
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

func notExecutedResults(count int) []types.TestResult {
	results := make([]types.TestResult, 0, count)
	for range count {
		results = append(results, types.TestResult{
			Status: "not_executed",
		})
	}

	return results
}
