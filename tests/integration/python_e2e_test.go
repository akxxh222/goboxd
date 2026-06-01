//go:build integration
// +build integration

package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "os/exec"
    "runtime"
    "strings"
    "testing"

    "github.com/thesouldev/goboxd/internal/httpapi"
    "github.com/thesouldev/goboxd/internal/types"
)

func TestPythonE2E(t *testing.T) {
    if runtime.GOOS == "windows" {
        t.Skip("integration tests require Linux / Docker environment")
    }
    if _, err := exec.LookPath("python3"); err != nil {
        t.Skip("python3 is not installed")
    }

    server := httptest.NewServer(httpapi.NewMux())
    defer server.Close()

    request := types.RunRequest{
        Language: "py3",
        Source:   "print(input())",
        Tests: []types.TestCase{
            {
                Stdin:          "hello\n",
                ExpectedStdout: "hello",
            },
        },
    }

    response := postRunResponse(t, server.URL+"/run", request)

    if response.Status != "accepted" {
        t.Fatalf("expected status accepted, got %q", response.Status)
    }
    if len(response.Tests) != 1 {
        t.Fatalf("expected 1 test result, got %d", len(response.Tests))
    }
    if response.Tests[0].Status != "accepted" {
        t.Fatalf("expected test status accepted, got %q", response.Tests[0].Status)
    }
    if strings.TrimSpace(response.Tests[0].Stdout) != "hello" {
        t.Fatalf("expected stdout 'hello', got %q", response.Tests[0].Stdout)
    }
}

func postRunResponse(t *testing.T, url string, request types.RunRequest) types.RunResponse {
    t.Helper()

    payload, err := json.Marshal(request)
    if err != nil {
        t.Fatalf("failed to marshal request: %v", err)
    }

    resp, err := http.Post(url, "application/json", bytes.NewReader(payload))
    if err != nil {
        t.Fatalf("POST /run failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Fatalf("expected HTTP 200, got %d", resp.StatusCode)
    }

    var runResponse types.RunResponse
    if err := json.NewDecoder(resp.Body).Decode(&runResponse); err != nil {
        t.Fatalf("failed to decode response: %v", err)
    }

    return runResponse
}
