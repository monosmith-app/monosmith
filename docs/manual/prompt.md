# smith prompt

`smith prompt` prints one line for a shell prompt: the smith version, the number of projects,
the failed checks and warnings of the last [`smith doctor`](doctor.md) run, and the projects
that the changes in the working tree touch. It is made to be called on every prompt, so it
answers in about 50 ms: it never runs doctor, never opens a session and never starts a plugin.

## When to use it

- To show the health of the workspace in Oh My Posh, Starship or powerlevel10k. The
  [shell prompt guide](../guides/shell-prompt.md) has a snippet for each.
- To feed a custom prompt segment or a status bar with `--json`.

## Synopsis

```sh
smith prompt [--json]
```

## Arguments and flags

<!-- smith:begin prompt -->
Print a one-line status for a shell prompt, from the last doctor run

Usage: `smith prompt [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print the facts as JSON for a custom prompt segment |

<!-- smith:end prompt -->

## What the line says

```text
smith 2.2.0 · 34 projects · × 3 ▲ 4 · actions, smith
```

| Part | Meaning |
|---|---|
| `smith 2.2.0` | The smith version. |
| `34 projects` | The projects of the workspace, as the last doctor run saw them. |
| `× 3` | Failed checks of the last doctor run. Left out when there are none. |
| `▲ 4` | Warnings of the last doctor run. Left out when there are none. |
| `✓` | Shown in place of the two counts when the last run had no failure and no warning. |
| `actions, smith` | The projects with changed files in the working tree (uncommitted and untracked), and the projects that depend on them. After three names the line says `+N`; `--json` lists all. |

The numbers are as old as the last doctor run. Run `smith doctor` to refresh them.

Before the first doctor run the line is the version alone (`smith 2.2.0`), and it never says
that something is wrong. Outside a smith workspace it prints nothing and exits 0.

## Exit codes

Always 0. A prompt must not fail: an unreadable summary, a folder that is not a git
repository and a folder outside a workspace all give a shorter line, or no line.

## Files and environment

Reads `.smith/prompt.json`, which a normal `smith doctor` run writes (not `--fix`). The file is
small, holds the workspace root it was written for (a copied folder is ignored), and is in
`.smith/`, which git ignores. It also runs one `git status`. It writes nothing.

## Examples

```sh
smith prompt
```

```text
smith 2.2.0 · 34 projects · × 3 ▲ 4 · actions, smith
```

```sh
smith prompt --json
```

```json
{"version":"2.2.0","projects":34,"doctor":{"failed":3,"warnings":4,"checkedAt":"2026-11-10T00:28:10Z"},"affected":["actions","smith"]}
```

Before the first doctor run:

```text
smith 2.2.0
```

## Related

[Shell prompt guide](../guides/shell-prompt.md), [`smith doctor`](doctor.md),
[`smith affected`](affected.md) (the same idea against a base branch, with the project graph).
