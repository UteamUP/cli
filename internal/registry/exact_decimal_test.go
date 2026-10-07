package registry

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestExactDecimalPreservesLiteralAndRejectsUnsupportedInput(t *testing.T) {
	for _, literal := range []string{"2147483647.123456", "0.000001", "1.230000", "-1", "0"} {
		number, err := parseExactDecimal(literal)
		if err != nil {
			t.Fatal(err)
		}
		body, err := json.Marshal(map[string]any{"quantity": number})
		if err != nil || string(body) != `{"quantity":`+literal+`}` {
			t.Fatalf("literal lost: %s %v", body, err)
		}
	}
	for _, literal := range []string{"", "NaN", "Infinity", "1e2", "01", "+1", " 1", "1 ", "true", "null", "{}", "[]", "1.", ".1", "0." + strings.Repeat("1", 29), strings.Repeat("1", 30), strings.Repeat("9", 1000)} {
		if _, err := parseExactDecimal(literal); err == nil {
			t.Fatalf("accepted %q", literal)
		}
	}
}

func TestTransferDecimalReachesActualRESTBodyWithoutFloatConversion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	var captured []byte
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/stock/transfers" {
			t.Errorf("wrong route %s %s", r.Method, r.URL.Path)
		}
		captured, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"transferGroupGuid":"ack"}`))
	}))
	defer server.Close()
	if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "isolated-test-token", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	logger := logging.New(logging.LevelError)
	format := "json"
	factory := func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{MaxRetries: 0}, logger), nil
	}
	action := *findStockAction(t, "transfer")
	command := buildActionCommand(findStockDomain(t), action, factory, logger, &format, nil)
	command.SetArgs([]string{"--stock-item-guid", "11111111-1111-4111-8111-111111111111", "--destination-stock-guid", "22222222-2222-4222-8222-222222222222",
		"--idempotency-key", "33333333-3333-4333-8333-333333333333", "--expected-updated-at", "2026-10-07T12:00:00.1234567Z",
		"--expected-source-catalog-sha256", strings.Repeat("A", 64), "--destination-expectation", "absent",
		"--expected-destination-stock-updated-at", "2026-10-07T12:00:01.1234567Z", "--quantity", "2147483647.123456", "--confirm"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(captured), `"quantity":2147483647.123456`) || strings.Contains(string(captured), `"confirm"`) {
		t.Fatalf("wrong exact body %s", captured)
	}
	if !strings.Contains(string(captured), `"expectedUpdatedAt":"2026-10-07T12:00:00.1234567Z"`) {
		t.Fatal("reviewed seventh tick lost")
	}
}

func TestDirectExecutionCannotBypassDecimalValidation(t *testing.T) {
	format := "json"
	for _, literal := range []string{"NaN", "1e2", "{}", "0." + strings.Repeat("1", 29)} {
		action := Action{Name: "create", Flags: []FlagDef{{Name: "quantity", Type: "decimal", Required: true}}}
		domain := &Domain{Name: "test"}
		command := buildActionCommand(domain, action, nil, nil, nil, nil)
		if err := command.Flags().Set("quantity", literal); err != nil {
			t.Fatal(err)
		}
		// No client/logger is supplied: refusal must occur before any request or logging.
		if err := executeAction(command, nil, domain, action, nil, nil, &format, nil); err == nil {
			t.Fatalf("bypassed decimal validation: %s", literal)
		}
	}
}

func TestExactDecimalDefaultAndDirectMissingInputFailClosed(t *testing.T) {
	format := "json"
	for _, input := range []FlagDef{
		{Name: "quantity", Type: "decimal", Required: true},
		{Name: "quantity", Type: "decimal", Default: 1.5},
		{Name: "quantity", Type: "decimal", Default: "1e2"},
	} {
		action := Action{Name: "create", Flags: []FlagDef{input}}
		domain := &Domain{Name: "test"}
		command := buildActionCommand(domain, action, nil, nil, nil, nil)
		if err := executeAction(command, nil, domain, action, nil, nil, &format, nil); err == nil {
			t.Fatalf("accepted unsupported missing/default decimal: %+v", input)
		}
	}
	action := Action{Name: "create", Flags: []FlagDef{{Name: "quantity", Type: "decimal", Default: "0.000001"}}}
	command := buildActionCommand(&Domain{Name: "test"}, action, nil, nil, nil, nil)
	values, err := exactDecimalFlags(command, action)
	if err != nil || values["quantity"] != json.Number("0.000001") {
		t.Fatalf("lost exact default: %v %v", values, err)
	}
}
