package app

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"

	webassets "geliumui/site/web"

	"geliumui/lib"
)

// defaultThemeClass is the theme applied when none is requested. The value must
// match a class owned by a theme that ships on disk (lib/themes/<name>.css).
const defaultThemeClass = "theme-material"

// themeDirection is one product-facing visual direction in the catalog.
type themeDirection struct {
	Class string
	Slug  string
	Label string
	// Fonts are the self-hosted webfonts this theme ships, emitted as
	// <link rel="preload"> in the document <head> so the browser discovers
	// them early (the @font-face src alone would fetch them too late,
	// causing FOUT + CLS). Relative to /static/fonts/<theme>/.
	Fonts []themeFont
}

// themeFont is one self-hosted webfont file a theme ships (served from
// /static/fonts/). File is the flat idempotent filename (e.g.
// "theme-alden-inter-400-latin.woff2"); Name is the display family label.
// Preload marks files the <head> should preload: only the above-the-fold body
// weight(s) — never every weight — so preloads don't compete with the LCP
// resource (contract §3.A.4 + §B.7, caps preloads). A font referenced by the
// theme's @font-face but not preloaded is still servable: it loads on demand
// via font-display: swap.
type themeFont struct {
	// File is the flat font filename served under /static/fonts/. Globally
	// idempotent across themes via the <theme-class>-<file> convention.
	File string
	// Name is the display family label (e.g. "Alden Sans").
	Name string
	// Preload true → layout.html emits this font in <head> preload links.
	Preload bool
}

// fontPreloadHref returns the absolute static path for a theme font file.
func (f themeFont) preloadHref() string {
	return "/static/fonts/" + f.File
}

// themeFontsFor returns the resolved theme's font set by theme class, or nil
// if the class is not allowlisted (falls back to the default theme's fonts,
// if any).
func themeFontsFor(themeClass string) []themeFont {
	for _, t := range availableThemes {
		if themeClass == t.Class {
			return t.Fonts
		}
	}
	for _, t := range availableThemes {
		if defaultThemeClass == t.Class {
			return t.Fonts
		}
	}
	return nil
}

// themePreloadFonts renders the <link rel="preload"> markup for the resolved
// theme class, or the empty string when the theme ships no fonts. Used by the
// layout <head> via the template func of the same name. The href is
// cache-busted by AssetsVersion at render time by the caller (layout passes
// the version through the func's second argument).
func themePreloadFonts(themeClass string, version string) template.HTML {
	fonts := themeFontsFor(themeClass)
	if len(fonts) == 0 {
		return ""
	}
	var b bytes.Buffer
	for _, f := range fonts {
		if !f.Preload {
			continue
		}
		fmt.Fprintf(&b,
			`<link rel="preload" as="font" href="%s?v=%s" type="font/woff2" crossorigin>`,
			f.preloadHref(), template.HTMLEscapeString(version))
	}
	return template.HTML(b.String())
}

// availableThemes is the product catalog of visual directions. Order is the
// switcher order. Adding a preset is one new row + theme file + app.css import.
var availableThemes = []themeDirection{
	{Class: "theme-material", Slug: "material", Label: "Material"},
	{Class: "theme-basecoat", Slug: "basecoat", Label: "Basecoat"},
	{Class: "theme-linear", Slug: "linear", Label: "Linear"},
	{Class: "theme-vercel", Slug: "vercel", Label: "Vercel"},
	{Class: "theme-baseui", Slug: "baseui", Label: "Base UI-inspired"},
	{Class: "theme-neubrutalism", Slug: "neubrutalism", Label: "Neubrutalism"},
	{
		Class: "theme-alden", Slug: "alden", Label: "Alden",
		// Self-hosted fonts shipped in lib/fonts/ (WOFF2, subsetted latin +
		// latin-ext). The full set declared here is servable (the theme's
		// @font-face references all of them). Only the body weight is
		// preloaded (Preload: true) to avoid competing with LCP; the serif
		// display and heavier weights load on demand via font-display: swap.
		Fonts: []themeFont{
			{File: "theme-alden-inter-400-latin.woff2", Name: "Inter 400 latin", Preload: true},
			{File: "theme-alden-inter-400-latin-ext.woff2", Name: "Inter 400 latin-ext", Preload: true},
			{File: "theme-alden-inter-500-latin.woff2", Name: "Inter 500 latin"},
			{File: "theme-alden-inter-500-latin-ext.woff2", Name: "Inter 500 latin-ext"},
			{File: "theme-alden-inter-600-latin.woff2", Name: "Inter 600 latin"},
			{File: "theme-alden-inter-600-latin-ext.woff2", Name: "Inter 600 latin-ext"},
			{File: "theme-alden-source-serif-4-400-latin.woff2", Name: "Source Serif 4 latin"},
			{File: "theme-alden-source-serif-4-400-latin-ext.woff2", Name: "Source Serif 4 latin-ext"},
		},
	},
}

// themeOptionView is one <option> in the native theme direction select.
type themeOptionView struct {
	Label    string
	Value    string // public ?theme= slug the select submits: material, basecoat
	Class    string // allowlisted theme class emitted as data-class for optimistic JS
	Selected bool
}

// themeSwitcherView is retained only for backwards-compatible test and
// integration callers. The product UI uses recipeSwitcherView below; legacy
// ?theme= links remain accepted by resolveDocumentSelection.
type themeSwitcherView struct {
	Label   string // accessible product label: "Skin" (name=theme remains compatible)
	Options []themeOptionView
	Scheme  string
	// Resolved selection is surfaced as read-only metadata while legacy ?theme=
	// remains the native compatibility control during Phase 1.
	Reference string
	Skin      string
}

type profileOptionView struct {
	Label    string
	Value    string
	Selected bool
}

// recipeOptionView is one safe visual recipe. A recipe maps the one visible
// visual select back to the reference/skin architecture on the server; users
// never have to combine low-level selectors themselves.
type recipeOptionView struct {
	Label    string
	Value    string
	Selected bool
}

// recipeSwitcherView is the native, no-JS Recipe form. Visual is deliberately
// a single control even though its resolved result has independent Reference
// and Skin concerns. A valid visual submission is canonicalized by the server
// to behavior/reference/skin/scheme/execution (and never emits legacy theme).
type recipeSwitcherView struct {
	ID      string
	Compact bool
	// RenderInTopbar is the single placement decision: compact selectors belong
	// to chrome; full selectors belong once in the primary article.
	RenderInTopbar bool
	Behavior       string
	Visual         string
	Scheme         string
	Execution      string
	Contract       string
	Behaviors      []profileOptionView
	Visuals        []recipeOptionView
	Executions     []profileOptionView
	Contracts      []profileOptionView
	Summary        string
}

// schemeSwitcherView is server-driven chrome for the light/dark control: a
// native switch (<input type="checkbox" role="switch">) inside a 0-JS GET
// form. Checked maps to ?scheme=dark; unchecked to ?scheme=light. Theme
// carries the current ?theme= slug so a hidden input preserves direction.
type schemeSwitcherView struct {
	Label     string // accessible product label: "Appearance"
	Checked   bool
	Selection documentSelection
	Execution accordionExecution
	// Theme preserves the public legacy appearance form contract when the
	// request arrived through ?theme=. Canonical is only true in recipe/gallery
	// contexts, where appearance must retain the resolved recipe instead.
	Theme     string
	Canonical bool
}

// themeContextKey carries the theme selected via the ?theme= query parameter
// through the request context (Phase H: selection from the document root, no JS).
type themeContextKey struct{}

// schemeContextKey carries the color scheme selected via ?scheme=light|dark.
type schemeContextKey struct{}
type referenceContextKey struct{}
type skinContextKey struct{}
type contractContextKey struct{}

// canonicalSelectionContextKey preserves whether this request arrived through
// the explicit Recipe/canonical contract. Native + none + none is neutral only
// when selected explicitly; ordinary pages retain the legacy Material fallback.
type canonicalSelectionContextKey struct{}

// referencePreset is Gelium's visual baseline. It is deliberately separate
// from behavior: it supplies only token direction, never interaction policy.
type referencePreset string

const (
	referenceNone     referencePreset = "none"
	referenceMaterial referencePreset = "material"
	referenceBasecoat referencePreset = "basecoat"
	referenceBaseUI   referencePreset = "baseui"
)

// productSkin is a product-owned visual overlay. `none` intentionally means
// no overlay; it is an allowlisted value, not an omitted/untrusted attribute.
type productSkin string

const (
	skinNone         productSkin = "none"
	skinMaterial     productSkin = "material"
	skinBasecoat     productSkin = "basecoat" // Vega default pack
	skinBasecoatNova productSkin = "basecoat-nova"
	skinBasecoatMaia productSkin = "basecoat-maia"
	skinBasecoatLyra productSkin = "basecoat-lyra"
	skinBasecoatMira productSkin = "basecoat-mira"
	skinBasecoatLuma productSkin = "basecoat-luma"
	skinBasecoatSera productSkin = "basecoat-sera"
	skinBasecoatRhea productSkin = "basecoat-rhea"
	skinBaseUI       productSkin = "baseui"
	skinAlden        productSkin = "alden"
	skinLinear       productSkin = "linear"
	skinVercel       productSkin = "vercel"
	skinNeubrutalism productSkin = "neubrutalism"
)

type documentSelection struct {
	Behavior  accordionBehavior
	Reference referencePreset
	Skin      productSkin
	Scheme    string
	Contract  selectionContract
	Canonical bool
}

// selectionContract decides whether Gelium platform floors apply on top of
// reference/skin anatomy. Default is gelium; source is an explicit opt-in for
// third-party-faithful density. Product/site tokens can still override either.
type selectionContract string

const (
	contractGelium selectionContract = "gelium"
	contractSource selectionContract = "source"
)

type visualRecipe struct {
	Value     string
	Label     string
	Reference string
	Skin      productSkin
}

// visualRecipes is the public two-select vocabulary. Basecoat style packs
// (Vega default + Nova/Maia/…) and product skins win over reference. Material
// and Base UI neutral are explicit reference presets; Base UI neutral is
// Gelium-authored and never claims to ship Base UI package CSS.
var visualRecipes = []visualRecipe{
	{Value: "default", Label: "Default for behavior", Reference: "auto", Skin: skinNone},
	{Value: "material", Label: "Material visual", Reference: string(referenceMaterial), Skin: skinNone},
	{Value: "basecoat", Label: "Basecoat Vega", Reference: "auto", Skin: skinBasecoat},
	{Value: "basecoat-nova", Label: "Basecoat Nova", Reference: "auto", Skin: skinBasecoatNova},
	{Value: "basecoat-maia", Label: "Basecoat Maia", Reference: "auto", Skin: skinBasecoatMaia},
	{Value: "basecoat-lyra", Label: "Basecoat Lyra", Reference: "auto", Skin: skinBasecoatLyra},
	{Value: "basecoat-mira", Label: "Basecoat Mira", Reference: "auto", Skin: skinBasecoatMira},
	{Value: "basecoat-luma", Label: "Basecoat Luma", Reference: "auto", Skin: skinBasecoatLuma},
	{Value: "basecoat-sera", Label: "Basecoat Sera", Reference: "auto", Skin: skinBasecoatSera},
	{Value: "basecoat-rhea", Label: "Basecoat Rhea", Reference: "auto", Skin: skinBasecoatRhea},
	{Value: "baseui", Label: "Base UI neutral", Reference: string(referenceBaseUI), Skin: skinNone},
	{Value: "vercel", Label: "Vercel", Reference: "auto", Skin: skinVercel},
	{Value: "alden", Label: "Alden", Reference: "auto", Skin: skinAlden},
	{Value: "linear", Label: "Linear", Reference: "auto", Skin: skinLinear},
	{Value: "neubrutalism", Label: "Neubrutalism", Reference: "auto", Skin: skinNeubrutalism},
}

