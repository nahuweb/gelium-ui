package app

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/yuin/goldmark"

	webassets "geliumui/site/web"
)

func openingTagWithID(t *testing.T, body, element, id string) string {
	t.Helper()
	pattern := `<` + element + `\b[^>]*\bid="` + regexp.QuoteMeta(id) + `"[^>]*>`
	tag := regexp.MustCompile(pattern).FindString(body)
	if tag == "" {
		t.Fatalf("body is missing <%s> with id %q", element, id)
	}
	return tag
}

func renderButton(t *testing.T, view buttonView) string {
	t.Helper()
	tmpl := parseTestTemplates(t, "templates/button.html")
	var rendered bytes.Buffer
	if err := tmpl.ExecuteTemplate(&rendered, "button", view); err != nil {
		t.Fatalf("execute button template: %v", err)
	}
	return rendered.String()
}

func renderTextField(t *testing.T, view textFieldView) string {
	t.Helper()
	tmpl := parseTestTemplates(t, "templates/text-field.html")
	var rendered bytes.Buffer
	if err := tmpl.ExecuteTemplate(&rendered, "text-field", view); err != nil {
		t.Fatalf("execute text field template: %v", err)
	}
	return rendered.String()
}

func TestHealthzReturnsPlainTextOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()

	New().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := res.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want plain UTF-8 text", got)
	}
	if got := res.Body.String(); got != "ok\n" {
		t.Errorf("body = %q, want %q", got, "ok\\n")
	}
}

