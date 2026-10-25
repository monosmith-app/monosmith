# ADR 0033: Doctor fixes carry steps and choices

- Status: accepted
- Date: 2026-10-24

## Context

User feedback: several doctor checks described how to fix a problem that smith could fix
itself.

## Options considered

Not recorded beyond the pull request.

## Decision

The plugin protocol's `Fix` gains `steps` (argv commands run from the workspace root, stopping
at the first failure), `choices` (a question whose options are fixes) and `applies_to` (the
projects). The change is backward compatible and `safe` keeps its meaning: only a safe fix
without choices is applied by `--yes`. The `doctorfix` engine folds the same fix on many
projects (`Group`), answers choices one at a time (`Resolve`), writes files and runs steps
(`Apply`), and drives the CLI flow (`Walk`). `smith doctor --fix` asks numbered questions in a
terminal; `--choose` and `--only` apply a fix without prompts, and `--json` carries `fixes`.
The terminal interface's fix wizard is built on the same engine
([#547](https://github.com/monosmith-app/monosmith/pull/547)).

## Consequences

Fixes were added for tracked generated files, projects without lint, imports a project does
not declare, packages the ESLint config imports, tools a script starts, and Node
dependencies not installed. The same engine drives the full-level conversion plan
([ADR 0034](0034-adopt-and-convert-depth-levels.md)).

## Sources

- Pull request [#546](https://github.com/monosmith-app/monosmith/pull/546)
- [`smith doctor`](../../manual/doctor.md), [Writing a plugin](../../guides/writing-a-plugin.md)
