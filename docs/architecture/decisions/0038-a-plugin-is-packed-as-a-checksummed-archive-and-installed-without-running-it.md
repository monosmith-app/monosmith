# ADR 0038: A plugin is packed as a checksummed archive and installed without running it

- Status: accepted
- Date: 2026-10-26

## Context

[Decision 0036](0036-external-plugins-are-trusted-and-get-a-minimal-environment.md) made an external
plugin start only when it is trusted, and in CI only when `smith.toml` pins its `sha256`. It left
the way a plugin reaches a machine unchanged: an author attached a zip to a release and every user
wrote the unpack step, with nothing that checked what arrived. This is the second step of
[#565](https://github.com/monosmith-app/monosmith/issues/565): one package format, a command that
makes it, and a command that installs it, checks it and records the pin. A registry stays out of
scope.

## Options considered

- **A tarball, or a custom container.** Rejected: `zip` is in Go's standard library, every
  operating system can open one, and GitHub release assets are routinely zips.
- **Only the checksums inside the archive.** Rejected as the sole defence: whoever changes a file
  changes the list too. The list protects against corruption and swaps inside one archive; the digest
  of the archive, obtained separately, is what ties it to the author's release.
- **Install by name, from a registry or an index.** Deferred here, and decided later by [ADR 0040](0040-plugins-are-found-by-name-in-a-marketplace-index.md). Reason for the deferral: it needs namespace ownership,
  moderation and signing infrastructure (#164).
- **Trust the plugin at install.** Rejected: fetching code is not the decision 0036 asks a person to
  make about running it. `--trust` makes it explicit in one step.
- **Run the plugin at install to read its name or check it.** Rejected: nothing of the plugin may run
  before it is trusted. The manifest is read as data, and the program is hashed, not started.
- **A software bill of materials in the archive.** Not done: smith's SBOM machinery describes a
  workspace's projects and their lockfiles, not an arbitrary plugin directory, so reusing it is not
  cheap. Follow-up.

## Decision

**Format.** A plugin archive is `<name>-<version>.smith-plugin.zip`. At its root: `plugin.json`, the
plugin's files, and `checksums.sha256`, the format of `sha256sum` (`<hex>  <path>`, sorted), listing
every other file. Beside it, `<archive>.sha256` holds the digest of the archive. The format lives in
the `pluginarchive` slice; `pluginpack` and `plugininstall` depend on it and never on each other, and
`pluginhost` hands both the manifest reader, so the manifest has one owner.

**Pack.** `smith plugin pack [dir] [--out dir]` packs every file of the directory except version
control folders, earlier archives, an old checksums file and what the directory's `.smithpluginignore`
lists (a gitignore-like list without negation; the file itself is not packed). Entries are sorted, timestamps are
fixed, modes are 0644 or 0755; the files the first element of the manifest's `command` names for any
platform (`{os}`, `{arch}` and `{exe}` as wildcards) are marked executable, so a package built on one
operating system installs runnable on another. A symbolic link or special file is refused. Packing
does not run anything. Entries are stored, not deflated, so the archive is the same bytes for the same
directory on any smith build: a deflate stream is not guaranteed stable across Go releases, and a
digest that anyone can reproduce from the source is worth more here than a smaller download.

**Install.** `smith plugin install <source>` takes a path (archive or plugin directory), an https URL,
or `github:owner/repo@tag[#asset]`. A tag is required. `http` is accepted for loopback only, and every
redirect must satisfy the same rule. Before anything is written, smith:

1. compares the archive's SHA-256 with `--sha256`, or with the release's `<archive>.sha256` asset
   when installing from GitHub without `--sha256`; a mismatch refuses the install, and so does an
   archive with no digest at all, before it is downloaded (a plugin directory is the person's own
   files and has no archive);
2. requires `checksums.sha256` to list exactly the archive's files, and each file to match it while
   it is written (a directory source is checked when it has one);
3. refuses an entry that is absolute, has `..`, a backslash, a colon, a control character, a
   trailing dot or space, a Windows device name, or differs from another only in case; refuses
   links and special files; writes 0644 or 0755 only, so no setuid bit survives;
4. bounds the download (512 MiB) and the extraction (5000 files, 512 MiB per file, 1 GiB total),
   counted while writing so an archive that lies about its sizes still stops;
5. reads `plugin.json` as data, validates the name (it becomes a directory name), and checks that
   the command's program exists for this platform, by hashing it.

The files go to a staging directory beside the destination and are moved into place at the end; an
installed plugin of the same name is replaced only with `--force` and is put back if the move fails.

**Where, and the pin.** Inside a workspace the plugin goes to `.smith/plugins/<name>`;
with `--pin`, `[plugins.external.<name>] sha256` is also written with the same text-preserving edit
`smith plugin trust --pin` uses, with the same value (the digest of the file the plugin runs), and
without it the command prints the digest to pin. Outside a workspace, or with `--global`, it goes
to `$SMITH_HOME/plugins`; a pin belongs to a workspace, so `--pin` there is an error. A plugin with
no program for this platform is refused wherever it would go.

**Trust.** Install records no trust decision unless `--trust` is given, which does what `smith
plugin trust` does for the workspace. On a person's machine the pin still does not make a plugin
trusted (decision 0036); in CI the pin is what lets the installed plugin start. The two digests are
different things: the archive's digest says "this is what the author published", the pin says "this is the
file the workspace runs".

## Consequences

An install proves that an archive is the one whose digest the person or the release published and
that it is intact; it does not prove who published it, and the digest taken from the same place as
the download adds nothing. Signatures and provenance attestations, a registry, an SBOM in the
archive and the canary checks of `smith plugin check` are not part of this step. Assets of private
GitHub repositories are not downloaded. Packing everything in the directory means an author must
keep it clean or list what to leave out in `.smithpluginignore`; the archives are larger than
compressed ones. A plugin that needs a runtime installed by another tool (Node.js, Python
and their dependencies) packs its script and launcher, not the dependencies, unless they are in the
directory.

## Sources

- Issue [#565](https://github.com/monosmith-app/monosmith/issues/565)
- [`smith plugin pack`](../../manual/plugin-pack.md), [`smith plugin install`](../../manual/plugin-install.md)
- [Packaging and distributing a plugin](../../plugins/creating/packaging-a-plugin.md)
- [Plugin security](../../plugins/using/plugin-security.md)
