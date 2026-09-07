# Agent workflow

How humans and LLMs **work with Gelium** without leaving the product ethos: server HTML, token themes, sourced criteria (GOV.UK / USWDS / M3 / NNG), progressive JS only.

This page adapts useful **agent process** ideas (brief → shape → build → audit → polish) — not free-form “make it award-winning” aesthetics that fight Material/Basecoat tokens.

## Start here: route the outcome

The package source of truth is [`skills/00-agent-routing.md`](https://github.com/cryptonahue/gelium-ui/blob/main/lib/skills/00-agent-routing.md).
It chooses `direct-exempt`, `delegated-direct`, `design-gated`, `escalate`, or
`full-sdd` before this downstream UX workflow is loaded. Delegation is per
action, and delivery remains under ordinary repository policy.

## Ethos (non-negotiable)

| Do | Don’t |
| --- | --- |
| Prefer `ui-*` partials + `--ui-*` tokens | Invent a parallel CSS design system per screen |
| Map `FEED-*` / `DATA-*` / `JOURNEY-*` | Toast-only validation; blank tables |
| One theme class on `<html>` | Random hex, purple gradients “for polish” |
| 0-JS first; enhance | Require a SPA runtime for basic CRUD |
| Cite handbook + sources | Override Forms/Feedback with vibe |

## Surface modes

Choose **one mode per URL/surface** (not per product company). Same product can have a Persuade landing and an Operate admin.

| Mode | Visitor success | Bias | Gelium defaults |
| --- | --- | --- | --- |
| **Operate** | Finish a task | Scan, density, consistency, native controls | Admin recipes, data-table, forms, FEED-VAL, compact/cozy |
| **Read** | Understand | Measure 65ch, clear H1, in-page nav | Handbook-style docs, long prose |
| **Persuade** | Decide and act | One primary CTA, honest hierarchy | Hub/start, hero, short proof — still tokens, not stock SaaS gradients |
| **Experience** | Be inside the artifact | Content leads; chrome recedes | Galleries/media; don’t fake it on dense CRUD |

**Operate wins ties** for tools, settings, queues, and tables.

## Workflow passes (bounded)

Do **not** open-ended self-QA loops. Fixed passes:

| Pass | ID | Goal | Exit when |
| --- | --- | --- | --- |
| **1. Brief** | `WF-BRIEF` | Plain-language audience, job, outcome, scope, and intent mode; generate/read PRODUCT.md + DESIGN.md and ask only missing decisions | Job sentence + surface mode or explicit open decision |
| **2. Shape** | `WF-SHAPE` | Screen type, journey, data pattern, FEED plan, and current-state audit when applicable — **no markup yet** | IDs listed (SCREEN / JOURNEY / DATA / FEED) |
| **3. Architecture** | `WF-ARCH` | Major-region purpose, hierarchy, action, revelation, recovery, and structural wireframe — **no components yet** | Every major region has a `SECTION-CONTRACT` ([Page + section architecture](/docs/page-section-architecture)) |
| **4. Direction** | `WF-DIRECTION` | Gelium theme, visual recipe/skin, registered interaction behaviors, copy/locales, and MEDIA-* direction | Direction is recorded and layout-affecting choices are reflected in the buildable packet |
| **5. Build** | `WF-BUILD` | Approved partials, theme/recipe, behavior, copy, media, and server contracts | Renders; happy path works |
| **6. Audit** | `WF-AUDIT` | Technical + contract checks, including finite `WF-SECTION-AUDIT` (a11y basics, responsive, FEED/DoD) | Detector script + checklist clean or waived with reason |
| **7. Polish** | `WF-POLISH` | Align to theme, spacing tokens, copy (Content style) | One pass only — then stop |
| **Harden** (as needed) | `WF-HARDEN` | Empty/error/i18n overflow, edge cases | Edge paths covered |
| **Onboard** (as needed) | `WF-ONBOARD` | First-run + empty activation | Empty has CTA |

`shape` owns discovery; `architecture` contracts major regions; `direction` chooses Gelium behavior and visual treatment; `build` owns code; `audit` is evidence; `polish` is alignment — not a redesign.

### Structural wireframes

For every design-gated screen, show a clean ASCII wireframe before markup. It
must include the global shell (header/navigation and footer, including bands
marked `existing-unchanged`), numbered purpose-bound regions, one primary
action, both `WIDE` and `NARROW` layouts, and state/recovery annotations.
Mark shell decisions as new, redesigned, existing, or out of scope; an
unresolved shell decision blocks approval. Use the notation and technical
annotation rules in `skills/12-wireframe-approval.md`. ASCII communicates
structure and reading order, not pixel polish or a generic collection of
cards.

## Anti-slop (Gelium-aware)

Adapted from common AI-frontend failure modes; **scoped so they don’t ban the design system itself**.

| Avoid | Why | Prefer |
| --- | --- | --- |
| Nested cards inside cards for simple forms | Noise, false hierarchy | One surface; `divider` / sections |
| Stock purple→blue hero gradients | Generic AI marketing look | Theme primary/surface tokens |
| Icon-in-rounded-square above every heading | Cliché template | Icon only when it carries meaning |
| Gray text on saturated fills | Contrast failure | Theme fg/on-color tokens; check a11y |
| Bounce/elastic motion everywhere | Dated; motion policy | Short token motion or none ([Motion](/docs/motion)) |
| “Make it bolder” on Operate tables | Hurts scan | Compact density + clear primary |
| New font stack per feature | Breaks system | Theme `--ui-font-*` unless DESIGN.md says otherwise |
| Layout tables for non-tabular content | GOV.UK table guidance | `DATA-LIST` / cards ([Data display](/docs/data-display)) |

**Allowed when the theme/brief says so:** Material type stack, Basecoat direction, product brand fonts recorded in consumer DESIGN.md.

## Consumer context files

For apps **using** gelium-ui (not the monorepo itself):

| File | Purpose |
| --- | --- |
| `PRODUCT.md` | Product outcome, audience/situation, jobs, roles, lifecycle, onboarding, voice, scope, success criteria |
| `DESIGN.md` | URL-to-job inventory, surface/screen choices, chrome, theme, density, states, references, anti-slop |

Templates: [`/docs/templates/product`](/docs/templates/product) and [`/docs/templates/design`](/docs/templates/design) (also under `site/web/content/templates/`).

Agents: read these before `WF-SHAPE` when present. If either is absent or incomplete,
run the plain-language brief in `/llms-ux.txt`; do not demand product vocabulary
or invent the missing decisions silently.

### When the answers are incomplete

The person does not need product or design vocabulary. Start with four plain
questions: what are you building or improving, who uses it and when, what should
they accomplish, and what is out of scope. The agent drafts the two context files,
reads its interpretation back, and asks only follow-ups that affect architecture.
`Unknown`, `To decide`, and `N/A` are valid answers when the reason and impact are
recorded. This keeps the process accessible without allowing invented product intent.

## Install the agent skill

Install Gelium UI's guidance as a project-local skill so every agent working in
an application can discover the same contract:

```bash
npm install gelium-ui
bash node_modules/gelium-ui/install-agents.sh
```

This creates `.agents/skills/gelium-ui/` with the canonical `SKILL.md`,
`AGENTS.md`, `llms-ux.txt`, `SKILLS.md`, and downstream routing skills. Commit
that directory when the project shares its agent guidance.

For a global host installation, use an explicit target such as
`--claude`, `--cursor`, or `--codex`; use `--target <path>` for other hosts.
The skill must be loaded before planning or implementing Gelium UI, and its
linked guidance files are the source of truth for the UX contracts.

## Detectors

Run from repo root:

```bash
bash scripts/ux-detect.sh
```

Deterministic checks for handbook contracts, agent pack IDs, and hard anti-patterns (e.g. masking overflow on `body`/`html` in library/site CSS). Exit non-zero on failure.

## Map to existing docs

| Pass | Deep links |
| --- | --- |
| Shape | [Screens](/docs/screens), [Page + section architecture](/docs/page-section-architecture), [Journeys](/docs/journeys), [Data display](/docs/data-display), [Patterns](/docs/patterns) |
| Architecture | [Page + section architecture](/docs/page-section-architecture) — `SECTION-CONTRACT` before components or tokens |
| Build | Components sidebar, [Forms](/docs/forms), [Server contracts](/docs/server-contracts) |
| Audit | [Page + section architecture](/docs/page-section-architecture) `WF-SECTION-AUDIT`, [UI definition of done](/docs/ui-definition-of-done), [Feedback](/docs/feedback), [Responsive](/docs/responsive), [Accessibility](/docs/accessibility) |
| Polish | [Themes](/docs/themes), [Tokens](/docs/tokens), [Content style](/docs/content-style), [Density](/docs/density) |

## See also

- Agent pack: [`/llms-ux.txt`](/llms-ux.txt)  
- [UI definition of done](/docs/ui-definition-of-done)
