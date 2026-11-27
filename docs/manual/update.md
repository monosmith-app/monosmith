# smith update

`smith update` updates the smith program itself, the way it was installed. It is not
`smith upgrade`: that one brings a workspace up to the smith you run.

## When to use it

- The interface or `smith --version` says a newer smith exists.
- You want the newest release now, without remembering the install command.

## What smith shows before it runs

`smith update` asks smith's releases for the newest smith on your channel (stable, or next when
you run a prerelease; GitHub, or the mirror in `SMITH_DOWNLOAD_URL`). A smith a launcher (npm,
PyPI or NuGet) started lives in `~/.smith/versions/<version>/`: `smith update` downloads the new version next to
it, checks it against the signed `checksums.txt`, and the launcher (or the [link](link.md)) runs it from then on. A
workspace that pins `[workspace] smith` keeps its version. For any other install it prints the
command that matches how smith was installed (the install script, or a package from before the
launchers: npm gets `npm i -g monosmith@latest`, and pip, uv, pipx and dotnet tool their upgrade
command, each of which installs the launcher). It then asks what to do. `--check` only looks and
changes nothing. Without a terminal it needs `--yes`. When smith cannot tell how it was
installed, it names the releases page instead of guessing.

## When the newer smith is installed already

A workspace that pins `[workspace] smith`, or a shell that sets `SMITH_VERSION`, makes the launcher
run that version. When the newer smith is in `~/.smith/versions` already, there is nothing to
download, and `smith update` changes nothing. It says what to change instead:

```text
smith 2.56.0 is installed, but this workspace runs 2.55.2: smith.toml pins it ([workspace] smith), or SMITH_VERSION names it. To use 2.56.0 here, set smith = "2.56.0" under [workspace] (or "latest").
```

The notice after a command says the same, and the interface offers no update for that version.

## In the interface

When smith knows a newer release, the interface offers it once per version in a dialog with the
command shown, an **Update now** button and a **Don't remind me about updates** box. Cancel keeps the
update one click away: a red "update available" in the bottom bar, which opens the same dialog. The
**Settings > Update** section shows the same facts, asks smith's releases on Enter, and has the same box
to bring the reminder back or switch it off. The box keeps only the reminder away: smith still
checks, and `smith update` still works. **Update now** closes the interface and runs
`smith update --yes`, so the running smith can be replaced.

## On Windows

An update through the launcher writes a new folder, so nothing has to be closed. For the other
installs, a running smith cannot be replaced. The update opens its own window, waits for `smith update`
to exit, and runs the command there. When smith starts directly (see [`smith link`](link.md), which
the first run of a version does by itself), smith runs from a link and does not hold the installed file,
so the update runs in place and other terminals need not be closed. Other smith programs of the same install (an editor's
`smith mcp`, `smith serve`) block the replacement, so `smith update` names them and stops until
they are closed.

## Synopsis

```sh
smith update [--check] [--yes]
```

## Arguments and flags

<!-- smith:begin update -->
Update smith itself to the newest release, the way it was installed

Usage: `smith update [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--check` | boolean | off | only say whether a newer smith exists and what would run; change nothing |
| `-y, --yes` | boolean | off | update without asking |

<!-- smith:end update -->
