package registry

import "testing"

func TestQualificationCatalogueTagsUsePublicGuids(t *testing.T) {
	for _, domain := range []string{"certificate", "skill"} {
		for _, name := range []string{"create", "update"} {
			action := findDomainAction(t, domain, name)
			flag := actionFlagByName(t, action, "tag-guids")
			if flag.BodyName != "tagGuids" || flag.Type != "stringSlice" {
				t.Fatalf("%s %s must bind the public GUID list: %+v", domain, name, flag)
			}
			if name == "update" {
				if len(action.Args) != 1 || action.Args[0].Name != "externalGuid" || action.Args[0].Type != "non-empty-uuid" {
					t.Fatalf("%s update must use a public GUID: %+v", domain, action.Args)
				}
				guid := "11111111-1111-4111-8111-111111111111"
				path, _ := buildRESTPath(findDomain(domain), *action, map[string]any{"externalGuid": guid})
				if path != "/api/"+domain+"/by-guid/"+guid {
					t.Fatalf("Wrong GUID update route: %s", path)
				}
			}
		}
	}
}
