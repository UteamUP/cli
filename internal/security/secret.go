package security

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func SecretFile(path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if info.Mode().Perm()&0077 != 0 {
		return "", fmt.Errorf("secret file must be owner-only")
	}
	bytes, err := ReadFile(path, 64*1024)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(bytes)), nil
}
func SecretStdin(reader io.Reader) (string, error) {
	bytes, err := ReadLimit(reader, 64*1024)
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(bytes))
	if value == "" {
		return "", fmt.Errorf("secret is empty")
	}
	return value, nil
}
func SafeText(value string) string {
	var b strings.Builder
	for _, c := range value {
		if c < 32 || (c >= 127 && c <= 159) {
			fmt.Fprintf(&b, "\\u%04x", c)
		} else {
			b.WriteRune(c)
		}
	}
	return b.String()
}
