//go:build unix

package security

import (
	"golang.org/x/sys/unix"
	"os"
)

func openSource(root *os.Root, name string) (*os.File, error) {
	return root.OpenFile(name, os.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
}
