//go:build !unix

package security

import "os"

func openSource(path string) (*os.File, error) { return os.Open(path) }
