package gps

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// GPSData holds extracted GPS coordinates from video metadata.
type GPSData struct {
	Latitude  float64
	Longitude float64
}

// ExtractGPS parses MP4/MOV container metadata to find GPS coordinates.
// It supports two common formats:
//   - Apple QuickTime: com.apple.quicktime.location.ISO6709 (e.g., "+37.7749-122.4194+000.000/")
//   - Android/Generic: ©xyz atom (e.g., "+37.7749-122.4194/")
//
// Returns the GPS data, whether GPS was found, and any error.
func ExtractGPS(path string) (data GPSData, found bool, err error) {
	f, err := os.Open(path)
	if err != nil {
		return GPSData{}, false, fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return GPSData{}, false, fmt.Errorf("stat file: %w", err)
	}

	// Try structured approach: parse moov -> udta -> ©xyz
	data, found, err = parseBoxes(f, info.Size(), []string{"moov"})
	if err != nil {
		return GPSData{}, false, err
	}
	if found {
		return data, true, nil
	}

	// Fallback: scan raw bytes for GPS patterns
	data, found, err = scanForGPS(f, info.Size())
	if err != nil {
		return GPSData{}, false, err
	}
	return data, found, nil
}

// readBoxHeader reads the 8-byte box header (4-byte size + 4-byte type).
// Returns the total box size, the box type string, and any error.
func readBoxHeader(r io.Reader) (size int64, boxType string, err error) {
	var buf [8]byte
	if _, err := io.ReadFull(r, buf[:]); err != nil {
		return 0, "", err
	}

	sz := binary.BigEndian.Uint32(buf[0:4])
	btype := string(buf[4:8])

	if sz == 1 {
		// Extended size: read 8 more bytes for 64-bit size
		var extBuf [8]byte
		if _, err := io.ReadFull(r, extBuf[:]); err != nil {
			return 0, "", err
		}
		return int64(binary.BigEndian.Uint64(extBuf[:])), btype, nil
	}

	return int64(sz), btype, nil
}

// headerSize returns the header size for a box given its declared size.
func headerSize(sz int64) int64 {
	if sz == 1 {
		return 16 // 8 (standard) + 8 (extended)
	}
	return 8
}

// xyzBoxType is the ©xyz atom type bytes: 0xA9 followed by 'x', 'y', 'z'.
var xyzBoxType = string([]byte{0xA9, 'x', 'y', 'z'})

// parseBoxes iterates over boxes within the region [current_pos, end) of the ReadSeeker.
// path tracks which container boxes we're looking for (e.g., ["moov"] means we need to find moov first).
func parseBoxes(r io.ReadSeeker, end int64, path []string) (GPSData, bool, error) {
	for {
		pos, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return GPSData{}, false, err
		}
		if pos >= end {
			break
		}

		size, boxType, err := readBoxHeader(r)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return GPSData{}, false, err
		}

		// Determine actual header size consumed
		hdrSize := int64(8)
		if size > 0 && binary.BigEndian.Uint32([]byte{
			byte(size >> 56), byte(size >> 48), byte(size >> 40), byte(size >> 32),
		}) > 0 || size == 0 {
			// This was a standard read; hdrSize is 8
		}
		// Re-check: if original 4-byte size was 1, we read 16 bytes total
		// We need to check the raw value. Since readBoxHeader already consumed
		// the right number of bytes, we compute based on current position.
		currentPos, _ := r.Seek(0, io.SeekCurrent)
		hdrSize = currentPos - pos

		// Handle size == 0 (extends to end of file/container)
		if size == 0 {
			size = end - pos
		}

		contentSize := size - hdrSize
		boxEnd := pos + size

		if len(path) > 0 && boxType == path[0] {
			// Enter this container box
			if len(path) == 1 {
				// We're at the target container, now look for GPS data inside
				return findGPSInContainer(r, boxEnd)
			}
			// Recurse deeper
			return parseBoxes(r, boxEnd, path[1:])
		}

		// Skip this box
		if contentSize > 0 {
			if _, err := r.Seek(boxEnd, io.SeekStart); err != nil {
				return GPSData{}, false, err
			}
		}
	}
	return GPSData{}, false, nil
}

// findGPSInContainer searches within a container (moov) for GPS data.
// It looks for udta -> ©xyz and also meta -> keys+ilst patterns.
func findGPSInContainer(r io.ReadSeeker, end int64) (GPSData, bool, error) {
	for {
		pos, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return GPSData{}, false, err
		}
		if pos >= end {
			break
		}

		size, boxType, err := readBoxHeader(r)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return GPSData{}, false, err
		}

		currentPos, _ := r.Seek(0, io.SeekCurrent)
		hdrSize := currentPos - pos

		if size == 0 {
			size = end - pos
		}

		boxEnd := pos + size
		_ = hdrSize

		switch boxType {
		case "udta":
			// Search inside udta for ©xyz
			data, found, err := findXYZInUDTA(r, boxEnd)
			if err != nil {
				return GPSData{}, false, err
			}
			if found {
				return data, true, nil
			}
		case "meta":
			// meta box has a 4-byte version/flags field after the header
			var versionFlags [4]byte
			if _, err := io.ReadFull(r, versionFlags[:]); err == nil {
				data, found, err := findGPSInMeta(r, boxEnd)
				if err != nil {
					return GPSData{}, false, err
				}
				if found {
					return data, true, nil
				}
			}
		}

		// Skip to next box
		if _, err := r.Seek(boxEnd, io.SeekStart); err != nil {
			return GPSData{}, false, err
		}
	}
	return GPSData{}, false, nil
}

