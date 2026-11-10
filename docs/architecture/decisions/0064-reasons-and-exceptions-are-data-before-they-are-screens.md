# ADR 0064: Reasons and exceptions are data before they are screens

- Status: accepted (command line and data); the editor screens are not built
- Date: 2026-11-01

## Context

Ideas F2 and F3 of the design review: show the reason for every default, so people trust it and
override it less blindly; make exceptions visible and cheap, with a calm list of divergences, each
with "why" and "undo".

## Decision

1. A reason is data next to the setting. The schema (`smith.schema.json`) gets an `x-reason` per
   setting, parsed into `configedit.Field.Reason`. It carries a reason only where one exists; a
   setting with none says "no reason is recorded" rather than a made-up one. Eleven settings
   have a reason now (commit_operations, release.branch, github_releases, api_check, cache size and
   directory, attest, agent files, remote, tag format, defaults vintage).
2. `smith why <setting>` answers for a setting (value and source, what it does, why the default,
   change, undo), next to the existing `smith why <file>`. A path that exists as a file is a file.
   `smith config list` appends the reason to the description.
3. `smith divergences` lists overrides, ejected files, drift and differing files, each with `why` and
   `undo`. It is a report: it never fails and never edits. "Never shame" is a rule for the wording
   and for the exit code.
4. The data is in JSON (`config list --json` has `reason`; `why <setting> --json`; `divergences
   --json`) so the TUI, the VS Code extension and the JetBrains plugin can show it. This brief
   builds no screen: the TUI is the work of Brief B, and the extensions follow it.

## Cost

A reason is a claim and must stay true when a default changes; a default changed on purpose now means
a reason edited in the same change. The `undo` of an ejected file is two steps (edit `smith.lock`,
then `smith derive --replace`) because smith has no command that gives a file back; that is a gap
to close with its own command, not to hide in prose.

## Not done

Reasons for the tables outside the fixed ones (`[tasks]`, `[dependencies]`, `[conventions]`), a
`doctor` section that links to `divergences`, and the screens.

## Sources

- `apps/smith/configedit/explain_setting_use_case.go`, `apps/smith/why/explain_setting_use_case.go`,
  `apps/smith/divergences/`
