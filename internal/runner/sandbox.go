package runner

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/thesouldev/goboxd/internal/config"
)

// SandboxOptions configures the resource limits passed to nsjail.
// These prevent denial-of-service attacks like fork bombs or memory exhaustion.
type SandboxOptions struct {
	TimeLimitSeconds string
	AddressSpaceMB   string
	FileSizeMB       string
	OpenFiles        string
	Processes        string
	ReadWriteDirs    []string
}

// sandboxedCommand is a convenience wrapper for sandboxedCommandWithOptions using defaults.
func sandboxedCommand(ctx context.Context, workDir string, command string, args ...string) *exec.Cmd {
	return sandboxedCommandWithOptions(ctx, workDir, SandboxOptions{
		TimeLimitSeconds: config.SandboxCPUSeconds,
		AddressSpaceMB:   config.SandboxAddressSpaceMB,
		FileSizeMB:       config.SandboxFileSizeMB,
		OpenFiles:        config.SandboxOpenFiles,
		Processes:        config.SandboxProcesses,
	}, command, args...)
}

// sandboxedCommandWithOptions wraps a command execution with nsjail isolation.
func sandboxedCommandWithOptions(ctx context.Context, workDir string, opts SandboxOptions, command string, args ...string) *exec.Cmd {
	// Windows local development fallback (nsjail is Linux-only).
	// This allows devs to run basic tests on Windows without isolation.
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, command, args...)
	}

	if _, err := exec.LookPath("nsjail"); err != nil {
		return exec.CommandContext(ctx, command, args...)
	}

	commandPath := command
	if path, err := exec.LookPath(command); err == nil {
		commandPath = path
	}

	// Construct the highly restrictive nsjail command line arguments
	nsjailArgs := []string{
		"-Mo",
		"--user", "65534",
		"--group", "65534",
		"--time_limit", opts.TimeLimitSeconds,
		"--rlimit_cpu", opts.TimeLimitSeconds,
		"--rlimit_as", opts.AddressSpaceMB,
		"--rlimit_fsize", opts.FileSizeMB,
		"--rlimit_nofile", opts.OpenFiles,
		"--rlimit_nproc", opts.Processes,
		"--chroot", "/",
		"-E", "PATH",
		// Ensure absolute TMPDIR is used to fix issues with R and other TMPDIR-dependent languages
		"-E", "TMPDIR=" + workDir,
	}

	// Explicitly map allowed read/write directories (by default, root is read-only)
	for _, rwDir := range opts.ReadWriteDirs {
		nsjailArgs = append(nsjailArgs, "-B", rwDir)
	}

	nsjailArgs = append(nsjailArgs,
		"--cwd", workDir,
		"--log", filepath.Join(workDir, "nsjail.log"),
		"--",
		commandPath,
	)
	nsjailArgs = append(nsjailArgs, args...)

	return exec.CommandContext(ctx, "nsjail", nsjailArgs...)
}

// PythonCommand resolves the correct python interpreter name across OS environments.
func PythonCommand() string {
	if runtime.GOOS == "windows" {
		return "python"
	}

	return "python3"
}
