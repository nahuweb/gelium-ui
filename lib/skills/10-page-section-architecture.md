# Skill: Page + Section Architecture

Use this skill after product reasoning and surface selection, before choosing
components or writing markup. It makes every major page region answer a specific
reader need instead of treating a page as an unordered component catalog.

## Required decision chain

Follow this order. Do not skip a layer or use a later layer to replace an earlier
decision.

| ID | Decide | Record |
| --- | --- | --- |
| `ARCH-PRODUCT` | Product intent | audience, user job, outcome, non-goal, and relevant lifecycle/event consequences from `skills/08-product-reasoning.md` |
| `ARCH-PAGE` | Page/Surface | URL, SURFACE mode, SCREEN type, one primary action, JOURNEY/DATA/FEED IDs, chrome, and server/no-JS contract |
| `ARCH-SECTION` | Section purpose | one purpose-bound contract for every major page region |
| `ARCH-COMPONENTS` | Components | registered semantic primitives that fulfill each section contract |
| `ARCH-TOKENS` | Tokens/Skin | existing semantic type, spacing, color, motion, and class-routed theme/recipe/skin tokens |

The chain is:

```text
Product intent → Page/Surface → Section → Components → Tokens/Skin
```

`skills/07-dod-and-antislop.md` remains the entry gate: read consumer
`PRODUCT.md`/`DESIGN.md` when present, or run the plain-language brief from
`llms-ux.txt` when they are absent or incomplete. Record unknowns and assumptions;
stop only when an unresolved decision would change page or flow architecture.
This protocol extends the existing product, registry-first, token, and DoD gates;
it does not replace them.

## SECTION-CONTRACT

Complete this contract for every major page region before selecting components.
A section is a coherent answer to one reader need within a page. It is not “a
card,” a decorative wrapper, or repeated card anatomy.

```text
SECTION-ID / name:
Purpose: What user need this region satisfies now.
Audience and moment: Who reads it, and at what point in the page/journey.
Input / evidence: Data, state, or context it receives.
Output / next move: What understanding or action the reader leaves with.
Hierarchy: Entry, primary, supporting, or recovery; why it has that rank.
Action policy: Primary / secondary / none; relation to the page’s single primary action.
Revelation: Always visible / conditional server-rendered state / user-invoked disclosure; why.
    Composition: Semantic landmark or element plus registered Gelium components.
    Interaction behavior: registered Gelium behavior/recipe, trigger, and no-JS/server fallback.
    State and recovery: Rest, loading, empty, error, success where applicable; no-JS behavior.

Boundary and rhythm: What belongs inside vs outside; relationship-based --ui-space-* gaps.
Accessibility/content: heading/label/landmark, reading order, announcement, concise factual copy.
Verification evidence: rendered-HTML assertions and mechanical checks that prove the contract.
```

## Protocol rules

### SECTION-HIERARCHY

- Assign every top-level page child a section purpose. Decorative wrappers and
  repeated card anatomy are not automatically sections.
- Establish entry → primary → supporting → recovery order. DOM order must match
  the intended reading order; a later support section cannot impersonate the
  primary task.
- Group elements that answer one need, and separate unrelated needs. Use semantic
  landmarks/headings and the existing spacing contract: `--ui-space-1/2` for an
  element and its metadata, `--ui-space-3` for sibling items,
  `--ui-space-4/6` for groups, and `--ui-space-8` between sections.

### SECTION-ACTION

- The page keeps one primary action.
- A section may own only a supporting or local action unless it is the declared
  primary-action section.
- Select components after the section purpose is known. A card is for a repeated,
  self-contained instance; it is not a substitute for page context or a section
  boundary.
- Resolve names in `lib/ui-vocabulary.md`, then compose registered components.
  Escalate a new primitive only when the vocabulary and registry process prove a
  real gap.

### SECTION-REVELATION

- Put identity and task-critical “now” information in the initial page scan.
- Use conditional server-rendered sections for meaningful state. Use native
  disclosure only for optional depth.
- Never hide a necessary action or recovery path behind hover or JavaScript.
- State which reader and moment a section serves. Do not mix reader context,
  authoring controls, and recovery as undifferentiated content.

### SECTION-RECOVERY

- A data or action section identifies its applicable rest, loading, empty, error,
  and success behavior plus the server/no-JS path.
