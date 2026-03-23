package fileutil

import (
	"fmt"
	"os"
)

// MIMEType represents detected file MIME type.
type MIMEType string

const (
	MIMETypeMP4         MIMEType = "video/mp4"
	MIMETypeMOV         MIMEType = "video/quicktime"
	MIMETypeGIF         MIMEType = "image/gif"
	MIMETypeUnsupported MIMEType = "unsupported"
)

// mp4Brands contains the set of known MP4 ftyp brand codes.
var mp4Brands = map[string]bool{
	"isom": true,
	"mp41": true,
	"mp42": true,
	"M4V ": true,
	"avc1": true,
	"mp71": true,
	"MSNV": true,
	"iso2": true,
	"iso3": true,
	"iso4": true,
	"iso5": true,
	"iso6": true,
	"f4v ": true,
}

// DetectMIME reads the first 12 bytes of a file and detects the MIME type
// via magic byte signatures.
//
// MP4: ftyp at offset 4 with brands isom, mp41, mp42, M4V, avc1, mp71, MSNV,
// iso2, iso3, iso4, iso5, iso6, f4v.
// MOV: ftyp at offset 4 with brand "qt  " (QuickTime).
// GIF: starts with GIF87a or GIF89a.
func DetectMIME(path string) (MIMEType, error) {
	f, err := os.Open(path)
	if err != nil {
		return MIMETypeUnsupported, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	buf := make([]byte, 12)
	n, err := f.Read(buf)
	if err != nil {
		return MIMETypeUnsupported, fmt.Errorf("read file header: %w", err)
	}
	if n == 0 {
		return MIMETypeUnsupported, fmt.Errorf("file is empty")
	}

	// GIF detection: first 6 bytes are "GIF87a" or "GIF89a"
	if n >= 6 {
		header := string(buf[:6])
		if header == "GIF87a" || header == "GIF89a" {
			return MIMETypeGIF, nil
		}
	}

	// ftyp detection: bytes 4-7 are "ftyp", bytes 8-11 are the brand
	if n >= 12 {
		ftyp := string(buf[4:8])
		if ftyp == "ftyp" {
			brand := string(buf[8:12])
			if brand == "qt  " {
				return MIMETypeMOV, nil
			}
			if mp4Brands[brand] {
				return MIMETypeMP4, nil
			}
		}
	}

	return MIMETypeUnsupported, nil
}

// IsVideo returns true if the MIME type is a supported video format (MP4 or MOV).
func IsVideo(mt MIMEType) bool {
	return mt == MIMETypeMP4 || mt == MIMETypeMOV
}

// IsGIF returns true if the MIME type is GIF.
func IsGIF(mt MIMEType) bool {
	return mt == MIMETypeGIF
}

// IsUnsupported returns true if the MIME type is not supported.
func IsUnsupported(mt MIMEType) bool {
	return mt == MIMETypeUnsupported
}
