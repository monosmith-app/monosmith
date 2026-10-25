# smith test

`smith test` runs the `test` task of each selected project: the test runner its stack
chose (Jest or Vitest for Node, `go test`, pytest, and so on). Where the task says
`^build` (Node projects do), the projects it imports are built first. A project whose inputs have not changed since a passing run is replayed
from the cache, with its recorded output. It is [`smith run test`](run.md) with the task
fixed.

## When to use it

- Locally, before a commit; in CI the same task runs inside [`smith ci verify`](ci.md).
- With `--affected` to test only what the changes touch.

## Synopsis

```sh
smith test [project]... [-p <projects>] [--affected [--base <branch>]] [--parallel <n>] [--bail] [--no-cache] [--json | --json-events]
```

## Arguments and flags

<!-- smith:begin test -->
Run the test task (same as `smith run test`)

Usage: `smith test [project]... [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `project` | no | projects to run (the same as `-p`) |

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

<!-- smith:end test -->

Selector syntax and flag interactions are on [`smith run`](run.md#arguments-and-flags).
`--no-cache` is the way to rerun a test that depends on something smith cannot see (the
network, the clock).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Every test task passed or came from the cache. |
| 1 | A test task failed, or the selection matched nothing. |

## Files and environment

Reads project files (a task's `inputs`) and the lockfile; writes the task cache and any
test output files the runner produces. Environment variables are those of
[`smith run`](run.md#files-and-environment).

## Examples

```sh
smith test
```

```text
✔ 4/4 tasks succeeded (0 from cache) in 2.694s
```

(The per-task log lines are above that summary; see [`smith run`](run.md#examples).)

A failing suite, its output and the summary:

```sh
smith test -p ui
```

```text
ui:test ▶ npm run test --silent
ui:test │ FAIL src/failing.spec.ts
ui:test │   ● fails
ui:test │ 
ui:test │     boom
ui:test │ 
ui:test │     > 1 | test("fails", () => { throw new Error("boom") })
ui:test │         |                             ^
ui:test │       2 |
ui:test │ 
ui:test │       at Object.<anonymous> (src/failing.spec.ts:1:29)
ui:test │ 
ui:test │ 
ui:test │ Test Suites: 1 failed, 1 passed, 2 total
ui:test │ Tests:       1 failed, 1 passed, 2 total
ui:test │ Snapshots:   0 total
ui:test │ Time:        0.265 s, estimated 1 s
ui:test │ Ran all test suites.
ui:test ✖ exited with code 1
  ✖ ui:test exited with code 1
✖ 0/1 tasks succeeded (0 from cache) in 999ms
```

The exit code is 1.

## Common errors

| Message | Cause and remedy |
|---|---|
| `✖ <project>:test exited with code 1` | The runner reported failures; the output above has them. |
| `no selected project has a [test] task` (exit 0) | None of the selected projects defines `test`; a stack without a test runner such as `node-ts` has none. `smith explain <project>` lists a project's tasks. (The wording is the one `smith run deploy` printed; it was not captured for `test`.) |

## Related

[`smith run`](run.md), [`smith lint`](lint.md), [`smith why-miss`](why-miss.md),
[Tasks, cache and affected](../reference/tasks-and-cache.md).
