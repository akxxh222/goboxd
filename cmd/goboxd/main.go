package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func main() {
	http.HandleFunc("/healthz", healthz)

	log.Println("server running on :8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}