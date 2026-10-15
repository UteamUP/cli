package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func vendorComplianceDeviceDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "vendor-compliance-device" {
			return domain
		}
	}
	t.Fatal("vendor-compliance-device domain missing")
	return nil
}

func TestVendorComplianceDeviceDiagnosticsPreserveOriginalAndHideSensitiveFields(t *testing.T) {
	for _, action := range vendorComplianceDeviceDomain(t).Actions {
		if err := validateActionDefinition(action); err != nil {
			t.Fatal(err)
		}
		original := " \n{\"operationGuid\":\"original\"}\n "
		args := map[string]any{"originalJson": original, "signature": "signature", "schemaVersion": 1}
		logged, _ := redactSensitiveActionValues(action, args, nil)
		if logged["originalJson"] != "[REDACTED]" || logged["signature"] != "[REDACTED]" || logged["schemaVersion"] != 1 {
			t.Fatal("unsafe diagnostic", logged)
		}
		if args["originalJson"] != original || args["signature"] != "signature" {
			t.Fatal("redaction changed signed request")
		}
	}
}

func TestVendorComplianceDeviceProtocolPostsExactRootDTO(t *testing.T) {
	for _, step := range []string{"proof", "register", "deliver"} {
		for _, operation := range []string{"create", "update", "delete", "review"} {
			t.Run(step+"/"+operation, func(t *testing.T) {
				home := t.TempDir()
				t.Setenv("HOME", home)
				t.Setenv("USERPROFILE", home)
				body := map[string]any{"schemaVersion": float64(1), "originalJson": " \n{\"retained\":true}\n ", "deviceGuid": "11111111-1111-4111-8111-111111111111", "enrollmentVersion": float64(7), "deviceKeyThumbprint": "thumbprint"}
				if step != "proof" {
					body["signature"] = "signature"
				}
				payload, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err)
				}
				file := filepath.Join(home, "request.json")
				if err := os.WriteFile(file, payload, 0600); err != nil {
					t.Fatal(err)
				}
				requests := 0
				server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests++
					if r.Method != "POST" || r.URL.Path != "/api/v1/vendor/compliance/reviewed/"+operation+"/"+step || r.URL.RawQuery != "" {
						t.Errorf("unexpected route %s %s", r.Method, r.URL)
					}
					var actual map[string]any
					if err := json.NewDecoder(r.Body).Decode(&actual); err != nil {
						t.Error(err)
					}
					if !reflect.DeepEqual(actual, body) {
						t.Errorf("signed DTO changed: %#v", actual)
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"schemaVersion":1}`))
				}))
				defer server.Close()
				if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "contract-token", ExpiresAt: time.Now().Add(time.Hour), TenantGUID: "55555555-5555-4555-8555-555555555555"}); err != nil {
					t.Fatal(err)
				}
				api := client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logging.New(logging.LevelError))
				format := "json"
				cmd := buildDomainCommand(vendorComplianceDeviceDomain(t), func() (*client.APIClient, error) { return api, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
				args := []string{step, operation, "--from-json", file}
				if step == "deliver" {
					args = append(args, "--confirm")
				}
				cmd.SetArgs(args)
				if err := cmd.Execute(); err != nil || requests != 1 {
					t.Fatalf("delivery err=%v requests=%d", err, requests)
				}
			})
		}
	}
}

func TestVendorComplianceDeviceRejectsUnconfirmedOrUnknownActionBeforeTransport(t *testing.T) {
	for _, args := range [][]string{{"deliver", "delete", "--from-json", "missing.json"}, {"deliver", "delete", "--from-json", "missing.json", "--confirm=false"}, {"proof", "admin", "--from-json", "missing.json"}} {
		created := false
		format := "json"
		cmd := buildDomainCommand(vendorComplianceDeviceDomain(t), func() (*client.APIClient, error) { created = true; return nil, nil }, logging.New(logging.LevelError), &format, &ExportConfig{})
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		cmd.SetArgs(args)
		if err := cmd.Execute(); err == nil || created {
			t.Fatalf("invalid request reached transport: %v", args)
		}
	}
}
