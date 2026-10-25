# ADR 0014: The convert end-to-end test runs on demand

- Status: accepted
- Date: 2026-10-17

## Context

Building the old-mnci and Nx workspaces that the test converts took 30 to 44 minutes on
Windows and held every pull request. Converting itself takes seconds, and nobody else
converts yet.

## Options considered

Running the test on every pull request (the previous behaviour) against running it by hand:
the second was chosen. Other options are not recorded.

## Decision

The `convert` job in `golden-path` runs only when started by hand. Unit tests and the other
end-to-end suites still run on every pull request.

## Consequences

Since 2026-10-17 the other end-to-end suites run once a day too
([ADR 0017](0017-end-to-end-suites-run-daily.md)). A change to convert reaches `main`
without this test unless someone starts it.

## Sources

- Issue [#235](https://github.com/monosmith-app/monosmith/issues/235)
- `.github/workflows/golden-path.yml`
