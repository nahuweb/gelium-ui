---
name: install-gelium
description: "Trigger: install Gelium, use Gelium UI, gelium-ui setup. Install the package and configure an agent-ready server-rendered UI project."
license: MIT
metadata:
  author: nahuweb
  version: "1.0"
---

# Install and use Gelium UI

Use this skill when a project needs to install, configure, or build with `gelium-ui`.

## Install

1. Install the package:

   ```bash
   npm install gelium-ui
   ```

2. Install project-local guidance for coding agents:

   ```bash
   bash node_modules/gelium-ui/install-agents.sh --project .
   ```

3. Commit `.agents/skills/gelium-ui/` when the project shares its agent guidance.

## Use

1. Import the shipped CSS:

   ```css
   @import "gelium-ui/dist/gelium.css";
   ```

2. Put a theme class such as `theme-material` or `theme-basecoat` on `<html>`.
3. Compose semantic, server-rendered HTML with Gelium classes and templates. Keep the primary flow functional with JavaScript disabled.
4. Add HTMX or `gelium.js` only as progressive enhancement.
5. Before changing UI, read the installed `AGENTS.md`, `llms-ux.txt`, and `SKILLS.md`. Run `bash scripts/ux-detect.sh` and the relevant project tests.

## Rules

- Use Gelium for HTML, CSS, themes, templates, and optional enhancement.
- Do not introduce a client-side component framework or duplicate Gelium tokens.
- For a new screen or flow, follow `skills/00-agent-routing.md` before writing markup.
