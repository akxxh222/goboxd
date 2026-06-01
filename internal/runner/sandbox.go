package runner

import (
	"context"
	"os/exec"
	"runtime"

	"github.com/thesouldev/goboxd/internal/config"
)

func sandboxedCommand(ctx context.Context, workDir string, command string, args ...string) *exec.Cmd {
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
		"--really_quiet",
		"--user", "65534",
		"--group", "65534",
		"--time_limit", config.SandboxCPUSeconds,
		"--rlimit_cpu", config.SandboxCPUSeconds,
		"--rlimit_as", config.SandboxAddressSpaceMB,
		"--rlimit_fsize", config.SandboxFileSizeMB,
		"--rlimit_nofile", config.SandboxOpenFiles,
		"--rlimit_nproc", config.SandboxProcesses,
		"--chroot", "/",
		"--cwd", workDir,
		"--",
		commandPath,
	}
	nsjailArgs = append(nsjailArgs, args...)

	return exec.CommandContext(ctx, "nsjail", nsjailArgs...)
}

func PythonCommand() string {
	if runtime.GOOS == "windows" {
		return "python"
	}

	return "python3"
}
