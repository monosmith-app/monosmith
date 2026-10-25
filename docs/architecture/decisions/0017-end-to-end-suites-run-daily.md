# ADR 0017: The end-to-end suites run once a day

- Status: accepted
- Date: 2026-10-17

## Context

With a plugin for every language, the golden-path suites took up to 45 minutes on Windows,
and every pull request waited for them. Several pull requests at once queued behind
GitHub's concurrency limits.

## Options considered

Running them on every pull request (the previous behaviour, see
[ADR 0014](0014-convert-end-to-end-test-on-demand.md)) against a daily schedule: the
second was chosen. Other options are not recorded.

## Decision

Pull requests and pushes run the `unit` job (vet and test of every module on three
operating systems, a few minutes). The golden-path suites run once a day at 03:00 UTC and
when started by hand.

## Consequences

A change can reach `main`, and a release, before the end-to-end suites have run on it; the
daily run catches it within a day, and a change the unit tests cannot judge gets a manual
run before merging.

## Sources

- `.github/workflows/ci.yml`, `.github/workflows/golden-path.yml`
- Rationale recorded in the first version of the decisions log, 2026-10-17; no issue number
  recorded
