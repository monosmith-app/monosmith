# smith run

`smith run <task>...` runs one or more tasks on the workspace's projects, in dependency
order, and replays a task from the cache when nothing it depends on changed. It is the
engine behind [`smith build`](build.md), [`smith test`](test.md), [`smith lint`](lint.md) and
[`smith typecheck`](typecheck.md), which are `smith run` with the task fixed. CI runs the
same thing through [`smith ci verify`](ci.md).

A task is a name that projects define: the language and stack plugins supply `build`,
`test`, `lint`, `typecheck`, `package`, `start`, `slice-check` and others; `smith.toml` can
add more. [`smith explain`](explain.md) shows what a project's tasks are and where each
setting came from.

## When to use it

- Any task other than the four shortcuts: `smith run package`, `smith run start -p web`,
  or several at once, `smith run lint test --affected`.
- A script or CI step that wants the JSON summary or the event stream
  (`--json`, `--json-events`).

For the four common tasks use their shortcut; for several apps at once with prefixed output,
[`smith dev`](dev.md).

## Synopsis

```sh
smith run <task>... [-p <projects>] [--affected [--base <branch>]] [--parallel <n>] [--bail] [--no-cache] [--json | --json-events]
```

## Arguments and flags

<!-- smith:begin run -->
Run tasks on projects, in dependency order, from cache when nothing changed

