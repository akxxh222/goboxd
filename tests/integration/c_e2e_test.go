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

func TestCE2E(t *testing.T) {
    if runtime.GOOS == "windows" {
        t.Skip("integration tests require Linux / Docker environment")
    }
    if _, err := exec.LookPath("gcc"); err != nil {
        t.Skip("gcc is not installed")
    }

    server := httptest.NewServer(httpapi.NewMux())
    defer server.Close()

    request := types.RunRequest{
        Language: "c",
        Source: `#include <stdio.h>
#include <string.h>
int main(void) {
    char buf[64];
    if (fgets(buf, sizeof(buf), stdin) == NULL) {
        return 0;
    }
    size_t len = strlen(buf);
    if (len && buf[len-1] == '\n') {
        buf[len-1] = '\0';
    }
    printf("%s", buf);
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
