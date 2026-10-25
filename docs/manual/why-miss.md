# smith why-miss

`smith why-miss <project>:<task>` says why a task would run again instead of coming from the
cache (a cache miss). It compares the task's current cache key inputs with those of its last
cached run and lists what changed: a file added, changed or removed, an environment variable,
the command, the plugin version.

## When to use it

- A task you expected to be a cache hit ran again, in CI or locally.
- You want to know what a change touches before running.

For the task's settings (what its command is, where a field came from) use
[`smith explain`](explain.md).

## Synopsis

```sh
smith why-miss <project>:<task> [--json]
```

## Arguments and flags

<!-- smith:begin why-miss -->
Say why a task would run again instead of coming from the cache (a cache miss; explain shows its settings)

Usage: `smith why-miss <project>:<task> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `project:task` | yes | the task to explain, such as `web:build` |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print JSON |

<!-- smith:end why-miss -->

The target is one word, `<project>:<task>`. (`smith commands --json` describes it as two
arguments; the command itself rejects `web test` with `smith why-miss takes a project and
task as <project>:<task> and nothing else`.)

## Exit codes

| Code | Meaning |
|---|---|
| 0 | An answer was printed, whether the task would hit or miss. |
| 1 | No such project or task, or a malformed target. |

## Files and environment

Reads the task cache (`.smith/cache` or `cache.dir`) and the current files. Writes nothing.

## Examples

Nothing changed:

```sh
smith why-miss web:test
```

```text
web:test would come from the cache: nothing changed since its last cached run.
```

After editing `apps/web/src/main.ts`:

```sh
smith why-miss web:build
```

```text
web:build would run: 1 change(s) since its last cached run (2026-10-25 09:07)
  file       apps/web/src/main.ts changed
```

When a file was added:

```text
ui:test would run: 1 change(s) since its last cached run (2026-10-25 09:07)
  file       libs/ui/src/failing.spec.ts added
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `smith why-miss takes a project and task as <project>:<task> and nothing else; got "test" too` | Write `web:build`. |
| `no project named "nope"` | `smith projects`. |
| `web has no nope task` | `smith explain web` lists the tasks. |

## Related

[`smith explain`](explain.md), [`smith run`](run.md) (`--no-cache`),
[Tasks, cache and affected](../reference/tasks-and-cache.md).
