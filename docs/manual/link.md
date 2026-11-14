# smith link

`smith link` makes the `smith` command start the program itself, with no launcher in front of it.

A package manager (npm, pnpm, pip, dotnet) starts smith through a launcher of its own language:
Node for npm, Python for pip. That launcher stays alive as long as smith runs. It costs about
110 ms at every start and about 40 MB of memory for every smith process, and on Windows it
keeps the installed `smith.exe` in use, so an update cannot replace it.

## What it does

`smith link` finds the smith entries of the installation that runs (`smith.cmd`, `smith.ps1`, a
shell script, a symbolic link, a small launcher program) and points each at the program. It keeps
the original of each, so `smith link --undo` gives them back. It touches only the entries of its
own installation, never another smith that is also on `PATH`.

On Windows the entries call a hard link of the program in `~/.smith/run/`. A hard link is a second
name for the same file, so it takes no disk space. Because smith runs from the link, the installed
file is free: `npm i -g monosmith@latest` works while smith runs in other terminals, and those keep
running the version they started with. On other systems the entries call the installed program,
which the system lets an update replace.

## The first run does it by itself

The first time a version runs behind its launcher, smith links once and says so at the end of the
command. It takes effect from the next command. The package manager writes its entries again when it
updates smith, so the first run of the next version links again.

It does not link in CI, without a terminal, for `serve`, `mcp`, `lsp`, `prompt`, `link`, `update` or
completion, or when `SMITH_NO_LINK=1` is set.

## Synopsis

```sh
smith link [--status] [--undo]
```

`--status` shows how smith starts now and what each entry does. `--undo` gives the entries back and
removes the links. Running `smith link` again changes nothing that is already direct.

## Arguments and flags

<!-- smith:begin link -->
Make the smith command start the program directly: faster to start, less memory

Usage: `smith link [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--status` | boolean | off | show how smith starts now and change nothing |
| `--undo` | boolean | off | give the entries back as the package manager wrote them |

<!-- smith:end link -->