- Do not copy a page-global state blindly into every section.
- Preserve a way forward from empty, error, and partial states; persistent
  feedback is not a toast-only outcome.

## Bounded critique workflow

### WF-ARCH — between `WF-SHAPE` and `WF-BUILD`

1. Read `PRODUCT.md` and `DESIGN.md` when present; otherwise run the
   plain-language brief from `llms-ux.txt`, then run the `skills/08`
   product-reasoning inventory.
2. Name the URL, SURFACE mode, SCREEN type, one primary action, and relevant
   `JOURNEY-*`, `DATA-*`, and `FEED-*` IDs.
3. Draw a numbered DOM outline of major regions only. Do not choose CSS or
   components yet.
4. Complete one `SECTION-CONTRACT` per major region. Delete, merge, split, or
   escalate any region without a distinct purpose.
5. Map sections to semantic elements, registered components, and existing tokens.
   Reject a new primitive unless the vocabulary/registry process proves the gap.

### WF-SECTION-AUDIT — bounded step inside `WF-AUDIT`

1. Run a five-second scan: title, context, primary action, and task-critical state
   are recognizable before optional depth.
2. Run action competition: count page-primary variants; section-local actions stay
   visually and semantically subordinate unless they are the declared page action.
3. Run revelation and recovery: identify the first visible state, every conditional
   state, the no-JS path, and a way out of empty, error, or partial states.
4. Inspect rendered HTML in narrow and class-routed dark conditions, then run
   tests/detectors. Record only failures with a protocol ID and rendered evidence.

Reviewers trace each rendered region back through `Section → Page/Surface →
Product intent`. A gap is missing intent, an uncontracted section, an unregistered
component, or token/skin misuse.

## Worked application: Public Feed

**Route:** `/recipes/public-feed`
**Source:** `site/web/templates/recipe-public-feed.html` and
`internal/app/recipe_public_feed.go`

| Ordered section | Purpose and existing contract |
| --- | --- |
| Page context | Establish scanning context with `recipe-pf-header`, its H1, and concise description before interaction or list content. |
| View selection | Change the feed lens with `nav.ui-tabs`, real GET links, `aria-current`, and `?view=` state; it does not replace the page’s read/scan job. |
| Activity feed | Deliver the primary task through `ol.recipe-pf-list` of `article.ui-card` items in reverse chronology. |
| Local reaction | Keep per-item native `POST /recipes/public-feed/{id}/react` → 303 reactions local and subordinate to discovery. |
| Collection recovery | Explain empty/loading/error conditions with a CTA/retry path; loading must not flash empty. |
| Refresh support | Offer a non-primary native POST refresh with no-JS and HTMX-consistent results. |

The feed list is the section. Each feed card is a repeated entity/event instance;
its header, body, and footer are card anatomy, not automatic page sections.

## Worked application: Post detail / rich article

**Route:** `/recipes/rich-article`
**Source:** `site/web/templates/recipe-rich-article.html` and
`internal/app/recipe_rich_article.go`

| Ordered section | Purpose and existing contract |
| --- | --- |
| Context and identity | Article header, eyebrow, one H1, lead, and byline identify the post and reading promise before prose. |
| Primary reading body | `recipe-rich-article-prose` provides readable structured content with descriptive H2/H3 headings. |
| Evidence/media | Local picture, native video/captions/fallback, audio/transcript, and safe-embed fallback support comprehension. |
| Supporting reference | The captioned data table provides structured facts without hijacking the read flow. |
| Related activity | `ui-list` adds adjacent context after the primary reading need. |
| Recoverable states | Skeleton, empty, and error alert states remain legible and actionable. |
| Related navigation | `nav[aria-label="Related content"]` gives the reader a next path after content and recovery. |

This is the existing post-detail/read-detail analogue, not a nonexistent
`/recipes/post-detail` route. It is an integration fixture with no mutation or
reply action, so it demonstrates read-detail section architecture—not a complete
forum-thread lifecycle.

## Pointers

- Human companion: `/docs/page-section-architecture`.
- Compact agent pack: `llms-ux.txt` (`ARCH-*`, `SECTION-*`, `WF-ARCH`, and
  `WF-SECTION-AUDIT`).
- Existing workflow: `skills/02-surface-and-screens.md` → `skills/08-product-reasoning.md`
  → this skill → `skills/07-dod-and-antislop.md` and
  `skills/09-usability-checklist.md`.