// findXYZInUDTA searches within a udta box for the ©xyz atom.
func findXYZInUDTA(r io.ReadSeeker, end int64) (GPSData, bool, error) {
	for {
		pos, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return GPSData{}, false, err
		}
		if pos >= end {
			break
		}

		size, boxType, err := readBoxHeader(r)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return GPSData{}, false, err
		}

		currentPos, _ := r.Seek(0, io.SeekCurrent)
		hdrSize := currentPos - pos

		if size == 0 {
			size = end - pos
		}

		boxEnd := pos + size

		if boxType == xyzBoxType {
			// Read the ©xyz data: 2-byte string length + 2-byte language + GPS string
			contentSize := size - hdrSize
			if contentSize < 4 {
				break
			}

			// Read the content
			content := make([]byte, contentSize)
			if _, err := io.ReadFull(r, content); err != nil {
				return GPSData{}, false, err
			}

			// Skip 2-byte length + 2-byte language code
			gpsStr := string(content[4:])
			gpsStr = strings.TrimSpace(gpsStr)

			lat, lng, err := parseISO6709(gpsStr)
			if err != nil {
				return GPSData{}, false, err
			}
			return GPSData{Latitude: lat, Longitude: lng}, true, nil
		}

		// Skip to next box
		if _, err := r.Seek(boxEnd, io.SeekStart); err != nil {
			return GPSData{}, false, err
		}
	}
	return GPSData{}, false, nil
}

// findGPSInMeta searches within a meta box for keys+ilst pattern with ISO6709 GPS.
func findGPSInMeta(r io.ReadSeeker, end int64) (GPSData, bool, error) {
	var keys []string
	var ilstPos int64
	var ilstSize int64

	// First pass: find keys and ilst boxes
	for {
		pos, err := r.Seek(0, io.SeekCurrent)
		if err != nil {
			return GPSData{}, false, err
		}
		if pos >= end {
			break
		}

		size, boxType, err := readBoxHeader(r)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return GPSData{}, false, err
		}

		currentPos, _ := r.Seek(0, io.SeekCurrent)
		hdrSize := currentPos - pos

		if size == 0 {
			size = end - pos
		}

		boxEnd := pos + size

		switch boxType {
		case "keys":
			contentSize := size - hdrSize
			if contentSize > 0 {
				content := make([]byte, contentSize)
				if _, err := io.ReadFull(r, content); err == nil {
					keys = parseKeysAtom(content)
				}
			}
		case "ilst":
			ilstPos = currentPos
			ilstSize = size - hdrSize
		}

		if _, err := r.Seek(boxEnd, io.SeekStart); err != nil {
			return GPSData{}, false, err
		}
	}

	// Find the GPS key index
	gpsKeyIndex := -1
	for i, key := range keys {
		if key == "com.apple.quicktime.location.ISO6709" {
			gpsKeyIndex = i
			break
		}
	}

	if gpsKeyIndex < 0 || ilstSize <= 0 {
		return GPSData{}, false, nil
	}

	// Read ilst to get the value at the GPS key index
	if _, err := r.Seek(ilstPos, io.SeekStart); err != nil {
		return GPSData{}, false, err
	}

	ilstEnd := ilstPos + ilstSize
	idx := 0
	for {
		pos, err := r.Seek(0, io.SeekCurrent)
		if err != nil || pos >= ilstEnd {
			break
		}

		size, _, err := readBoxHeader(r)
		if err != nil {
			break
		}

		currentPos, _ := r.Seek(0, io.SeekCurrent)
		hdrSize := currentPos - pos

		if size == 0 {
			size = ilstEnd - pos
		}

		boxEnd := pos + size

		if idx == gpsKeyIndex {
			// Read this ilst entry's data box
			contentSize := size - hdrSize
			if contentSize > 0 {
				content := make([]byte, contentSize)
				if _, err := io.ReadFull(r, content); err == nil {
					// ilst entries contain a "data" sub-box
					gpsStr := extractILSTValue(content)
					if gpsStr != "" {
						lat, lng, err := parseISO6709(gpsStr)
						if err == nil {
							return GPSData{Latitude: lat, Longitude: lng}, true, nil
						}
					}
				}
			}
		}

		idx++
		if _, err := r.Seek(boxEnd, io.SeekStart); err != nil {
			break
		}
	}

	return GPSData{}, false, nil
}

