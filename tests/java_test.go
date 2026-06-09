//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/thesouldev/goboxd/internal/types"
)

func TestJavaExecution(t *testing.T) {
	cases := []struct {
		name               string
		source             string
		stdin              string
		expectedStdout     string
		expectedStatus     string
		expectedBuildState string
		expectedTestState  string
	}{
		{
			name: "accepted",
			source: `public class Main {
    public static void main(String[] args) {
        System.out.print("hi");
    }
}`,
			stdin:              "",
			expectedStdout:     "hi",
			expectedStatus:     "accepted",
			expectedBuildState: "ok",
			expectedTestState:  "accepted",
		},
		{
			name: "build_failed",
			source: `public class Main {
    public static void main(String[] args) {
        System.out.print("hi") // missing semicolon
    }
}`,
			stdin:              "",
			expectedStdout:     "hi",
			expectedStatus:     "build_failed",
			expectedBuildState: "failed",
			expectedTestState:  "not_executed",
		},
		{
			name: "runtime_error",
			source: `public class Main {
    public static void main(String[] args) {
        int[] arr = new int[2];
        System.out.print(arr[5]);
    }
}`,
			stdin:              "",
			expectedStdout:     "",
			expectedStatus:     "runtime_error",
			expectedBuildState: "ok",
			expectedTestState:  "runtime_error",
		},
		{
			name: "stdin_handling",
			source: `import java.util.Scanner;
public class Main {
    public static void main(String[] args) {
        Scanner scanner = new Scanner(System.in);
        if (scanner.hasNextLine()) {
            System.out.print(scanner.nextLine());
        }
    }
}`,
			stdin:              "hello java",
			expectedStdout:     "hello java",
			expectedStatus:     "accepted",
			expectedBuildState: "ok",
			expectedTestState:  "accepted",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := types.RunRequest{
				Language: "java",
				Source:   tc.source,
				Tests: []types.TestCase{
					{
						Stdin:          tc.stdin,
						ExpectedStdout: tc.expectedStdout,
					},
				},
			}

			apiUrl := os.Getenv("API_URL")
			if apiUrl == "" {
				apiUrl = "http://localhost:8080"
			}

			jsonData, _ := json.Marshal(reqBody)
			resp, err := http.Post(apiUrl+"/run", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				t.Fatalf("Failed to make POST request: %v", err)
			}
			defer resp.Body.Close()

			var runResp types.RunResponse
			if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if runResp.Status != tc.expectedStatus {
				t.Errorf("Expected overall status '%s', got '%s'", tc.expectedStatus, runResp.Status)
			}
			if runResp.Build != nil && runResp.Build.Status != tc.expectedBuildState {
				t.Errorf("Expected build status '%s', got '%s'", tc.expectedBuildState, runResp.Build.Status)
			}
			if len(runResp.Tests) > 0 && runResp.Tests[0].Status != tc.expectedTestState {
				t.Errorf("Expected test status '%s', got '%s'", tc.expectedTestState, runResp.Tests[0].Status)
			}
		})
	}
}
