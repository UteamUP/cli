package cleanup

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func keysOf(c Catalog, typ string) map[string]bool {
	out := map[string]bool{}
	for _, e := range c.Entries {
		if e.Type == typ {
			out[e.Key] = true
		}
	}
	return out
}

func TestScan_DerivesMatchingKeys(t *testing.T) {
	root := t.TempDir()

	write(t, root, "UteamUP_Backend/UteamUP_API/Controllers/AssetController.cs", `
public class AssetController : ControllerBase
{
    public AssetController(IDeps d) {}

    [HttpGet("{id}")]
    public async Task<IActionResult> GetSummary(int id) { return Ok(); }

    [NonAction]
    public void Helper() {}

    [HttpPost]
    public async Task<IActionResult> Create([FromBody] X x) { return Ok(); }
}`)

	write(t, root, "UteamUP_Backend/UteamUP_API/Repositories/Interfaces/IAssetRepository.cs", `
public interface IAssetRepository
{
    Task<List<Asset>> GetAllAsync(Guid tenantGuid);
    Task<Asset?> GetByGuidAsync(Guid guid);
}`)

	write(t, root, "UteamUP_Frontend/pages/workorders/[id]/index.vue", "<template></template>")
	write(t, root, "UteamUP_Frontend/pages/index.vue", "<template></template>")
	write(t, root, "UteamUP_Frontend/components/Foo.vue", "<template></template>")
	write(t, root, "UteamUP_Frontend/brand_components/Bar.vue", "<template></template>")
	// Foo is instrumented via a v-usage somewhere; Bar is not.
	write(t, root, "UteamUP_Frontend/pages/uses-foo.vue", `<template><UteamupCard v-usage="'Foo'" /></template>`)

	write(t, root, "UteamUP_Mobile/lib/core/router/app_router.dart", `
final r = GoRouter(routes: [
  GoRoute(path: '/home'),
  GoRoute(path: '/work/workorder/:guid'),
]);`)
	write(t, root, "UteamUP_Mobile/lib/widgets/u_card.dart", `class UCard extends StatelessWidget {}`)

	cat := Scan(root)

	be := keysOf(cat, TypeBackendEndpoint)
	if !be["Asset.GetSummary"] || !be["Asset.Create"] {
		t.Errorf("backend endpoints missing: %v", be)
	}
	if be["Asset.Helper"] {
		t.Errorf("[NonAction] Helper must not be tracked: %v", be)
	}

	repo := keysOf(cat, TypeBackendRepository)
	if !repo["AssetRepository.GetAllAsync"] || !repo["AssetRepository.GetByGuidAsync"] {
		t.Errorf("repository methods missing: %v", repo)
	}

	pages := keysOf(cat, TypeFrontendPage)
	if !pages["/workorders/[id]"] || !pages["/"] {
		t.Errorf("frontend pages missing: %v", pages)
	}

	var fooInstrumented, barInstrumented, fooFound, barFound bool
	for _, e := range cat.Entries {
		if e.Type == TypeFrontendComponent && e.Key == "Foo" {
			fooFound = true
			fooInstrumented = e.Instrumented
		}
		if e.Type == TypeFrontendComponent && e.Key == "Bar" {
			barFound = true
			barInstrumented = e.Instrumented
		}
	}
	if !fooFound || !barFound {
		t.Errorf("frontend components missing (foo=%v bar=%v)", fooFound, barFound)
	}
	if !fooInstrumented {
		t.Errorf("Foo should be instrumented (has v-usage)")
	}
	if barInstrumented {
		t.Errorf("Bar should NOT be instrumented (no v-usage)")
	}

	mp := keysOf(cat, TypeMobilePage)
	if !mp["/work/workorder/[guid]"] || !mp["/home"] {
		t.Errorf("mobile pages missing: %v", mp)
	}
	mc := keysOf(cat, TypeMobileComponent)
	if !mc["UCard"] {
		t.Errorf("mobile components missing: %v", mc)
	}
}
