# Skill: Definition of Done and anti-slop

Run this checklist before calling a surface done. If any item is NO, fix it — do
not pass.

## Step 0 — brief gate

Read consumer `PRODUCT.md` / `DESIGN.md` when present. If either is absent or
incomplete, run the plain-language brief in `llms-ux.txt` before implementation.
Accept `Unknown`, `To decide`, and `N/A` with a reason; record assumptions and
open decisions. Do not silently invent product intent, chrome, or visual direction.

Proceed when the brief establishes the user job, scope boundary, intent mode (`new-product`, `redesign`, or explicit `visual-migration`), and either the
relevant surface/theme decisions or an explicit decision to resolve them during
Shape/Architecture. Redesign also needs a current-state audit; visual migration needs preserved UX/server contracts. Stop only when an unresolved decision would change page or
flow architecture.

## UI Definition of Done (product/structure/states/contracts)

**Product**

- [ ] One primary action per screen.
- [ ] Each screen serves the user's job, not the org chart.
- [ ] Empty/error/loading/success all handled.

**Structure**

- [ ] Semantic HTML (native elements), logical heading order (H1 → H2 → H3).
- [ ] Every major region has a purpose-bound `SECTION-CONTRACT` before component selection.
- [ ] Gelium interaction behavior, theme/recipe/skin, copy/locales, and media direction are recorded before Build.
- [ ] Page shell composes registered components (`ui-container`,
      `ui-navigation-bar`/drawer primitives). Custom shell CSS is limited to
      spacing/width — no hand-rolled nav headers or sticky shells.
- [ ] Correct list semantics for lists; tables only for tabular data.
- [ ] Links are real `<a href>`; actions are real buttons/forms.

**States**

- [ ] Empty state explains what/why/what-to-do, with a CTA.
- [ ] Error copy names the fix (action pattern), never blames the user.
- [ ] Loading is readable; content doesn't jump on load.

**Contracts**

- [ ] Read = GET + URL state.
- [ ] Mutate = POST + 303 (or 422 + `X-Gelium-Validation` re-render).
- [ ] No-JS flow completes end-to-end.
- [ ] Chat/feedback uses the FEED matrix, persistent ≠ toast.

## Anti-slop (Gelium-aware) — far side of the coin

Avoid:

- nested cards for simple forms;
- stock purple-blue hero gradients;
- an icon tile above every H1;
- gray on saturated fills (low contrast);
- bounce/babel motions;
- "bolder for its own sake" on Operate tables;
- a new font per feature;
- layout tables for non-data; "lorem" filler that hints at content.

OK:

- Material/Basecoat/Neubrutalism token stacks from the theme;
- brand fonts only where `DESIGN.md` says so;
- spacing from the core `--ui-space-*` scale (no feature-specific scale).

## Verification

Run `scripts/ux-detect.sh` from the consumer repo root (ships inside the package,
e.g. `bash node_modules/@gelium/ui/scripts/ux-detect.sh`). It greps for the
mechanical checks: no media-query dark overrides, no hand-rolled page shells, no
`overflow-x: hidden` on body, no one-off color literals, validation-summary hooks
on every form. Where the consumer repo also mirrors this package's monorepo gates,
run those too (styles_*_test, copy/contrast contracts). At minimum confirm:

- Dark mode is class-routed: `theme-dark` on `<html>` — grep for
  `prefers-color-scheme:\s*dark` overrides (forbidden; no media-query dark hex).
- Page shells compose registered components; custom shell CSS is only
  spacing/width.
- A validation-summary + inline errors on every form.
- No one-off color literals in shipped markup.

## Do not

Skip to pretty CSS. Claim pass without the checks. Redesign under "polish". End
in an endless loop — one polish pass, then stop.
