# ADR 0084: One concept, one flag

- Status: accepted
- Date: 2026-11-04

## Context

The same idea had different flags in different commands. Selecting a project was `-p, --projects` in
`smith run`, `-w, --workspace` in `smith install` and `smith uninstall`, and `--project` (with `-p` in only some
commands) in `smith docker description`, `smith pipeline`, `smith release assets`, `smith release pub` and
`smith tauri icon`. In smith the workspace is the whole repository, so `--workspace` for one project is wrong.
The same drift existed for the language (`--lang` and `--language`), the output path (`--output` and `--out`),
"do not install" (`--no-install` and `--skip-install`) and `-y` (missing on three commands that take `--yes`).

Why it happened. `-p/--projects` came with the first task runner (`d6d1cb83`, 2026-10-15). `-w/--workspace` came one
day later with `install` and `uninstall` (`9bd55b02`, issue #100) and copied the word of npm's `--workspace`.
Each later `--project` was added by a feature commit that copied the nearest neighbour or named its own flag
(`71b806b3` pub, `db216228` Docker Hub description, `80d0b070` release assets, `dc04bbcf` tauri). No document
named the flags, no helper registered them, and no test compared commands. ADR 0062 audited the nouns of the
documentation, not the flags.

## Decision

1. **The vocabulary is data.** `libs/actions/flagvocabulary/flag_vocabulary.json` declares, for each shared
   concept, the flag, its shorthand, its synonyms, its value type, its meaning and the commands that take it.
   [The reference table](../../reference/flag-vocabulary.md) is generated from it.
2. **One concept, one flag.** A command registers a shared flag through package `apps/smith/commandflags`
   (`AddProjectFlag`, `AddLangFlag`, `AddOutputFlag`, `AddYesFlag`, `AddNoInstallFlag`). It does not choose a name.
   A new concept goes into the vocabulary first.
3. **Project selection is `-p, --project`**: repeatable and comma separated. `-w`, `--workspace` and `--projects`
   are synonyms with the same behaviour on every command that takes the flag. `--help` shows the flag once and
   names the synonyms: `(also -w, --workspace, --projects)`. Removing a synonym is one line in the JSON.
4. **Clients take the names from the vocabulary.** The TUI, the guided forms, the MCP tools and the hints
   call `flagvocabulary.Arg("project")`. The VS Code extension and the JetBrains plugin use constants that
   `go run ./flaggen` (in `libs/actions`) writes from the JSON: `flag.project` and `SmithFlags.PROJECT`.
   Every command line smith prints or builds uses the canonical spelling (`-p`).
5. **Tests guard it.** `TestEveryCommandUsesTheFlagVocabulary` builds the command tree in memory (no session, no
   plugin) and fails when a command defines a near-synonym of a shared concept in plain view, a shared flag
   with the wrong shorthand or without its synonyms, a command list that differs from the vocabulary, or one
   shorthand for two flags. A flag that looks like a shared one on purpose goes in `exceptions` of the JSON with its
   reason. `TestGeneratedFilesAreUpToDate` fails when a generated file is stale, and
   `TestClientsDoNotHandTypeSharedFlags` fails when a client types a shared flag by hand.

## Consequences

- A script that uses an old spelling keeps working. The synonyms are the only backward compatibility, and
  they are declared in one place.
- Flags that mean something else stay as they are and are listed in the exceptions: `graph --focus` (a
  neighbourhood, not a target), `add --like` (the project to copy, not the one to act on), `graph --format` and
  `release --format` (a layout, not the JSON switch).
- A person who adds a flag meets the test before the review does.

## Sources

- Owner request of 2026-11-04: "it makes no sense at all that each command uses different parameters for the same thing".
- [Flag vocabulary](../../reference/flag-vocabulary.md), [ADR 0029](0029-one-list-of-actions.md) (the same idea for actions).
