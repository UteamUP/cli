package registry

import "testing"

func findOnCallDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "oncall" {
			return d
		}
	}
	t.Fatal("expected oncall domain to be registered")
	return nil
}

func TestOnCallDomainRegistered(t *testing.T) {
	d := findOnCallDomain(t)
	if d.Description == "" {
		t.Error("oncall domain must have a Description")
	}
	if d.APIPath != "/api/oncall" {
		t.Errorf("oncall APIPath = %q, want %q", d.APIPath, "/api/oncall")
	}
	hasAlias := false
	for _, a := range d.Aliases {
		if a == "on-call" {
			hasAlias = true
		}
	}
	if !hasAlias {
		t.Errorf("oncall domain missing 'on-call' alias, got %v", d.Aliases)
	}
}

func TestOnCallWhoActionWired(t *testing.T) {
	d := findOnCallDomain(t)
	var who *Action
	for i := range d.Actions {
		if d.Actions[i].Name == "who" {
			who = &d.Actions[i]
		}
	}
	if who == nil {
		t.Fatal("expected 'who' action on oncall domain")
	}
	if who.ToolName != "UteamupOnCallWho" {
		t.Errorf("who ToolName = %q, want %q", who.ToolName, "UteamupOnCallWho")
	}
	if who.HTTPMethod != "GET" {
		t.Errorf("who HTTPMethod = %q, want GET", who.HTTPMethod)
	}
	if who.RESTPath != "{schedule-guid}/who" {
		t.Errorf("who RESTPath = %q, want %q", who.RESTPath, "{schedule-guid}/who")
	}
	// schedule-guid is a required uuid positional arg
	if len(who.Args) != 1 || who.Args[0].Name != "schedule-guid" || !who.Args[0].Required || who.Args[0].Type != "uuid" {
		t.Errorf("who must take a required uuid 'schedule-guid' arg, got %+v", who.Args)
	}
	// optional 'at' query flag
	hasAt := false
	for _, f := range who.Flags {
		if f.Name == "at" {
			hasAt = true
		}
	}
	if !hasAt {
		t.Errorf("who must expose an optional 'at' flag, got %+v", who.Flags)
	}
}
