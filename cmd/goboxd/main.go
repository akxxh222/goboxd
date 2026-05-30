package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/thesouldev/goboxd/internal/types"
)

const (
	buildTimeout = 10 * time.Second
	runTimeout   = 3 * time.Second

	maxRequestBodyBytes  = 1 << 20
	maxSourceBytes       = 256 << 10
	maxTests             = 25
	maxTestInputBytes    = 64 << 10
	maxExpectedBytes     = 64 << 10
	maxCapturedOutputLen = 64 << 10

	outputTruncatedMarker = "\n[output truncated]\n"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func readyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	checks := map[string]string{
		"nsjail": "ok",
		"python": "ok",
		"g++":    "ok",
	}

	if _, err := exec.LookPath("nsjail"); err != nil {
		checks["nsjail"] = "missing"
	}

	if _, err := exec.LookPath(pythonCommand()); err != nil {
		checks["python"] = "missing"
	}

	if _, err := exec.LookPath("g++"); err != nil {
		checks["g++"] = "missing"
	}

	status := "ready"
	for _, check := range checks {
		if check != "ok" {
			status = "not_ready"
			w.WriteHeader(http.StatusServiceUnavailable)
			break
		}
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status": status,
		"checks": checks,
	})
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]any{
		"name": "goboxd",
		"languages": []map[string]string{
			{
				"id":   "py3",
				"name": "Python 3",
			},
			{
				"id":   "cpp",
				"name": "C++",
			},
		},
		"endpoints": []string{
			"GET /healthz",
			"GET /readyz",
			"GET /info",
			"POST /run",
		},
	})
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	var req types.RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		message := "invalid JSON request"
		if strings.Contains(err.Error(), "request body too large") {
			message = "request body too large"
		}
		json.NewEncoder(w).Encode(map[string]string{
			"error": message,
		})
		return
	}

	if errorMessage := validateRunRequest(req); errorMessage != "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": errorMessage,
		})
		return
	}

	tempDir, err := os.MkdirTemp("", "goboxd-*")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "failed to create temp directory",
		})
		return
	}
	defer os.RemoveAll(tempDir)
	if err := os.Chmod(tempDir, 0755); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "failed to prepare temp directory",
		})
		return
	}

	var response map[string]any
	switch req.Language {
	case "py3":
		response = runPython(tempDir, req)
	case "cpp":
		response = runCpp(tempDir, req)
	default:
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown language",
		})
		return
	}

	json.NewEncoder(w).Encode(response)
}

func validateRunRequest(req types.RunRequest) string {
	if req.Language == "" {
		return "language is required"
	}

	if req.Source == "" {
		return "source is required"
	}

	if len(req.Source) > maxSourceBytes {
		return "source is too large"
	}

	if len(req.Tests) == 0 {
		return "at least one test is required"
	}

	if len(req.Tests) > maxTests {
		return "too many tests"
	}

	for _, test := range req.Tests {
		if len(test.Stdin) > maxTestInputBytes {
			return "test stdin is too large"
		}

		if len(test.ExpectedStdout) > maxExpectedBytes {
			return "expected stdout is too large"
		}
	}

	return ""
}

func runPython(tempDir string, req types.RunRequest) map[string]any {
	sourcePath := filepath.Join(tempDir, "solution.py")
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		return map[string]any{
			"error": "failed to write source file",
		}
	}

	results := make([]map[string]any, 0, len(req.Tests))
	overallStatus := "accepted"

	for _, test := range req.Tests {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
		cmd := sandboxedCommand(ctx, tempDir, pythonCommand(), "solution.py")
		cmd.Dir = tempDir

		stdout := newCappedBuffer(maxCapturedOutputLen)
		stderr := newCappedBuffer(maxCapturedOutputLen)
		cmd.Stdin = strings.NewReader(test.Stdin)
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		testStatus := "accepted"
		if err := cmd.Run(); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				testStatus = "time_limit_exceeded"
			} else {
				testStatus = "runtime_error"
			}
		} else if strings.TrimSpace(stdout.String()) != strings.TrimSpace(test.ExpectedStdout) {
			testStatus = "wrong_output"
		}
		cancel()

		if testStatus != "accepted" && overallStatus == "accepted" {
			overallStatus = testStatus
		}

		results = append(results, map[string]any{
			"status":           testStatus,
			"stdout":           stdout.String(),
			"stderr":           stderr.String(),
			"stdout_truncated": stdout.Truncated(),
			"stderr_truncated": stderr.Truncated(),
			"duration_ms":      time.Since(start).Milliseconds(),
		})
	}

	return map[string]any{
		"status": overallStatus,
		"tests":  results,
	}
}

