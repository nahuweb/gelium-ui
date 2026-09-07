# Gelium UI

> Themeable, open-code UI components for Tailwind CSS and HTMX.

Current release: **v0.6.7** (prepared; npm publication pending) — [npm `gelium-ui`](https://www.npmjs.com/package/gelium-ui).

Gelium UI is an open-code, server-rendered UI library for HTML-first products.
It gives teams reusable components, token-based themes, accessible states, and
a real no-JavaScript baseline without requiring React, Vue, hydration, or a
client-side component runtime.

## Is Gelium for you?

| Choose Gelium if you need… | Gelium may not be the right fit if you need… |
| --- | --- |
| Go, Rails, Laravel, Django, or plain HTML integration | A React/Vue component runtime or hydration-first architecture |
| Copyable HTML and server-rendered templates | A closed, JavaScript-only component abstraction |
| HTML forms and URLs that work without JavaScript | Client-only state with no server-rendered fallback |
| Tailwind CSS 4, HTMX as enhancement, and class-based themes | A bundled design brand or a mandatory frontend framework |

## What you get

The published package contains CSS, themes, HTML partials, optional enhancement
JavaScript, portable agent guidance, and text reference fichas:

```text
gelium-ui/
├── dist/       prebuilt CSS bundle
├── styles/     source CSS, tokens, and component styles
├── themes/     Material, Basecoat, Base UI-inspired, and product skins
├── templates/  copyable server-rendered HTML partials
├── js/         optional toast, validation, slider, and transition enhancements
├── skills/     decision and implementation guidance for agents
└── references/ portable structural references for npm-only consumers
```

### The repository docs consumer

The accompanying docs consumer includes HTMX 4, an **on-this-page rail**,
**prev/next pagination**, demo-first documentation, and mobile foundations.
These are examples of the repository dogfood site, not runtime requirements of
the npm package.

## Quick start

### Install (consumer)

```bash
npm install gelium-ui
```

Drop-in CSS bundle (themes + components):

```css
@import "gelium-ui/dist/gelium.css";
```

Or pick themes + source styles in your own Tailwind entry:

```css
@import "tailwindcss";
@import "gelium-ui/styles/index.css";
@import "gelium-ui/themes/theme-material.css";
@import "gelium-ui/themes/theme-basecoat.css";
```

Use a registered component in your server-rendered HTML:

```html
<button class="ui-button ui-button-primary" type="submit">
  Save changes
</button>
```

Then choose a visual direction on the document root:

```html
<html class="theme-material">
```

See the [component registry](docs/gelium-ui-component-registry.md) for the
canonical inventory and the [package README](lib/README.md) for templates,
themes, and optional JavaScript.

## Core contract

- **HTML-first:** native elements before ARIA or custom controls.
- **No-JS first:** the main flow works with JavaScript disabled.
- **Server-first state:** URLs, forms, HTTP statuses, and validation remain real.
- **Progressive enhancement:** HTMX and `js/gelium.js` enhance; they do not own
  the primary flow.
- **Theme by tokens:** themes change semantic `--ui-*` values without requiring
  different component markup.
- **States included:** loading, empty, error, success, validation, and recovery
  are part of the component contract where applicable.
- **Accessible by default:** focus, keyboard, contrast, responsive behavior,
  reduced motion, and forced colors are tested contracts.

## Components and themes

The library includes buttons, text fields, selects, checkboxes, radios, switches,
sliders, dialogs, toasts, cards, lists, data tables, navigation, tabs, pagination,
media, feedback, and layout primitives. Themes and skins share the same semantic
markup and server contracts; documented divergences are intentional and are not
claimed as official vendor parity.

Optional progressive-enhancement JS (toast, 422 validation swap, slider fill,
view-transition guard):

```html
<script defer src="node_modules/gelium-ui/js/gelium.js"></script>
```

Theme selection can still use a class on the document root, e.g.
`<html class="theme-material">` (add `theme-dark` / `data-theme="dark"` for
dark). Copyable HTML partials live under `node_modules/gelium-ui/templates/`.

Package page: [npmjs.com/package/gelium-ui](https://www.npmjs.com/package/gelium-ui)
· consumer README inside the package.

## Architecture (selection model)

Gelium separates concerns so one component can combine systems without forks:

```text
Core
  semantic HTML, a11y, tokens, no-JS baseline

Behavior
  interaction policy only (native | material | basecoat | baseui)

Reference visual preset
  default visual direction of a reference system
  (material | basecoat | baseui | none)

Product skin
  overlay that wins over reference
  (basecoat Vega/Nova/Maia/…, baseui, alden, linear, vercel, …)

Contract
  gelium  → platform floors (default; e.g. touch ≥44px)
  source  → third-party-faithful density from the active skin/reference

Site / product
  always can override tokens on top
```

### Public query / root attrs

```text
?behavior=basecoat
&reference=auto
&skin=basecoat-mira
&contract=source
&scheme=dark
&execution=native
```

Root attributes (resolved server-side, allowlisted):

```html
<html
  class="theme-basecoat"
  data-gelium-reference="basecoat"
  data-gelium-skin="basecoat-mira"
  data-gelium-contract="source"
  data-gelium-scheme="dark">
```

Legacy `?theme=` remains compatible.

### Contract

| Value | Meaning |
| --- | --- |
| `gelium` (default) | Gelium platform rules win over denser source sizes |
| `source` | Use skin/reference authored density; product CSS may still override |

Example: Basecoat Mira wants ~`h-7`. With `contract=gelium` the button stays
≥ touch target. With `contract=source` the pack density can show, and a site
token can still raise it again.

Recipe UI exposes Contract under **Advanced**.

### Basecoat style packs

Basecoat’s real look lives in style packs, not only `base`:

| Gelium skin | Basecoat pack |
| --- | --- |
| `basecoat` | Vega (default) |
| `basecoat-nova` | Nova |
| `basecoat-maia` | Maia |
| `basecoat-lyra` | Lyra |
| `basecoat-mira` | Mira |
| `basecoat-luma` | Luma |
| `basecoat-sera` | Sera |
| `basecoat-rhea` | Rhea |

Color foundations still use `.theme-basecoat`. Pack differences are anatomy
tokens under `data-gelium-skin`. Evidence packages live in
`docs/audit/packages/`.

### Base UI note

Official `@base-ui/react` is headless (no visual CSS). Gelium’s Base UI visuals
are **docs-inspired / Gelium-authored**, never claimed as official package CSS.

## Inspiration and attribution

Gelium is informed by public design systems, open-code libraries, UX research,
visual studies, and the [Gentle AI](https://github.com/Gentleman-Programming/gentle-ai)
agent workflow. The roles are different, and the table below records the
adaptation rather than implying affiliation or runtime dependency.

| Area | Sources studied | Gelium adaptation |
| --- | --- | --- |
| Design foundations | [Material Design 3](https://m3.material.io), [USWDS](https://designsystem.digital.gov), [GOV.UK Design System](https://design-system.service.gov.uk), [Mozilla Protocol](https://protocol.mozilla.org) | Semantic `--ui-*` tokens, content guidance, accessibility, readable measure, state/elevation/type rules |
| Component behavior and open code | [Base UI](https://base-ui.com), [Basecoat UI](https://basecoatui.com), [Naive UI](https://naiveui.com), [shadcn/ui](https://ui.shadcn.com), [templ](https://templ.guide) | Composable behavior, copyable HTML, demos-first docs, and server-rendered/no-JS adaptations—not their React runtime |
| Naming, references, and icons | [Name That UI](https://namethatui.com), [Material Web](https://github.com/material-components/material-web), [Material Symbols](https://github.com/google/material-design-icons), [Tabler Icons](https://github.com/tabler/tabler-icons) | Alternative names, agent-readable references, audited token families, curated inline SVG icons |
| Visual directions | Basecoat style packs, [Linear](https://linear.app), [Vercel](https://vercel.com), Alden, and [Refero](https://styles.refero.design) studies | Token-only skins over unchanged Gelium markup and server contracts; documented divergences instead of claimed parity |
| UX and content research | [Nielsen Norman Group](https://www.nngroup.com), [WCAG](https://www.w3.org/WAI/standards-guidelines/wcag/) | Reading, hierarchy, plain language, focus, contrast, responsive, and recovery checks |
| Agent workflow | [Gentle AI](https://github.com/Gentleman-Programming/gentle-ai) | Outcome-first routing, bounded delegation, optional SDD, candidate-first review, and separated delivery authority; not full RDD |

Gelium reimplements ideas as original semantic HTML, `--ui-*` tokens,
server-rendered templates, and optional progressive enhancement. It does not
ship vendor branding, copied markup, screenshots, or runtime dependencies.
See the [full acknowledgments matrix](site/web/content/handbook-acknowledgments.md)
for sources, adaptations, licenses, and explicit boundaries.

## Two audiences

- **Install and use** (developer or LLM): `npm install gelium-ui` — CSS, themes,
  templates, `dist/gelium.css`, `js/gelium.js`. The embedded Go handler is the
  advanced demo mode.
- **Contribute** (this repo): npm workspaces monorepo — `lib/` is the published
  package, `site/` imports it by package name (dogfooding).

## Requirements

- Go 1.24 or later.
- Node.js 20 or later with npm.

## Setup and reproducible build (contributor)

From the project root:

```bash
npm install
go mod download
npm run build
```

`npm run build` (root) chains: site build (`tailwindcss` on
`site/web/styles/app.css` → `site/web/static/app.css`, copy `htmx.min.js` and
icons) + library dist (`lib/styles/dist-entry.css` → `lib/dist/gelium.css`) +
copy `lib/js/gelium.js` → `site/web/static/gelium.js`. Final artifacts under
`site/web/static/` and `lib/dist/` are build artifacts embedded by the Go binary
via `embed`.

## Tests and checks (gates)

```bash
go test ./internal/... ./site/... ./lib/...
go vet ./internal/... ./site/... ./lib/...
npm run build
git diff --check
gofmt -l internal site lib
```

Do **not** use bare `go test ./...` if the tree includes intentionally excluded
pre-split packages. Tests use `httptest` and verify handlers, URL-encoded forms,
HTTP 422/200, Markdown render, component HTML/ARIA contracts, HTMX integration,
CSS/tokens, and embedded assets without full-page snapshots.

## Real structure

```text
gelium-ui/
├── lib/                        ← publishable npm package (gelium-ui)
│   ├── assets.go               (//go:embed templates + styles → geliumui/lib)
│   ├── version.go              (AssetsVersion: centralized cache-bust)
│   ├── package.json            (exports map, files, sideEffects)
│   ├── dist/gelium.css         (prebuilt bundle: preflight + themes + components)
│   ├── js/gelium.js            (consumer JS: toast, 422 contract, VT, slider)
│   ├── styles/                 (tokens, base, components, index manifest)
│   ├── templates/              (component partials)
│   ├── themes/                 (material, basecoat, baseui, alden, linear, vercel)
│   └── *_test.go               (component contract suites)
├── site/                       ← demo consumer (docs, landing, recipes)
│   ├── package.json            (depends on gelium-ui by package name)
│   └── web/
│       ├── assets.go           (//go:embed templates + content + static)
│       ├── content/            (markdown docs)
│       ├── static/             (compiled app.css, gelium.js, chrome JS)
│       ├── styles/             (app entry, selection/reference/skin adapters, docs chrome)
│       └── templates/          (shell / landing / recipe chrome)
├── internal/app/               (full Go docs server + selection resolver)
├── docs/                       (contracts, audits, package evidence)
├── scripts/                    (copy-htmx, copy-icons, copy-lib-js, build-lib-dist)
├── go.mod                      (module geliumui)
└── package.json                (workspaces: ["lib", "site"])
```

### Library vs docs separation

- `lib/` = reusable package: `ui-*` components, tokens, themes, consumer JS, dist.
- `site/` = first consumer: docs shell, landing, recipes, blog. Imports the library
  **by package name** — the same contract external consumers get.
- `internal/app` = Go handler serving the site (dual embed: `geliumui/site/web` +
  `geliumui/lib`).

### Note on `embed`

Go `//go:embed` cannot read parent paths with `..`. There are two embeds:
`lib/assets.go` and `site/web/assets.go`. `buildTemplates()` merges both with
`template.ParseFS` — names are disjoint by construction (collision-guard test).

### Cache-bust

Asset cache-bust is a single value: `lib.AssetsVersion`, rendered as
`?v={{.AssetsVersion}}`. Coherence tests pin npm version == constant and forbid
hardcoded `?v=`. Bump it when a static asset changes.

## Run

```bash
go run ./cmd/gelium        # serves on :8787 (PORT=3000 for another port)
```

You can also mount the docs site from your own app with `internal/app.New()`:

```go
package main

import (
 "log"
 "net/http"

 "geliumui/internal/app"
)

func main() {
 log.Fatal(http.ListenAndServe(":8787", app.New()))
}
```

## Wire contract (`gelium:*` / `X-Gelium-*`)

Server-driven wire contracts use the product prefix: `gelium:toast`
(HX-Trigger event + `#gelium-toast-region`) and header
`X-Gelium-Validation: true`. The served HTMX hook (`lib/js/gelium.js`) and tests
pin these as canonical.
Reference: [`docs/gelium-ui-wire-compatibility.md`](docs/gelium-ui-wire-compatibility.md).

## Scope

Current scope includes open-code components (Button, Text field, Dialog, Toast,
Data table, Chips, Navigation, Accordion, …) with server-rendered demos, multiple
visual directions (Material, Basecoat packs, Base UI-inspired, Alden, Linear,
Vercel), selection architecture (behavior/reference/skin/contract), and
server-side HTMX validation patterns. It does not ship Lit, Shadow DOM, Astro,
`templ`, SQLite, SSE, a CLI, or a third-party component registry runtime.

## License

Gelium UI is distributed under the [MIT license](LICENSE).
