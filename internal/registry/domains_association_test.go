package registry

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/uteamup/cli/internal/auth"
	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

const (
	testMeetingGUID = "32000000-0000-0000-0000-000000000001"
	testVoteGUID    = "32000000-0000-0000-0000-000000000002"
	testNoticeGUID  = "32000000-0000-0000-0000-000000000003"
)

// The association commands mirror AssociationController (api/association); every record is
// addressed by GUID and the tenant comes from the session, never from an argument.
func TestAssociationCommandsMatchTheBackendRoutes(t *testing.T) {
	d := findDomain("association")
	if d == nil {
		t.Fatal("expected association domain to be registered")
	}

	cases := []struct {
		action, method, tool string
		args                 map[string]any
		path                 string
	}{
		{"meetings-list", "GET", "uteamup_association_meetings_list", nil, "/api/association/meetings"},
		{"meeting-get", "GET", "uteamup_association_meeting_get", map[string]any{"meetingGuid": testMeetingGUID}, "/api/association/meetings/" + testMeetingGUID},
		{"meeting-create", "POST", "uteamup_association_meeting_create", nil, "/api/association/meetings"},
		{"meeting-update", "PUT", "uteamup_association_meeting_update", map[string]any{"meetingGuid": testMeetingGUID}, "/api/association/meetings/" + testMeetingGUID},
		{"meeting-delete", "DELETE", "uteamup_association_meeting_delete", map[string]any{"meetingGuid": testMeetingGUID}, "/api/association/meetings/" + testMeetingGUID},
		{"votes-list", "GET", "uteamup_association_votes_list", nil, "/api/association/votes"},
		{"vote-create", "POST", "uteamup_association_vote_create", nil, "/api/association/votes"},
		{"ballot-cast", "POST", "uteamup_association_ballot_cast", map[string]any{"voteGuid": testVoteGUID}, "/api/association/votes/" + testVoteGUID + "/ballot"},
		{"vote-close", "POST", "uteamup_association_vote_close", map[string]any{"voteGuid": testVoteGUID}, "/api/association/votes/" + testVoteGUID + "/close"},
		{"notices-list", "GET", "uteamup_association_notices_list", nil, "/api/association/notices"},
		{"notice-post", "POST", "uteamup_association_notice_post", nil, "/api/association/notices"},
		{"notice-delete", "DELETE", "uteamup_association_notice_delete", map[string]any{"noticeGuid": testNoticeGUID}, "/api/association/notices/" + testNoticeGUID},
	}
	if len(d.Actions) != len(cases) {
		t.Errorf("association has %d actions, want %d", len(d.Actions), len(cases))
	}
	for _, c := range cases {
		action := findAction(d, c.action)
		if action == nil {
			t.Errorf("expected %s action on association", c.action)
			continue
		}
		if action.MCPOnly {
			t.Errorf("%s has a REST endpoint and must not be MCPOnly", c.action)
		}
		if action.HTTPMethod != c.method || action.ToolName != c.tool {
			t.Errorf("%s = %s via %s, want %s via %s", c.action, action.HTTPMethod, action.ToolName, c.method, c.tool)
		}
		args := c.args
		if args == nil {
			args = map[string]any{}
		}
		if path, _ := buildRESTPath(d, *action, args); path != c.path {
			t.Errorf("%s path = %q, want %q", c.action, path, c.path)
		}
		if len(action.Args) != len(c.args) {
			t.Errorf("%s takes %d positional args, want %d", c.action, len(action.Args), len(c.args))
		}
		for _, arg := range action.Args {
			if arg.Type != "non-empty-uuid" {
				t.Errorf("%s positional %s must be a non-empty GUID", c.action, arg.Name)
			}
		}
		for name := range flagsToMap(action.Flags) {
			switch name {
			case "id", "tenant-id", "tenant-guid", "user-id", "user-guid":
				t.Errorf("%s must not take an identity or integer id flag (--%s)", c.action, name)
			}
		}
	}
}

