# ADR 0058: Folder rules are task groups with a path selector

- Status: rejected (no `[folders]` table)
- Date: 2026-11-01

## Context

Idea A2 of the design review: replace selector rules with folders. Configuration would inherit down
the tree (workspace, then `packages/`, then one project), the nearest rule winning, and a rule for
"all packages" would be a rule placed on `packages/`. The first slice proposed `[folders."packages"]`
(or a `smith.folder.toml`) taking the keys of `[projects.<name>]`.

## Finding

The folder rule already exists. A `[[task_groups]]` selector may be a glob over a project's folder:
`select = ["packages/*"]` is the rule placed on `packages/`
(`projectgraph.matches` tests the name, the root and `path.Match` of both). `smith explain` already
names it as the source: `from smith.toml [[task_groups]] select = ["packages/*"]`. The selectors
`kind:`, `lang:` and `tag:` are what a folder cannot say: a Node library and a Go library live in
one `libs/` folder, and a group "all Node libraries" has no folder.

## Decision

Do not add `[folders]`. A second table that also sets task fields would give three layers that
overlap (`[tasks]`, a folder, a task group) and two ways to write the same rule. The "nearest wins"
rule is also not an improvement: task groups are ordered by the file, which a person can read; depth
is implicit.

What a folder rule cannot do today is set `tags` or release settings for a folder. That is a real
gap but a small one. If the owner wants it, extend `[[task_groups]]` (rename to `[[groups]]`) with
`tags`, the same selectors and the same order, rather than add a table.

## Done

The manual says that a folder is a selector. `[boundaries]` and `--projects` already read the same
selector words, so one syntax stays.

## Cost of the rejection

A newcomer must learn that `libs/*` is a selector. The configuration reference now says so where task
groups are introduced.
