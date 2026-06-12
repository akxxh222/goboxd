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

// runJava is a Stage 1 legacy executor for Java code.
// It serves as a fallback if java is removed from languages.yaml.
func runJava(tempDir string, req types.RunRequest) types.RunResponse {
	sourcePath := filepath.Join(tempDir, "Main.java")
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		return types.RunResponse{
			Status: "internal_error",
		}
	}

	// 1. Compile Phase using javac
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

	// 2. Execution Phase using java
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
		AddressSpaceMB:   "max",
		FileSizeMB:       "10",
		OpenFiles:        "max",
		Processes:        "max",
		ReadWriteDirs:    []string{tempDir},
	}, "javac", "-J-Xmx512m", "Main.java")
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
	stderr.Write([]byte(processStderr(rawStderr, "javac build")))
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

	cmd := sandboxedCommandWithOptions(ctx, tempDir, SandboxOptions{
		TimeLimitSeconds: config.SandboxCPUSeconds,
		AddressSpaceMB:   "max",
		FileSizeMB:       config.SandboxFileSizeMB,
		OpenFiles:        "max",
		Processes:        "max",
		ReadWriteDirs:    []string{tempDir},
	}, "java", "-Xmx512m", "-cp", ".", "Main")
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
	stderr.Write([]byte(processStderr(rawStderr, "java run")))
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
