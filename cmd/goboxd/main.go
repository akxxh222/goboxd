package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
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

	// Configure and start the HTTP server with strict timeouts to prevent connection exhaustion
	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewMux(),
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	// Listen for OS signals to trigger graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the server in a separate goroutine
	go func() {
		log.Println("server running on :8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve error: %v", err)
		}
	}()

	// Block until a shutdown signal is received
	<-stop
	log.Println("shutdown signal received, waiting up to 30s for active runs to finish...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	log.Println("server stopped cleanly")
}
