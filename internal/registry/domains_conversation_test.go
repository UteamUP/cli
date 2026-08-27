package registry

import "testing"

// requireConversationDomain finds the registered domain or fails the test.
func requireConversationDomain(t *testing.T) *Domain {
	t.Helper()
	for _, domain := range DefaultRegistry.Domains() {
		if domain.Name == "conversation" {
			return domain
		}
	}
	t.Fatal("conversation domain is not registered")
	return nil
}

// The registry routes by APIPath + RESTPath, not by ToolName. A domain that omits APIPath
// gets one fabricated from its name, which 404s silently as soon as the name and the route
// diverge - so this asserts the route is stated explicitly.
func TestConversationDomainDeclaresApiPath(t *testing.T) {
	domain := requireConversationDomain(t)

	if domain.APIPath != "/api/conversation" {
		t.Errorf("APIPath = %q, want %q", domain.APIPath, "/api/conversation")
	}
}

func TestConversationDomainActionsRouteCorrectly(t *testing.T) {
	domain := requireConversationDomain(t)

	want := map[string]struct {
		method   string
		restPath string
	}{
		"list":    {"GET", ""},
		"get":     {"GET", "{conversationGuid}"},
		"send":    {"POST", "{conversationGuid}/messages"},
		"search":  {"GET", "{conversationGuid}/messages/search"},
		"since":   {"GET", "{conversationGuid}/messages"},
		"archive": {"PATCH", "{conversationGuid}/inbox-state"},
	}

	got := map[string]bool{}
	for _, action := range domain.Actions {
		expected, known := want[action.Name]
		if !known {
			t.Errorf("unexpected action %q", action.Name)
			continue
		}
		got[action.Name] = true

		if action.HTTPMethod != expected.method {
			t.Errorf("action %q: HTTPMethod = %q, want %q", action.Name, action.HTTPMethod, expected.method)
		}
		if action.RESTPath != expected.restPath {
			t.Errorf("action %q: RESTPath = %q, want %q", action.Name, action.RESTPath, expected.restPath)
		}
	}

	for name := range want {
		if !got[name] {
			t.Errorf("missing action %q", name)
		}
	}
}

// Identity comes from the bearer token. A user argument or flag would let a caller read
// someone else's inbox or post as them, so assert none of them exists.
func TestConversationDomainNeverAcceptsAUserIdentity(t *testing.T) {
	domain := requireConversationDomain(t)

	forbidden := []string{"userid", "userguid", "user", "senderid", "senderguid", "sender", "onbehalfof"}
	contains := func(haystack string) bool {
		for _, needle := range forbidden {
			if normalize(haystack) == needle {
				return true
			}
		}
		return false
	}

	for _, action := range domain.Actions {
		for _, arg := range action.Args {
			if contains(arg.Name) {
				t.Errorf("action %q accepts identity arg %q - the caller must come from the token", action.Name, arg.Name)
			}
		}
		for _, flag := range action.Flags {
			if contains(flag.Name) || contains(flag.BodyName) {
				t.Errorf("action %q accepts identity flag %q - the caller must come from the token", action.Name, flag.Name)
			}
		}
	}
}

// normalize lowercases and strips separators so "user-guid", "userGuid" and "user_guid" all
// collapse to the same token.
func normalize(value string) string {
	out := make([]rune, 0, len(value))
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z':
			out = append(out, r+('a'-'A'))
		case r == '-' || r == '_':
		default:
			out = append(out, r)
		}
	}
	return string(out)
}