func visualRecipeByValue(value string) (visualRecipe, bool) {
	for _, recipe := range visualRecipes {
		if recipe.Value == strings.ToLower(strings.TrimSpace(value)) {
			return recipe, true
		}
	}
	return visualRecipe{}, false
}

func normalizeReference(raw string) (referencePreset, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "auto", "default":
		return "", true // resolved against behavior below
	case string(referenceNone):
		return referenceNone, true
	case string(referenceMaterial):
		return referenceMaterial, true
	case string(referenceBasecoat):
		return referenceBasecoat, true
	case string(referenceBaseUI):
		return referenceBaseUI, true
	default:
		return "", false
	}
}

func normalizeSkin(raw string) (productSkin, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", string(skinNone):
		return skinNone, true
	case string(skinMaterial):
		return skinMaterial, true
	case string(skinBasecoat):
		return skinBasecoat, true
	case string(skinBasecoatNova):
		return skinBasecoatNova, true
	case string(skinBasecoatMaia):
		return skinBasecoatMaia, true
	case string(skinBasecoatLyra):
		return skinBasecoatLyra, true
	case string(skinBasecoatMira):
		return skinBasecoatMira, true
	case string(skinBasecoatLuma):
		return skinBasecoatLuma, true
	case string(skinBasecoatSera):
		return skinBasecoatSera, true
	case string(skinBasecoatRhea):
		return skinBasecoatRhea, true
	case string(skinBaseUI):
		return skinBaseUI, true
	case string(skinAlden):
		return skinAlden, true
	case string(skinLinear):
		return skinLinear, true
	case string(skinVercel):
		return skinVercel, true
	case string(skinNeubrutalism):
		return skinNeubrutalism, true
	default:
		return "", false
	}
}

func normalizeContract(raw string) selectionContract {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(contractSource):
		return contractSource
	default:
		return contractGelium
	}
}

func isBasecoatFamilySkin(skin productSkin) bool {
	s := string(skin)
	return s == string(skinBasecoat) || strings.HasPrefix(s, "basecoat-")
}

func defaultReferenceForBehavior(behavior accordionBehavior) referencePreset {
	switch behavior {
	case accordionBehaviorMaterial:
		return referenceMaterial
	case accordionBehaviorBasecoat:
		return referenceBasecoat
	case accordionBehaviorBaseUI:
		return referenceBaseUI
	default:
		return referenceNone
	}
}

// canonicalSelectionRequested accepts only valid canonical inputs. Malformed
// query noise must not turn a legacy/default document into a neutral document.
func canonicalSelectionRequested(q url.Values) bool {
	if q.Has("behavior") {
		switch strings.ToLower(strings.TrimSpace(q.Get("behavior"))) {
		case string(accordionBehaviorNative), string(accordionBehaviorMaterial), string(accordionBehaviorBasecoat), string(accordionBehaviorBaseUI):
			return true
		}
	}
	if q.Has("execution") {
		switch strings.ToLower(strings.TrimSpace(q.Get("execution"))) {
		case string(accordionExecutionNative), string(accordionExecutionHTMX):
			return true
		}
	}
	if q.Has("reference") {
		if _, ok := normalizeReference(q.Get("reference")); ok {
			return true
		}
	}
	if q.Has("skin") {
		if _, ok := normalizeSkin(q.Get("skin")); ok {
			return true
		}
	}
	if q.Has("contract") {
		switch strings.ToLower(strings.TrimSpace(q.Get("contract"))) {
		case string(contractGelium), string(contractSource):
			return true
		}
	}
	return false
}

// resolveDocumentSelection is the single server-side resolver for public
// selection parameters. Legacy ?theme= is mapped per visual concern first;
// explicit reference/skin values then override only their own concern.
func resolveDocumentSelection(q url.Values) documentSelection {
	s := documentSelection{
		Behavior: normalizeAccordionBehavior(q.Get("behavior")),
		Skin:     skinNone,
		Scheme:   normalizeScheme(q.Get("scheme")),
		Contract: normalizeContract(q.Get("contract")),
	}
	legacySource := false
	if legacy, ok := themeBySlugOrClass(q.Get("theme")); ok {
		legacySource = true
		switch legacy.Slug {
		case "material":
			s.Reference = referenceMaterial
		case "basecoat":
			s.Reference = referenceBasecoat
		case "baseui":
			s.Reference = referenceBaseUI
		default:
			s.Skin = productSkin(legacy.Slug)
		}
	}
	if q.Has("reference") {
		if ref, ok := normalizeReference(q.Get("reference")); ok {
			s.Reference = ref
		}
	}
	if q.Has("skin") {
		if skin, ok := normalizeSkin(q.Get("skin")); ok {
			s.Skin = skin
		}
	}
	if s.Reference == "" {
		s.Reference = defaultReferenceForBehavior(s.Behavior)
	}
	s.Canonical = !legacySource && canonicalSelectionRequested(q)
	return s
}

// canonicalReferenceValue preserves the intentional "Default for behavior"
// contract across links. The resolver still exposes the effective preset in
// data-gelium-reference; only the URL uses auto where it is semantically true.
func canonicalReferenceValue(s documentSelection) string {
	if s.Skin == skinNone && s.Reference == defaultReferenceForBehavior(s.Behavior) {
		return "auto"
	}
	return string(s.Reference)
}

func canonicalSelectionQuery(s documentSelection, execution accordionExecution) string {
	q := url.Values{}
	q.Set("behavior", string(normalizeAccordionBehavior(string(s.Behavior))))
	q.Set("reference", canonicalReferenceValue(s))
	q.Set("skin", string(s.Skin))
	if scheme := normalizeScheme(s.Scheme); scheme != "" {
		q.Set("scheme", scheme)
	}
	// Only emit non-default contract to keep ordinary URLs short.
	if normalizeContract(string(s.Contract)) == contractSource {
		q.Set("contract", string(contractSource))
	}
	q.Set("execution", string(normalizeAccordionExecution(string(execution))))
	return q.Encode()
}

func canonicalSelectionHref(path string, s documentSelection, execution accordionExecution) string {
	return path + "?" + canonicalSelectionQuery(s, execution)
}

// navigationSelection describes which *source* of selection is safe to carry
// into links. Legacy theme/scheme selections stay in their legacy query form;
// canonical recipe selections stay canonical across the docs IA so sidebar
// navigation does not silently reset behavior, visual recipe, or appearance.
type navigationSelection struct {
	themeSlug string
	scheme    string
	canonical bool
	selection documentSelection
	execution accordionExecution
}

func recipeNavigationContext(path string) bool {
	return path == "/components/accordion" || path == "/docs/themes/gallery"
}

func navigationSelectionFor(r *http.Request, _ string) navigationSelection {
	if r == nil || r.URL == nil {
		return navigationSelection{}
	}
	q := r.URL.Query()
	n := navigationSelection{scheme: normalizeScheme(q.Get("scheme"))}
	if theme, ok := themeBySlugOrClass(q.Get("theme")); ok {
		n.themeSlug = theme.Slug
		return n
	}
	if q.Has("behavior") || q.Has("reference") || q.Has("skin") || q.Has("execution") {
		n.canonical = true
		n.selection = resolveDocumentSelection(q)
		n.execution = normalizeAccordionExecution(q.Get("execution"))
	}
	return n
}

func (n navigationSelection) href(path string) string {
	if n.canonical {
		return canonicalSelectionHref(path, n.selection, n.execution)
	}
	return chromeHref(path, n.themeSlug, n.scheme)
}

func (n navigationSelection) query() string {
	if n.canonical {
		return "?" + canonicalSelectionQuery(n.selection, n.execution)
	}
	return chromeQuery(n.themeSlug, n.scheme)
}

func (n navigationSelection) hasLegacyChrome() bool {
	return n.themeSlug != "" || n.scheme != ""
}

func (s documentSelection) legacyThemeClass() string {
	// The explicit Native "Default for behavior" is Gelium neutral, not an
	// implicit Material adapter. The same resolved values from no selection or
	// malformed legacy input retain the historical default class below.
	if s.Canonical && s.Reference == referenceNone && s.Skin == skinNone {
		return ""
	}
	if s.Skin != skinNone {
		// Basecoat style packs share the Basecoat color theme sheet; pack
		// differences live in data-gelium-skin anatomy tokens.
		if isBasecoatFamilySkin(s.Skin) {
			return "theme-basecoat"
		}
		return "theme-" + string(s.Skin)
	}
	if s.Reference != referenceNone {
		return "theme-" + string(s.Reference)
	}
	return defaultThemeClass
}

func referenceFromRequest(r *http.Request) referencePreset {
	if r == nil {
		return referenceNone
	}
	if v, ok := r.Context().Value(referenceContextKey{}).(referencePreset); ok {
		return v
	}
	return referenceNone
}

func skinFromRequest(r *http.Request) productSkin {
	if r == nil {
		return skinNone
	}
	if v, ok := r.Context().Value(skinContextKey{}).(productSkin); ok {
		return v
	}
	return skinNone
}

func contractFromRequest(r *http.Request) selectionContract {
	if r == nil {
		return contractGelium
	}
	if v, ok := r.Context().Value(contractContextKey{}).(selectionContract); ok {
		return normalizeContract(string(v))
	}
	return contractGelium
}

func canonicalSelectionFromRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	if v, ok := r.Context().Value(canonicalSelectionContextKey{}).(bool); ok {
		return v
	}
	return false
}

// themeFromRequest returns the theme selected by ?theme= if present and valid,
// otherwise the empty string (so callers fall back to themeClass("") → default).
func themeFromRequest(r *http.Request) string {
	if v, ok := r.Context().Value(themeContextKey{}).(string); ok {
		return v
	}
	return ""
}

// schemeFromRequest returns "light" or "dark" when ?scheme= is allowlisted,
// otherwise empty (OS prefers-color-scheme media route applies).
func schemeFromRequest(r *http.Request) string {
	if v, ok := r.Context().Value(schemeContextKey{}).(string); ok {
		return v
	}
	return ""
}

// normalizeScheme returns "light", "dark", or "" for unknown/empty values.
func normalizeScheme(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "light":
		return "light"
	case "dark":
		return "dark"
	default:
		return ""
	}
}

// recipeScheme makes recipe changes deterministic when no explicit appearance
// has been chosen yet. The appearance switch still lets users opt into dark.
func recipeScheme(raw string) string {
	if scheme := normalizeScheme(raw); scheme != "" {
		return scheme
	}
	return "light"
}

// themeBySlugOrClass resolves a ?theme= value to a catalog entry.
func themeBySlugOrClass(raw string) (themeDirection, bool) {
	name := raw
	if !strings.HasPrefix(name, "theme-") {
		name = "theme-" + name
	}
	for _, t := range availableThemes {
		if t.Class == name || t.Slug == raw {
			return t, true
		}
	}
	return themeDirection{}, false
}

