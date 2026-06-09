package runner

import (
	"bytes"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/types"
)

func Run(tempDir string, req types.RunRequest) (types.RunResponse, bool) {
	switch req.Language {
	case "py3":
		return runPython(tempDir, req), true
	case "cpp":
		return runCpp(tempDir, req), true
	case "c":
		return runC(tempDir, req), true
	case "java":
		return runJava(tempDir, req), true
	case "bash":
		return runBash(tempDir, req), true
	case "node":
		return runNode(tempDir, req), true
	default:
		return types.RunResponse{}, false
	}
}

func firstNonAccepted(current string, next string) string {
	if current == "accepted" && next != "accepted" {
		return next
	}

	return current
}

type cappedBuffer struct {
	buffer    bytes.Buffer
	limit     int
	truncated bool
}

func newCappedBuffer(limit int) *cappedBuffer {
	return &cappedBuffer{
		limit: limit,
	}
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	remaining := b.limit - b.buffer.Len()
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}

	if len(p) > remaining {
		b.buffer.Write(p[:remaining])
		b.truncated = true
		return len(p), nil
	}

	b.buffer.Write(p)
	return len(p), nil
}

func (b *cappedBuffer) String() string {
	if b.truncated {
		return b.buffer.String() + config.OutputTruncatedMarker
	}

	return b.buffer.String()
}

func (b *cappedBuffer) Truncated() bool {
	return b.truncated
}

func notExecutedResults(count int) []types.TestResult {
	results := make([]types.TestResult, 0, count)
	for i := 0; i < count; i++ {
		results = append(results, types.TestResult{
			Status: "not_executed",
		})
	}

	return results
}
