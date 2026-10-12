// Package cleanup natively scans the UteamUP monorepo to build the complete catalog of code units
// (backend controller actions, backend repository interface methods, frontend pages, frontend
// components, mobile pages, mobile components), then diffs that catalog against runtime usage so the
// `uteamup cleanup` command can report what is never exercised. No external script — everything is
// here in the CLI.
package cleanup

import (
	"errors"
	"fmt"
	"github.com/uteamup/cli/internal/security"
	"io"
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
	Entries    []CatalogEntry
	Warnings   []string
	Incomplete bool
	scan       *scanSession
}

var (
	csClassRe    = regexp.MustCompile(`\bclass\s+(\w+)Controller\b`)
	csMethodRe   = regexp.MustCompile(`(\w+)\s*\(`)
	ifaceRe      = regexp.MustCompile(`\binterface\s+(I\w+Repository)\b`)
	ifaceMethRe  = regexp.MustCompile(`(\w+)\s*\(`)
	goRouteRe    = regexp.MustCompile(`path:\s*'([^']+)'`)
	dartWidgetRe = regexp.MustCompile(`\bclass\s+(\w+)\s+extends\s+(?:Consumer)?(?:Stateless|Stateful)Widget\b`)
	vUsageRe     = regexp.MustCompile(`v-usage="'([^']+)'"`)
	dartCompRe   = regexp.MustCompile(`trackComponentUsage\(\s*'([^']+)'\s*\)`)
	routeParamRe = regexp.MustCompile(`:([A-Za-z0-9_]+)`)
)

// Scan walks the monorepo rooted at root (the dir containing UteamUP_Backend/ etc.).
func Scan(root string) Catalog {
	return scanWithLimits(root, scanLimits{bytes: 256 * 1024 * 1024, entries: 250000, units: 100000, depth: 128})
}

func scanWithLimits(root string, limits scanLimits) Catalog {
	anchor, err := security.OpenDirectory(root)
	if err != nil {
		return Catalog{Incomplete: true, Warnings: []string{"Scan incomplete: root must be an existing directory without symbolic links."}}
	}
	defer anchor.Close()
	cat := Catalog{scan: &scanSession{root: anchor, absolute: root, limits: limits, cache: make(map[string]string)}}
	cat.scanBackend(root)
	cat.scanFrontend(root)
	cat.scanMobile(root)
	if cat.scan.failure != "" {
		cat.Incomplete = true
		cat.Warnings = append(cat.Warnings, "Scan incomplete: "+cat.scan.failure+"; unused results are not trustworthy.")
	}
	cat.scan = nil
	return cat
}

func (c *Catalog) add(t, key, file string, instrumented bool) {
	if key == "" {
		return
	}
	if c.scan != nil && (c.scan.failure != "" || len(c.Entries) >= c.scan.limits.units) {
		if c.scan.failure == "" {
			c.scan.failure = "code-unit budget exhausted"
		}
		return
	}
	c.Entries = append(c.Entries, CatalogEntry{Type: t, Key: key, File: file, Instrumented: instrumented})
}

// --- Backend ---

