package main

import (
	"errors"
	"os"
	"os/exec"

	"github.com/axispx/zeta/internal/update"
)

// relaunch starts the just-installed zeta as a child and exits with its status.
// Windows has no exec(2) to swap the process image, so this process stays alive
// as a thin parent until the child quits. Only returns on failure to start.
func relaunch() error {
	exe, err := update.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	err = cmd.Run()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		os.Exit(ee.ExitCode())
	}
	if err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
