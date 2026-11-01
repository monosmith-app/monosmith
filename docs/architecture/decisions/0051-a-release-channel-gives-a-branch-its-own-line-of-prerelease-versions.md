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

The channel is built in three steps. The first covers the configuration, the version computation
and the branch guard. The second publishes. The third, last in this record, is the pipeline and the
pull request preview.

Publishing sends the channel's name to the plugin as `PublishParams.tag` (`dist_tag`, default the
prerelease identifier, never `latest`). A plugin that publishes under it declares the new
`channels` capability. It is a capability that no method needs, because it changes how `publish`
behaves. The Go SDK declares it with the `ChannelPublisher` marker interface, and the Node plugin
passes the name to `npm publish --tag`. The generated protocol types of the TypeScript, Python and
C# SDKs carry the field, but those SDKs have no way yet for a plugin to declare it.

Before it publishes, `smith release` refuses a channel release when a project in the plan has a
plugin without `channels`, a deploy plugin, or binaries sent to package managers or registries of
their own. Each of these would move a stable pointer. The GitHub Release of any version with a
prerelease part is created as a prerelease.

The pipeline and the preview (step 3): the generated pipeline of every provider triggers on a
channel branch and runs the `release` step there, and only that step. Pack, upload, attest and the
sign-in steps stay on the release branch, because the checks above already refuse what they would
publish. A workspace with no channel gets the same files as before. The plan carries the channel's
name, and `smith release --dry-run`, the report and the pull request comment say it. A pull request
build checks out a detached merge commit, so a plan then uses the branch the pull request targets
(`GITHUB_BASE_REF` and its equivalents).

Alternative not chosen: a flag on `smith release` (`--prerelease beta`). It leaves the pipeline and
the guard without a statement of which branch releases what, and each run would repeat the choice.

## Consequences

- A long-lived branch can plan and release its own line, for Node packages.
- A project whose latest channel tag has a different major than the channel starts the counter
  again at `.1`. Changing `major` starts a new line on purpose.
- Promotion is a merge into the release branch: the next stable release there is computed from the
  stable tags as usual.
- Not done: the other plugins' registries (each needs its own spelling of a prerelease and a stable pointer), and
  a way to declare `channels` in the TypeScript, Python and C# SDKs.

## Sources

- Issue [#845](https://github.com/monosmith-app/monosmith/issues/845), epic
  [#819](https://github.com/monosmith-app/monosmith/issues/819)
- `apps/smith/release/release_channel_model.go`, `apps/smith/workspace/release_channels_validator.go`