func TestHomeRendersMarketingLanding(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()

	New().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`class="ui-landing"`,
		`class="ui-hero`,
		`Build server-rendered UI your LLM can understand.`,
		`class="ui-button ui-button-primary"`,
		`Install gelium-ui`,
		`href="#landing-install"`,
		`href="/docs/agent-workflow"`,
		`class="ui-feature-card`,
		`class="ui-split`,
		`class="ui-footer"`,
		`Product`,
		`Guidance`,
		`class="site-header"`,
		`class="site-header-mobile-menu"`,
		`<summary>Menu</summary>`,
		`aria-label="Appearance"`,
		`src="/static/htmx.min.js?v=0.6.6"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("home does not contain contract %q", contract)
		}
	}
	// Must not keep the old Markdown-only home article as the primary surface.
	if strings.Contains(body, `<article class="prose">`) {
		t.Error("marketing landing must not render prose article shell")
	}
	// Must not use the docs two-pane chrome on home.
	if strings.Contains(body, `class="docs-chrome"`) {
		t.Error("home must not render docs shell topbar")
	}
}

func TestLayoutCacheBustsEmbeddedAssetsAcrossExeUpgrades(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	body := res.Body.String()

	for _, asset := range []string{
		`href="/static/app.css?v=0.6.6"`,
		`src="/static/htmx.min.js?v=0.6.6"`,
		`src="/static/app.js?v=0.6.6"`,
	} {
		if !strings.Contains(body, asset) {
			t.Errorf("layout must cache-bust upgraded embedded asset %s", asset)
		}
	}
}

func TestStaticBuildArtifactsAreServedFromEmbeddedFilesystem(t *testing.T) {
	tests := []struct {
		path        string
		contentType string
		contract    string
	}{
		{path: "/static/app.css", contentType: "text/css; charset=utf-8", contract: ".ui-button"},
		{path: "/static/htmx.min.js", contentType: "text/javascript; charset=utf-8", contract: "htmx"},
		{path: "/static/gelium.js", contentType: "text/javascript; charset=utf-8", contract: "X-Gelium-Validation"},
		{path: "/static/app.js", contentType: "text/javascript; charset=utf-8", contract: "applyOptimisticChrome"},
		{path: "/static/search.js", contentType: "text/javascript; charset=utf-8", contract: "search"},
		{path: "/static/morph-afterswap.js", contentType: "text/javascript; charset=utf-8", contract: "_geliumPageGeneration"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			res := httptest.NewRecorder()
			New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
			}
			if got := res.Header().Get("Content-Type"); got != tt.contentType {
				t.Errorf("Content-Type = %q, want %q", got, tt.contentType)
			}
			if got := res.Header().Get("Cache-Control"); got != "no-cache" {
				t.Errorf("Cache-Control = %q, want revalidation with no-cache", got)
			}
			if !strings.Contains(res.Body.String(), tt.contract) {
				t.Errorf("asset does not contain build contract %q", tt.contract)
			}
		})
	}
}

func TestMaterialDarkThemeKeepsFilledFieldDistinctFromSurface(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/static/app.css", nil))
	css := res.Body.String()

	// Single dark mechanism: each dark value is declared exactly once in the
	// compiled bundle (the explicit .theme-material.theme-dark class route),
	// never duplicated by a dark media block.
	if got := strings.Count(css, "--ui-color-surface:#211f26"); got != 1 {
		t.Fatalf("compiled dark theme surface declarations = %d, want 1 (single class route)", got)
	}
	if got := strings.Count(css, "--ui-field-container:#36343b"); got != 1 {
		t.Errorf("compiled dark theme filled container declarations = %d, want 1 (single class route)", got)
	}
	if strings.Contains(css, "--ui-field-container:#211f26") {
		t.Error("dark filled field container must differ from the #211f26 surface")
	}
}

func TestMaterialThemeDefinesTextFieldTypescaleTokens(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/static/app.css", nil))
	css := res.Body.String()

	// Phase B decomposition (R1): the compiled bundle carries the decomposed
	// per-step tokens (size/weight/line-height) in the theme, and the core
	// alias composes them into the font: shorthand consumers use.
	for _, token := range []string{
		`--ui-type-body-lg-size:1rem`,
		`--ui-type-body-lg-weight:400`,
		`--ui-type-body-lg:var(--ui-type-body-lg-weight) var(--ui-type-body-lg-size)/var(--ui-type-body-lg-line-height) var(--ui-type-body-lg-family)`,
		`--ui-type-body-sm-size:.75rem`,
		`--ui-type-body-sm-weight:400`,
		`--ui-type-label-sm-size:.75rem`,
		`--ui-type-label-sm-weight:500`,
	} {
		if !strings.Contains(css, token) {
			t.Errorf("compiled Material theme is missing text-field typescale definition %q", token)
		}
	}
}

func TestMaterialThemeExposesSemanticFoundationContracts(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/static/app.css", nil))
	css := res.Body.String()

	for _, token := range []string{
		"--ui-color-primary:",
		"--ui-type-display-sm:",
		"--ui-radius-full:",
		"--ui-shadow-5:",
		"--ui-focus-thickness:3px",
		"--ui-focus-offset:2px",
		"--ui-field-container:",
		"--ui-field-border:",
		"--ui-field-error:",
		`.theme-material.dark`,
	} {
		if !strings.Contains(css, token) {
			t.Errorf("compiled Material theme is missing %q", token)
		}
	}
}
func TestThemeClassIsServerDrivenAndAllowlisted(t *testing.T) {
	tests := []struct {
		name  string
		theme string
		want  string
	}{
		{name: "empty falls back to default", theme: "", want: "theme-material"},
		{name: "default passes the allowlist", theme: "theme-material", want: "theme-material"},
		{name: "basecoat passes the allowlist", theme: "theme-basecoat", want: "theme-basecoat"},
		{name: "unknown theme falls back to default", theme: "theme-unknown", want: "theme-material"},
		{name: "arbitrary injection falls back to default", theme: "onload=alert(1)", want: "theme-material"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := themeClass(tt.theme); got != tt.want {
				t.Errorf("themeClass(%q) = %q, want %q", tt.theme, got, tt.want)
			}
		})
	}
}

func TestLayoutRendersThemeClassOnHTMLElement(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	body := res.Body.String()
	for _, contract := range []string{
		`<html lang="en" class="theme-material" data-gelium-reference="none" data-gelium-skin="none" data-gelium-contract="gelium" data-gelium-scheme="system">`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("home does not render the server-driven theme class %q", contract)
		}
	}
}

func renderToast(t *testing.T, view toastView) string {
	t.Helper()
	tmpl := parseTestTemplates(t, "templates/toast.html")
	var rendered bytes.Buffer
	if err := tmpl.ExecuteTemplate(&rendered, "toast", view); err != nil {
		t.Fatalf("execute toast template: %v", err)
	}
	return rendered.String()
}

func renderBanner(t *testing.T, view bannerView) string {
	t.Helper()
	tmpl := parseTestTemplates(t, "templates/banner.html")
	var rendered bytes.Buffer
	if err := tmpl.ExecuteTemplate(&rendered, "banner", view); err != nil {
		t.Fatalf("execute banner template: %v", err)
	}
	return rendered.String()
}

// renderInlineAlert drives the inline-alert partial directly with the
// production view model (Phase D pattern 3), pinning the tone→role
// derivation (error → alert, rest → status).
func renderInlineAlert(t *testing.T, view inlineAlertView) string {
	t.Helper()
	tmpl := parseTestTemplates(t, "templates/inline-alert.html")
	var rendered bytes.Buffer
	if err := tmpl.ExecuteTemplate(&rendered, "inline-alert", view); err != nil {
		t.Fatalf("execute inline-alert template: %v", err)
	}
	return rendered.String()
}

// TestLayoutOmitsBannerWhenNil proves the pageView.Banner slot is optional: a
// nil banner renders nothing and the layout still builds cleanly.
func TestLayoutOmitsBannerWhenNil(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if strings.Contains(res.Body.String(), "ui-banner") {
		t.Error("layout must not render a banner when pageView.Banner is nil")
	}
}

// TestLayoutRendersBannerSlotBetweenHeaderAndMain proves the global banner slot
// sits between the </header> and <main> landmarks and renders the partial when
// pageView.Banner is set.
func TestLayoutRendersBannerSlotBetweenHeaderAndMain(t *testing.T) {
	tmpl := parseTestTemplates(t, "templates/*.html")
	var page bytes.Buffer
	data := pageView{
		Title:      "banner slot",
		ThemeClass: "theme-material",
		Nav:        []navLink{{Path: "/", Label: "Home"}},
		Banner: &bannerView{
			Tone:  "info",
			Title: "Scheduled maintenance",
			Body:  "Tonight 22:00-23:00 UTC.",
		},
	}
	if err := tmpl.ExecuteTemplate(&page, "layout", data); err != nil {
		t.Fatalf("execute layout with banner: %v", err)
	}
	body := page.String()
	for _, contract := range []string{
		`class="ui-banner ui-banner--info"`,
		`role="status"`,
		`class="ui-banner-title">Scheduled maintenance`,
		`class="ui-banner-body">Tonight 22:00-23:00 UTC.`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("layout does not render banner contract %q", contract)
		}
	}

	headerEnd := strings.Index(body, "</header>")
	mainStart := strings.Index(body, "<main")
	bannerPos := strings.Index(body, `class="ui-banner ui-banner--info"`)
	if headerEnd < 0 || mainStart < 0 {
		t.Fatal("layout is missing the header or main landmarks")
	}
	if bannerPos < headerEnd || bannerPos > mainStart {
		t.Error("banner slot must render between </header> and <main>")
	}
}

// TestBannerRoleIsDerivedFromTone proves role="alert" is reserved for the
// error tone and every other tone announces politely as status.
func TestBannerRoleIsDerivedFromTone(t *testing.T) {
	for _, tone := range []string{"error", "success", "info", "warning"} {
		got := renderBanner(t, bannerView{Tone: tone, Body: "x"})
		want := "role=\"status\""
		if tone == "error" {
			want = "role=\"alert\""
		}
		if !strings.Contains(got, want) {
			t.Errorf("tone %q must render %s", tone, want)
		}
	}
}

// TestInlineAlertRoleIsDerivedFromTone proves the inline-alert partial mirrors
// the banner contract: role="alert" is reserved for the error tone and every
// other tone announces politely as status — success included. This is the
// render-level confirmation for the success persistent reuse (the tone CSS
// alone was already pinned by TestInlineAlertTonesUseCoreTokens).
func TestInlineAlertRoleIsDerivedFromTone(t *testing.T) {
	for _, tone := range []string{"error", "success", "info", "warning"} {
		got := renderInlineAlert(t, inlineAlertView{Tone: tone, Body: "x"})
		want := "role=\"status\""
		if tone == "error" {
			want = "role=\"alert\""
		}
		if !strings.Contains(got, want) {
			t.Errorf("tone %q must render %s", tone, want)
		}
	}
}

// TestBannerMarkupContracts proves the decorative icon is aria-hidden, the CTA
// is a real link, the dismiss is a POST form with a submit button, and an empty
// DismissHref omits the dismiss form entirely.
func TestBannerMarkupContracts(t *testing.T) {
	icon := template.HTML(`<svg aria-hidden="true"><title>info</title></svg>`)
	got := renderBanner(t, bannerView{
		Tone:        "error",
		Icon:        icon,
		Title:       "Session expired",
		Body:        "Re-authenticate to continue.",
		CTA:         true,
		CTAHref:     "/login",
		CTALabel:    "Re-authenticate",
		DismissHref: "/dismiss",
		DismissIcon: template.HTML("×"),
	})
	for _, contract := range []string{
		`<span class="ui-banner-icon" aria-hidden="true">`,
		`class="ui-banner-title">Session expired`,
		`class="ui-banner-body">Re-authenticate to continue.`,
		`<a class="ui-button" href="/login">Re-authenticate</a>`,
		`<form class="ui-banner-dismiss" method="post" action="/dismiss">`,
		`<button class="ui-icon-button" type="submit" aria-label="Dismiss">×</button>`,
	} {
		if !strings.Contains(got, contract) {
			t.Errorf("banner markup is missing contract %q", contract)
		}
	}

	noDismiss := renderBanner(t, bannerView{Tone: "success", Body: "Saved."})
	if strings.Contains(noDismiss, "ui-banner-dismiss") {
		t.Error("banner must omit the dismiss form when DismissHref is empty")
	}
}

func renderErrorState(t *testing.T, view errorStateView) string {
	t.Helper()
	tmpl := parseTestTemplates(t, "templates/error-state.html")
	var rendered bytes.Buffer
	if err := tmpl.ExecuteTemplate(&rendered, "error-state", view); err != nil {
		t.Fatalf("execute error-state template: %v", err)
	}
	return rendered.String()
}

// TestErrorStateMarkupContracts proves the status code is the decorative
// anchor (aria-hidden), the title is the single h1, the body is the muted
// message, and the retry is a real GET link rendered only when Retry is set.
func TestErrorStateMarkupContracts(t *testing.T) {
	got := renderErrorState(t, errorStateView{
		StatusCode: 404,
		Title:      "Page not found",
		Body:       "The page you are looking for does not exist or has moved.",
		Retry:      true,
		Href:       "/",
		Label:      "Back to home",
	})
	for _, contract := range []string{
		`class="ui-error-state" role="alert"`,
		`class="ui-error-state-code" aria-hidden="true">404`,
		`<h1 class="ui-error-state-title">Page not found</h1>`,
		`class="ui-error-state-body">The page you are looking for does not exist or has moved.`,
		`<a class="ui-button" href="/">Back to home</a>`,
	} {
		if !strings.Contains(got, contract) {
			t.Errorf("error-state markup is missing contract %q", contract)
		}
	}
	if count := strings.Count(got, "<h1"); count != 1 {
		t.Errorf("error-state must render exactly one h1, got %d", count)
	}

	noRetry := renderErrorState(t, errorStateView{
		StatusCode: 500,
		Title:      "Something went wrong",
		Body:       "This page could not be loaded. Please try again later.",
	})
	if strings.Contains(noRetry, "ui-button") {
		t.Error("error-state must omit the retry link when Retry is false")
	}
}

// TestUnknownRouteRendersErrorStatePage proves the catch-all serves the ERROR
// STATE slot with the real 404 status and the full layout, instead of the
// plain-text net/http default.
func TestUnknownRouteRendersErrorStatePage(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/does-not-exist", nil))
	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`class="ui-error-state" role="alert"`,
		`class="ui-error-state-code" aria-hidden="true">404`,
		`<h1 class="ui-error-state-title">Page not found</h1>`,
		`class="ui-error-state-body">The page you are looking for does not exist or has moved.`,
		`<a class="ui-button" href="/">Back to home</a>`,
		`<main`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("404 page is missing contract %q", contract)
		}
	}
	if strings.Contains(body, "404 page not found") {
		t.Error("404 must not fall back to the plain net/http body")
	}
}

// TestErrorSlotOmittedWhenNil proves the Error slot is optional: a nil
// pageView.Error renders the normal content and no error state at all.
func TestErrorSlotOmittedWhenNil(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if strings.Contains(res.Body.String(), "ui-error-state") {
		t.Error("layout must not render an error state when pageView.Error is nil")
	}
}

// TestHomeRendersServerDrivenMetadata proves the home page emits the full
// server-driven metadata contract: description, clean canonical, index robots,
// a minimal OG set and the WebSite JSON-LD (SEO contract §1, §3, §6, §12).
func TestHomeRendersServerDrivenMetadata(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<meta name="description" content="Gelium UI — themeable,`,
		`<link rel="canonical" href="https://gelium-ui.example/">`,
		`<meta name="robots" content="index, follow">`,
		`<meta property="og:type" content="website">`,
		`<meta property="og:url" content="https://gelium-ui.example/">`,
		`<script type="application/ld+json">{"@context":"https://schema.org","@type":"WebSite"`,
		`class="theme-material"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("home is missing metadata contract %q", contract)
		}
	}
	// OG title follows page title for the marketing landing.
	if !strings.Contains(body, `<meta property="og:title"`) {
		t.Error("home must emit og:title")
	}
}

// TestComponentPageRendersMetadata proves component pages resolve per-route
// metadata (article OG type, clean canonical, index robots) from the render
// choke point without any handler change.
func TestComponentPageRendersMetadata(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/button", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<link rel="canonical" href="https://gelium-ui.example/components/button">`,
		`<meta name="robots" content="index, follow">`,
		`<meta property="og:type" content="article">`,
		`<meta property="og:title" content="Button · Gelium UI">`,
		`<meta property="og:url" content="https://gelium-ui.example/components/button">`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("button docs is missing metadata contract %q", contract)
		}
	}
}

