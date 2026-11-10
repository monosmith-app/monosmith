# Using the plugin marketplace

The marketplace is a list of plugins that smith can install by name. It is a JSON file
(`index.json`) in the repository [monosmith-app/plugins](https://github.com/monosmith-app/plugins).
Each entry gives a plugin's description, what it does and, for every version, the URL of its archive
and the archive's SHA-256. The marketplace holds no code. The reasons are in
[decision 0040](../../architecture/decisions/0040-plugins-are-found-by-name-in-a-marketplace-index.md).

## Find and install a plugin

```sh
smith plugin search               # every plugin
smith plugin search deploy go     # plugins that mention both words
smith plugin install widget       # the newest version that is not a pre-release
smith plugin install widget@1.2.0 # one version
```

`install` downloads the archive, checks it against the digest in the index, and installs it by the same
rules as any other archive ([`smith plugin install`](../../manual/plugin-install.md)): it checks the
`checksums.sha256` inside, refuses unsafe paths and large archives, and never starts the plugin.

Installing is not trusting. Run `smith plugin trust <name>` after you read the plugin, or install with
`--trust`. For CI, install with `--pin` and commit the pin it writes to `smith.toml`.
A listed plugin is not a reviewed plugin: the digest says the archive is the one that was listed, not
that the code is safe. See [Plugin security](plugin-security.md).

A file or directory in the current directory with the same name as a plugin is a path, not a name.
Write `./widget` to be clear.

## Use another marketplace

smith looks for the index in this order:

1. the `SMITH_MARKETPLACE` environment variable;
2. `marketplace` under `[plugins]` in `smith.toml`;
3. the index of monosmith-app/plugins.

```toml
[plugins]
marketplace = "https://example.com/plugins/index.json"
```

The value is an https URL, an `http` URL on this machine, or a file path. A company can host its own
list that way, or a team can test an index from a local file. Read the line smith prints before the
install (`Found widget@1.2.0 in the marketplace ...`): a `smith.toml` from a repository you cloned can
name any index.

## A private index

While an index is in a private GitHub repository, smith asks GitHub again with a token when the first
request is refused. The token comes from `GITHUB_TOKEN`, `GH_TOKEN` or `gh auth token`. It goes to GitHub
only. Without a token, smith says that the index is private or missing and what to set. The archives the
index lists must be public: smith does not send a token to a download.

## Publish a plugin

Write the plugin, pack it with [`smith plugin pack`](../../manual/plugin-pack.md), publish the archive at
a stable https URL, and open a pull request on monosmith-app/plugins that adds the entry. The steps
and the checks are in that repository's `CONTRIBUTING.md`. The packing format is in
[Packaging and distributing a plugin](../creating/packaging-a-plugin.md).
