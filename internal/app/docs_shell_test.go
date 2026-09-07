package app

import (
	"html"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"geliumui/lib"
)

// TestDocsShellFrameOnDocsAndComponents proves the two-pane docs chrome
// (topbar, dual sidebar/nav, main) on shell routes while home stays on the
// legacy site-header layout (tasks 2.1).
func TestDocsShellFrameOnDocsAndComponents(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantShell  bool
		wantStatus int
	}{
		{name: "docs hub", path: "/docs", wantShell: true, wantStatus: http.StatusOK},
		{name: "component button", path: "/components/button", wantShell: true, wantStatus: http.StatusOK},
		{name: "home", path: "/", wantShell: false, wantStatus: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", res.Code, tt.wantStatus)
			}
			body := res.Body.String()
			if tt.wantShell {
				for _, contract := range []string{
					`class="docs-chrome"`,
					`class="docs-nav-mobile"`,
					`<summary`,
					`class="docs-sidebar-desktop"`,
					`aria-label="Docs"`,
					`<main id="main-content"`,
					`href="#main-content"`,
					`ui-list`,
					`ui-list-item-link`,
					`ui-divider`,
					`ui-text-field`,
					`ui-button ui-button-text`,
					">" + lib.AssetsVersion + "<",
					`Gelium UI`,
					`type="search"`,
					`aria-label="Component recipe"`,
					`aria-label="Appearance"`,
				} {
					if !strings.Contains(body, contract) {
						t.Errorf("%s missing shell contract %q", tt.path, contract)
					}
				}
				// Shell must not keep the legacy primary header nav chrome.
				if strings.Contains(body, `class="site-header"`) {
					t.Errorf("%s must not render legacy site-header when docs shell is on", tt.path)
				}
			} else {
				for _, forbidden := range []string{
					`class="docs-chrome"`,
					`class="docs-nav-mobile"`,
					`class="docs-sidebar-desktop"`,
				} {
					if strings.Contains(body, forbidden) {
						t.Errorf("home must not render docs shell chrome %q", forbidden)
					}
				}
				if !strings.Contains(body, `class="site-header"`) {
					t.Error("home must keep legacy site-header")
				}
			}
		})
	}
}

// TestDocsStubRoutesAndShellChrome proves Patterns/Handbook routes return 200
// with shell chrome, honest disabled search, landmarks, and dual mobile
// markup (2.2).
func TestDocsStubRoutesAndShellChrome(t *testing.T) {
	for _, path := range []string{"/docs/patterns", "/docs/themes", "/docs/compare"} {
		t.Run(path, func(t *testing.T) {
			res := httptest.NewRecorder()
			New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
			}
			body := res.Body.String()
			for _, contract := range []string{
				`class="docs-chrome"`,
				`class="docs-nav-mobile"`,
				`class="docs-sidebar-desktop"`,
				`<nav aria-label="Docs"`,
				`<main id="main-content"`,
				`href="#main-content"`,
				`type="search"`,
				`name="q"`,
			} {
				if !strings.Contains(body, contract) {
					t.Errorf("%s missing contract %q", path, contract)
				}
			}
			// Search is a live GET form to the /docs hub: 0-JS Enter submits
			// /docs?q=<term>; with JS, search.js filters the nav index client-side.
			if !strings.Contains(body, `<form class="docs-search" method="get" action="/docs" role="search">`) {
				t.Errorf("%s must wrap search in a live GET form to /docs", path)
			}
		})
	}
}

// TestDocsShellActiveAndIAGroups proves sidebar IA groups and aria-current on
// the active destination for shell pages (R2/R3 smoke for PR2).
func TestDocsShellActiveAndIAGroups(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/button", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, group := range []string{
		"Getting started",
		"Foundation",
		"Actions",
		"Patterns",
		"Recipes",
		"Core",
		"System",
		"Meta",
	} {
		if !strings.Contains(body, group) {
			t.Errorf("sidebar missing IA group %q", group)
		}
	}
	// Button current: at least one aria-current="page" on the Button link.
	if !strings.Contains(body, `aria-current="page"`) {
		t.Error("active route must set aria-current=page on the current nav item")
	}
	if !strings.Contains(body, `href="/components/button"`) {
		t.Error("sidebar must link to /components/button")
	}
	// Recipes stay outbound real paths.
	for _, path := range []string{
		"/recipes/admin-resource",
		"/recipes/ops-queue",
		"/recipes/public-feed",
	} {
		if !strings.Contains(body, `href="`+path+`"`) {
			t.Errorf("sidebar missing recipe href %q", path)
		}
	}
}

