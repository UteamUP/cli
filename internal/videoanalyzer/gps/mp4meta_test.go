package gps

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// makeBox creates an MP4 box with the given type and content.
func makeBox(boxType []byte, content []byte) []byte {
	size := uint32(8 + len(content))
	buf := make([]byte, size)
	binary.BigEndian.PutUint32(buf[0:4], size)
	copy(buf[4:8], boxType)
	copy(buf[8:], content)
	return buf
}

// makeFtypBox creates a minimal ftyp box.
func makeFtypBox() []byte {
	return makeBox([]byte("ftyp"), []byte("isom"))
}

// makeXYZContent creates the content for a ©xyz atom.
// Format: 2-byte string length (big-endian) + 2-byte language (0x0000) + GPS string.
func makeXYZContent(gpsStr string) []byte {
	gpsBytes := []byte(gpsStr)
	content := make([]byte, 4+len(gpsBytes))
	binary.BigEndian.PutUint16(content[0:2], uint16(len(gpsBytes)))
	content[2] = 0x00
	content[3] = 0x00
	copy(content[4:], gpsBytes)
	return content
}

// buildMP4WithXYZ creates a minimal MP4 file with moov/udta/©xyz atom containing the given GPS string.
func buildMP4WithXYZ(t *testing.T, gpsStr string) string {
	t.Helper()

	xyzType := []byte{0xA9, 'x', 'y', 'z'}
	xyzBox := makeBox(xyzType, makeXYZContent(gpsStr))
	udtaBox := makeBox([]byte("udta"), xyzBox)
	moovBox := makeBox([]byte("moov"), udtaBox)

	ftyp := makeFtypBox()

	data := append(ftyp, moovBox...)

	tmpFile := filepath.Join(t.TempDir(), "test_xyz.mp4")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("failed to write test MP4: %v", err)
	}
	return tmpFile
}

// buildMP4WithISO6709 creates a minimal MP4 file with Apple's ISO6709 metadata
// embedded in a way the binary scanner will find.
func buildMP4WithISO6709(t *testing.T, gpsStr string) string {
	t.Helper()

	// Build a keys+ilst structure inside meta inside moov
	// keys atom: version(4) + count(4) + entry(4 size + 4 namespace + key string)
	keyStr := "com.apple.quicktime.location.ISO6709"
	keyEntry := make([]byte, 8+len(keyStr))
	binary.BigEndian.PutUint32(keyEntry[0:4], uint32(8+len(keyStr)))
	copy(keyEntry[4:8], "mdta")
	copy(keyEntry[8:], keyStr)

	keysContent := make([]byte, 8+len(keyEntry))
	// version + flags = 0
	binary.BigEndian.PutUint32(keysContent[4:8], 1) // count = 1
	copy(keysContent[8:], keyEntry)
	keysBox := makeBox([]byte("keys"), keysContent)

	// ilst: contains one entry indexed by key position (1-based)
	// Entry box type is the index as big-endian uint32 (so index 0 -> 0x00000001)
	gpsBytes := []byte(gpsStr)
	dataContent := make([]byte, 8+len(gpsBytes))
	// type indicator (4 bytes) + locale (4 bytes)
	// type 1 = UTF-8
	binary.BigEndian.PutUint32(dataContent[0:4], 1)
	// locale = 0
	copy(dataContent[8:], gpsBytes)
	dataBox := makeBox([]byte("data"), dataContent)

	entryType := make([]byte, 4)
	binary.BigEndian.PutUint32(entryType, 1) // 1-based index
	entryBox := makeBox(entryType, dataBox)

	ilstBox := makeBox([]byte("ilst"), entryBox)

	// meta box: version/flags(4) + children
	metaContent := make([]byte, 4+len(keysBox)+len(ilstBox))
	copy(metaContent[4:], keysBox)
	copy(metaContent[4+len(keysBox):], ilstBox)
	metaBox := makeBox([]byte("meta"), metaContent)

	moovBox := makeBox([]byte("moov"), metaBox)
	ftyp := makeFtypBox()
	data := append(ftyp, moovBox...)

	tmpFile := filepath.Join(t.TempDir(), "test_iso6709.mp4")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("failed to write test MP4: %v", err)
	}
	return tmpFile
}

// buildMP4NoGPS creates a minimal valid MP4 file without GPS data.
func buildMP4NoGPS(t *testing.T) string {
	t.Helper()

	// Just ftyp + moov with an empty udta
	udtaBox := makeBox([]byte("udta"), nil)
	moovBox := makeBox([]byte("moov"), udtaBox)
	ftyp := makeFtypBox()
	data := append(ftyp, moovBox...)

	tmpFile := filepath.Join(t.TempDir(), "test_nogps.mp4")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatalf("failed to write test MP4: %v", err)
	}
	return tmpFile
}

func almostEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func TestExtractGPS_XYZ_PositiveLatPositiveLng(t *testing.T) {
	path := buildMP4WithXYZ(t, "+37.7749+122.4194/")
	data, found, err := ExtractGPS(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected GPS to be found")
	}
	if !almostEqual(data.Latitude, 37.7749, 0.0001) {
		t.Errorf("latitude = %f, want 37.7749", data.Latitude)
	}
	if !almostEqual(data.Longitude, 122.4194, 0.0001) {
		t.Errorf("longitude = %f, want 122.4194", data.Longitude)
	}
}