// themeQueryMiddleware resolves every public selection at the server boundary.
// Unknown values cannot cross into templates, classes, or document attributes.
func themeQueryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The single visual <select> sends a short-lived `visual` recipe value.
		// Native HTML cannot change two independent input names from one select;
		// canonicalize it at the server boundary so the browser ends on a stable
		// behavior/reference/skin/scheme/execution URL with no legacy ?theme=.
		if recipe, ok := visualRecipeByValue(r.URL.Query().Get("visual")); ok {
			q := url.Values{}
			q.Set("behavior", string(normalizeAccordionBehavior(r.URL.Query().Get("behavior"))))
			q.Set("reference", recipe.Reference)
			q.Set("skin", string(recipe.Skin))
			if scheme := normalizeScheme(r.URL.Query().Get("scheme")); scheme != "" {
				q.Set("scheme", scheme)
			}
			if normalizeContract(r.URL.Query().Get("contract")) == contractSource {
				q.Set("contract", string(contractSource))
			}
			q.Set("execution", string(normalizeAccordionExecution(r.URL.Query().Get("execution"))))
			target := r.URL.Path + "?" + q.Encode()
			http.Redirect(w, r, target, http.StatusSeeOther)
			return
		}
		selection := resolveDocumentSelection(r.URL.Query())
		ctx := context.WithValue(r.Context(), referenceContextKey{}, selection.Reference)
		ctx = context.WithValue(ctx, skinContextKey{}, selection.Skin)
		ctx = context.WithValue(ctx, canonicalSelectionContextKey{}, selection.Canonical)
		ctx = context.WithValue(ctx, schemeContextKey{}, selection.Scheme)
		ctx = context.WithValue(ctx, contractContextKey{}, selection.Contract)
		ctx = context.WithValue(ctx, accordionBehaviorContextKey{}, selection.Behavior)
		ctx = context.WithValue(ctx, accordionExecutionContextKey{}, normalizeAccordionExecution(r.URL.Query().Get("execution")))
		// Retain ThemeClass only as a legacy selector adapter during migration.
		if class := selection.legacyThemeClass(); class != "" {
			ctx = context.WithValue(ctx, themeContextKey{}, class)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// themeClass resolves a requested theme to a safe CSS class. Theme identity is
// server-driven and validated against an allowlist of themes that exist on
// disk; unknown values fall back to the default so no page can inject an
// arbitrary class or depend on a theme that does not exist.
//
// Allowlist rule (Phase H): a theme class only enters this list together with
// its bundle entry — an import in site/web/styles/app.css AND lib/themes/<name>.css
// on disk. theme-basecoat joined the list in Phase I, at the same commit that
// created lib/themes/theme-basecoat.css and its app.css import. Adding a string before
// the theme exists would let a page select a theme that is not in the bundle.
func themeClass(theme string) string {
	for _, t := range availableThemes {
		if theme == t.Class {
			return theme
		}
	}
	return defaultThemeClass
}

// requestPath returns the URL path for chrome links, defaulting to "/".
func requestPath(r *http.Request) string {
	if r != nil && r.URL != nil && r.URL.Path != "" {
		return r.URL.Path
	}
	return "/"
}

// chromeQuery builds the closed-vocabulary query string for docs chrome links.
// Only allowlisted theme + scheme keys are emitted (never arbitrary request query).
func chromeQuery(themeSlug, scheme string, profiles ...string) string {
	q := url.Values{}
	if themeSlug != "" {
		if t, ok := themeBySlugOrClass(themeSlug); ok {
			q.Set("theme", t.Slug)
		}
	}
	if s := normalizeScheme(scheme); s != "" {
		q.Set("scheme", s)
	}
	if len(profiles) > 0 {
		if b := normalizeAccordionBehavior(profiles[0]); b != accordionBehaviorNative || strings.EqualFold(strings.TrimSpace(profiles[0]), string(accordionBehaviorNative)) {
			q.Set("behavior", string(b))
		}
	}
	if len(profiles) > 1 {
		if e := normalizeAccordionExecution(profiles[1]); e != accordionExecutionNative || strings.EqualFold(strings.TrimSpace(profiles[1]), string(accordionExecutionNative)) {
			q.Set("execution", string(e))
		}
	}
	enc := q.Encode()
	if enc == "" {
		return ""
	}
	return "?" + enc
}

// chromeHref is path + allowlisted theme/scheme query for shell navigation.
func chromeHref(path, themeSlug, scheme string, profiles ...string) string {
	return path + chromeQuery(themeSlug, scheme, profiles...)
}

// themeSwitcherFor builds the direction switcher for the current request.
// Labels are product directions (Material, Basecoat, …), never internal class
// names. The native select submits ?theme=<slug> to the same URL (0-JS GET
// form); the hidden scheme input preserves the current ?scheme= when set.
// Other query params are dropped, matching the old link-list contract.
func themeSwitcherFor(r *http.Request, currentClass, _ string, scheme string) *themeSwitcherView {
	current := themeClass(currentClass)
	if r != nil {
		if q := themeFromRequest(r); q != "" {
			current = q
		}
	}
	opts := make([]themeOptionView, 0, len(availableThemes))
	for _, t := range availableThemes {
		opts = append(opts, themeOptionView{
			Label:    t.Label,
			Value:    t.Slug,
			Class:    t.Class,
			Selected: t.Class == current,
		})
	}
	return &themeSwitcherView{
		Label: "Skin", Options: opts, Scheme: normalizeScheme(scheme),
		Reference: string(referenceFromRequest(r)), Skin: string(skinFromRequest(r)),
	}
}

func visualForSelection(r *http.Request, s documentSelection) string {
	// A valid legacy direction has a specific visual intent even where its
	// resolved reference happens to match a behavior default.
	if r != nil {
		if legacy, ok := themeBySlugOrClass(r.URL.Query().Get("theme")); ok {
			if _, exists := visualRecipeByValue(legacy.Slug); exists {
				return legacy.Slug
			}
		}
	}
	if s.Skin != skinNone {
		if _, ok := visualRecipeByValue(string(s.Skin)); ok {
			return string(s.Skin)
		}
	}
	switch s.Reference {
	case referenceMaterial:
		if s.Behavior != accordionBehaviorMaterial {
			return "material"
		}
	case referenceBaseUI:
		if s.Behavior != accordionBehaviorBaseUI {
			return "baseui"
		}
	case referenceBasecoat:
		if s.Behavior != accordionBehaviorBasecoat {
			return "basecoat"
		}
	}
	return "default"
}

func behaviorLabel(b accordionBehavior) string {
	for _, option := range recipeBehaviorOptions(b) {
		if option.Selected {
			return option.Label
		}
	}
	return "Native"
}

func recipeBehaviorOptions(b accordionBehavior) []profileOptionView {
	return []profileOptionView{
		{Label: "Native", Value: "native", Selected: b == accordionBehaviorNative},
		{Label: "Material", Value: "material", Selected: b == accordionBehaviorMaterial},
		{Label: "Basecoat", Value: "basecoat", Selected: b == accordionBehaviorBasecoat},
		{Label: "Base UI", Value: "baseui", Selected: b == accordionBehaviorBaseUI},
	}
}

func recipeSwitcherFor(r *http.Request, selection documentSelection, execution accordionExecution, compact bool, id string) *recipeSwitcherView {
	b := normalizeAccordionBehavior(string(selection.Behavior))
	e := normalizeAccordionExecution(string(execution))
	c := normalizeContract(string(selection.Contract))
	visual := visualForSelection(r, documentSelection{Behavior: b, Reference: selection.Reference, Skin: selection.Skin, Scheme: selection.Scheme, Contract: c})
	visuals := make([]recipeOptionView, 0, len(visualRecipes))
	visualLabel := "Default for behavior"
	for _, recipe := range visualRecipes {
		selected := recipe.Value == visual
		if selected {
			visualLabel = recipe.Label
		}
		visuals = append(visuals, recipeOptionView{Label: recipe.Label, Value: recipe.Value, Selected: selected})
	}
	executionLabel := "Native baseline"
	if e == accordionExecutionHTMX {
		executionLabel = "HTMX optional enhancement"
	}
	contractLabel := "Gelium defaults"
	if c == contractSource {
		contractLabel = "Source-faithful density"
	}
	return &recipeSwitcherView{
		ID: id, Compact: compact, RenderInTopbar: compact, Behavior: string(b), Visual: visual,
		Scheme: recipeScheme(selection.Scheme), Execution: string(e), Contract: string(c),
		Behaviors: recipeBehaviorOptions(b), Visuals: visuals,
		Executions: []profileOptionView{{Label: "Native baseline", Value: "native", Selected: e == accordionExecutionNative}, {Label: "HTMX optional server enhancement", Value: "htmx", Selected: e == accordionExecutionHTMX}},
		Contracts: []profileOptionView{
			{Label: "Gelium defaults (touch floors)", Value: string(contractGelium), Selected: c == contractGelium},
			{Label: "Source-faithful density", Value: string(contractSource), Selected: c == contractSource},
		},
		Summary: behaviorLabel(b) + " behavior · " + visualLabel + " · " + contractLabel + " · " + executionLabel,
	}
}

// schemeSwitcherFor builds the Light/Dark appearance control: a native switch
// in a 0-JS GET form. Checked → ?scheme=dark, unchecked → ?scheme=light
// (a hidden light twin after the checkbox supplies the light value, since a
// cleared checkbox submits nothing). The hidden theme input keeps the current
// ?theme= slug. No scheme yet = light (matches default light tokens).
func schemeSwitcherFor(_ *http.Request, themeSlug, scheme string) *schemeSwitcherView {
	selection := resolveDocumentSelection(url.Values{"theme": []string{themeSlug}, "scheme": []string{scheme}})
	switcher := schemeSwitcherForSelection(selection, accordionExecutionNative)
	switcher.Theme = themeSlug
	return switcher
}

func schemeSwitcherForNavigation(selection documentSelection, execution accordionExecution, navigation navigationSelection) *schemeSwitcherView {
	switcher := schemeSwitcherForSelection(selection, execution)
	switcher.Theme = navigation.themeSlug
	switcher.Canonical = navigation.canonical
	return switcher
}

func schemeSwitcherForSelection(selection documentSelection, execution accordionExecution) *schemeSwitcherView {
	return &schemeSwitcherView{
		Label: "Appearance", Selection: selection,
		Execution: normalizeAccordionExecution(string(execution)),
		Checked:   normalizeScheme(selection.Scheme) == "dark",
	}
}

// applyDocumentRootScheme mutates ThemeClass / DataTheme for explicit scheme.
// dark → append theme-dark (class route). light → data-theme="light" so the
// prefers-color-scheme media block's :not([data-theme="light"]) guard skips.
func applyDocumentRootScheme(data *pageView, scheme string) {
	data.Scheme = normalizeScheme(scheme)
	switch data.Scheme {
	case "dark":
		if !strings.Contains(data.ThemeClass, "theme-dark") {
			data.ThemeClass = strings.TrimSpace(data.ThemeClass + " theme-dark")
		}
		data.DataTheme = "dark"
	case "light":
		data.DataTheme = "light"
	default:
		data.DataTheme = ""
	}
}

// applyDocumentSelection writes the resolved visual contract to a page view.
// ThemeClass remains a temporary selector adapter for legacy theme sheets; the
// data-gelium-* attributes are the stable Phase-1 selection contract.
func applyDocumentSelection(data *pageView, r *http.Request) documentSelection {
	selection := documentSelection{
		Behavior:  accordionBehaviorFromRequest(r),
		Reference: referenceFromRequest(r),
		Skin:      skinFromRequest(r),
		Scheme:    schemeFromRequest(r),
		Contract:  contractFromRequest(r),
		Canonical: canonicalSelectionFromRequest(r),
	}
	data.Reference = string(selection.Reference)
	data.Skin = string(selection.Skin)
	data.Contract = string(normalizeContract(string(selection.Contract)))
	data.ThemeClass = selection.legacyThemeClass()
	applyDocumentRootScheme(data, selection.Scheme)
	return selection
}

// bannerView is the server-driven view model for the page-level BANNER slot
// (Phase D pattern 5). The partial is a primitive: no handler wires it yet —
// Phase G will inject it when a page/site condition (expired session,
// maintenance, pending consent) is detected. A nil Banner on pageView renders
// no banner. Tone is a closed vocabulary (info|success|warning|error), the
// CTA is a real link and the dismiss a POST+303 action, both 0 JS.
type bannerView struct {
	Tone        string
	Icon        template.HTML
	Title       string
	Body        string
	CTA         bool
	CTAHref     string
	CTALabel    string
	DismissHref string
	DismissIcon template.HTML
}

// breadcrumbView is the server-driven view model for the BREADCRUMB public
// content pattern (Phase F). A nil Breadcrumb on pageView renders no breadcrumb.
// The last item with Current=true renders as <span aria-current="page">; the
// rest are real links. Zero JS, native <nav>/<ol> semantics.
type breadcrumbView struct {
	Items []breadcrumbItem
}

// breadcrumbItem is one crumb: a real link (Href) or the current page marker
// (Current=true renders the label without a link).
type breadcrumbItem struct {
	Href    string
	Label   string
	Current bool
}

// defaultBreadcrumb builds the standard Home > <label> trail for a page label.
// Components pages can override with a deeper trail by setting pageView.Breadcrumb.
func defaultBreadcrumb(label string) *breadcrumbView {
	return &breadcrumbView{Items: []breadcrumbItem{
		{Href: "/", Label: "Home"},
		{Label: label, Current: true},
	}}
}

// breadcrumbWithChrome returns a copy of bc whose non-current link hrefs carry
// allowlisted ?theme= / ?scheme= chrome query. Current crumbs stay unlinked.
// Home "/" is left bare — outside the docs shell.
func breadcrumbWithChrome(bc *breadcrumbView, themeSlug, scheme string) *breadcrumbView {
	if bc == nil {
		return bc
	}
	if themeSlug == "" && normalizeScheme(scheme) == "" {
		return bc
	}
	out := &breadcrumbView{Items: make([]breadcrumbItem, len(bc.Items))}
	copy(out.Items, bc.Items)
	for i := range out.Items {
		it := &out.Items[i]
		if it.Current || it.Href == "" || it.Href == "/" {
			continue
		}
		// Only rewrite plain path hrefs (no existing query).
		if strings.Contains(it.Href, "?") {
			continue
		}
		it.Href = chromeHref(it.Href, themeSlug, scheme)
	}
	return out
}

// breadcrumbWithSelection is the recipe-aware variant for docs shell links.
func breadcrumbWithSelection(bc *breadcrumbView, selection documentSelection, execution accordionExecution) *breadcrumbView {
	if bc == nil {
		return nil
	}
	out := &breadcrumbView{Items: make([]breadcrumbItem, len(bc.Items))}
	copy(out.Items, bc.Items)
	for i := range out.Items {
		item := &out.Items[i]
		if item.Current || item.Href == "" || item.Href == "/" || strings.Contains(item.Href, "?") {
			continue
		}
		item.Href = canonicalSelectionHref(item.Href, selection, execution)
	}
	return out
}

// inlineAlertView is the server-driven view model for the section/form-level
// INLINE ALERT slot (Phase D pattern 6). A nil InlineAlert on pageView renders
// no alert. Tone is a closed vocabulary (info|success|warning|error) and the
// role derives from it (error → alert, everything else → status). It is a
// persistent server-rendered signal, never a transient toast.
type inlineAlertView struct {
	Tone  string
	Icon  template.HTML
	Title string
	Body  string
}

// errorStateView is the server-driven view model for the page-level ERROR
// STATE slot (Phase D pattern 7). A nil Error on pageView renders no error
// state and the normal content instead. The status code is the canonical
// attribute: the handler picks the real HTTP status (404/500) and the copy
// per status. Retry is optional — a real GET link back to a known URL —
// and everything works with 0 JS.
// footerView is the server-driven view model for the page-level FOOTER slot
// (Phase F public content pattern). The footer is the site chrome (brand,
// secondary nav, legal) that every layout page renders; a nil Footer omits it
// so partial layouts and tests can opt out. Sections collapse natively with
// <details>/<summary> on narrow screens and stay forced-open on desktop —
// zero JS. Links reuse navLink so the chrome can never drift from the nav.
type footerView struct {
	Brand    string
	Sections []footerSection
	Legal    string
}

// footerSection is one collapsible link group inside the footer secondary nav.
type footerSection struct {
	Title string
	Links []navLink
}

// defaultFooter is the shared public chrome data used by the landing and docs.
// The docs sidebar remains the full IA surface; the footer stays intentionally
// concise. Injected at render choke points; a consumer may replace it per page
// by setting pageView.Footer explicitly.
func defaultFooter() *footerView {
	return marketingFooter()
}

type errorStateView struct {
	StatusCode int
	Title      string
	Body       string
	Retry      bool
	Href       string
	Label      string
}

// siteBaseURL is the canonical origin for absolute URLs (canonical, og:url,
// JSON-LD). Gelium UI ships with no production host yet — the docs are
// embedded and self-hosted — so this is a documented placeholder origin
// matching the SEO contract examples. Canonical URLs are built from it plus
// the clean route path and never from the request Host, so metadata stays
// stable across every deployment. The origin is env-configurable through
// BASE_URL at startup (SEO contract §2); defaultBaseURL is the shipped
// placeholder. Swap the default for the real origin when a public domain
// exists.
const defaultBaseURL = "https://gelium-ui.example"

// resolveBaseURL resolves the canonical origin from the BASE_URL environment
// variable. Empty input falls back to defaultBaseURL; trailing slashes are
// trimmed so canonical URLs never double-slash. Pure function — unit-testable
// without env pollution in parallel tests.
func resolveBaseURL(env string) string {
	if env == "" {
		return defaultBaseURL
	}
	return strings.TrimRight(env, "/")
}

// siteBaseURL is resolved once at startup so every absolute URL family
// (canonical, og:url, og:image, JSON-LD, sitemap) derives from one origin.
var siteBaseURL = resolveBaseURL(os.Getenv("BASE_URL"))

// licenseURL is the link target of the article provenance license reference
// (GEO §15). Design D5: the OSI MIT page — no /LICENSE route exists.
const licenseURL = "https://opensource.org/licenses/MIT"

const (
	// homeDescription is the system description for the landing page.
	homeDescription = "Gelium UI — themeable, open-code UI components for server-rendered applications. Native HTML semantics, zero component JavaScript, Material 3 design."
	// docsDescription is the system description for the /docs index.
	docsDescription = "Gelium UI component library, organized by category. Every page is dogfooded: it renders the real component it documents."
	// defaultMetaDescription is the fallback description for any route without
	// its own. Component pages derive a per-route description from their label
	// so every URL stays unique (contract: one description per URL).
	defaultMetaDescription = "Gelium UI — open-code, server-rendered UI components with native HTML semantics."
)

// metaView is the server-driven metadata model emitted into <head>. It is
// theme-agnostic by contract: switching themes must not change a single byte
// of SEO-relevant markup. A zero metaView renders a valid minimal head (the
// layout guards every tag behind {{if}}), so error pages and partials keep
// working without metadata.
type metaView struct {
	Title         string // mirrors pageView.Title (single source in the handler)
	Description   string
	Canonical     string
	Robots        string // default "index, follow"
	OGTitle       string
	OGDescription string
	OGType        string
	OGImage       string
	TwitterCard   string
	TwitterTitle  string
	TwitterDesc   string
	TwitterURL    string
	JSONLD        template.JS // trusted structured data; emitted inside <script type="application/ld+json">
	Lang          string      // default "en"
}

// ogImagePlaceholder is the default social image URL emitted as og:image on
// every layout page. The actual PNG is not shipped yet — the placeholder origin
// is documented in the SEO contract so the markup contract stays stable until
// a real asset lands. It is a var (not const) because it derives from the
// startup-resolved siteBaseURL.
var ogImagePlaceholder = siteBaseURL + "/og.png"

// jsonLDPublisher and webSiteLD are the home-page WebSite structured-data
// types. They are marshaled with encoding/json so escaping and validity are
// guaranteed (contract §12: no JSON built by string concatenation).
type jsonLDPublisher struct {
	Type string `json:"@type"`
	Name string `json:"name"`
}

type webSiteLD struct {
	Context    string          `json:"@context"`
	Type       string          `json:"@type"`
	Name       string          `json:"name"`
	URL        string          `json:"url"`
	InLanguage string          `json:"inLanguage"`
	Publisher  jsonLDPublisher `json:"publisher"`
}

// websiteJSONLD is the home-page WebSite block, built once at startup. The
// value is template.JS on purpose: inside a <script type="application/ld+json">
// element, html/template escapes template.HTML as a JSON string, while
// template.JS (a valid JS expression; JSON is valid JS) is emitted verbatim.
var websiteJSONLD = func() template.JS {
	b, err := json.Marshal(webSiteLD{
		Context:    "https://schema.org",
		Type:       "WebSite",
		Name:       "Gelium UI",
		URL:        siteBaseURL + "/",
		InLanguage: "en",
		Publisher:  jsonLDPublisher{Type: "Organization", Name: "Gelium UI"},
	})
	if err != nil {
		return template.JS("")
	}
	return template.JS(b) // #nosec G203 -- trusted, system-generated JSON.
}()

// jsonLDBreadcrumb is the BreadcrumbList entity emitted on component pages.
// It mirrors the visible breadcrumb trail (Home > Components > page) so the
// structured data never disagrees with the rendered navigation (contract §12).
type jsonLDBreadcrumb struct {
	Type  string           `json:"@type"`
	Items []jsonLDListItem `json:"itemListElement"`
}

// jsonLDListItem is one crumb of the BreadcrumbList. Item is the canonical
// absolute URL; the current page is the last item.
type jsonLDListItem struct {
	Type     string `json:"@type"`
	Position int    `json:"position"`
	Name     string `json:"name"`
	Item     string `json:"item"`
}

// jsonLDArticle is the TechArticle/Article entity emitted on component pages
// with the page headline, canonical URL, and — when the date table has an
// entry for the slug — ISO-8601 datePublished/dateModified that mirror the
// visible provenance line (GEO §7). Authored and published by the Gelium UI
// organization entity.
type jsonLDArticle struct {
	Type          string          `json:"@type"`
	Headline      string          `json:"headline"`
	URL           string          `json:"url"`
	InLanguage    string          `json:"inLanguage"`
	DatePublished string          `json:"datePublished,omitempty"`
	DateModified  string          `json:"dateModified,omitempty"`
	Author        jsonLDPublisher `json:"author"`
	Publisher     jsonLDPublisher `json:"publisher"`
}

type jsonLDBlogPosting struct {
	Context          string          `json:"@context"`
	Type             string          `json:"@type"`
	Headline         string          `json:"headline"`
	DatePublished    string          `json:"datePublished"`
	Author           jsonLDPublisher `json:"author"`
	Publisher        jsonLDPublisher `json:"publisher"`
	MainEntityOfPage string          `json:"mainEntityOfPage"`
}

type jsonLDWebPage struct {
	Context     string `json:"@context"`
	Type        string `json:"@type"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

func blogJSONLD(routePath string) template.JS {
	slug := strings.TrimPrefix(routePath, "/blog/")
	for _, post := range blogPosts {
		if post.Slug != slug {
			continue
		}
		b, err := json.Marshal(jsonLDBlogPosting{
			Context: "https://schema.org", Type: "BlogPosting", Headline: post.Title,
			DatePublished:    post.Date,
			Author:           jsonLDPublisher{Type: "Person", Name: post.Author},
			Publisher:        jsonLDPublisher{Type: "Organization", Name: "Gelium UI"},
			MainEntityOfPage: siteBaseURL + routePath,
		})
		if err != nil {
			return template.JS("")
		}
		return template.JS(b) // #nosec G203 -- trusted, registry-derived JSON.
	}
	return template.JS("")
}

func docsJSONLD(routePath, title, description string) template.JS {
	typeName := "WebPage"
	if routePath == "/docs" {
		typeName = "CollectionPage"
	}
	b, err := json.Marshal(jsonLDWebPage{
		Context: "https://schema.org", Type: typeName, Name: title,
		URL: siteBaseURL + routePath, Description: description,
	})
	if err != nil {
		return template.JS("")
	}
	return template.JS(b) // #nosec G203 -- trusted, system-generated JSON.
}

// jsonLDSoftwareApplication is the SoftwareApplication entity emitted on every
// /components/* page (GEO §14): name, category, the single-source
// docsShellVersion, the operating system scope, and the MIT license.
type jsonLDSoftwareApplication struct {
	Type                string `json:"@type"`
	Name                string `json:"name"`
	ApplicationCategory string `json:"applicationCategory"`
	SoftwareVersion     string `json:"softwareVersion"`
	OperatingSystem     string `json:"operatingSystem"`
	License             string `json:"license"`
}

// jsonLDGraph wraps the per-page entities in a single @graph document so the
// layout emits one <script type="application/ld+json"> block per page.
type jsonLDGraph struct {
	Context string `json:"@context"`
	Graph   []any  `json:"@graph"`
}

// componentJSONLD builds the @graph structured data for a registered
// /components/* page: the BreadcrumbList trail plus a TechArticle carrying the
// page headline, canonical URL and (when present in the date table) the
// published/modified dates, plus a SoftwareApplication entity describing the
// library (GEO §7, §14). Built with encoding/json so escaping and validity are
// guaranteed (contract §12: no JSON by string concatenation).
func componentJSONLD(routePath string) template.JS {
	label := componentLabel(routePath)
	canonical := siteBaseURL + routePath
	article := jsonLDArticle{
		Type:       "TechArticle",
		Headline:   label + " · Gelium UI",
		URL:        canonical,
		InLanguage: "en",
		Author:     jsonLDPublisher{Type: "Organization", Name: "Gelium UI"},
		Publisher:  jsonLDPublisher{Type: "Organization", Name: "Gelium UI"},
	}
	if dates, ok := docDatesFor(strings.TrimPrefix(routePath, "/components/")); ok {
		article.DatePublished = dates.Published
		article.DateModified = dates.Modified
	}
	graph := jsonLDGraph{
		Context: "https://schema.org",
		Graph: []any{
			jsonLDBreadcrumb{
				Type: "BreadcrumbList",
				Items: []jsonLDListItem{
					{Type: "ListItem", Position: 1, Name: "Home", Item: siteBaseURL + "/"},
					{Type: "ListItem", Position: 2, Name: "Components", Item: siteBaseURL + "/docs"},
					{Type: "ListItem", Position: 3, Name: label, Item: canonical},
				},
			},
			article,
			jsonLDSoftwareApplication{
				Type:                "SoftwareApplication",
				Name:                "Gelium UI",
				ApplicationCategory: "DeveloperApplication",
				SoftwareVersion:     docsShellVersion,
				OperatingSystem:     "Any",
				License:             "MIT",
			},
		},
	}
	b, err := json.Marshal(graph)
	if err != nil {
		return template.JS("")
	}
	return template.JS(b) // #nosec G203 -- trusted, system-generated JSON.
}

// routePathForContent maps an embedded markdown file to its public route path.
// Content files live at content/<route-segment>.md and render at
// /components/<route-segment>; content/index.md renders at /. The convention
// is what keeps the canonical URL clean and query-free.
func routePathForContent(contentPath string) string {
	base := path.Base(contentPath)
	if base == "index.md" {
		return "/"
	}
	return "/components/" + strings.TrimSuffix(base, ".md")
}

// componentLabel resolves the navigation label for a /components/* route from
// the single route registry, so metadata never drifts from the nav.
func componentLabel(routePath string) string {
	for _, r := range componentRoutes() {
		if r.Path == routePath {
			return r.Label
		}
	}
	return routePath
}

// componentRouteLabel returns the registry label for a registered component
// docs page and whether the path is one. Unregistered /components/* paths
// (e.g. /components/dialog/confirm) report false so they never fabricate a
// component identity in the breadcrumb or structured data.
func componentRouteLabel(routePath string) (string, bool) {
	for _, r := range componentRoutes() {
		if r.Path == routePath {
			return r.Label, true
		}
	}
	return "", false
}

// componentBreadcrumb is the trail for a registered component docs page:
// Home > Components > <label>. It mirrors the BreadcrumbList JSON-LD emitted
// for the same page, so the visible breadcrumb and the structured data always
// agree (contract §11). The Components crumb links to the /docs index.
func componentBreadcrumb(label, _ string) *breadcrumbView {
	return &breadcrumbView{Items: []breadcrumbItem{
		{Href: "/", Label: "Home"},
		{Href: "/docs", Label: "Components"},
		{Label: label, Current: true},
	}}
}

// componentDescription derives a unique per-route description for a component
// page from its registry label (contract: one description per URL).
func componentDescription(label string) string {
	return "The " + label + " component in Gelium UI — server-rendered docs with native HTML semantics, zero component JavaScript."
}

// resolveMeta computes the page metadata for a route path at the single render
// choke point. Fields already set by the handler (data.Meta) win; resolution
// only fills what is still empty. Demos and examples are never indexed; every
// other route defaults to "index, follow". The canonical is always the clean
// route path — no query string can leak into it (contract §16).
func resolveMeta(data pageView, routePath string) metaView {
	meta := data.Meta
	meta.Title = data.Title
	if meta.Lang == "" {
		meta.Lang = "en"
	}
	if meta.Robots == "" {
		meta.Robots = indexabilityPolicy(routePath)
	}

	meta.Canonical = siteBaseURL + routePath

	if meta.Description == "" {
		switch {
		case routePath == "/":
			meta.Description = homeDescription
		case routePath == "/docs":
			meta.Description = docsDescription
		case strings.HasPrefix(routePath, "/components/"):
			meta.Description = componentDescription(componentLabel(routePath))
		default:
			meta.Description = defaultMetaDescription
		}
	}
	if meta.OGType == "" {
		switch {
		case routePath == "/" || routePath == "/docs":
			meta.OGType = "website"
		case strings.HasPrefix(routePath, "/components/"):
			meta.OGType = "article"
		default:
			meta.OGType = "website"
		}
	}
	if meta.OGTitle == "" {
		if routePath == "/" {
			meta.OGTitle = "Gelium UI"
		} else {
			meta.OGTitle = meta.Title + " · Gelium UI"
		}
	}
	if meta.OGDescription == "" {
		meta.OGDescription = meta.Description
	}
	if meta.OGImage == "" {
		meta.OGImage = ogImagePlaceholder
	}
	if meta.TwitterCard == "" {
		if meta.OGImage != "" {
			meta.TwitterCard = "summary_large_image"
		} else {
			meta.TwitterCard = "summary"
		}
	}
	if meta.TwitterTitle == "" {
		meta.TwitterTitle = meta.OGTitle
	}
	if meta.TwitterDesc == "" {
		meta.TwitterDesc = meta.OGDescription
	}
	if meta.TwitterURL == "" {
		meta.TwitterURL = meta.Canonical
	}
	if routePath == "/" && meta.JSONLD == "" {
		meta.JSONLD = websiteJSONLD
	}
	if meta.JSONLD == "" {
		if _, ok := componentRouteLabel(routePath); ok {
			meta.JSONLD = componentJSONLD(routePath)
		}
	}
	if meta.JSONLD == "" && strings.HasPrefix(routePath, "/blog/") {
		meta.JSONLD = blogJSONLD(routePath)
	}
	if meta.JSONLD == "" && strings.HasPrefix(routePath, "/docs") {
		meta.JSONLD = docsJSONLD(routePath, meta.Title, meta.Description)
	}
	return meta
}

// provenanceView is the server-side provenance line rendered inside component
// articles (GEO §8, §15): version, license link, source reference and the
// ISO dates from the date table. One struct → one {{if .Provenance}} guard in
// the layout; nil on non-component and error pages, so nothing is emitted.
type provenanceView struct {
	Version    string
	LicenseURL string
	Source     string
	Published  string
	Modified   string
}

// tocEntry is one "On this page" section: the goldmark auto heading ID plus
// the heading text. Level 2 is top-level, level 3 nests under it.
type tocEntry struct {
	ID    string
	Text  string
	Level int
}

// prevNextLink is one pagination destination (GOV.UK previous/next pattern).
type prevNextLink struct {
	Href  string
	Label string
}

// prevNextView is the previous/next pagination; Prev or Next is nil on the
// first/last IA boundary so the template renders a spacer.
type prevNextView struct {
	Prev *prevNextLink
	Next *prevNextLink
}

type pageView struct {
	Meta  metaView
	Title string
	// AssetsVersion is the single cache-busting version for static asset URLs
	// (rendered as ?v={{.AssetsVersion}} in templates). Centralized in
	// geliumui/lib so package, Go, and template surfaces cannot drift.
	AssetsVersion string
	// H1 is the leading level-1 heading of a component page, rendered as its
	// own block so the live demo can sit between the title and the body
	// (Base UI/Naive UI order: show the component, then the rules). Empty on
	// pages whose markdown does not lead with an H1 (handbook, hub, landing).
	H1      template.HTML
	Content template.HTML
	// ContentRest is the tail of a pilot component page's markdown (from
	// "## Guidance" onward), rendered AFTER the server-injected Examples and
	// API reference sections. Empty on non-pilot pages.
	ContentRest template.HTML
	ThemeClass  string
	// Reference and Skin are resolved, closed-vocabulary document metadata.
	// They never contain raw query values; `none` is emitted explicitly.
	Reference string
	Skin      string
	// Contract is gelium (default floors) or source (third-party-faithful density).
	Contract string
	Scheme   string
	// Provenance is the article provenance line (version, license, source,
	// dates) rendered inside the article on component pages only.
	Provenance *provenanceView
	// DataTheme is the optional data-theme attribute on <html> (light|dark).
	// Empty leaves OS prefers-color-scheme in control via theme CSS media.
	DataTheme string
	Nav       []navLink
	// DocsNav enables the two-pane docs shell when non-nil (docs + components).
	DocsNav *docsNavView
	// OnThisPage is the server-built TOC of the rendered article (h2/h3 with
	// goldmark auto heading IDs). Rendered as the sticky right rail on shell
	// routes; empty on pages without headings.
	OnThisPage []tocEntry
	// PrevNext is the previous/next pagination across the docs IA (GOV.UK
	// pattern), derived from the SAME ordered model as the sidebar. nil on
	// non-shell routes and on first/last IA boundaries.
	PrevNext *prevNextView
	// RecipeSwitcher is the two-select native GET control. It is compact in
	// chrome and full on the gallery and Accordion documentation page.
	RecipeSwitcher *recipeSwitcherView
	// ThemeSwitcher is a retired compatibility view kept so older internal
	// callers compile; templates intentionally no longer render it.
	ThemeSwitcher *themeSwitcherView
	// SchemeSwitcher is the native Light/Dark switch (docs topbar / site header).
	SchemeSwitcher *schemeSwitcherView
	// Landing enables the marketing home composition (hero, features, recipes).
	// When non-nil, layout renders the landing template instead of Markdown prose.
	Landing *landingView
	// Blog enables the blog space (index + posts). When non-nil, layout
	// renders the blog template — a separate surface from the docs shell.
	Blog       *blogView
	Banner     *bannerView
	Breadcrumb *breadcrumbView
	Footer     *footerView
	// Examples renders the pilot "## Examples" section (Base UI pattern):
	// live demos rendered through the real component partials, each paired
	// with the actual Go template invocation that produced it. Nil on
	// non-pilot pages so the layout emits nothing.
	Examples []exampleBlock
	// APIRef renders the pilot "## API reference" table from the component's
	// real view-struct fields. Nil on non-pilot pages.
	APIRef               *apiRefView
	Error                *errorStateView
	InlineAlert          *inlineAlertView
	CTA                  *buttonView
	Buttons              []buttonView
	TextFields           []textFieldView
	ValidationForm       *validationFormView
	Dialog               *dialogView
	DialogConfirm        *dialogConfirmView
	Toasts               []toastView
	ToastDemo            *toastDemoView
	ElevationDemo        *elevationDemo
	FocusRingDemo        *focusRingDemo
	IconDemo             *iconDemo
	DividerDemo          *dividerDemo
	CardDemo             *cardDemo
	BadgeDemo            *badgeDemo
	CheckboxDemo         *checkboxDemo
	RadioDemo            *radioDemo
	SwitchDemo           *switchDemo
	SelectDemo           *selectDemo
	SelectMenuDemo       *selectMenuDemo
	SelectPopup          *selectPopupExperiment
	SliderDemo           *sliderDemo
	ProgressDemo         *progressDemo
	IconButtonDemo       *iconButtonDemo
	FabDemo              *fabDemo
	ListDemo             *listDemo
	ChipsDemo            *chipsDemo
	TabsDemo             *tabsDemo
	NavigationBarDemo    *navigationBarDemo
	NavigationTabDemo    *navigationTabDemo
	SegmentedButtonDemo  *segmentedButtonDemo
	MenuDemo             *menuDemo
	NavigationDrawerDemo *navigationDrawerDemo
	DataTableDemo        *dataTableDemo
	TooltipDemo          *tooltipDemo
	Newsletter           *newsletterView
	ThemeGallery         *themeGalleryView
	AccordionDemo        *accordionView
}

type server struct {
	templates *template.Template
	markdown  goldmark.Markdown
	assets    fs.FS

	// recipeAdminAuthorize is supplied by the consumer in a real application.
	// The documentation demo uses an explicit allow-all policy so its recipes
	// remain runnable without inventing authentication.
	recipeAdminAuthorize recipeAdminAuthorizer
}

// buildTemplates parses the merged template set: the site shell/chrome first
// (site/web), then the component library (lib). Template names are disjoint by
// construction (registry_sync enforces the boundary), so the merge is a plain
// superset; a collision would fail loudly here instead of silently shadowing.
func renderRecipeOptions(options interface{}) template.HTML {
	var b strings.Builder
	appendOption := func(value, label string, selected bool) {
		b.WriteString(`<option value="`)
		b.WriteString(template.HTMLEscapeString(value))
		b.WriteString(`"`)
		if selected {
			b.WriteString(` selected`)
		}
		b.WriteString(`>`)
		b.WriteString(template.HTMLEscapeString(label))
		b.WriteString(`</option>`)
	}
	switch values := options.(type) {
	case []profileOptionView:
		for _, option := range values {
			appendOption(option.Value, option.Label, option.Selected)
		}
	case []recipeOptionView:
		for _, option := range values {
			appendOption(option.Value, option.Label, option.Selected)
		}
	}
	return template.HTML(b.String())
}

