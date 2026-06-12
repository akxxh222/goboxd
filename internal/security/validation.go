package security

import (
	"strings"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/types"
)

func ValidateRunRequest(req types.RunRequest) string {
	if req.Language == "" {
		return "language is required"
	}

	if req.Source == "" {
		return "source is required"
	}

	if len(req.Source) > config.MaxSourceBytes {
		return "source is too large"
	}

	if len(req.Tests) == 0 {
		return "at least one test is required"
	}

	if len(req.Tests) > config.MaxTests {
		return "too many tests"
	}

	for _, test := range req.Tests {
		if len(test.Stdin) > config.MaxTestInputBytes {
			return "test stdin is too large"
		}

		if len(test.ExpectedStdout) > config.MaxExpectedBytes {
			return "expected stdout is too large"
		}
	}

	for _, flag := range req.BuildFlags {
		if !strings.HasPrefix(flag, "-") {
			return "build flags must start with - or --"
		}
		if strings.ContainsAny(flag, ";&|<>$\n\r") {
			return "invalid character in build flag"
		}
	}

	for _, flag := range req.RunFlags {
		if !strings.HasPrefix(flag, "-") {
			return "run flags must start with - or --"
		}
		if strings.ContainsAny(flag, ";&|<>$\n\r") {
			return "invalid character in run flag"
		}
	}

	return ""
}
