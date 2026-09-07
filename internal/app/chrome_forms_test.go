package app

import (
	"strings"
	"testing"
)

// TestChromeFormViewModels proves the server-side switcher models drive the
// native controls: theme options carry the public ?theme= slugs (Material →
// theme-material, Basecoat → theme-basecoat) with the current one selected,
// and the scheme switch maps checked → dark / unchecked → light while each
// form preserves the OTHER chrome parameter via hidden inputs.
func TestChromeFormViewModels(t *testing.T) {
	t.Run("theme select defaults to material", func(t *testing.T) {
		sw := themeSwitcherFor(nil, "theme-material", "", "")
		if sw.Label != "Skin" {
			t.Errorf("Label = %q, want Skin", sw.Label)
		}
		want := []themeOptionView{
			{Label: "Material", Value: "material", Class: "theme-material", Selected: true},
			{Label: "Basecoat", Value: "basecoat", Class: "theme-basecoat", Selected: false},
			{Label: "Linear", Value: "linear", Class: "theme-linear", Selected: false},
			{Label: "Vercel", Value: "vercel", Class: "theme-vercel", Selected: false},
			{Label: "Base UI-inspired", Value: "baseui", Class: "theme-baseui", Selected: false},
			{Label: "Neubrutalism", Value: "neubrutalism", Class: "theme-neubrutalism", Selected: false},
			{Label: "Alden", Value: "alden", Class: "theme-alden", Selected: false},
		}
		if len(sw.Options) != len(want) {
			t.Fatalf("options = %d, want %d", len(sw.Options), len(want))
		}
		for i, opt := range want {
			if sw.Options[i] != opt {
				t.Errorf("options[%d] = %+v, want %+v", i, sw.Options[i], opt)
			}
		}
		if sw.Scheme != "" {
			t.Errorf("Scheme = %q, want empty when no scheme set", sw.Scheme)
		}
	})

	t.Run("theme select basecoat selected and scheme preserved", func(t *testing.T) {
		// Mirrors the server call site: the request theme is already applied to
		// data.ThemeClass by the render path, so the builder receives the
		// resolved class plus the slug and scheme.
		sw := themeSwitcherFor(nil, "theme-basecoat", "basecoat", "dark")
		for _, opt := range sw.Options {
			if opt.Value == "basecoat" && !opt.Selected {
				t.Error("basecoat must be selected under ?theme=basecoat")
			}
			if opt.Value == "material" && opt.Selected {
				t.Error("material must not be selected under ?theme=basecoat")
			}
		}
		if sw.Scheme != "dark" {
			t.Errorf("Scheme = %q, want dark (hidden input preserves it)", sw.Scheme)
		}
	})

	t.Run("scheme switch checked maps to dark", func(t *testing.T) {
		sw := schemeSwitcherFor(nil, "basecoat", "dark")
		if sw.Label != "Appearance" {
			t.Errorf("Label = %q, want Appearance", sw.Label)
		}
		if !sw.Checked {
			t.Error("switch must be checked under ?scheme=dark")
		}
		if sw.Theme != "basecoat" {
			t.Errorf("Theme = %q, want basecoat (hidden input preserves it)", sw.Theme)
		}
	})

	t.Run("scheme switch unchecked maps to light", func(t *testing.T) {
		sw := schemeSwitcherFor(nil, "", "")
		if sw.Checked {
			t.Error("switch must be unchecked by default (light)")
		}
	})
}

// TestThemeSelectValuesProduceSameChromeURLs proves the native select values
// build the exact ?theme=<slug> URLs the old link-list switcher produced, with
// the same closed-vocabulary preservation (theme + scheme only).
func TestThemeSelectValuesProduceSameChromeURLs(t *testing.T) {
	for _, tt := range []struct {
		slug   string
		scheme string
		want   string
	}{
		{slug: "material", scheme: "", want: "/docs?theme=material"},
		{slug: "basecoat", scheme: "", want: "/docs?theme=basecoat"},
		{slug: "material", scheme: "dark", want: "/docs?scheme=dark&theme=material"},
	} {
		if got := chromeHref("/docs", tt.slug, tt.scheme); got != tt.want {
			t.Errorf("chromeHref(%q, %q) = %q, want %q", tt.slug, tt.scheme, got, tt.want)
		}
	}
}

