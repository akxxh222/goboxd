package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
)

func TestRunMethodNotAllowed(t *testing.T) {
	mux := NewMux()
	req := httptest.NewRequest(http.MethodGet, "/run", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestRunBodyTooLarge(t *testing.T) {
	mux := NewMux()
	jsonBody := strings.Repeat("x", config.MaxRequestBodyBytes+1)
	body := bytes.NewBufferString(`{"language":"py3","source":"` + jsonBody + `","tests":[{"stdin":"","expected_stdout":""}]}`)
	req := httptest.NewRequest(http.MethodPost, "/run", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if !strings.Contains(w.Body.String(), "error") {
		t.Fatalf("expected error response, got %q", w.Body.String())
	}
}
