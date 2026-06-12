package main

import (
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/httpapi"
	"github.com/thesouldev/goboxd/internal/runner"
)

func main() {
	if runtime.GOOS != "linux" {
		log.Fatalf("goboxd only supports running on Linux (current OS: %s)", runtime.GOOS)
	}
	if _, err := exec.LookPath("nsjail"); err != nil {
		log.Fatalf("nsjail is not installed or not in PATH: %v", err)
	}

	registryPath := os.Getenv("LANGUAGES_PATH")
	if registryPath == "" {
		registryPath = "languages.yaml"
	}
	if err := runner.LoadRegistry(registryPath); err != nil {
		log.Printf("warning: failed to load %s: %v", registryPath, err)
	}

	if removed, err := runner.CleanupStaleTempDirs(os.TempDir(), config.StaleTempDirAge); err != nil {
		log.Printf("stale temp dir cleanup failed: %v", err)
	} else if removed > 0 {
		log.Printf("removed %d stale temp dirs", removed)
	}

	log.Println("server running on :8080")

	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewMux(),
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
