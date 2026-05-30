package httpapi

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/runner"
	"github.com/thesouldev/goboxd/internal/security"
	"github.com/thesouldev/goboxd/internal/types"
)

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
		"python": "ok",
		"g++":    "ok",
	}

	if _, err := exec.LookPath("nsjail"); err != nil {
		checks["nsjail"] = "missing"
	}

	if _, err := exec.LookPath(runner.PythonCommand()); err != nil {
		checks["python"] = "missing"
	}

	if _, err := exec.LookPath("g++"); err != nil {
		checks["g++"] = "missing"
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

	writeJSON(w, http.StatusOK, map[string]any{
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

func run(w http.ResponseWriter, r *http.Request) {
	if !requireMethod(w, r, http.MethodPost) {
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
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "unknown language",
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
