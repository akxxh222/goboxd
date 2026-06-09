//go:build integration
// +build integration

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/thesouldev/goboxd/internal/types"
)

func TestNodeExecution(t *testing.T) {
	cases := []struct {
		name           string
		req            types.RunRequest
		expectedStatus string
	}{
		{
			name: "accepted",
			req: types.RunRequest{
				Language: "node",
				Source:   "console.log('Hello Node');",
				Tests: []types.TestCase{
					{
						Stdin:          "",
						ExpectedStdout: "Hello Node\n",
					},
				},
			},
			expectedStatus: "accepted",
		},
		{
			name: "runtime_error",
			req: types.RunRequest{
				Language: "node",
				Source:   "throw new Error('boom');",
				Tests: []types.TestCase{
					{
						Stdin:          "",
						ExpectedStdout: "",
					},
				},
			},
			expectedStatus: "runtime_error",
		},
		{
			name: "stdin_handling",
			req: types.RunRequest{
				Language: "node",
				Source:   "const fs = require('fs');\nconst input = fs.readFileSync('/dev/stdin', 'utf-8');\nconsole.log(`Got: ${input.trim()}`);",
				Tests: []types.TestCase{
					{
						Stdin:          "input data",
						ExpectedStdout: "Got: input data\n",
					},
				},
			},
			expectedStatus: "accepted",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.req)
			resp, err := http.Post("http://goboxd:8080/run", "application/json", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer resp.Body.Close()

			var runResp types.RunResponse
			json.NewDecoder(resp.Body).Decode(&runResp)
			if runResp.Status != tc.expectedStatus {
				t.Errorf("Expected status %q, got %q", tc.expectedStatus, runResp.Status)
			}
		})
	}
}