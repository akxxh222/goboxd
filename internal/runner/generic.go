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

// applyOverrides applies per-request resource limits (e.g., memory, processes)
// dynamically. This is required for languages like R that consume heavy threads (OpenMP).
func applyOverrides(opts *SandboxOptions, overrides *types.ResourceOverrides) {
	if overrides == nil {
		return
	}
	if overrides.TimeLimitSeconds != "" {
		opts.TimeLimitSeconds = overrides.TimeLimitSeconds
	}
	if overrides.AddressSpaceMB != "" {
		opts.AddressSpaceMB = overrides.AddressSpaceMB
	}
	if overrides.FileSizeMB != "" {
		opts.FileSizeMB = overrides.FileSizeMB
	}
	if overrides.OpenFiles != "" {
		opts.OpenFiles = overrides.OpenFiles
	}
	if overrides.Processes != "" {
		opts.Processes = overrides.Processes
	}
}

// runGeneric orchestrates the full lifecycle (build + test execution) for a language
// defined entirely through the YAML configuration.
func runGeneric(tempDir string, req types.RunRequest, def LanguageDef) types.RunResponse {
	// 1. Write the source code payload to the securely isolated temporary directory
	sourcePath := filepath.Join(tempDir, def.SourceFile)
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		return types.RunResponse{Status: "internal_error"}
	}

	// 2. Optional: Build phase (only for compiled languages like C++, OCaml)
	var buildResult *types.BuildResult
	if def.IsCompiled && len(def.BuildCmd) > 0 {
		b := buildGeneric(tempDir, req, def)
		buildResult = &b
		if b.Status != "ok" {
			return types.RunResponse{
				Status: "build_failed",
				Build:  buildResult,
				Tests:  notExecutedResults(len(req.Tests)),
			}
		}
	}

	// 3. Execution phase: Run each test case in its own isolated sandbox
	results := make([]types.TestResult, 0, len(req.Tests))
	overallStatus := "accepted"

	for _, test := range req.Tests {
		result := runGenericTest(tempDir, req, def, test)
		overallStatus = firstNonAccepted(overallStatus, result.Status)
		results = append(results, result)
	}

	return types.RunResponse{
		Status: overallStatus,
		Build:  buildResult,
		Tests:  results,
	}
}

// buildGeneric executes the compiler command securely inside an nsjail sandbox.
func buildGeneric(tempDir string, req types.RunRequest, def LanguageDef) types.BuildResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.BuildTimeout)
	defer cancel()

	opts := SandboxOptions{
		TimeLimitSeconds: "10",
		AddressSpaceMB:   "max",
		FileSizeMB:       "10",
		OpenFiles:        "128",
		Processes:        "32",
		ReadWriteDirs:    []string{tempDir},
	}
	applyOverrides(&opts, req.ResourceOverrides)

	args := append([]string{}, def.BuildCmd...)
	args = append(args, req.BuildFlags...)

	cmd := sandboxedCommandWithOptions(ctx, tempDir, opts, args[0], args[1:]...)
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
	stderr.Write([]byte(processStderr(rawStderr, def.Name+" build")))

	return types.BuildResult{
		Status:          status,
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		StdoutTruncated: stdout.Truncated(),
		StderrTruncated: stderr.Truncated(),
		DurationMS:      time.Since(start).Milliseconds(),
	}
}

// runGenericTest executes the runtime command (e.g., Python, compiled binary) against a single test case.
func runGenericTest(tempDir string, req types.RunRequest, def LanguageDef, test types.TestCase) types.TestResult {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), config.RunTimeout)
	defer cancel()

	opts := SandboxOptions{
		TimeLimitSeconds: config.SandboxCPUSeconds,
		AddressSpaceMB:   config.SandboxAddressSpaceMB,
		FileSizeMB:       config.SandboxFileSizeMB,
		OpenFiles:        config.SandboxOpenFiles,
		Processes:        config.SandboxProcesses,
		ReadWriteDirs:    []string{tempDir},
	}
	applyOverrides(&opts, req.ResourceOverrides)

	args := append([]string{}, def.RunCmd...)
	args = append(args, req.RunFlags...)

	cmd := sandboxedCommandWithOptions(ctx, tempDir, opts, args[0], args[1:]...)
	cmd.Dir = tempDir

	stdout := newCappedBuffer(config.MaxCapturedOutputLen)
	stderr := newCappedBuffer(config.MaxCapturedOutputLen)
	cmd.Stdin = strings.NewReader(test.Stdin)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// Interpret exit codes and timeouts to provide standardized API statuses
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
	stderr.Write([]byte(processStderr(rawStderr, def.Name)))

	// Validate standard output against the expected answer
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