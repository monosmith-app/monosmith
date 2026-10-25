# smith explain

`smith explain` shows how a project's tasks resolved and where each of their fields came
from: the language plugin, a stack's plugin, a convention, or `smith.toml`. For a project
made from a recipe it also lists what the recipe decided for it (its choices, the plugin
at each level and the plugins' settings), each with the recipe file and line it came from.

The text starts with one line per task (its command, whether it is cached, what it waits
for), then each task's fields. Where a value came from is said once for all of them and
marked only where it differs.

## When to use it

- A task does something unexpected and you want to know which plugin or setting decided it.
- You want a task's exact command, working directory, inputs, outputs and cache setting.

For why a task would **run again** rather than come from the cache, use
[`smith why-miss`](why-miss.md).

## Synopsis

```sh
smith explain <project>[:<task>] [--task <task>] [--recipe] [--json]
```

## Arguments and flags

<!-- smith:begin explain -->
Show how a project's tasks resolved, and where each setting came from (why-miss says why one reran)

Usage: `smith explain <project>[:<task>] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `project[:task]` | yes | the project, optionally with one of its tasks, such as web:build |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print JSON |
| `--recipe` | boolean | off | print only what the project's recipe decided, and where |
| `--task <task>` | string |  | explain only this task (same as <project>:<task>) |

<!-- smith:end explain -->

- `<project>:<task>` and `--task <task>` are the same; either explains one task.
- `--recipe` prints only what the project's recipe decided, and where; with `--json` it
  prints that for tools. It is an error for a project that was not made from a recipe.
- `--json` prints an array of tasks, each `{ "id", "description", "fields": [ { "name",
  "value", "source" } ] }`, with the fields `command`, `cwd`, `dependsOn`, `inputs`,
  `outputs`, `env` and `cache`.

The JSON and text show these task fields:

| Field | Meaning |
|---|---|
| `command` | What runs. `{workspaceRoot}` and `{smith}` are placeholders. |
| `cwd` | The directory it runs in, workspace-relative. |
| `dependsOn` | What runs first: `^build` is the build of the projects this one imports; a bare name is another task of this project. |
| `inputs` | The files whose content goes into the cache key. |
| `outputs` | The directories the cache stores and restores. |
| `env` | Environment variables part of the key. |
| `cache` | Whether results are cached. |

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Printed. |
| 1 | No such project or task, or `--recipe` for a project without a recipe. |

## Files and environment

Reads `smith.toml` and asks the plugins. Writes nothing.

## Examples

One task:

```sh
smith explain web:test
```

```text
web:test — Run unit tests
  command    npm run test --silent
  cwd        apps/web
  dependsOn  ^build (build of the projects it imports, first)
  inputs     **/*, {workspaceRoot}/package-lock.json, {workspaceRoot}/tsconfig.base.json, {workspaceRoot}/eslint.config.mjs
  outputs    —
  env        —
  cache      true

Every value above comes from plugin node, except where marked.
```

The whole project starts with the summary:

```sh
smith explain web
```

```text
web has 7 tasks:
  build      npm run build --silent  (cached, after ^build)
  lint       npm run lint --silent  (cached, after ^build)
  package    smith npm-pack apps/web  (after build)
  prune      smith npm-prune apps/web  (after build, ^build)
  start      npm run start --silent  (after build)
  test       npm run test --silent  (cached, after ^build)
  typecheck  npm run typecheck --silent  (cached, after ^build)

web:build — Compile into the workspace's dist/<project dir>
  command    npm run build --silent
...
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `smith explain needs a project name` | Give one; `smith projects` lists them. |
| `no project named "nope" (projects: api, tools, ui, web)` | Check the spelling. |
| `web was not made from a recipe: smith.toml records none for it` | `--recipe` only applies to recipe projects. |

## Related

[`smith why-miss`](why-miss.md), [`smith projects`](projects.md), [`smith recipe check`](recipe-check.md),
[Tasks, cache and affected](../reference/tasks-and-cache.md).
