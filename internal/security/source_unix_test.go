//go:build unix

package security

import (
	"golang.org/x/sys/unix"
	"path/filepath"
	"testing"
)

func TestOpenSourceDoesNotBlockOnFIFO(t *testing.T) {
	path := filepath.Join(t.TempDir(), "payload")
	if err := unix.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	file, err := openSource(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().IsRegular() {
		t.Fatal("FIFO treated as regular")
	}
}
