package web

import (
	"strings"
	"testing"
)

func TestMarketingShellUsesInlineStartMobileDrawer(t *testing.T) {
	css := strings.Join(strings.Fields(sourceComponentCSS(t, "docs-chrome.css")), " ")
	for _, contract := range []string{
		".site-header-mobile-menu { display: none; }",
		"@media (max-width: 48rem)",
		".site-header-primary { display: none; }",
		".site-header-mobile-menu { display: block;",
		"position: absolute;",
		"inset-inline-start: 0;",
		"width: min(20rem, calc(100vw - 2rem));",
	} {
		if !strings.Contains(css, contract) {
			t.Errorf("marketing shell is missing responsive drawer contract %q", contract)
		}
	}
}

// TestDocsShellSafeAreaAndNavHeightContract pins the Layer 3 (mobile/runtime)
// contracts that live in docs-shell.css and must reach the compiled bundle:
// the sticky topbar pads into the device safe areas (GOV.UK
// viewport-fit=cover pattern) so controls never sit under a notch, and the
// mobile nav disclosure caps its open height with dvh — never vh, which
// overflows the visible viewport on mobile browsers — scrolling internally
// (Base UI SideNavViewport / GOV.UK hidden-toggle pattern).
func TestDocsShellSafeAreaAndNavHeightContract(t *testing.T) {
	css := sourceComponentCSS(t, "docs-shell.css")
	for _, contract := range []string{
		"env(safe-area-inset-top)",
		"env(safe-area-inset-right)",
		"env(safe-area-inset-left)",
	} {
		if !strings.Contains(css, contract) {
			t.Errorf("docs-shell.css topbar is missing the safe-area contract %q", contract)
		}
	}

	start := strings.Index(css, ".docs-nav-mobile nav {")
	if start < 0 {
		t.Fatal("docs-shell.css must declare the .docs-nav-mobile nav rule")
	}
	navBlock := css[start:]
	if end := strings.Index(navBlock, "}"); end >= 0 {
		navBlock = navBlock[:end]
	}
	for _, contract := range []string{"100dvh", "overflow-y: auto"} {
		if !strings.Contains(navBlock, contract) {
			t.Errorf(".docs-nav-mobile nav is missing %q (dvh cap with internal scroll)", contract)
		}
	}
	if strings.Contains(navBlock, "vh") && !strings.Contains(navBlock, "dvh") {
		t.Error(".docs-nav-mobile nav must not keep a legacy vh cap (dvh only)")
	}

	// The served bundle carries both contracts (minifier keeps units/selectors).
	compiled := compiledAppCSS(t)
	for _, contract := range []string{"100dvh", "env(safe-area-inset-top)"} {
		if !strings.Contains(compiled, contract) {
			t.Errorf("compiled app.css is missing the mobile safe-area/nav contract %q", contract)
		}
	}
}

// TestViewTransitionFadeAndNoCrossDocumentAtRule pins the CSS side of the
// Layer 3 same-document view transitions: a default cross-fade for the root
// snapshot in base.css, and an absence proof that the cross-document
// @view-transition at-rule never ships — automatic cross-document navigation
// is still flag-gated in Firefox (Baseline 2025), so it stays same-document
// only. The prefers-reduced-motion gate lives in app.js; the media block is
// the CSS-side belt.
func TestViewTransitionFadeAndNoCrossDocumentAtRule(t *testing.T) {
	base := sourceComponentCSS(t, "base.css")
	for _, contract := range []string{
		"::view-transition-old(root)",
		"::view-transition-new(root)",
		"animation-duration",
	} {
		if !strings.Contains(base, contract) {
			t.Errorf("base.css is missing the view-transition cross-fade contract %q", contract)
		}
	}
	if strings.Contains(base, "@view-transition") {
		t.Error("base.css must never ship the cross-document @view-transition at-rule (same-document VT only)")
	}
	compiled := compiledAppCSS(t)
	if !strings.Contains(compiled, "::view-transition-old(root)") {
		t.Error("compiled app.css must carry the view-transition cross-fade")
	}
	if strings.Contains(compiled, "@view-transition") {
		t.Error("compiled app.css must never ship the cross-document @view-transition at-rule")
	}
}

// TestFocusRingVisibleInEveryTheme is the Layer 3 focus matrix (GOV.UK focus
// rule, WCAG 2.2 1.4.11): focus must stay visible on ANY background, in EVERY
// theme and in both schemes. The global :focus-visible rule ships in
// focus-ring.css and compiles into the served bundle; each theme must define
// the focus ring color in its light scheme AND in its explicit dark class
// route, plus the shared thickness/offset tokens. Presence only — never a
// concrete hex value. The test may read theme-basecoat but never modifies it.
func TestFocusRingVisibleInEveryTheme(t *testing.T) {
	focus := sourceComponentCSS(t, "focus-ring.css")
	if !strings.Contains(focus, ":focus-visible {") {
		t.Error("focus-ring.css must declare the global :focus-visible rule")
	}
	compiled := compiledAppCSS(t)
	if !strings.Contains(compiled, ":focus-visible") {
		t.Error("compiled app.css must carry a :focus-visible rule")
	}
	if strings.Contains(compiled, "@view-transition") {
		t.Error("compiled app.css must never ship the cross-document @view-transition at-rule")
	}

	for _, theme := range availableThemes(t) {
		t.Run(theme, func(t *testing.T) {
			light, darkClass, darkMedia := splitThemeSchemes(t, theme)
			if darkMedia != "" {
				t.Errorf("%s must not define focus values in a dark media block (single dark mechanism is the class route)", theme)
			}
			if !strings.Contains(light, "--ui-color-focus-ring:") {
				t.Errorf("%s light scheme must define --ui-color-focus-ring", theme)
			}
			if !strings.Contains(darkClass, "--ui-color-focus-ring:") {
				t.Errorf("%s dark class route must define --ui-color-focus-ring", theme)
			}
			css := themeCSS(t, theme)
			for _, token := range []string{"--ui-focus-thickness:", "--ui-focus-offset:"} {
				if !strings.Contains(css, token) {
					t.Errorf("%s must define %s (shared focus geometry token)", theme, token)
				}
			}
		})
	}
}