func TestAssociationFlagsMapToTheBackendModels(t *testing.T) {
	d := findDomain("association")
	if d == nil {
		t.Fatal("expected association domain to be registered")
	}

	want := map[string]map[string]struct {
		body     string
		required bool
	}{
		"meeting-create": {"title": {"title", true}, "kind": {"kind", true}, "held-on": {"heldOn", true}, "location": {"location", false}, "minutes-text": {"minutesText", false}, "minutes-document-guid": {"minutesDocumentGuid", false}},
		"meeting-update": {"title": {"title", true}, "kind": {"kind", true}, "held-on": {"heldOn", true}, "location": {"location", false}, "minutes-text": {"minutesText", false}, "minutes-document-guid": {"minutesDocumentGuid", false}},
		"vote-create":    {"title": {"title", true}, "description": {"description", false}, "closes-at": {"closesAt", false}},
		"ballot-cast":    {"choice": {"choice", true}},
		"notice-post":    {"title": {"title", true}, "body": {"body", true}},
	}
	for actionName, flags := range want {
		got := flagsToMap(findAction(d, actionName).Flags)
		if len(got) != len(flags) {
			t.Errorf("%s has %d flags, want %d", actionName, len(got), len(flags))
		}
		for name, expected := range flags {
			flag, ok := got[name]
			if !ok || flag.BodyName != expected.body || flag.Required != expected.required {
				t.Errorf("--%s on %s must map to %s (required %v), got %+v", name, actionName, expected.body, expected.required, flag)
			}
		}
	}

	kind := flagsToMap(findAction(d, "meeting-create").Flags)["kind"]
	if len(kind.AllowedValues) != 3 || !containsExact(kind.AllowedValues, "residents") {
		t.Errorf("--kind must allow general, board and residents, got %v", kind.AllowedValues)
	}
	choice := flagsToMap(findAction(d, "ballot-cast").Flags)["choice"]
	for _, value := range []string{"yes", "no", "abstain"} {
		if !containsExact(choice.AllowedValues, value) {
			t.Errorf("--choice must allow %s, got %v", value, choice.AllowedValues)
		}
	}
	if len(choice.AllowedValues) != 3 {
		t.Errorf("--choice must allow exactly yes, no and abstain, got %v", choice.AllowedValues)
	}
}

func associationTestFactory(t *testing.T, handler http.HandlerFunc) (APIClientFactory, *logging.Logger) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("UTEAMUP_TENANT_GUID", "52e73fd7-cd7b-45c6-9872-b6be9415203c")
	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)
	if err := auth.SaveToken(&auth.TokenData{APIOrigin: server.URL, AccessToken: "test-token", AuthMethod: "apikey", ExpiresAt: time.Now().Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	logger := logging.New(logging.LevelError)
	return func() (*client.APIClient, error) {
		return client.NewAPIClient(server.URL, time.Second, true, client.RetryOptions{}, logger), nil
	}, logger
}

func TestAssociationBallotCastSendsTheChoiceToTheVoteRoute(t *testing.T) {
	var body map[string]any
	factory, logger := associationTestFactory(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/association/votes/"+testVoteGUID+"/ballot" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"guid":"` + testVoteGUID + `","myChoice":"abstain"}`))
	})
	format := "json"
	command := buildActionCommand(findDomain("association"), *findAction(findDomain("association"), "ballot-cast"), factory, logger, &format, nil)
	command.SetArgs([]string{testVoteGUID, "--choice", "abstain"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if body["choice"] != "abstain" {
		t.Errorf("choice = %v, want abstain", body["choice"])
	}
	if _, leaked := body["voteGuid"]; leaked {
		t.Error("voteGuid belongs in the path, not the body")
	}
}

func TestAssociationRejectsBadInputBeforeAnyRequest(t *testing.T) {
	cases := []struct {
		name, action string
		args         []string
	}{
		{"invalid ballot choice", "ballot-cast", []string{testVoteGUID, "--choice", "maybe"}},
		{"integer vote id", "ballot-cast", []string{"42", "--choice", "yes"}},
		{"non-GUID minutes document", "meeting-create", []string{"--title", "Fundur", "--kind", "general", "--held-on", "2026-11-05T18:00:00Z", "--minutes-document-guid", "7"}},
		{"invalid meeting kind", "meeting-create", []string{"--title", "Fundur", "--kind", "annual", "--held-on", "2026-11-05T18:00:00Z"}},
		{"notice without body", "notice-post", []string{"--title", "Vatnslaust"}},
		{"extra positional argument", "votes-list", []string{testVoteGUID}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			factory, logger := associationTestFactory(t, func(w http.ResponseWriter, r *http.Request) {
				t.Errorf("no request may be sent, got %s %s", r.Method, r.URL.Path)
			})
			format := "json"
			d := findDomain("association")
			command := buildActionCommand(d, *findAction(d, c.action), factory, logger, &format, nil)
			command.SilenceUsage = true
			command.SilenceErrors = true
			command.SetArgs(c.args)
			if err := command.Execute(); err == nil {
				t.Fatal("expected the command to fail")
			}
		})
	}
}
