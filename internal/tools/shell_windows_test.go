package tools

import (
	"os"
	"strings"
	"testing"
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
