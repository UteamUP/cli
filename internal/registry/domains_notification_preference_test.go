package registry

import "testing"

func findNotificationPreferenceDomain(t *testing.T) *Domain {
	t.Helper()
	for _, d := range DefaultRegistry.Domains() {
		if d.Name == "notification-preference" {
			return d
		}
	}
	t.Fatal("expected notification-preference domain to be registered")
	return nil
}

func findNotificationPreferenceAction(t *testing.T, name string) *Action {
	t.Helper()
	d := findNotificationPreferenceDomain(t)
	for i := range d.Actions {
		if d.Actions[i].Name == name {
			return &d.Actions[i]
		}
	}
	t.Fatalf("expected %q action on notification-preference domain", name)
	return nil
}

func TestNotificationPreferenceDomainRegistered(t *testing.T) {
	d := findNotificationPreferenceDomain(t)
	if d.Description == "" {
		t.Error("notification-preference domain must have a Description")
	}
	// APIPath must be explicit: the backend route is the current-user-scoped
	// /api/notificationpreference (no path id).
	if d.APIPath != "/api/notificationpreference" {
		t.Errorf("notification-preference APIPath = %q, want %q", d.APIPath, "/api/notificationpreference")
	}
	wantAliases := map[string]bool{
		"notification-preferences": true,
		"notif-pref":               true,
		"notification-prefs":       true,
	}
	for _, a := range d.Aliases {
		delete(wantAliases, a)
	}
	if len(wantAliases) != 0 {
		t.Errorf("notification-preference domain missing aliases: %v (got %v)", wantAliases, d.Aliases)
	}
}

func TestNotificationPreferenceActionsWired(t *testing.T) {
	d := findNotificationPreferenceDomain(t)
	expected := map[string]string{
		"get": "UteamupNotificationPreferenceGet",
		"set": "UteamupNotificationPreferenceSet",
	}
	got := map[string]string{}
	for _, a := range d.Actions {
		got[a.Name] = a.ToolName
	}
	if len(d.Actions) != len(expected) {
		t.Errorf("notification-preference action count = %d, want %d (%v)", len(d.Actions), len(expected), got)
	}
	for action, tool := range expected {
		if got[action] != tool {
			t.Errorf("expected action %q to map to %q, got %q", action, tool, got[action])
		}
	}
}

// get → GET /api/notificationpreference (no path id, no RESTPath, no args).
func TestNotificationPreferenceGetAction(t *testing.T) {
	action := findNotificationPreferenceAction(t, "get")

	if action.RESTPath != "" {
		t.Errorf("get RESTPath = %q, want empty (GET base path)", action.RESTPath)
	}
	if action.HTTPMethod != "" {
		t.Errorf("get HTTPMethod = %q, want empty (defaults to GET via action-name map)", action.HTTPMethod)
	}
	if HTTPMethod["get"] != "GET" {
		t.Errorf("get resolves to %q, want GET", HTTPMethod["get"])
	}
	if len(action.Args) != 0 {
		t.Errorf("get must take no positional args (current-user scoped), got %+v", action.Args)
	}
	if len(action.Flags) != 0 {
		t.Errorf("get must take no flags, got %+v", action.Flags)
	}
}

// set → PUT /api/notificationpreference. "set" is not in the action-name verb
// map and is not an `update-` prefix, so HTTPMethod must be set explicitly to PUT;
// otherwise runCommand would default it to GET.
func TestNotificationPreferenceSetAction(t *testing.T) {
	action := findNotificationPreferenceAction(t, "set")

	if action.HTTPMethod != "PUT" {
		t.Errorf("set HTTPMethod = %q, want PUT (explicit; 'set' is not in the verb map)", action.HTTPMethod)
	}
	if action.RESTPath != "" {
		t.Errorf("set RESTPath = %q, want empty (PUT to base path, no id)", action.RESTPath)
	}
	if len(action.Args) != 0 {
		t.Errorf("set must take no positional args, got %+v", action.Args)
	}

	flagByName := func(name string) *FlagDef {
		for i := range action.Flags {
			if action.Flags[i].Name == name {
				return &action.Flags[i]
			}
		}
		return nil
	}

	// Flags must carry NO Default so an unchanged flag is omitted from the PUT
	// body — that is what makes `set` a true partial update. BodyName must pin
	// each flag onto the backend WorkorderNotificationPreferenceUpdateModel field
	// (kebab->camel auto-conversion would not produce these names).
	for _, want := range []struct {
		name     string
		typ      string
		bodyName string
	}{
		{"due-window-start", "int", "dueDateWindowStartHours"},
		{"due-window-end", "int", "dueDateWindowEndHours"},
		{"start-window-start", "int", "startDateWindowStartHours"},
		{"start-window-end", "int", "startDateWindowEndHours"},
		{"notify-on-due-date", "bool", "notifyOnDueDate"},
		{"notify-on-start-date", "bool", "notifyOnStartDate"},
		{"notify-on-change", "bool", "notifyOnChange"},
		{"notify-on-comment", "bool", "notifyOnComment"},
	} {
		f := flagByName(want.name)
		if f == nil {
			t.Fatalf("set must expose a %q flag", want.name)
		}
		if f.Type != want.typ {
			t.Errorf("%s flag type = %q, want %q", want.name, f.Type, want.typ)
		}
		if f.BodyName != want.bodyName {
			t.Errorf("%s flag BodyName = %q, want %q", want.name, f.BodyName, want.bodyName)
		}
		if f.Default != nil {
			t.Errorf("%s flag Default = %v, want nil (partial update must not send unchanged flags)", want.name, f.Default)
		}
		if f.Required {
			t.Errorf("%s flag must not be Required (partial update)", want.name)
		}
	}

	if len(action.Flags) != 8 {
		t.Errorf("set flag count = %d, want 8", len(action.Flags))
	}
}