// TestCanonicalIsCleanWithoutQuery proves the canonical never carries query
// state: a GET with ?foo=bar still resolves to the clean route path (contract
// §16), because the canonical derives from the route, not the request query.
// renderBreadcrumb drives the breadcrumb partial directly using the
// production breadcrumbItem view model (server.go).
func renderBreadcrumb(t *testing.T, items []breadcrumbItem) string {
	t.Helper()
	tmpl := parseTestTemplates(t, "templates/breadcrumb.html")
	var rendered bytes.Buffer
	if err := tmpl.ExecuteTemplate(&rendered, "breadcrumb", struct{ Items []breadcrumbItem }{Items: items}); err != nil {
		t.Fatalf("execute breadcrumb template: %v", err)
	}
	return rendered.String()
}

func renderFooter(t *testing.T, view footerView) string {
	t.Helper()
	tmpl := parseTestTemplates(t, "templates/footer.html")
	var rendered bytes.Buffer
	if err := tmpl.ExecuteTemplate(&rendered, "footer", view); err != nil {
		t.Fatalf("execute footer template: %v", err)
	}
	return rendered.String()
}

// TestBreadcrumbMarkupContracts proves the P1 markup contract
// (seo-patterns.md:50-64): nav → ol → li, every crumb but the last is a real
// link, and the current crumb is a span with aria-current="page" — never an <a>.
func TestBreadcrumbMarkupContracts(t *testing.T) {
	got := renderBreadcrumb(t, []breadcrumbItem{
		{Href: "/", Label: "Home"},
		{Href: "/docs", Label: "Docs"},
		{Label: "Button", Current: true},
	})
	for _, contract := range []string{
		`<nav aria-label="Breadcrumb">`,
		`<ol class="ui-breadcrumb">`,
		`<li class="ui-breadcrumb-item"><a href="/">Home</a></li>`,
		`<li class="ui-breadcrumb-item"><a href="/docs">Docs</a></li>`,
		`<li class="ui-breadcrumb-item"><span aria-current="page">Button</span></li>`,
	} {
		if !strings.Contains(got, contract) {
			t.Errorf("breadcrumb markup is missing contract %q", contract)
		}
	}

	// A trail with no current crumb renders only links.
	noCurrent := renderBreadcrumb(t, []breadcrumbItem{{Href: "/docs", Label: "Docs"}})
	if !strings.Contains(noCurrent, `<a href="/docs">Docs</a>`) {
		t.Error("breadcrumb must render non-current crumbs as links")
	}
	if strings.Contains(noCurrent, "aria-current") {
		t.Error("breadcrumb must not render aria-current when no crumb is current")
	}
}

