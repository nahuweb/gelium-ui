# Skill: Design criteria — visual judgment

Use this skill after page/section architecture and before markup or visual
polish. It complements skills 01–10: it does not replace product reasoning,
section contracts, component registry, token discipline, accessibility, server
contracts, or the DoD. Visual judgment is a constrained decision process, not a
request to make every page distinctive.

## Scope and conversational aliases

The names `/g*` used in this document are **conversational workflow aliases**
for agent/user dialogue (for example, `/gbrief`, `/gshape`, `/gcheck`). They
are not runtime slash commands, routes, browser commands, or Gelium features.
Record the alias in a design note only when it helps a person follow the
conversation; implementation must use the repository's actual files, routes,
components, and contracts.

## Token-first, escalate on mismatch

Start from the existing Gelium contract and tokens. Prefer a registered
component, semantic HTML, and the existing `--ui-*` role that matches the
intended relationship. Treat the token scale as the first hypothesis, not an
absolute rule: when a token or component does not express the required
hierarchy, density, state, or accessibility outcome, record the mismatch and
escalate it for a deliberate extension or exception. Never hide a mismatch by
adding a page-local literal or overriding a component's anatomy casually.

Check:

- Theme and mode are class-routed on `<html>` (`theme-*`, optionally
  `theme-dark`), with no one-off color, type, radius, shadow, or spacing
  literals.
- Spacing expresses a known relationship: element/metadata, list sibling,
  group, or page section. Components own internal padding; composition owns
  inter-component gaps.
- Existing component anatomy, states, and registered primitives are reused
  before new markup or CSS is proposed.
- Every proposed exception names the failed token/component fit, the smallest
  safe scope, and how it remains compatible with existing Gelium contracts.

## Visual + interaction direction

After the structural wireframe and before the buildable wireframe, record the
approved direction for the screen. This is not permission to redesign under
polish; it is the decision layer that makes the Gelium composition concrete.

- **Theme:** selected class-routed `theme-*` family and light/dark route.
- **Visual recipe / skin:** existing project recipe, registered pattern, or an
  explicit `none`; never invent a parallel token system.
- **Interaction behavior:** registered Gelium component or pattern for each
  disclosure, tab, accordion, dialog, select, validation, loading, feedback, and
  recovery need, including the no-JS/server fallback.
- **Copy and voice:** brand tone, terminology, labels, state messages, and words
  to avoid; use realistic lengths before Build.
- **Localization:** locales, fallback, expansion, pluralization, date/number, and
  RTL considerations where applicable.
- **Media:** `MEDIA-*` role, source/license, ratio, crop/focal point, alt or
  transcript, wide/narrow behavior, loading, and fallback. Never invent metadata.

If a behavior, recipe, theme, copy constraint, or media ratio changes hierarchy,
section boundaries, or responsive reflow, update the buildable wireframe and ask
for approval again.

## Criteria plan (prebuild)

Before markup for a design-gated change, record the intended hierarchy, DOM
order, primary/supporting actions, section boundaries, responsive intent,
token/theme/recipe choices, interaction behaviors, copy/localization, media,
applicable states, accessibility/no-JS path, preserved server contracts, and the
`DESIGN-MEMORY` reuse decision. This plan is a prebuild decision record; it must
not claim rendered evidence that cannot exist yet. Record any unresolved mismatch
as an escalation or bounded exception.

## Rendered audit (postbuild)

After a candidate exists, evaluate the rendered screen at narrow and wide
widths, in the selected light and dark class-routed themes, and with realistic
content. Mark each item **pass**, **fail**, **not applicable**,
**pass-with-escalation**, or **exception**, with a short evidence note.

### Hierarchy and composition

- [ ] In a five-second scan, the page purpose, location/context, primary task,
      and task-critical state are identifiable.
- [ ] DOM and visual order match the intended reading order; optional depth does
      not precede identity or primary work.
- [ ] Major regions have distinct purpose and boundaries from their
      `SECTION-CONTRACT`; decorative wrappers are not presented as product
      sections.
- [ ] There is one visually dominant page-primary action; supporting and
      owner-only actions do not compete with reader content.
- [ ] Related items are grouped by proximity and unrelated intentions have a
      clear boundary; spacing is evidence of the relationship, not a substitute
      for one.
- [ ] Responsive behavior preserves hierarchy, readable measure, and contained
      scrolling without clipping or masking content.

### Type, color, and surface

- [ ] Typography uses the existing semantic type roles and creates a clear
      heading/content/metadata distinction.
- [ ] Contrast and emphasis remain sufficient in every selected theme and
      state; meaning is never communicated by color alone.
- [ ] Surface, border, radius, and elevation choices clarify grouping or
      interaction; decorative cards are not added just to fill empty space.
- [ ] State styling is consistent with existing Gelium feedback patterns and
      remains understandable in forced colors and high-contrast conditions.
- [ ] Motion is purposeful, optional enhancement only, and respects
      `prefers-reduced-motion`.

### Accessibility and resilience

- [ ] Native semantics, landmarks, headings, labels, focus order, and
      `:focus-visible` behavior remain intact; custom styling does not erase
      affordances.
- [ ] Interactive targets meet the Gelium touch-target contract (normally
      `--ui-touch-target`, at least 44px), and keyboard users can reach every
      action without hover-only behavior.
