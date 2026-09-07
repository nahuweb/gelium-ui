# Skill: Wireframe approval gate

Use this skill to decide when a page or flow needs an architecture and wireframe
approval before markup or CSS changes. It complements skills 01–11 and keeps
small, low-risk fixes moving. Approval is a human decision about a proposed
structure, not approval of pixel polish or a runtime slash command.

## When approval is required

Require approval **before implementation** when the change introduces:

- a new screen or route;
- a new user flow, lifecycle, or multi-step interaction; or
- a substantial redesign of an existing page, meaning its information
  hierarchy, major regions, primary action, reading order, or owner/reader
  boundary materially changes.

For these changes, follow `ROUTE → ORIENT → PLAN → ARCHITECT → APPROVE` before
markup. Orient reads the product/design artifacts, vocabulary, registry, and
hard contracts; Plan defines the user need; Architect inspects the real route,
template, handler, CSS, components, states, data, permissions, and no-JS/server
contract. Do not edit implementation files during Orient, Plan, or Architect.

### Two wireframes, one approval target

- **Intent wireframe (Plan)** records the user job, audience, major regions,
  primary action, states, and non-goals without inventing exact components or
  data.
- **Buildable wireframe (Architect)** reconciles that intent with real routes,
  data, permissions, templates, registered components, and server/no-JS
  contracts. Human approval applies to this version.

If Architect finds a material mismatch, return to Plan or escalate it; do not
hide it as CSS polish.

## Explicit exemptions

Do not block implementation approval for:

- accessibility fixes or inclusive responsive improvements;
- bug fixes that restore the documented behavior;
- corrections to an existing Gelium, URL, server, component, or data contract;
- component-level or mechanical changes with no page/flow architecture change
  (for example token replacement, selector repair, copy correction, or a
  mechanical migration); or
- small adjustments that are already covered by an approved architecture or
  wireframe and do not materially change its scope.

An exempt change still follows skills 01–11, the existing contracts, and the
DoD. If its implementation reveals a new screen, flow, or substantial
architectural change, stop at that boundary and apply this gate to the newly
expanded scope.

## Approval packet

For a gated change, copy `skills/templates/wireframe-approval-packet.md` into
the change record and complete it in English technical artifact convention. Its
Plan section contains the intent wireframe; its Architect section contains the
buildable wireframe that is actually approved. The packet records:

```text
Packet version:
Change:
Product job / audience:
Existing route and contracts:
Scope: new screen | new flow | substantial redesign
Intent mode: new-product | redesign | visual-migration
Section inventory: ordered regions and purpose
Primary action and action hierarchy:
Plan — intent wireframe:
Architect — buildable wireframe:
States and recovery:
Accessibility and no-JS behavior:
Reuse / DESIGN-MEMORY decision:
Open questions and explicit trade-offs:
Decision: approved | changes-requested | declined | exception
```

Wireframes are structural, not decorative. Use ASCII or another reviewable
text form to show reading order, major regions, primary/supporting actions,
disclosures, recovery, and owner-only areas. Do not spend the approval packet on
exact colors, shadows, or pixel dimensions; apply skill 11 after structure is
approved.

## Gelium structural ASCII wireframes

Show the wireframe in the conversation **before** markup. ASCII is a structural
contract, not a generic “pretty box” or a substitute for `SECTION-CONTRACT`.
It must map the chosen SURFACE + SCREEN blocks from `skills/02` and
`llms-ux.txt`, and must use real product content or explicit placeholders.

For every design-gated screen, include all of the following:

1. **Screen identity** — URL, SURFACE, SCREEN type, user job, one primary action,
   and relevant `JOURNEY-*`, `DATA-*`, and `FEED-*` IDs.
2. **Global shell** — header/navigation, current-location treatment, utility or
   account actions when applicable, and footer. Mark each as `new`, `redesigned`,
   `existing-unchanged`, or `out-of-scope`. An unresolved shell decision blocks
   approval; an existing shell still appears as a labelled band in the map.
3. **Page structure** — numbered major regions in DOM/reading order, each with its
   purpose, boundary, hierarchy, and primary/supporting action policy.
4. **ASCII notation** — use `┌ ┐ └ ┘ │ ─` (or `+ - |`), `[Button text]` for
   buttons, `[________________]` for inputs, and `→ Link text` for links. Keep
   alignment and indentation consistent; distinguish interactive controls from
   static content. A box does not automatically mean a card or a section.
5. **Responsive maps** — show both `WIDE` and `NARROW` versions. Preserve the
   same regions and reading order; show columns stacking and navigation reflow.
   Use actual project breakpoints only when verified; do not invent pixel values.
6. **States and recovery** — show rest, loading, empty, error, success, and
   validation/partial states when applicable, with a visible way forward and a
   no-JS path.
