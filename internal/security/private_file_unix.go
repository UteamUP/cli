//go:build !windows

package security

import (
	"fmt"
	"os"
)

func validatePrivateFile(_ *os.File, info os.FileInfo) error {
	if info.Mode().Perm()&0077 != 0 {
		return fmt.Errorf("private file must be owner-only")
	}
	return nil
}
