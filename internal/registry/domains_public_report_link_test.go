package registry

import "testing"

func TestReportLinkDomainIsRoutedAndGuidOnly(t *testing.T) {
	domain := findDomain("report-link")
	if domain == nil {
		t.Fatal("report-link domain is not registered")
	}

	// A domain without APIPath silently 404s at runtime rather than failing to
	// build, so assert the route base explicitly.
	if domain.APIPath != "/api/tenant/reportlinks" {
		t.Fatalf("report-link APIPath = %q", domain.APIPath)
	}

	cases := []struct {
		action string
		method string
		path   string
	}{
		{"list", "GET", ""},
		{"get", "GET", "{guid}"},
		{"create", "POST", ""},
		{"update", "PUT", "{guid}"},
		{"regenerate-token", "POST", "{guid}/regenerate-token"},
		{"delete", "DELETE", "{guid}"},
	}

	for _, tc := range cases {
		action := findAction(domain, tc.action)
		if action == nil {
			t.Fatalf("report-link %s action is missing", tc.action)
		}
		if action.HTTPMethod != tc.method || action.RESTPath != tc.path {
			t.Fatalf("report-link %s route = method %q path %q, want %q %q",
				tc.action, action.HTTPMethod, action.RESTPath, tc.method, tc.path)
		}
		if action.ToolName == "" {
			t.Fatalf("report-link %s has no ToolName, so no MCP tool is exposed", tc.action)
		}
	}
}

func TestReportLinkIdentifiersAreGuidsNotInts(t *testing.T) {
	domain := findDomain("report-link")
	if domain == nil {
		t.Fatal("report-link domain is not registered")
	}

	guidFlagActions := []string{"get", "update", "regenerate-token", "delete"}
	for _, name := range guidFlagActions {
		flags := flagsToMap(findAction(domain, name).Flags)
		flag, ok := flags["guid"]
		if !ok || flag.Type != "uuid" || !flag.Required {
			t.Fatalf("report-link %s guid flag = %+v", name, flag)
		}
		if _, exists := flags["id"]; exists {
			t.Fatalf("report-link %s exposes a legacy int id flag", name)
		}
	}

	createFlags := flagsToMap(findAction(domain, "create").Flags)
	for _, name := range []string{"default-category-guid", "default-location-guid", "default-asset-guid", "notify-team-guid"} {
		if flag, ok := createFlags[name]; !ok || flag.Type != "uuid" {
			t.Fatalf("report-link create %s flag = %+v", name, flag)
		}
	}
}

// The submit endpoint is unauthenticated, captcha-gated and rate-limited by
// design. A scripted authenticated path would be an abuse vector and could not
// pass the captcha regardless, so its absence is intentional and asserted.
func TestReportLinkExposesNoSubmitAction(t *testing.T) {
	domain := findDomain("report-link")
	if domain == nil {
		t.Fatal("report-link domain is not registered")
	}

	for _, forbidden := range []string{"submit", "report", "send", "create-report"} {
		if findAction(domain, forbidden) != nil {
			t.Fatalf("report-link exposes %q — the public submit endpoint must not be scriptable", forbidden)
		}
	}
}

func TestReportLinkContactDefaultsToAnonymous(t *testing.T) {
	domain := findDomain("report-link")
	if domain == nil {
		t.Fatal("report-link domain is not registered")
	}

	flags := flagsToMap(findAction(domain, "create").Flags)
	contact, ok := flags["require-contact-details"]
	if !ok {
		t.Fatal("create is missing require-contact-details")
	}

	// Anonymity is the feature. A default of true here would quietly turn every
	// CLI-created poster into a named-reporting form.
	if contact.Default == true {
		t.Fatal("require-contact-details defaults to true — anonymous reporting must be the default")
	}
}
