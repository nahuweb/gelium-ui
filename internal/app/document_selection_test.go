package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveDocumentSelectionPrecedenceAndLegacyCompatibility(t *testing.T) {
	tests := []struct {
		query       string
		behavior    accordionBehavior
		reference   referencePreset
		skin        productSkin
		legacyClass string
		canonical   bool
	}{
		{"theme=material", accordionBehaviorNative, referenceMaterial, skinNone, "theme-material", false},
		{"theme=basecoat", accordionBehaviorNative, referenceBasecoat, skinNone, "theme-basecoat", false},
		{"theme=alden", accordionBehaviorNative, referenceNone, skinAlden, "theme-alden", false},
		{"behavior=material&reference=auto", accordionBehaviorMaterial, referenceMaterial, skinNone, "theme-material", true},
		{"behavior=baseui&reference=default", accordionBehaviorBaseUI, referenceBaseUI, skinNone, "theme-baseui", true},
		{"behavior=native&reference=auto", accordionBehaviorNative, referenceNone, skinNone, "", true},
		{"theme=material&reference=basecoat&skin=vercel", accordionBehaviorNative, referenceBasecoat, skinVercel, "theme-vercel", false},
		{"theme=alden&reference=material&skin=none", accordionBehaviorNative, referenceMaterial, skinNone, "theme-material", false},
		{"theme=basecoat&reference=<script>&skin=onload%3Dalert(1)", accordionBehaviorNative, referenceBasecoat, skinNone, "theme-basecoat", false},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			q, err := url.ParseQuery(tt.query)
			if err != nil {
				t.Fatal(err)
			}
			got := resolveDocumentSelection(q)
			if got.Behavior != tt.behavior || got.Reference != tt.reference || got.Skin != tt.skin || got.Canonical != tt.canonical || got.legacyThemeClass() != tt.legacyClass {
				t.Fatalf("resolveDocumentSelection(%q) = %+v class=%q", tt.query, got, got.legacyThemeClass())
			}
		})
	}
}

func TestExplicitNativeDefaultRendersNeutralRootWithoutLegacyMaterial(t *testing.T) {
	const path = "/components/accordion?behavior=native&execution=native&reference=auto&scheme=dark&skin=none"
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d", res.Code)
	}
	body := res.Body.String()
	if !strings.Contains(body, `data-gelium-reference="none" data-gelium-skin="none" data-gelium-contract="gelium" data-gelium-scheme="dark"`) {
		t.Fatalf("native default must render its resolved neutral attributes: %s", htmlClassSnippet(body))
	}
	if strings.Contains(body, `theme-material`) {
		t.Fatalf("explicit native default must not carry legacy Material class: %s", htmlClassSnippet(body))
	}
	if !strings.Contains(body, `class="theme-dark"`) {
		t.Fatalf("explicit dark scheme must retain the generic dark class: %s", htmlClassSnippet(body))
	}
}

func TestRecipeSubmissionCanonicalizesVisualAndDropsLegacyTheme(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/components/accordion?theme=basecoat&behavior=material&visual=vercel&scheme=dark&execution=native&foo=bar&unsafe=%3Cscript%3E", nil))
	if res.Code != http.StatusSeeOther {
		t.Fatalf("Recipe submission status = %d, want %d", res.Code, http.StatusSeeOther)
	}
	if got, want := res.Header().Get("Location"), "/components/accordion?behavior=material&execution=native&reference=auto&scheme=dark&skin=vercel"; got != want {
		t.Fatalf("Recipe submission location = %q, want %q", got, want)
	}

	unsafe := httptest.NewRecorder()
	New().ServeHTTP(unsafe, httptest.NewRequest(http.MethodGet, "/components/accordion?visual=%3Cscript%3E&skin=%3Cscript%3E", nil))
	if unsafe.Code != http.StatusOK || strings.Contains(unsafe.Body.String(), "ZgotmplZ") || strings.Contains(unsafe.Body.String(), "&lt;script&gt;") {
		t.Fatalf("unknown visual must remain a safe normal request; status=%d", unsafe.Code)
	}
}

// TestRecipePrimaryRoutesRenderOneFullNativeForm proves the primary Recipe
// surfaces give the full selector sole ownership: the docs topbar does not
// render a second compact copy that would duplicate form controls or IDs.
func TestRecipePrimaryRoutesRenderOneFullNativeForm(t *testing.T) {
	for _, path := range []string{"/components/accordion", "/docs/themes/gallery"} {
		t.Run(path, func(t *testing.T) {
			res := httptest.NewRecorder()
			New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
			}
			body := res.Body.String()
			if got := strings.Count(body, `class="ui-recipe-switcher ui-recipe-switcher--full"`); got != 1 {
				t.Fatalf("full Recipe forms = %d, want 1", got)
			}
			if got := strings.Count(body, `ui-recipe-switcher--compact`); got != 0 {
				t.Fatalf("compact Recipe forms = %d, want 0 on a primary route", got)
			}
			for _, id := range []string{"docs-recipe-behavior", "docs-recipe-visual", "docs-recipe-execution"} {
				if got := strings.Count(body, `id="`+id+`"`); got != 1 {
					t.Errorf("id %q occurs %d times, want 1", id, got)
				}
			}
			for _, label := range []string{
				`for="docs-recipe-behavior"`,
				`for="docs-recipe-visual"`,
				`for="docs-recipe-execution"`,
			} {
				if got := strings.Count(body, label); got != 1 {
					t.Errorf("label %q occurs %d times, want 1", label, got)
				}
			}
			if got := strings.Count(body, `name="behavior"`); got != 1 {
				t.Errorf("behavior selects = %d, want 1", got)
			}
			if got := strings.Count(body, `name="visual"`); got != 1 {
				t.Errorf("visual selects = %d, want 1", got)
			}
			if !strings.Contains(body, `<button`) || !strings.Contains(body, `type="submit"`) || !strings.Contains(body, `>Apply</span>`) {
				t.Error("full Recipe form must retain its native GET submit control")
			}
		})
	}
}

