package security

import (
	"strings"
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/types"
)

func TestValidateRunRequest(t *testing.T) {
	validTest := types.TestCase{Stdin: "1 2\n", ExpectedStdout: "3\n"}
	validReq := types.RunRequest{Language: "cpp", Source: "int main(){}", Tests: []types.TestCase{validTest}}

	cases := []struct {
		name string
		req  types.RunRequest
		want string
	}{
		{name: "valid request", req: validReq, want: ""},
		{name: "missing language", req: types.RunRequest{Source: "x", Tests: []types.TestCase{validTest}}, want: "language is required"},
		{name: "missing source", req: types.RunRequest{Language: "cpp", Tests: []types.TestCase{validTest}}, want: "source is required"},
		{name: "source too large", req: types.RunRequest{Language: "cpp", Source: strings.Repeat("a", config.MaxSourceBytes+1), Tests: []types.TestCase{validTest}}, want: "source is too large"},
		{name: "no tests", req: types.RunRequest{Language: "cpp", Source: "int main(){}", Tests: nil}, want: "at least one test is required"},
		{name: "too many tests", req: types.RunRequest{Language: "cpp", Source: "int main(){}", Tests: make([]types.TestCase, config.MaxTests+1)}, want: "too many tests"},
		{name: "stdin too large", req: types.RunRequest{Language: "cpp", Source: "int main(){}", Tests: []types.TestCase{{Stdin: strings.Repeat("a", config.MaxTestInputBytes+1), ExpectedStdout: ""}}}, want: "test stdin is too large"},
		{name: "expected stdout too large", req: types.RunRequest{Language: "cpp", Source: "int main(){}", Tests: []types.TestCase{{Stdin: "", ExpectedStdout: strings.Repeat("a", config.MaxExpectedBytes+1)}}}, want: "expected stdout is too large"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
		got := ValidateRunRequest(tc.req)
		if got != tc.want {
			t.Fatalf("expected %q, got %q", tc.want, got)
		}
		if tc.want == "" && got != "" {
			t.Fatalf("expected no error for valid request, got %q", got)
		}
	})
	}
}
