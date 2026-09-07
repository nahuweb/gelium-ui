package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestDocsScreensPageRendersSourcedCriteria proves GET /docs/screens serves
// screen types, hierarchy, nav patterns, and cites external sources.
func TestDocsScreensPageRendersSourcedCriteria(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/docs/screens", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		">Screens</h1>",
		"Screen types",
		"Build checklist",
		"Primary action",
		"Nav decision cues",
		"GOV.UK",
		"USWDS",
		"design-system.service.gov.uk",
		"designsystem.digital.gov",
		"nngroup.com",
		`href="/docs/feedback"`,
		`href="/docs/journeys"`,
		`href="/llms-ux.txt"`,
		`href="/docs/screens"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("screens page missing %q", contract)
		}
	}
}

// TestDocsFeedbackPageRendersDecisionMatrix proves GET /docs/feedback serves
// the sourced feedback matrix (GOV.UK error summary / banner rules).
func TestDocsFeedbackPageRendersDecisionMatrix(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/docs/feedback", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	for _, contract := range []string{
		">Feedback</h1>",
		"Decision matrix",
		"error summary",
		"FEED-VAL",
		"FEED-PARTIAL",
		"Toast rules",
		"design-system.service.gov.uk",
		`href="/llms-ux.txt"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("feedback page missing %q", contract)
		}
	}
}

// TestDocsJourneysPageRendersShapes proves GET /docs/journeys serves journey
// shapes and post-submit landing rules from GOV.UK-adapted criteria.
func TestDocsJourneysPageRendersShapes(t *testing.T) {
	body := getOKBody(t, "/docs/journeys")
	for _, contract := range []string{
		">Journeys</h1>",
		"JOURNEY-LINEAR",
		"JOURNEY-TASKLIST",
		"After submit",
		"design-system.service.gov.uk",
		`href="/docs/feedback"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("journeys page missing %q", contract)
		}
	}
}

// TestDocsDataDisplayPageRendersPatterns proves GET /docs/data-display serves
// DATA-* when/when-not rules.
func TestDocsDataDisplayPageRendersPatterns(t *testing.T) {
	body := getOKBody(t, "/docs/data-display")
	for _, contract := range []string{
		">Data display</h1>",
		"DATA-TABLE",
		"DATA-CARDS",
		"FEED-EMPTY",
		"design-system.service.gov.uk",
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("data-display page missing %q", contract)
		}
	}
}

// TestDocsPatternsDomainSkeletons proves patterns is no longer a stub and
// includes forum/catalog skeletons plus recipe links.
func TestDocsPatternsDomainSkeletons(t *testing.T) {
	body := getOKBody(t, "/docs/patterns")
	for _, contract := range []string{
		">Patterns</h1>",
		"SKEL-FORUM",
		"SKEL-CATALOG",
		"SKEL-ADMIN-RESOURCE",
		`href="/recipes/admin-resource"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("patterns page missing %q", contract)
		}
	}
}

// TestDocsDensityMotionAndDoD proves supporting UX criteria pages render.
func TestDocsDensityMotionAndDoD(t *testing.T) {
	cases := map[string][]string{
		"/docs/density": {
			"Density and shell",
			"Density modes",
			"--ui-touch-target",
		},
		"/docs/motion": {
			">Motion</h1>",
			"prefers-reduced-motion",
			"MOTION-NONE",
		},
		"/docs/ui-definition-of-done": {
			"UI definition of done",
			"DoD checklist",
			"FEED-VAL",
		},
	}
	for path, contracts := range cases {
		body := getOKBody(t, path)
		for _, c := range contracts {
			if !strings.Contains(body, c) {
				t.Errorf("%s missing %q", path, c)
			}
		}
	}
}

// TestLlmsUXTxtServesAgentDecisionPack proves GET /llms-ux.txt serves dense
// agent tables linked from the human UX handbook.
func TestLlmsUXTxtServesAgentDecisionPack(t *testing.T) {
	res := httptest.NewRecorder()
	New().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/llms-ux.txt", nil))
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := res.Header().Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want plain UTF-8 text", got)
	}
	body := res.Body.String()
	for _, contract := range []string{
		"UX decision pack",
		"SURFACE",
		"WF-SHAPE",
		"Operate",
		"ANTI-SLOP",
		"SCREEN TYPE",
		"JOURNEY-LINEAR",
		"DATA-TABLE",
		"FEED-VAL",
		"SKEL-FORUM",
		"DOC-H1",
		"DOC-H2",
		"DOC-LIST",
		"DoD",
		"/docs/agent-workflow",
		"/docs/journeys",
		"/docs/content-style",
		"/docs/ui-definition-of-done",
		"REFERENCE-SELECTION",
		"semantic match",
		"component reference",
		"constant state",
		"visibility semantics",
		"first-plane priority",
		"progressive disclosure",
		"reference diversity",
		"PLAIN-LANGUAGE BRIEF",
		"Unknown",
		"open decisions",
	} {

		if !strings.Contains(body, contract) {
			t.Errorf("llms-ux.txt missing %q", contract)
		}
	}
}

