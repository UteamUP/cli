// Package cleanup natively scans the UteamUP monorepo to build the complete catalog of code units
// (backend controller actions, backend repository interface methods, frontend pages, frontend
// components, mobile pages, mobile components), then diffs that catalog against runtime usage so the
// `uteamup cleanup` command can report what is never exercised. No external script — everything is
// here in the CLI.
package cleanup

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Usage type names — must match the backend UsageType enum string values exactly.
const (
	TypeBackendEndpoint   = "BackendEndpoint"
	TypeBackendRepository = "BackendRepository"
	TypeFrontendPage      = "FrontendPage"
	TypeFrontendComponent = "FrontendComponent"
	TypeMobilePage        = "MobilePage"
	TypeMobileComponent   = "MobileComponent"
)

// CatalogEntry is one code unit found by the scanner.
type CatalogEntry struct {
	Type         string `json:"type"`
	Key          string `json:"key"`
	File         string `json:"file"`
	Instrumented bool   `json:"instrumented"` // opt-in types only (components): is it wired to report?
}

// Catalog is the full scan plus any non-fatal warnings (e.g. duplicate component basenames).
type Catalog struct {
	Entries  []CatalogEntry
	Warnings []string
}

var (
	csClassRe   = regexp.MustCompile(`\bclass\s+(\w+)Controller\b`)
	csMethodRe  = regexp.MustCompile(`(\w+)\s*\(`)
	ifaceRe     = regexp.MustCompile(`\binterface\s+(I\w+Repository)\b`)
	ifaceMethRe = regexp.MustCompile(`(\w+)\s*\(`)
	goRouteRe   = regexp.MustCompile(`path:\s*'([^']+)'`)
	dartWidgetRe = regexp.MustCompile(`\bclass\s+(\w+)\s+extends\s+(?:Consumer)?(?:Stateless|Stateful)Widget\b`)
	vUsageRe    = regexp.MustCompile(`v-usage="'([^']+)'"`)
	dartCompRe  = regexp.MustCompile(`trackComponentUsage\(\s*'([^']+)'\s*\)`)
	routeParamRe = regexp.MustCompile(`:([A-Za-z0-9_]+)`)
)

// Scan walks the monorepo rooted at root (the dir containing UteamUP_Backend/ etc.).
func Scan(root string) Catalog {
	cat := Catalog{}
	cat.scanBackend(root)
	cat.scanFrontend(root)
	cat.scanMobile(root)
	return cat
}

func (c *Catalog) add(t, key, file string, instrumented bool) {
	if key == "" {
		return
	}
	c.Entries = append(c.Entries, CatalogEntry{Type: t, Key: key, File: file, Instrumented: instrumented})
}

// --- Backend ---

func (c *Catalog) scanBackend(root string) {
	ctrlDir := filepath.Join(root, "UteamUP_Backend", "UteamUP_API", "Controllers")
	walkFiles(ctrlDir, ".cs", func(path string, content string) {
		c.parseController(path, content, root)
	})
	ifaceDir := filepath.Join(root, "UteamUP_Backend", "UteamUP_API", "Repositories", "Interfaces")
	walkFiles(ifaceDir, ".cs", func(path string, content string) {
		c.parseRepositoryInterface(path, content, root)
	})
}

func (c *Catalog) parseController(path, content, root string) {
	m := csClassRe.FindStringSubmatch(content)
	if m == nil {
		return
	}
	controller := m[1] // class name minus the "Controller" suffix, matching ControllerActionDescriptor.ControllerName
	rel := relPath(root, path)

	pendingHttp := false
	skipNext := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		if strings.Contains(trimmed, "[NonAction]") {
			skipNext = true
			continue
		}
		if strings.Contains(trimmed, "[Http") {
			pendingHttp = true
			continue
		}
		// An action method declaration: a public member with a parameter list, not the class line.
		isMethod := strings.HasPrefix(trimmed, "public ") && !strings.Contains(trimmed, " class ") && strings.Contains(trimmed, "(")
		if isMethod {
			if pendingHttp && !skipNext {
				if name := firstCall(csMethodRe, trimmed); name != "" {
					c.add(TypeBackendEndpoint, controller+"."+name, rel, true)
				}
			}
			pendingHttp = false
			skipNext = false
		}
	}
}

func (c *Catalog) parseRepositoryInterface(path, content, root string) {
	m := ifaceRe.FindStringSubmatch(content)
	if m == nil {
		return
	}
	repo := strings.TrimPrefix(m[1], "I") // IAssetRepository -> AssetRepository
	rel := relPath(root, path)

	inBody := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if !inBody {
			if strings.Contains(trimmed, "{") {
				inBody = true
			}
			continue
		}
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}
		// Interface method declaration: has a parameter list. (Properties have no "(".)
		if strings.Contains(trimmed, "(") {
			if name := firstCall(ifaceMethRe, trimmed); name != "" {
				c.add(TypeBackendRepository, repo+"."+name, rel, true)
			}
		}
	}
}

