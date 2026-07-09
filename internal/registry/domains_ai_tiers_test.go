package registry

import "testing"

func TestAiProviderDomainUsesTenantProviderRoute(t *testing.T) {
	d := findDomain("ai-provider")
	if d == nil {
		t.Fatal("expected ai-provider domain to be registered")
	}
	if d.APIPath != "/api/tenant-ai-provider" {
		t.Fatalf("ai-provider APIPath = %q, want /api/tenant-ai-provider", d.APIPath)
	}

	a := findDomainAction(t, "ai-provider", "test-connection")
	if a.RESTPath != "test-connection" || a.HTTPMethod != "POST" {
		t.Fatalf("test-connection route = method %q path %q, want POST test-connection", a.HTTPMethod, a.RESTPath)
	}
}

func TestAiTierDomainsExposeNewBackendEndpoints(t *testing.T) {
	ocr := findDomainAction(t, "meter-reading", "ocr")
	if ocr.RESTPath != "{asset-guid}/meter-readings/{attribute-definition-guid}/ocr" || ocr.HTTPMethod != "POST" {
		t.Fatalf("meter-reading ocr route = method %q path %q", ocr.HTTPMethod, ocr.RESTPath)
	}

	brief := findDomainAction(t, "workforce-ai", "daily-brief")
	if brief.RESTPath != "daily-brief" || brief.HTTPMethod != "POST" {
		t.Fatalf("daily-brief route = method %q path %q", brief.HTTPMethod, brief.RESTPath)
	}

	prefill := findDomainAction(t, "work-permit-ai", "prefill")
	if prefill.RESTPath != "by-guid/{work-permit-guid}/ai-prefill" || prefill.HTTPMethod != "POST" {
		t.Fatalf("prefill route = method %q path %q", prefill.HTTPMethod, prefill.RESTPath)
	}

	usage := findDomainAction(t, "ai-usage", "summary")
	if usage.RESTPath != "summary" || usage.HTTPMethod != "" {
		t.Fatalf("ai-usage summary route = method %q path %q, want derived GET summary", usage.HTTPMethod, usage.RESTPath)
	}
}
