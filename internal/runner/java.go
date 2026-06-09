package runner

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/types"
)

func runJava(tempDir string, req types.RunRequest) types.RunResponse {
	sourcePath := filepath.Join(tempDir, "Main.java")
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		return types.RunResponse{
			Status: "internal_error",
		}
	}

	build := buildJava(tempDir)
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
		result := runJavaTest(tempDir, test)
		overallStatus = firstNonAccepted(overallStatus, result.Status)
		results = append(results, result)
	}

	return types.RunResponse{
		Status: overallStatus,
		Build:  &build,
		Tests:  results,
	}
}

func buildJava(tempDir string) types.BuildResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.BuildTimeout)
	defer cancel()

	cmd := sandboxedCommandWithOptions(ctx, tempDir, SandboxOptions{
		TimeLimitSeconds: "10",
		AddressSpaceMB:   "1024",
		FileSizeMB:       "10",
		OpenFiles:        "128",
		Processes:        "32",
		ReadWriteDirs:    []string{tempDir},
	}, "javac", "Main.java")
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

	stderr = newCappedBuffer(config.MaxCapturedOutputLen)
	stderr.Write([]byte(processStderr(stderr.String(), "javac build")))
	logNsjailFile(tempDir, "javac build")

	return types.BuildResult{
		Status:          status,
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		StdoutTruncated: stdout.Truncated(),
		StderrTruncated: stderr.Truncated(),
		DurationMS:      time.Since(start).Milliseconds(),
	}
}

func runJavaTest(tempDir string, test types.TestCase) types.TestResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.RunTimeout)
	defer cancel()

	cmd := sandboxedCommand(ctx, tempDir, "java", "-cp", ".", "Main")
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

	stderr = newCappedBuffer(config.MaxCapturedOutputLen)
	stderr.Write([]byte(processStderr(stderr.String(), "java run")))
	logNsjailFile(tempDir, "java run")

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
