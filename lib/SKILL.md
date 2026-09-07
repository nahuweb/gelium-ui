---
name: gelium-ui
description: "Trigger: gelium-ui, Gelium UI, UI screen, UI component, Tailwind, HTMX. Load Gelium's project-local guidance before planning or implementing server-rendered HTML UI."
version: 0.6.7
license: MIT
---

# Gelium UI agent skill

## Activation Contract

Activate when Gelium UI is installed, mentioned, or used to build a Tailwind/HTMX interface. This skill is project-local guidance, not optional package documentation.

## Hard Rules

1. Before planning or writing UI, read `AGENTS.md`, `llms-ux.txt`, and `SKILLS.md` from this skill directory.
2. Use `skills/00-agent-routing.md` first. It selects the smallest safe route: `direct-exempt`, `delegated-direct`, `design-gated`, `escalate`, or `full-sdd`.
3. After routing, read only the downstream skills required by that route.
4. Never claim Gelium guidance was applied unless the required files were actually loaded; report missing or unreadable files and stop when the route requires them.

Gelium is server-rendered HTML + Tailwind CSS + HTMX, not a client-side framework. The linked files remain the source of truth for UX and delivery contracts.
