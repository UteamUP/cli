package client

import (
	"context"
	"encoding/json"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/logging"
)

func parseForm(t *testing.T, body []byte, contentType string) map[string][]string {
	t.Helper()
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "multipart/form-data" {
		t.Fatalf("content type = %q (%v), want multipart/form-data", contentType, err)
	}
	form, err := multipart.NewReader(strings.NewReader(string(body)), params["boundary"]).ReadForm(1 << 20)
	if err != nil {
		t.Fatal(err)
	}
	return form.Value
}

func TestEncodeFormFieldsMatchesTheFromFormBinder(t *testing.T) {
	body, contentType, err := EncodeFormFields(map[string]any{
		"id":                   "ignored",
		"title":                "ChatGPT Team",
		"type":                 json.Number("5"),
		"totalValue":           json.Number("4990.50"),
		"isAutoRenewal":        true,
		"userGuids":            []any{"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"},
		"responsibleUserGuid":  nil,
		"billingFrequencyDays": float64(30),
	})
	if err != nil {
		t.Fatal(err)
	}

	got := parseForm(t, body, contentType)
	want := map[string][]string{
		"Title":                {"ChatGPT Team"},
		"Type":                 {"5"},
		"TotalValue":           {"4990.50"},
		"IsAutoRenewal":        {"true"},
		"UserGuids":            {"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"},
		"BillingFrequencyDays": {"30"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("form fields = %v, want %v", got, want)
	}
}

func TestEncodeFormFieldsRejectsNestedObjectsInsteadOfDroppingThem(t *testing.T) {
	_, _, err := EncodeFormFields(map[string]any{"vendor": map[string]any{"name": "OpenAI"}})
	if err == nil || !strings.Contains(err.Error(), `"vendor"`) {
		t.Fatalf("err = %v, want a nested-object error naming the field", err)
	}
	_, _, err = EncodeFormFields(map[string]any{"assets": []any{map[string]any{"guid": "x"}}})
	if err == nil {
		t.Fatal("a list of objects must be rejected too")
	}
}

func TestContractPayerFormDistinguishesExplicitClearFromOmission(t *testing.T) {
	for _, fields := range []map[string]any{
		{"title": "Legacy agreement"},
		{"payerType": "", "payerName": nil},
		{"payerType": "External", "payerName": "Named payer"},
	} {
		body, contentType, err := EncodeFormFields(fields)
		if err != nil {
			t.Fatal(err)
		}
		form := parseForm(t, body, contentType)
		for _, key := range []string{"payerType", "payerName"} {
			want, present := fields[key]
			got, sent := form[pascalCase(key)]
			if !present || want == nil {
				if sent {
					t.Fatalf("omitted %s unexpectedly sent: %v", key, got)
				}
			} else if !reflect.DeepEqual(got, []string{want.(string)}) {
				t.Fatalf("%s = %v, want %q", key, got, want)
			}
		}
	}
}

func TestCallRESTFormSendsMultipartWithAuthTenantAndCSRFHeaders(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	const tenant = "b966b8c7-04a4-45d4-aa51-519ecf2ef13a"

	var seen map[string][]string
	var headers http.Header
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("parsing form: %v", err)
		} else {
			seen = r.MultipartForm.Value
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"externalGuid":"33333333-3333-4333-8333-333333333333"}`))
	}))
	defer server.Close()
	if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "test-token", AuthMethod: "apikey",
		ExpiresAt: time.Now().Add(time.Hour), TenantGUID: tenant}); err != nil {
		t.Fatal(err)
	}
	client := NewAPIClient(server.URL, time.Second, true, RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))

	result, err := client.CallRESTForm(context.Background(), "POST", "/api/contracts",
		map[string]any{"title": "Claude Max", "userGuids": []any{"11111111-1111-4111-8111-111111111111"}}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(result), "33333333-3333-4333-8333-333333333333") {
		t.Errorf("result = %s", result)
	}
	if !reflect.DeepEqual(seen, map[string][]string{"Title": {"Claude Max"}, "UserGuids": {"11111111-1111-4111-8111-111111111111"}}) {
		t.Errorf("server saw %v", seen)
	}
	if headers.Get("Authorization") != "Bearer test-token" || headers.Get("X-Requested-With") != "XMLHttpRequest" ||
		headers.Get("X-Tenant-Guid") != tenant {
		t.Errorf("headers = %v", headers)
	}
}
