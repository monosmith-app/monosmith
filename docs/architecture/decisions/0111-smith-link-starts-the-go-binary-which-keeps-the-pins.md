# ADR 0111: smith link starts the Go binary, which keeps the pins

- Status: accepted
- Date: 2026-11-27
- Amends: [ADR 0101](0101-the-registry-packages-are-launchers-not-releases.md) (a linked entry chooses the version in Go)

## Context

Since ADR 0101 the npm package `monosmith` is a Node launcher. It chooses the version of a run
(`SMITH_VERSION`, then the `[workspace] smith` pin, then the newest installed), downloads it when it
is missing, and starts `~/.smith/versions/<version>/smith`. Node stays in front of every run: about
110 ms more at each start and about 40 MB more memory.

`smith link` and the first-run link were made to remove that cost, but they looked for the old npm
layout (the program inside `node_modules/monosmith`). Under the launcher they found no entry and did
nothing. A plain link to one version would also break the pins: every workspace would run that version.

Owner decision of 2026-11-27: "Go launcher, keep pins".

## Decision

1. **The link target.** `smith link` and the first-run link point the smith entries of the launcher
   packages at a hard link of the running version in the run folder: `~/.smith/run/smith-<version>`
   (`.exe` on Windows), on every system. A copy replaces the hard link where the system cannot link.
2. **The entries.** For a program in `~/.smith/versions/`, smith looks in the folders of `PATH` for
   entries that name the `monosmith` package or call the run folder. That finds the npm entries (the
   `smith` symbolic link in `<prefix>/bin` on Unix; `smith.cmd`, `smith.ps1` and the `smith` shell
   script in the npm prefix on Windows), the pnpm entries and the pip entry script on Unix. Another
   program named smith is never changed, and a program started by its own path never takes the smith
   command over. A small launcher program (pip or dotnet on Windows) keeps its launcher: a copy of
   smith in its place would not run from the run folder. The backups, `link.json`, `--status` and
   `--undo` stay as they were.
3. **The hand-off.** At the start of every run, before the command is read, a program that runs from
   the run folder, with no `SMITH_LAUNCHER` set, chooses the version as the Node launcher does: the same
   version pattern, the same error for a bad `SMITH_VERSION`, the nearest `smith.toml`, and the newest
   installed version without a pin. When that is not the running version it hands the run to
   `~/.smith/versions/<version>/smith`: `exec` on Unix, and on Windows a child on the same console
   whose exit code it returns. Only a pinned version that is missing uses the network: it is downloaded
   and checked against the signed `checksums.txt`, with one line on stderr.
4. **No loop, and the update.** The next program gets `SMITH_LAUNCHER=link@<version>`. It does not
   hand off again and does not link again. It runs from the versions folder, so `smith update`
   installs the new version there, as under a launcher. The link target is never rewritten: the next
   run hands off to the newest installed version.
5. **The launchers stay.** The Node, Python and .NET launchers are unchanged. The link only removes
   them from the runs of the smith command.

## Consequences

- A linked run starts with no Node in front. A run whose version is not the linked one starts two Go
  programs (one `exec` on Unix).
- A workspace pin keeps working after `smith link`, and `SMITH_VERSION` too.
- When npm writes its entries again (a new launcher), the first run of the next smith version links
  them again. `SMITH_NO_LINK=1` stops the first-run link.
- On Windows, the pip and dotnet launcher programs still start their runtime first.
