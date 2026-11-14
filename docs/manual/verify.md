# smith verify

`smith verify` checks the workspace the way CI does: the dependency boundaries first, then
lint, typecheck, test and build of every project, and a check that no file git tracks was
changed. It is the verify phase of the generated pipelines, the same as
[`smith ci verify`](ci.md), so a green run here is the run CI makes.

## When to use it

- Before you push, to find what CI would find.
- When a run passed on your machine and failed in CI: repeat it with `--no-cache`. A clean
  CI agent has no cache, and your machine replays a result when its inputs did not change.

## What it prints at the end

A table of the projects with the tasks that ran and how long they took, then, for each project:

| Column | What it says |
|---|---|
| Tests | how many tests passed and failed, as the runner counted them (jest, vitest, pytest, dotnet, cargo; for `go test`, packages) |
| Lint | `passed`, or the number of issues the linter reported (eslint, golangci-lint, ruff) |
| Build | the size of what the build wrote |
| Coverage | the share of lines covered |

A second table gives the paths: the build output, the test results and the coverage report.

A number smith cannot read is left as `-`, never guessed. **smith never turns coverage or test
results on.** They show when a report is already in place: `coverage/coverage-summary.json`,
`lcov.info`, `coverage.xml` or a cobertura file, a Go coverage profile, and a JUnit or `.trx` file.
A report older than the run shows its age (`81.2% (2d old)`), so an old file is not taken for the
result of this run. When no project has coverage, a line says so. The job summary of the CI carries
the same tables.

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