func htmlCloseTag() template.HTML {
	return template.HTML("</html>")
}

func htmlRootAttrs(v pageView) template.HTMLAttr {
	lang := v.Meta.Lang
	if lang == "" {
		lang = "en"
	}
	reference := v.Reference
	if reference == "" {
		reference = "none"
	}
	skin := v.Skin
	if skin == "" {
		skin = "none"
	}
	contract := v.Contract
	if contract == "" {
		contract = "gelium"
	}
	scheme := v.Scheme
	if scheme == "" {
		scheme = "system"
	}
	attrs := ` lang="` + template.HTMLEscapeString(lang) + `" class="` + template.HTMLEscapeString(v.ThemeClass) + `" data-gelium-reference="` + template.HTMLEscapeString(reference) + `" data-gelium-skin="` + template.HTMLEscapeString(skin) + `" data-gelium-contract="` + template.HTMLEscapeString(contract) + `" data-gelium-scheme="` + template.HTMLEscapeString(scheme) + `"`
	if v.DataTheme != "" {
		attrs += ` data-theme="` + template.HTMLEscapeString(v.DataTheme) + `"`
	}
	return template.HTMLAttr(attrs)
}

func buildTemplates() *template.Template {
	tmpl := template.New("geliumui").Funcs(template.FuncMap{
		"htmlCloseTag":        htmlCloseTag,
		"htmlRootAttrs":       htmlRootAttrs,
		"renderRecipeOptions": renderRecipeOptions,
		// themePreloadFonts emits self-hosted font <link rel="preload">
		// tags for the resolved theme class (empty string when the theme
		// ships no fonts). AssetsVersion is threaded through so font hrefs
		// participate in the same cache-busting as app.css.
		"themePreloadFonts": func(themeClass, version string) template.HTML {
			return themePreloadFonts(themeClass, version)
		},
	})
	template.Must(tmpl.ParseFS(webassets.Assets, "templates/*.html"))
	template.Must(tmpl.ParseFS(lib.LibAssets, "templates/*.html"))
	return tmpl
}

