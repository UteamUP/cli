package registry

import (
	"strings"
	"testing"
)

func maintenancePlanAction(t *testing.T, name string) (*Domain, Action) {
	t.Helper()
	domain := findDomain("asset-maintenance-plan")
	if domain == nil {
		t.Fatal("asset-maintenance-plan domain is not registered")
	}

	for _, action := range domain.Actions {
		if action.Name == name {
			return domain, action
		}
	}

	t.Fatalf("asset-maintenance-plan action %q is not registered", name)
	return nil, Action{}
}

func TestAssetMaintenancePlanRoutesAreGuidOnly(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args map[string]any
		path string
	}{
		{name: "get", args: map[string]any{"planExternalGuid": "plan-guid"}, path: "/api/v1/maintenanceplans/plan-guid"},
		{name: "create", args: map[string]any{"assetExternalGuid": "asset-guid"}, path: "/api/v1/maintenanceplans/asset/asset-guid"},
		{name: "update", args: map[string]any{"planExternalGuid": "plan-guid"}, path: "/api/v1/maintenanceplans/plan-guid"},
		{name: "delete", args: map[string]any{"planExternalGuid": "plan-guid"}, path: "/api/v1/maintenanceplans/plan-guid"},
		{name: "items", args: map[string]any{"planExternalGuid": "plan-guid"}, path: "/api/v1/maintenanceplans/plan-guid/items"},
		{name: "item-add", args: map[string]any{"planExternalGuid": "plan-guid"}, path: "/api/v1/maintenanceplans/plan-guid/items"},
		{name: "item-update", args: map[string]any{"itemExternalGuid": "item-guid"}, path: "/api/v1/maintenanceplans/items/item-guid"},
		{name: "item-delete", args: map[string]any{"itemExternalGuid": "item-guid"}, path: "/api/v1/maintenanceplans/items/item-guid"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			domain, action := maintenancePlanAction(t, testCase.name)
			path, consumed := buildRESTPath(domain, action, testCase.args)
			if path != testCase.path {
				t.Fatalf("path = %q, want %q", path, testCase.path)
			}
			if len(consumed) != 1 {
				t.Fatalf("consumed args = %v, want exactly one GUID argument", consumed)
			}
			if strings.Contains(action.RESTPath, "{id}") {
				t.Fatalf("RESTPath exposes an integer identity: %q", action.RESTPath)
			}
			for _, arg := range action.Args {
				if arg.Type != "uuid" || strings.EqualFold(arg.Name, "id") {
					t.Fatalf("public identity arg is not GUID-only: %+v", arg)
				}
			}
		})
	}
}

func TestAssetMaintenancePlanFlagsMirrorBackendModels(t *testing.T) {
	t.Parallel()

	_, list := maintenancePlanAction(t, "list")
	if len(list.Flags) != 1 || list.Flags[0].BodyName != "assetExternalGuid" {
		t.Fatalf("list asset filter = %+v, want assetExternalGuid", list.Flags)
	}

	for _, actionName := range []string{"create", "update"} {
		_, action := maintenancePlanAction(t, actionName)
		assertMaintenanceFlag(t, action, "name", "name", "string", true)
		assertMaintenanceFlag(t, action, "is-active", "isActive", "bool", false)
	}

	for _, actionName := range []string{"item-add", "item-update"} {
		_, action := maintenancePlanAction(t, actionName)
		assertMaintenanceFlag(t, action, "trigger-type", "triggerType", "int", false)
		assertMaintenanceFlag(t, action, "calendar-interval-days", "calendarIntervalDays", "int", false)
		assertMaintenanceFlag(t, action, "meter-interval-value", "meterIntervalValue", "float", false)
		assertMaintenanceFlag(t, action, "workorder-template-external-guid", "workorderTemplateExternalGuid", "string", false)
		assertMaintenanceFlag(t, action, "required-for-warranty", "isRequiredForWarranty", "bool", false)
		assertMaintenanceFlag(t, action, "required-for-certification", "isRequiredForCertification", "bool", false)
	}
}

func assertMaintenanceFlag(
	t *testing.T,
	action Action,
	name string,
	bodyName string,
	flagType string,
	required bool,
) {
	t.Helper()
	for _, flag := range action.Flags {
		if flag.Name != name {
			continue
		}

		actualBodyName := flag.BodyName
		if actualBodyName == "" {
			actualBodyName = toCamelCase(flag.Name)
		}
		if actualBodyName != bodyName || flag.Type != flagType || flag.Required != required {
			t.Fatalf("%s flag = %+v, want body=%q type=%q required=%t", name, flag, bodyName, flagType, required)
		}
		return
	}

	t.Fatalf("action %q is missing flag %q", action.Name, name)
}