// TestFooterMarkupContracts proves the footer partial renders the brand, the
// secondary nav with native details/summary sections (open initially),
// and the legal line — zero JS.
func TestFooterMarkupContracts(t *testing.T) {
	got := renderFooter(t, footerView{
		Brand: "Gelium UI",
		Sections: []footerSection{
			{Title: "Documentation", Links: []navLink{{Path: "/docs", Label: "Docs"}}},
		},
		Legal: "© 2026 Gelium UI · MIT",
	})
	for _, contract := range []string{
		`<footer class="ui-footer">`,
		`<p class="ui-footer-brand">Gelium UI</p>`,
		`<nav class="ui-footer-nav" aria-label="Footer">`,
		`<section class="ui-footer-section">`,
		`<details class="ui-footer-details" open>`,
		`<summary class="ui-footer-heading">Documentation</summary>`,
		`<ul class="ui-footer-list">`,
		`<a href="/docs">Docs</a>`,
		`<p class="ui-footer-legal">© 2026 Gelium UI · MIT</p>`,
	} {
		if !strings.Contains(got, contract) {
			t.Errorf("footer markup is missing contract %q", contract)
		}
	}
	if !strings.Contains(got, "<details class=\"ui-footer-details\" open>") {
		t.Error("footer must render groups open initially so desktop links are visible")
	}

	// Nil-safe: brand and legal are optional and omitted when empty.
	minimal := renderFooter(t, footerView{Sections: []footerSection{{Title: "S", Links: []navLink{{Path: "/", Label: "Home"}}}}})
	if strings.Contains(minimal, "ui-footer-brand") || strings.Contains(minimal, "ui-footer-legal") {
		t.Error("footer must omit brand and legal when empty")
	}
}

// TestLayoutOmitsFooterWhenNil proves the pageView.Footer slot is optional: a
// nil footer renders no <footer> and the layout still builds cleanly.
func TestLayoutOmitsFooterWhenNil(t *testing.T) {
	tmpl := parseTestTemplates(t, "templates/*.html")
	var page bytes.Buffer
	data := pageView{Title: "no footer", ThemeClass: "theme-material", Nav: []navLink{{Path: "/", Label: "Home"}}}
	if err := tmpl.ExecuteTemplate(&page, "layout", data); err != nil {
		t.Fatalf("execute layout without footer: %v", err)
	}
	if strings.Contains(page.String(), "<footer") {
		t.Error("layout must not render a footer when pageView.Footer is nil")
	}
}

