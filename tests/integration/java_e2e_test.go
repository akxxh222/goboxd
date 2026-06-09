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

func TestJavaE2E(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("integration tests require Linux / Docker environment")
	}
	if _, err := exec.LookPath("javac"); err != nil {
		t.Skip("javac is not installed")
	}
	if _, err := exec.LookPath("java"); err != nil {
		t.Skip("java is not installed")
	}

	server := httptest.NewServer(httpapi.NewMux())
	defer server.Close()

	request := types.RunRequest{
		Language: "java",
		Source: `public class Main {
	public static void main(String[] args) throws java.io.IOException {
		java.io.BufferedReader reader = new java.io.BufferedReader(new java.io.InputStreamReader(System.in));
		String line = reader.readLine();
		if (line == null) {
			return;
		}
		System.out.print("Received: " + line);
	}
}`,
		Tests: []types.TestCase{
			{
				Stdin:          "hello world\n",
				ExpectedStdout: "Received: hello world",
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
	if strings.TrimSpace(response.Tests[0].Stdout) != "Received: hello world" {
		t.Fatalf("expected stdout 'Received: hello world', got %q", response.Tests[0].Stdout)
	}
}
