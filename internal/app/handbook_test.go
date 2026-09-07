package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// handbookRoute is one destination in the Gelium Handbook IA. Contract is a
// content marker unique to the page — the thing that makes the page real
// rather than a stub; H1 is the rendered heading level 1 of that page.
type handbookRoute struct {
	path     string
	label    string
	h1       string
	contract string
}

// handbookRoutes is the single source of truth for the Handbook pages
// (Handbook sections + Design Principles), mirroring the sidebar group.
var handbookRoutes = []handbookRoute{
	{path: "/docs/information-architecture", label: "Information architecture", h1: "Information architecture", contract: "concept before reference"},
	{path: "/docs/screens", label: "Screens", h1: "Screens", contract: "Build checklist"},
	{path: "/docs/page-section-architecture", label: "Page + section architecture", h1: "Page + section architecture", contract: "SECTION-CONTRACT"},
	{path: "/docs/section-references", label: "Section references", h1: "Section references", contract: "not a moodboard"},
	{path: "/docs/journeys", label: "Journeys", h1: "Journeys", contract: "JOURNEY-LINEAR"},
	{path: "/docs/data-display", label: "Data display", h1: "Data display", contract: "DATA-TABLE"},
	{path: "/docs/feedback", label: "Feedback", h1: "Feedback", contract: "error summary"},
	{path: "/docs/density", label: "Density", h1: "Density and shell", contract: "Density modes"},
	{path: "/docs/motion", label: "Motion", h1: "Motion", contract: "prefers-reduced-motion"},
	{path: "/docs/ui-definition-of-done", label: "UI definition of done", h1: "UI definition of done", contract: "DoD checklist"},
	{path: "/docs/agent-workflow", label: "Agent workflow", h1: "Agent workflow", contract: "WF-SHAPE"},
	{path: "/docs/choose-the-right-control", label: "Choose the right control", h1: "Choose the right control", contract: "Radio vs Select"},
	{path: "/docs/forms", label: "Forms", h1: "Forms contract", contract: "inputmode"},
	{path: "/docs/compare", label: "Why Gelium", h1: "Why Gelium (comparison)", contract: "When NOT to choose Gelium"},
	{path: "/docs/performance", label: "Performance", h1: "Performance stance", contract: "CSS is the biggest asset by design"},
	{path: "/docs/responsive", label: "Responsive", h1: "Design for screen sizes, not devices", contract: "must not mask"},
	{path: "/docs/media", label: "Media", h1: "Media", contract: "MEDIA-IMAGE"},
	{path: "/docs/themes", label: "Themes", h1: "Themes", contract: "?theme=basecoat"},
	{path: "/docs/tokens", label: "Tokens", h1: "Tokens", contract: "--ui-color-primary"},
	{path: "/docs/server-contracts", label: "Server contracts", h1: "Server contracts", contract: "HX-Trigger"},
	{path: "/docs/accessibility", label: "Accessibility", h1: "Accessibility", contract: "forced-colors"},
	{path: "/docs/principles", label: "Design principles", h1: "Design principles", contract: "Native semantics first"},
	{path: "/docs/browser-support", label: "Browser support", h1: "Browser support", contract: "Popover API"},
	{path: "/docs/content-style", label: "Content style", h1: "Content style", contract: "Content structure"},
	{path: "/docs/acknowledgments", label: "Acknowledgments", h1: "Acknowledgments", contract: "inspired by, not a copy"},
	{path: "/docs/contributing", label: "Contributing", h1: "Contributing", contract: "Conventional commits"},
	{path: "/docs/changelog", label: "Changelog", h1: "Changelog", contract: "Keep a Changelog"},
	{path: "/docs/roadmap", label: "Roadmap", h1: "Roadmap", contract: "phases A–J"},
}

