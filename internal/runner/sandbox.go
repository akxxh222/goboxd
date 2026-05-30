package runner

import (
	"context"
	"os/exec"
	"runtime"
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