7. **Technical and direction annotations** — map regions to registered Gelium
   components, behavior patterns, tokens, selected theme/recipe/skin, copy and
   locale constraints, media references, route/server behavior,
   accessibility/focus order, and any deliberate mismatch or open decision.

Use the sequence `entry → primary work → supporting content → recovery`. Keep
one highlighted page-primary action. Row navigation is a link/chevron, never a
repeated primary button. Before approval, map mixed spacing relationships to
`skills/01-foundations.md` rather than using one wrapper `gap`.

Bad: two generic cards, seven `[Open]` buttons, or a page map with no shell.
Good: labelled shell + H1/context, purpose-bound regions, one primary action,
wide/narrow maps, and state/recovery annotations.

## Visible packet

Required for design-gated work: the human must **see** the buildable desktop and
mobile wireframes in the conversation and explicitly approve **that packet**
before markup or CSS.

- Saving the packet only to a plan file is not approval.
- “Make the page”, “oks dale”, `continua`, or a resume after a model/context
  switch is not approval unless the human has already been shown that wireframe
  and said yes to it.
- If the packet was never shown, stop with `Needs your decision` and show it.
- Record the approved packet version, date, and channel after a real yes.

## Approval workflow

1. **Classify and Orient.** Name the route/trigger, read the required
   artifacts and hard contracts, and confirm whether an exemption applies.
   Completion: the route classification and reading attestations are recorded.
2. **Plan.** Record intent mode, job, audience, surface, states, non-goals,
   section purposes, and the structural intent wireframe. For redesign, record
   the current-state audit; for visual-migration, record preserved contracts.
   Completion: another person can understand what is proposed without markup or
   invented component details.
3. **Architect.** Inspect route, data, permissions, templates, components, and
   server/no-JS behavior; first resolve Gelium theme/recipe/skin, registered
   interaction behaviors, copy/locales, and MEDIA-* direction. Then produce the
   buildable wireframe and section/component mapping. Completion: material
   incompatibilities are resolved or escalated.

4. **Request a decision.** Show the buildable wireframe in the conversation,
   then record `approved`, `changes-requested`, `declined`, or bounded
   `exception`, with date/author and packet version. Completion: gated work has
   an explicit blocking outcome before implementation. `pending` blocks Build.
5. **Implement within scope.** Follow the approved section order and reuse
   existing components/tokens. Completion: implementation matches the packet,
   or any material deviation is recorded and re-approved.
6. **Verify.** Run rendered-HTML, responsive, theme, state, accessibility,
   no-JS, and mechanical checks from skills 07, 09, and 11. Completion: all
   failures are fixed, escalated, or explicitly documented.

## Intentional gate bypass / exception record

An exempt change does not need a ceremonial wireframe. When a change that would
normally be gated intentionally proceeds without prior approval, record an
explicit exception rather than silently bypassing the gate. This is a short
note in the change/PR/task record, not a new blocking workflow:

```text
WIREFRAME gate exception: yes
Scope that would normally require approval:
Reason for bypass:
Risk boundary / files and routes:
Maintainer or user decision:
Follow-up review required: yes | no
Verification evidence:
```

Use this only when timing, incident recovery, a tiny contained expansion, or
another stated constraint makes prior approval impractical. Keep the scope
bounded and complete the record as part of implementation. Accessibility fixes,
bug fixes, contract corrections, component/mechanical changes, and already
approved small adjustments remain ordinary exemptions and do not require an
exception record unless their scope expands into a gated change.

## Handoff

After approval, hand off to implementation with the packet and its decision
record. The implementation handoff is:

```text
skills/01–02 → skills/08 → skills/10 → skill 11 → skill 12 (when gated)
→ registered components → tokens/skin → skill 11 critique → skills/09 → skills/07
```

For exempt small work, use the narrowest relevant path and do not manufacture a
wireframe. In all cases, plan before markup, preserve semantic HTML and
existing Gelium contracts, and keep the primary flow functional with JS
 disabled.

## Verification checklist

- [ ] Scope is new screen, new flow, substantial redesign, or documented exemption.
- [ ] Gated work has a visible packet in conversation and an approved packet before markup.
- [ ] Exempt work is not delayed by an unnecessary approval ceremony.
- [ ] Any intentional bypass is recorded as an explicit exception with bounded scope.
- [ ] Packet covers intent mode, current-state audit when applicable, desktop/mobile
  order, section contracts, Gelium behavior direction, theme/recipe/skin,
  copy/locales, media, states, recovery, accessibility, no-JS behavior, and
  DESIGN-MEMORY reuse.

- [ ] Rendered implementation matches the approved structure or records a
      material deviation and decision.
- [ ] Skills 07, 09, and 11 verification is complete.
