//go:build !windows

package checkpoint

import (
	"os"
	"syscall"
)

// processAlive reports whether pid belongs to a running process.
//
// On Unix os.FindProcess always succeeds regardless of whether the pid exists,
// so signal 0 is the actual probe: it runs the kernel's permission and
// existence checks without delivering anything.
func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
