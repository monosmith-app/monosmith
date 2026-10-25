# ADR 0015: npm audit blocks only what can be fixed

- Status: accepted
- Date: 2026-10-17

## Context

An advisory with no upstream fix keeps CI red until the upstream package is patched, and
nothing in the repository can change that.

## Options considered

Failing on every advisory of moderate severity or above (the strict alternative) is
implied by the context; other options are not recorded.

## Decision

`smith ci audit` fails on a production advisory of moderate severity or above only when a
plain `npm audit fix` would change something. An advisory with no upstream fix, one whose
only fix is a semver-major `--force`, and dev-only advisories are printed as notes. This
ports the old mnci's rule.

## Consequences

CI no longer stays red while an upstream package is unpatched. The answer to a real finding
is a targeted `overrides` entry. Stricter policy checks are opt-in through `[ci]
audit_policies` (node, go, python and rust policies were added on 2026-10-20 in pull
requests [#388](https://github.com/monosmith-app/monosmith/pull/388) and
[#389](https://github.com/monosmith-app/monosmith/pull/389)).

## Sources

- Issue [#246](https://github.com/monosmith-app/monosmith/issues/246)
- [CI pipelines](../../reference/ci.md)