func TestNavigationSelectionIsSourceAware(t *testing.T) {
	tests := []struct {
		name, path, want string
	}{
		{
			name: "legacy theme and scheme stay legacy in ordinary docs",
			path: "/components/button?theme=basecoat&scheme=dark",
			want: "/components/icon-button?scheme=dark&theme=basecoat",
		},
		{
			name: "canonical selection follows ordinary docs navigation",
			path: "/components/button?behavior=material&reference=baseui&skin=vercel&execution=native",
			want: "/components/icon-button?behavior=material&execution=native&reference=baseui&skin=vercel",
		},
		{
			name: "canonical selection stays in the recipe component context",
			path: "/components/accordion?behavior=material&reference=baseui&skin=vercel&execution=native",
			want: "/components/button?behavior=material&execution=native&reference=baseui&skin=vercel",
		},
		{
			name: "canonical selection stays in the gallery context",
			path: "/docs/themes/gallery?behavior=baseui&reference=auto&skin=vercel&execution=native",
			want: "/docs/themes?behavior=baseui&execution=native&reference=baseui&skin=vercel",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if got := navigationSelectionFor(req, req.URL.Path).href(strings.Split(tt.want, "?")[0]); got != tt.want {
				t.Fatalf("navigation href = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultAndInvalidSelectionsKeepMaterialClass(t *testing.T) {
	for _, path := range []string{
		"/components/accordion",
		"/components/accordion?theme=not-a-theme",
		"/components/accordion?reference=invalid&skin=invalid",
	} {
		t.Run(path, func(t *testing.T) {
			res := httptest.NewRecorder()
			New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d", res.Code)
			}
			if !strings.Contains(res.Body.String(), `class="theme-material"`) {
				t.Fatalf("default or invalid selection must retain theme-material: %s", htmlClassSnippet(res.Body.String()))
			}
		})
	}
}

func TestDocumentSelectionRendersOnlySafeResolvedAttributes(t *testing.T) {
	server := New()
	for _, tt := range []struct{ path, want string }{
		{"/components/accordion?theme=material", `class="theme-material" data-gelium-reference="material" data-gelium-skin="none" data-gelium-contract="gelium" data-gelium-scheme="system"`},
		{"/components/accordion?behavior=material&reference=material&skin=basecoat&execution=native", `class="theme-basecoat" data-gelium-reference="material" data-gelium-skin="basecoat" data-gelium-contract="gelium" data-gelium-scheme="system"`},
		{"/components/accordion?behavior=baseui&reference=baseui&skin=vercel&execution=native", `class="theme-vercel" data-gelium-reference="baseui" data-gelium-skin="vercel" data-gelium-contract="gelium" data-gelium-scheme="system"`},
		{"/components/accordion?behavior=basecoat&reference=basecoat&skin=alden&execution=htmx&scheme=dark", `class="theme-alden theme-dark" data-gelium-reference="basecoat" data-gelium-skin="alden" data-gelium-contract="gelium" data-gelium-scheme="dark" data-theme="dark"`},
		{"/components/accordion?reference=%22%20onload%3Dalert(1)&skin=%3Cscript%3E&scheme=evil", `data-gelium-reference="none" data-gelium-skin="none" data-gelium-contract="gelium" data-gelium-scheme="system"`},
	} {
		t.Run(tt.path, func(t *testing.T) {
			res := httptest.NewRecorder()
			server.ServeHTTP(res, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if res.Code != http.StatusOK {
				t.Fatalf("status = %d", res.Code)
			}
			body := res.Body.String()
			if !strings.Contains(body, tt.want) {
				t.Fatalf("missing %q\n%s", tt.want, body[:min(len(body), 1000)])
			}
			for _, unsafe := range []string{"onload=alert", "&lt;script&gt;", "ZgotmplZ"} {
				if strings.Contains(body, unsafe) {
					t.Fatalf("unsafe query content leaked: %q", unsafe)
				}
			}
		})
	}
}

func TestExactAccordionSelectionHasMatchingCompiledCascade(t *testing.T) {
	const path = "/components/accordion?behavior=material&reference=material&skin=basecoat&execution=native"
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
	if res.Code != http.StatusOK {
		t.Fatalf("%s status = %d", path, res.Code)
	}
	if !strings.Contains(res.Body.String(), `data-gelium-reference="material" data-gelium-skin="basecoat"`) {
		t.Fatalf("exact selection did not render both root visual contracts:\n%s", res.Body.String()[:min(len(res.Body.String()), 1000)])
	}

	css, err := os.ReadFile(filepath.Join("..", "..", "site", "web", "static", "app.css"))
	if err != nil {
		t.Fatalf("read compiled app.css: %v", err)
	}
	bundle := strings.ReplaceAll(string(css), " ", "")
	for _, selector := range []string{`html[data-gelium-reference=material]{`, `html[data-gelium-skin=basecoat]{`} {
		if !strings.Contains(bundle, selector) {
			t.Fatalf("exact selection has no compiled adapter selector %q (run npm run build)", selector)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
