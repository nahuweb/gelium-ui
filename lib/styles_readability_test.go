package lib

// Prose readability contract tests (readability-improvements).
//
// The docs text column is the primary reading surface of the Gelium UI
// documentation, so its typography is a contract, not a taste:
//
//  1. Measure — .prose caps line length at 65ch (the 65-75ch typographic
//     standard; USWDS uses ~68ch, GOV.UK ~30em). ch is the correct unit for
//     measure because it scales with the rendered font: the column always
//     carries roughly the same number of characters regardless of type size.
//  2. Text wrapping — body copy uses text-wrap: pretty (avoids widows and
//     orphaned final lines), headings use text-wrap: balance (even line
//     breaks on short heading blocks). Both are Baseline 2024.
//  3. Hyphenation — .prose sets hyphens: auto and inherits the document
//     language from <html lang>, so justified-ish narrow columns hyphenate
//     per language instead of overflowing.
//  4. Progressive leading trim — headings set text-box-trim: trim-both with
//     text-box-edge: cap alphabetic where supported (Chromium); unsupported
//     browsers ignore the pair, so this is a pure enhancement.
//  5. Readable line height — prose body copy keeps line-height >= 1.6.

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// readSourceStyle reads one file from the embedded web/styles source tree.
func readSourceStyle(t *testing.T, name string) string {
	t.Helper()
	css, err := sourceStyles.ReadFile("styles/" + name)
	if err == nil {
		return string(css)
	}
	// docs-shell.css is site-owned; lib readability tests read it from the repo.
	return repositoryFile(t, "site", "web", "styles", name)
}

// cssBlock returns the rule block for an exact selector prefix (e.g. ".prose"
// or ".prose p"), from the opening brace to the matching close brace. The
// style files keep these rules flat (no nested braces), so brace counting is
// sufficient.
func cssBlock(t *testing.T, css, selector string) string {
	t.Helper()
	idx := strings.Index(css, selector+" {")
	if idx < 0 {
		// Formatting may place comma-separated selectors on separate lines.
		pattern := regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(selector), " ", `\\s+`) + `\\s*\\{`)
		loc := pattern.FindStringIndex(css)
		if loc != nil {
			idx = loc[0]
		}
	}
	if idx < 0 {
		t.Fatalf("CSS is missing rule %q", selector)
	}
	brace := strings.Index(css[idx:], "{")
	depth := 0
	for i := idx + brace; i < len(css); i++ {
		switch css[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return css[idx : i+1]
			}
		}
	}
	t.Fatalf("rule %q is unterminated", selector)
	return ""
}

// TestProseMeasureUsesChUnits is the readability contract: the prose text
// column must cap line length at 65ch, and must not fall back to a rem-based
// measure (the old 48rem ≈ 90 chars at the body size — too wide for comfort).
func TestProseMeasureUsesChUnits(t *testing.T) {
	prose := cssBlock(t, readSourceStyle(t, "docs-chrome.css"), ".prose")
	if !strings.Contains(prose, "max-width: 65ch") {
		t.Errorf(".prose must cap line length at 65ch, got block: %s", prose)
	}
	if strings.Contains(prose, "48rem") {
		t.Errorf(".prose must use a ch measure, not rem: %s", prose)
	}
}

// TestProseBodyTextWrapPretty is the widows/orphans contract: body copy must
// set text-wrap: pretty so the final line of a paragraph never dangles alone.
func TestProseBodyTextWrapPretty(t *testing.T) {
	p := cssBlock(t, readSourceStyle(t, "docs-chrome.css"), ".prose p")
	if !strings.Contains(p, "text-wrap: pretty") {
		t.Errorf(".prose p must set text-wrap: pretty, got block: %s", p)
	}
}

// TestProseBodyLineHeightStaysReadable pins the readable line-height floor:
// prose body copy must keep line-height >= 1.6 (current: 1.7).
func TestProseBodyLineHeightStaysReadable(t *testing.T) {
	p := cssBlock(t, readSourceStyle(t, "docs-chrome.css"), ".prose p")
	m := regexp.MustCompile(`line-height:\s*([0-9]+(?:\.[0-9]+)?)`).FindStringSubmatch(p)
	if m == nil {
		t.Fatalf(".prose p must set a numeric line-height >= 1.6, got block: %s", p)
	}
	lh, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		t.Fatalf("parse line-height %q: %v", m[1], err)
	}
	if lh < 1.6 {
		t.Errorf(".prose p line-height = %v, want >= 1.6", lh)
	}
}

// TestProseHeadingsTextWrapBalance is the balanced-headings contract: h1-h3
// must set text-wrap: balance so multi-line headings break evenly.
func TestProseHeadingsTextWrapBalance(t *testing.T) {
	heads := cssBlock(t, strings.ReplaceAll(singleSpaceCSS(readSourceStyle(t, "docs-chrome.css")), "0.", "."), ".prose h1, .prose h2, .prose h3")
	if !strings.Contains(heads, "text-wrap: balance") {
		t.Errorf(".prose h1/h2/h3 must set text-wrap: balance, got block: %s", heads)
	}
}

// TestProseHyphensAuto is the hyphenation contract: .prose must set
// hyphens: auto and inherit the document language from <html lang>.
func TestProseHyphensAuto(t *testing.T) {
	prose := cssBlock(t, readSourceStyle(t, "docs-chrome.css"), ".prose")
	if !strings.Contains(prose, "hyphens: auto") {
		t.Errorf(".prose must set hyphens: auto, got block: %s", prose)
	}
}

