package security

import (
	"os"
	"path/filepath"
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
	if info.Mode().Perm() != 0600 {
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
	if _, err := SecretFile(path); err == nil {
		t.Fatal("public secret file accepted")
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