// TestChromeFormsPreserveEachOther proves the rendered forms carry the hidden
// cross-parameter inputs: the theme form keeps ?scheme=, the scheme form keeps
// ?theme=, so a GET submission never silently resets direction or light/dark.
func TestChromeFormsPreserveEachOther(t *testing.T) {
	body := getOKBody(t, "/docs?theme=basecoat&scheme=dark")
	if !strings.Contains(body, `<input type="hidden" name="scheme" value="dark">`) {
		t.Error("theme form must carry hidden scheme=dark to preserve light/dark")
	}
	if !strings.Contains(body, `<input type="hidden" name="theme" value="basecoat">`) {
		t.Error("scheme form must carry hidden theme=basecoat to preserve direction")
	}
	// The switch reflects the dark state.
	if !strings.Contains(body, `role="switch" name="scheme" value="dark" checked`) {
		t.Error("scheme switch must render checked under ?scheme=dark")
	}
}

// TestChromeFormsRenderSubmitButtons keeps the native GET fallback in the
// DOM while the compact header hides its submit affordance after enhancement.
func TestChromeFormsRenderSubmitButtons(t *testing.T) {
	body := getOKBody(t, "/docs")
	if got := strings.Count(body, `<button type="submit"`); got != 1 {
		t.Errorf("chrome submit buttons = %d, want 1 (Appearance fallback)", got)
	}
	for _, contract := range []string{
		`<form class="ui-recipe-switcher ui-recipe-switcher--compact" method="get" data-chrome-form hx-boost="false" aria-label="Component recipe">`,
		`<form class="ui-theme-switcher ui-scheme-switcher" method="get" data-chrome-form hx-swap="none" aria-label="Appearance">`,
		`<label class="ui-select-label" for="docs-recipe-behavior">Behavior</label>`,
		`<label class="ui-select-label" for="docs-recipe-visual">Visual recipe</label>`,
		`ui-select ui-select-outlined`,
		`ui-switch`,
		`name="behavior"`, `name="visual"`, `value="basecoat"`, `Apply`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("recipe form missing %q", contract)
		}
	}
}

// TestChromeFormsAlwaysCarryPreserveInputs proves both chrome forms render
// their cross-parameter hidden preserve input ALWAYS (even empty). With
// hx-swap=none the body never re-renders, so a conditional input would stay
// stale after an optimistic toggle and the next submission would silently
// forget the other parameter (theme form forgetting scheme, or vice versa).
func TestChromeFormsAlwaysCarryPreserveInputs(t *testing.T) {
	body := getOKBody(t, "/docs")
	if !strings.Contains(body, `<input type="hidden" name="scheme" value="light">`) {
		t.Error("recipe form must default its hidden scheme preserve input to light")
	}
	if !strings.Contains(body, `<input type="hidden" name="theme" value="">`) {
		t.Error("scheme form must always render its hidden theme preserve input (empty value allowed)")
	}
}

// TestChromeIconsAreInlineSvgWithoutLibrary proves the chrome icons are plain
// inline SVG (no icon-library asset, no network dependency) and decorative:
// aria-hidden + focusable=false. The scheme form carries the sun/moon pair so
// CSS toggles which one is visible from the server-driven .theme-dark class.
func TestChromeIconsAreInlineSvgWithoutLibrary(t *testing.T) {
	body := getOKBody(t, "/docs?theme=basecoat&scheme=dark")
	for _, contract := range []string{
		`<svg class="ui-scheme-icon ui-scheme-icon-sun"`,
		`<svg class="ui-scheme-icon ui-scheme-icon-moon"`,
		`aria-hidden="true" focusable="false"`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("chrome icons are missing contract %q", contract)
		}
	}
}

// TestRecipeFormCarriesOnlyClosedValues proves the new selector stays native
// GET, offers only allowlisted values, and has no optimistic custom-JS contract.
func TestRecipeFormCarriesOnlyClosedValues(t *testing.T) {
	body := getOKBody(t, "/components/accordion?theme=basecoat&scheme=dark")
	for _, contract := range []string{
		`name="behavior"`, `value="native"`, `value="material"`, `value="baseui"`,
		`name="visual"`, `value="default"`, `value="basecoat"`, `value="vercel"`,
		`<input type="hidden" name="scheme" value="dark">`,
	} {
		if !strings.Contains(body, contract) {
			t.Errorf("recipe form missing closed-value contract %q", contract)
		}
	}
	start := strings.Index(body, `class="ui-recipe-switcher`)
	if start < 0 {
		t.Fatal("missing recipe form")
	}
	end := strings.Index(body[start:], "</form>")
	if end < 0 {
		t.Fatal("unclosed recipe form")
	}
	recipe := body[start : start+end]
	if strings.Contains(recipe, `hx-swap="none"`) || strings.Contains(recipe, `data-class=`) {
		t.Error("native Recipe form must not depend on optimistic or class-mapping JavaScript")
	}
}
