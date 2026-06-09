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

func TestVerilogExecution(t *testing.T) {
	cases := []struct {
		name               string
		req                types.RunRequest
		expectedStatus     string
		expectedBuildState string
	}{
		{
			name: "accepted",
			req: types.RunRequest{
				Language: "verilog",
				Source:   "module main;\ninitial begin\n$display(\"Hello Verilog\");\n$finish;\nend\nendmodule",
				Tests: []types.TestCase{
					{
						Stdin:          "",
						ExpectedStdout: "Hello Verilog\n",
					},
				},
			},
			expectedStatus:     "accepted",
			expectedBuildState: "ok",
		},
		{
			name: "build_failed",
			req: types.RunRequest{
				Language: "verilog",
				Source:   "module main;\ninitial begin\n$display(\"Hello Verilog\") // missing semicolon\nend\nendmodule",
				Tests: []types.TestCase{
					{
						Stdin:          "",
						ExpectedStdout: "",
					},
				},
			},
			expectedStatus:     "build_failed",
			expectedBuildState: "failed",
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
			if runResp.Build != nil && runResp.Build.Status != tc.expectedBuildState {
				t.Errorf("Expected build status %q, got %q", tc.expectedBuildState, runResp.Build.Status)
			}
		})
	}
}