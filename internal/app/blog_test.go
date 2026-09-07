package app

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	webassets "geliumui/site/web"
)

// blogPostContracts is a per-slug content marker that makes each post page
// real rather than a stub (same pattern as handbookRoutes). The marker must
// appear in the rendered post body.
var blogPostContracts = map[string]string{
	"introducing": "zero required JavaScript",
	"genesis":     "zero required JS",
	"roadmap":     "contract-tested",
}

// TestBlogIndexListsPosts proves /blog renders the separate blog space: an
// index of every registered post as a Gelium card grid (dogfooding ui-card
// markup), with title, summary and date per card. It must keep the legacy
// site-header chrome — the blog is NOT the docs two-pane shell.
func TestBlogIndexListsPosts(t *testing.T) {
	body := getOKBody(t, "/blog")
	// Separate space: legacy site-header chrome, not the docs shell.
	if strings.Contains(body, `class="docs-chrome"`) {
		t.Error("blog must not render docs-topbar")
	}
	if !strings.Contains(body, `class="site-header"`) {
		t.Error("blog must keep the site-header chrome")
	}
	if !strings.Contains(body, `class="ui-blog-main"`) {
		t.Error("blog index must render the blog frame (ui-blog-main)")
	}
	if !strings.Contains(body, ">Gelium blog</h1>") {
		t.Error("blog index must render <h1>Gelium blog</h1>")
	}
	if !strings.Contains(body, `class="ui-card ui-card-outlined ui-blog-card"`) {
		t.Error("blog index must dogfood Gelium card markup (ui-card link card)")
	}
	if len(blogPosts) == 0 {
		t.Fatal("blogPosts registry is empty")
	}
	for _, p := range blogPosts {
		if !strings.Contains(body, p.Title) {
			t.Errorf("blog index missing post title %q", p.Title)
		}
		if !strings.Contains(body, p.Summary) {
			t.Errorf("blog index missing summary %q", p.Summary)
		}
		if !strings.Contains(body, p.Date) {
			t.Errorf("blog index missing date %q for %s", p.Date, p.Slug)
		}
		if !strings.Contains(body, `href="/blog/`+p.Slug+`"`) {
			t.Errorf("blog index missing card link /blog/%s", p.Slug)
		}
	}
}

// TestBlogPostsRender proves every registered post renders 200 with exactly
// one h1 (the post title), the long-form prose article, and the
// author/date/reading-time provenance line plus a link back to the index.
func TestBlogPostsRender(t *testing.T) {
	for _, p := range blogPosts {
		t.Run(p.Slug, func(t *testing.T) {
			res := httptest.NewRecorder()
			New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/blog/"+p.Slug, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("GET /blog/%s status = %d, want 200", p.Slug, res.Code)
			}
			body := res.Body.String()
			if got := strings.Count(body, "<h1"); got != 1 {
				t.Errorf("post page h1 count = %d, want 1", got)
			}
			if !strings.Contains(body, ">"+p.Title+"</h1>") {
				t.Errorf("post page must render <h1>%s</h1>", p.Title)
			}
			if !strings.Contains(body, `class="prose ui-blog-post"`) {
				t.Error("post prose must reuse the .prose readability class")
			}
			if !strings.Contains(body, `class="ui-blog-provenance"`) {
				t.Error("post must render the ui-blog-provenance line")
			}
			if !strings.Contains(body, p.Author) {
				t.Errorf("post provenance missing author %q", p.Author)
			}
			if !strings.Contains(body, "min read") {
				t.Error("post provenance must include the reading time")
			}
			if !strings.Contains(body, `href="/blog"`) {
				t.Error("post page must link back to the blog index")
			}
			if marker, ok := blogPostContracts[p.Slug]; ok {
				if !strings.Contains(body, marker) {
					t.Errorf("post %s is missing its content contract %q", p.Slug, marker)
				}
			}
		})
	}
}

// TestBlogRegistryIntegrity proves every registry entry points at a real
// embedded content file and carries complete metadata: unique slug, ISO-8601
// date, author, summary, and the conventional content/blog-<slug>.md path.
func TestBlogRegistryIntegrity(t *testing.T) {
	if len(blogPosts) < 2 {
		t.Fatalf("blogPosts has %d entries, want at least 2", len(blogPosts))
	}
	seen := map[string]bool{}
	for _, p := range blogPosts {
		if p.Slug == "" {
			t.Error("post with empty slug")
		}
		if seen[p.Slug] {
			t.Errorf("duplicate blog slug %q", p.Slug)
		}
		seen[p.Slug] = true
		if p.Title == "" {
			t.Errorf("post %s has empty title", p.Slug)
		}
		if p.Date == "" {
			t.Errorf("post %s has empty date", p.Slug)
		}
		if _, err := time.Parse("2006-01-02", p.Date); err != nil {
			t.Errorf("post %s date %q is not ISO-8601: %v", p.Slug, p.Date, err)
		}
		if p.Author == "" {
			t.Errorf("post %s has empty author", p.Slug)
		}
		if p.Summary == "" {
			t.Errorf("post %s has empty summary", p.Slug)
		}
		wantPath := "content/blog-" + p.Slug + ".md"
		if p.ContentPath != wantPath {
			t.Errorf("post %s ContentPath = %q, want %q", p.Slug, p.ContentPath, wantPath)
		}
		if _, err := fs.ReadFile(webassets.Assets, p.ContentPath); err != nil {
			t.Errorf("post %s content file %s unreadable: %v", p.Slug, p.ContentPath, err)
		}
	}
	if !seen["genesis"] || !seen["roadmap"] {
		t.Errorf("blogPosts must include genesis and roadmap; got %v", seen)
	}
}

