package registry

import "testing"

func TestMeterDisplayFlagsPreserveOptionalPrecision(t *testing.T) {
	add := findDomainAction(t, "asset-type-meter", "add")
	for _, expected := range []struct{ name, body, kind string }{
		{"decimal-places", "decimalPlaces", "int"},
		{"rounding-mode", "roundingMode", "string"},
		{"chart-type", "chartType", "string"},
		{"monitoring-visibility", "monitoringVisibility", "string"},
	} {
		found := false
		for _, flag := range add.Flags {
			if flag.Name == expected.name {
				found = true
				if flag.BodyName != expected.body || flag.Type != expected.kind || flag.Required {
					t.Fatalf("unexpected display flag: %+v", flag)
				}
			}
		}
		if !found {
			t.Fatalf("missing display flag %s", expected.name)
		}
	}
}