// New builds the Gelium UI documentation HTTP handler from embedded assets.
// The demo deliberately uses an explicit allow-all policy; consumers should
// provide their own authorization boundary around the recipe handlers.
func New() http.Handler {
	return newWithRecipeAdminAuthorizer(recipeAdminAllowAll)
}

func newWithRecipeAdminAuthorizer(authorize recipeAdminAuthorizer) http.Handler {
	templates := buildTemplates()
	s := &server{
		templates: templates,
		markdown: goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				highlighting.NewHighlighting(
					highlighting.WithStyle("monokai"),
					highlighting.WithFormatOptions(html.WithClasses(true)),
				),
			),
			// Auto heading IDs power the "On this page" rail anchors. The IDs
			// land on the heading nodes during parsing, so the TOC builder
			// reads them from the AST instead of duplicating the algorithm.
			goldmark.WithParserOptions(parser.WithAutoHeadingID()),
			goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()),
		),
		assets:               webassets.Assets,
		recipeAdminAuthorize: authorize,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /sitemap.xml", s.sitemap)
	mux.HandleFunc("GET /components.json", s.componentsJSON)
	mux.HandleFunc("GET /robots.txt", s.robots)
	mux.HandleFunc("GET /llms.txt", s.llmsTxt)
	mux.HandleFunc("GET /llms-ux.txt", s.llmsUXTxt)
	mux.HandleFunc("GET /{$}", s.home)
	mux.HandleFunc("GET /docs", s.docsIndex)
	mux.HandleFunc("GET /docs/patterns", s.docsPatterns)
	mux.HandleFunc("GET /docs/information-architecture", s.docsInformationArchitecture)
	mux.HandleFunc("GET /docs/screens", s.docsScreens)
	mux.HandleFunc("GET /docs/page-section-architecture", s.docsPageSectionArchitecture)
	mux.HandleFunc("GET /docs/section-references", s.docsSectionReferences)
	mux.HandleFunc("GET /docs/section-references/{id}", s.docsSectionReferenceDetail)
	mux.HandleFunc("GET /docs/journeys", s.docsJourneys)
	mux.HandleFunc("GET /docs/data-display", s.docsDataDisplay)
	mux.HandleFunc("GET /docs/admin-recipes", s.docsAdminRecipes)
	mux.HandleFunc("GET /docs/feedback", s.docsFeedback)
	mux.HandleFunc("GET /docs/density", s.docsDensity)
	mux.HandleFunc("GET /docs/motion", s.docsMotion)
	mux.HandleFunc("GET /docs/ui-definition-of-done", s.docsUIDefinitionOfDone)
	mux.HandleFunc("GET /docs/agent-workflow", s.docsAgentWorkflow)
	mux.HandleFunc("GET /docs/templates/product", s.docsTemplateProduct)
	mux.HandleFunc("GET /docs/templates/design", s.docsTemplateDesign)
	mux.HandleFunc("GET /docs/choose-the-right-control", s.docsChooseTheRightControl)
	mux.HandleFunc("GET /docs/forms", s.docsForms)
	mux.HandleFunc("GET /docs/compare", s.docsCompare)
	mux.HandleFunc("GET /docs/performance", s.docsPerformance)
	mux.HandleFunc("GET /docs/responsive", s.docsResponsive)
	mux.HandleFunc("GET /docs/media", s.docsMedia)
	mux.HandleFunc("GET /docs/themes", s.docsThemes)
	mux.HandleFunc("GET /docs/themes/gallery", s.docsThemeGallery)
	mux.HandleFunc("GET /docs/tokens", s.docsTokens)
	mux.HandleFunc("GET /docs/typography", s.docsTypography)
	mux.HandleFunc("GET /docs/spacing", s.docsSpacing)
	mux.HandleFunc("GET /docs/colors", s.docsColors)
	mux.HandleFunc("GET /docs/server-contracts", s.docsServerContracts)
	mux.HandleFunc("GET /docs/accessibility", s.docsAccessibility)
	mux.HandleFunc("GET /docs/principles", s.docsPrinciples)
	mux.HandleFunc("GET /docs/browser-support", s.docsBrowserSupport)
	mux.HandleFunc("GET /docs/content-style", s.docsContentStyle)
	mux.HandleFunc("GET /docs/acknowledgments", s.docsAcknowledgments)
	mux.HandleFunc("GET /docs/contributing", s.docsContributing)
	mux.HandleFunc("GET /docs/changelog", s.docsChangelog)
	mux.HandleFunc("GET /docs/roadmap", s.docsRoadmap)
	mux.HandleFunc("GET /docs/seo", s.docsSEO)
	mux.HandleFunc("GET /docs/aeo", s.docsAEO)
	// Blog space: separate surface with its own look. /blog is the index;
	// /blog/{slug} renders a post from the registry (unknown slugs 404).
	mux.HandleFunc("GET /blog", s.blogIndex)
	mux.HandleFunc("GET /blog/{slug}", s.blogPost)
	for _, r := range componentRoutes() {
		r := r
		mux.HandleFunc("GET "+r.Path, func(w http.ResponseWriter, req *http.Request) {
			r.Handler(s, w, req)
		})
	}
	mux.HandleFunc("POST /examples/text-field/validate", s.validateTextField)
	mux.HandleFunc("POST /examples/toast/demo", s.toastDemo)
	mux.HandleFunc("POST /examples/select/menu", s.selectMenu)
	mux.HandleFunc("POST /examples/select/popup", s.selectPopup)
	mux.HandleFunc("GET /components/dialog/confirm", s.dialogConfirm)
	mux.HandleFunc("POST /components/dialog/confirm", s.dialogConfirmPost)
	mux.HandleFunc("POST /examples/chips/remove", s.chipsRemoveDemo)
	mux.HandleFunc("POST /examples/data-table/refresh", s.dataTableRefreshDemo)
	mux.HandleFunc("GET /examples/newsletter", s.newsletterExample)
	mux.HandleFunc("POST /examples/newsletter", s.newsletterSubscribe)
	mux.HandleFunc("GET /demo/whatsapp", s.whatsAppDemo)
	mux.HandleFunc("GET /demo/whatsapp/admin", s.whatsAppAdmin)
	mux.HandleFunc("POST /demo/whatsapp/send", s.whatsAppSend)
	mux.HandleFunc("POST /demo/whatsapp/send-template", s.whatsAppSendTemplate)
	mux.HandleFunc("POST /demo/whatsapp/typing", s.whatsAppTyping)
	mux.HandleFunc("POST /demo/whatsapp/read", s.whatsAppRead)
	mux.HandleFunc("POST /demo/whatsapp/admin/webhook", s.whatsAppWebhookSave)
	// Admin Resource screen recipe (Phase G): the list shares its path with the
	// create action (GET renders, POST mutates), the edit/delete routes pair a
	// GET form/confirm page with a POST mutation, and refresh is POST-only.
	mux.HandleFunc("GET /recipes/admin-resource", s.recipeAdminResourceList)
	mux.HandleFunc("GET /recipes/admin-dashboard", s.recipeAdminDashboard)
	mux.HandleFunc("POST /recipes/admin-resource", s.recipeAdminResourceCreate)
	mux.HandleFunc("GET /recipes/admin-resource/new", s.recipeAdminResourceNew)
	mux.HandleFunc("GET /recipes/admin-resource/import", s.recipeAdminResourceImport)
	mux.HandleFunc("POST /recipes/admin-resource/import", s.recipeAdminResourceImportCSV)
	mux.HandleFunc("GET /recipes/admin-resource/{id}", s.recipeAdminResourceDetail)
	mux.HandleFunc("GET /recipes/admin-resource/{id}/tasks", s.recipeAdminResourceTasks)
	mux.HandleFunc("GET /recipes/admin-resource/{id}/activity", s.recipeAdminResourceActivity)
	mux.HandleFunc("POST /recipes/admin-resource/{id}/transition", s.recipeAdminResourceTransition)
	mux.HandleFunc("POST /recipes/admin-resource/{id}/tasks", s.recipeAdminResourceTaskCreate)
	mux.HandleFunc("GET /recipes/admin-resource/{id}/edit", s.recipeAdminResourceEdit)
	mux.HandleFunc("POST /recipes/admin-resource/{id}/edit", s.recipeAdminResourceUpdate)
	mux.HandleFunc("GET /recipes/admin-resource/{id}/delete", s.recipeAdminResourceDeleteConfirm)
	mux.HandleFunc("POST /recipes/admin-resource/{id}/delete", s.recipeAdminResourceDelete)
	mux.HandleFunc("GET /recipes/admin-resource/bulk-delete", s.recipeAdminResourceBulkDeleteConfirm)
	mux.HandleFunc("POST /recipes/admin-resource/bulk-delete", s.recipeAdminResourceBulkDelete)
	mux.HandleFunc("POST /recipes/admin-resource/refresh", s.recipeAdminResourceRefresh)
	// Ops Queue screen recipe (Phase G): the list is a GET page (filter by
	// status/kind + server-side pagination), every transition is a POST+303
	// mutation with a persistent success banner, and refresh is POST-only.
	mux.HandleFunc("GET /recipes/search", s.recipeGlobalSearch)
	mux.HandleFunc("GET /recipes/ops-queue", s.recipeOpsQueueList)
	mux.HandleFunc("GET /recipes/ops-queue/export.csv", s.recipeOpsQueueExport)
	mux.HandleFunc("GET /recipes/ops-queue/{id}", s.recipeOpsQueueDetail)
	mux.HandleFunc("POST /recipes/ops-queue/{id}/advance", s.recipeOpsQueueAdvance)
	mux.HandleFunc("POST /recipes/ops-queue/{id}/dequeue", s.recipeOpsQueueDequeue)
	mux.HandleFunc("POST /recipes/ops-queue/refresh", s.recipeOpsQueueRefresh)
	// Public/Social Feed screen recipe (Phase G): the feed is a GET page (view
	// tabs + server-side pagination), reactions are POST+303 with a flash
	// toast, and refresh is POST-only.
	mux.HandleFunc("GET /recipes/public-feed", s.recipePublicFeedList)
	mux.HandleFunc("GET /recipes/rich-article", s.recipeRichArticle)
	mux.HandleFunc("POST /recipes/public-feed/{id}/react", s.recipePublicFeedReact)
	mux.HandleFunc("POST /recipes/public-feed/refresh", s.recipePublicFeedRefresh)
	mux.HandleFunc("GET /static/fonts/{file}", s.fontAsset)
	mux.HandleFunc("GET /static/{path...}", s.staticAsset)
	// 404 catch-all: any unknown GET path falls back to the styled ERROR STATE
	// page (the mux gives the more specific patterns above priority). Post-only
	// routes register a GET 405 companion below so a GET to them keeps the
	// mux's modern method-mismatch semantics instead of being swallowed by the
	// catch-all.
	for _, path := range postOnlyPaths() {
		mux.HandleFunc("GET "+path, methodNotAllowed)
	}
	mux.HandleFunc("GET /{path...}", s.notFound)
	return themeQueryMiddleware(mux)
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

