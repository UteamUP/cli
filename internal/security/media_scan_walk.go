package security

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	mediaScanReadDirBatch = 128
	maximumMediaScanDepth = 64
)

// WalkMediaTree visits regular directory entries in bounded batches while
// anchoring each directory read to an os.Root. Directories are counted even
// when recursion is disabled.
func WalkMediaTree(
	path string,
	recursive bool,
	budget *MediaScanBudget,
	visit func(path string, entry fs.DirEntry) error,
) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	root, err := OpenDirectory(absolute)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := budget.AddEntry(); err != nil {
		return err
	}
	return walkMediaDirectory(root, path, 0, recursive, budget, visit)
}

func walkMediaDirectory(
	root *os.Root,
	path string,
	depth int,
	recursive bool,
	budget *MediaScanBudget,
	visit func(path string, entry fs.DirEntry) error,
) error {
	directory, err := root.Open(".")
	if err != nil {
		return err
	}
	defer directory.Close()

	for {
		entries, readErr := directory.ReadDir(mediaScanReadDirBatch)
		for _, entry := range entries {
			if err := budget.AddEntry(); err != nil {
				return err
			}
			entryPath := filepath.Join(path, entry.Name())
			if !entry.IsDir() {
				if err := visit(entryPath, entry); err != nil {
					return err
				}
				continue
			}
			if !recursive {
				continue
			}
			if depth+1 >= maximumMediaScanDepth {
				return fmt.Errorf("media scan exceeds the %d-directory depth limit", maximumMediaScanDepth)
			}
			if err := walkMediaChild(root, entry.Name(), entryPath, depth+1, recursive, budget, visit); err != nil {
				return err
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return readErr
		}
	}
}

func walkMediaChild(
	parent *os.Root,
	name string,
	path string,
	depth int,
	recursive bool,
	budget *MediaScanBudget,
	visit func(path string, entry fs.DirEntry) error,
) error {
	before, err := parent.Lstat(name)
	if err != nil {
		return err
	}
	if !before.IsDir() {
		return nil
	}
	child, err := parent.OpenRoot(name)
	if err != nil {
		return err
	}
	defer child.Close()
	actual, err := child.Stat(".")
	if err != nil || !os.SameFile(before, actual) {
		return fmt.Errorf("media scan directory identity changed")
	}
	after, err := parent.Lstat(name)
	if err != nil || !after.IsDir() || !os.SameFile(before, after) {
		return fmt.Errorf("media scan directory identity changed")
	}
	return walkMediaDirectory(child, path, depth, recursive, budget, visit)
}
