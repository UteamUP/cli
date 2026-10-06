package registry

import "testing"

func TestAiTierDomainsExposeNewBackendEndpoints(t *testing.T) {
	// Placeholders use the camelCase key, so the dashed positional args carry a matching
	// BodyName: executeAction stores positional args under BodyName, else their raw Name, and
	// expandPathTemplate looks them up by that key. TestEveryRESTPathPlaceholderResolvesToAStoredArgKey
	// guards the pairing for every domain.
	ocr := findDomainAction(t, "meter-reading", "ocr")
	if ocr.RESTPath != "{assetGuid}/meter-readings/{attributeDefinitionGuid}/ocr" || ocr.HTTPMethod != "POST" {
		t.Fatalf("meter-reading ocr route = method %q path %q", ocr.HTTPMethod, ocr.RESTPath)
	}

	brief := findDomainAction(t, "workforce-ai", "daily-brief")
	if brief.RESTPath != "daily-brief" || brief.HTTPMethod != "POST" {
		t.Fatalf("daily-brief route = method %q path %q", brief.HTTPMethod, brief.RESTPath)
	}
	briefFlags := make(map[string]FlagDef)
	for _, f := range brief.Flags {
		briefFlags[f.Name] = f
	}
	if briefFlags["currentLatitude"].Type != "float" || briefFlags["currentLongitude"].Type != "float" {
		t.Fatalf("daily-brief GPS flags missing or wrong type: %#v", briefFlags)
	}

	prefill := findDomainAction(t, "work-permit-ai", "prefill")
	if prefill.RESTPath != "by-guid/{workPermitGuid}/ai-prefill" || prefill.HTTPMethod != "POST" {
		t.Fatalf("prefill route = method %q path %q", prefill.HTTPMethod, prefill.RESTPath)
	}

	usage := findDomainAction(t, "ai-usage", "summary")
	if usage.RESTPath != "summary" || usage.HTTPMethod != "" {
		t.Fatalf("ai-usage summary route = method %q path %q, want derived GET summary", usage.HTTPMethod, usage.RESTPath)
	}
}
