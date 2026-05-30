package main

import (
	"log"
	"net/http"

	"github.com/thesouldev/goboxd/internal/httpapi"
)

func main() {
	log.Println("server running on :8080")

	log.Fatal(http.ListenAndServe(":8080", httpapi.NewMux()))
}
