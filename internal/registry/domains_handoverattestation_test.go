package registry

import (
	"strings"
	"testing"
)

func findHandoverAttestationDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "handoverattestation" {
			return d
		}
	}
	t.Fatal("expected handoverattestation domain to be registered")
	return nil
}

func TestHandoverAttestationDomainRegistered(t *testing.T) {
	d := findHandoverAttestationDomain(t)
	if d.APIPath != "/api/handoverattestation" {
		t.Errorf("handoverattestation APIPath = %q, want %q", d.APIPath, "/api/handoverattestation")
	}
}

func TestHandoverAttestationActionsWired(t *testing.T) {
	d := findHandoverAttestationDomain(t)
	byName := map[string]*Action{}
	for i := range d.Actions {
		byName[d.Actions[i].Name] = &d.Actions[i]
	}

	issue, ok := byName["issue"]
	if !ok || issue.HTTPMethod != "POST" || issue.RESTPath != "{handover-guid}/issue" {
		t.Fatalf("issue must be POST \"{handover-guid}/issue\", got %+v", issue)
	}
	if len(issue.Args) != 1 || issue.Args[0].Name != "handover-guid" || issue.Args[0].Type != "uuid" {
		t.Errorf("issue must take a uuid 'handover-guid' arg, got %+v", issue.Args)
	}
	if !issue.DisableResponseExport {
		t.Error("issue must disable response export because it returns transfer secrets")
	}

	redeem, ok := byName["redeem"]
	if !ok || redeem.HTTPMethod != "POST" || redeem.RESTPath != "redeem" {
		t.Fatalf("redeem must be POST \"redeem\", got %+v", redeem)
	}
	var tokenFlag *FlagDef
	for i := range redeem.Flags {
		if redeem.Flags[i].Name == "token" {
			tokenFlag = &redeem.Flags[i]
		}
	}
	if tokenFlag == nil || !tokenFlag.Required || !tokenFlag.Sensitive {
		t.Errorf("redeem must have a required sensitive 'token' flag, got %+v", redeem.Flags)
	}

	redeemCode, ok := byName["redeem-code"]
	if !ok || redeemCode.HTTPMethod != "POST" || redeemCode.RESTPath != "redeem-code" {
		t.Fatalf("redeem-code must be POST \"redeem-code\", got %+v", redeemCode)
	}
	if redeemCode.ToolName != "UteamupHandoverAttestationRedeemCode" {
		t.Errorf("redeem-code tool = %q", redeemCode.ToolName)
	}
	if !strings.Contains(redeemCode.Description, "5 attempts/minute") || !strings.Contains(redeemCode.Description, "0 AI credits") || !strings.Contains(redeemCode.Description, "review only") {
		t.Errorf("redeem-code description must disclose limits, credit cost, and review-only behavior: %q", redeemCode.Description)
	}
	var codeFlag *FlagDef
	for i := range redeemCode.Flags {
		if redeemCode.Flags[i].Name == "code" {
			codeFlag = &redeemCode.Flags[i]
		}
	}
	if codeFlag == nil || !codeFlag.Required || !codeFlag.Sensitive || codeFlag.BodyName != "code" {
		t.Errorf("redeem-code must have a required sensitive 'code' body flag, got %+v", redeemCode.Flags)
	}

	verify, ok := byName["verify"]
	if !ok || verify.HTTPMethod != "POST" || verify.RESTPath != "redeem" {
		t.Fatalf("verify compatibility alias must redeem atomically, got %+v", verify)
	}
	if verify.ToolName != redeem.ToolName {
		t.Errorf("verify alias tool = %q, want %q", verify.ToolName, redeem.ToolName)
	}
}
