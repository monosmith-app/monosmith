# ADR 0104: A channel without a major releases the next version

- Status: accepted
- Date: 2026-11-20
- Amends: [ADR 0051](0051-a-release-channel-gives-a-branch-its-own-line-of-prerelease-versions.md)

## Context

A team deploys each commit of a `dev` branch to a test environment, then merges `dev` into the
release branch to release the same version to the next environments. A release channel (ADR 0051)
needed a `major` and numbered `<major>.0.0-<id>.N`, a line for the next major. With 6.1.0 released
and `major = 6`, smith refused the first channel release (`6.0.0-beta.1 is not above the released
6.1.0`). Smith could not give `6.2.0-beta.N`, and nothing checked that the release branch released
the version that `dev` tested.

## Decision

1. `major` is optional. A channel without it numbers from the next planned stable version: the
   commits since the last stable tag give the target (`6.2.0`), and the channel releases
   `6.2.0-beta.1`, `6.2.0-beta.2`. The counter rises while the target stays the same and a
   releasing commit follows the latest prerelease. A new target starts again at `.1`.
2. A channel with a major keeps the ADR 0051 behaviour.
3. `[release] promote` (`warn` by default, `strict`, `off`) compares a plan on the release branch
   with the highest prerelease that a channel without a major released and that HEAD contains.
   A project with no such prerelease, or with one of another version, gives a warning (`warn`) or
   stops the release in the preflight (`strict`).
4. The release branch builds the release again. It does not reuse the files the channel built:
   that needs a lookup in each CI system and gives little more than the version check.

## Consequences

- One `[[release.channels]]` entry without `major` gives the dev-to-production flow.
- A commit that goes straight to the release branch, or a release without a channel build, is named
  before it ships. Teams that do not use a channel this way set `promote = "off"` or keep `major`.
- The files the release branch deploys come from a new build of the same commits. They are not the
  same bytes that were tested.
