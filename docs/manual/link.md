# smith link

`smith link` makes the `smith` command start the program itself, with no launcher in front of it.

The registry packages (npm, pnpm, PyPI, NuGet) install a launcher of their own language
([ADR 0101](../architecture/decisions/0101-the-registry-packages-are-launchers-not-releases.md)):
Node for npm, Python for pip. It chooses the smith version of a run and starts it from
`~/.smith/versions/<version>/`. That launcher stays alive as long as smith runs. It costs about
110 ms at every start and about 40 MB of memory for every smith process.

## What it does

`smith link` finds the smith entries of the launcher packages on `PATH` and points each at a hard
link of the running smith in `~/.smith/run/` (`smith-<version>`, `.exe` on Windows). A hard link is
a second name for the same file, so it takes no disk space. It keeps the original of each entry, so
`smith link --undo` gives them back. It changes only entries that name the `monosmith` package or call
the run folder, never another program named smith.

| System | Entries that `smith link` changes |
|---|---|
| Linux, macOS | the npm symbolic link `<prefix>/bin/smith`, the pnpm entries, the pip entry script |
| Windows | `smith.cmd`, `smith.ps1` and the `smith` shell script in the npm prefix, the pnpm entries |

On Windows, the pip and dotnet entries are small programs, not scripts. They keep their launcher, and
`smith link` lists them as skipped.

## The pins still work

The program in the run folder chooses the version as the launcher does ([ADR
0111](../architecture/decisions/0111-smith-link-starts-the-go-binary-which-keeps-the-pins.md)):
`SMITH_VERSION`, then the `[workspace] smith` pin of the nearest `smith.toml`, then the newest
version in `~/.smith/versions`. When that is not its own version, it hands the run to that version.
On Linux and macOS it replaces itself with the other program. On Windows it starts the other program on
the same console and returns its exit code. A pinned version that is missing is downloaded and checked
against the signed `checksums.txt` first, with one line on stderr. Nothing else uses the network.

`smith update` installs the new version in `~/.smith/versions`, as under the launcher. The entries
are not changed again: the next run hands off to the newest version.

## The first run does it by itself

The first time a version runs behind its launcher, smith links once and says so at the end of the
command. It takes effect from the next command. A run that the link handed to another version does not
link again.

It does not link in CI, without a terminal, for `serve`, `mcp`, `lsp`, `prompt`, `link`, `update` or
completion, or when `SMITH_NO_LINK=1` is set.

In the interface, the first run that linked tells so in a dialog and offers **Restart now**, because
only a new start uses the link. **Later** keeps the interface as it is.

## Synopsis

```sh
smith link [--status] [--undo]
```

`--status` shows how smith starts now and what each entry does. `--undo` gives the entries back and
removes the links, so the launcher starts smith again. Running `smith link` again changes nothing that is already direct.

## Arguments and flags

<!-- smith:begin link -->
Make the smith command start the program directly: faster to start, less memory

Usage: `smith link [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--status` | boolean | off | show how smith starts now and change nothing |
| `--undo` | boolean | off | give the entries back as the package manager wrote them |

<!-- smith:end link -->
