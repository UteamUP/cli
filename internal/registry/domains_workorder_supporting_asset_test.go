package registry

import (
	"reflect"
	"testing"
)

func workorderSupportingAssetAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("workorder-supporting-asset")
	if domain == nil {
		t.Fatal("workorder-supporting-asset domain is not registered")
	}
	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}
	t.Fatalf("workorder-supporting-asset action %q is not registered", name)
	return nil, Action{}
}

func TestWorkorderSupportingAssetRoutesUseGuidIdentity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		args     map[string]any
		path     string
		consumed []string
	}{
		{
			name:     "list",
			args:     map[string]any{"workorderGuid": "wo-guid"},
			path:     "/api/workorder/wo-guid/supporting-assets",
			consumed: []string{"workorderGuid"},
		},
		{
			name:     "create",
			args:     map[string]any{"workorderGuid": "wo-guid"},
			path:     "/api/workorder/wo-guid/supporting-assets",
			consumed: []string{"workorderGuid"},
		},
		{
			name: "update",
			args: map[string]any{
				"workorderGuid":   "wo-guid",
				"requirementGuid": "requirement-guid",
			},
			path:     "/api/workorder/wo-guid/supporting-assets/requirement-guid",
			consumed: []string{"workorderGuid", "requirementGuid"},
		},
		{
			name: "delete",
			args: map[string]any{
				"workorderGuid":   "wo-guid",
				"requirementGuid": "requirement-guid",
			},
			path:     "/api/workorder/wo-guid/supporting-assets/requirement-guid",
			consumed: []string{"workorderGuid", "requirementGuid"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			domain, action := workorderSupportingAssetAction(t, testCase.name)
			path, consumed := buildRESTPath(domain, action, testCase.args)
			if path != testCase.path {
				t.Fatalf("path = %q, want %q", path, testCase.path)
			}
			if !reflect.DeepEqual(consumed, testCase.consumed) {
				t.Fatalf("consumed = %v, want %v", consumed, testCase.consumed)
			}
			for _, arg := range action.Args {
				if arg.Type != "uuid" {
					t.Fatalf("public identity arg is not UUID typed: %+v", arg)
				}
			}
		})
	}
}

func TestWorkorderSupportingAssetMutationFlagsMirrorBackend(t *testing.T) {
	t.Parallel()

	_, create := workorderSupportingAssetAction(t, "create")
	_, update := workorderSupportingAssetAction(t, "update")
	if !reflect.DeepEqual(create.Flags, update.Flags) {
		t.Fatal("create and update must send the same replacement-model fields")
	}

	want := map[string]struct {
		bodyName string
		flagType string
	}{
		"name":                          {bodyName: "name", flagType: "string"},
		"exact-asset-guid":              {bodyName: "exactAssetGuid", flagType: "uuid"},
		"asset-type-guid":               {bodyName: "assetTypeGuid", flagType: "uuid"},
		"quantity":                      {bodyName: "quantity", flagType: "int"},
		"mandatory":                     {bodyName: "isMandatory", flagType: "bool"},
		"active":                        {bodyName: "isActive", flagType: "bool"},
		"availability-validated-at-utc": {bodyName: "availabilityValidatedAtUtc", flagType: "string"},
		"max-evidence-age-hours":        {bodyName: "maxEvidenceAgeHours", flagType: "int"},
	}
	if len(create.Flags) != len(want) {
		t.Fatalf("mutation flags = %d, want %d", len(create.Flags), len(want))
	}
	for _, flag := range create.Flags {
		expected, ok := want[flag.Name]
		if !ok {
			t.Fatalf("unexpected mutation flag: %+v", flag)
		}
		if flag.BodyName != expected.bodyName || flag.Type != expected.flagType {
			t.Fatalf("flag %q = %+v, want %+v", flag.Name, flag, expected)
		}
	}
}

func TestWorkorderSupportingAssetToolNamesMirrorBackend(t *testing.T) {
	t.Parallel()

	want := map[string]string{
		"list":   "UteamupWorkorderSupportingAssetRequirementList",
		"create": "UteamupWorkorderSupportingAssetRequirementCreate",
		"update": "UteamupWorkorderSupportingAssetRequirementUpdate",
		"delete": "UteamupWorkorderSupportingAssetRequirementDelete",
	}
	for actionName, toolName := range want {
		_, action := workorderSupportingAssetAction(t, actionName)
		if action.ToolName != toolName {
			t.Fatalf("%s tool = %q, want %q", actionName, action.ToolName, toolName)
		}
	}
}
