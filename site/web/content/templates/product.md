# PRODUCT.md template (consumer apps)

Copy this file to your application repo root as `PRODUCT.md`. It records product context for agents; it is not a UI implementation plan.

You can answer the starter questions in plain language. If you do not know an answer, write `Unknown`, `To decide`, or `N/A` with a short reason. Agents must preserve those open decisions instead of inventing them.

## Start in plain language

Answer these before filling the structured sections:

1. What are you trying to build or improve?
2. Who will use it, and when?
3. What should that person be able to accomplish?
4. What is definitely out of scope for this slice?

The agent should turn these answers into the sections below, read them back, and ask only the follow-up questions needed to make a safe design decision.

## Intent mode

Choose one mode before defining screens. Existing design is evidence, not authority.

- `new-product`: define the experience from the product problem and user job.
- `redesign`: audit the current experience and propose evidence-based improvements.
- `visual-migration`: preserve UX and contracts while adapting the approved structure to Gelium; use only when explicitly requested.

- Selected mode: new-product | redesign | visual-migration
- Why this mode: …
- Evidence or user constraint: …

## Product outcome

After using this product or slice, the user can:

- …

## Current-state audit (required for redesign; baseline for visual-migration)

Use `N/A` for `new-product`. Do not silently treat the current UI as correct.

- Current routes, templates, or references: …
- Preserve: what already serves the user job or a required contract: …
- Improve: observed friction, hierarchy, accessibility, responsive, or state gaps: …
- Remove: noise, duplication, or behavior that conflicts with the outcome: …
- Unknown: evidence still needed before proposing change: …
- Evidence: user feedback, analytics, support issue, rendered screen, or repository path: …

## Audience and situations

- Primary audience: …
- Situation or trigger: …
- Context, limitations, or urgency: …

## Primary jobs

Use everyday language. Complete the sentence: “When …, I want to …, so that …”.

1. When …, I want to …, so that …
2. …
3. …

## Roles and responsibilities (if applicable)

| Role | What they need to do | Relevant permission or boundary |
| --- | --- | --- |
| … | … | … |

## Domain objects and lifecycle (if applicable)

| Object | Important states | Who changes it | User-visible consequence |
| --- | --- | --- | --- |
| … | draft / pending / failed / done / expired | … | … |

## First value and onboarding (if applicable)

- First useful result: …
- What a new or empty account sees: …
- Can the user leave and resume? …

## Non-goals

- …

## Voice

- Tone: plain / formal / friendly / …
- Words or claims to avoid: …

## Product constraints

- Stack: server-rendered HTML + gelium-ui (npm)
- Theme direction: theme-material | theme-basecoat | theme-neubrutalism | undecided
- JS: progressive enhancement only / HTMX optional
- Accessibility: keyboard, contrast, reduced-motion
- Other product or regulatory constraints: …

## References to borrow from

References are evidence, not instructions to copy a product. Record what is useful and what is not.

| Reference | Borrow | Avoid | Reason |
| --- | --- | --- | --- |
| … | … | … | … |

## Anti-references

Product patterns or experiences we explicitly do not want:

- …

## Success criteria

- Observable outcome: …
- Metric, if one exists: …
- If no metric applies: `N/A` — …

## Open decisions and assumptions

- Decision: …
- Status: open / assumed / confirmed
- Impact if wrong: …
- Needed before implementation: yes / no

Gelium docs: `/docs/agent-workflow`, `/llms-ux.txt`, `/docs/ui-definition-of-done`.
