//go:build windows

package checkpoint

import "os"

// processAlive reports whether pid belongs to a running process.
//
// Windows has no signals: (*os.Process).Signal rejects everything except Kill
// with "not supported by windows". Probing with signal 0 therefore reports even
// a live process as dead, which made every lock look stale and let concurrent
// runs delete each other's checkpoint.
//
// os.FindProcess is the real probe here — unlike Unix it calls OpenProcess and
// returns an error when the pid is not running. The handle it opens must be
// released or the process leaks one per check.
func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	_ = proc.Release()
	return true
}
