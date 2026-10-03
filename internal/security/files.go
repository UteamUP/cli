package security

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var sourceIdentities sync.Map

func BindSource(path string, info os.FileInfo) {
	absolute, _ := filepath.Abs(path)
	sourceIdentities.Store(absolute, info)
}

// OpenRegular validates the opened descriptor, so a pathname swap cannot redirect a read.
func OpenRegular(path string, maximum int64) (*os.File, error) {
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	absolute, _ := filepath.Abs(path)
	if expected, ok := sourceIdentities.Load(absolute); ok {
		original := expected.(os.FileInfo)
		if !os.SameFile(original, before) || original.Size() != before.Size() || !original.ModTime().Equal(before.ModTime()) {
			return nil, fmt.Errorf("scanned source identity changed")
		}
	}
	if !before.Mode().IsRegular() || before.Size() > maximum {
		return nil, fmt.Errorf("source is not a bounded regular file")
	}
	file, err := openSource(path)
	if err != nil {
		return nil, err
	}
	after, err := file.Stat()
	if err != nil || !after.Mode().IsRegular() || !os.SameFile(before, after) || after.Size() > maximum {
		file.Close()
		return nil, fmt.Errorf("source identity changed")
	}
	return file, nil
}
func ReadFile(path string, maximum int64) ([]byte, error) {
	file, err := OpenRegular(path, maximum)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ReadLimit(file, maximum)
}

// openDirectory anchors every traversal step to a descriptor and rejects symbolic links.
func openDirectory(path string) (*os.Root, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	// macOS ships these root-owned aliases; all user-controlled components still
	// undergo descriptor-anchored, no-symlink traversal below.
	if runtime.GOOS == "darwin" {
		for _, alias := range []string{"/var", "/tmp"} {
			if absolute == alias || strings.HasPrefix(absolute, alias+"/") {
				target, err := os.Readlink(alias)
				if err == nil && target == "private"+alias {
					absolute = "/private" + absolute
				}
			}
		}
	}
	volume := filepath.VolumeName(absolute)
	root, err := os.OpenRoot(volume + string(filepath.Separator))
	if err != nil {
		return nil, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(absolute, volume+string(filepath.Separator)), string(filepath.Separator)) {
		if part == "" {
			continue
		}
		info, err := root.Lstat(part)
		if err != nil || !info.IsDir() {
			root.Close()
			return nil, fmt.Errorf("output parent must be an existing directory without symlinks")
		}
		next, err := root.OpenRoot(part)
		if err != nil {
			root.Close()
			return nil, err
		}
		actual, err := next.Stat(".")
		root.Close()
		if err != nil || !os.SameFile(info, actual) {
			next.Close()
			return nil, fmt.Errorf("output directory identity changed")
		}
		root = next
	}
	return root, nil
}

// WriteFile atomically replaces a regular destination without following its links.
func WriteFile(path string, data []byte) error {
	root, err := openDirectory(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer root.Close()
	name := filepath.Base(path)
	info, err := root.Lstat(name)
	if err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("output destination is not a regular file")
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return err
	}
	temporary := ".uteamup-output-" + hex.EncodeToString(random)
	file, err := root.OpenFile(temporary, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer root.Remove(temporary)
	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return root.Rename(temporary, name)
}

// CopyFile uses the same replacement boundary for generated streaming content.
func CopyFile(path string, source io.Reader, maximum int64) error {
	bytes, err := ReadLimit(source, maximum)
	if err != nil {
		return err
	}
	return WriteFile(path, bytes)
}
