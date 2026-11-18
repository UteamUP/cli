package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

const testJournalGUID = "3f2a9c1e-5b7d-4e8f-9a0b-1c2d3e4f5a6b"

func writeTestImage(t *testing.T, dir, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidateJournalAddImagesInputRejectsBadInput(t *testing.T) {
	dir := t.TempDir()
	png := writeTestImage(t, dir, "photo.png", []byte("\x89PNG\r\n\x1a\nimage"))
	svg := writeTestImage(t, dir, "logo.svg", []byte("<svg/>"))
	pdf := writeTestImage(t, dir, "doc.pdf", []byte("%PDF-1.7"))
	empty := writeTestImage(t, dir, "empty.jpg", nil)

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"no files", []string{testJournalGUID}, "at least one image"},
		{"bad guid", []string{"not-a-guid", png}, "must be a non-empty GUID"},
		{"empty guid", []string{"00000000-0000-0000-0000-000000000000", png}, "must be a non-empty GUID"},
		{"svg", []string{testJournalGUID, svg}, "SVG images are not supported"},
		{"pdf", []string{testJournalGUID, pdf}, "unsupported image type"},
		{"missing", []string{testJournalGUID, filepath.Join(dir, "nope.jpg")}, "file not found"},
		{"empty file", []string{testJournalGUID, empty}, "file is empty"},
		{"directory", []string{testJournalGUID, filepath.Join(dir, "sub.png")}, "not a regular file"},
		{"duplicate", []string{testJournalGUID, png, png}, "listed more than once"},
	}
	if err := os.Mkdir(filepath.Join(dir, "sub.png"), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := validateJournalAddImagesInput(tc.args)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestValidateJournalAddImagesInputRejectsOversizedImage(t *testing.T) {
	dir := t.TempDir()
	big := filepath.Join(dir, "big.jpg")
	f, err := os.Create(big)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Truncate(journalImageMaxBytes + 1); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, _, err := validateJournalAddImagesInput([]string{testJournalGUID, big}); err == nil || !strings.Contains(err.Error(), "25 MB") {
		t.Fatalf("error = %v, want 25 MB limit", err)
	}
}

func newJournalImagesTestClient(t *testing.T, handler http.HandlerFunc) func() (*client.APIClient, error) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "test-token", AuthMethod: "apikey", ExpiresAt: time.Now().Add(time.Hour), TenantGUID: "b966b8c7-04a4-45d4-aa51-519ecf2ef13a"}); err != nil {
		t.Fatal(err)
	}
	return func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, 5*time.Second, true, client.RetryOptions{}, logging.New(logging.LevelError)), nil
	}
}

func TestJournalAddImagesSendsRepeatedImagesParts(t *testing.T) {
	dir := t.TempDir()
	first := writeTestImage(t, dir, "before.png", []byte("\x89PNG\r\n\x1a\nfirst"))
	second := writeTestImage(t, dir, "after.JPG", []byte("\xff\xd8\xffsecond"))

	var gotPath, gotMethod string
	var gotParts []string
	factory := newJournalImagesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		reader, err := r.MultipartReader()
		if err != nil {
			t.Errorf("multipart reader: %v", err)
			return
		}
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("next part: %v", err)
				return
			}
			body, _ := io.ReadAll(part)
			gotParts = append(gotParts, part.FormName()+"|"+part.FileName()+"|"+part.Header.Get("Content-Type")+"|"+string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"journalGuid":"` + testJournalGUID + `","imagesImported":2}`))
	})

	cmd := newJournalAddImagesCmd(factory)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{testJournalGUID, first, second})
	previous := outputFormat
	outputFormat = "json"
	t.Cleanup(func() { outputFormat = previous })
	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if gotMethod != http.MethodPost || gotPath != "/api/journal/"+testJournalGUID+"/images" {
		t.Fatalf("request = %s %s", gotMethod, gotPath)
	}
	want := []string{
		"images|before.png|image/png|\x89PNG\r\n\x1a\nfirst",
		"images|after.JPG|image/jpeg|\xff\xd8\xffsecond",
	}
	if strings.Join(gotParts, "\n") != strings.Join(want, "\n") {
		t.Fatalf("parts = %q, want %q", gotParts, want)
	}
}

func TestJournalAddImagesExplainsServerRejections(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"not found", http.StatusNotFound, `{"error":"Journal not found"}`, "journal entry " + testJournalGUID + " not found"},
		{"svg", http.StatusUnsupportedMediaType, `{"error":"svg_not_supported","detail":"svg"}`, "SVG is not supported"},
		{"mismatch", http.StatusUnsupportedMediaType, `{"error":"mime_magic_byte_mismatch"}`, "does not match a PNG, JPEG, GIF or WebP image"},
		{"forbidden", http.StatusForbidden, `{}`, "Journal.Import and Document.Upload"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			image := writeTestImage(t, dir, "photo.webp", []byte("RIFF0000WEBPdata"))
			factory := newJournalImagesTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.Copy(io.Discard, r.Body)
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			cmd := newJournalAddImagesCmd(factory)
			cmd.SilenceUsage, cmd.SilenceErrors = true, true
			cmd.SetArgs([]string{testJournalGUID, image})
			err := cmd.Execute()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestJournalCommandHasAddImages(t *testing.T) {
	for _, domain := range rootCmd.Commands() {
		if domain.Name() != "journal" {
			continue
		}
		for _, action := range domain.Commands() {
			if action.Name() == "add-images" {
				return
			}
		}
	}
	t.Fatal("journal add-images is not registered")
}
