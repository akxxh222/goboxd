//go:build integration
// +build integration

package integration

import (
    "net/http/httptest"
    "os/exec"
    "runtime"
    "strings"
    "testing"

    "github.com/thesouldev/goboxd/internal/httpapi"
    "github.com/thesouldev/goboxd/internal/types"
)

func TestCppE2E(t *testing.T) {
    if runtime.GOOS == "windows" {
        t.Skip("integration tests require Linux / Docker environment")
    }
    if _, err := exec.LookPath("g++"); err != nil {
        t.Skip("g++ is not installed")
    }

    server := httptest.NewServer(httpapi.NewMux())
    defer server.Close()

    request := types.RunRequest{
        Language: "cpp",
        Source: `#include <iostream>
#include <string>
int main() {
    std::string s;
    std::getline(std::cin, s);
    std::cout << s;
    return 0;
}`,
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
    if response.Build == nil {
        t.Fatal("expected build metadata, got nil")
    }
    if response.Build.Status != "ok" {
        t.Fatalf("expected build status ok, got %q", response.Build.Status)
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
