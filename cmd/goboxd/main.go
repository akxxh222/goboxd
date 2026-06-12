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
	// Enforce Linux execution environment since nsjail requires Linux namespaces
	if runtime.GOOS != "linux" {
		log.Fatalf("goboxd only supports running on Linux (current OS: %s)", runtime.GOOS)
	}
	// Ensure the nsjail binary is available in the system PATH before starting
	if _, err := exec.LookPath("nsjail"); err != nil {
		log.Fatalf("nsjail is not installed or not in PATH: %v", err)
	}

	// Load the dynamic language registry (YAML) for plug-and-play language support
	registryPath := os.Getenv("LANGUAGES_PATH")
	if registryPath == "" {
		registryPath = "languages.yaml"
	}
	if err := runner.LoadRegistry(registryPath); err != nil {
		log.Printf("warning: failed to load %s: %v", registryPath, err)
	}

	// Security: Clean up any orphaned temporary directories from previous crashed runs
	if removed, err := runner.CleanupStaleTempDirs(os.TempDir(), config.StaleTempDirAge); err != nil {
		log.Printf("stale temp dir cleanup failed: %v", err)
	} else if removed > 0 {
		log.Printf("removed %d stale temp dirs", removed)
	}

	log.Println("server running on :8080")

	// Configure and start the HTTP server with strict timeouts to prevent connection exhaustion
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
