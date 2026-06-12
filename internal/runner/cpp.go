package runner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/types"
)

// runCpp is a Stage 1 legacy executor for C++ code.
// It serves as a fallback if cpp is removed from languages.yaml.
func runCpp(tempDir string, req types.RunRequest) types.RunResponse {
	sourcePath := filepath.Join(tempDir, "solution.cpp")
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		return types.RunResponse{
			Status: "internal_error",
		}
	}

	binaryName := cppBinaryName()
	// 1. Compile Phase using g++
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

	// 2. Execution Phase
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

	cmd := sandboxedCommandWithOptions(ctx, tempDir, SandboxOptions{
		TimeLimitSeconds: "10",
		AddressSpaceMB:   "1024", // g++ needs more memory to compile
		FileSizeMB:       "10",   // to write binary
		OpenFiles:        "128",
		Processes:        "32",
		ReadWriteDirs:    []string{tempDir},
	}, "g++", "-w", "solution.cpp", "-o", binaryName)
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
	stderr.Write([]byte(processStderr(rawStderr, "g++ build")))
	logNsjailFile(tempDir, "g++ build")

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
	}

	rawStderr := stderr.String()
	stderr = newCappedBuffer(config.MaxCapturedOutputLen)
	stderr.Write([]byte(processStderr(rawStderr, "cpp run")))
	logNsjailFile(tempDir, "cpp run")

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