// TestLayoutRendersFooterAfterMain proves the footer slot renders after the
// </main> landmark and before the toast region when pageView.Footer is set.
func TestLayoutRendersFooterAfterMain(t *testing.T) {
	tmpl := parseTestTemplates(t, "templates/*.html")
	var page bytes.Buffer
	data := pageView{
		Title:      "footer slot",
		ThemeClass: "theme-material",
		Nav:        []navLink{{Path: "/", Label: "Home"}},
		Footer: &footerView{
			Brand: "Gelium UI",
			Sections: []footerSection{
				{Title: "Components", Links: []navLink{{Path: "/components/button", Label: "Button"}}},
			},
			Legal: "© 2026 Gelium UI · MIT",
		},
	}
	if err := tmpl.ExecuteTemplate(&page, "layout", data); err != nil {
		t.Fatalf("execute layout with footer: %v", err)
	}
	body := page.String()
	for _, contract := range []string{
		`<footer class="ui-footer">`,
		`<p class="ui-footer-brand">Gelium UI</p>`,
		`class="ui-footer-heading">Components</summary>`,
		`class="ui-footer-legal">© 2026 Gelium UI · MIT</p>`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("layout does not render footer contract %q", contract)
		}
	}

	mainEnd := strings.Index(body, "</main>")
	toastPos := strings.Index(body, "gelium-toast-region")
	footerPos := strings.Index(body, `<footer class="ui-footer">`)
	if mainEnd < 0 || toastPos < 0 {
		t.Fatal("layout is missing the main landmark or toast region")
	}
	if footerPos < mainEnd || footerPos > toastPos {
		t.Error("footer slot must render after </main> and before the toast region")
	}
}

// TestHomeRendersDefaultFooter proves the real home page ships the footer
// chrome with the default site data: brand, docsNavFor IA groups (Getting
// started + docsSections + Patterns/Recipes/Core-System-Meta), and the legal line.
func TestHomeRendersDefaultFooter(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<footer class="ui-footer">`,
		`<p class="ui-footer-brand">Gelium UI</p>`,
		`<nav class="ui-footer-nav" aria-label="Footer">`,
		`<summary class="ui-footer-heading">Product</summary>`,
		`<summary class="ui-footer-heading">Guidance</summary>`,
		`<summary class="ui-footer-heading">Source</summary>`,
		`<a href="/docs">Docs</a>`,
		`<a href="/docs/agent-workflow">Agent workflow</a>`,
		`<a href="https://github.com/cryptonahue/gelium-ui">GitHub</a>`,
		`<p class="ui-footer-legal">© 2026 Gelium UI · MIT</p>`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("home is missing footer contract %q", contract)
		}
	}
	if got := strings.Count(body, `<details class="ui-footer-details" open>`); got != 3 {
		t.Errorf("landing footer rendered %d expanded groups, want 3", got)
	}
}

func TestCanonicalIsCleanWithoutQuery(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/button?foo=bar", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<link rel="canonical" href="https://gelium-ui.example/components/button">`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("button docs is missing clean canonical %q", contract)
		}
	}
	if strings.Contains(body, "?foo=bar") {
		t.Error("canonical must not carry query state")
	}
}

// TestRobotsTxtPolicy proves GET /robots.txt serves the crawl policy over plain
// text: the public docs site is allowed, the demo/example/recipe surfaces are
// disallowed, and the generated sitemap is advertised (SEO contract §4).
func TestRobotsTxtPolicy(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := res.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want plain UTF-8 text", got)
	}
	body := res.Body.String()
	for _, contract := range []string{
		"User-agent: *",
		"Allow: /",
		"Disallow: /demo/",
		"Disallow: /examples/",
		"Disallow: /recipes/",
		"Sitemap: https://gelium-ui.example/sitemap.xml",
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("robots.txt is missing %q", contract)
		}
	}
}

// TestLlmsTxtServesAgentBrief proves GET /llms.txt serves the embedded
// agent-oriented project brief (install path, wire contracts, component list).
func TestLlmsTxtServesAgentBrief(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/llms.txt", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := res.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want plain UTF-8 text", got)
	}
	body := res.Body.String()
	for _, contract := range []string{
		"# Gelium UI",
		"npm install gelium-ui",
		"gelium-ui/dist/gelium.css",
		"X-Gelium-Validation",
		"gelium:toast",
		"/docs",
		"/docs/compare",
		"theme-material",
		"When NOT to use",
		"/llms-ux.txt",
		"/docs/screens",
		"/docs/feedback",
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("llms.txt is missing %q", contract)
		}
	}
}

// TestSitemapXMLDerivedFromRegistry proves GET /sitemap.xml lists exactly the
// indexable pages from the route registry — home, /docs and every component —
// once each, with absolute URLs, and never lists the noindex/recipe surfaces.
func TestSitemapXMLDerivedFromRegistry(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := res.Header().Get("Content-Type"); got != "application/xml; charset=utf-8" {
		t.Errorf("Content-Type = %q, want application/xml", got)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`,
		`<loc>https://gelium-ui.example/</loc>`,
		`<loc>https://gelium-ui.example/docs</loc>`,
		`<loc>https://gelium-ui.example/docs/patterns</loc>`,
		`<loc>https://gelium-ui.example/docs/information-architecture</loc>`,
		`<loc>https://gelium-ui.example/docs/themes</loc>`,
		`<loc>https://gelium-ui.example/docs/tokens</loc>`,
		`<loc>https://gelium-ui.example/docs/server-contracts</loc>`,
		`<loc>https://gelium-ui.example/docs/accessibility</loc>`,
		`<loc>https://gelium-ui.example/docs/principles</loc>`,
		`<loc>https://gelium-ui.example/docs/content-style</loc>`,
		`<loc>https://gelium-ui.example/docs/acknowledgments</loc>`,
		`<loc>https://gelium-ui.example/components/button</loc>`,
		`<loc>https://gelium-ui.example/components/data-table</loc>`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("sitemap is missing %q", contract)
		}
	}
	// home + /docs + patterns + handbook pages + all components. The handbook
	// count derives from handbookNavLinks() (same registry the sitemap uses), so
	// adding a handbook page can never drift this total again.
	if got := strings.Count(body, "<url>"); got != len(componentRoutes())+len(handbookNavLinks())+8+len(blogPosts)+len(publishedSectionRefPaths()) {
		t.Errorf("sitemap <url> entries = %d, want %d (base pages + handbook + section-ref fichas + components + blog registry)", got, len(componentRoutes())+len(handbookNavLinks())+8+len(blogPosts)+len(publishedSectionRefPaths()))
	}
	for _, excluded := range []string{"/demo/", "/examples/", "/recipes/", "/components/dialog/confirm"} {
		if strings.Contains(body, excluded) {
			t.Errorf("sitemap must not list noindex/form surface %q", excluded)
		}
	}
}

