# Wireframe approval packet

> Copy this packet into the change record for a `design-gated` change. It is a
> review artifact, not permission to commit, publish, deploy, or claim rendered
> quality before Build and Audit.

## Identity

```text
Packet version:
Change:
Date:
Author:
Route: design-gated
Scope: new screen | new flow | substantial redesign
Intent mode: new-product | redesign | visual-migration
Existing route and contracts:
Product job / audience:
Primary action:
Non-goals:
Shell scope: new | redesigned | existing-unchanged | out-of-scope
Header / navigation decision:
Footer decision:
```

## Orient receipts

```text
Product/design artifacts:
Gelium entrypoint / decision pack / skills:
Vocabulary and component registry:
Nearby reusable patterns:
Hard route, permission, data, and no-JS/server contracts:
Open uncertainty:
```

## Plan — intent wireframe

Describe the requested user outcome without inventing unavailable components,
data, media metadata, or pixel polish. ASCII maps SCREEN blocks from skill 02.

```text
Screen / SURFACE:
SCREEN type:
User job and audience:
Global shell treatment (header/navigation/footer):
Major regions in reading order:
Primary and supporting actions:
States and recovery:
ASCII notation and legend:
WIDE structural wireframe:
NARROW structural wireframe:
Constraints and non-goals:
Current-state audit or preserved-contract summary:
```

## Architect — buildable wireframe

Reconcile the intent with actual route, data, permissions, templates, registered
components, tokens, and no-JS/server contracts. This is the wireframe approved
for Build.

```text
Route / handler / template:
Data and permission boundary:
Global shell mapping and unchanged/new boundary:
Section inventory and SECTION-CONTRACT mapping:
Gelium theme class and visual recipe/skin:
Gelium behavior and state mapping:
Copy, voice, and locale constraints:
Media direction and MEDIA-* references:
Component and token mapping:
URL, form, POST+303, 422, and validation contracts:
No-JS and accessibility behavior:
WIDE buildable wireframe:
NARROW buildable wireframe:
State/recovery wireframes or annotations:
Material mismatch, exception, or escalation:
```

## Criteria plan (prebuild)

```text
Hierarchy and DOM order:
Action hierarchy and boundaries:
Responsive and class-routed theme/recipe intent:
Interaction behavior and no-JS fallback:
Copy/localization and media constraints:
States and recovery:
Accessibility and no-JS:
Preserved contracts:
DESIGN-MEMORY reuse decision:
```

## Decision

```text
Decision: approved | changes-requested | declined | exception
Approver:
Date / channel:
Approved scope or bounded exception:
Reason, risk boundary, owner, and follow-up (if exception):
```

A chat decision is valid only when the human was shown the buildable wireframe
in the conversation and this packet records the exact approved scope, approver,
date/channel, and packet version. `changes-requested`, `declined`, and a packet
that was never shown block Build. An exception must remain bounded and receive a
rendered audit.

## Postbuild handoff

After Build, attach the rendered audit: wide/narrow, selected class-routed
themes, realistic content, states, focus/touch/keyboard, no-JS/server behavior,
detector output, tests/builds, and any exceptions or limitations.
