# Skill: Product reasoning (discovery before screens)

Run this AFTER picking a SURFACE mode (`skills/02`) and BEFORE writing markup.
It finds MISSING product-level UX — events with no surface, feedback with no
timing, jobs with no path — that screen-by-screen work silently skips. It feeds
the DoD (`skills/07`); it does not replace it.

Six steps. Step 0 selects the relationship to any existing experience; the
remaining steps are enumeration or yes/no questions. If a step surfaces gaps,
name them and ask the user before designing around them.

## Step 0 — Intent mode

Choose exactly one mode before screen architecture. Existing UI is evidence, not
a specification, and a reference is never permission to copy.

| Mode | Discovery obligation | Change boundary |
| --- | --- | --- |
| `new-product` | derive the experience from the product outcome and user job | define the target from first principles and known contracts |
| `redesign` | audit the current experience and classify `preserve`, `improve`, `remove`, or `unknown` with evidence | propose improvements, preserve valid contracts, and obtain approval for scope |
| `visual-migration` | inventory current routes, structure, behavior, content, and contracts | keep UX stable while mapping the approved experience to Gelium; use only when explicitly requested |

Record the selected mode, why it is appropriate, evidence, and scope boundary in
`PRODUCT.md` and `DESIGN.md`. Never infer `visual-migration` merely because a
screen already exists. For `redesign`, do not propose a target until the current
state and its evidence are recorded. For `visual-migration`, still escalate
critical accessibility, responsive, security, or contract failures instead of
silently preserving them.

## Sources (prior art)

| Topic | Source |
| --- | --- |
| Canonical task/event inventory | [GOV.UK patterns: "Ask users for…" / "Help users to…"](https://design-system.service.gov.uk/patterns/) |
| Notification inventory + importance grouping | [Material Design: notifications](https://m2.material.io/design/communication/notifications.html) |
| Post-launch gap signals | [HEART framework](https://kerryrodden.com/heart/) — Goals→Signals→Metrics |
| Ready-made inventories per product shape | Gelium domain skeletons: [Patterns](/docs/patterns) (`SKEL-FORUM`, `SKEL-CATALOG`, `SKEL-ADMIN-RESOURCE`, …) |

If your product matches a Gelium domain skeleton, start Step 1 from its table —
reuse the solved problem before re-deriving it.

## Step 1 — Inventory

Enumerate, purely from the domain (no UI yet):

- **Events**: "comment created", "follow added", "media deleted", "export
  failed", "invite expired"…
- **Actors**: who does each event — user, another user, system, scheduled job.
- **Artifacts**: what each event produces or touches — comment, follow,
  upload, report.
- **Lifecycle states** per artifact: draft / pending / failed / done / expired.

No design decisions here. A short flat list is the deliverable.

## Step 2 — Map coverage

For every **event × artifact × state**, name the UI surface that represents it
(view, list row, status badge, empty state). An unmapped cell is a candidate
gap.

Key question: *which objects in the data model have no view, list, or status
anywhere?* Those are the screens you were about to forget.

## Step 3 — Classify timing

Per event, pick one:

| Timing | Use when |
| --- | --- |
| Toast now | relevant only while the user is present (< 1s of attention) |
| Badge + count on persistent chrome | user must act eventually; count > 0 |
| Inbox on return | worth reading later, not urgent |
| Silent by design | routine success the user already saw happen |

Rules:

- Anything the user must **eventually act on** needs an always-visible
  indicator when count > 0. The badge is hidden at zero — never render a
  literal "0".
- Live feedback only for events relevant while the user is present; on-return
  events go to a page/inbox, not a toast.

## Step 4 — Walkthrough narratives

Narrate 2–3 end-to-end jobs out loud, **including failure and abandonment**:
"I follow someone — how do I learn they followed back?" "My upload fails at 80%
— what do I see tomorrow?"

Check per narrative:

- Will a first-time user know what to do?
- Will they see progress while it runs?
- Will they understand the result?
- Does it survive closing the tab mid-way?

## Step 5 — Audit pass

One-line self-checks:

- **Nielsen**: Is system status always visible? Recognition over recall?
  Error messages offer recovery? Consistent with the rest of the app?
- **Cognitive accessibility**: plain language? Undo available? Destructive
  actions confirmed? Help findable from where it's needed?
- **Kano must-be tier**: does every core job have a fallback/recovery path?
- The stranger test: could someone who has never seen this page grasp its
  purpose in 5 seconds?

## Worked micro-example: notifications

Data model has `notification {read bool}`. Plain text link in the nav fails:

- Step 3: unread count invisible → user must eventually act but has no badge.
- Step 4: "someone replies while I'm away — how do I ever find out?" → no.

Passing shape: bell icon in persistent chrome + unread badge (hidden at 0) +
live toast for events while present + history page for on-return. Every cell of
event × state now maps to a surface.

## Compact checklist

0. Intent mode selected and evidence/scope boundary recorded?
1. Events / actors / artifacts / lifecycle states enumerated?
2. Every event × artifact × state mapped to a surface? Any data-model object
   with no view anywhere?
3. Timing chosen per event? Action-required events have a badge when count > 0
   (hidden at 0)? On-return events on a page, not a toast?
4. 2–3 narratives walked including failure/abandonment? Survive tab close?
5. Audit pass clean: status visible, recognition over recall, error recovery,
   undo, confirmations, help findable, fallback paths, 5-second purpose test?

Any NO → name the gap and resolve it (ask the user if it's a product decision)
before markup.
