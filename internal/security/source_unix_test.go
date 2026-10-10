//go:build unix

package security

import (
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenSourceDoesNotBlockOnFIFO(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "payload")
	if err := unix.Mkfifo(path, 0600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	file, err := openSource(root, "payload")
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
