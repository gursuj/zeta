package tools

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

// shellDesc names the shell in the bash tool description.
const shellDesc = "the user's shell: Git Bash if installed (or $SHELL), else PowerShell, else cmd.exe"

type shellKind int

const (
	kindBash shellKind = iota
	kindPowerShell
	kindCmd
)

// runShell runs script in the best shell found, from dir, sending stdout and
// stderr to out. ctx cancellation kills the shell and everything it started.
func runShell(ctx context.Context, script, dir string, out io.Writer) error {
	c := shellCommand(ctx, script)
	c.Dir = dir
	c.Stdout = out
	c.Stderr = out

	// By default Go kills only the shell, so a child like `go test` keeps
	// running and holds the output pipe open, leaving Wait blocked. A job object
	// tracks the whole tree, even through Git Bash's own process model, which
	// taskkill /T cannot follow.
	var job atomic.Uintptr
	c.Cancel = func() error {
		if h := windows.Handle(job.Load()); h != 0 {
			if err := windows.TerminateJobObject(h, 1); err == nil {
				return nil
			}
		}
		return c.Process.Kill()
	}
	// Backstop: if something still holds the pipe, stop waiting for it.
	c.WaitDelay = 2 * time.Second

	if err := c.Start(); err != nil {
		return err
	}
	if h, err := newJob(c.Process.Pid); err == nil {
		job.Store(uintptr(h))
		defer windows.CloseHandle(h)
	}
	return c.Wait()
}

// newJob creates a job object holding the process pid. Children it starts
// afterwards join the job automatically.
func newJob(pid int) (windows.Handle, error) {
	h, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return 0, err
	}
	p, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		_ = windows.CloseHandle(h)
		return 0, err
	}
	defer windows.CloseHandle(p)
	if err := windows.AssignProcessToJobObject(h, p); err != nil {
		_ = windows.CloseHandle(h)
		return 0, err
	}
	return h, nil
}

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