// parseKeysAtom parses the content of a keys atom.
// Format: 4-byte version/flags, 4-byte entry count, then entries:
//
//	each entry: 4-byte key size, 4-byte namespace, key string
func parseKeysAtom(data []byte) []string {
	if len(data) < 8 {
		return nil
	}
	count := binary.BigEndian.Uint32(data[4:8])
	var keys []string
	offset := 8
	for i := uint32(0); i < count; i++ {
		if offset+8 > len(data) {
			break
		}
		keySize := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		// namespace is 4 bytes after size
		if offset+keySize > len(data) || keySize < 8 {
			break
		}
		keyStr := string(data[offset+8 : offset+keySize])
		keys = append(keys, keyStr)
		offset += keySize
	}
	return keys
}

// extractILSTValue extracts the string value from an ilst entry's content.
// The content typically contains a "data" sub-box with: 4-byte size, "data", 4-byte type, 4-byte locale, value.
func extractILSTValue(content []byte) string {
	if len(content) < 16 {
		return ""
	}
	// Look for "data" box
	if string(content[4:8]) == "data" {
		// Skip: 4 size + 4 "data" + 4 type + 4 locale = 16 bytes header
		if len(content) > 16 {
			return strings.TrimSpace(string(content[16:]))
		}
	}
	return ""
}

// scanForGPS scans the first 10MB of the file for GPS coordinate patterns.
func scanForGPS(r io.ReadSeeker, fileSize int64) (GPSData, bool, error) {
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return GPSData{}, false, err
	}

	scanSize := fileSize
	const maxScan = 10 * 1024 * 1024
	if scanSize > maxScan {
		scanSize = maxScan
	}

	buf := make([]byte, scanSize)
	n, err := io.ReadFull(r, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return GPSData{}, false, err
	}
	buf = buf[:n]
	content := string(buf)

	// Look for Apple ISO6709 pattern
	const iso6709Key = "com.apple.quicktime.location.ISO6709"
	if idx := strings.Index(content, iso6709Key); idx >= 0 {
		// Search for ISO 6709 coordinate after the key
		remaining := content[idx+len(iso6709Key):]
		lat, lng, err := findISO6709InString(remaining)
		if err == nil {
			return GPSData{Latitude: lat, Longitude: lng}, true, nil
		}
	}

	// Look for ©xyz pattern (0xA9 + "xyz")
	xyzMarker := string([]byte{0xA9, 'x', 'y', 'z'})
	if idx := strings.Index(content, xyzMarker); idx >= 0 {
		remaining := content[idx+4:]
		if len(remaining) > 4 {
			// Skip 2-byte length + 2-byte language
			remaining = remaining[4:]
			lat, lng, err := findISO6709InString(remaining)
			if err == nil {
				return GPSData{Latitude: lat, Longitude: lng}, true, nil
			}
		}
	}

	return GPSData{}, false, nil
}

// findISO6709InString searches for an ISO 6709 coordinate pattern in a string.
func findISO6709InString(s string) (lat, lng float64, err error) {
	// Look for pattern like +DD.DDDD-DDD.DDDD or +DD.DDDD+DDD.DDDD
	for i := 0; i < len(s) && i < 200; i++ {
		if s[i] == '+' || s[i] == '-' {
			candidate := s[i:]
			if len(candidate) > 5 {
				lat, lng, err := parseISO6709(candidate)
				if err == nil {
					return lat, lng, nil
				}
			}
		}
	}
	return 0, 0, fmt.Errorf("no ISO 6709 coordinate found")
}

// parseISO6709 parses an ISO 6709 coordinate string.
// Format: [+-]DD.DDDD[+-]DDD.DDDD[+-]AAA.AAA/ or without trailing slash.
// Examples:
//
//	"+37.7749-122.4194/"
//	"+37.7749-122.4194+026.543/"
//	"-33.8688+151.2093/"
func parseISO6709(s string) (lat, lng float64, err error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, 0, fmt.Errorf("empty ISO 6709 string")
	}

	// Must start with + or -
	if s[0] != '+' && s[0] != '-' {
		return 0, 0, fmt.Errorf("ISO 6709 string must start with + or -")
	}

	// Find the second +/- sign (start of longitude)
	secondSign := -1
	for i := 1; i < len(s); i++ {
		if s[i] == '+' || s[i] == '-' {
			secondSign = i
			break
		}
	}
	if secondSign < 0 {
		return 0, 0, fmt.Errorf("could not find longitude in ISO 6709 string: %s", s)
	}

	latStr := s[:secondSign]

	// Find the end of longitude: next +/- (altitude) or / or end of string
	remaining := s[secondSign:]
	lngEnd := len(remaining)
	for i := 1; i < len(remaining); i++ {
		if remaining[i] == '+' || remaining[i] == '-' || remaining[i] == '/' {
			lngEnd = i
			break
		}
	}
	lngStr := remaining[:lngEnd]

	lat, err = strconv.ParseFloat(latStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing latitude %q: %w", latStr, err)
	}

	lng, err = strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing longitude %q: %w", lngStr, err)
	}

	return lat, lng, nil
}
