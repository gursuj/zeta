package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// shellDesc names the shell in the bash tool description.
const shellDesc = "the user's shell: Git Bash if installed (or $SHELL), else PowerShell, else cmd.exe"

type shellKind int

const (
	kindBash shellKind = iota
	kindPowerShell
	kindCmd
)

// shellCommand builds the command that runs script in the best shell found.
func shellCommand(ctx context.Context, script string) *exec.Cmd {
	path, kind := findShell()
	switch kind {
	case kindPowerShell:
		return exec.CommandContext(ctx, path, "-NoProfile", "-NonInteractive", "-Command", script)
	case kindCmd:
		c := exec.CommandContext(ctx, path)
		// Go's default argv quoting breaks cmd.exe, which parses its own line.
		// Pass the line raw: /s strips exactly the outer quotes.
		c.SysProcAttr = &syscall.SysProcAttr{CmdLine: `"` + path + `" /d /s /c "` + script + `"`}
		return c
	default:
		return exec.CommandContext(ctx, path, "-c", script)
	}
}

// findShell picks a shell. Bash comes first because the permission parser
// (internal/policy) splits commands with POSIX rules.
func findShell() (string, shellKind) {
	if s := strings.TrimSpace(os.Getenv("SHELL")); s != "" {
		if p := resolveExe(s); p != "" {
			return p, kindOf(p)
		}
	}
	if p := gitBash(); p != "" {
		return p, kindBash
	}
	for _, name := range []string{"pwsh", "powershell"} {
		if p, err := exec.LookPath(name); err == nil {
			return p, kindPowerShell
		}
	}
	if p := os.Getenv("ComSpec"); p != "" {
		return p, kindCmd
	}
	return "cmd.exe", kindCmd
}

func resolveExe(s string) string {
	if filepath.IsAbs(s) {
		if _, err := os.Stat(s); err == nil {
			return s
		}
		return ""
	}
	if p, err := exec.LookPath(s); err == nil {
		return p
	}
	return ""
}

func kindOf(path string) shellKind {
	base := strings.ToLower(strings.TrimSuffix(strings.ToLower(filepath.Base(path)), ".exe"))
	switch base {
	case "pwsh", "powershell":
		return kindPowerShell
	case "cmd":
		return kindCmd
	}
	return kindBash
}

// gitBash finds Git for Windows' bash. A bash.exe in System32 is the WSL
// launcher, which runs commands in Linux, so it is skipped.
func gitBash() string {
	var candidates []string
	for _, env := range []string{"ProgramFiles", "ProgramW6432", "ProgramFiles(x86)"} {
		if dir := os.Getenv(env); dir != "" {
			candidates = append(candidates, filepath.Join(dir, "Git", "bin", "bash.exe"))
		}
	}
	if dir := os.Getenv("LocalAppData"); dir != "" {
		candidates = append(candidates, filepath.Join(dir, "Programs", "Git", "bin", "bash.exe"))
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	p, err := exec.LookPath("bash")
	if err != nil {
		return ""
	}
	if root := os.Getenv("SystemRoot"); root != "" &&
		strings.HasPrefix(strings.ToLower(p), strings.ToLower(root)+`\`) {
		return ""
	}
	return p
}