// --- Frontend ---

func (c *Catalog) scanFrontend(root string) {
	fe := filepath.Join(root, "UteamUP_Frontend")

	pagesDir := filepath.Join(fe, "pages")
	walkFiles(pagesDir, ".vue", func(path string, _ string) {
		if key := nuxtRouteFromPath(pagesDir, path); key != "" {
			c.add(TypeFrontendPage, key, relPath(root, path), true)
		}
	})

	// Component instrumentation set: any `v-usage="'Name'"` anywhere makes that component eligible.
	instrumented := map[string]bool{}
	for _, dir := range []string{filepath.Join(fe, "pages"), filepath.Join(fe, "components"), filepath.Join(fe, "brand_components"), filepath.Join(fe, "layouts")} {
		walkFiles(dir, ".vue", func(_ string, content string) {
			for _, mm := range vUsageRe.FindAllStringSubmatch(content, -1) {
				instrumented[mm[1]] = true
			}
		})
	}

	seenBasename := map[string]string{}
	for _, dir := range []string{filepath.Join(fe, "components"), filepath.Join(fe, "brand_components")} {
		walkFiles(dir, ".vue", func(path string, _ string) {
			name := strings.TrimSuffix(filepath.Base(path), ".vue")
			if prev, dup := seenBasename[name]; dup {
				c.Warnings = append(c.Warnings, "duplicate component basename '"+name+"' ("+prev+" and "+relPath(root, path)+") — usage key contract requires unique component names")
			}
			seenBasename[name] = relPath(root, path)
			c.add(TypeFrontendComponent, name, relPath(root, path), instrumented[name])
		})
	}
}

// --- Mobile ---

func (c *Catalog) scanMobile(root string) {
	mobile := filepath.Join(root, "UteamUP_Mobile", "lib")
	if _, err := os.Stat(mobile); err != nil {
		return // mobile not present in this checkout
	}

	// MobilePage catalog = GoRoute path strings (the route observer reports the templated path).
	router := filepath.Join(mobile, "core", "router", "app_router.dart")
	if content, err := os.ReadFile(router); err == nil {
		for _, mm := range goRouteRe.FindAllStringSubmatch(string(content), -1) {
			p := mm[1]
			if !strings.HasPrefix(p, "/") {
				continue
			}
			c.add(TypeMobilePage, normalizeRouteParams(p), relPath(root, router), true)
		}
	}

	// MobileComponent catalog = widget classes; instrumented if they call trackComponentUsage('Name').
	instrumented := map[string]bool{}
	walkFiles(mobile, ".dart", func(_ string, content string) {
		for _, mm := range dartCompRe.FindAllStringSubmatch(content, -1) {
			instrumented[mm[1]] = true
		}
	})
	walkFiles(mobile, ".dart", func(path string, content string) {
		for _, mm := range dartWidgetRe.FindAllStringSubmatch(content, -1) {
			name := mm[1]
			c.add(TypeMobileComponent, name, relPath(root, path), instrumented[name])
		}
	})
}

// --- helpers ---

func walkFiles(dir, ext string, fn func(path, content string)) {
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ext) {
			return nil
		}
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		fn(path, string(b))
		return nil
	})
}

func firstCall(re *regexp.Regexp, line string) string {
	m := re.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	name := m[1]
	// Filter obvious non-method keywords that can precede a "(".
	switch name {
	case "if", "for", "foreach", "while", "switch", "catch", "lock", "using", "return", "get", "set":
		return ""
	}
	return name
}

// nuxtRouteFromPath maps pages/workorders/[id]/index.vue -> /workorders/[id], keeping bracket params.
func nuxtRouteFromPath(pagesDir, path string) string {
	rel, err := filepath.Rel(pagesDir, path)
	if err != nil {
		return ""
	}
	rel = filepath.ToSlash(rel)
	rel = strings.TrimSuffix(rel, ".vue")
	rel = strings.TrimSuffix(rel, "/index")
	if rel == "index" || rel == "" {
		return "/"
	}
	return "/" + rel
}

// normalizeRouteParams turns go_router ":guid" into Nuxt-style "[guid]" so mobile page keys match
// what the mobile reporter sends.
func normalizeRouteParams(p string) string {
	p = routeParamRe.ReplaceAllString(p, "[$1]")
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

func relPath(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil {
		return filepath.ToSlash(rel)
	}
	return path
}