// TestDocsShellColorSchemeSwitcher proves Light/Dark chrome on the docs topbar
// (right side), document-root scheme application, and query preservation.
func TestDocsShellColorSchemeSwitcher(t *testing.T) {
	t.Run("dark scheme class and data-theme", func(t *testing.T) {
		body := getOKBody(t, "/components/button?scheme=dark")
		if !(strings.Contains(body, `theme-material`) && strings.Contains(body, `theme-dark`)) {
			t.Fatalf("dark scheme must set theme-material + theme-dark on document root; snippet=%q", htmlClassSnippet(body))
		}
		if !strings.Contains(body, `data-theme="dark"`) {
			t.Error(`expected data-theme="dark"`)
		}
		if !strings.Contains(body, `aria-label="Appearance"`) {
			t.Error("docs topbar must expose Appearance (scheme) switcher")
		}
		// Native switch: checkbox with role=switch, checked → ?scheme=dark.
		if !strings.Contains(body, `<input type="checkbox" role="switch" name="scheme" value="dark" checked`) {
			t.Error("scheme switch must render checked under ?scheme=dark")
		}
		// Unchecked state maps to ?scheme=light via the hidden twin after the
		// checkbox (first value wins server-side: dark when checked, light when not).
		if !strings.Contains(body, `<input type="hidden" name="scheme" value="light">`) {
			t.Error("scheme switch must carry the hidden light twin for the unchecked state")
		}
		if !strings.Contains(body, `name="scheme"`) {
			t.Error("scheme switch must submit name=scheme")
		}
	})

	t.Run("light forces data-theme light", func(t *testing.T) {
		body := getOKBody(t, "/docs?scheme=light")
		if !strings.Contains(body, `data-theme="light"`) {
			t.Error(`?scheme=light must set data-theme="light" so OS dark media is suppressed`)
		}
		// The document root must not carry theme-dark under ?scheme=light.
		// Scope to the <html> tag: the docs prose may legitimately mention the
		// class route while the root class itself stays light.
		if strings.Contains(htmlClassSnippet(body), "theme-dark") {
			t.Error("light scheme must not add theme-dark class to the document root")
		}
	})

	t.Run("theme and scheme preserved together", func(t *testing.T) {
		body := getOKBody(t, "/components/button?theme=basecoat&scheme=dark")
		if !strings.Contains(body, `theme-basecoat`) || !strings.Contains(body, `theme-dark`) {
			t.Fatalf("expected basecoat+dark root classes; got %q", htmlClassSnippet(body))
		}
		// Sidebar list links keep both query keys (order may vary).
		if !strings.Contains(body, `href="/components/icon-button?`) {
			t.Fatal("missing sidebar peer href")
		}
		// Find one themed peer and require both keys.
		idx := strings.Index(body, `href="/components/icon-button?`)
		end := strings.Index(body[idx+6:], `"`)
		href := body[idx : idx+6+end+1]
		if !strings.Contains(href, `theme=basecoat`) || !strings.Contains(href, `scheme=dark`) {
			t.Errorf("sidebar peer must preserve theme+scheme; got %s", href)
		}
		// Appearance Dark stays current while Theme Basecoat stays current.
		if !strings.Contains(body, `aria-label="Appearance"`) {
			t.Error("missing Appearance switcher")
		}
	})

	t.Run("home marketing chrome has appearance without docs shell", func(t *testing.T) {
		body := getOKBody(t, "/")
		// Marketing landing may expose Appearance on the site header, but must
		// not use the docs two-pane shell chrome.
		if strings.Contains(body, `class="docs-chrome"`) {
			t.Error("home must not render docs-topbar")
		}
		if !strings.Contains(body, `class="site-header"`) {
			t.Error("home must keep site-header")
		}
	})
}

