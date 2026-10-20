package security

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestPrivateAtomicOutputRejectsSymlinks(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	os.WriteFile(target, []byte("preserve"), 0600)
	link := filepath.Join(dir, "output")
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}
	if WriteFile(link, []byte("changed")) == nil {
		t.Fatal("link accepted")
	}
	got, _ := os.ReadFile(target)
	if string(got) != "preserve" {
		t.Fatal("target overwritten")
	}
	regular := filepath.Join(dir, "regular")
	if err := WriteFile(regular, []byte("ok")); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(regular)
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatal(info.Mode())
	}
}
func TestBoundMediaSourceRejectsReplacementAndSymlink(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input")
	os.WriteFile(path, []byte("original"), 0600)
	info, _ := os.Stat(path)
	BindSource(path, info)
	os.Rename(path, path+".old")
	os.WriteFile(path, []byte("different"), 0600)
	if _, err := ReadFile(path, 100); err == nil {
		t.Fatal("replacement accepted")
	}
	os.Remove(path)
	os.Symlink(path+".old", path)
	if _, err := ReadFile(path, 100); err == nil {
		t.Fatal("symlink accepted")
	}
}
func TestSecretAndTerminalBoundaries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret")
	os.WriteFile(path, []byte("credential"), 0644)
	if runtime.GOOS != "windows" {
		if _, err := SecretFile(path); err == nil {
			t.Fatal("public secret file accepted")
		}
	}
	os.Chmod(path, 0600)
	if got, err := SecretFile(path); err != nil || got != "credential" {
		t.Fatal(err)
	}
	got := SafeText("safe\x1b]52;x\a\r\n\u009b")
	for _, r := range got {
		if r < 32 || r >= 127 && r <= 159 {
			t.Fatal("control survived")
		}
	}
	if !strings.Contains(got, "safe") {
		t.Fatal(got)
	}
}

func TestReadPrivateFileRejectsPublicPermissionsAndSymlinks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "private")
	if err := os.WriteFile(path, []byte("credential"), 0644); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		if _, err := ReadPrivateFile(path, 1024); err == nil {
			t.Fatal("public file accepted")
		}
	}
	if err := os.Chmod(path, 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := ReadPrivateFile(path, 1024); err != nil || string(got) != "credential" {
		t.Fatalf("private file rejected: %v", err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(path, link); err != nil {
		t.Skip(err)
	}
	if _, err := ReadPrivateFile(link, 1024); err == nil {
		t.Fatal("symlink accepted")
	}
}

func TestReadPrivateFilePreservesMissingParentError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing", "secret")
	if _, err := ReadPrivateFile(path, 1024); !os.IsNotExist(err) {
		t.Fatalf("missing parent error = %v, want not-exist", err)
	}
}

func TestRemoveFileRejectsSymlinksAndRemovesRegularFiles(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}
	if err := RemoveFile(link); err == nil {
		t.Fatal("symlink accepted")
	}
	if got, err := os.ReadFile(target); err != nil || string(got) != "preserve" {
		t.Fatalf("symlink target changed: %q, %v", got, err)
	}
	if err := RemoveFile(target); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("regular file remains: %v", err)
	}
}

func TestSourceRejectsIntermediateDirectorySymlink(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	source := filepath.Join(outside, "input")
	if err := os.WriteFile(source, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "redirect")
	if err := os.Symlink(outside, link); err != nil {
		t.Skip(err)
	}
	if _, err := ReadFile(filepath.Join(link, "input"), 100); err == nil {
		t.Fatal("intermediate symlink followed")
	}
	if got, err := ReadFile(source, 100); err != nil || string(got) != "private" {
		t.Fatalf("ordinary bounded file rejected: %v", err)
	}
}
