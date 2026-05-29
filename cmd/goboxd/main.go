package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/thesouldev/goboxd/internal/types"
)
func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
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

	sourcePath := tempDir + "/solution.py"
	if err := os.WriteFile(sourcePath, []byte(req.Source), 0644); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "failed to write source file",
		})
		return
	}

	results := make([]map[string]string, 0, len(req.Tests))
	overallStatus := "accepted"

	for _, test := range req.Tests {
		cmd := exec.Command("python", "solution.py")
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

		results = append(results, map[string]string{
			"status": testStatus,
			"stdout": stdout.String(),
			"stderr": stderr.String(),
		})
	}

	json.NewEncoder(w).Encode(map[string]any{
		"status": overallStatus,
		"tests":  results,
	})
}

func main() {
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/run", runHandler)

	log.Println("server running on :8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
