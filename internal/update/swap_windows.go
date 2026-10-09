package update

import (
	"errors"
	"os"
)

// swapIn moves the finished temp file over path. Windows will not overwrite or
// delete a running .exe, but it does allow renaming one. So the live binary is
// moved aside to path+".old" first, and the new one takes its name.
func swapIn(tmp, path string) error {
	old := path + ".old"
	// A leftover from the previous update is no longer running by now.
	_ = os.Remove(old)
	if err := os.Rename(path, old); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		// Put the old binary back so a failed update leaves zeta usable.
		_ = os.Rename(old, path)
		return err
	}
	return nil
}
