package tools

import (
	"context"
	"strings"
	"testing"
)

// A stray \r makes the terminal paint over the row, so CRLF must not reach the
// transcript or the model.
func TestBashNormalisesCRLF(t *testing.T) {
	out := Run(context.Background(), Build(), t.TempDir(), Bash, mustRaw(t, map[string]any{
		"command": `printf 'a\r\nb\r\n'`,
	}))
	if strings.Contains(out, "\r") {
		t.Fatalf("CR left in output: %q", out)
	}
	if !strings.HasPrefix(out, "a\nb\n") {
		t.Fatalf("got %q", out)
	}
}
