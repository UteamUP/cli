package scanner

import (
	"encoding/binary"
	"hash/crc32"
	"os"
	"path/filepath"
	"testing"
)

func TestPerceptualHashRejectsOversizedPNGHeader(t *testing.T) {
	data := []byte("\x89PNG\r\n\x1a\n")
	chunk := make([]byte, 25)
	binary.BigEndian.PutUint32(chunk, 13)
	copy(chunk[4:], "IHDR")
	binary.BigEndian.PutUint32(chunk[8:], 20000)
	binary.BigEndian.PutUint32(chunk[12:], 20000)
	chunk[16] = 8
	chunk[17] = 2
	binary.BigEndian.PutUint32(chunk[21:], crc32.ChecksumIEEE(chunk[4:21]))
	data = append(data, chunk...)
	path := filepath.Join(t.TempDir(), "bomb.png")
	os.WriteFile(path, data, 0600)
	if _, err := ComputePerceptualHash(path); err == nil {
		t.Fatal("oversized dimensions accepted")
	}
}
