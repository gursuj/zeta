package tools

import (
	"runtime"
	"testing"
)

// On Windows a stray \r paints over the row in the TUI, so CRLF is flattened.
// Elsewhere output is passed through untouched.
func TestNormaliseOutput(t *testing.T) {
	in := "a\r\nb\r\nc\rd"
	want := in
	if runtime.GOOS == "windows" {
		want = "a\nb\nc\rd" // lone \r (progress bars) is kept
	}
	if got := normaliseOutput(in); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
