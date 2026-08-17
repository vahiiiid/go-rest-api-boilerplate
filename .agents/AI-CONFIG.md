# GRAB — AI Configuration Guide

This document explains where AI configuration lives, which tool reads which file, and where to put new content. It exists so that when a project convention changes, the person (or agent) making the change knows every file that needs to move together.

## File Inventory

| File | Tool(s) | Loading condition | Content type |
|---|---|---|---|
| `AGENTS.md` | Codex, JetBrains AI, Claude Code (via `@import`), any AGENTS.md-compliant tool | Always | **Source of truth.** Full project guidelines: architecture, workflow, conventions, common tasks |
| `CLAUDE.md` | Claude Code | Always | Single line: `@AGENTS.md` |
| `CLAUDE.local.md` | Claude Code only | Always (gitignored) | Personal overrides — never committed |
| `.github/copilot-instructions.md` | GitHub Copilot (VS Code, JetBrains, Visual Studio) | Always | Condensed copy of `AGENTS.md`, manually synced |
| `.cursor/rules/grab.mdc` | Cursor | Always (`alwaysApply: true`) | Condensed copy of `AGENTS.md`, manually synced |
| `.windsurf/rules/grab.md` | Windsurf | Always ("Always On") | Condensed copy of `AGENTS.md`, manually synced |

## Why Condensed Copies Instead of Imports

`AGENTS.md` is the canonical, comprehensive guide. Claude Code officially supports importing it with `@AGENTS.md` syntax in `CLAUDE.md`, so that file stays a one-liner.

Copilot, Cursor, and Windsurf don't support that kind of cross-file import, and their rule files are meant to be short (large rule files either get truncated or eat context budget on every prompt). So `.github/copilot-instructions.md`, `.cursor/rules/grab.mdc`, and `.windsurf/rules/grab.md` are deliberately condensed, hand-maintained adapters — not full copies of `AGENTS.md`.

## Maintenance Rule

**`AGENTS.md` is the source of truth. When you change a project-wide convention there — architecture pattern, error-handling API, migration naming, testing pattern, commit format — check whether the same convention appears in the three condensed adapters and update it there too:**

- `.github/copilot-instructions.md`
- `.cursor/rules/grab.mdc`
- `.windsurf/rules/grab.md`

Each of those files carries a comment at the top pointing back to `AGENTS.md` as the source. They don't need to match `AGENTS.md` word-for-word (they're condensed on purpose), but they must not contradict it.

Content that's only relevant to one tool (a Cursor-only setting, a Copilot IDE-support note) can live only in that tool's file.

## Personal Local Overrides (Not Committed)

### `CLAUDE.local.md` (Claude Code)

Copy `CLAUDE.local.md.example` to `CLAUDE.local.md` in the repo root. Claude Code reads it automatically in addition to `CLAUDE.md`. It's gitignored.

### Copilot (VS Code Settings)

Copilot personal preferences are set in VS Code User Settings (not per-repo, so not gitignored):

1. Open VS Code Settings → search for `github.copilot.chat.codeGeneration.instructions`
2. Add entries like:
   ```json
   "github.copilot.chat.codeGeneration.instructions": [
     { "text": "I prefer table-driven tests with explicit subtests." }
   ]
   ```

### Cursor

Create `.cursor/rules/personal.mdc` (gitignored per your own global gitignore, or keep it out of version control manually) with `alwaysApply: true` and your preferences.

## How Each Tool Reads Configuration

- **Codex**: reads `AGENTS.md` natively, no import needed.
- **Claude Code**: reads `CLAUDE.md` (which imports `AGENTS.md` via `@AGENTS.md`), plus `CLAUDE.local.md` if present.
- **GitHub Copilot**: reads `.github/copilot-instructions.md` automatically in Copilot Chat (VS Code, JetBrains IDEs, Visual Studio).
- **Cursor**: reads `.cursor/rules/*.mdc` files; `grab.mdc` has `alwaysApply: true`.
- **Windsurf**: reads `.windsurf/rules/*.md` files marked "Always On".
- **JetBrains AI Assistant**: reads `AGENTS.md` directly.