// TestProseHeadingTextBoxTrimProgressive is the progressive-enhancement
// contract: headings trim visual leading with text-box-trim/text-box-edge
// where supported (Chromium only for now). The pair must be declared on the
// heading rule so a future browser ships it automatically.
func TestProseHeadingTextBoxTrimProgressive(t *testing.T) {
	heads := cssBlock(t, strings.ReplaceAll(singleSpaceCSS(readSourceStyle(t, "docs-chrome.css")), "0.", "."), ".prose h1, .prose h2, .prose h3")
	if !strings.Contains(heads, "text-box-trim: trim-both") {
		t.Errorf(".prose h1/h2/h3 must set text-box-trim: trim-both (progressive), got block: %s", heads)
	}
	if !strings.Contains(heads, "text-box-edge: cap alphabetic") {
		t.Errorf(".prose h1/h2/h3 must set text-box-edge: cap alphabetic (progressive), got block: %s", heads)
	}
}

// TestProseVerticalRhythm is the vertical-rhythm contract: a label line
// (breadcrumb, provenance) or a lower heading must never collide with the
// next heading. h2/h3/h4 carry top AND bottom margins; the provenance line
// has its own label style with breathing room before the title.
func TestProseVerticalRhythm(t *testing.T) {
	css := strings.ReplaceAll(singleSpaceCSS(readSourceStyle(t, "docs-chrome.css")), "0.", ".")

	h2 := cssBlock(t, css, ".prose h2")
	if !strings.Contains(h2, "margin: 2.5rem 0 1rem") {
		t.Errorf(".prose h2 must carry top+bottom margins (2.5rem 0 1rem), got block: %s", h2)
	}
	h3 := cssBlock(t, css, ".prose h3")
	if !strings.Contains(h3, "margin: 1.75rem 0 .75rem") {
		t.Errorf(".prose h3 must carry top+bottom margins (1.75rem 0 .75rem), got block: %s", h3)
	}
	// The provenance line must be label-styled, not body copy, and spaced
	// from the title below it.
	prov := cssBlock(t, css, ".article-provenance")
	if !strings.Contains(prov, "font: var(--ui-type-label-md)") {
		t.Errorf(".article-provenance must use label type, got block: %s", prov)
	}
	if !strings.Contains(prov, "margin: 0 0 1.25rem") {
		t.Errorf(".article-provenance must have bottom breathing room (0 0 1.25rem), got block: %s", prov)
	}
}

// TestBreadcrumbClearsPageTitle is the shell rhythm contract: the breadcrumb
// nav above a page title must not collide with the h1.
func TestBreadcrumbClearsPageTitle(t *testing.T) {
	shell := readSourceStyle(t, "docs-shell.css")
	idx := strings.Index(shell, `nav[aria-label="Breadcrumb"]`)
	if idx < 0 {
		t.Fatal("docs-shell.css must style the breadcrumb nav with bottom margin")
	}
	block := shell[idx:]
	end := strings.Index(block, "}")
	if end < 0 {
		t.Fatal("breadcrumb rule is unterminated")
	}
	rule := block[:end+1]
	if !strings.Contains(rule, "margin-bottom: 1rem") {
		t.Errorf("breadcrumb nav must have margin-bottom: 1rem, got rule: %s", rule)
	}
}

// TestProseTablesRender is the GFM-tables contract: markdown tables in the
// prose must be styled as real tables (borders, header emphasis, mobile
// scroll) — they were previously unstyled after GFM was enabled.
func TestProseTablesRender(t *testing.T) {
	css := strings.ReplaceAll(singleSpaceCSS(readSourceStyle(t, "docs-chrome.css")), "0.", ".")
	table := cssBlock(t, css, ".prose table")
	if !strings.Contains(table, "border-collapse: collapse") {
		t.Errorf(".prose table must collapse borders, got block: %s", table)
	}
	cells := cssBlock(t, css, ".prose th, .prose td")
	if !strings.Contains(cells, "border-bottom: 1px solid var(--ui-color-border)") {
		t.Errorf(".prose th/td must carry a token border, got block: %s", cells)
	}
	if !strings.Contains(cells, "padding: .5rem .75rem") {
		t.Errorf(".prose th/td must pad cells, got block: %s", cells)
	}
}

// TestProseCodeBlocksRender is the fenced-code-block contract: markdown
// <pre> blocks must be styled (border, surface-container, mono) and carry
// chroma syntax-highlighting classes driven by theme tokens.
func TestProseCodeBlocksRender(t *testing.T) {
	css := singleSpaceCSS(readSourceStyle(t, "docs-chrome.css"))
	pre := cssBlock(t, css, ".prose pre")
	if !strings.Contains(pre, "background: var(--ui-color-surface-container)") {
		t.Errorf(".prose pre must use the surface-container token, got block: %s", pre)
	}
	if !strings.Contains(pre, "font-family: var(--ui-font-mono") {
		t.Errorf(".prose pre must use the mono font, got block: %s", pre)
	}
	// Chroma selectors are comma-grouped, so assert by substring instead of
	// a standalone block.
	if !strings.Contains(css, ".prose .chroma .k, ") && !strings.Contains(css, ".prose .chroma .k ") {
		t.Errorf("chroma keyword selectors must exist, got css without .chroma .k")
	}
	if !strings.Contains(css, ".prose .chroma .k") || !strings.Contains(css, "var(--ui-color-primary)") {
		t.Errorf("chroma keywords must map to the primary token")
	}
	if !strings.Contains(css, ".prose .chroma .c,") && !strings.Contains(css, ".prose .chroma .c ") {
		t.Errorf("chroma comment selectors must exist, got css without .chroma .c")
	}
	if !strings.Contains(css, ".prose .chroma .c") || !strings.Contains(css, "var(--ui-color-fg-muted)") {
		t.Errorf("chroma comments must map to the muted token")
	}
}
