# smith dev

`smith dev` starts several projects together, such as a frontend and the API it calls.
What they depend on is built first; then every project's `start` task runs at once, each
line of output prefixed with the project's name, and Ctrl+C stops them all. It is
`smith run start -p <projects>` with room for every one to run at once, and `--dry-run`
prints that command.

## When to use it

- Local development of more than one long-running project.

For a single project, `smith run start -p web` is enough. For anything that finishes, use
[`smith run`](run.md).

## Synopsis

```sh
smith dev [project]... [--all] [--dry-run]
```

## Arguments and flags

<!-- smith:begin dev -->
Start several projects together, such as a frontend and the API it calls

Usage: `smith dev [projects...] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `projects...` | no | the projects to start; each needs a `start` task |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--all` | boolean | off | start every project that has a start task |
| `--dry-run` | boolean | off | print the equivalent smith run command and start nothing |

<!-- smith:end dev -->

- Name the projects, pass `--all` for every project that has a `start` task, or give neither
  and pick from a list in a terminal.
- A project without a `start` task is refused.
- The printed command carries a `--parallel` value large enough for every start task to run
  at once (6 for two projects and 7 for three in the trials; how it is computed was not
  verified).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Stopped by you, or `--dry-run` printed the command. |
| 1 | A project has no `start` task, nothing was named outside a terminal, or a started task failed. |

## Files and environment

Reads project tasks; the started processes write wherever they write. No smith-specific
environment variable. Tasks run through the same engine as [`smith run`](run.md), so the
cache and `SMITH_*` task variables apply.

## Examples

```sh
smith dev --all --dry-run
```

```text
→ smith run start -p api,web,web2 --parallel 7
```

```sh
smith dev nope --dry-run
```

```text
smith: nope has no start task (projects with one: api, web, web2)
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `name the projects to start, or pass --all (projects with a start task: api, web)` | Nothing was named and there is no terminal to pick in. |
| `<project> has no start task (projects with one: ...)` | Only those projects can be started. `smith explain <project>` shows the tasks. |

## Related

[`smith run`](run.md), [`smith explain`](explain.md), [Serving an app for a task](../reference/tasks-and-cache.md#serving-an-app-for-a-task).
