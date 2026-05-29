package main

import (
	"encoding/json"
	"log"
	"net/http"

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

	json.NewEncoder(w).Encode(map[string]any{
		"language": req.Language,
		"source":   req.Source,
		"tests":    req.Tests,
	})
}

func main() {
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/run", runHandler)

	log.Println("server running on :8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