// robotsTxt is the fixed robots.txt policy. The public docs site is fully
// crawlable except the demo, example and recipe surfaces (noindex or form
// flows), and the sitemap is advertised so crawlers discover every indexable
// page in one hop (SEO contract §4). A var because the sitemap URL derives
// from the startup-resolved siteBaseURL.
var robotsTxt = "User-agent: *\n" +
	"Allow: /\n" +
	"Disallow: /demo/\n" +
	"Disallow: /examples/\n" +
	"Disallow: /recipes/\n" +
	"Sitemap: " + siteBaseURL + "/sitemap.xml\n"

// robots serves the robots.txt policy over plain text.
func (s *server) robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(robotsTxt))
}

// llmsTxt serves /llms.txt — the agent-oriented project brief (llmstxt.org
// style). Content lives in the embedded static/llms.txt so it ships with the
// docs binary and stays one file to edit.
func (s *server) llmsTxt(w http.ResponseWriter, r *http.Request) {
	s.servePlainStatic(w, r, "static/llms.txt")
}

// llmsUXTxt serves /llms-ux.txt — dense screen/feedback decision tables for
// agents (companion to the human handbook at /docs/screens and /docs/feedback).
func (s *server) llmsUXTxt(w http.ResponseWriter, r *http.Request) {
	s.servePlainStatic(w, r, "static/llms-ux.txt")
}

