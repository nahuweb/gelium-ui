package app

import (
	"html/template"
	"testing"

	"geliumui/lib"

	webassets "geliumui/site/web"
)

// parseTestTemplates parses the given template patterns from BOTH embed
// filesystems (site shell + lib components). A glob like "templates/*.html"
// always parses both trees so the merged name set is the same as buildTemplates
// uses; an individual pattern resolves in whichever tree owns it, failing only
// if it exists in neither. This mirrors the dual buildTemplates() used by New().
func parseTestTemplates(t *testing.T, patterns ...string) *template.Template {
	t.Helper()
	// The layout <head> calls themePreloadFonts; the helper must register the
	// same func as buildTemplates() or ParseFS fails on an undefined function
	// and the layout template never enters the set.
	tmpl := template.New("test").Funcs(template.FuncMap{
		"htmlCloseTag":        htmlCloseTag,
		"htmlRootAttrs":       htmlRootAttrs,
		"renderRecipeOptions": renderRecipeOptions,
		"themePreloadFonts": func(themeClass, version string) template.HTML {
			return themePreloadFonts(themeClass, version)
		},
	})
	for _, pattern := range patterns {
		okSite := true
		if _, err := tmpl.ParseFS(webassets.Assets, pattern); err != nil {
			okSite = false
		}
		okLib := true
		if _, err := tmpl.ParseFS(lib.LibAssets, pattern); err != nil {
			okLib = false
		}
		if !okSite && !okLib {
			t.Fatalf("parse %s: not found in site/web or lib", pattern)
		}
	}
	return tmpl
}
