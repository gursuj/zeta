//go:build !windows

package update

import "os"

// swapIn moves the finished temp file over path. A running binary can be
// replaced by rename on unix, since the old inode stays alive until it exits.
func swapIn(tmp, path string) error {
	return os.Rename(tmp, path)
}
