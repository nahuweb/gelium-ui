---
name: gelium-ui-component-implementation
description: Use when implementing or reviewing a new Gelium UI component in the library. Apply the canonical route, component contract, TDD, ownership, no-JS, and verification workflow.
version: 0.6.7
license: MIT
---

# Gelium UI component implementation

This skill governs adding or materially changing one component in the Gelium UI
library. It is downstream of `00-agent-routing.md`: route the task first. It is
not for composing a consumer screen or for changing an existing component's
copy/token/selector without architectural impact.

## Scope and route

Implement exactly one component per work unit. Use the smallest route:

- `direct-exempt`: a bounded correction with no new component anatomy or flow.
- `delegated-direct`: broad research or multi-file work without architecture
  change; use one bounded worker and parent verification.
- `design-gated`: a new component changes public anatomy, interaction, states,
  or ownership boundaries; obtain the relevant architecture decision before
  implementation.
- `full-sdd`: only when durable proposal/spec/design/tasks reduce substantial
  cross-cutting ambiguity.

The component writer never owns approval or delivery. Do not commit, push,
publish, or deploy from a worker handoff.

## Required discovery before writing

Read in this order and record exact paths:

1. `lib/skills/00-agent-routing.md`.
2. `lib/AGENTS.md`, `lib/llms-ux.txt`, and applicable skills.
3. `docs/gelium-ui-component-registry.md` and `docs/gelium-ui-vocabulary.md`.
4. `README.md`, `lib/README.md`, and the relevant component contract.
5. One or more similar real components: Go view model/handler, template, CSS,
   tests, docs, and build entrypoint.
6. The current package/version and generated-asset contract.

Resolve the component's canonical vocabulary before naming it. If the registry
has no valid row, stop and record the product/architecture decision needed; do
not invent a parallel component category.

Before writing, produce a compact ownership table:

| Path | New/existing | Owned/shared | Writer | Verification |
|---|---|---|---|---|

Shared paths include routes, layout, central app CSS/JS, registries, package
metadata, generated assets, and `styles_contract_test.go`. A component task does
not implicitly own them.

## Component contract

Define the smallest complete contract before implementation:

- purpose and audience;
- canonical name and category;
- anatomy and semantic HTML root;
- variants and states: rest, hover, focus-visible, active, disabled, loading,
  error, and empty where applicable;
- keyboard/focus and accessible naming behavior;
- `--ui-*` token ownership and light/dark behavior;
- responsive, reduced-motion, forced-colors, and RTL behavior where applicable;
- server/HTMX contract, if any;
- complete no-JS path;
- docs and registry impact;
- alternatives rejected and any `UNKNOWN`/`BLOCKED` evidence.

Use native HTML before ARIA. Use concrete Go view models and closed vocabularies.
Do not use React, JSX, Lit, Shadow DOM, custom runtime elements, CDN assets,
or new dependencies without an explicit approved exception.

## Platform-first and no-JS rules

Before adding JavaScript, compare semantic HTML, CSS, server-rendered forms and
navigation, and HTMX enhancement. JavaScript is allowed only for a demonstrated
platform gap, with a real no-JS fallback and a failing test first. Never make
reading, navigation, submission, validation, or the primary flow depend on JS.

Preserve Gelium contracts:

- GET for reads and navigable state;
- POST + 303 for mutations;
- 422 + `X-Gelium-Validation` for validation;
- `gelium:toast` only for transient action feedback;
- persistent/contextual feedback stays in the page or section.

Do not fabricate identity, media dimensions, states, or browser support.

## Implementation order

1. **RED:** NO PRODUCTION CODE WITHOUT A FAILING TEST FIRST. Add one focused
   failing test for one behavior and run only it.
2. Confirm the failure is caused by the missing behavior, not setup or a typo.
3. **GREEN:** implement the smallest production change.
4. Run the focused test, then the related package/handler suite.
5. **REFACTOR:** simplify only while green; repeat for each independent behavior.
6. Add the component template under `lib/templates/<slug>.html` and CSS under
   `lib/styles/<slug>.css`; update the source style manifest, not generated CSS.
7. Add/update the corresponding `lib/styles_<slug>_test.go` and any Go view-model
   tests. Keep shared test helpers in `lib/styles_contract_test.go`; do not move
   Go tests into a separate `test/` directory.
8. Update the real registry, dependency metadata, and docs only when the
   component contract requires them.
9. Regenerate `lib/dist/gelium.css` and any embedded site assets through the
   existing build, never by hand.

## Verification

Run from the repository root:

```bash
go test ./lib/... -count=1
go test ./internal/... ./site/... -count=1
go vet ./internal/... ./site/... ./lib/...
npm run build
bash scripts/ux-detect.sh
git diff --check
```

For server behavior, test ordinary requests and `HX-Request: true` separately.
For visual/interaction behavior, inspect light/dark, narrow/wide, keyboard,
focus, disabled/loading/error/empty, reduced motion, forced colors, and the
supported browser matrix. Mark unobserved coverage as a limitation; never call
it a clean visual pass.

Before handoff, confirm:

- only the approved component scope changed;
- no generated bundle was hand-edited;
- no shared path changed without ownership;
- the component is discoverable in the registry/package when intended;
- no-JS and server contracts remain functional;
- tests, build, vet, detector, and diff checks have real output;
- unresolved findings have an owner and bounded follow-up.

## Handoff status

Use one truthful status:

- `READY_FOR_INTEGRATION`: worker changes and evidence are ready; shared
  integration remains with the parent.
- `COMPLETE_AWAITING_USER_ACCEPTANCE`: implementation, review, and required
  smoke evidence are complete, but explicit acceptance remains.
- `BLOCKED`: a required product, platform, ownership, data, or verification
  decision is missing.
- `ABORTED_ON_DRIFT`: an owned baseline changed unexpectedly; report path,
  expected state, observed state, and stop.
