package main

import (
	"bytes"
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

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func readyz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ready",
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

	var req types.RunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid JSON request",
		})
		return
	}

	if req.Language == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "language is required",
		})
		return
	}

	if req.Source == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "source is required",
		})
		return
	}

	if len(req.Tests) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "at least one test is required",
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
		cmd := exec.Command(pythonCommand(), "solution.py")
		cmd.Dir = tempDir

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdin = strings.NewReader(test.Stdin)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		testStatus := "accepted"
		if err := cmd.Run(); err != nil {
			testStatus = "runtime_error"
		} else if strings.TrimSpace(stdout.String()) != strings.TrimSpace(test.ExpectedStdout) {
			testStatus = "wrong_output"
		}

		if testStatus != "accepted" && overallStatus == "accepted" {
			overallStatus = testStatus
		}

		results = append(results, map[string]any{
			"status":      testStatus,
			"stdout":      stdout.String(),
			"stderr":      stderr.String(),
			"duration_ms": time.Since(start).Milliseconds(),
		})
	}

	return map[string]any{
		"status": overallStatus,
		"tests":  results,
	}
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
	buildCmd := exec.Command("g++", "solution.cpp", "-o", binaryName)
	buildCmd.Dir = tempDir

	var buildStdout bytes.Buffer
	var buildStderr bytes.Buffer
	buildCmd.Stdout = &buildStdout
	buildCmd.Stderr = &buildStderr

	if err := buildCmd.Run(); err != nil {
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
				"status":      "failed",
				"stdout":      buildStdout.String(),
				"stderr":      buildStderr.String(),
				"duration_ms": time.Since(buildStart).Milliseconds(),
			},
			"tests": results,
		}
	}

	results := make([]map[string]any, 0, len(req.Tests))
	overallStatus := "accepted"
	executable := "." + string(os.PathSeparator) + binaryName

	for _, test := range req.Tests {
		start := time.Now()
		cmd := exec.Command(executable)
		cmd.Dir = tempDir

		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdin = strings.NewReader(test.Stdin)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		testStatus := "accepted"
		if err := cmd.Run(); err != nil {
			testStatus = "runtime_error"
		} else if strings.TrimSpace(stdout.String()) != strings.TrimSpace(test.ExpectedStdout) {
			testStatus = "wrong_output"
		}

		if testStatus != "accepted" && overallStatus == "accepted" {
			overallStatus = testStatus
		}

		results = append(results, map[string]any{
			"status":      testStatus,
			"stdout":      stdout.String(),
			"stderr":      stderr.String(),
			"duration_ms": time.Since(start).Milliseconds(),
		})
	}

	return map[string]any{
		"status": overallStatus,
		"build": map[string]any{
			"status":      "ok",
			"stdout":      buildStdout.String(),
			"stderr":      buildStderr.String(),
			"duration_ms": time.Since(buildStart).Milliseconds(),
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
