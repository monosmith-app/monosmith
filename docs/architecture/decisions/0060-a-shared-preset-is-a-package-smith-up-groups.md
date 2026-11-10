# ADR 0060: A shared preset is a package that smith up groups

- Status: accepted
- Date: 2026-11-01

## Context

Idea E1 of the design review: a company publishes its taste as a versioned package. Workspaces
depend on it, and a bump spreads without meetings.

## Challenge

A preset needs no new file format: a package that exports a config is already a preset, and
registries, versions and lockfiles already carry it. A "preset" concept inside smith would be a
second packaging system next to npm, PyPI and NuGet. The one thing smith can add is to treat a bump
of such a package as one visible decision.

## Decision

`[dependencies] presets = [...]` names the packages that carry shared configuration. `smith up`
(and the picker) shows their updates first, as a group named `preset`, whatever their semver level;
`--json` carries `preset: true`. `smith up` already changes every project that declares a package
in one change, so the bump is one change with one plan, and `smith try up <name>` previews it.
The list is the only new state.

Not built: `smith preset new` (a preset is an ordinary package; `smith add package` makes it) and
a bot that opens the upgrade pull request in each repository (that is the CI of the team, running
`smith up --yes`). Both would be additions without a decision behind them. The guide
`docs/guides/shared-presets.md` writes the pattern down.

## Cost

One more setting. A package that is in `presets` but is not in any project shows nothing.

## Sources

- `apps/smith/dependencies/find_updates_algorithm.go`, `docs/guides/shared-presets.md`
