package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/runner"
	"github.com/thesouldev/goboxd/internal/security"
	"github.com/thesouldev/goboxd/internal/types"
)

// runSemaphore limits the number of concurrent execution environments.
// 6 concurrent runs * ~300MB per Java run = 1.8GB, safely under the 2GB limit.
// Prevents OOM kills and enables graceful degradation via queuing.
var runSemaphore = make(chan struct{}, 6)

func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthz)
	mux.HandleFunc("/readyz", readyz)
	mux.HandleFunc("/info", info)
	mux.HandleFunc("/run", run)

	return mux
}

func healthz(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func readyz(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	checks := map[string]string{
		"nsjail": "ok",
	}

	if _, err := exec.LookPath("nsjail"); err != nil {
		checks["nsjail"] = "missing"
	}

	for _, def := range runner.Registry {
		if len(def.BuildCmd) > 0 {
			cmd := def.BuildCmd[0]
			if !strings.HasPrefix(cmd, "./") && !strings.HasPrefix(cmd, "/") {
				if _, err := exec.LookPath(cmd); err != nil {
					checks[cmd] = "missing"
				} else {
					checks[cmd] = "ok"
				}
			}
		}
		if len(def.RunCmd) > 0 {
			cmd := def.RunCmd[0]
			if !strings.HasPrefix(cmd, "./") && !strings.HasPrefix(cmd, "/") {
				if _, err := exec.LookPath(cmd); err != nil {
					checks[cmd] = "missing"
				} else {
					checks[cmd] = "ok"
				}
			}
		}
	}

	if len(runner.Registry) == 0 {
		legacyCmds := []string{runner.PythonCommand(), "g++", "gcc", "javac", "java", "bash", "node", "iverilog", "vvp"}
		for _, cmd := range legacyCmds {
			if _, err := exec.LookPath(cmd); err != nil {
				checks[cmd] = "missing"
			} else {
				checks[cmd] = "ok"
			}
		}
	}

	status := "ready"
	statusCode := http.StatusOK
	for _, check := range checks {
		if check != "ok" {
			status = "not_ready"
			statusCode = http.StatusServiceUnavailable
			break
		}
	}

	writeJSON(w, statusCode, map[string]any{
		"status": status,
		"checks": checks,
	})
}

func info(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodGet) {
		return
	}

	var langs []map[string]string
	for id, def := range runner.Registry {
		langs = append(langs, map[string]string{
			"id":   id,
			"name": def.Name,
		})
	}

	if len(langs) == 0 {
		langs = []map[string]string{
			{"id": "py3", "name": "Python 3"},
			{"id": "cpp", "name": "C++"},
			{"id": "c", "name": "C"},
			{"id": "java", "name": "Java"},
			{"id": "bash", "name": "Bash"},
			{"id": "node", "name": "JavaScript (Node.js)"},
			{"id": "verilog", "name": "Verilog"},
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"name": "goboxd",
		"languages": langs,
		"endpoints": []string{
			"GET /healthz",
			"GET /readyz",
			"GET /info",
			"POST /run",
		},
		"limits": map[string]any{
			"build_timeout_ms":       config.BuildTimeout.Milliseconds(),
			"run_timeout_ms":         config.RunTimeout.Milliseconds(),
			"max_request_body_bytes": config.MaxRequestBodyBytes,
			"max_source_bytes":       config.MaxSourceBytes,
			"max_tests":              config.MaxTests,
			"max_test_input_bytes":   config.MaxTestInputBytes,
			"max_expected_bytes":     config.MaxExpectedBytes,
			"max_captured_output":    config.MaxCapturedOutputLen,
			"sandbox": map[string]string{
				"cpu_seconds":      config.SandboxCPUSeconds,
				"address_space_mb": config.SandboxAddressSpaceMB,
				"file_size_mb":     config.SandboxFileSizeMB,
				"open_files":       config.SandboxOpenFiles,
				"processes":        config.SandboxProcesses,
			},
		},
	})
}

func run(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
		return
	}

	// Enforce 10-second request timeout (including queue time)
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Acquire concurrency slot for graceful degradation
	select {
	case runSemaphore <- struct{}{}:
		defer func() { <-runSemaphore }()
	case <-ctx.Done():
		// Request queued too long
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"error": "server at capacity, request queued too long",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, config.MaxRequestBodyBytes)

	var req types.RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		message := "invalid JSON request"
		if strings.Contains(err.Error(), "request body too large") {
			message = "request body too large"
		}

		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": message,
		})
		return
	}

	if errorMessage := security.ValidateRunRequest(req); errorMessage != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": errorMessage,
		})
		return
	}

	tempDir, err := os.MkdirTemp("", "goboxd-*")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to create temp directory",
		})
		return
	}
	defer os.RemoveAll(tempDir)

	if err := os.Chmod(tempDir, 0755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to prepare temp directory",
		})
		return
	}

	response, ok := runner.Run(tempDir, req)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "failed to execute runner",
		})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}

	w.Header().Set("Allow", method)
	writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
		"error": "method not allowed",
	})
	return false
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(value)
}
