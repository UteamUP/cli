package registry

import (
	"strings"
	"testing"
)

func TestSustainabilityDomainRoutesThroughMcpTools(t *testing.T) {
	t.Parallel()
	domain := findDomain("sustainability")
	if domain == nil {
		t.Fatal("sustainability domain is not registered")
	}
	want := map[string]string{
		"energy-reading-list":    "UteamupSustainabilityEnergyReadingList",
		"energy-reading-get":     "UteamupSustainabilityEnergyReadingGet",
		"energy-reading-create":  "UteamupSustainabilityEnergyReadingCreate",
		"energy-reading-update":  "UteamupSustainabilityEnergyReadingUpdate",
		"water-reading-list":     "UteamupSustainabilityWaterReadingList",
		"water-reading-get":      "UteamupSustainabilityWaterReadingGet",
		"water-reading-create":   "UteamupSustainabilityWaterReadingCreate",
		"water-reading-update":   "UteamupSustainabilityWaterReadingUpdate",
		"waste-record-list":      "UteamupSustainabilityWasteRecordList",
		"waste-record-get":       "UteamupSustainabilityWasteRecordGet",
		"waste-record-create":    "UteamupSustainabilityWasteRecordCreate",
		"waste-record-update":    "UteamupSustainabilityWasteRecordUpdate",
		"carbon-emission-list":   "UteamupSustainabilityCarbonEmissionList",
		"carbon-emission-get":    "UteamupSustainabilityCarbonEmissionGet",
		"carbon-emission-create": "UteamupSustainabilityCarbonEmissionCreate",
		"carbon-emission-update": "UteamupSustainabilityCarbonEmissionUpdate",
		"emission-factor-list":   "UteamupSustainabilityEmissionFactorList",
		"emission-factor-get":    "UteamupSustainabilityEmissionFactorGet",
		"benchmark-list":         "UteamupSustainabilityEnergyBenchmarkList",
		"benchmark-get":          "UteamupSustainabilityEnergyBenchmarkGet",
		"goal-list":              "UteamupSustainabilityGoalList",
		"goal-get":               "UteamupSustainabilityGoalGet",
		"anomaly-list":           "UteamupSustainabilityEnergyAnomalyList",
		"anomaly-get":            "UteamupSustainabilityEnergyAnomalyGet",
	}
	if len(domain.Actions) != len(want) {
		t.Fatalf("actions = %d, want %d", len(domain.Actions), len(want))
	}
	for _, action := range domain.Actions {
		if want[action.Name] != action.ToolName || !action.MCPOnly || !action.RejectExtraArgs {
			t.Fatalf("action contract = %+v", action)
		}
	}
}

func TestSustainabilityGetAndUpdateRequireAGuid(t *testing.T) {
	t.Parallel()
	domain := findDomain("sustainability")
	for _, action := range domain.Actions {
		name := action.Name
		if !strings.HasSuffix(name, "-get") && !strings.HasSuffix(name, "-update") {
			continue
		}
		if len(action.Args) != 1 || !action.Args[0].Required || action.Args[0].Type != "non-empty-uuid" {
			t.Fatalf("%s must take one required public GUID: %+v", name, action.Args)
		}
	}
}

func TestSustainabilityCreateRequiresTheRestRequiredFields(t *testing.T) {
	t.Parallel()
	domain := findDomain("sustainability")
	want := map[string][]string{
		"energy-reading-create":  {"utility-meter-guid", "reading-value", "reading-date"},
		"water-reading-create":   {"location-guid", "reading-date", "consumption"},
		"waste-record-create":    {"location-guid", "record-date", "waste-category", "quantity"},
		"carbon-emission-create": {"scope", "category", "source-description", "period-start", "period-end", "co2e-kg", "activity-data", "emission-factor-value"},
	}
	for _, action := range domain.Actions {
		required, ok := want[action.Name]
		if !ok {
			if strings.HasSuffix(action.Name, "-update") {
				for _, flag := range action.Flags {
					if flag.Required {
						t.Fatalf("%s flag %s must be optional so omitted values are kept", action.Name, flag.Name)
					}
				}
			}
			continue
		}
		got := map[string]bool{}
		for _, flag := range action.Flags {
			if flag.Required {
				got[flag.Name] = true
			}
		}
		if len(got) != len(required) {
			t.Fatalf("%s required flags = %v, want %v", action.Name, got, required)
		}
		for _, name := range required {
			if !got[name] {
				t.Fatalf("%s must require --%s", action.Name, name)
			}
		}
	}
}
