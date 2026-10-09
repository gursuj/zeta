//go:build !windows

package tools

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// shellDesc names the shell in the bash tool description.
const shellDesc = "the user's default shell ($SHELL)"

// normaliseOutput leaves command output as the shell wrote it.
func normaliseOutput(s string) string { return s }

// runShell runs script in the user's shell from dir, sending stdout and stderr
// to out. ctx cancellation kills the shell.
func runShell(ctx context.Context, script, dir string, out io.Writer) error {
	c := exec.CommandContext(ctx, userShell(), "-c", script)
	c.Dir = dir
	c.Stdout = out
	c.Stderr = out
	return c.Run()
}

// userShell returns $SHELL when it looks usable, else bash, else sh.
func userShell() string {
	if s := strings.TrimSpace(os.Getenv("SHELL")); s != "" {
		if filepath.IsAbs(s) {
			if _, err := os.Stat(s); err == nil {
				return s
			}
		} else if p, err := exec.LookPath(s); err == nil {
			return p
		}
	}
	if p, err := exec.LookPath("bash"); err == nil {
		return p
	}
	return "sh"
}
