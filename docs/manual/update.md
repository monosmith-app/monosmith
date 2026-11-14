# smith update

`smith update` updates the smith program itself, the way it was installed. It is not
`smith upgrade`: that one brings a workspace up to the smith you run.

## When to use it

- The interface or `smith --version` says a newer smith exists.
- You want the newest release now, without remembering the install command.

## What smith shows before it runs

`smith update` asks the npm registry for the newest smith on your channel (stable, or next when
you run a prerelease) and prints the command that matches how smith was installed: npm, dotnet
tool, pip, uv, pipx or the install script. It then asks what to do. `--check` only looks and
changes nothing. Without a terminal it needs `--yes`. When smith cannot tell how it was
installed, it names the releases page instead of guessing.

## On Windows

A running smith cannot be replaced. The update opens its own window, waits for `smith update`
to exit, and runs the command there. Other smith programs of the same install (an editor's
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