func containsDocsNavHref(body, href string) bool {
	pattern := `(?s)<a\s+class="ui-list-item-link(?:\s+is-current)?"\s+href="` + regexp.QuoteMeta(href) + `"`
	return regexp.MustCompile(pattern).FindStringIndex(body) != nil
}

func countCurrentDocsNavLinks(body, href string) int {
	pattern := `(?s)<a\s+class="ui-list-item-link\s+is-current"\s+href="` + regexp.QuoteMeta(href) + `"\s+aria-current="page"`
	return len(regexp.MustCompile(pattern).FindAllStringIndex(body, -1))
}

func htmlClassSnippet(body string) string {
	i := strings.Index(body, "<html")
	if i < 0 {
		return ""
	}
	j := strings.Index(body[i:], ">")
	if j < 0 {
		return body[i:min(i+120, len(body))]
	}
	return body[i : i+j+1]
}

// TestDocsShellSidebarPreservesTheme proves in-shell sidebar navigation keeps
// the selected ?theme= direction (Basecoat must not reset to Material).
func TestDocsShellSidebarPreservesTheme(t *testing.T) {
	body := getOKBody(t, "/components/button?theme=basecoat")
	if !strings.Contains(body, `class="theme-basecoat"`) {
		t.Fatal(`expected class="theme-basecoat" on document root`)
	}
	// Dual nav (mobile + desktop) must carry ?theme=basecoat on IA destinations.
	// Footer may still use bare paths — only the list-item sidebar links are
	// required to preserve theme (that is what drops direction on click).
	for _, path := range []string{
		"/components/button",
		"/components/icon-button",
		"/components/fab",
		"/docs",
		"/docs/patterns",
		"/docs/information-architecture",
		"/docs/themes",
		"/docs/tokens",
		"/docs/server-contracts",
		"/docs/accessibility",
		"/docs/principles",
		"/docs/content-style",
		"/recipes/admin-resource",
	} {
		if !containsDocsNavHref(body, path+"?theme=basecoat") {
			t.Errorf("sidebar missing theme-preserving list href for %q", path)
		}
	}
	// Active Button marker uses the themed href in both nav trees.
	if got := countCurrentDocsNavLinks(body, "/components/button?theme=basecoat"); got != 2 {
		t.Fatalf("themed Button aria-current markers = %d, want 2", got)
	}
	// No bare list-item href for Button (would drop theme on the current item).
	if countCurrentDocsNavLinks(body, "/components/button") > 0 {
		t.Error("current Button list link must not be a bare path under ?theme=basecoat")
	}
	// Breadcrumb non-current shell crumbs also keep theme.
	if !strings.Contains(body, `href="/docs?theme=basecoat"`) {
		t.Error("breadcrumb shell crumb should preserve ?theme=basecoat on /docs")
	}
}

