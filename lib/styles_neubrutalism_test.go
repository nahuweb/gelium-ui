package lib

import (
	"regexp"
	"strings"
	"testing"
)

func singleSpaceCSS(css string) string {
	return regexp.MustCompile(`\s+`).ReplaceAllString(css, " ")
}

func TestNeubrutalismDepthTokensCoverPrimarySurfaces(t *testing.T) {
	css := singleSpaceCSS(themeCSS(t, "theme-neubrutalism"))
	for _, contract := range []string{
		`--ui-dialog-border: var(--ui-border-width-1) var(--ui-border-style-solid) var(--ui-color-border);`,
		`--ui-dialog-shadow: var(--ui-shadow-3);`,
		`--ui-card-elevated-border: var(--ui-border-width-1) var(--ui-border-style-solid) var(--ui-color-border);`,
		`--ui-card-elevated-shadow: var(--ui-shadow-2);`,
		`--ui-card-filled-border: var(--ui-border-width-1) var(--ui-border-style-solid) var(--ui-color-border);`,
		`--ui-card-filled-shadow: var(--ui-shadow-1);`,
		`--ui-card-outlined-shadow: var(--ui-shadow-2);`,
		`--ui-accordion-shadow: var(--ui-shadow-2);`,
		`--ui-shadow-md: var(--ui-shadow-2);`,
		`--ui-toast-border: var(--ui-border-width-1) var(--ui-border-style-solid) var(--ui-color-border);`,
		`--ui-icon-button-border: var(--ui-border-width-1) var(--ui-border-style-solid) var(--ui-color-border);`,
		`--ui-icon-button-shadow: var(--ui-shadow-1);`,
		`--ui-icon-button-hover-shadow: none;`,
		`--ui-icon-button-hover-transform: translate(3px, 3px);`,
		`--ui-icon-button-active-transform: translate(4px, 4px);`,
	} {
		if !strings.Contains(css, contract) {
			t.Errorf("Neubrutalism depth contract is missing %q", contract)
		}
	}
}

func TestNeubrutalismTypographyUsesHeavyHierarchy(t *testing.T) {
	css := themeCSS(t, "theme-neubrutalism")
	for _, contract := range []string{
		`--ui-type-display-lg-weight: 700;`,
		`--ui-type-display-sm-weight: 700;`,
		`--ui-type-headline-sm-weight: 700;`,
		`--ui-type-title-lg-weight: 700;`,
		`--ui-type-title-md-weight: 700;`,
		`--ui-type-body-lg-weight: 500;`,
		`--ui-type-body-md-weight: 500;`,
		`--ui-type-body-sm-weight: 500;`,
		`--ui-type-dialog-headline-weight: 700;`,
		`--ui-type-dialog-body-weight: 500;`,
	} {
		if !strings.Contains(css, contract) {
			t.Errorf("Neubrutalism typography contract is missing %q", contract)
		}
	}
}

func TestDepthBearingComponentsConsumeThemeTokens(t *testing.T) {
	contracts := map[string][]string{
		"dialog.css": {
			`border: var(--ui-dialog-border, 0);`,
			`box-shadow: var(--ui-dialog-shadow, none);`,
			`.ui-dialog-page {`,
		},
		"card.css": {
			`.ui-card-elevated { background: var(--ui-card-container-elevated); border: var(--ui-card-elevated-border, 0); box-shadow: var(--ui-card-elevated-shadow, var(--ui-shadow-1)); }`,
			`.ui-card-filled { background: var(--ui-card-container-filled); border: var(--ui-card-filled-border, 0); box-shadow: var(--ui-card-filled-shadow, none); }`,
			`.ui-card-outlined { background: var(--ui-card-container-outlined); border: var(--ui-card-outlined-border, var(--ui-border-width-1) var(--ui-border-style-solid) var(--ui-card-outline-color)); box-shadow: var(--ui-card-outlined-shadow, none); }`,
		},
		"toast.css": {
			`border: var(--ui-toast-border, 1px solid var(--ui-color-border-strong));`,
		},
		"icon-button.css": {
			`border: var(--ui-icon-button-border, var(--ui-border-width-1) var(--ui-border-style-solid) transparent);`,
			`box-shadow: var(--ui-icon-button-shadow, none);`,
			`transform: var(--ui-icon-button-transform, none);`,
			`transform: var(--ui-icon-button-hover-transform, var(--ui-icon-button-transform, none));`,
			`transform: var(--ui-icon-button-active-transform, var(--ui-icon-button-hover-transform, var(--ui-icon-button-transform, none)));`,
			`@media (prefers-reduced-motion: reduce) { .ui-icon-button { transition: none; transform: none; } }`,
		},
	}
	for file, expected := range contracts {
		css := singleSpaceCSS(readSourceStyle(t, file))
		for _, contract := range expected {
			if !strings.Contains(css, contract) {
				t.Errorf("%s is missing %q", file, contract)
			}
		}
	}
}
