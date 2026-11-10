# ADR 0067: `[[groups]]` replaces `[[task_groups]]` and also sets tags and release

- Status: accepted
- Date: 2026-11-01

## Context

`[[task_groups]]` sets task overrides for the projects its selectors name, including folder globs
([ADR 0058](0058-folder-rules-are-task-groups-with-a-path-selector.md)). ADR 0058 refused a
`[folders]` table and left one gap: a folder rule cannot set `tags` or release settings. It said the
owner could extend the group and rename it. The owner decided to do both.

## Decision

- The table is `[[groups]]`. Its selector is unchanged (`select = [...]`: a name, a glob such as
  `packages/*`, `kind:`, `lang:`, `tag:`; a project must match all of them; groups apply in file
  order).
- A group sets at least one of `tasks`, `tags` and `release`. `tags` are added to the tags of each
  selected project. `release` is the release table of each selected project that sets none of its
  own; a project's own `[projects.<name>.release]` replaces it whole, with no field merge. A group
  cannot name the published package (`npm`, `pypi`, `nuget`, `rubygems`, `pub`), because that name is
  one project's.
- Tags apply in file order, so a later group can select by `tag:` a tag an earlier group gave.
- `smith.toml` has no `task_groups` any more, and the parser has no alias for it. A file that still
  has it does not load with a list of parse errors. It stops with one sentence, in the style of the
  newer-smith message: `smith.toml uses [[task_groups]], which is now [[groups]]: run smith doctor
  --fix`.
- `smith doctor` is the one command that runs on such a file. It reads `smith.toml` as text, reports
  the old key as one finding, and `smith doctor --fix` (with `--yes` or after a confirmation) renames
  the table headers with a diff. The edit is on the text, so comments, order and every other line stay.
  Then doctor opens the workspace and runs its other checks.
- The list of renamed keys is a table in `workspace` (`removed_keys_policy.go`). A later rename adds a
  row; the message, the finding and the fix follow from it.

## Consequences

- One word, `group`, for the selector table. The doctor check is "every group names a project".
- The rename is not a schema bump and not a defaults vintage
  ([ADR 0054](0054-defaults-have-vintages.md)): the schema version means a change smith cannot read, and the
  vintage means defaults that moved. Here the file is read the same way once the name is changed, and
  doctor does the change.
- `smith convert` and `smith doctor --fix` (task folding) write `[[groups]]`.
- The VS Code extension and the JetBrains plugin do not print the key name, so they need no change.

## Sources

- Owner decision of 2026-11-01, after the design review (idea A2).
- [ADR 0058](0058-folder-rules-are-task-groups-with-a-path-selector.md).
