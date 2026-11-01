# ADR 0051: A release channel gives a branch its own line of prerelease versions

- Status: accepted
- Date: 2026-11-01

## Context

`smith release` releases from one branch (`[release] branch`) and only stable versions. A prerelease
such as `-rc.1` is understood when it is published, but the release plan never makes one. A team
that builds the next major version on a long-lived branch cannot publish it from there. It holds
the work back, or merges it into the release branch and releases it as part of the current line.
MonoSmith met this with its own `smith-v2` branch (issue
[#845](https://github.com/monosmith-app/monosmith/issues/845)).

## Decision

A `[[release.channels]]` entry names a branch, a prerelease identifier and a major version:

```toml
[[release.channels]]
branch = "smith-v2"
prerelease = "beta"
major = 2
```

On a channel branch the release plan counts the releases of the channel, not those of the release
branch:

- The first release of a project is `<major>.0.0-<prerelease>.1`, whatever stable versions exist.
- Each later release raises the counter (`2.0.0-beta.2`) and keeps the numbers before it. A
  breaking change inside a channel does not start a new major, because the major is the channel's.
- The channel's last release is the highest `<major>.x.y-<prerelease>.<n>` tag that the checked out
  branch contains (`git tag --merged HEAD`). Tags of other branches do not count. The commits that
  decide whether to release are those since that tag, or since the last stable tag when the
  channel has none.
- The stable plan ignores prerelease tags, as before. Prerelease versions compare by the semver
  rules, so `beta.9` sorts before `beta.10`.

The channel is part of this decision in three steps. This record and its first step cover the
configuration, the version computation and the branch guard. Until the publish step exists, a real
release from a channel branch is refused, so that a prerelease can never move `latest`. Planning
(`smith release --dry-run`, the report, the pull request preview) works. The publish step
(dist-tag, GitHub prerelease, no move of `latest`) and the pipeline step are decided in the
records that add them.

Alternative not chosen: a flag on `smith release` (`--prerelease beta`). It leaves the pipeline and
the guard without a statement of which branch releases what, and each run would repeat the choice.

## Consequences

- A long-lived branch can plan and, after the publish step, release its own line.
- A project whose latest channel tag has a different major than the channel starts the counter
  again at `.1`. Changing `major` starts a new line on purpose.
- Promotion is a merge into the release branch: the next stable release there is computed from the
  stable tags as usual.
- Not done in this step: publishing, the dist-tag, the GitHub prerelease flag, the generated
  pipeline and the pull request preview naming the channel.

## Sources

- Issue [#845](https://github.com/monosmith-app/monosmith/issues/845), epic
  [#819](https://github.com/monosmith-app/monosmith/issues/819)
- `apps/smith/release/release_channel_model.go`, `apps/smith/workspace/release_channels_validator.go`
