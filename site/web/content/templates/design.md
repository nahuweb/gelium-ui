# DESIGN.md template (consumer apps)

Copy this file to your application repo root as `DESIGN.md`. It records the visual and interaction direction without forking Gelium's token system.

You can answer the starter questions in plain language. If you do not know an answer, write `Unknown`, `To decide`, or `N/A` with a short reason. The agent may propose a safe default, but must record it as an assumption and ask for confirmation when it affects page architecture.

## Start in plain language

1. Which screens or URLs do you expect to have first?
2. Should each surface mainly help people **do a task**, **read**, **decide**, or **experience** something?
3. Should navigation be mainly a top bar, a side navigation, tabs, or something else?
4. What products or sites have a useful quality to borrow, and what should we avoid copying?

The agent should turn these answers into the structured sections below and ask only the follow-up questions needed for a safe implementation.

## Design intent

Choose one mode before drawing screens. Existing design is evidence to audit, not an
instruction to copy.

- `new-product`: design from the product outcome and user job.
- `redesign`: inspect the current experience, then propose preserve/improve/remove/unknown decisions.
- `visual-migration`: keep UX and server contracts stable while applying Gelium; use only when explicitly requested.

- Selected mode: new-product | redesign | visual-migration
- Product/design evidence: …
- Scope boundary: …

## Current-state audit (redesign and visual-migration)

Use `N/A` for `new-product`. For redesign, record evidence before proposing the target.

| Area | Current behavior | Finding | Decision | Evidence |
| --- | --- | --- | --- | --- |
| Shell / navigation | … | … | preserve / improve / remove / unknown | … |
| Content hierarchy | … | … | preserve / improve / remove / unknown | … |
| Responsive behavior | … | … | preserve / improve / remove / unknown | … |
| States and recovery | … | … | preserve / improve / remove / unknown | … |
| Accessibility | … | … | preserve / improve / remove / unknown | … |
| Theme / visual treatment | … | … | preserve / improve / remove / unknown | … |

## Screen inventory

One row per important URL or surface. The agent can propose the Gelium mode and screen type; do not invent them when the product decision is unclear.

| URL | Surface: Operate / Read / Persuade / Experience | Screen type | User job | Primary action |
| --- | --- | --- | --- | --- |
| `/` | … | hub / list / detail / form / confirm / settings / queue / result | … | … |

## Chrome and navigation

Resolve the shared shell before drawing individual screens. Use `undecided` only
when the decision is explicitly carried into the architecture gate.

- Shell scope: new / redesigned / existing-unchanged / out-of-scope / undecided
- Header: top bar / side navigation / other / none / undecided
- Navigation model: top bar / side navigation / tabs / other / undecided
- Footer: none / legal / contact / links / undecided
- What stays stable across screens: …
- Role-specific navigation or boundaries: …
- Narrow behavior: stack / collapse / disclosure / horizontal scroll / undecided
- Current-location treatment: breadcrumb / active nav / page heading / other

For a new product or product redesign, `Shell scope`, header, navigation, footer,
and narrow behavior are architecture decisions. They must be resolved or marked
as open decisions before the buildable wireframe can be approved.

## Theme

- HTML class: `theme-material` | `theme-basecoat` | `theme-neubrutalism` | undecided
- Dark mode: explicit `theme-dark` class route; never media-only as the sole authority
- Brand direction in one sentence: …

## Visual and interaction direction

Resolve this after the structural wireframe and before the buildable wireframe.
If a choice changes hierarchy, region boundaries, or reflow, update the wireframe
and record the decision again.

### Gelium visual direction

- Theme class on `<html>`: `theme-material` | `theme-basecoat` | `theme-neubrutalism` | undecided
- Light/dark route: explicit `theme-dark` class when applicable
- Visual recipe / skin: existing project recipe | none | to decide
- Why this theme/recipe fits the user job and product direction: …
- Token or component mismatch to escalate: …

### Gelium interaction behaviors

Choose registered components and their behavior, not generic UI names.

| Screen / section | Behavior need | Gelium component or pattern | States | No-JS / server fallback |
| --- | --- | --- | --- | --- |
| … | disclosure / tabs / accordion / dialog / select / validation / loading / feedback | … | … | … |

### Copy and localization

- Voice and tone: …
- Terminology and words to avoid: …
- Locales and fallback: …
- Expansion, pluralization, date/number, and RTL considerations: …
- Copy still provisional or approved: …

### Media direction

Use `MEDIA-*` references for images, illustrations, video, or audio. Do not invent
assets or metadata. If media changes layout, its ratio must appear in the wireframe.

| ID | Role / section | Source or license | Ratio / crop | Alt or transcript | Wide / narrow behavior | Loading / fallback |
| --- | --- | --- | --- | --- | --- | --- |
| MEDIA-01 | … | … | … | … | … | … |

## Density

- Operate surfaces: cozy | compact
- Read/Persuade surfaces: comfortable | cozy
- Exception and reason: …

## Data display defaults

- Collections: table / list / cards / decide per screen
- Detail records: description list / sections / other
- Filters, sorting, or pagination: …
- Never use a display choice that conflicts with the data's purpose.

## Components and composition

Prefer registered Gelium `ui-*` partials from `gelium-ui/templates`. A new primitive needs a token-first rationale and a vocabulary/registry check.

- Reusable components or recipes: …
- Components explicitly not wanted: …

## States and recovery defaults

For every applicable screen, plan rest, loading, empty, error, success, and partial states.

- Empty state: explain what is empty, why it matters, and the next action.
- Error state: explain the problem and a recovery action.
- Success state: use a result page when the outcome must persist; do not rely on a toast alone.
- No-JS path: …

## Motion and accessibility

- Motion: `MOTION-NONE` by default on Operate; honor `prefers-reduced-motion`.
- Keyboard and focus: …
- Contrast and non-color cues: …
- Narrow layout and touch target notes: …

## Anti-slop checklist

Keep Gelium's anti-slop rules and check project-specific constraints before markup.

- [ ] No purple-blue gradient used as a default identity.
- [ ] No centered hero plus three identical feature cards by default.
- [ ] No invented metrics, avatars, testimonials, or “trusted by” logos.
- [ ] No rounded surface and shadow applied to every element.
- [ ] One highlighted primary action per screen.
- [ ] Operate surfaces do not use decorative motion.
- [ ] Empty, loading, error, and success states have a server-rendered path.
- [ ] No one-off feature colors, fonts, or spacing scales.
- [ ] …

## References to borrow from

| Reference | Borrow | Avoid | Reason |
| --- | --- | --- | --- |
| … | … | … | … |

## Brand notes

Logo, typeface, and any extra brand color must map into Gelium tokens. Do not hardcode one-off hex values in features.

- Logo: …
- Typeface: use theme `--ui-font-*` unless a brand choice is recorded here.
- Brand tokens: …

## Visual definition of done

- [ ] A stranger understands the page purpose in five seconds.
- [ ] The screen inventory has one primary action per URL.
- [ ] Desktop and narrow layouts preserve the same regions and reading order.
- [ ] Theme, states, focus, contrast, and reduced motion are checked.
- [ ] Recovery works without JavaScript where the flow requires it.

## Open decisions and assumptions

- Decision: …
- Status: open / assumed / confirmed
- Impact if wrong: …
- Needed before implementation: yes / no

Gelium docs: `/docs/themes`, `/docs/tokens`, `/docs/density`, `/docs/agent-workflow`.