// TestBlogStaysOutOfDocsShell proves the blog is a separate space: /blog and
// its posts never enter the docs shell nav model or the two-pane chrome.
func TestBlogStaysOutOfDocsShell(t *testing.T) {
	if usesDocsShell("/blog") {
		t.Error("usesDocsShell(/blog) must be false")
	}
	if usesDocsShell("/blog/genesis") {
		t.Error("usesDocsShell(/blog/genesis) must be false")
	}
	nav := docsNavFor("/docs", "", "")
	for _, g := range nav.Groups {
		for _, l := range g.Links {
			if l.Path == "/blog" || strings.HasPrefix(l.Path, "/blog/") {
				t.Errorf("docs nav model must not contain blog path %q", l.Path)
			}
		}
	}
	body := getOKBody(t, "/components/button")
	if strings.Contains(body, `class="ui-list-item-link" href="/blog`) {
		t.Error("docs sidebar must not link into the blog")
	}
}

// TestBlogPostsKeepSentencesUnder25Words applies the copy contract (content
// style guide §Paragraphs and sentences) to the blog posts: every sentence at
// most 25 words, using the exact same mechanical counting as the component
// pages (componentProse). The component contract itself is untouched.
func TestBlogPostsKeepSentencesUnder25Words(t *testing.T) {
	offenders := 0
	for _, p := range blogPosts {
		for _, sentence := range componentProse(t, "blog-"+p.Slug) {
			if n := len(strings.Fields(sentence)); n > 25 {
				offenders++
				t.Errorf("blog-%s.md has a %d-word sentence (> 25): %q", p.Slug, n, sentence)
			}
		}
	}
	if offenders > 0 {
		t.Fatalf("found %d sentences over 25 words across blog posts", offenders)
	}
}

// TestBlogTopbarLinkKeepsChromeQuery proves the docs topbar Blog link carries
// the allowlisted chrome query under ?theme=/?scheme=, so boosted navigation
// from the docs shell never silently resets direction or light/dark (owner
// decision 2026-08-15: the blog is a separate space but honors the chrome on
// direct navigation). It stays plain when no theme/scheme is selected.
func TestBlogTopbarLinkKeepsChromeQuery(t *testing.T) {
	body := getOKBody(t, "/components/button?theme=basecoat&scheme=dark")
	if !strings.Contains(body, `class="ui-button ui-button-text docs-chrome-blog"`) ||
		!strings.Contains(body, `href="/blog?scheme=dark&amp;theme=basecoat"`) {
		t.Error("docs topbar Blog link must carry the chrome query under ?theme=/?scheme=")
	}
	plain := getOKBody(t, "/components/button")
	if !strings.Contains(plain, `class="ui-button ui-button-text docs-chrome-blog"`) ||
		!strings.Contains(plain, `href="/blog"`) {
		t.Error("docs topbar Blog link must stay plain without theme/scheme")
	}
	// The blog space itself still honors the chrome query on its own pages.
	blog := getOKBody(t, "/blog?theme=basecoat")
	if !strings.Contains(blog, `class="theme-basecoat"`) {
		t.Error("blog page must apply ?theme=basecoat to the document root")
	}
}

// TestBlogCssInBundle proves the blog stylesheet compiles into the single
// served CSS bundle (blog.css is imported by app.css).
func TestBlogCssInBundle(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/static/app.css", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("app.css status = %d, want 200", res.Code)
	}
	for _, want := range []string{".ui-blog-main", ".ui-blog-card-grid", ".ui-blog-provenance"} {
		if !strings.Contains(res.Body.String(), want) {
			t.Errorf("served app.css is missing %s (blog.css not in bundle)", want)
		}
	}
}

// TestBlogUnknownSlug404 proves the /blog/{slug} route rejects slugs outside
// the registry with the styled error page instead of rendering a ghost post.
func TestBlogUnknownSlug404(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/blog/not-a-post", nil))
	if res.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotFound)
	}
}

// TestReadingTimeMinutes proves the reading-time estimate is a pure function
// of the markdown word count: at least one minute, rounded up at 200 wpm.
func TestReadingTimeMinutes(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   int
	}{
		{name: "empty", source: "", want: 1},
		{name: "short post", source: "One short paragraph with a handful of words.", want: 1},
		{name: "about 400 words", source: strings.Repeat("word ", 400), want: 2},
		{name: "about 450 words", source: strings.Repeat("word ", 450), want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readingTimeMinutesFor(tt.source); got != tt.want {
				t.Errorf("readingTimeMinutesFor(%q) = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
	// Real posts must yield a positive reading time shown on the page.
	for _, p := range blogPosts {
		src, err := fs.ReadFile(webassets.Assets, p.ContentPath)
		if err != nil {
			t.Fatalf("read %s: %v", p.ContentPath, err)
		}
		if got := readingTimeMinutesFor(string(src)); got < 1 {
			t.Errorf("post %s reading time = %d, want >= 1", p.Slug, got)
		}
	}
}
