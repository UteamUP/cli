package registry

import (
	"strings"
	"testing"
)

func TestUpmateFieldServiceDomainMirrorsExactMCPTools(t *testing.T) {
	domain := findDomain("upmate-field-service")
	if domain == nil {
		t.Fatal("expected upmate-field-service domain to be registered")
	}
	expected := map[string]string{
		"schedule-preview":             "UteamupUpmateScheduleOptimizationPreview",
		"schedule-explain":             "UteamupUpmateScheduleOptimizationExplain",
		"maintenance-suggest":          "UteamupUpmateMaintenancePlanSuggest",
		"maintenance-due-explain":      "UteamupUpmateMaintenanceDueExplain",
		"fieldnote-transcribe":         "UteamupUpmateFieldnoteTranscribe",
		"portal-request-classify-cost": "UteamupUpmatePortalRequestClassifyCost",
		"portal-request-classify":      "UteamupUpmatePortalRequestClassify",
		"service-billing-review":       "UteamupUpmateServiceBillingReview",
	}

	for _, action := range domain.Actions {
		want, exists := expected[action.Name]
		if !exists {
			t.Errorf("unexpected action %q", action.Name)
			continue
		}
		if action.ToolName != want {
			t.Errorf("%s tool = %q, want %q", action.Name, action.ToolName, want)
		}
		if !action.MCPOnly {
			t.Errorf("%s must use the MCP transport", action.Name)
		}
		if action.RESTBasePath != "" || action.RESTPath != "" || action.HTTPMethod != "" {
			t.Errorf("%s unexpectedly declares a REST adapter", action.Name)
		}
		delete(expected, action.Name)
	}
	for missing := range expected {
		t.Errorf("missing field-service action %q", missing)
	}
}

func TestUpmateFieldServicePublicInputsAreGUIDOnlyAndActorNeutral(t *testing.T) {
	domain := findDomain("upmate-field-service")
	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if strings.HasSuffix(arg.Name, "Guid") && arg.Type != "uuid" {
				t.Errorf("%s arg %s is not a UUID", action.Name, arg.Name)
			}
			assertFieldServiceInputIsActorNeutral(t, action.Name, arg.Name)
		}
		for _, flag := range action.Flags {
			if strings.HasSuffix(flag.BodyName, "Guid") && flag.Type != "uuid" {
				t.Errorf("%s flag %s is not a UUID", action.Name, flag.Name)
			}
			assertFieldServiceInputIsActorNeutral(t, action.Name, flag.BodyName)
		}
	}
}

func TestUpmateFieldServiceBoundedCollectionsAndObjectiveDefaults(t *testing.T) {
	preview := findUpmateFieldServiceAction(t, "schedule-preview")
	flags := make(map[string]FlagDef)
	for _, flag := range preview.Flags {
		flags[flag.Name] = flag
	}

	for _, name := range []string{"workorder-guids", "technician-guids"} {
		if flags[name].Type != "stringSlice" || !flags[name].Required {
			t.Errorf("%s must be a required bounded GUID collection", name)
		}
	}
	for _, name := range []string{
		"competency-weight",
		"travel-weight",
		"lateness-weight",
		"overtime-weight",
		"workload-balance-weight",
		"continuity-weight",
		"repeat-visit-weight",
		"disruption-weight",
	} {
		if flags[name].Type != "float" || flags[name].Default != 1.0 {
			t.Errorf("%s default = %v, want 1.0", name, flags[name].Default)
		}
	}
}

func TestUpmateFieldServiceCostPreviewTakesNoInputsAndDisclosesBeforeCharge(t *testing.T) {
	cost := findUpmateFieldServiceAction(t, "portal-request-classify-cost")
	if len(cost.Args) != 0 || len(cost.Flags) != 0 {
		t.Errorf("cost preview must be tenant/actor scoped via the MCP context, not inputs")
	}
	if !strings.Contains(strings.ToLower(cost.Description), "cost") {
		t.Errorf("cost preview description must state it discloses the credit cost: %q", cost.Description)
	}
}

func TestUpmateFieldServiceDescriptionsKeepMutationsOutOfScope(t *testing.T) {
	domain := findDomain("upmate-field-service")
	for _, action := range domain.Actions {
		description := strings.ToLower(action.Description)
		if !strings.Contains(description, "without") &&
			!strings.Contains(description, "never") &&
			!strings.Contains(description, "non-executable") {
			t.Errorf("%s description does not state its safety boundary: %q", action.Name, action.Description)
		}
	}
}

func findUpmateFieldServiceAction(t *testing.T, name string) Action {
	t.Helper()
	domain := findDomain("upmate-field-service")
	if domain == nil {
		t.Fatal("expected upmate-field-service domain")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return action
		}
	}
	t.Fatalf("expected field-service action %q", name)
	return Action{}
}

func assertFieldServiceInputIsActorNeutral(t *testing.T, action string, name string) {
	t.Helper()
	normalized := strings.ToLower(name)
	for _, forbidden := range []string{
		"tenantid",
		"tenantguid",
		"userid",
		"userguid",
		"actor",
		"apikey",
		"provider",
		"model",
	} {
		if strings.Contains(normalized, forbidden) {
			t.Errorf("%s exposes forbidden caller-controlled field %q", action, name)
		}
	}
}