func (c *Catalog) scanBackend(root string) {
	ctrlDir := filepath.Join(root, "UteamUP_Backend", "UteamUP_API", "Controllers")
	c.walkFiles(ctrlDir, ".cs", func(path string, content string) {
		c.parseController(path, content, root)
	})
	ifaceDir := filepath.Join(root, "UteamUP_Backend", "UteamUP_API", "Repositories", "Interfaces")
	c.walkFiles(ifaceDir, ".cs", func(path string, content string) {
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

	pendingHTTP := false
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
			pendingHTTP = true
			continue
		}
		// An action method declaration: a public member with a parameter list, not the class line.
		isMethod := strings.HasPrefix(trimmed, "public ") && !strings.Contains(trimmed, " class ") && strings.Contains(trimmed, "(")
		if isMethod {
			if pendingHTTP && !skipNext {
				if name := firstCall(csMethodRe, trimmed); name != "" {
					c.add(TypeBackendEndpoint, controller+"."+name, rel, true)
				}
			}
			pendingHTTP = false
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
	c.walkFiles(pagesDir, ".vue", func(path string, _ string) {
		if key := nuxtRouteFromPath(pagesDir, path); key != "" {
			c.add(TypeFrontendPage, key, relPath(root, path), true)
		}
	})

	// Component instrumentation set: any `v-usage="'Name'"` anywhere makes that component eligible.
	instrumented := map[string]bool{}
	for _, dir := range []string{filepath.Join(fe, "pages"), filepath.Join(fe, "components"), filepath.Join(fe, "brand_components"), filepath.Join(fe, "layouts")} {
		c.walkFiles(dir, ".vue", func(_ string, content string) {
			for _, mm := range vUsageRe.FindAllStringSubmatch(content, -1) {
				instrumented[mm[1]] = true
			}
		})
	}

	seenBasename := map[string]string{}
	for _, dir := range []string{filepath.Join(fe, "components"), filepath.Join(fe, "brand_components")} {
		c.walkFiles(dir, ".vue", func(path string, _ string) {
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

	// MobilePage catalog = GoRoute path strings (the route observer reports the templated path).
	router := filepath.Join(mobile, "core", "router", "app_router.dart")
	if content, err := c.scan.readPath(router); err == nil {
		for _, mm := range goRouteRe.FindAllStringSubmatch(string(content), -1) {
			p := mm[1]
			if !strings.HasPrefix(p, "/") {
				continue
			}
			c.add(TypeMobilePage, normalizeRouteParams(p), relPath(root, router), true)
		}
	} else if !os.IsNotExist(err) && c.scan.failure == "" {
		c.scan.failure = "unsafe, oversized, or unreadable mobile router"
	}

	// MobileComponent catalog = widget classes; instrumented if they call trackComponentUsage('Name').
	instrumented := map[string]bool{}
	c.walkFiles(mobile, ".dart", func(_ string, content string) {
		for _, mm := range dartCompRe.FindAllStringSubmatch(content, -1) {
			instrumented[mm[1]] = true
		}
	})
	c.walkFiles(mobile, ".dart", func(path string, content string) {
		for _, mm := range dartWidgetRe.FindAllStringSubmatch(content, -1) {
			name := mm[1]
			c.add(TypeMobileComponent, name, relPath(root, path), instrumented[name])
		}
	})
}

// --- helpers ---

type scanLimits struct {
	bytes                 int64
	entries, units, depth int
}
type scanSession struct {
	root     *os.Root
	absolute string
	limits   scanLimits
	cache    map[string]string
	failure  string
}

var errScanBudget = errors.New("scan budget exhausted")

// Each subdirectory is opened through its parent descriptor. Links and directory
// swaps cannot redirect enumeration outside the selected monorepo.
func openScanDirectory(root *os.Root, relative string) (*os.Root, error) {
	current, err := root.OpenRoot(".")
	if err != nil {
		return nil, err
	}
	for _, part := range strings.Split(filepath.Clean(relative), string(filepath.Separator)) {
		if part == "." {
			continue
		}
		if part == ".." || filepath.IsAbs(part) {
			current.Close()
			return nil, fmt.Errorf("path escapes scan root")
		}
		before, err := current.Lstat(part)
		if err != nil {
			current.Close()
			return nil, err
		}
		if !before.IsDir() {
			current.Close()
			return nil, fmt.Errorf("directory is not a regular directory")
		}
		next, err := current.OpenRoot(part)
		if err != nil {
			current.Close()
			return nil, err
		}
		after, err := next.Stat(".")
		current.Close()
		if err != nil || !os.SameFile(before, after) {
			next.Close()
			return nil, fmt.Errorf("directory identity changed")
		}
		current = next
	}
	return current, nil
}

func (s *scanSession) readPath(path string) ([]byte, error) {
	if s.failure != "" {
		return nil, errScanBudget
	}
	relative, err := filepath.Rel(s.absolute, path)
	if err != nil {
		return nil, err
	}
	dir, err := openScanDirectory(s.root, filepath.Dir(relative))
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	return s.read(dir, filepath.Base(relative), path)
}

func (s *scanSession) read(dir *os.Root, name, path string) ([]byte, error) {
	if content, ok := s.cache[path]; ok {
		return []byte(content), nil
	}
	before, err := dir.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || before.Size() > 4*1024*1024 {
		return nil, fmt.Errorf("source is not a bounded regular file")
	}
	if before.Size() > s.limits.bytes {
		s.failure = "shared source-byte budget exhausted"
		return nil, errScanBudget
	}
	data, err := security.ReadRootFile(dir, name, min(s.limits.bytes, int64(4*1024*1024)))
	if err != nil {
		return nil, err
	}
	s.limits.bytes -= int64(len(data))
	s.cache[path] = string(data)
	return data, nil
}

func (c *Catalog) walkFiles(dir, ext string, fn func(path, content string)) {
	s := c.scan
	if s.failure != "" {
		return
	}
	relative, err := filepath.Rel(s.absolute, dir)
	if err != nil {
		s.failure = "invalid source directory"
		return
	}
	anchor, err := openScanDirectory(s.root, relative)
	if err != nil {
		if !os.IsNotExist(err) {
			s.failure = "unsafe or unreadable source directory"
		}
		return
	}
	defer anchor.Close()
	var walk func(*os.Root, string, int) error
	walk = func(root *os.Root, path string, depth int) error {
		if depth > s.limits.depth {
			s.failure = "directory-depth budget exhausted"
			return errScanBudget
		}
		stream, err := root.Open(".")
		if err != nil {
			return err
		}
		defer stream.Close()
		for {
			batch, err := stream.ReadDir(128)
			if err != nil && err != io.EOF {
				return err
			}
			for _, entry := range batch {
				if s.limits.entries <= 0 {
					s.failure = "shared directory-entry budget exhausted"
					return errScanBudget
				}
				s.limits.entries--
				childPath := filepath.Join(path, entry.Name())
				if entry.IsDir() {
					child, err := openScanDirectory(root, entry.Name())
					if err != nil {
						return err
					}
					err = walk(child, childPath, depth+1)
					child.Close()
					if err != nil {
						return err
					}
				} else if entry.Type().IsRegular() && strings.HasSuffix(entry.Name(), ext) {
					data, err := s.read(root, entry.Name(), childPath)
					if err != nil {
						return err
					}
					fn(childPath, string(data))
					if s.failure != "" {
						return errScanBudget
					}
				}
			}
			if err == io.EOF {
				return nil
			}
		}
	}
	if err := walk(anchor, dir, 0); err != nil && s.failure == "" {
		s.failure = "unsafe, oversized, or unreadable source"
	}
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
