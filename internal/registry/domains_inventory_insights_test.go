package registry

import (
	"strings"
	"testing"
)

func TestInventoryInsightsDomainUsesGuidFirstRoutes(t *testing.T) {
	domain := findInventoryInsightsDomain(t)
	if domain.APIPath != "/api/inventory/insights" {
		t.Fatalf("APIPath = %q, want /api/inventory/insights", domain.APIPath)
	}

	expected := map[string]struct {
		method string
		path   string
	}{
		"get":        {method: "GET", path: "{entityType}/by-guid/{entityGuid}"},
		"rca-draft":  {method: "POST", path: "asset/by-guid/{assetGuid}/root-cause-analysis/draft"},
		"rca-create": {method: "POST", path: "asset/by-guid/{assetGuid}/root-cause-analysis"},
	}

	for name, want := range expected {
		action := findInventoryInsightsAction(t, domain, name)
		if action.HTTPMethod != want.method {
			t.Errorf("%s HTTPMethod = %q, want %q", name, action.HTTPMethod, want.method)
		}
		if action.RESTPath != want.path {
			t.Errorf("%s RESTPath = %q, want %q", name, action.RESTPath, want.path)
		}
		if strings.Contains(strings.ToLower(action.RESTPath), "{id}") {
			t.Errorf("%s exposes an integer identity in %q", name, action.RESTPath)
		}
		for _, argument := range action.Args {
			if strings.HasSuffix(argument.Name, "Guid") && argument.Type != "uuid" {
				t.Errorf("%s argument %s must be typed uuid, got %q", name, argument.Name, argument.Type)
			}
		}
	}
}

func TestInventoryInsightsGetConstrainsEntityType(t *testing.T) {
	domain := findInventoryInsightsDomain(t)
	action := findInventoryInsightsAction(t, domain, "get")

	if len(action.Args) != 2 || action.Args[0].Name != "entityType" {
		t.Fatalf("get must take entityType then entityGuid, got %+v", action.Args)
	}
	allowed := strings.Join(action.Args[0].AllowedValues, ",")
	if allowed != "asset,part,tool,chemical" {
		t.Errorf("entityType AllowedValues = %q, want asset,part,tool,chemical", allowed)
	}
	if action.ToolName != "UteamupAssetInsightsGet" {
		t.Errorf("get ToolName = %q, want UteamupAssetInsightsGet", action.ToolName)
	}
}

func TestInventoryInsightsDraftFlagsMapToBodyFields(t *testing.T) {
	domain := findInventoryInsightsDomain(t)
	action := findInventoryInsightsAction(t, domain, "rca-draft")

	bodyNames := map[string]string{}
	for _, flag := range action.Flags {
		bodyNames[flag.Name] = flag.BodyName
	}
	if bodyNames["months-back"] != "monthsBack" {
		t.Errorf("months-back BodyName = %q, want monthsBack", bodyNames["months-back"])
	}
	if bodyNames["focus"] != "focus" {
		t.Errorf("focus BodyName = %q, want focus", bodyNames["focus"])
	}
	if action.ToolName != "UteamupAssetRcaDraft" {
		t.Errorf("rca-draft ToolName = %q, want UteamupAssetRcaDraft", action.ToolName)
	}
}

func findInventoryInsightsDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "insights" {
			return domain
		}
	}
	t.Fatal("insights domain is not registered")
	return nil
}

func findInventoryInsightsAction(t *testing.T, domain *Domain, name string) *Action {
	t.Helper()
	for index := range domain.Actions {
		if domain.Actions[index].Name == name {
			return &domain.Actions[index]
		}
	}
	t.Fatalf("%s action is not registered", name)
	return nil
}
