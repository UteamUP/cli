package registry

import (
	"errors"
	"strings"
	"testing"

	"github.com/uteamup/cli/internal/client"
	"github.com/uteamup/cli/internal/logging"
)

func TestBodyMimeEnumAcceptsDeclaredValuesBeforeTransport(t *testing.T) {
	for _, contentType := range []string{"image/png", "image/jpeg", "image/gif"} {
		t.Run(contentType, func(t *testing.T) {
			reached := errors.New("transport reached")
			called := false
			format := "json"
			action := Action{Name: "upload-image", Flags: []FlagDef{{Name: "content-type", BodyName: "contentType",
				Type: "string", Required: true, AllowedValues: []string{"image/png", "image/jpeg"}}}}
			command := buildActionCommand(&Domain{Name: "part"}, action, func() (*client.APIClient, error) {
				called = true
				return nil, reached
			}, logging.New(logging.LevelError), &format, nil)
			command.SetArgs([]string{"--content-type", contentType})
			err := command.Execute()
			if contentType == "image/gif" {
				if called || err == nil || !strings.Contains(err.Error(), "must be one of") {
					t.Fatalf("undeclared MIME type reached transport: called=%v err=%v", called, err)
				}
			} else if !called || !errors.Is(err, reached) {
				t.Fatalf("declared MIME type was blocked: called=%v err=%v", called, err)
			}
		})
	}
}

func TestRouteEnumStillRejectsMimeSeparatorsWithEveryFlagMapping(t *testing.T) {
	for _, placeholder := range []string{"contentType", "content-type", "mime"} {
		flag := FlagDef{Name: "content-type", Type: "string", AllowedValues: []string{"image/png"}}
		if placeholder == "mime" {
			flag.BodyName = "mime"
		}
		action := Action{Name: "get", RESTPath: "files/{" + placeholder + "}", Flags: []FlagDef{flag}}
		if err := validateActionDefinition(action); err == nil {
			t.Fatalf("unsafe route enum accepted for %s", placeholder)
		}
	}
}
