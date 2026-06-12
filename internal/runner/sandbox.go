package runner

import (
	"context"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/thesouldev/goboxd/internal/config"
)

type SandboxOptions struct {
	TimeLimitSeconds string
	AddressSpaceMB   string
	FileSizeMB       string
	OpenFiles        string
	Processes        string
	ReadWriteDirs    []string
}

func sandboxedCommand(ctx context.Context, workDir string, command string, args ...string) *exec.Cmd {
	return sandboxedCommandWithOptions(ctx, workDir, SandboxOptions{
		TimeLimitSeconds: config.SandboxCPUSeconds,
		AddressSpaceMB:   config.SandboxAddressSpaceMB,
		FileSizeMB:       config.SandboxFileSizeMB,
		OpenFiles:        config.SandboxOpenFiles,
		Processes:        config.SandboxProcesses,
	}, command, args...)
}

func sandboxedCommandWithOptions(ctx context.Context, workDir string, opts SandboxOptions, command string, args ...string) *exec.Cmd {
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
		"-E", "TMPDIR=" + workDir,
	}

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

func PythonCommand() string {
	if runtime.GOOS == "windows" {
		return "python"
	}

	return "python3"
}
