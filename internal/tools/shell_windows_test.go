package tools

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestKindOf(t *testing.T) {
	cases := map[string]shellKind{
		`C:\Program Files\Git\bin\bash.exe`:                         kindBash,
		`C:\Program Files\PowerShell\7\pwsh.exe`:                    kindPowerShell,
		`C:\Windows\System32\WindowsPowerShell\v1.0\POWERSHELL.EXE`: kindPowerShell,
		`C:\Windows\System32\cmd.exe`:                               kindCmd,
	}
	for path, want := range cases {
		if got := kindOf(path); got != want {
			t.Errorf("kindOf(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestFindShellSkipsWSLBash(t *testing.T) {
	t.Setenv("SHELL", "")
	path, kind := findShell()
	if kind == kindBash {
		if root := os.Getenv("SystemRoot"); strings.HasPrefix(strings.ToLower(path), strings.ToLower(root)+`\`) {
			t.Fatalf("picked the WSL launcher: %s", path)
		}
	}
	if path == "" {
		t.Fatal("no shell found")
	}
}

// A timed-out command must take its child processes with it. Without the
// tree kill, the orphaned sleep holds the output pipe and Run blocks ~30s.
func TestBashTimeoutKillsChildren(t *testing.T) {
	root := t.TempDir()
	start := time.Now()
	out := Run(context.Background(), Build(), root, Bash, mustRaw(t, map[string]any{
		"command": "sleep 30; echo done", "timeout_ms": 500,
	}))
	if !strings.Contains(out, "exit: timeout") {
		t.Fatalf("want timeout, got %q", out)
	}
	if d := time.Since(start); d > 10*time.Second {
		t.Fatalf("took %s, child likely survived", d)
	}
}
