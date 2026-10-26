# smith typecheck

`smith typecheck` runs the `typecheck` task of each selected project: it checks types
without producing output (`tsc --noEmit` for Node stacks). Languages whose compiler checks
as part of `build` have no separate typecheck task. It is [`smith run typecheck`](run.md)
with the task fixed.

## When to use it

- A quick correctness check that does not write `dist/`.
- In CI through [`smith ci verify`](ci.md), which runs `lint`, `typecheck`, `test` and
  `build` in that set.

## Synopsis

```sh
smith typecheck [project]... [-p <projects>] [--affected [--base <branch>]] [--parallel <n>] [--bail] [--no-cache] [--json | --json-events]
```

## Arguments and flags

<!-- smith:begin typecheck -->
Run the typecheck task (same as `smith run typecheck`)

Usage: `smith typecheck [project]... [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `project...` | no | projects to run (the same as `-p`) |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--affected` | boolean | off | only projects affected by changes since --base |
| `--bail` | boolean | off | stop starting tasks after the first failure |
| `--base <base>` | string | the pull request target, else the release branch | base branch for --affected (default: the PR target, else the release branch) |
| `--json` | boolean | off | print the summary as JSON |
| `--json-events` | boolean | off | print the run as JSON lines, one event each (see docs/reference/serve.md), and nothing else |
| `--no-cache` | boolean | off | run every task even if a cached result exists |
| `--parallel <parallel>` | integer | the number of CPUs | tasks to run at once (default: CPU count) |
| `-p, --projects <projects>` | string, repeatable |  | projects to run: names, globs, kind:app, lang:go, tag:x, !exclude |

<!-- smith:end typecheck -->

Selector syntax and flag interactions are on [`smith run`](run.md#arguments-and-flags).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Every typecheck passed or came from the cache. |
| 1 | A typecheck failed, or the selection matched nothing. |

## Files and environment

Reads sources and `tsconfig.json`; writes the task cache. Environment variables are those
of [`smith run`](run.md#files-and-environment).

## Examples

```sh
smith typecheck
```

```text
web:typecheck ▶ npm run typecheck --silent
ui:typecheck ▶ npm run typecheck --silent
ui:typecheck ✔ 1.058s
web:typecheck ✔ 1.058s
✔ 2/2 tasks succeeded (0 from cache) in 1.063s
```

Only Node projects (`web`, `ui`) have the task in this workspace. With a type error in
`ui`:

```text
ui:typecheck │ src/failing.spec.ts(1,1): error TS2593: Cannot find name 'test'. Do you need to install type definitions for a test runner? ...
ui:typecheck ✖ exited with code 2
  ✖ ui:typecheck exited with code 2
✖ 1/2 tasks succeeded (1 from cache, 1.058s saved) in 931ms
```

The exit code of smith is 1; `2` is the compiler's.

## Common errors

| Message | Cause and remedy |
|---|---|
| `✖ <project>:typecheck exited with code 2` | `tsc` found type errors; the VS Code extension also puts them in the Problems panel ([the extension](vscode.md)). |

## Related

[`smith run`](run.md), [`smith lint`](lint.md), [`smith build`](build.md).
