package scanner

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/uteamup/cli/internal/security"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"

	"github.com/corona10/goimagehash"

	"github.com/uteamup/cli/internal/imageanalyzer/imageutil"
)

const hashChunkSize = 8192 // 8 KiB

// ComputeSHA256 computes the SHA-256 hash of the file at filePath,
// reading in streaming 8 KiB chunks. Returns the lowercase hex string.
func ComputeSHA256(filePath string) (string, error) {
	f, err := security.OpenRegular(filePath, 100*1024*1024)
	if err != nil {
		return "", fmt.Errorf("open file for sha256: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	buf := make([]byte, hashChunkSize)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			h.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read file for sha256: %w", err)
		}
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ComputePerceptualHash computes an average perceptual hash of the image
// at filePath using goimagehash. For HEIC files it converts to JPEG
// first via imageutil. Returns the hex string of the hash.
//
// If the image cannot be decoded for perceptual hashing (e.g. corrupt or
// unsupported sub-format), it returns an empty string and nil error --
// this mirrors the Python behaviour where phash failure is non-fatal.
func ComputePerceptualHash(filePath string) (string, error) {
	var img image.Image

	if imageutil.IsHEIC(filePath) {
		jpegBytes, err := imageutil.ConvertHEICToJPEG(filePath)
		if err != nil {
			log.Printf("perceptual hash: HEIC conversion failed for %s: %v", filePath, err)
			return "", nil
		}
		if err := imageutil.ValidateImageHeader(bytes.NewReader(jpegBytes)); err != nil {
			return "", err
		}
		decoded, _, err := image.Decode(bytes.NewReader(jpegBytes))
		if err != nil {
			log.Printf("perceptual hash: decode converted HEIC failed for %s: %v", filePath, err)
			return "", nil
		}
		img = decoded
	} else {
		f, err := security.OpenRegular(filePath, 100*1024*1024)
		if err != nil {
			return "", fmt.Errorf("open file for phash: %w", err)
		}
		defer f.Close()

		if err := imageutil.ValidateImageHeader(f); err != nil {
			return "", err
		}
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return "", err
		}
		decoded, _, err := image.Decode(f)
		if err != nil {
			log.Printf("perceptual hash: decode failed for %s: %v", filePath, err)
			return "", nil
		}
		img = decoded
	}

	hash, err := goimagehash.AverageHash(img)
	if err != nil {
		log.Printf("perceptual hash: average hash failed for %s: %v", filePath, err)
		return "", nil
	}

	return fmt.Sprintf("%016x", hash.GetHash()), nil
}