type cappedBuffer struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func newCappedBuffer(limit int) *cappedBuffer {
	return &cappedBuffer{
		limit: limit,
	}
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	remaining := b.limit - b.buffer.Len()
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}

	if len(p) > remaining {
		b.buffer.Write(p[:remaining])
		b.truncated = true
		return len(p), nil
	}

	b.buffer.Write(p)
	return len(p), nil
}

func (b *cappedBuffer) String() string {
	if b.truncated {
		return b.buffer.String() + outputTruncatedMarker
	}

	return b.buffer.String()
}

func (b *cappedBuffer) Truncated() bool {
	return b.truncated
}

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

func pythonCommand() string {
	if runtime.GOOS == "windows" {
		return "python"
	}

	return "python3"
}

func runCpp(tempDir string, req types.RunRequest) map[string]any {
	sourcePath := filepath.Join(tempDir, "solution.cpp")
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		return map[string]any{
			"error": "failed to write source file",
		}
	}

	binaryName := "solution"
	if runtime.GOOS == "windows" {
		binaryName = "solution.exe"
	}

	buildStart := time.Now()
	buildCtx, buildCancel := context.WithTimeout(context.Background(), buildTimeout)
	defer buildCancel()
	buildCmd := exec.CommandContext(buildCtx, "g++", "solution.cpp", "-o", binaryName)
	buildCmd.Dir = tempDir

	buildStdout := newCappedBuffer(maxCapturedOutputLen)
	buildStderr := newCappedBuffer(maxCapturedOutputLen)
	buildCmd.Stdout = buildStdout
	buildCmd.Stderr = buildStderr

	if err := buildCmd.Run(); err != nil {
		buildStatus := "failed"
		if buildCtx.Err() == context.DeadlineExceeded {
			buildStatus = "time_limit_exceeded"
		}

		results := make([]map[string]any, 0, len(req.Tests))
		for range req.Tests {
			results = append(results, map[string]any{
				"status":      "not_executed",
				"stdout":      "",
				"stderr":      "",
				"duration_ms": int64(0),
			})
		}

		return map[string]any{
			"status": "build_failed",
			"build": map[string]any{
				"status":           buildStatus,
				"stdout":           buildStdout.String(),
				"stderr":           buildStderr.String(),
				"stdout_truncated": buildStdout.Truncated(),
				"stderr_truncated": buildStderr.Truncated(),
				"duration_ms":      time.Since(buildStart).Milliseconds(),
			},
			"tests": results,
		}
	}

	results := make([]map[string]any, 0, len(req.Tests))
	overallStatus := "accepted"
	executable := "." + string(os.PathSeparator) + binaryName

	for _, test := range req.Tests {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), runTimeout)
		cmd := sandboxedCommand(ctx, tempDir, executable)
		cmd.Dir = tempDir

		stdout := newCappedBuffer(maxCapturedOutputLen)
		stderr := newCappedBuffer(maxCapturedOutputLen)
		cmd.Stdin = strings.NewReader(test.Stdin)
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		testStatus := "accepted"
		if err := cmd.Run(); err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				testStatus = "time_limit_exceeded"
			} else {
				testStatus = "runtime_error"
			}
		} else if strings.TrimSpace(stdout.String()) != strings.TrimSpace(test.ExpectedStdout) {
			testStatus = "wrong_output"
		}
		cancel()

		if testStatus != "accepted" && overallStatus == "accepted" {
			overallStatus = testStatus
		}

		results = append(results, map[string]any{
			"status":           testStatus,
			"stdout":           stdout.String(),
			"stderr":           stderr.String(),
			"stdout_truncated": stdout.Truncated(),
			"stderr_truncated": stderr.Truncated(),
			"duration_ms":      time.Since(start).Milliseconds(),
		})
	}

	return map[string]any{
		"status": overallStatus,
		"build": map[string]any{
			"status":           "ok",
			"stdout":           buildStdout.String(),
			"stderr":           buildStderr.String(),
			"stdout_truncated": buildStdout.Truncated(),
			"stderr_truncated": buildStderr.Truncated(),
			"duration_ms":      time.Since(buildStart).Milliseconds(),
		},
		"tests": results,
	}
}

func main() {
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/readyz", readyz)
	http.HandleFunc("/info", infoHandler)
	http.HandleFunc("/run", runHandler)

	log.Println("server running on :8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