- [ ] The primary flow, navigation, disclosure, and recovery path work with JS
      disabled using HTML, CSS, and server rendering; HTMX/JS only enhances.
- [ ] Loading, empty, error, success, validation, and partial states retain
      the same hierarchy and have persistent recovery/next-step feedback.
- [ ] Existing URL, POST+303, 422 validation, `X-Gelium-Validation`, metadata,
      component, and theme contracts are preserved or an intentional change is
      explicitly recorded.
- [ ] If a glyph is required, it is a Gelium `.ui-icon` referenced by a
      **string literal** from the product's chosen catalog (`data-gelium-icon="chevron_right"`
      or `icons.SVG("tabler:chevron-right")`), next to visible text; not `›`,
      emoji, an icon font, or an ad-hoc SVG. The consumer embed is produced by
      `extract-used-icons`.

## Gelium icon allowlist

The **library** catalogs are Material Symbols rounded
(`@material-symbols/svg-400`) and Tabler Icons (`@tabler/icons`). The
**consumer** chooses one default set and the binary embeds only names
referenced as string literals. Do not compile either catalog into the app,
and do not invent a character or one-off path.

Pick the product set once (`--set material|tabler`; default `material`).
Unprefixed names resolve against that set. Prefixes override per glyph:

```html
<span data-gelium-icon="chevron_right"></span>
<span data-gelium-icon="tabler:chevron-right"></span>
```

```go
icons.SVG("settings")
icons.SVG("ms:home")
icons.SVG("tabler-filled:star")
```

Then generate the embed (only used glyphs):

```text
node node_modules/gelium-ui/scripts/extract-used-icons.mjs \
  --scan . --out internal/icons/icons.go --package icons \
  --set material
```

Use `--set tabler` when the product chose Tabler. Prefer one set per product;
prefixes are for an explicit mixed screen, not a default.

Contract: `.ui-icon`, `aria-hidden` + `focusable="false"` when decorative,
visible text names the control, Material uses `fill="currentColor"`, Tabler
outline uses `stroke="currentColor"` + `data-gelium-set="tabler"`. No
user-built markup. Import `icon.css` (or the Gelium index that already
includes it).

If extract errors `unknown Material Symbol` or `unknown Tabler icon`, the
name is not in that catalog — look it up on <https://fonts.google.com/icons>
or <https://tabler.io/icons>, pick another name, or use text only. Do not paste
a random SVG. The docs gallery still uses a small curated Material demo set
(`scripts/copy-icons.mjs`); that is not the consumer catalog.

## DESIGN-MEMORY: reuse with evidence

Before inventing a new macro-structure, inspect nearby screens and the
repository's existing patterns. Flag repeated macro-structures (for example,
a recurring page header, filter/action band, detail layout, or recovery zone)
and state whether the reuse is intentional because the user job, hierarchy,
or contract is shared. Reuse is the default when it improves recognition and
consistency; do **not** force variation merely to make screens look different.

Record:

```text
Pattern / screens:
Shared user job or contract:
Intentional reuse: yes | no | partial
Reason for reuse or justified divergence:
Evidence: existing route/component/token/template:
```

A divergence is warranted when the audience, task mode, content density,
permission boundary, state model, or responsive need materially differs. Keep
the divergence at the smallest layer: composition before new component,
component behavior before new token family, and a new primitive only after the
registry/vocabulary gap is proven.

## Optional composition heuristic

The 60–30–10 split (dominant/supporting/accent emphasis) is an **optional
heuristic**, not a requirement or a measurement target. Use it only as a quick
way to discuss emphasis when it clarifies a real hierarchy. Do not force colors,
percentages, or visual variety to satisfy it; token roles, contrast,
accessibility, content, and the page's purpose take precedence.

## Workflow aliases (conversation only)

Use these as short dialogue labels, never as runtime commands:

1. `/gbrief` — confirm product intent, audience, surface, theme, and existing
   constraints; completion: the relevant artifact and assumptions are named.
2. `/gshape` — apply skills 02, 08, and 10; completion: page/section contracts
   and reading order exist before markup.
3. `/gcriteria` — record the prebuild criteria plan, including hierarchy,
   states, contracts, tokens, accessibility/no-JS, and DESIGN-MEMORY; completion:
   unresolved mismatches are escalated or bounded.
4. `/gcheck` — run the rendered audit after a candidate exists at required
   widths, themes, states, and no-JS; completion: every item is pass, fail,
   not applicable, pass-with-escalation, or exception with evidence.

## Verification record

Before calling a screen complete, record:

- criteria plan and the rendered-audit evidence inspected;
- selected theme classes, viewport widths, and representative states;
- token/component choices and any escalated mismatches;
- DESIGN-MEMORY reuse decision;
- accessibility and no-JS checks;
- exact failures, fixes, or explicit exceptions.

Then run `skills/09-usability-checklist.md` and
`skills/07-dod-and-antislop.md`. Visual polish cannot waive either contract.

## Sources

| Source | Use |
| --- | --- |
| Gelium skills 01–10 | Existing tokens, architecture, contracts, accessibility, states, and DoD |
| Material 3 and GOV.UK patterns | Role-based emphasis, density, resilient and inclusive service design |
| Gelium registry and `lib/ui-vocabulary.md` | Reuse and canonical component naming before invention |
