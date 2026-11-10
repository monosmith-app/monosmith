# smith verify

`smith verify` checks the workspace the way CI does: the dependency boundaries first, then
lint, typecheck, test and build of every project, and a check that no file git tracks was
changed. It is the verify phase of the generated pipelines, the same as
[`smith ci verify`](ci.md), so a green run here is the run CI makes.

## When to use it

- Before you push, to find what CI would find.
- When a run passed on your machine and failed in CI: repeat it with `--no-cache`. A clean
  CI agent has no cache, and your machine replays a result when its inputs did not change.

## Synopsis

```sh
smith verify [--no-cache]
```

## Arguments and flags

<!-- smith:begin verify -->
Check the workspace as CI does: boundaries, lint, typecheck, test and build

Usage: `smith verify [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--no-cache` | boolean | off | run every task even if a cached result exists, as a clean CI agent does |

<!-- smith:end verify -->

## See also

[`smith ci`](ci.md), [`smith run`](run.md), [Tasks and the cache](../reference/tasks-and-cache.md).
