package scanner

import (
	"log"
	"os"

	"github.com/rwcarlsen/goexif/exif"
)

// ExtractEXIF extracts useful EXIF fields from the image at filePath.
// It returns a map with keys "date_taken", "camera_make", "camera_model"
// when available. Returns an empty map on any error (never fails).
func ExtractEXIF(filePath string) map[string]interface{} {
	result := make(map[string]interface{})

	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("exif: cannot open %s: %v", filePath, err)
		return result
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		// Many images lack EXIF; this is not an error worth logging at
		// default level.
		return result
	}

	// DateTimeOriginal (tag 0x9003)
	if dt, err := x.Get(exif.DateTimeOriginal); err == nil {
		if val, err := dt.StringVal(); err == nil {
			result["date_taken"] = val
		}
	} else if dt, err := x.Get(exif.DateTime); err == nil {
		if val, err := dt.StringVal(); err == nil {
			result["date_taken"] = val
		}
	}

	// Camera Make (tag 0x010F)
	if mk, err := x.Get(exif.Make); err == nil {
		if val, err := mk.StringVal(); err == nil {
			result["camera_make"] = val
		}
	}

	// Camera Model (tag 0x0110)
	if md, err := x.Get(exif.Model); err == nil {
		if val, err := md.StringVal(); err == nil {
			result["camera_model"] = val
		}
	}

	return result
}
