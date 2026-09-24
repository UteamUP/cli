package registry

import "testing"

// The weather domain is pinned to WeatherController's real templates. It is deliberately not in
// TestPhantomPathDomainsNeverDeriveTheStrippedFallback: its APIPath equals the derived
// /api/weather, so the real routes are proven here by exact path instead.
func TestWeatherDomainResolvesToRealControllerRoutes(t *testing.T) {
	t.Parallel()
	const locationGUID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	tests := []struct {
		action, tool string
		args         map[string]any
		path         string
	}{
		{"list", "UteamupWeatherGetAll", nil, "/api/weather/locations"},
		{"get", "UteamupWeatherGetForLocation", map[string]any{"locationGuid": locationGUID}, "/api/weather/locations/" + locationGUID},
		{"alerts", "UteamupWeatherGetAlerts", nil, "/api/weather/alerts"},
		{"windows", "UteamupWeatherSiteWindows", nil, "/api/weather/windows"},
	}

	domain := findDomain("weather")
	if domain == nil {
		t.Fatal("weather domain is not registered")
	}
	if domain.APIPath != "/api/weather" {
		t.Fatalf("APIPath = %q, want /api/weather", domain.APIPath)
	}
	if len(domain.Actions) != len(tests) {
		t.Fatalf("weather has %d actions, want %d (no unrouted verbs)", len(domain.Actions), len(tests))
	}

	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			action := findAction(domain, test.action)
			if action == nil {
				t.Fatalf("action %q is not registered", test.action)
			}
			if action.MCPOnly || action.HTTPMethod != "GET" || action.ToolName != test.tool {
				t.Fatalf("%s = MCPOnly %v method %q tool %q, want REST GET %q",
					test.action, action.MCPOnly, action.HTTPMethod, action.ToolName, test.tool)
			}
			args := test.args
			if args == nil {
				args = map[string]any{}
			}
			path, _ := buildRESTPath(domain, *action, args)
			if path != test.path {
				t.Fatalf("path = %q, want %q", path, test.path)
			}
			for _, arg := range action.Args {
				if arg.Type == "int" || arg.Name == "id" || arg.Name == "locationId" {
					t.Errorf("%s exposes an integer identifier %+v", test.action, arg)
				}
			}
		})
	}
}

func TestWeatherGetTakesANonEmptyLocationGuid(t *testing.T) {
	t.Parallel()
	action := findAction(findDomain("weather"), "get")
	if action == nil || len(action.Args) != 1 {
		t.Fatal("weather get must take exactly one positional argument")
	}
	arg := action.Args[0]
	if arg.Name != "locationGuid" || !arg.Required || arg.Type != "non-empty-uuid" {
		t.Fatalf("weather get arg = %+v, want required non-empty-uuid locationGuid", arg)
	}
}

func TestWeatherWindowsSendsTheLocationGuidAsAQueryParameter(t *testing.T) {
	t.Parallel()
	domain := findDomain("weather")
	action := findAction(domain, "windows")
	if action == nil || len(action.Flags) != 1 {
		t.Fatal("weather windows must declare exactly one flag")
	}
	flag := action.Flags[0]
	if flag.Name != "location-guid" || flag.QueryName != "locationGuid" || flag.Type != "non-empty-uuid" || flag.Required {
		t.Fatalf("weather windows flag = %+v, want optional non-empty-uuid --location-guid as ?locationGuid", flag)
	}

	path, _ := buildRESTPath(domain, *action, map[string]any{})
	const locationGUID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	got := appendQueryParameters(path, map[string]any{flag.QueryName: locationGUID})
	if want := "/api/weather/windows?locationGuid=" + locationGUID; got != want {
		t.Fatalf("windows URL = %q, want %q", got, want)
	}
	if got := appendQueryParameters(path, nil); got != "/api/weather/windows" {
		t.Fatalf("windows URL without a location = %q, want /api/weather/windows", got)
	}
}
