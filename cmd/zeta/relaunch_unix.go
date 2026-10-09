//go:build !windows

package main

import (
	"os"
	"syscall"

	"github.com/axispx/zeta/internal/update"
)

// relaunch replaces this process with a fresh zeta so the terminal is handed to
// the just-installed binary. Only returns on failure.
func relaunch() error {
	exe, err := update.Executable()
	if err != nil {
		return err
	}
	return syscall.Exec(exe, os.Args, os.Environ())
}