// TestDocsTopbarLinksCarryChromeQuery proves the brand, blog and changelog
// links keep the selected theme/scheme. With hx-boost, a plain href would
// navigate without ?theme=&scheme= and the server would silently render the
// default theme (material/light). The GitHub link stays external.
func TestDocsTopbarLinksCarryChromeQuery(t *testing.T) {
	body := getOKBody(t, "/docs?theme=basecoat&scheme=dark")
	for _, contract := range []string{
		`href="/?scheme=dark&amp;theme=basecoat"`,
		`href="/blog?scheme=dark&amp;theme=basecoat"`,
		`href="/docs/changelog?scheme=dark&amp;theme=basecoat"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("topbar link missing chrome query %q", contract)
		}
	}
	plain := getOKBody(t, "/docs")
	if !strings.Contains(plain, `href="/blog"`) {
		t.Error("topbar blog link must stay plain when no theme/scheme is selected")
	}
	if !strings.Contains(plain, `href="https://github.com/cryptonahue/gelium-ui"`) {
		t.Error("topbar GitHub link must remain external and unboosted")
	}
}

// TestDocsShellChromeActivePeersAndIA (task 3.1) proves exact active-route peers,
// full docsSections IA labels, theme switcher ?theme=-only links, and root class
// for ?theme=basecoat on shell pages.
func TestDocsShellChromeActivePeersAndIA(t *testing.T) {
	t.Run("button peers not current dual nav", func(t *testing.T) {
		body := getOKBody(t, "/components/button")
		// Dual mobile+desktop nav: Button is current in both trees.
		if got := countCurrentDocsNavLinks(body, "/components/button"); got != 2 {
			t.Fatalf("Button aria-current nav markers = %d, want 2 (mobile+desktop)", got)
		}
		// Actions peers must never carry aria-current="page".
		for _, peer := range []string{
			"/components/icon-button",
			"/components/fab",
			"/components/chips",
			"/components/segmented-button",
			"/components/menu",
		} {
			if countCurrentDocsNavLinks(body, peer) > 0 {
				t.Errorf("peer %q must not be aria-current=page", peer)
			}
			// Peer links exist but use the non-current list-item class.
			if !containsDocsNavHref(body, peer) {
				t.Errorf("sidebar missing non-current peer link %q", peer)
			}
		}
		// Docs hub must not be current on a component page.
		if strings.Contains(body, `href="/docs" aria-current="page"`) {
			t.Error("/docs must not be current on /components/button")
		}
	})

	t.Run("docs hub current", func(t *testing.T) {
		body := getOKBody(t, "/docs")
		if got := countCurrentDocsNavLinks(body, "/docs"); got != 2 {
			t.Fatalf("Documentation hub aria-current markers = %d, want 2", got)
		}
		if strings.Contains(body, `href="/components/button" aria-current="page"`) {
			t.Error("component links must not be current on /docs")
		}
	})

	t.Run("all docsSections group titles in chrome", func(t *testing.T) {
		body := getOKBody(t, "/docs")
		// Top-level IA blocks + every docsSections category title.
		want := []string{
			"Getting started",
			"Core",
			"System",
			"Meta",
			"Patterns",
			"Recipes",
		}
		for _, section := range docsSections {
			want = append(want, section.Title)
		}
		for _, group := range want {
			// Group labels use the docs-nav-group-label marker (not prose h2 only).
			// Titles may contain "&" which html/template escapes to &amp;.
			marker := `class="docs-nav-group-label">` + html.EscapeString(group) + `<`
			if !strings.Contains(body, marker) {
				t.Errorf("sidebar missing group label marker for %q", group)
			}
		}
	})

	t.Run("recipe select has only closed values", func(t *testing.T) {
		// The Recipe form deliberately drops unrelated query state and submits a
		// closed visual key; middleware performs the canonical GET redirect.
		body := getOKBody(t, "/components/button?foo=bar&theme=material")
		if !strings.Contains(body, `<form class="ui-recipe-switcher ui-recipe-switcher--compact" method="get"`) {
			t.Error("topbar Recipe switcher must be a native GET form")
		}
		for _, opt := range []string{`<option value="native"`, `<option value="material"`, `<option value="basecoat"`, `<option value="baseui"`} {
			if !strings.Contains(body, opt) {
				t.Errorf("Recipe selector missing option %q", opt)
			}
		}
		if !strings.Contains(body, `name="visual"`) || !strings.Contains(body, `value="material" selected`) {
			t.Error("legacy material URL must map to the matching safe visual recipe")
		}
		if strings.Contains(body, `name="foo"`) || strings.Contains(body, "?foo=bar") {
			t.Error("Recipe form must not preserve unrelated query parameters")
		}
		if strings.Contains(body, "ui-theme-switcher-option") {
			t.Error("retired theme link-list options must not render")
		}
	})

	t.Run("basecoat theme class on root without client script", func(t *testing.T) {
		body := getOKBody(t, "/components/button?theme=basecoat")
		if !strings.Contains(body, `class="theme-basecoat"`) {
			t.Error(`?theme=basecoat must set class="theme-basecoat" on the document root`)
		}
		// Shell still present; theme does not drop chrome.
		if !strings.Contains(body, `class="docs-chrome"`) {
			t.Error("theme query must keep docs shell chrome")
		}
	})
}

// TestDocsShellPathsStableNoRedirect (task 3.3) proves shell paths stay 200
// without a redirect matrix when the chrome ships.
func TestDocsShellPathsStableNoRedirect(t *testing.T) {
	for _, path := range []string{
		"/docs",
		"/docs/patterns",
		"/docs/information-architecture",
		"/docs/themes",
		"/docs/tokens",
		"/docs/server-contracts",
		"/docs/accessibility",
		"/docs/principles",
		"/docs/content-style",
		"/components/button",
		"/components/data-table",
	} {
		t.Run(path, func(t *testing.T) {
			res := httptest.NewRecorder()
			New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d (no redirect)", res.Code, http.StatusOK)
			}
			if loc := res.Header().Get("Location"); loc != "" {
				t.Errorf("unexpected redirect Location=%q", loc)
			}
		})
	}
}

// TestDocsShellFooterAndJSONLDRegressions (task 3.3) keeps footer IA and
// component structured data stable under the docs shell.
func TestDocsShellFooterAndJSONLDRegressions(t *testing.T) {
	body := getOKBody(t, "/components/button")
	for _, contract := range []string{
		`<footer class="ui-footer">`,
		`<p class="ui-footer-brand">Gelium UI</p>`,
		`class="ui-footer-heading">Product</summary>`,
		`class="ui-footer-heading">Guidance</summary>`,
		`class="ui-footer-heading">Source</summary>`,
		`<a href="/docs">Docs</a>`,
		`<a href="/docs/agent-workflow">Agent workflow</a>`,
		`<a href="/docs/server-contracts">Server contracts</a>`,
		`<a href="https://github.com/cryptonahue/gelium-ui">GitHub</a>`,
		`"@type":"BreadcrumbList"`,
		`"@type":"TechArticle"`,
		`"item":"https://gelium-ui.example/components/button"`,
		`"url":"https://gelium-ui.example/components/button"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("shell page missing footer/JSON-LD contract %q", contract)
		}
	}
}