// TestComponentPageRendersStructuredData proves every /components/* page emits
// a single valid JSON-LD block with the BreadcrumbList trail
// (Home > Components > page) and a TechArticle carrying the page headline and
// canonical URL (SEO contract §12).
func TestComponentPageRendersStructuredData(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/button", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<script type="application/ld+json">`,
		`"@context":"https://schema.org"`,
		`"@type":"BreadcrumbList"`,
		`"name":"Home"`,
		`"name":"Components"`,
		`"item":"https://gelium-ui.example/components/button"`,
		`"@type":"TechArticle"`,
		`"headline":"Button · Gelium UI"`,
		`"url":"https://gelium-ui.example/components/button"`,
		`"name":"Gelium UI"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("button docs JSON-LD is missing %q", contract)
		}
	}
	if !json.Valid([]byte(extractJSONLD(t, body))) {
		t.Error("component JSON-LD must parse with encoding/json")
	}
}

// extractJSONLD returns the text inside the first application/ld+json script.
func extractJSONLD(t *testing.T, body string) string {
	t.Helper()
	const open = `<script type="application/ld+json">`
	start := strings.Index(body, open)
	if start < 0 {
		t.Fatal("body has no ld+json script")
	}
	start += len(open)
	end := strings.Index(body[start:], `</script>`)
	if end < 0 {
		t.Fatal("ld+json script is not closed")
	}
	return body[start : start+end]
}

// visibleUpdatedDate returns the ISO date rendered on the visible provenance
// line ("Updated <time>…</time>") of a component page, if present.
func visibleUpdatedDate(t *testing.T, body string) string {
	t.Helper()
	m := regexp.MustCompile(`Updated <time[^>]*>([0-9]{4}-[0-9]{2}-[0-9]{2})</time>`).FindStringSubmatch(body)
	if m == nil {
		t.Fatal("body has no visible Updated date line")
	}
	return m[1]
}

// TestComponentPageRendersSoftwareApplicationAndDates proves every
// /components/* page includes a SoftwareApplication node in its JSON-LD
// @graph (name "Gelium UI", applicationCategory "DeveloperApplication",
// softwareVersion from the single docsShellVersion source, operatingSystem
// "Any", license "MIT" — GEO §14) and that the TechArticle
// datePublished/dateModified equal the visible provenance line dates
// (GEO §7).
func TestComponentPageRendersSoftwareApplicationAndDates(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/button", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()

	var graph struct {
		Context string           `json:"@context"`
		Graph   []map[string]any `json:"@graph"`
	}
	if err := json.Unmarshal([]byte(extractJSONLD(t, body)), &graph); err != nil {
		t.Fatalf("parse JSON-LD: %v", err)
	}
	var appNode, articleNode map[string]any
	for _, node := range graph.Graph {
		switch node["@type"] {
		case "SoftwareApplication":
			appNode = node
		case "TechArticle":
			articleNode = node
		}
	}
	if appNode == nil {
		t.Fatal("JSON-LD @graph has no SoftwareApplication node")
	}
	for _, want := range []struct{ key, value string }{
		{key: "name", value: "Gelium UI"},
		{key: "applicationCategory", value: "DeveloperApplication"},
		{key: "softwareVersion", value: docsShellVersion},
		{key: "operatingSystem", value: "Any"},
		{key: "license", value: "MIT"},
	} {
		if got, _ := appNode[want.key].(string); got != want.value {
			t.Errorf("SoftwareApplication %s = %q, want %q", want.key, got, want.value)
		}
	}
	if articleNode == nil {
		t.Fatal("JSON-LD @graph has no TechArticle node")
	}
	published, _ := articleNode["datePublished"].(string)
	modified, _ := articleNode["dateModified"].(string)
	if published == "" || modified == "" {
		t.Fatalf("TechArticle dates missing: datePublished=%q dateModified=%q", published, modified)
	}
	visible := visibleUpdatedDate(t, body)
	if published != visible || modified != visible {
		t.Errorf("TechArticle dates (%s, %s) must match the visible Updated line (%s)", published, modified, visible)
	}
}

// TestErrorPagesCarryMetadata proves 404/500 pages populate Meta through
// resolveMeta so every error page carries description, canonical, robots and
// OG tags per the route contract, with the real status preserved (SEO §16).
func TestErrorPagesCarryMetadata(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/does-not-exist", nil))

	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<meta name="description" content="Gelium UI — open-code, server-rendered UI components with native HTML semantics.">`,
		`<link rel="canonical" href="https://gelium-ui.example/does-not-exist">`,
		`<meta name="robots" content="noindex, nofollow">`,
		`<meta property="og:title" content="Page not found · Gelium UI">`,
		`<meta property="og:url" content="https://gelium-ui.example/does-not-exist">`,
		`<meta property="og:image" content="https://gelium-ui.example/og.png">`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("404 page is missing %q", contract)
		}
	}

	// 500 path: renderErrorPage must populate the same head metadata through
	// resolveMeta and preserve the internal-error status. A server is built
	// directly (no natural HTTP route 500s) and the page-level failure path
	// exercised with a real route identity.
	templates := parseTestTemplates(t, "templates/*.html")
	s := &server{templates: templates, markdown: goldmark.New(), assets: webassets.Assets}
	res500 := httptest.NewRecorder()
	s.renderErrorPage(res500, nil, http.StatusInternalServerError, "Something went wrong", "boom", true, "/", "Back to home", "/recipes/ops-queue")
	if res500.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", res500.Code, http.StatusInternalServerError)
	}
	body500 := res500.Body.String()
	for _, contract := range []string{
		`<link rel="canonical" href="https://gelium-ui.example/recipes/ops-queue">`,
		`<meta name="robots" content="noindex, nofollow">`,
	} {
		if !strings.Contains(body500, contract) {
			t.Errorf("500 page is missing %q", contract)
		}
	}
}

