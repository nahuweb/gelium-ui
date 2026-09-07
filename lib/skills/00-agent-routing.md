# Gelium agent routing

> Canonical first decision layer for agents using Gelium UI. This document
> selects the smallest useful implementation route; it does not replace the
> Gelium UX, server, accessibility, or component contracts.

## Start with the outcome

Ask what outcome the user needs, then classify the current action before loading
all downstream workflow material. File count is a discovery signal, not a risk
score and never an automatic reason to use SDD.

| Route | Use when | Required behavior |
| --- | --- | --- |
| `direct-exempt` | The change is already understood, bounded, and does not change screen/flow architecture. Typical examples are copy, token, selector, accessibility, bug, or existing-contract corrections. | Work inline. Load only the narrow skills and checks. No wireframe or ledger ceremony. |
| `delegated-direct` | Understanding needs four or more files, broad read-only research, a writer changing two or more non-trivial files, or fresh context materially improves the action. | Delegate one bounded exploration or one focused writer. Pass exact paths and applicable skill files. The parent owns scope, integration, and verification. No SDD state is created merely by delegating. |
| `design-gated` | A new screen, new flow, or substantial redesign changes hierarchy, major regions, primary action, reading order, or ownership boundary. | Use `ROUTE → ORIENT → PLAN → ARCHITECT → APPROVE → BUILD → AUDIT → RELEASE`. Build waits for the visible, human-approved Architecture packet. |
| `escalate` | Product intent, permissions, scope, data, external effect, or residual risk cannot converge safely from current artifacts. | Stop at the smallest concrete decision. State the cause, impact, and choices. |
| `full-sdd` | Durable proposal, specification, design, and task artifacts materially reduce ambiguity across normative layers or repositories. | Use OpenSpec/SDD explicitly. It is optional for ordinary screens and is never inferred from file count alone. |

Delegation is **per action**, not a project-wide mode. A delegated exploration
can be followed by direct implementation; a delegated writer does not receive
approval authority. Use at most two focused read-only explorers, one production
writer, and a fresh reviewer when the candidate warrants it.

## Public progress states

Expose only the state needed for the user's next decision:

- **Working** — the candidate can still change.
- **Needs your decision** — automatic convergence is unsafe; present the cause,
  impact, and concrete choices.
- **Checking** — tests, detector, rendered evidence, or bounded review are being
  collected.
- **Ready** — the selected route has sufficient evidence for ordinary repository
  delivery review.

These are projections, not a replacement for detailed ledger statuses.

## Machine-readable route handoff

Repository tooling may validate the agent's selected route before creating a
ledger or loading downstream skills:

```bash
go run ./cmd/gelium-preflight route --route delegated-direct --format json
```

The command reports the route, status, next action, and whether the Gelium
design gate is required. It validates the selected route; it does not infer the
route from file count, inspect user intent, or authorize delivery. An
`escalate` result exits non-zero so a hook can stop before implementation.

## Optional startup hook

Projects that use Gentle AI can run the packaged hook once at agent startup:

```bash
./node_modules/gelium-ui/scripts/agent-start.sh .
```

When `gentle-ai` is installed, it runs the idempotent project-local skill
registry refresh. Without Gentle AI, it reports that the refresh was skipped
and exits successfully. It never chooses a task route, enables RDD/SDD, edits
the repository, or replaces the agent's own startup hooks.

## Design-gated Gelium sequence

Only after the route is `design-gated`:

1. **Orient** — read the relevant Gelium entrypoint, decision pack, skills,
   vocabulary/registry, consumer `PRODUCT.md`/`DESIGN.md` when present, and hard
   route, permission, server, and no-JS contracts. If either consumer artifact is
   missing or incomplete, run the plain-language brief and record missing decisions
   honestly.
2. **Plan** — select `new-product`, `redesign`, or explicit `visual-migration`; state the user job, audience, SURFACE/SCREEN, primary action,
   states, constraints, non-goals, current-state evidence when applicable, and intent wireframe. Do not fabricate
   components, data, identities, or media metadata.
3. **Architect** — inspect the real route, handler, templates, data,
   permissions, shared chrome, references, components, responsive behavior,
   themes, accessibility, and no-JS fallback. Resolve Gelium theme/recipe/skin, registered interaction behaviors, copy/locales, and MEDIA-* direction before producing the buildable packet and
   section contracts.
4. **Approve** — show the complete desktop/mobile packet in the conversation.
   Record the approved scope, packet version, approver, date, and channel.
   `continua`, a model switch, or a plan file alone is not approval.
5. **Build** — implement only A plus explicitly accepted C items, preserving
   Gelium contracts and unrelated worktree changes.
6. **Audit** — collect rendered evidence after a candidate exists: wide/narrow,
   themes, states, keyboard/focus/touch, no-JS, detector, tests, and build.
7. **Release** — report clean pass, pass-with-exceptions, or failure. Evidence
   never authorizes commit, push, publish, deploy, or archive.

`direct-exempt` and `delegated-direct` do not inherit the design gate. If work
reveals a material architecture change, re-route it before writing further.

## Evidence and authority boundaries

- Ledgers attest structured facts and evidence references; they cannot prove
  cognitive reading or subjective approval.
- Memory may seed discovery, but current repository artifacts override it.
- Review follows a candidate and is bounded by the candidate's scope. It does
  not own delivery.
- OpenSpec/SDD artifacts are durable planning records. A screen ledger is
  operational evidence, not a second SDD transaction system.
- Ordinary repository policy owns commit, push, pull request, release, and
  deployment decisions.

## Downstream Gelium documents

After routing, load only what the selected route needs:

- `AGENTS.md` → `llms-ux.txt` → `SKILLS.md` for the Gelium decision chain.
- `skills/01`–`skills/09` for applicable foundations, surface, forms, states,
  server, mobile/a11y, DoD, product reasoning, and usability checks.
- `skills/10`–`skills/13` for architecture, criteria, approval, and references
  on design-gated work.
- `docs/gelium-ui-agent-prompts.md` for downstream component, recipe, theme,
  and documentation-specific procedures; it is not an alternate route selector.
- `docs/gelium/agent-contract.md` for composition artifacts after routing; it is
  not a replacement for this routing layer.

This contract adopts Gentle AI's organic routing, compact public states,
per-action delegation, candidate-first review, and delivery boundary. It does
not import Gentle AI's RDD transaction lineages, locks, hashes-as-authority, or
runtime delivery controls.
