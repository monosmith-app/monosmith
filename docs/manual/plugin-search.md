<!-- lore-master
parent: plugin.md
-->
# smith plugin search

`smith plugin search` lists the plugins of the marketplace: the index of plugins that
[`smith plugin install <name>`](plugin-install.md) can fetch. It matches a term against each
plugin's name, description, roles, kinds and languages, and shows the newest version of every match.

## When to use it

- You look for a plugin for a language, a project kind or a role, and you do not know its name.
- You want to see which versions a name has before you install one.

## Synopsis

```sh
smith plugin search [term] [--json]
```

## Arguments and flags

<!-- smith:begin plugin search -->
Search the plugin marketplace

Usage: `smith plugin search [term] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `term` | no | words to find in the name, description, roles, kinds or languages; default every plugin |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print JSON |

<!-- smith:end plugin search -->

A term of several words matches a plugin that contains every word, in any order, ignoring case.
`--json` prints an array of the matching entries, each with its `name` (what `smith plugin install`
takes), its `newest` stable version, and its `versions`, `url` and `sha256`; an empty result is `[]`.

## Which marketplace

In order: the `SMITH_MARKETPLACE` environment variable, `marketplace` under `[plugins]` in `smith.toml`,
then the index of [monosmith-app/plugins](https://github.com/monosmith-app/plugins). The value is an
https URL, an `http` URL on this machine, or a file path. Searching reads the index only; it downloads no
plugin.

While the index is in a private repository, smith asks GitHub for it again with a token when GitHub
answers 401, 403 or 404. It uses `GITHUB_TOKEN`, then `GH_TOKEN`, then `gh auth token` if the GitHub CLI
is installed. The token goes to GitHub hosts only. Without a token, smith tells you what to set.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The index was read, with or without matches. |
| 1 | The index cannot be read (network, private without a token, a file that is not a valid index of this schema). |

## Files and environment

Reads `smith.toml` (`[plugins] marketplace`) and the index. Uses `SMITH_MARKETPLACE`, `GITHUB_TOKEN`,
`GH_TOKEN` and the `gh` command. Writes nothing.

## Examples

```sh
smith plugin search deploy
```

```text
PLUGIN             VERSION   ROLES                  DESCRIPTION
moon-deploy        0.2.0     deploy                 Deploys to the moon

Install one with: smith plugin install <name>[@version]
```

(The plugin in this example is made up: the marketplace may list none yet.)

## Common errors

| Message | Cause and remedy |
|---|---|
| `the marketplace at ... answered 404: it is private, or it does not exist` | Set `GITHUB_TOKEN`, run `gh auth login`, or point `SMITH_MARKETPLACE` at another index. |
| `the marketplace index has schema N; this smith reads schema 1` | The index is newer than this smith. Update smith. |
| `cannot read the marketplace index` | The file path in `SMITH_MARKETPLACE` or `smith.toml` does not exist. |

## Related

[`smith plugin install`](plugin-install.md), [Using the plugin marketplace](../plugins/using/marketplace.md),
[Decision 0040](../architecture/decisions/0040-plugins-are-found-by-name-in-a-marketplace-index.md).
