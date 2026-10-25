<!-- lore-master
parent: plugin.md
-->
# smith plugin install

`smith plugin install` puts a packaged plugin on your machine: from an archive made by
[`smith plugin pack`](plugin-pack.md), from a plugin directory, from an https URL, or from a GitHub
release. It checks the archive before it writes anything, **never starts the plugin**, and, inside a
workspace, pins the digest of the file the plugin runs in `smith.toml`, which is what CI needs.

## When to use it

- A plugin author gave you a `.smith-plugin.zip`, a URL or a GitHub repository and tag.
- A workspace needs a plugin in CI: install it on your machine, commit the pin in `smith.toml`.

## Synopsis

```sh
smith plugin install <source> [--sha256 <digest>] [--global] [--force] [--trust]
```

## Arguments and flags

<!-- smith:begin plugin install -->
Install a packaged plugin and pin its digest

Usage: `smith plugin install <source> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `source` | yes | an archive or plugin directory path, an https URL, or `github:owner/repo@tag` (optionally `#<asset>`) |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--force` | boolean | off | replace an installed plugin of the same name |
| `--global` | boolean | off | install into $SMITH_HOME/plugins even inside a workspace, and pin nothing |
| `--sha256 <sha256>` | string |  | the SHA-256 the archive must have; the install is refused if it differs |
| `--trust` | boolean | off | also record that you trust the plugin in this workspace |

<!-- smith:end plugin install -->

## Where it comes from

| Source | Notes |
|---|---|
| a path to a `.smith-plugin.zip` | the archive must have a `checksums.sha256` |
| a path to a directory | a quick way to try a plugin you are writing; if it has a `checksums.sha256` it is checked, otherwise the files are installed as they are |
| `https://...` | the file is downloaded to a temporary file, at most 512 MiB, with at most five redirects; every redirect must be https too. `http://` is accepted for `localhost` only |
| `github:owner/repo@tag` | the release `tag` of the repository: its one asset ending in `.smith-plugin.zip`, or `github:owner/repo@tag#<asset>` when it has several. A tag is required. `GITHUB_TOKEN` and `GITHUB_API_URL` are honoured for the lookup; assets of private repositories are not supported |

## What smith checks before it installs

1. **The archive's digest.** `--sha256 <digest>` must equal the archive's SHA-256; for a GitHub
   release without `--sha256`, the release's `<archive>.sha256` asset (which `smith plugin pack`
   writes) is used when it has one. A difference stops the install. Without either, smith prints the
   archive's digest and says it was not compared with anything: a download is then only as
   trustworthy as where it came from.
2. **Every file against `checksums.sha256`.** The list must name exactly the files in the archive,
   and each file's SHA-256 must match. This catches corruption and a file swapped inside the
   archive; it cannot tell you who made the archive, which is what step 1 is for.
3. **Paths and file types.** An entry may not be absolute, contain `..`, a backslash, a drive letter
   or a Windows device name, or differ from another only in case. Symbolic links and special files
   are refused. Files are written with mode 0644, or 0755 for an executable: no setuid bit survives.
4. **Sizes.** At most 5000 files, 512 MiB per file and 1 GiB in all, measured while writing, so an
   archive that lies about its sizes still stops.
5. **The plugin.** `plugin.json` must be at the archive root, with a lowercase `name`, and its
   `command` must name a program that exists for this platform.

The files are written to a staging directory beside the destination and moved into place only when
every check has passed. A failure installs nothing and leaves an installed plugin of the same name as
it was.

## Where it goes

- Inside a workspace: `.smith/plugins/<name>`, and `[plugins.external.<name>] sha256 = "..."` is
  written to `smith.toml` with the same text-preserving edit `smith plugin trust --pin` uses. The
  digest is that of the file the plugin runs (the script, when the command runs an interpreter on
  one), not of the archive.
- Outside a workspace, or with `--global`: `$SMITH_HOME/plugins/<name>` (`~/.smith/plugins`), and
  nothing is pinned, because a pin belongs to a workspace.

An installed plugin of the same name is replaced only with `--force`.

## Installing is not trusting

Installing never starts the plugin and does not record that you trust it. The first command that
needs it asks you in a terminal, or you run `smith plugin trust <name>`; with `--trust` the install
records that decision for the workspace too. On a person's machine a pin does not make a plugin
trusted (it arrived with the repository); in CI the pin is what lets the plugin start. A trusted plugin runs
with your privileges and smith does not sandbox it.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The plugin is installed. |
| 1 | A digest or checksum did not match, an entry was unsafe, a limit was exceeded, the download failed, the plugin has no program for this platform, or it is already installed and `--force` was not given. Nothing was installed. |

## Files and environment

Reads the source; writes under `.smith/plugins` or `$SMITH_HOME/plugins`, and `[plugins.external.<name>]`
in `smith.toml`; with `--trust`, `trust.json` in the smith home. Uses `GITHUB_TOKEN` and
`GITHUB_API_URL` for the release lookup only; the token is never sent to a download.

## Examples

```sh
smith plugin install ./widget-0.1.0.smith-plugin.zip --sha256 db7fb4c0...2204d
```

```text
Installed widget 0.1.0 in ~/work/inst/.smith/plugins/widget (2 files, from ./widget-0.1.0.smith-plugin.zip).
  archive sha256 db7fb4c0...2204d matches the digest you gave or the release published.
  every file matches checksums.sha256.
  the plugin was not started.
Pinned sha256 9e36a074...e7087 in smith.toml [plugins.external.widget], which is what CI checks.
smith will ask before it first starts widget (or run `smith plugin trust widget`).
```

From a release, trusting it for this workspace in the same step:

```sh
smith plugin install github:acme/smith-plugin-widget@v0.1.0 --trust
```

A digest that does not match:

```text
✖ the archive's sha256 is db7fb4c0...2204d, but --sha256 says abc
  → nothing was installed: download the archive again, or check where the digest came from
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `✖ the archive's sha256 is X, but --sha256 says Y` (or `the release's ...sha256 says Y`) | The archive is not the one that was vouched for. Download it again; if it persists, do not install it. |
| `the archive has no checksums.sha256, so nothing can be verified` | Not made by `smith plugin pack`. Ask the author for a packed archive. |
| `the archive does not match its checksums.sha256: ...` | A file was added, removed or changed after packing. |
| `unsafe entry name "..."` | The archive tries to write outside the plugin's folder. Do not use it. |
| `✖ plugin already installed: <dir>; use --force to replace it` | A plugin of that name is there; `--force` replaces it. |
| `the plugin has no program for this platform` | The archive was packed without a binary for your operating system and architecture. |
| `release v1 has several plugin archives (...)` | Name one: `github:owner/repo@v1#<asset>`. |
| `only https URLs are accepted (http only for localhost)` | Use an https link. |

## Related

[`smith plugin pack`](plugin-pack.md), [`smith plugin trust`](plugin-trust.md),
[`smith plugin list`](plugin-list.md), [Plugin security](../plugins/using/plugin-security.md),
[Installing, enabling and configuring plugins](../plugins/using/managing-plugins.md),
[Decision 0038](../architecture/decisions/0038-a-plugin-is-packed-as-a-checksummed-archive-and-installed-without-running-it.md).
