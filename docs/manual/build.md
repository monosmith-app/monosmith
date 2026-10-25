# smith build

`smith build` runs the `build` task: it compiles or bundles each selected project into the
workspace's `dist/<project dir>` folder, after building the projects it imports. It is
[`smith run build`](run.md) with the task fixed.

## When to use it

- Before you run an app, package it, or check that everything still compiles.
- To produce build outputs another task depends on (`^build` in
  [`smith explain`](explain.md) means "the build of the projects this one imports first").

For tests, lint and type checks use their own shortcuts; for something else, `smith run`.

## Synopsis

```sh
smith build [project]... [-p <projects>] [--affected [--base <branch>]] [--parallel <n>] [--bail] [--no-cache] [--json | --json-events]
```

## Arguments and flags

<!-- smith:begin build -->
Run the build task (same as `smith run build`)

Usage: `smith build [project]... [flags]`

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

<!-- smith:end build -->

Positional `project` names and `-p` are two ways to select projects; the selector syntax
(`web`, `apps/*`, `kind:app`, `lang:go`, `tag:x`, `!api`) and every interaction between the
flags are described on [`smith run`](run.md#arguments-and-flags).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Every build succeeded or came from the cache. |
| 1 | A build failed or was skipped because a dependency failed, or the selection matched nothing. |

## Files and environment

Writes each project's declared outputs, normally `dist/<project dir>` (`smith doctor`
checks that every build writes there), and the task cache. Environment variables are those
of [`smith run`](run.md#files-and-environment).

## Examples

One project; `web` has no changes since its last build, so a second run is a cache hit:

```sh
smith build -p web
```

```text
web:build ▶ npm run build --silent
web:build ✔ 935ms
✔ 1/1 tasks succeeded (0 from cache) in 959ms
```

```sh
smith build web
```

```text
web:build ✔ from cache
...
```

Every publishable package:

```sh
smith build -p kind:package
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `no project matches "nope" (projects: ...)` | `smith projects` lists the names. |
| `ui:build → the command tsc was not found: install it, ...` | Dependencies are not installed. `smith install`. |
| `✖ web:build exited with code 2` | The compiler failed; read the lines above it. |

## Related

[`smith run`](run.md), [`smith test`](test.md), [`smith ci verify`](ci.md),
[`smith why-miss`](why-miss.md).