// servePlainStatic writes an embedded text asset as text/plain with no-cache.
func (s *server) servePlainStatic(w http.ResponseWriter, r *http.Request, name string) {
	asset, err := fs.ReadFile(s.assets, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(asset)
}

// sitemapURL is one <url> entry in the server-generated sitemap.
type sitemapURL struct {
	XMLName xml.Name `xml:"url"`
	Loc     string   `xml:"loc"`
}

// urlset is the sitemap document emitted at /sitemap.xml with the sitemaps.org
// namespace every crawler expects.
type urlset struct {
	XMLName xml.Name     `xml:"urlset"`
	Xmlns   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

// indexabilityPolicy is the one route policy consumed by both robots metadata
// and sitemap generation. Interactive/demo/stateful surfaces are never indexed.
func indexabilityPolicy(routePath string) string {
	if strings.HasPrefix(routePath, "/demo/") || strings.HasPrefix(routePath, "/examples/") ||
		strings.HasPrefix(routePath, "/recipes/") || routePath == "/components/dialog/confirm" {
		return "noindex, nofollow"
	}
	return "index, follow"
}

func isIndexablePath(routePath string) bool { return indexabilityPolicy(routePath) == "index, follow" }

// sitemapPaths is the canonical list of public pages filtered through the same
// indexability policy used by resolveMeta.
func sitemapPaths() []string {
	paths := []string{"/", "/docs", "/docs/patterns", "/docs/seo", "/docs/aeo", "/docs/templates/product", "/docs/templates/design", "/blog"}
	for _, l := range handbookNavLinks() {
		paths = append(paths, l.Path)
	}
	paths = append(paths, publishedSectionRefPaths()...)
	for _, r := range componentRoutes() {
		paths = append(paths, r.Path)
	}
	for _, p := range blogPosts {
		paths = append(paths, "/blog/"+p.Slug)
	}
	filtered := make([]string, 0, len(paths))
	for _, p := range paths {
		if isIndexablePath(p) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// sitemap serves the server-generated sitemap.xml built from the route
// registry: one <url><loc> per indexable page, absolute URLs, no lastmod
// (static content). Content-Type is application/xml per the sitemap protocol.
func (s *server) sitemap(w http.ResponseWriter, r *http.Request) {
	urls := make([]sitemapURL, 0, len(sitemapPaths()))
	for _, p := range sitemapPaths() {
		urls = append(urls, sitemapURL{Loc: siteBaseURL + p})
	}
	doc, err := xml.MarshalIndent(urlset{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}, "", "  ")
	if err != nil {
		http.Error(w, "sitemap unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	_, _ = w.Write(append([]byte(xml.Header), doc...))
}

// postOnlyPaths are the routes that only accept POST. A GET to any of them
// must stay a 405 (the mux's modern method-mismatch semantics) even after the
// catch-all 404, so each path also registers a GET methodNotAllowed companion.
func postOnlyPaths() []string {
	return []string{
		"/examples/text-field/validate",
		"/examples/toast/demo",
		"/examples/select/menu",
		"/examples/chips/remove",
		"/examples/data-table/refresh",
		"/demo/whatsapp/send",
		"/demo/whatsapp/send-template",
		"/demo/whatsapp/typing",
		"/demo/whatsapp/read",
		"/demo/whatsapp/admin/webhook",
		"/recipes/admin-resource/refresh",
		"/recipes/ops-queue/refresh",
		"/recipes/public-feed/refresh",
	}
}

// methodNotAllowed answers a GET to a POST-only route with the net/http 405
// Method Not Allowed response, preserving the Allow header the ServeMux would
// otherwise emit.
func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "POST")
	http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
}

// notFound serves the styled ERROR STATE page for unknown routes, replacing
// the plain-text net/http default. It renders the full docs layout with the
// Error slot set, the real 404 status, and route metadata resolved for the
// requested path (SEO §16).
func (s *server) notFound(w http.ResponseWriter, r *http.Request) {
	s.renderErrorPage(w,
		r,
		http.StatusNotFound,
		"Page not found",
		"The page you are looking for does not exist or has moved.",
		true,
		"/",
		"Back to home",
		r.URL.Path,
	)
}

// renderErrorPage renders the full docs layout with the ERROR STATE slot set
// and the real HTTP status. It is the page-level failure path for the 404
// catch-all and resource 500s; the layout template is already parsed, so a
// failed template exec still falls back to a minimal plain response. Meta is
// populated through resolveMeta for the failing routePath so error pages carry
// description, canonical, robots and OG tags per the route contract (SEO §16).
func (s *server) renderErrorPage(w http.ResponseWriter, r *http.Request, status int, title, body string, retry bool, href, label, routePath string) {
	var page bytes.Buffer
	data := pageView{
		Title:         title,
		ThemeClass:    themeClass(""),
		Nav:           navLinks(),
		Footer:        defaultFooter(),
		AssetsVersion: lib.AssetsVersion,
		Error: &errorStateView{
			StatusCode: status,
			Title:      title,
			Body:       body,
			Retry:      retry,
			Href:       href,
			Label:      label,
		},
	}
	data.Meta = resolveMeta(data, routePath)
	// Non-200 responses must never be indexed: a soft-404 (or a soft-500)
	// that Google crawls undermines the indexability contract (SEO §4).
	if status != http.StatusOK {
		data.Meta.Robots = "noindex, nofollow"
	}
	// Docs shell chrome for shell routes, mirroring renderMarkdownStatus:
	// grouped sidebar + search + theme/scheme switchers. Shell routes get the
	// two-pane chrome instead of the flat legacy header (Nav must be nil so
	// layout.html does not render both navigations on the failure page).
	// r is nil for a few legacy helpers (recipe not-found) whose routes never
	// use the shell; they keep the flat header path unchanged.
	if r != nil {
		selection := applyDocumentSelection(&data, r)
		if usesDocsShell(routePath) {
			execution := accordionExecutionFromRequest(r)
			navigation := navigationSelectionFor(r, routePath)
			nav := docsNavForNavigation(routePath, navigation)
			data.DocsNav = &nav
			data.RecipeSwitcher = recipeSwitcherFor(r, selection, execution, true, "docs-recipe")
			data.SchemeSwitcher = schemeSwitcherForNavigation(selection, execution, navigation)
			data.Nav = nil
		}
	}
	if err := s.templates.ExecuteTemplate(&page, "layout", data); err != nil {
		http.Error(w, "documentation unavailable", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(page.Bytes())
}

// fontAsset serves a self-hosted, embedded webfont file (WOFF2) from the lib
// theme-font bundle, bypassing the site static tree so fonts travel with the
// theme in lib/fonts/ (self-hosted, no CDN). Only whitelisted theme fonts
// referenced by availableThemes are servable: a font path is allowed only if
// some allowlisted theme ships a font with that exact filename (this keeps the
// /static/fonts namespace closed and prevents arbitrary asset reads).
func (s *server) fontAsset(w http.ResponseWriter, r *http.Request) {
	file := r.PathValue("file")
	if file == "" || !strings.HasSuffix(file, ".woff2") || !themeShipsFontFile(file) {
		http.NotFound(w, r)
		return
	}
	// fonts/<file> maps 1:1 onto the flat lib embed glob fonts/*.
	asset, err := fs.ReadFile(lib.LibAssets, "fonts/"+file)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "font/woff2")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(asset)
}

// themeShipsFontFile reports whether some allowlisted theme declares a font
// with the given flat filename. Guards the /static/fonts/ namespace against
// arbitrary reads.
func themeShipsFontFile(file string) bool {
	for _, t := range availableThemes {
		for _, f := range t.Fonts {
			if f.File == file {
				return true
			}
		}
	}
	return false
}

func (s *server) staticAsset(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("path")
	asset, err := fs.ReadFile(s.assets, "static/"+name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", staticContentType(name))
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(asset)
}

// staticContentType maps a static asset extension to its Content-Type. The
// static directory is part of the embedded build surface; any asset that
// exists in the embed is servable, never a hardcoded allowlist.
func staticContentType(name string) string {
	switch mime.TypeByExtension(filepath.Ext(name)) {
	case "":
		return "application/octet-stream"
	case "text/javascript":
		return "text/javascript; charset=utf-8"
	case "text/css":
		return "text/css; charset=utf-8"
	case "image/svg+xml", "application/json", "text/xml":
		return mime.TypeByExtension(filepath.Ext(name)) + "; charset=utf-8"
	default:
		return mime.TypeByExtension(filepath.Ext(name))
	}
}

func (s *server) renderMarkdownPage(w http.ResponseWriter, r *http.Request, data pageView, contentPath string) {
	s.renderMarkdownPageStatus(w, r, data, contentPath, http.StatusOK)
}

func (s *server) renderMarkdownPageStatus(w http.ResponseWriter, r *http.Request, data pageView, contentPath string, status int) {
	source, err := fs.ReadFile(s.assets, contentPath)
	if err != nil {
		s.renderErrorPage(w, r, http.StatusInternalServerError, "Something went wrong", "This page could not be loaded. Please try again later.", true, "/", "Back to home", routePathForContent(contentPath))
		return
	}
	s.renderMarkdownStatus(w, r, data, string(source), routePathForContent(contentPath), status)
}

// renderMarkdownPageAt renders an embedded markdown file under an explicit
// PUBLIC route path. Content-only docs pages (handbook) must pass the real
// /docs/* path: the derived /components/* path would break sidebar current
// marking and previous/next pagination (both key off the IA route).
func (s *server) renderMarkdownPageAt(w http.ResponseWriter, r *http.Request, data pageView, contentPath, routePath string) {
	source, err := fs.ReadFile(s.assets, contentPath)
	if err != nil {
		s.renderErrorPage(w, r, http.StatusInternalServerError, "Something went wrong", "This page could not be loaded. Please try again later.", true, "/", "Back to home", routePath)
		return
	}
	s.renderMarkdownStatus(w, r, data, string(source), routePath, http.StatusOK)
}

// renderMarkdown converts an in-memory Markdown string and renders it into the
// docs layout. Used by pages that build their content programmatically (the
// /docs index) as well as by pages served from embedded files. routePath is
// the public route identity used to resolve metadata.
func (s *server) renderMarkdown(w http.ResponseWriter, r *http.Request, data pageView, source, routePath string) {
	s.renderMarkdownStatus(w, r, data, source, routePath, http.StatusOK)
}

// buildTOC walks the markdown AST of source and collects the h2/h3 headings
// with their auto-generated ids (parser.WithAutoHeadingID). The rail links
// point at those anchors; ids are read from the heading nodes, so the TOC can
// never drift from the rendered output.
func (s *server) buildTOC(source string) []tocEntry {
	sourceBytes := []byte(source)
	doc := s.markdown.Parser().Parse(text.NewReader(sourceBytes))
	toc := make([]tocEntry, 0, 8)
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok || (h.Level != 2 && h.Level != 3) {
			return ast.WalkContinue, nil
		}
		var id string
		if v, ok := h.AttributeString("id"); ok {
			if b, ok := v.([]byte); ok {
				id = string(b)
			}
		}
		if id == "" {
			return ast.WalkContinue, nil
		}
		toc = append(toc, tocEntry{ID: id, Text: string(h.Text(sourceBytes)), Level: h.Level})
		return ast.WalkContinue, nil
	})
	return toc
}

// splitLeadingH1 extracts the first level-1 heading (a line starting with
// "# ") from an embedded markdown source and converts it to its own HTML
// block, returning the remaining markdown plus the H1 markup. Pages whose
// markdown does not lead with an H1 (handbook, hub, landing, blog) return the
// source unchanged with an empty H1. The split happens on the markdown line
// so goldmark's auto-heading-id still applies to the rendered H1.
func splitLeadingH1(source string) (rest, h1 string) {
	lines := strings.SplitN(source, "\n", 2)
	first := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(first, "# ") {
		return source, ""
	}
	var buf bytes.Buffer
	if err := (goldmark.New(goldmark.WithParserOptions(parser.WithAutoHeadingID()))).Convert([]byte(first), &buf); err != nil {
		return source, ""
	}
	rest = ""
	if len(lines) == 2 {
		rest = lines[1]
	}
	return rest, buf.String()
}

func (s *server) renderMarkdownStatus(w http.ResponseWriter, r *http.Request, data pageView, source, routePath string, status int) {
	// Demo-first: split the leading H1 out of the markdown so the layout can
	// render title → live demo → body (Base UI/Naive UI order: show the
	// component, then the rules). Non-component pages (no leading H1) keep
	// the whole source in Content as before.
	source, h1 := splitLeadingH1(source)
	data.AssetsVersion = lib.AssetsVersion
	var rendered bytes.Buffer
	if err := s.markdown.Convert([]byte(source), &rendered); err != nil {
		s.renderErrorPage(w, r, http.StatusInternalServerError, "Something went wrong", "This page could not be rendered. Please try again later.", true, "/", "Back to home", routePath)
		return
	}
	if h1 != "" {
		data.H1 = template.HTML(h1) // #nosec G203 -- the H1 is parsed from trusted embedded markdown.
	}
	// On this page: server-built TOC from the FULL source (before the pilot
	// split), so the rail covers the whole article.
	data.OnThisPage = s.buildTOC(source)

	var page bytes.Buffer
	data.Nav = navLinks()
	if data.Footer == nil {
		data.Footer = defaultFooter()
	}
	if data.Breadcrumb == nil && routePath != "/" {
		if label, ok := componentRouteLabel(routePath); ok {
			data.Breadcrumb = componentBreadcrumb(label, routePath)
		} else {
			data.Breadcrumb = defaultBreadcrumb(data.Title)
		}
	}
	// Article provenance (GEO §8, §15): registered component pages render the
	// version, MIT license link, source reference and (when dated) the
	// published/modified ISO dates. The layout guards the whole line behind
	// {{if .Provenance}}, so non-component pages emit nothing.
	if _, ok := componentRouteLabel(routePath); ok {
		slug := strings.TrimPrefix(routePath, "/components/")
		provenance := &provenanceView{
			Version:    docsShellVersion,
			LicenseURL: licenseURL,
			Source:     slug + ".md",
		}
		if dates, ok := docDatesFor(slug); ok {
			provenance.Published = dates.Published
			provenance.Modified = dates.Modified
		}
		data.Provenance = provenance
	}
	data.Meta = resolveMeta(data, routePath)
	// Pilot pages split their markdown so Examples + API reference sit
	// between the intro and the Guidance block (Base UI order: show the
	// component, then the rules). The split happens at "## Guidance": the
	// intro (answer-first lead) renders first, then the server-rendered
	// Examples/API sections, then the rest (Guidance onward) closes the
	// article. Non-pilot pages keep the whole source in Content.
	if len(data.Examples) > 0 {
		rest := ""
		if idx := strings.Index(source, "## Guidance"); idx >= 0 {
			rest = source[idx:]
			source = source[:idx]
		}
		var introHTML bytes.Buffer
		if err := s.markdown.Convert([]byte(source), &introHTML); err != nil {
			s.renderErrorPage(w, r, http.StatusInternalServerError, "Something went wrong", "This page could not be rendered. Please try again later.", true, "/", "Back to home", routePath)
			return
		}
		var restHTML bytes.Buffer
		if rest != "" {
			if err := s.markdown.Convert([]byte(rest), &restHTML); err != nil {
				s.renderErrorPage(w, r, http.StatusInternalServerError, "Something went wrong", "This page could not be rendered. Please try again later.", true, "/", "Back to home", routePath)
				return
			}
		}
		data.Content = template.HTML(introHTML.String())    // #nosec G203 -- markdown is trusted (embedded or generated).
		data.ContentRest = template.HTML(restHTML.String()) // #nosec G203 -- markdown is trusted (embedded or generated).
	} else {
		data.Content = template.HTML(rendered.String()) // #nosec G203 -- markdown is trusted (embedded or generated).
	}
	// Pilot Examples sections render their live demos at the choke point:
	// every page that carries Examples (the three pilot components) gets the
	// demos executed through the real component partials before layout, so
	// the code blocks shown next to each demo can never drift from the
	// partials that produced them. Non-pilot pages pass nil and skip this.
	if len(data.Examples) > 0 {
		examples, err := s.renderExampleDemos(data.Examples)
		if err != nil {
			s.renderErrorPage(w, r, http.StatusInternalServerError, "Something went wrong", "This page could not be loaded. Please try again later.", true, "/", "Back to home", routePath)
			return
		}
		data.Examples = examples
	}
	// Resolve the independent document contract once, before building chrome.
	selection := applyDocumentSelection(&data, r)
	// Docs shell navigation preserves legacy chrome where it arrived as legacy;
	// canonical recipe values only remain inside the recipe/gallery surfaces.
	if usesDocsShell(routePath) {
		execution := accordionExecutionFromRequest(r)
		navigation := navigationSelectionFor(r, routePath)
		nav := docsNavForNavigation(routePath, navigation)
		data.DocsNav = &nav
		fullRecipe := routePath == "/docs/themes/gallery" || routePath == "/components/accordion"
		data.RecipeSwitcher = recipeSwitcherFor(r, selection, execution, !fullRecipe, "docs-recipe")
		data.SchemeSwitcher = schemeSwitcherForNavigation(selection, execution, navigation)
		if pn := prevNextForNavigation(routePath, navigation); pn != nil {
			data.PrevNext = pn
		}
		// Existing breadcrumb compatibility applies only to legacy chrome. Canonical
		// recipe queries never turn ordinary IA anchors into recipe URLs.
		if data.Breadcrumb != nil && navigation.hasLegacyChrome() {
			data.Breadcrumb = breadcrumbWithChrome(data.Breadcrumb, navigation.themeSlug, navigation.scheme)
		}
	}
	if err := s.templates.ExecuteTemplate(&page, "layout", data); err != nil {
		http.Error(w, "documentation unavailable", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(page.Bytes())
}
