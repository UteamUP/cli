package registry

import "testing"

func TestLogbookImportGetUsesGuidRoute(t *testing.T) {
	domain := findDomain("logbook-import")
	if domain == nil {
		t.Fatal("expected logbook-import domain to be registered")
	}
	if domain.APIPath != "/api/logbookimport" {
		t.Fatalf("APIPath = %q, want /api/logbookimport", domain.APIPath)
	}

	action := findDomainAction(t, "logbook-import", "get")
	if len(action.Args) != 1 {
		t.Fatalf("get args = %+v, want one GUID positional arg", action.Args)
	}
	argument := action.Args[0]
	if argument.Name != "externalGuid" || argument.Type != "string" || !argument.Required {
		t.Fatalf("get identity = %+v, want required externalGuid string", argument)
	}
	if len(action.Flags) != 0 {
		t.Fatalf("get flags = %+v, want no integer batch-id flag", action.Flags)
	}

	guid := "77777777-7777-4777-8777-777777777777"
	path, consumed := buildRESTPath(
		domain,
		*action,
		map[string]any{"externalGuid": guid},
	)
	if path != "/api/logbookimport/"+guid {
		t.Fatalf("resolved path = %q, want GUID logbook route", path)
	}
	if len(consumed) != 1 || consumed[0] != "externalGuid" {
		t.Fatalf("consumed args = %v, want [externalGuid]", consumed)
	}
}