// TestComponentPageWithoutDateEntryOmitsDates proves a component route absent
// from the date table emits no dates: the TechArticle carries no
// datePublished/dateModified and the article shows no provenance line
// (GEO §7 scenario 2). Unregistered /components/* paths fall back to the
// styled 404 and must not fabricate a component identity or dates.
func TestComponentPageWithoutDateEntryOmitsDates(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/ghost", nil))
	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
	body := res.Body.String()
	if strings.Contains(body, `"datePublished"`) || strings.Contains(body, `"dateModified"`) {
		t.Error("route without date-table entry must not emit datePublished/dateModified")
	}
	if strings.Contains(body, "Updated") {
		t.Error("route without date-table entry must not render a provenance line")
	}

	// The table lookup itself misses unknown slugs and hits known ones.
	if _, ok := docDatesFor("ghost"); ok {
		t.Error("docDatesFor(ghost) must report not found")
	}
	d, ok := docDatesFor("button")
	if !ok {
		t.Fatal("docDatesFor(button) must be found")
	}
	if d.Published != "2026-08-09" || d.Modified != "2026-08-09" {
		t.Errorf("docDatesFor(button) = published %q modified %q, want 2026-08-09/2026-08-09", d.Published, d.Modified)
	}
}

// TestResolveBaseURL proves the canonical origin resolves from BASE_URL at
// startup: empty input falls back to the default const, a trailing slash is
// trimmed, and every absolute URL family (canonical, og:url, JSON-LD item,
// sitemap loc) derives from the resolved origin (SEO §2).
func TestResolveBaseURL(t *testing.T) {
	if got := resolveBaseURL(""); got != defaultBaseURL {
		t.Errorf(`resolveBaseURL("") = %q, want %q`, got, defaultBaseURL)
	}
	if got := resolveBaseURL("https://docs.example.com/"); got != "https://docs.example.com" {
		t.Errorf(`resolveBaseURL("https://docs.example.com/") = %q, want origin without trailing slash`, got)
	}
	if got := resolveBaseURL("https://docs.example.com"); got != "https://docs.example.com" {
		t.Errorf(`resolveBaseURL("https://docs.example.com") = %q, want unchanged origin`, got)
	}
	if got := resolveBaseURL("https://docs.example.com///"); got != "https://docs.example.com" {
		t.Errorf(`resolveBaseURL("https://docs.example.com///") = %q, want all trailing slashes trimmed`, got)
	}

	// With no BASE_URL in the environment, every absolute URL family derives
	// from the resolved default origin rather than a hardcoded literal.
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/button", nil))
	body := res.Body.String()
	for _, contract := range []string{
		`<link rel="canonical" href="https://gelium-ui.example/components/button">`,
		`<meta property="og:url" content="https://gelium-ui.example/components/button">`,
		`"item":"https://gelium-ui.example/components/button"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("button page missing %q (must derive from resolved origin)", contract)
		}
	}
	sm := httptest.NewRecorder()
	New().ServeHTTP(sm, httptest.NewRequest(http.MethodGet, "/sitemap.xml", nil))
	if !strings.Contains(sm.Body.String(), "<loc>https://gelium-ui.example/components/button</loc>") {
		t.Error("sitemap loc must derive from the resolved origin")
	}
}

// answerFirstIntroSignal is a deterministic proxy for the GEO §1-4
// answer-first rule: the intro paragraph must state what the component is
// (it names the component) AND when to use it. A when-clause ("Use X when…",
// "…when you need…") or an explicit "use … for/to" purpose expresses that.
var answerFirstIntroSignal = regexp.MustCompile(`\bwhen\b|(?i)\buse\b.{0,80}\b(for|to|when)\b`)

// firstProseParagraph returns the first prose paragraph of a markdown doc:
// the first run of consecutive non-empty lines after the H1 whose first line
// is not a non-paragraph block (heading, blockquote, list, table, code fence,
// raw HTML). GEO §1-4 reads this paragraph in isolation.
func firstProseParagraph(t *testing.T, source string) string {
	t.Helper()
	lines := strings.Split(source, "\n")
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "# ") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		t.Fatal("document has no H1")
	}
	for i := start; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") ||
			strings.HasPrefix(trimmed, ">") ||
			strings.HasPrefix(trimmed, "-") ||
			strings.HasPrefix(trimmed, "*") ||
			strings.HasPrefix(trimmed, "|") ||
			strings.HasPrefix(trimmed, "`") ||
			strings.HasPrefix(trimmed, "<") ||
			regexp.MustCompile(`^\d+\.`).MatchString(trimmed) {
			continue
		}
		var para []string
		for j := i; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "" {
				break
			}
			para = append(para, strings.TrimSpace(lines[j]))
		}
		if len(para) > 0 {
			return strings.Join(para, " ")
		}
	}
	t.Fatal("document has no prose paragraph after the H1")
	return ""
}

