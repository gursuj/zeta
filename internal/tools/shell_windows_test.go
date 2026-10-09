package tools

import (
	"bytes"
	"context"
	"os"
	"os/exec"
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

// The fallback shells go through different argv handling than bash, so run a
// quoted command through each to prove the invocation works.
func TestFallbackShellsRun(t *testing.T) {
	cases := []struct {
		name, shell, script, want string
	}{
		{"pwsh", "pwsh", `Write-Output 'a b'`, "a b"},
		{"powershell", "powershell", `Write-Output 'a b'`, "a b"},
		{"cmd", "cmd", `echo "a b"`, `"a b"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := exec.LookPath(tc.shell); err != nil {
				t.Skipf("%s not installed", tc.shell)
			}
			t.Setenv("SHELL", tc.shell)
			var buf bytes.Buffer
			if err := runShell(context.Background(), tc.script, t.TempDir(), &buf); err != nil {
				t.Fatalf("run: %v (out %q)", err, buf.String())
			}
			if got := strings.TrimSpace(buf.String()); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// PowerShell 7 colours errors even when piped. The escape codes must not reach
// the model.
func TestPowerShellOutputHasNoColourCodes(t *testing.T) {
	for _, shell := range []string{"pwsh", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			if _, err := exec.LookPath(shell); err != nil {
				t.Skipf("%s not installed", shell)
			}
			t.Setenv("SHELL", shell)
			var buf bytes.Buffer
			_ = runShell(context.Background(), "Get-ChildItem -NoSuchParam; Get-Location", t.TempDir(), &buf)
			if strings.Contains(buf.String(), "") {
				t.Fatalf("escape codes in output: %q", buf.String())
			}
			if buf.Len() == 0 {
				t.Fatal("no output")
			}
		})
	}
}