// TestHandbookPagesRender proves every Handbook destination returns 200 under
// the docs shell, renders its own content (h1 + a page-specific contract
// marker), and is reachable from a Core/System/Meta sidebar group.
func TestHandbookPagesRender(t *testing.T) {
	for _, hb := range handbookRoutes {
		t.Run(hb.path, func(t *testing.T) {
			res := httptest.NewRecorder()
			New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, hb.path, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
			}
			body := res.Body.String()
			if !strings.Contains(body, ">"+hb.h1+"</h1>") {
				t.Errorf("page must render <h1>%s</h1>", hb.h1)
			}
			if !strings.Contains(body, hb.contract) {
				t.Errorf("page is missing its content contract %q", hb.contract)
			}
			hasTier := strings.Contains(body, `class="docs-nav-group-label">Core<`) ||
				strings.Contains(body, `class="docs-nav-group-label">System<`) ||
				strings.Contains(body, `class="docs-nav-group-label">Meta<`)
			if !hasTier {
				t.Error("sidebar must include Core/System/Meta handbook tiers")
			}
			if !strings.Contains(body, `href="`+hb.path+`"`) {
				t.Errorf("sidebar must link to %s", hb.path)
			}
		})
	}
}

// TestDocsNavForHandbookGroup proves Core/System/Meta carry every handbook
// route with correct labels and current marking on /docs/tokens (System).
func TestDocsNavForHandbookGroup(t *testing.T) {
	nav := docsNavFor("/docs/tokens", "", "")
	flat := make([]docsNavLink, 0)
	var system *docsNavGroup
	for i := range nav.Groups {
		switch nav.Groups[i].Title {
		case "Core", "System", "Meta":
			flat = append(flat, nav.Groups[i].Links...)
		}
		if nav.Groups[i].Title == "System" {
			system = &nav.Groups[i]
		}
	}
	if system == nil {
		t.Fatal("docsNavFor must include a System group")
	}
	if len(flat) != len(handbookRoutes)+4 {
		t.Fatalf("Core+System+Meta links = %d, want %d", len(flat), len(handbookRoutes)+4)
	}
	for _, hb := range handbookRoutes {
		found := false
		for _, link := range flat {
			if link.Path != hb.path {
				continue
			}
			found = true
			if link.Label != hb.label {
				t.Errorf("handbook link %s label = %q, want %q", hb.path, link.Label, hb.label)
			}
			if link.Path == "/docs/tokens" && !link.Current {
				t.Error("Tokens must be current on /docs/tokens")
			}
			if link.Path != "/docs/tokens" && link.Current {
				t.Errorf("peer %s must not be current on /docs/tokens", link.Path)
			}
		}
		if !found {
			t.Errorf("handbook tiers missing %s", hb.path)
		}
	}
}

// TestHandbookGroupPrecedesComponentSections proves concept tiers come right
// after Getting started and BEFORE component reference sections.
func TestHandbookGroupPrecedesComponentSections(t *testing.T) {
	t.Run("nav model index", func(t *testing.T) {
		nav := docsNavFor("", "", "")
		if len(nav.Groups) < 5 {
			t.Fatalf("docsNavFor groups = %d, want at least 5", len(nav.Groups))
		}
		if nav.Groups[0].Title != "Getting started" {
			t.Fatalf("groups[0] = %q, want Getting started", nav.Groups[0].Title)
		}
		if nav.Groups[1].Title != "Core" || nav.Groups[2].Title != "System" || nav.Groups[3].Title != "Meta" {
			t.Fatalf("groups[1..3] = %q/%q/%q, want Core/System/Meta", nav.Groups[1].Title, nav.Groups[2].Title, nav.Groups[3].Title)
		}
		if nav.Groups[4].Title != "Foundation" {
			t.Fatalf("groups[4] = %q, want Foundation (first component section)", nav.Groups[4].Title)
		}
	})

	t.Run("sidebar render order", func(t *testing.T) {
		body := getOKBody(t, "/components/button")
		gettingStarted := strings.Index(body, `class="docs-nav-group-label">Getting started<`)
		core := strings.Index(body, `class="docs-nav-group-label">Core<`)
		system := strings.Index(body, `class="docs-nav-group-label">System<`)
		meta := strings.Index(body, `class="docs-nav-group-label">Meta<`)
		foundation := strings.Index(body, `class="docs-nav-group-label">Foundation<`)
		if gettingStarted < 0 || core < 0 || system < 0 || meta < 0 || foundation < 0 {
			t.Fatalf("sidebar must render Getting started, Core, System, Meta, Foundation (idxs %d,%d,%d,%d,%d)", gettingStarted, core, system, meta, foundation)
		}
		if !(gettingStarted < core && core < system && system < meta && meta < foundation) {
			t.Errorf("sidebar order must be Getting started < Core < System < Meta < Foundation")
		}
	})

	t.Run("hub render order", func(t *testing.T) {
		body := getOKBody(t, "/docs")
		start := strings.Index(body, ">Start here</h2>")
		try := strings.Index(body, ">Try a full screen</h2>")
		if start < 0 || try < 0 {
			t.Fatalf("docs hub must render Start here and Try a full screen (idxs %d, %d)", start, try)
		}
		if start > try {
			t.Errorf("hub must lead with Start here before recipes/demos, got Start %d > Try %d", start, try)
		}
		if strings.Contains(body, ">Foundation</h2>") || strings.Contains(body, ">Actions</h2>") {
			t.Error("docs hub body must not dump component category headings")
		}
	})
}