Usage: `smith run <task>... [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `task...` | yes | the tasks to run, in order, such as `lint test` |

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
| `--verbose` | boolean | off | smith build: show the log of every project, not only the ones that fail |

<!-- smith:end run -->

The same flags apply to `build`, `test`, `lint` and `typecheck`, which take project names
as positional arguments (`smith test web ui`) instead of tasks.

Project selection (`-p`, comma-separated or repeated; positional names on the shortcuts):

| Selector | Selects |
|---|---|
| `web` | the project by name |
| `apps/*`, `web*` | a glob over names or directories |
| `kind:app`, `kind:lib`, `kind:package` | by kind |
| `lang:go` | by language |
| `tag:x` | by tag in `smith.toml` |
| `!api` | excludes what it matches |

A task also runs in the projects the selected ones depend on when the task is defined to
wait for them (`^build` in the output of `smith explain`).

Interactions:

- `--affected` runs only projects the changes since `--base` affect. `--base` defaults to
  the pull request target taken from the environment (`SMITH_BASE`, `GITHUB_BASE_REF`,
  `SYSTEM_PULLREQUEST_TARGETBRANCH`, `CI_MERGE_REQUEST_TARGET_BRANCH_NAME`, in that
  order), else the release branch. When smith cannot decide, every project is affected;
  the first line says why. See [`smith affected`](affected.md).
- `--json` and `--json-events` are mutually exclusive (`if any flags in the group [json
  json-events] are set none of the others can be`).
- `--json-events` prints one JSON object per line and nothing else on standard output (task
  output arrives as `task.output` events). The events are the ones
  [`smith serve`](serve.md) sends.
- `--no-cache` runs every task even when a cached result exists.
- `--bail` stops starting tasks after the first failure.
- `--parallel` is a task count, default the number of CPUs.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Every selected task succeeded or came from the cache, or nothing was selected. |
| 1 | A task failed or was skipped because a dependency failed, a selector matched nothing, or a flag is invalid. |

## Files and environment

Reads `smith.toml` and every project's manifest; writes the task cache (`.smith/cache`
unless `cache.dir` or `SMITH_CACHE_DIR` says otherwise) and each task's declared outputs
(usually under `dist/`). A task's process gets `SMITH_TASK` (the id, `web:test`),
`SMITH_TASK_KEY` (its cache key) and `SMITH_WORKSPACE_ROOT`. The remote cache is read from
`SMITH_CACHE_REMOTE` and signed with `SMITH_CACHE_KEY`; `SMITH_CACHE_READ_ONLY` makes it
read-only. See [Tasks, cache and affected](../reference/tasks-and-cache.md) and
[Configuration](../reference/configuration.md).

Output goes to standard output as `<project>:<task> <symbol> ...` lines. With `--json` the
same task log lines are printed first and the JSON summary follows on standard output
(checked with `2>NUL` and `>NUL`: nothing is written to standard error), so a parser must
take the document after the log, or use `--json-events`, which prints nothing else.

## Examples

All tests, twice. The second run replays from the cache:

```sh
smith test
```

```text
api:test ▶ go test ./...
tools:test ▶ {workspaceRoot}/.venv/Scripts/python.exe -m pytest -q -p no:cacheprovider
ui:test ▶ npm run test --silent
web:test ▶ npm run test --silent
api:test │ ?   	github.com/acme/acme/apps/api	[no test files]
tools:test │ .                                                                        [100%]
tools:test │ 1 passed in 0.01s
tools:test ✔ 1.644s
api:test │ ok  	github.com/acme/acme/apps/api/greeting	0.435s
api:test ✔ 2.262s
...
ui:test ✔ 2.671s
web:test ✔ 2.688s
✔ 4/4 tasks succeeded (0 from cache) in 2.694s
```

```sh
smith test
```

```text
api:test │ ?   	github.com/acme/acme/apps/api	[no test files]
api:test │ ok  	github.com/acme/acme/apps/api/greeting	0.435s
api:test ✔ from cache
...
✔ 4/4 tasks succeeded (4 from cache, 9.266s saved) in 4ms
```

Only what changed since `main`, after editing `apps/web/src/main.ts`:

```sh
smith run lint test --affected
```

```text
affected: changes since main → 1 project(s)
web:lint ▶ npm run lint --silent
web:test ▶ npm run test --silent
web:test │ Test Suites: 1 passed, 1 total
web:test │ Tests:       1 passed, 1 total
web:test │ Snapshots:   0 total
web:test │ Time:        0.257 s, estimated 1 s
web:test │ Ran all test suites.
web:test ✔ 1.268s
web:lint ✔ 4.243s
✔ 2/2 tasks succeeded (0 from cache) in 4.246s
```

The JSON summary of a failed run, printed after the task log (the exit code is 1; a cached
task has `"status": "cached"` and a `savedNs` field):

```sh
smith test --json -p ui
```

```json
{
  "results": [
    {
      "task": "ui:test",
      "status": "failure",
      "key": "b670e175c30730cb267c0de91a81d6516946f460386b0f8a47c7521a885dd511",
      "durationNs": 1192561100,
      "reason": "exited with code 1"
    }
  ],
  "durationNs": 1193593200
}
```

The event stream, for a tool:

```sh
smith test --json-events -p ui
```

```text
{"event":"run.started","run":"run-20261025T090946-1","seq":1,"time":"...","command":"smith run test -p ui","tasks":["ui:test"]}
{"event":"task.queued","run":"run-20261025T090946-1","seq":2,"time":"...","task":"ui:test"}
{"event":"task.started","run":"run-20261025T090946-1","seq":3,"time":"...","task":"ui:test"}
{"event":"task.output","run":"run-20261025T090946-1","seq":4,"time":"...","task":"ui:test","stream":"combined","line":"Test Suites: 1 passed, 1 total"}
...
{"event":"task.finished","run":"run-20261025T090946-1","seq":9,"time":"...","task":"ui:test","status":"cached","cacheKey":"e99ec6ac..."}
{"event":"run.finished","run":"run-20261025T090946-1","seq":10,"time":"...","ok":true,"summary":{"passed":0,"failed":0,"cached":1,"skipped":0,"durationMs":3}}
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `requires at least 1 arg(s), only received 0` | `smith run` needs a task: `smith run build`. |
| `no project matches "nope" (projects: api, tools, ui, web)` | The selector matched nothing. `smith projects` lists the names. |
| `no selected project has a [deploy] task` (exit 0) | No selected project defines the task; nothing ran. `smith explain <project>` lists its tasks. |
| `the command tsc was not found: install it, or on a package.json script run smith i so node_modules/.bin has it` | The task's program is not installed. Run `smith install`. |
| `invalid argument "x" for "--parallel" flag` | `--parallel` takes a whole number. |
| `if any flags in the group [json json-events] are set none of the others can be` | Pick one of `--json` and `--json-events`. |
| A task shows `✖ exited with code N` | The tool failed; its output is above. `smith why-miss <project>:<task>` says what changed since it last passed. |

## Related

[`smith build`](build.md), [`smith test`](test.md), [`smith lint`](lint.md),
[`smith typecheck`](typecheck.md), [`smith dev`](dev.md), [`smith affected`](affected.md),
[`smith why-miss`](why-miss.md), [`smith explain`](explain.md),
[Tasks, cache and affected](../reference/tasks-and-cache.md).