func TestExtractGPS_XYZ_PositiveLatNegativeLng(t *testing.T) {
	path := buildMP4WithXYZ(t, "+37.7749-122.4194/")
	data, found, err := ExtractGPS(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected GPS to be found")
	}
	if !almostEqual(data.Latitude, 37.7749, 0.0001) {
		t.Errorf("latitude = %f, want 37.7749", data.Latitude)
	}
	if !almostEqual(data.Longitude, -122.4194, 0.0001) {
		t.Errorf("longitude = %f, want -122.4194", data.Longitude)
	}
}

func TestExtractGPS_XYZ_NegativeLatPositiveLng(t *testing.T) {
	path := buildMP4WithXYZ(t, "-33.8688+151.2093/")
	data, found, err := ExtractGPS(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected GPS to be found")
	}
	if !almostEqual(data.Latitude, -33.8688, 0.0001) {
		t.Errorf("latitude = %f, want -33.8688", data.Latitude)
	}
	if !almostEqual(data.Longitude, 151.2093, 0.0001) {
		t.Errorf("longitude = %f, want 151.2093", data.Longitude)
	}
}

func TestExtractGPS_XYZ_WithAltitude(t *testing.T) {
	path := buildMP4WithXYZ(t, "+37.7749-122.4194+026.543/")
	data, found, err := ExtractGPS(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected GPS to be found")
	}
	if !almostEqual(data.Latitude, 37.7749, 0.0001) {
		t.Errorf("latitude = %f, want 37.7749", data.Latitude)
	}
	if !almostEqual(data.Longitude, -122.4194, 0.0001) {
		t.Errorf("longitude = %f, want -122.4194", data.Longitude)
	}
}

func TestExtractGPS_NoGPS(t *testing.T) {
	path := buildMP4NoGPS(t)
	_, found, err := ExtractGPS(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected GPS to not be found")
	}
}

func TestExtractGPS_NonMP4File(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	if err := os.WriteFile(tmpFile, []byte("this is not an mp4 file"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	_, found, err := ExtractGPS(tmpFile)
	if err != nil {
		t.Fatalf("unexpected error for non-MP4 file: %v", err)
	}
	if found {
		t.Error("expected GPS to not be found in text file")
	}
}

func TestExtractGPS_NonexistentFile(t *testing.T) {
	_, _, err := ExtractGPS("/nonexistent/file.mp4")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestExtractGPS_EmptyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "empty.mp4")
	if err := os.WriteFile(tmpFile, []byte{}, 0644); err != nil {
		t.Fatalf("failed to write empty file: %v", err)
	}
	_, found, err := ExtractGPS(tmpFile)
	// Either no error with found=false, or an error — both are acceptable
	if err == nil && found {
		t.Error("expected GPS to not be found in empty file")
	}
}

func TestExtractGPS_ISO6709_AppleFormat(t *testing.T) {
	path := buildMP4WithISO6709(t, "+48.8566+002.3522/")
	data, found, err := ExtractGPS(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !found {
		t.Fatal("expected GPS to be found in ISO6709 format")
	}
	if !almostEqual(data.Latitude, 48.8566, 0.0001) {
		t.Errorf("latitude = %f, want 48.8566", data.Latitude)
	}
	if !almostEqual(data.Longitude, 2.3522, 0.0001) {
		t.Errorf("longitude = %f, want 2.3522", data.Longitude)
	}
}

func TestParseISO6709(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLat float64
		wantLng float64
		wantErr bool
	}{
		{
			name:    "positive lat positive lng",
			input:   "+37.7749+122.4194/",
			wantLat: 37.7749,
			wantLng: 122.4194,
		},
		{
			name:    "positive lat negative lng",
			input:   "+37.7749-122.4194/",
			wantLat: 37.7749,
			wantLng: -122.4194,
		},
		{
			name:    "negative lat positive lng",
			input:   "-33.8688+151.2093/",
			wantLat: -33.8688,
			wantLng: 151.2093,
		},
		{
			name:    "negative lat negative lng",
			input:   "-33.8688-151.2093/",
			wantLat: -33.8688,
			wantLng: -151.2093,
		},
		{
			name:    "with altitude",
			input:   "+37.7749-122.4194+026.543/",
			wantLat: 37.7749,
			wantLng: -122.4194,
		},
		{
			name:    "no trailing slash",
			input:   "+37.7749-122.4194",
			wantLat: 37.7749,
			wantLng: -122.4194,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "no sign prefix",
			input:   "37.7749-122.4194/",
			wantErr: true,
		},
		{
			name:    "only latitude",
			input:   "+37.7749",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lat, lng, err := parseISO6709(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got lat=%f lng=%f", lat, lng)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !almostEqual(lat, tc.wantLat, 0.0001) {
				t.Errorf("latitude = %f, want %f", lat, tc.wantLat)
			}
			if !almostEqual(lng, tc.wantLng, 0.0001) {
				t.Errorf("longitude = %f, want %f", lng, tc.wantLng)
			}
		})
	}
}
