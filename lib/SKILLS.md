# gelium-ui — Skills for agents

Actionable decision skills. Read `skills/00-agent-routing.md` first, then load
the ones relevant to the selected route. Do not impose design-gated ceremony on
direct-exempt work. The full decision-id pack (SURFACE / SCREEN / ARCH / SECTION / WF / DATA /
FEED / JOURNEY / MEDIA / SKEL) is in `llms-ux.txt`.

| # | Skill | What it answers |
| --- | --- | --- |
| 00 | `00-agent-routing.md` | outcome-first route selection and delegation boundary |
| 01 | `01-foundations.md` | tokens, themes by class, 0-JS, layout, mobile guardrails |
| 02 | `02-surface-and-screens.md` | SURFACE mode + screen types → layout blocks |
| 03 | `03-forms-and-controls.md` | which control, form layout, 422 validation |
| 04 | `04-state-and-feedback.md` | states + FEED channel decision matrix |
| 05 | `05-server-contracts.md` | GET / POST+303 / 422 / gelium:toast wire |
| 06 | `06-mobile-and-a11y.md` | touch, keyboard, semantics, contrast, media |
| 07 | `07-dod-and-antislop.md` | Definition of Done checklist + anti-slop |
| 08 | `08-product-reasoning.md` | discovery workflow: find missing product-level UX before drawing |
| 09 | `09-usability-checklist.md` | per-screen binary usability checklist |
| 10 | `10-page-section-architecture.md` | purpose-first page and section contracts before components |
| 11 | `11-design-criteria.md` | checkable visual judgment, reuse, and resilient polish |
| 12 | `12-wireframe-approval.md` | conditional approval gate for new screens, flows, and substantial redesigns |
| 13 | `13-reference-selection.md` | semantic section/component reference selection before wireframes |
| 14 | `14-component-implementation.md` | new component contract, TDD, ownership, no-JS, and verification |
| 15 | `15-install-and-use.md` | install the npm package, install agent guidance, and compose a consumer app |

## Proportional workflow and records

Classify first: `direct-exempt` for an understood bounded fix without a page or
flow architecture change; `delegated-direct` for broad context or multi-file
work without a design change; `design-gated` for a new screen, new flow, or
substantial redesign; `escalate` for a decision that cannot safely be inferred;
and `full-sdd` only when cross-cutting work benefits from durable OpenSpec
artifacts. The design-gated path is:

```text
ROUTE → ORIENT → PLAN → ARCHITECT → APPROVE → BUILD → AUDIT → RELEASE
```

Orient reads product/design artifacts, vocabulary, registry, and hard contracts.
Plan selects `new-product`, `redesign`, or explicit `visual-migration`, then produces the job, audience, states, section purposes, and an intent wireframe. Redesign records a current-state audit; visual migration records preserved UX and server contracts. Architect
checks real routes, data, permissions, templates, components, and no-JS/server
contracts, then resolves Gelium behavior patterns, theme/recipe/skin, copy/locales,
and MEDIA-* direction before producing the buildable wireframe for approval. Prebuild
records decisions and a criteria plan; rendered evidence belongs in Audit after
Build. Start records from `skills/templates/gate-ledger.md` and
`skills/templates/wireframe-approval-packet.md`. For design-gated work the
packet, visible wireframe, and human approval of that packet are required
before markup. These are attestations and review records, not proof of reading
or authority to commit, publish, or deploy.

## Required architecture handoff

For a design-gated change, keep the gates in order: foundations → surface →
product reasoning → page/section → design criteria §1–§3 → conditional
visual/interaction direction → conditional wireframe approval → registered components
→ tokens/skin → rendered design
criteria audit → usability → DoD. A direct-exempt change uses the narrowest
relevant skills and checks. If `PRODUCT.md` or `DESIGN.md` is absent or
incomplete in a consumer repo, run the plain-language brief before a
design-gated sequence; stop only for unresolved decisions that change page or
flow architecture.

Protocol IDs: `ARCH-PRODUCT`, `ARCH-PAGE`, `ARCH-SECTION`, `ARCH-COMPONENTS`,
`ARCH-TOKENS`, `SECTION-CONTRACT`, `SECTION-HIERARCHY`, `SECTION-ACTION`,
`SECTION-REVELATION`, `SECTION-RECOVERY`, `WF-CURRENT-STATE`, `WF-DIRECTION`, `WF-ARCH`, `WF-SECTION-AUDIT`.

## Description-to-name resolver

**Before creating any new component**, resolve informal UI descriptions ("the
dark see-through layer behind a popup", "the eye icon on the password field")
against `lib/ui-vocabulary.md` — an attributed glossary mapping ~76 UI elements
to canonical names, common aliases, and API symbols. Then check
`docs/gelium-ui-vocabulary.md` §8 (resolved naming conflicts) and compose a
registered component instead of inventing one. See the "Vocabulary resolution"
step in `skills/02-surface-and-screens.md`.

Vocabulary source: namethatui.com (used with attribution; removal on request).

## Install these skills into your agent tool

Copy this folder into your agent's skill directory so the LLM loads it in any
project that uses `gelium-ui`:

```bash
bash install-agents.sh          # copies skills/ + AGENTS.md + llms-ux.txt
```

Targets detected: `~/.hermes/skills/gelium-ui`, `~/.cursor/skills/gelium-ui`,
`~/.claude/skills/gelium-ui`. See `install-agents.sh --help`.
