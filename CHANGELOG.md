# Changelog

All notable changes to the Gelium UI project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project uses [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Changed

- **Documentation chrome**: the header now uses link-based navigation, keeps GitHub/version together, and preserves behavior, recipe, reference, skin, execution, and appearance selections across navigation.
- **HTMX navigation**: compact recipe changes update the document without a full-page reload, preserve scroll position, and keep hidden fallback submit controls from causing layout shifts.
- **Docs layout**: breadcrumbs now share the content container and remain aligned with the article reading measure.
- **Gelium design discovery protocol**: design-gated work now distinguishes new products, evidence-based redesigns, and explicit visual migrations; records current-state findings, section contracts, Gelium behavior recipes, theme/skin direction, copy/locales, and media before final wireframe approval.
- **Structural wireframe protocol**: design-gated screens now require labelled global shell bands, purpose-bound regions, WIDE/NARROW ASCII layouts, state/recovery annotations, and explicit shell decisions before approval.

### Fixed

- **Dark theme select contrast**: Neubrutalism and Vercel select text, labels, carets, borders, and native option menus now use readable dark-scheme colors.

## [0.6.6] — 2026-09-06

### Added

- **Plain-language product and design brief**: consumer workflows now start with accessible questions about the job, audience, outcome, and scope instead of requiring product or design vocabulary.
- **Expanded consumer templates**: `PRODUCT.md` and `DESIGN.md` now capture screen inventory, navigation, lifecycle, states, accessibility, success criteria, references, and open decisions.

### Changed

- **Agent workflow guidance**: missing or incomplete consumer context files now trigger a plain-language brief, preserving unknowns and assumptions while stopping only for unresolved architectural decisions.
- **Neubrutalism guidance**: package documentation and examples now include the shipped `theme-neubrutalism` theme.
- **UX contract coverage**: detector behavior and docs-site tests cover the new brief and template contracts.

## [0.6.5] — 2026-09-04

### Added

- **Inspiration map in package documentation**: root and npm README files now distinguish design foundations, component/open-code references, visual studies, UX/content research, and Gentle AI workflow influence.

### Changed

- **README accuracy**: the published version, package boundaries, implementation stack, attribution boundaries, and non-affiliation disclaimer are explicit; the full source matrix remains in the docs-site acknowledgments page.

## [0.6.4] — 2026-09-03

### Added

- **Canonical agent routing**: root `AGENTS.md`, discoverable `lib/SKILL.md`, and `lib/skills/00-agent-routing.md` now route work as `direct-exempt`, `delegated-direct`, `design-gated`, `escalate`, or `full-sdd` before loading downstream Gelium guidance.
- **Component implementation skill**: `lib/skills/14-component-implementation.md` replaces the legacy root prompt as the operational contract for new components, covering ownership, platform-first decisions, no-JS, TDD, verification, and handoff states.
- **Route handoff and startup discovery**: `gelium-preflight route` provides machine-readable next actions, while `lib/scripts/agent-start.sh` optionally refreshes a Gentle AI project registry.

### Changed

- **Agent guidance projections**: `site/web/static/llms.txt` and `llms-ux.txt` are now copied from `lib/` during the build and protected by equality tests.
- **Component workflow references**: active documentation points to skill 14; `AI-COMPONENT-IMPLEMENTER-PROMPT.md` was removed after its references were migrated.
- **Package metadata**: the docs site consumes `gelium-ui@0.6.4`; HTMX tests and asset cache-busting expectations follow the actual `4.0.0` runtime and `0.6.4` asset version.

## [0.6.3] — 2026-09-03

### Fixed

- **Portable reference fichas were stubs**: 0.6.2 shipped `references/` but most Markdown files were ~6-line placeholders. Agents still needed the monorepo for usable feed/shell/detail guidance. Fichas are now full structural notes (especially `REF-SOCIAL-FEED`) so npm-only consumers can run skill 13 without docs or git checkout.
- **Reference quality contract**: `TestPortableReferenceCatalogQuality` requires catalog IDs, on-disk fichas, minimum substantive size, social-feed filter content, and forbids shipping screenshot binaries in the package.

### Changed

- **Skill 13**: package `references/catalog.json` + fichas are the authority for npm consumers; monorepo screenshots/`/docs/section-references` are optional supplements only.

## [0.6.2] — 2026-09-03

### Added

- **Portable agent reference catalog**: npm consumers now receive `references/catalog.json`, section fichas, component-reference guidance, and social-feed pattern notes without requiring the docs site or monorepo.
- **Package completeness contract**: tests and pack verification require the portable reference catalog to be declared and included in the published files.

### Changed

- **Reference resolution**: `skills/13-reference-selection.md` reads the npm catalog first and uses docs-site fichas only as a richer supplement.

## [0.6.1] — 2026-09-03

### Added

- **Semantic reference selection for agents**: matching section/component references are consulted before architecture and wireframes; visual affordances are fished before product filtering and recorded as existing contract, B, C, fork, or rejected.
- **Social-feed visual reference set**: development-only responsive captures, provenance manifest, and audit notes for X-like and Reddit-like feed patterns.

### Changed

- **Gelium decision pack**: added identity/account grouping, action-versus-navigation, constant-state, visibility-semantics, detail-entry, and responsive-reference lenses.
- **Reference selection skill**: added `skills/13-reference-selection.md` and synchronized `AGENTS.md`, `SKILLS.md`, and static agent guidance.

## [0.6.0] — 2026-08-25

### Added

- **Agent guidance hardening**: step-0 artifacts gate (no `PRODUCT.md`/`DESIGN.md` → STOP and ask the user for job, SURFACE mode, and theme+skin before generating UI), registry-first page-shell rule (compose `ui-container`/nav primitives; custom shell CSS is spacing/width only), and a portable DoD verifier (`lib/scripts/ux-detect.sh`, shipped in the npm package) that mechanically greps for media-query dark overrides, hand-rolled page shells, body `overflow-x:hidden`, one-off color literals, and forms without validation-summary. Golden rules 8–9 added to `AGENTS.md`.
- **Launch blog post** (`/blog/introducing`): "Introducing Gelium UI" — the announcement post covering the problem, the 0-JS contract, token themes, and who the library is for; registered in the blog registry with its content contract.
- **Component catalog grown from 28 to 47 pages**: the composition, content, and feedback primitives that the handbook named but nobody could reach now have `/components/*` pages — hero, avatar, breadcrumb, footer, pagination, section-heading, feature-card, split, image, media, video, newsletter, banner, inline-alert, callout, skeleton, empty-state, error-state, validation-summary. Each page carries a live raw-HTML specimen, When to use / When not to use guidance, anatomy, and a11y notes; sidebar, search index, sitemap, and prev/next auto-derive from the one nav model.
- **No-JS docs search actually searches**: `GET /docs?q=` now renders server-built results from the same nav model the client-side index is built from (the documented 0-JS fallback previously reloaded the hub without querying).
- **`/static/{path...}`**: the static route now serves multi-segment assets (media fixtures, subfolders).
- **Media fixtures**: `empty-audio.ogg` (Opus) + `editorial-placeholder.svg` — the `/docs/media` demo no longer 404s.
- **Template pages in sitemap**: `/docs/templates/product` + `/docs/templates/design` were indexable but invisible to crawler discovery; both are now in the sitemap.

### Changed

- **Error pages render through the docs shell** on `/docs*` and `/components*` routes (grouped sidebar + search + theme/scheme switchers) instead of the flat legacy header, and force `noindex, nofollow` on any non-200 response (removes the soft-404 indexing risk).
- **Specimen spacing**: full-bleed component specimens now live in a `.specimen-block` docs surface (2rem vertical rhythm, card look) — component styles reset their own margins by design, so the container owns the air between demo and prose.
- **`ui-media` figure rhythm**: figcaption and the transcript link own their spacing below the full-width audio player (8px/12px/40px measured), and the same rule ships in the library CSS.
- **llms.txt synced**: image + media in the component list, rich-article among the recipes.

## [0.5.3] — 2026-08-17

> 0.5.1 and 0.5.2 were packaging continuity steps (npm workspaces scaffolding, publish-auth hardening); this entry documents the series as one release.

### Added

- **Publishable `lib/` package** (`gelium-ui` on npm): npm workspaces monorepo, exports map (`./styles/*`, `./themes/*`, `./templates/*`, `./js/*`, `./dist/*`), cache-busting centralized in `lib/version.go`, dogfood import of the package by the docs site.
- **Server-rendered Media component** (`media`): audio + transcript + embed templates with a zero-JS contract, accessible fallback copy, embed consent boundary, and media style contract.
- **Coherent SEO/AEO metadata**: llms.txt + llms-ux.txt served, JSON-LD (CollectionPage, TechArticle, BreadcrumbList), canonical/robots/sitemap coherence.

## [0.5.0] — 2026-08-16

### Added

- **On this page rail** (`/docs/*`): server-built TOC reading real goldmark AST heading ids (`parser.WithAutoHeadingID`), nested h2/h3, sticky ≥64rem, hidden on mobile, zero algorithm duplication.
- **Prev/Next pagination**: GOV.UK pattern driven by the same `orderedDocsNav` model as the sidebar; no Previous on the first page / no Next on the last; honors theme/scheme.
- **Sidebar current**: the active entry stays marked while navigating.
- **Optimistic theme/scheme toggle**: the class/`data-theme` applies instantly (both theme CSS already in the bundle), round-trip persists, `hx-swap="none"` + `requestSubmit`, reconciled with the server in `htmx:before:swap` (before paint). No JS = classic submit.
- **Inline SVG icons** in the theme and scheme switchers.
- **Centered content**: single `min(65ch,100%)` column with `margin-inline:auto` for article, breadcrumbs, prev/next (crumbs/banner moved inside `main`).

### Changed

- **HTMX 2.0.10 → 4.0.0-beta6**: namespaced events (`htmx:before:swap`, `htmx:before:request`, `htmx:response:error`, `htmx:error`), `ctx.response`/`ctx.text` instead of `xhr`, `hx-boost:inherited` + `innerMorph` + `morph-afterswap.js`.
- **Chrome sync moved to `before:swap`** (was `after:swap`): removes the one-frame flash when navigating dark/light.
- **`staticAsset`**: serves any embedded asset with extension-based Content-Type (the 3-file allowlist removed).
- **Cache-busting to `?v=0.5.0`** on all assets.
- **Standalone recipes** now apply theme+scheme (`applyRequestChrome` + `data-theme`) — the sidebar no longer flips dark→light.
- **Topbar (brand/Blog/changelog)**: preserves the chrome query with scheme-first order (`?scheme=dark&theme=basecoat`).

### Fixed

- `search.js` and `morph-afterswap.js` were 404 + `text/plain` rejected (hardcoded allowlist; search.js was never served on `main`).
- 3 legacy htmx-2 listeners (`htmx:beforeRequest`, `htmx:responseError`, `htmx:sendError`) broke scroll save and error toasts.
- 11 handbook handlers used content routes (`/components/handbook-*`) instead of `/docs/*` — sidebar current and Prev/Next returned nil.

### Enforced (new contract tests)

- `htmx4_navigation_test.go`: deprecated htmx-2 API list rejected, `htmx:before:swap`/`ctx.response`/`ctx.text` contracts, 0.5.0 cache-busting.
- `on_this_page_test.go`: rail and Prev/Next on served pages.
- `recipe_chrome_test.go`: recipes honor `?scheme=`.
- `styles_on_this_page_test.go`: `margin:0auto` contract (compactCSS).
- `docs_shell_test.go`: `TestProduceSameChromeURLs`, hrefs with `&amp;`.
- `server_test.go`: static table with `search.js`/`morph-afterswap.js`.

## [0.4.0] — 2026-08-15

### Added

- **Content style guide** (`/docs/content-style`): copy rules for errors (recognize + recover, never blame the user — NNG), toasts (verb + result), empty states (what it is / why it's empty / what to do), banners, validation summary. Editorial writing patterns: plain English, active voice, no "please", AP Style.
- **Screen-reading rules**: "Reading on screen" section with NNG eye-tracking research (most people scan, not read), F pattern (important content top-left), web copy shorter than print. "Paragraphs and sentences": 2–4 sentence paragraphs, sentences ≤ 25 words, inverted pyramid, lists over prose.
- **Copy length contract**: `TestComponentPagesKeepSentencesUnder25Words` — no component page may have sentences over 25 words (179 rewritten).
- **Copy contract test**: `TestRecipeErrorCopyUsesActionPattern` + `TestRecipeEmptyStatesCarryActionLanguage` — errors with an action pattern, actionable empty states.
- **Acknowledgments page** (`/docs/acknowledgments`): explicit credit for all inspiration sources — Material Design 3, USWDS, GOV.UK, Mozilla Protocol, Base UI, Basecoat UI, Naive UI, Name That UI, Material Web, shadcn/templ — what was taken, how it was adapted, and the license.
- **Information architecture page** (`/docs/information-architecture`): concept-before-reference rule, page-addition criteria (who navigates / what task / concept-vs-reference) and an agent prompt for LLMs auditing docs IA.
- **Choose the right control page** (`/docs/choose-the-right-control`): decision table for picking the correct input component (Radio vs Select vs Checkbox vs Switch vs Slider vs Text field vs Menu) with rules of thumb.
- **Guidance sections on all 28 component pages**: When to use / When not to use / Usability / Accessibility, cross-linked to the decision page.
- **6-page handbook**: Themes, Tokens, Server contracts, Accessibility, Design principles, Information architecture — concept before reference.
- **Client-side docs search**: `#docs-search-index` JSON + `search.js`, 0-JS fallback to GET `/docs?q=`.
- **Native theme switcher** (`<select name="theme">`) and **real scheme switcher** (`<input type="checkbox" role="switch">`) — both 0-JS with progressive enhancement.
- **GitHub link in the docs topbar.**

### Changed

- **Wire contract migrated**: `loom:*` / `X-Loom-*` → `gelium:*` / `X-Gelium-*` (owner decision: the project is new, migrate now). The wire prefix now matches the product name.
- **Prose readability**: measure 48rem (~90 chars) → **65ch**, `text-wrap: pretty` (no orphans), `text-wrap: balance` on headings, `hyphens: auto`, progressive `text-box-trim`.
- **Vertical rhythm**: breadcrumb → title, provenance → title, h2 → h3 with bidirectional margins (previously one-directional).
- **AA contrast**: Basecoat-light `fg-muted` corrected (4.35:1 → 4.75:1) for WCAG AA; contract test in both themes.
- **Unified error copy** to the action pattern: "Name is required." → "Enter the project name."; "Choose a valid status." → "Choose a status from the list." (previously two voices).
- **Public feed empty state**: now with a CTA ("Follow more people to fill this feed.").
- **Sidebar hierarchy**: Handbook moved to position 2 (after Getting started, before components).
- **Rename residues**: "Loom UI" → "Gelium UI" in 4 root docs (MATERIAL-WEB-PROGRESS, prompts, roadmap).

### Fixed

- `dependency-metadata.md` stale claim about Phase I.
- Roadmap missing completed-phase markers (Phase I and J now DONE).

### Enforced (new contract tests)

- `styles_readability_test.go`: 65ch, text-wrap, hyphens, text-box-trim, line-height ≥ 1.6, vertical rhythm, breadcrumb.
- `styles_prose_contrast_test.go`: WCAG AA prose in both themes.
- `copy_contract_test.go`: action pattern in errors, actionable empty states, sentences ≤ 25 words.
- `content_name_that_ui_test.go`: Name That UI sections (Alternative names + Agent prompt) on served pages.
- `handbook_test.go`: Handbook pages render, in nav, in sitemap.

## [0.3.0] — 2026-08-14

### Added

- **Full A–J roadmap SDD**: formal verification of the 10 phases (A–J) with RED→GREEN pairs.
- **Registry sync guards**: `registry_sync_test.go` — registries (component, pattern, theme, dependency, agent-prompts, screen-composition) must reference real files.
- **Improved landing**: FAQ (Base UI), claims with checkmarks (Naive UI), visual demo card (Basecoat), GitHub link, documented BASE_URL.
- **Docs root explainer**: what Gelium UI is (and is NOT), 4-step Quick start, themes explained (Material, Basecoat, Base UI as vocabulary, never runtime).
- **Basecoat theme** (Phase I, PR #19) — complete theme, light + dark single class.

### Changed

- `index.md` wired to `/docs` (it was embedded but no route served it).
- Wire contract documented as canonical (`gelium-ui-wire-compatibility.md` rewritten).

## [0.2.0] — 2026-08-13

### Added

- **Theme mechanism** (Phase H, PR #18): class selection on the document root (`<html>`), dark via single-class route (no `@media prefers-color-scheme`), themes swappable without touching markup.
- **Screen recipes** (Phase G): Admin Resource, Ops Queue, Public Feed — composition patterns.
- **Public content patterns** (Phase F): 14 content patterns with card slots.
- **Server contracts**: stable GET+query, POST+303, 422 + `X-Loom-Validation` header (later `X-Gelium-Validation`), `loom:toast` (later `gelium:toast`).

## [0.1.0] — 2026-08-12

### Added

- **Gelium UI core**: server-rendered components in Go + HTMX, zero component JS.
- **6-layer architecture**: core tokens → themes → components → patterns → recipes → screens.
- **`--ui-*` tokens**: typography, color roles, spacing, elevation vocabulary.
- **Two themes in one bundle**: Material (default, M3) + Basecoat.
- **Access and verification contracts** (Phases A–D): native semantics, focus rings, aria-* across the whole surface.
- **Docs shell**: navigation, sidebar, breadcrumbs, search (disabled), theme/scheme switchers.

[Unreleased]: https://github.com/cryptonahue/gelium-ui/compare/v0.6.6...HEAD
[0.6.6]: https://github.com/cryptonahue/gelium-ui/releases/tag/v0.6.6
[0.6.5]: https://github.com/cryptonahue/gelium-ui/releases/tag/v0.6.5
[0.6.4]: https://github.com/cryptonahue/gelium-ui/releases/tag/v0.6.4
[0.6.3]: https://github.com/cryptonahue/gelium-ui/releases/tag/v0.6.3
[0.6.2]: https://github.com/cryptonahue/gelium-ui/releases/tag/v0.6.2
[0.6.0]: https://github.com/cryptonahue/gelium-ui/releases/tag/v0.6.0
[0.5.3]: https://github.com/cryptonahue/gelium-ui/releases
[0.5.0]: https://github.com/cryptonahue/gelium-ui/releases
[0.4.0]: https://github.com/cryptonahue/gelium-ui/releases
[0.3.0]: https://github.com/cryptonahue/gelium-ui/releases
[0.2.0]: https://github.com/cryptonahue/gelium-ui/releases
