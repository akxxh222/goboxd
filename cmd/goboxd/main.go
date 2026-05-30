package main

import (
	"log"
	"net/http"
	"os"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/httpapi"
	"github.com/thesouldev/goboxd/internal/runner"
)

func main() {
	if removed, err := runner.CleanupStaleTempDirs(os.TempDir(), config.StaleTempDirAge); err != nil {
		log.Printf("stale temp dir cleanup failed: %v", err)
	} else if removed > 0 {
		log.Printf("removed %d stale temp dirs", removed)
	}

	log.Println("server running on :8080")

	log.Fatal(http.ListenAndServe(":8080", httpapi.NewMux()))
}