// TestDocsAgentWorkflowEthosSafe proves agent workflow documents surface modes,
// WF passes, and Gelium-aware anti-slop without abandoning system ethos.
func TestDocsAgentWorkflowEthosSafe(t *testing.T) {
	body := getOKBody(t, "/docs/agent-workflow")
	for _, c := range []string{
		">Agent workflow</h1>",
		"Ethos",
		"Surface modes",
		"WF-SHAPE",
		"Operate",
		"Anti-slop",
		"ux-detect",
		"plain-language brief",
		`href="/docs/templates/product"`,
		`href="/docs/ui-definition-of-done"`,
	} {
		if !strings.Contains(body, c) {
			t.Errorf("agent-workflow missing %q", c)
		}
	}
}

// TestConsumerDesignTemplatesRender proves PRODUCT/DESIGN templates are served.
func TestConsumerDesignTemplatesRender(t *testing.T) {
	for _, path := range []string{"/docs/templates/product", "/docs/templates/design"} {
		body := getOKBody(t, path)
		if !strings.Contains(body, "gelium") && !strings.Contains(body, "Gelium") {
			t.Errorf("%s should mention Gelium", path)
		}
	}
	if body := getOKBody(t, "/docs/templates/product"); !strings.Contains(body, "PRODUCT.md") {
		t.Error("product template missing PRODUCT.md")
	}
	if body := getOKBody(t, "/docs/templates/design"); !strings.Contains(body, "theme-material") {
		t.Error("design template missing theme-material")
	}
	for _, contract := range []string{
		"plain language",
		"Unknown",
		"Success criteria",
		"theme-neubrutalism",
	} {
		if body := getOKBody(t, "/docs/templates/product"); !strings.Contains(body, contract) {
			t.Errorf("product template missing %q", contract)
		}
	}
	for _, contract := range []string{
		"Screen inventory",
		"Design intent",
		"Current-state audit",
		"Chrome and navigation",
		"Shell scope",
		"Narrow behavior",
		"Visual and interaction direction",
		"Media direction",
		"Anti-slop checklist",
		"theme-neubrutalism",
	} {
		if body := getOKBody(t, "/docs/templates/design"); !strings.Contains(body, contract) {
			t.Errorf("design template missing %q", contract)
		}
	}
}

// TestUxDetectScriptPasses proves deterministic detectors stay green on main.
func TestUxDetectScriptPasses(t *testing.T) {
	root := moduleRoot(t)
	cmd := exec.Command("bash", filepath.Join(root, "scripts/ux-detect.sh"))
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ux-detect.sh failed: %v\n%s", err, out)
	}
}

func TestShippedLegacyUxDetectKeepsPositionalResultContract(t *testing.T) {
	root := moduleRoot(t)
	consumer := t.TempDir()
	good := filepath.Join(consumer, "screen.html")
	if err := os.WriteFile(good, []byte(`<main><h1>Safe screen</h1></main>`), 0o644); err != nil {
		t.Fatal(err)
	}
	shipped := filepath.Join(root, "lib/scripts/ux-detect.sh")
	cmd := exec.Command("bash", shipped, good)
	cmd.Dir = consumer
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "== RESULT: PASSED ==") {
		t.Fatalf("legacy positional clean result changed: err=%v output=%s", err, out)
	}

	bad := filepath.Join(consumer, "form.html")
	if err := os.WriteFile(bad, []byte(`<form><input name="email"></form>`), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd = exec.Command("bash", shipped, bad)
	cmd.Dir = consumer
	out, err = cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "form without validation-summary contract") || !strings.Contains(string(out), "== RESULT: FAILED ==") {
		t.Fatalf("legacy positional failure wording changed: err=%v output=%s", err, out)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found from test wd")
		}
		dir = parent
	}
}
