package app

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// TestDocsSearchEnabledInShell proves the docs search input is a real,
// enabled control: a live GET form to the /docs hub (0-JS fallback submits
// /docs?q=<term>) with no disabled state and no "coming soon" placeholder.
func TestDocsSearchEnabledInShell(t *testing.T) {
	body := getOKBody(t, "/docs")
	for _, contract := range []string{
		`<form class="docs-search" method="get" action="/docs" role="search">`,
		`id="docs-search"`,
		`type="search"`,
		`name="q"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("docs shell search missing contract %q", contract)
		}
	}
	input := openingTagWithID(t, body, "input", "docs-search")
	if strings.Contains(input, "disabled") || strings.Contains(input, "aria-disabled") {
		t.Error("docs search input must not be disabled")
	}
	if strings.Contains(body, "coming soon") {
		t.Error("search must not advertise coming soon")
	}
	// The same live form renders on component pages too.
	comp := getOKBody(t, "/components/button")
	if !strings.Contains(comp, `<form class="docs-search" method="get" action="/docs" role="search">`) {
		t.Error("component pages must render the live docs search form")
	}
}

// TestDocsSearchIndexEmitted proves the shell emits a JSON index of the nav
// model ({title, href, group}) that search.js can filter client-side, and that
// the entries match the real nav destinations.
func TestDocsSearchIndexEmitted(t *testing.T) {
	body := getOKBody(t, "/docs")
	m := regexp.MustCompile(`(?s)<script type="application/json" id="docs-search-index">(.*?)</script>`).FindStringSubmatch(body)
	if len(m) != 2 {
		t.Fatal("docs shell must emit the search index script")
	}
	var entries []struct {
		Title string `json:"title"`
		Href  string `json:"href"`
		Group string `json:"group"`
	}
	if err := json.Unmarshal([]byte(m[1]), &entries); err != nil {
		t.Fatalf("search index must be valid JSON: %v", err)
	}
	if len(entries) < 10 {
		t.Fatalf("search index entries = %d, want at least 10 (whole nav model)", len(entries))
	}
	byHref := map[string]struct{ title, group string }{}
	for _, e := range entries {
		byHref[e.Href] = struct{ title, group string }{e.Title, e.Group}
	}
	for _, want := range []struct{ href, title, group string }{
		{"/docs", "Documentation", "Getting started"},
		{"/docs/themes", "Themes", "System"},
		{"/components/button", "Button", "Actions"},
	} {
		got, ok := byHref[want.href]
		if !ok {
			t.Errorf("search index missing %s", want.href)
			continue
		}
		if got.title != want.title || got.group != want.group {
			t.Errorf("index[%s] = {title:%q group:%q}, want {title:%q group:%q}",
				want.href, got.title, got.group, want.title, want.group)
		}
	}
}

// TestDocsSearchPreservesChrome proves the search fallback and the index keep
// the current theme/scheme: the GET form carries hidden theme/scheme inputs
// and the index hrefs keep the chrome query (same rule as the sidebar links).
func TestDocsSearchPreservesChrome(t *testing.T) {
	body := getOKBody(t, "/docs?theme=basecoat&scheme=dark")
	if !strings.Contains(body, `<input type="hidden" name="theme" value="basecoat">`) {
		t.Error("search form must carry hidden theme=basecoat for the GET fallback")
	}
	if !strings.Contains(body, `<input type="hidden" name="scheme" value="dark">`) {
		t.Error("search form must carry hidden scheme=dark for the GET fallback")
	}
	if !strings.Contains(body, `theme=basecoat`) || !strings.Contains(body, `scheme=dark`) {
		t.Error("search index hrefs must preserve the chrome query")
	}
}

// TestDocsTopbarGitHubLinkStaysClean proves the external repository link in
// the docs topbar is never chrome-rewritten: ?theme= / ?scheme= must not leak
// onto the GitHub href (same rule as the landing external CTA).
func TestDocsTopbarGitHubLinkStaysClean(t *testing.T) {
	body := getOKBody(t, "/components/button?theme=basecoat&scheme=dark")
	if !strings.Contains(body, `href="https://github.com/cryptonahue/gelium-ui"`) {
		t.Error("docs topbar must link https://github.com/cryptonahue/gelium-ui")
	}
	if !strings.Contains(body, "View source") {
		t.Error("topbar GitHub link must be labelled View source")
	}
	if strings.Contains(body, "github.com/cryptonahue/gelium-ui?") {
		t.Error("external topbar GitHub href must not carry the chrome query string")
	}
	// The landing's own GitHub CTA keeps working on non-shell routes.
	home := getOKBody(t, "/")
	if !strings.Contains(home, `href="https://github.com/cryptonahue/gelium-ui"`) {
		t.Error("landing must keep its GitHub link")
	}
}
