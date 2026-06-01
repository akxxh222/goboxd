package runner

import (
	"reflect"
	"strings"
	"testing"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/types"
)

func TestFirstNonAccepted(t *testing.T) {
	cases := []struct {
		current string
		next    string
		want    string
	}{
		{current: "accepted", next: "accepted", want: "accepted"},
		{current: "accepted", next: "wrong_output", want: "wrong_output"},
		{current: "wrong_output", next: "accepted", want: "wrong_output"},
		{current: "runtime_error", next: "time_limit_exceeded", want: "runtime_error"},
	}

	for _, tc := range cases {
		got := firstNonAccepted(tc.current, tc.next)
		if got != tc.want {
			t.Fatalf("firstNonAccepted(%q, %q) = %q, want %q", tc.current, tc.next, got, tc.want)
		}
	}
}

func TestCappedBufferTruncates(t *testing.T) {
	buf := newCappedBuffer(10)
	_, err := buf.Write([]byte("0123456789ABCDEF"))
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if !buf.Truncated() {
		t.Fatal("expected buffer to be truncated")
	}

	got := buf.String()
	if !strings.HasSuffix(got, config.OutputTruncatedMarker) {
		t.Fatalf("expected truncated marker, got %q", got)
	}

	content := strings.TrimSuffix(got, config.OutputTruncatedMarker)
	if len(content) != 10 {
		t.Fatalf("expected truncated content length 10, got %d", len(content))
	}
}

func TestRunRequestTypeDoesNotExposeFilename(t *testing.T) {
	rt := reflect.TypeOf(types.RunRequest{})
	if _, ok := rt.FieldByName("Filename"); ok {
		t.Fatal("RunRequest should not expose a Filename field")
	}
}