// TestHandbookPagesInSitemap proves every Handbook page is an indexable route
// in the generated sitemap (SEO contract: one <url> per public docs page).
func TestHandbookPagesInSitemap(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, hb := range handbookRoutes {
		want := "<loc>https://gelium-ui.example" + hb.path + "</loc>"
		if !strings.Contains(body, want) {
			t.Errorf("sitemap is missing %q", want)
		}
	}
}

// TestPageSectionArchitectureDocsModel pins the new Core entry to the docs
// model that drives the sidebar, search, footer, and previous/next pagination.
func TestPageSectionArchitectureDocsModel(t *testing.T) {
	const path = "/docs/page-section-architecture"
	nav := docsNavFor(path, "", "")
	var core []docsNavLink
	for _, group := range nav.Groups {
		if group.Title == "Core" {
			core = group.Links
			break
		}
	}
	if len(core) == 0 {
		t.Fatal("docs navigation must include the Core group")
	}
	found := false
	for i, link := range core {
		if link.Path != path {
			continue
		}
		found = true
		if !link.Current {
			t.Error("page and section architecture must be current on its route")
		}
		if i == 0 || i == len(core)-1 || core[i-1].Path != "/docs/screens" || core[i+1].Path != "/docs/section-references" {
			t.Fatalf("Core order around %s = %v, want Screens → Page + section architecture → Section references", path, core)
		}
		break
	}
	if !found {
		t.Fatalf("Core group must include %s", path)
	}
	if !strings.Contains(string(nav.SearchIndex), `"title":"Page + section architecture","href":"/docs/page-section-architecture","group":"Core"`) {
		t.Error("search index must include the Core handbook entry")
	}
	pn := prevNextFor(path, "", "")
	if pn == nil || pn.Prev == nil || pn.Next == nil || pn.Prev.Href != "/docs/screens" || pn.Next.Href != "/docs/section-references" {
		t.Fatalf("previous/next = %+v, want Screens and Section references", pn)
	}
	body := getOKBody(t, path)
	footer := strings.Index(body, `<footer class="ui-footer">`)
	if footer < 0 || !strings.Contains(body[footer:], `<a href="/docs">Docs</a>`) {
		t.Error("footer must include the shared docs entry point")
	}
}

// TestDocsIndexListsHandbook proves the /docs hub points readers into the
// Handbook without dumping every handbook destination (those live in the
// sidebar). A curated Start here set is enough.
func TestDocsIndexListsHandbook(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	if !strings.Contains(body, ">Start here</h2>") {
		t.Error("docs hub must include a Start here section")
	}
	if !strings.Contains(body, "sidebar") {
		t.Error("docs hub must tell readers the full catalog is in the sidebar")
	}
	if !strings.Contains(body, "Core") || !strings.Contains(body, "System") || !strings.Contains(body, "Meta") {
		t.Error("docs hub must name Core / System / Meta sidebar tiers")
	}
	// Curated high-value handbook entry points (not the full handbookRoutes list).
	for _, path := range []string{"/docs/screens", "/docs/page-section-architecture", "/docs/journeys", "/docs/data-display", "/docs/feedback", "/docs/compare", "/docs/forms", "/docs/themes", "/docs/tokens", "/docs/performance", "/llms.txt", "/llms-ux.txt"} {
		if !strings.Contains(body, `href="`+path+`"`) {
			t.Errorf("docs hub Start here must link to %s", path)
		}
	}
	// No full-catalog H2 dump (sidebar owns tiers + component groups).
	if strings.Contains(body, ">Handbook</h2>") || strings.Contains(body, ">Foundation</h2>") {
		t.Error("docs hub must not dump Handbook/Foundation catalog headings (sidebar owns that)")
	}
}
