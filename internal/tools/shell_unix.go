//go:build !windows

package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// shellDesc names the shell in the bash tool description.
const shellDesc = "the user's default shell ($SHELL)"

// shellCommand builds the command that runs script in the user's shell.
func shellCommand(ctx context.Context, script string) *exec.Cmd {
	return exec.CommandContext(ctx, userShell(), "-c", script)
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
