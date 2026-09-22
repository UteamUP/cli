package registry

import "testing"

func TestSystemStatusRouteMirrorsGlobalAdminController(t *testing.T) {
	t.Parallel()
	domain := findDomain("system-status")
	if domain == nil {
		t.Fatal("system-status domain is not registered")
	}
	if len(domain.Actions) != 1 {
		t.Fatalf("system-status actions = %d, want 1 (read-only domain)", len(domain.Actions))
	}
	action := domain.Actions[0]
	if action.Name != "get" || action.HTTPMethod != "GET" {
		t.Fatalf("action = %s %s, want get GET", action.Name, action.HTTPMethod)
	}
	if action.ToolName != "UteamupExternalServiceStatus" {
		t.Fatalf("tool = %q, want UteamupExternalServiceStatus", action.ToolName)
	}
	path, _ := buildRESTPath(domain, action, map[string]any{})
	if path != "/api/globaladmin/external-services/status" {
		t.Fatalf("path = %q", path)
	}
}

func TestSystemStatusRefreshIsABoolQueryFlag(t *testing.T) {
	t.Parallel()
	domain := findDomain("system-status")
	if domain == nil {
		t.Fatal("system-status domain is not registered")
	}
	flags := domain.Actions[0].Flags
	if len(flags) != 1 {
		t.Fatalf("flags = %d, want only refresh", len(flags))
	}
	refresh := flags[0]
	if refresh.Name != "refresh" || refresh.QueryName != "refresh" || refresh.Type != "bool" || refresh.Required {
		t.Fatalf("refresh flag = %+v", refresh)
	}
	if got := appendQueryParameters("/api/globaladmin/external-services/status", map[string]any{"refresh": true}); got != "/api/globaladmin/external-services/status?refresh=true" {
		t.Fatalf("query = %q", got)
	}
}

func TestSystemStatusDeclaresNoIdentifier(t *testing.T) {
	t.Parallel()
	domain := findDomain("system-status")
	if domain == nil {
		t.Fatal("system-status domain is not registered")
	}
	for _, action := range domain.Actions {
		if len(action.Args) != 0 {
			t.Fatalf("action %q declares args %+v; the status read takes no identifier", action.Name, action.Args)
		}
	}
}