// TestContentFilesLeadWithAnswerFirst proves every component content file
// leads with an answer-first summary (GEO §1-4): the first prose paragraph
// after the H1 is non-empty, at least 40 chars, names the component, and
// states when to use it — self-contained, with no back-references. The slug
// table grows in two chunks: chunk A here, chunk B added by task 4.1.
func TestContentFilesLeadWithAnswerFirst(t *testing.T) {
	slugs := []string{
		// Chunk A (PR2): data entry, selection and feedback.
		"text-field", "checkbox", "radio", "switch", "select", "slider",
		"progress", "segmented-button", "menu", "chips", "tabs", "list",
		"data-table",
		// Chunk B (PR3): display, navigation and atoms.
		"badge", "button", "card", "dialog", "divider", "elevation", "fab",
		"focus-ring", "icon", "icon-button", "toast", "tooltip",
		"navigation-bar", "navigation-drawer", "navigation-tab", "index",
	}
	for _, slug := range slugs {
		t.Run(slug, func(t *testing.T) {
			source, err := fs.ReadFile(webassets.Assets, "content/"+slug+".md")
			if err != nil {
				t.Fatalf("read content/%s.md: %v", slug, err)
			}
			para := firstProseParagraph(t, string(source))
			if len(para) < 40 {
				t.Errorf("answer-first intro is only %d chars (want ≥40): %q", len(para), para)
			}
			name := strings.ReplaceAll(slug, "-", " ")
			// index.md is the landing page, not a component — there is no
			// component name to match, only the answer-first intro itself.
			if slug != "index" && !strings.Contains(strings.ToLower(para), name) {
				t.Errorf("answer-first intro must name the component %q: %q", name, para)
			}
			if !answerFirstIntroSignal.MatchString(para) {
				t.Errorf("answer-first intro must state when to use the component: %q", para)
			}
		})
	}
}
func TestHomeRendersOGImageAndTwitterCard(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<meta property="og:image" content="https://gelium-ui.example/og.png">`,
		`<meta name="twitter:card" content="summary_large_image">`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("home is missing %q", contract)
		}
	}
}

// TestComponentPageRendersOGImage proves component pages also carry the default
// og:image placeholder so every indexable page ships a social image.
func TestComponentPageRendersOGImage(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/button", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<meta property="og:image" content="https://gelium-ui.example/og.png">`,
		`<meta name="twitter:card" content="summary_large_image">`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("button docs is missing %q", contract)
		}
	}
}

// TestLayoutRendersSkipLinkToMain proves every layout page ships the skip link
// as the first focusable element targeting the main landmark (G7). Home keeps
// the legacy centered column; docs shell routes use main.docs-shell-content
// inside the two-pane frame (task 3.2).
func TestLayoutRendersSkipLinkToMain(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		mainClass string
	}{
		{
			name:      "home marketing landing main",
			path:      "/",
			mainClass: `<main id="main-content" class="ui-landing-main">`,
		},
		{
			name:      "docs shell content column",
			path:      "/docs",
			mainClass: `<main id="main-content" class="docs-shell-content">`,
		},
		{
			name:      "component shell content column",
			path:      "/components/button",
			mainClass: `<main id="main-content" class="docs-shell-content">`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
			}
			body := res.Body.String()
			for _, contract := range []string{
				`<a class="ui-skip-link" href="#main-content">Skip to main content</a>`,
				tt.mainClass,
			} {
				if !strings.Contains(body, contract) {
					t.Errorf("%s is missing %q", tt.path, contract)
				}
			}
			// Shell pages must not keep the home-only centered main utility.
			if tt.path != "/" {
				if strings.Contains(body, `class="docs-shell docs-content"`) {
					t.Errorf("%s must not use home main.docs-shell.docs-content", tt.path)
				}
				if strings.Contains(body, `class="site-header"`) {
					t.Errorf("%s must not render legacy site-header", tt.path)
				}
			}
			skipPos := strings.Index(body, `class="ui-skip-link"`)
			mainPos := strings.Index(body, `id="main-content"`)
			if skipPos < 0 || mainPos < 0 || skipPos > mainPos {
				t.Error("skip link must render before the main landmark")
			}
		})
	}
}

// TestComponentBreadcrumbTrailMatchesStructuredData proves the visible
// breadcrumb on a component page is Home > Components > <label>, matching the
// BreadcrumbList JSON-LD so the rendered navigation and the structured data
// never disagree (SEO contract §11).
func TestComponentBreadcrumbTrailMatchesStructuredData(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/button", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		`<nav aria-label="Breadcrumb">`,
		`<a href="/">Home</a>`,
		`<a href="/docs">Components</a>`,
		`<span aria-current="page">Button</span>`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("component breadcrumb is missing %q", contract)
		}
	}
}

// TestThemeQueryParamSelectsTheme proves the document-root theme selection
// (Phase H): a valid ?theme= query renders that theme's class on the document
// root without JS, an unknown theme falls back to the default, and the
// default stays theme-material when no query is present.
func TestThemeQueryParamSelectsTheme(t *testing.T) {
	cases := []struct {
		name  string
		query string
		want  string
	}{
		{name: "basecoat via query", query: "?theme=basecoat", want: `class="theme-basecoat"`},
		{name: "material via query", query: "?theme=material", want: `class="theme-material"`},
		{name: "unknown falls back to default", query: "?theme=unknown", want: `class="theme-material"`},
		{name: "no query keeps default", query: "", want: `class="theme-material"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/button"+tc.query, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
			}
			if !strings.Contains(res.Body.String(), tc.want) {
				t.Errorf("rendered page missing %q", tc.want)
			}
		})
	}
}

// TestLayoutUsesHxBoostForInternalNavigation pins the SPA-style navigation
// contract: the document body carries hx-boost so internal links navigate
// via AJAX (htmx swaps the body, URL and title update via pushState), which
// removes the full-page reload that made docs navigation feel rough.
func TestLayoutUsesHxBoostForInternalNavigation(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/docs/tokens", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	if !strings.Contains(body, `hx-boost:inherited="true"`) {
		t.Errorf("layout must inherit boost for SPA-style navigation, got: %s", firstLineContaining(body, "<body"))
	}
}

func firstLineContaining(haystack, needle string) string {
	for _, line := range strings.Split(haystack, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	return ""
}
