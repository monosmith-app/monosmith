# smith lint

`smith lint` runs the `lint` task of each selected project: ESLint with
`@monosmith/eslint-config` for Node stacks, `golangci-lint` for Go, `ruff check` for Python, and
the linter each other stack names. It is [`smith run lint`](run.md) with the task fixed.

smith owns the lint setup, so a repository needs no linter configuration: for Go and
Python smith writes its default to `.smith/defaults/` and passes it to the tool. A
`.golangci.yml` (Go) or a `ruff.toml` / `[tool.ruff]` table (Python) in the repository is the
override and replaces the default; see [Go](../plugins/using/catalogue/go.md#lint-configuration)
and [Python](../plugins/using/catalogue/python.md#tool-configuration).

A lint task is also how the vertical-slice rules are enforced: for Node, the shared
ESLint config carries them; Go, Python and the other languages get a separate
`slice-check` task (see [`smith slices check`](slices-check.md)).

## When to use it

- Before a commit, or in CI through [`smith ci verify`](ci.md).
- With `--affected` to lint only the projects the changes touch.

## Synopsis

```sh
smith lint [project]... [-p <projects>] [--affected [--base <branch>]] [--parallel <n>] [--bail] [--no-cache] [--json | --json-events]
```

## Arguments and flags

<!-- smith:begin lint -->
Run the lint task (same as `smith run lint`)

Usage: `smith lint [project]... [flags]`

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

<!-- smith:end lint -->

Selector syntax and flag interactions are on [`smith run`](run.md#arguments-and-flags).
There is no `--fix` flag; apply fixes with the linter itself (`npx eslint --fix`) or the
command `smith explain <project>:lint` prints.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Every lint task passed or came from the cache. |
| 1 | A lint task failed, or the selection matched nothing. |

## Files and environment

Reads source files and lint configuration (`eslint.config.mjs`, `.golangci.yml`, `ruff.toml`); writes the
task cache and, for Go and Python, smith's default config under `.smith/defaults/`. Environment variables are those of [`smith run`](run.md#files-and-environment).

## Examples

```sh
smith lint -p ui
```

```text
ui:lint ▶ npm run lint --silent
ui:lint ✔ 9.376s
✔ 1/1 tasks succeeded (0 from cache) in 9.382s
```

A failure, in a file that breaks the project's style rules:

```text
ui:lint │ ~/work\acme\libs\ui\src\failing.spec.ts
ui:lint │   1:6   error  Strings must use singlequote  @stylistic/quotes
ui:lint │   1:39  error  Strings must use singlequote  @stylistic/quotes
ui:lint │
ui:lint │ ✖ 2 problems (2 errors, 0 warnings)
ui:lint │   2 errors and 0 warnings potentially fixable with the `--fix` option.
ui:lint ✖ exited with code 1
  ✖ ui:lint exited with code 1
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `✖ <project>:lint exited with code 1` | The linter reported problems; fix them or run the linter's fixer. |
| `lib has no lint task, so CI never checks it` (a `smith doctor` warning) | The project's stack has no lint; add a lint task or a stack that has one. |

## Related

[`smith run`](run.md), [`smith typecheck`](typecheck.md), [`smith slices check`](slices-check.md),
[`smith doctor`](doctor.md).
