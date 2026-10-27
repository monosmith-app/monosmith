---
name: cli-output-design
description: Design what smith prints in a terminal or a CI log (pipeline steps, command results, reports): banners, tables, status lines, wording. Use when adding or changing any text a smith command prints, especially `smith ci <phase>`, `smith release`, `smith doctor` and `smith convert` output.
---

# Designing smith's text output

The TUI set the visual language; a command line and a CI log use the same one in plain text. Read
`apps/smith/tui/theme/icons_config.go` first: the symbols (`✓ ✗ ▲ ● ○ ⚡ ↷`, and the ASCII set
`[v] [x] [!] [~] [ ] [c] [-]`) are the single source. Do not invent new ones.

## What a reader wants

Someone opens a failed or a green pipeline and wants three answers in this order, without scrolling:
**what ran**, **what happened**, **what to do next**. Everything else is detail and goes in a
collapsed group.

## A step has three parts

1. **Banner** (one to three lines): the step's name and, in plain words, what it is about to do and on
   what. `Verify: checking 23 projects (lint, typecheck, test)`.
2. **Status lines** while it runs: one line per unit of work, only when the unit finishes or fails.
   Never stream a tool's raw output outside a group.
3. **Result**: a table or a few lines, then a closing line that says the outcome and the time.
   `✓ Verify passed in 3m 12s: 23 projects, 412 tests`. A failure closes with the cause and the
   remedy (`→ run smith run test -p <project> to see it`).

Say so when nothing happened and why: `Release: nothing to release (no feat: or fix: commit since
the last release)`. Silence reads as a bug.

## Banners

A banner marks the start and the end of a step and the things that need a decision. Plain text:

```
══ Setup ═══════════════════════════════════════════
Installing dependencies for node and go, then preparing 23 projects.
```

- Width at most 80 columns, so it survives a narrow log viewer. One banner style; the step's name
  inside the rule. ASCII fallback: `== Setup ==`.
- Highlight what matters with the CI system's own annotations, not with colour: GitHub
  `::notice::`, `::warning::`, `::error::`; Azure `##vso[task.logissue type=warning;]` and
  `type=error`. Use them for things a person should see in the run summary (a warning, a failure, a
  release's result), never for progress.
- Use the job summary when the CI has one (GitHub `$GITHUB_STEP_SUMMARY`, Azure
  `##vso[task.uploadsummary]`): the same result table in Markdown, once per phase.

## Tables

Use a table when there are several things with the same properties (projects, versions, packages).
Rules:

- Header row, then one row per item, columns padded with spaces (not tabs). No box drawing in
  tables: it breaks when copied.
- Left-align text, right-align numbers and durations. Name first, status second, detail last.
- Status is a symbol plus a word (`✓ passed`), so it still reads without colour or the symbol font.
- Order: failures first, then the rest alphabetically; or the order of the pipeline if it matters.
- Cap the rows (about 30): beyond that print the failures and a count (`... 18 more passed`).
- One truncation rule: cut the middle of long names with `…`, never the status.

```
Project        Lint  Types  Tests         Time
core           ✓     ✓      ✓ 42 passed   12s
atlassian      ✓     ✓      ✗ 3 failed    41s
logger         ✓     ✓      ○ none         -
```

For a release, the row is the decision, with its reason:

```
Package    Was     Now     Why
core       1.4.2   1.5.0   feat: add search (3 commits)
logger     2.0.1   2.0.2   fix: handle empty input
```

## Words

- Plain language, verb first, present tense for what is happening, past for what is done.
- Name the thing and the number: `23 projects`, not `all projects`; `smith 1.102.2 on linux/amd64`.
- Explain a term the first time it appears or do not use it. No internal names (`{smith}`, plugin
  ids, step keys) in text a person reads.
- Every warning and error names the remedy on the next line, starting with `→`.
- Durations as `12s`, `3m 12s`; sizes as `4.2 MB`; counts with the noun.

## CI log rules

- Plain text only: no ANSI colour, no cursor movement, no spinners, no progress bars (a bar is
  noise in a log). Check `NO_COLOR` and "not a terminal" and keep the same words.
- Fold detail with the CI's groups (`::group::` / `##[group]`) via the existing `group()` and
  `endGroup()` helpers in `apps/smith/cipipeline`. The banner and the result stay outside the group.
- Raw tool output (npm, test runners) goes inside a group; when a step fails, print the tail
  outside it (see `step_output_tail_algorithm.go`).
- Deterministic: the same run prints the same lines, so logs can be diffed. No timestamps in the
  text (the CI adds them).
- Keep a successful run short: a step is about 5 to 15 lines outside groups.

## Before you ship output changes

1. Run the command and read it as a stranger: do the first three lines say what this step is?
2. Run it with `NO_COLOR=1` and piped to a file: still readable?
3. Put a real sample in the PR body, so the owner can judge the wording and the layout.
4. A golden or test pins the lines that matter (the banner, the table header, the closing line).
