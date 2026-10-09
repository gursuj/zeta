//go:build !windows

package tools

import (
	"os"
	"os/exec"
	"testing"
)

func TestUserShell(t *testing.T) {
	t.Setenv("SHELL", "/nonexistent/shell")
	got := userShell()
	if got == "/nonexistent/shell" {
		t.Fatalf("should fall back, got %q", got)
	}
	if p, err := exec.LookPath("bash"); err == nil {
		if got != p && got != "sh" {
			t.Fatalf("fallback: %q", got)
		}
	}

	sh := t.TempDir() + "/mysh"
	_ = os.WriteFile(sh, []byte("#!/bin/sh\n"), 0o755)
	t.Setenv("SHELL", sh)
	if got := userShell(); got != sh {
		t.Fatalf("got %q want %q", got, sh)
	}
}
