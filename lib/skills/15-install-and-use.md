# Skill: Install and use Gelium UI

Use this skill when a project needs to install, configure, or build with the `gelium-ui` npm package.

## Install

1. Install the package from the project root:

   ```bash
   npm install gelium-ui
   ```

2. Install Gelium's project-local agent guidance:

   ```bash
   bash node_modules/gelium-ui/install-agents.sh --project .
   ```

3. Commit `.agents/skills/gelium-ui/` when the project shares its agent guidance.

## Use

1. Import the shipped stylesheet from the application's CSS entrypoint:

   ```css
   @import "gelium-ui/dist/gelium.css";
   ```

2. Put one shipped theme class on `<html>`, such as `theme-material` or `theme-basecoat`.
3. Compose server-rendered HTML with Gelium's native classes and templates. Start with semantic HTML; keep the primary flow functional with JavaScript disabled.
4. Add HTMX or `gelium.js` only as progressive enhancement, not as a required runtime.
5. Read the installed `AGENTS.md`, `llms-ux.txt`, and `SKILLS.md` before planning UI, then run `bash scripts/ux-detect.sh` and the project's relevant tests.

## Decision gates

- Use the package for HTML, CSS, themes, templates, and optional progressive enhancement.
- Do not introduce a client-side component framework or duplicate Gelium tokens in consumer CSS.
- If the project needs a new screen or flow, follow `skills/00-agent-routing.md` before writing markup.

## Output contract

A successful setup has `gelium-ui` in `package.json`, the package stylesheet imported, a theme class on `<html>`, and project-local agent guidance installed when agents will modify the UI.

## References

- `README.md` — package API and consumer examples.
- `install-agents.sh` — project-local and host-specific guidance installation.
- `AGENTS.md` — canonical Gelium UI workflow and contracts.
