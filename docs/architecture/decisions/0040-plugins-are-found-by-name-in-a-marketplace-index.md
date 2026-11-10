# ADR 0040: Plugins are found by name in a marketplace index

- Status: accepted
- Date: 2026-10-31

## Context

[Decision 0038](0038-a-plugin-is-packed-as-a-checksummed-archive-and-installed-without-running-it.md)
made a packed plugin installable from a path, a URL or a GitHub release, and deferred "install by name,
from a registry or an index". A person still has to know where a plugin lives and where its digest is.
Issue [#826](https://github.com/monosmith-app/monosmith/issues/826) asks for a repository that lists
installable plugins (`monosmith-app/plugins`, private for now), and for `smith plugin search` and
`smith plugin install <name>` to resolve names through it. It is part of the smith 2.0 epic
([#819](https://github.com/monosmith-app/monosmith/issues/819)).

## Options considered

- **A registry service with accounts, namespaces and moderation.** Rejected for now: it is the
  infrastructure 0038 deferred, and a Git repository with pull requests gives review and history for free.
- **One folder or file per plugin, merged by a build.** Rejected: more moving parts for a list that is
  small. One `index.json`, sorted by name, is easy to review and to mirror.
- **The index holds the plugin's files.** Rejected: the index only points to archives made by
  `smith plugin pack`. The format that smith checks does not change.
- **Install by name trusts the plugin, because the index is curated.** Rejected: listing is not review.
  Decision 0036 still decides whether a plugin starts.
- **Send a GitHub token to whatever host the index names.** Rejected: a token goes to GitHub hosts only.

## Decision

**The index.** `index.json` has `schema: 1` and `plugins`, an object keyed by plugin name. Each plugin has
`description`, `license`, optional `roles`, `kinds`, `languages` and `homepage`, and `versions`: an object
keyed by version, each with the archive `url` (https; http only for this machine) and its lowercase hex
`sha256`. The file `index.schema.json` in the repository describes it, and its workflow checks the schema,
the order of names, that each archive is named `<name>-<version>.smith-plugin.zip`, and (with network)
that each archive has the digest the index states. smith itself validates what it reads (the `pluginmarket`
slice): the schema number, safe names and versions, a digest and a URL for every version.

**Where.** `SMITH_MARKETPLACE`, else `[plugins] marketplace` in `smith.toml`, else the raw
`index.json` of `monosmith-app/plugins` on its default branch. The value is an https URL, a loopback http
URL or a file path.

**Search.** `smith plugin search [term]` lists the plugins whose name, description, roles, kinds or
languages contain every word of the term, with the newest version.

**Install.** An install argument that is a valid plugin name, optionally `@version`, and is not an existing
path is a marketplace name. smith picks the version (the newest that is not a pre-release, or the one
asked for), and hands plugininstall the archive URL and the index's digest as if the person had given
`--sha256`. So the whole check of 0038 runs: archive digest, `checksums.sha256`, safe paths, limits, manifest.
`--sha256` with a name is an error, because the index is the source of the digest. Nothing is trusted:
`--pin` and `--trust` keep their meaning.

**Private index.** The first request carries no credentials. If a GitHub host answers 401, 403 or 404,
smith asks again with a token from `GITHUB_TOKEN`, `GH_TOKEN` or `gh auth token`. With none, it says
what to set. The token is never sent to another host, and never to an archive download.

## Consequences

The digest in the index ties the archive to the entry that was reviewed, not to a person. Whoever controls
the index controls what a name means, so a `marketplace` in a `smith.toml` that arrives with a repository can
point `smith plugin install <name>` at another index. The install still needs its own digest to match and
the plugin still needs a trust decision, but a person should read the location the command prints
(`Found widget@1.2.0 in the marketplace ...`). Archives in private repositories cannot be downloaded yet:
the index can be private, the archives it lists must be reachable without a token. Signatures, yanked
versions, dependency between plugins, and update checks (`smith plugin outdated`) are not part of this step.

## Sources

- Issue [#826](https://github.com/monosmith-app/monosmith/issues/826), epic
  [#819](https://github.com/monosmith-app/monosmith/issues/819)
- [`smith plugin search`](../../manual/plugin-search.md), [`smith plugin install`](../../manual/plugin-install.md)
- [Using the plugin marketplace](../../plugins/using/marketplace.md)
- Repository `monosmith-app/plugins` (`index.schema.json`, `CONTRIBUTING.md`)