func getOKBody(t *testing.T, path string) string {
	t.Helper()
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
	if res.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want %d", path, res.Code, http.StatusOK)
	}
	return res.Body.String()
}

// TestHomeUnchangedByDocsShell is the negative home contract for PR2 focused runs.
func TestHomeUnchangedByDocsShell(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, forbidden := range []string{
		`class="docs-chrome"`,
		`class="docs-nav-mobile"`,
		`class="docs-sidebar-desktop"`,
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("home must not contain %q", forbidden)
		}
	}
	if !strings.Contains(body, `class="site-header"`) {
		t.Error("home must keep site-header")
	}
}

func TestUsesDocsShell(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "docs hub", path: "/docs", want: true},
		{name: "docs patterns stub", path: "/docs/patterns", want: true},
		{name: "docs themes stub", path: "/docs/themes", want: true},
		{name: "docs nested path", path: "/docs/anything", want: true},
		{name: "component button", path: "/components/button", want: true},
		{name: "component data table", path: "/components/data-table", want: true},
		{name: "home", path: "/", want: false},
		{name: "recipe admin", path: "/recipes/admin-resource", want: false},
		{name: "recipe ops", path: "/recipes/ops-queue", want: false},
		{name: "demo", path: "/demo/whatsapp", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := usesDocsShell(tt.path); got != tt.want {
				t.Fatalf("usesDocsShell(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestDocsNavFor(t *testing.T) {
	t.Run("topbar slots and IA blocks", func(t *testing.T) {
		nav := docsNavFor("/docs", "", "")
		if nav.Version != lib.AssetsVersion {
			t.Errorf("Version = %q, want %q (derived from the single cache-bust constant)", nav.Version, lib.AssetsVersion)
		}
		// The nav model drives client-side search: a JSON index of every
		// destination is emitted so search.js can filter without a server.
		if nav.SearchIndex == "" || !strings.Contains(string(nav.SearchIndex), `"title":"Documentation"`) {
			t.Error("SearchIndex must be a non-empty JSON index including the Documentation hub")
		}
		if nav.ThemeSlug != "" || nav.Scheme != "" {
			t.Errorf("ThemeSlug/Scheme = %q/%q, want empty without chrome query", nav.ThemeSlug, nav.Scheme)
		}

		titles := docsNavGroupTitles(nav)
		for _, want := range []string{
			"Getting started",
			"Core",
			"System",
			"Meta",
			"Patterns",
			"Recipes",
		} {
			if !containsString(titles, want) {
				t.Errorf("docsNavFor missing IA block %q; got %v", want, titles)
			}
		}
		for _, section := range docsSections {
			if !containsString(titles, section.Title) {
				t.Errorf("docsNavFor missing docsSections group %q; got %v", section.Title, titles)
			}
		}
		// Five top-level IA blocks: Getting started, Components (via docsSections),
		// Patterns, Recipes, Handbook — Components expands to one group per section.
		if len(nav.Groups) < 5 {
			t.Fatalf("docsNavFor groups = %d, want at least 5 IA blocks", len(nav.Groups))
		}
	})

	t.Run("current on docs hub", func(t *testing.T) {
		nav := docsNavFor("/docs", "", "")
		current := currentDocsNavLinks(nav)
		if len(current) != 1 {
			t.Fatalf("current links = %v, want exactly one", current)
		}
		if current[0].Path != "/docs" || current[0].Label != "Documentation" {
			t.Errorf("current = %+v, want Path=/docs Label=Documentation", current[0])
		}
	})

	t.Run("current on component button", func(t *testing.T) {
		nav := docsNavFor("/components/button", "", "")
		current := currentDocsNavLinks(nav)
		if len(current) != 1 {
			t.Fatalf("current links = %v, want exactly one", current)
		}
		if current[0].Path != "/components/button" || current[0].Label != "Button" {
			t.Errorf("current = %+v, want Path=/components/button Label=Button", current[0])
		}
		// Peers in the same Actions group must not be current.
		for _, g := range nav.Groups {
			if g.Title != "Actions" {
				continue
			}
			for _, link := range g.Links {
				if link.Path == "/components/button" {
					continue
				}
				if link.Current {
					t.Errorf("peer %q must not be current", link.Path)
				}
			}
		}
	})

	t.Run("recipes are real outbound paths", func(t *testing.T) {
		nav := docsNavFor("/docs", "", "")
		var recipes *docsNavGroup
		for i := range nav.Groups {
			if nav.Groups[i].Title == "Recipes" {
				recipes = &nav.Groups[i]
				break
			}
		}
		if recipes == nil {
			t.Fatal("missing Recipes group")
		}
		wantPaths := map[string]string{
			"/recipes/admin-resource": "Admin Resource",
			"/recipes/ops-queue":      "Ops Queue",
			"/recipes/public-feed":    "Public Feed",
			"/recipes/rich-article":   "Rich Article",
		}
		if len(recipes.Links) != len(wantPaths) {
			t.Fatalf("Recipes links = %d, want %d", len(recipes.Links), len(wantPaths))
		}
		for _, link := range recipes.Links {
			label, ok := wantPaths[link.Path]
			if !ok {
				t.Errorf("unexpected recipe path %q", link.Path)
				continue
			}
			if link.Label != label {
				t.Errorf("recipe %q label = %q, want %q", link.Path, link.Label, label)
			}
		}
	})

	t.Run("patterns and themes stub paths", func(t *testing.T) {
		nav := docsNavFor("/docs/patterns", "", "")
		if !hasDocsNavLink(nav, "/docs/patterns", true) {
			t.Error("Patterns group must mark /docs/patterns current")
		}
		if !hasDocsNavLink(nav, "/docs/themes", false) {
			t.Error("Themes group must include /docs/themes")
		}
	})
}

func TestDefaultFooter(t *testing.T) {
	footer := defaultFooter()
	if footer == nil {
		t.Fatal("defaultFooter() returned nil")
	}
	if footer.Brand != "Gelium UI" {
		t.Errorf("Brand = %q, want %q", footer.Brand, "Gelium UI")
	}
	if footer.Legal != "© 2026 Gelium UI · MIT" {
		t.Errorf("Legal = %q, want %q", footer.Legal, "© 2026 Gelium UI · MIT")
	}
	if got := len(footer.Sections); got != 3 {
		t.Fatalf("default footer groups = %d, want 3", got)
	}
	for i, want := range []string{"Product", "Guidance", "Source"} {
		if footer.Sections[i].Title != want {
			t.Errorf("footer group %d = %q, want %q", i, footer.Sections[i].Title, want)
		}
	}
	for _, path := range []string{
		"/docs",
		"/components/button",
		"/docs/themes/gallery",
		"/docs/agent-workflow",
		"/docs/server-contracts",
		"https://www.npmjs.com/package/gelium-ui",
		"https://github.com/cryptonahue/gelium-ui",
	} {
		if !containsString(footerFlatPaths(footer), path) {
			t.Errorf("shared footer missing path %q", path)
		}
	}
}

func TestNavLinksDerivedFromDocsSections(t *testing.T) {
	links := navLinks()
	if len(links) == 0 || links[0].Path != "/docs" || links[0].Label != "Docs" {
		t.Fatalf("navLinks()[0] = %+v, want Docs → /docs", firstNavLink(links))
	}
	// Component entries must match docsSections order and labels — not componentRoutes.
	want := make([]navLink, 0)
	for _, section := range docsSections {
		want = append(want, section.Links...)
	}
	got := links[1:]
	if len(got) != len(want) {
		t.Fatalf("component nav links = %d, want %d from docsSections", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("navLinks()[%d] = %+v, want %+v (docsSections order)", i+1, got[i], want[i])
		}
	}
}

// TestComponentPageDemoFirst proves the demo-first layout contract: a
// component page renders title → live demo → body. The live demo preview
// must sit between the H1 and the first prose paragraph (Base UI/Naive UI
// order: show the component, then the rules).
func TestComponentPageDemoFirst(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/button", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()

	h1 := strings.Index(body, ">Button</h1>")
	if h1 < 0 {
		t.Fatal("button page must render <h1>Button</h1>")
	}
	demo := strings.Index(body, `aria-label="Button examples"`)
	if demo < 0 {
		t.Fatal("button page must render the live demo preview")
	}
	// The first prose paragraph after the H1 is the lead sentence of the
	// markdown body ("Use buttons when..." in button.md).
	prose := strings.Index(body, "<p>")
	if h1 > demo {
		t.Errorf("demo must come after the H1 (h1 idx %d, demo idx %d)", h1, demo)
	}
	if demo > prose {
		t.Errorf("demo must come before the body prose (demo idx %d, prose idx %d)", demo, prose)
	}
	// A handbook page (no live demo) must not regress: H1 then prose directly.
	res2 := httptest.NewRecorder()
	New().ServeHTTP(res2, httptest.NewRequest(http.MethodGet, "/docs/themes", nil))
	if res2.Code != http.StatusOK {
		t.Fatalf("handbook status = %d, want %d", res2.Code, http.StatusOK)
	}
	body2 := res2.Body.String()
	h1Themes := strings.Index(body2, ">Themes</h1>")
	proseThemes := strings.Index(body2, "<p>")
	if h1Themes < 0 || proseThemes < 0 || h1Themes > proseThemes {
		t.Errorf("handbook page must keep H1 before prose (h1 idx %d, prose idx %d)", h1Themes, proseThemes)
	}
}

func docsNavGroupTitles(nav docsNavView) []string {
	titles := make([]string, 0, len(nav.Groups))
	for _, g := range nav.Groups {
		titles = append(titles, g.Title)
	}
	return titles
}

func currentDocsNavLinks(nav docsNavView) []docsNavLink {
	var out []docsNavLink
	for _, g := range nav.Groups {
		for _, link := range g.Links {
			if link.Current {
				out = append(out, link)
			}
		}
	}
	return out
}

func hasDocsNavLink(nav docsNavView, path string, current bool) bool {
	for _, g := range nav.Groups {
		for _, link := range g.Links {
			if link.Path == path && link.Current == current {
				return true
			}
		}
	}
	return false
}

func footerFlatPaths(f *footerView) []string {
	var paths []string
	for _, section := range f.Sections {
		for _, link := range section.Links {
			paths = append(paths, link.Path)
		}
	}
	return paths
}

func containsString(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

func firstNavLink(links []navLink) navLink {
	if len(links) == 0 {
		return navLink{}
	}
	return links[0]
}
