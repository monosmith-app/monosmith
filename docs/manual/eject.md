# smith eject

`smith eject` copies a default configuration that smith passes to a tool into the repository, so
the repository owns it. smith runs lint, type checks and tests with defaults it ships (written to
`.smith/defaults/`), which means a repository needs no linter, type checker or test configuration.
When you want different rules, eject the default and edit the copy: the file lands where the tool
reads it, replaces the default whole, and smith stops passing a configuration of its own.

## When to use it

- You want to change a lint rule, the type checker's strictness or the test options, and the
  default does not allow it.
- You want the configuration visible in the repository, for people and editors that read it.

Without arguments it only lists what could be ejected, so it is safe to run to see what smith owns.

## Synopsis

```sh
smith eject [tool...] [--all] [--force]
```

## Arguments and flags

<!-- smith:begin eject -->
Copy smith's default configuration of a tool into the repository, to own and edit it

Usage: `smith eject [tool...] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `tool...` | no |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--all` | boolean | off | eject every default that applies |
| `--force` | boolean | off | replace a file that exists already |

<!-- smith:end eject -->

The tools with a default, the file the default is copied to (in the workspace root) and the
language that runs them:

| Tool | Copied to | Language |
|---|---|---|
| `golangci-lint` | `.golangci.yml` | Go |
| `ruff` | `ruff.toml` | Python |
| `mypy` | `mypy.ini` | Python |
| `pytest` | `pytest.ini` | Python |
| `rustfmt` | `rustfmt.toml` | Rust |
| `phpstan` | `phpstan.neon`, in each package that uses the default (PHPStan reads only the package's own directory) | PHP |

- Only defaults that apply to the workspace now can be ejected: a tool the repository already
  configures (in the project or above it), or one no project uses, has none. `smith eject` lists
  the ones that do.
- A file at the target path is left alone: if it differs from the default the command refuses and
  names `--force`; if it is identical there is nothing to do.
- The copy goes to the workspace root, so it applies to every project below it, the way a
  root-level `ruff.toml` or `.golangci.yml` does. Move it into a project for that project alone.
- smith deletes and rewrites `.smith/defaults/` while it runs, so what is listed is what applies
  now, not what applied the last time.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The defaults were copied, or the list was printed. |
| 1 | Not in a workspace, an unknown tool, a tool with no default to copy, or a differing file in the way without `--force`. |

## Files and environment

Reads `.smith/defaults/` and writes the target file in the workspace root. It rewrites
`.smith/defaults/` (git-ignored) as every command that discovers projects does. No network request
is made.

## Examples

See what smith owns:

```sh
smith eject
```

```text
Defaults smith passes to tools in this workspace (smith eject <tool> copies one to the repository):
  ruff           python  -> ruff.toml
  mypy           python  -> mypy.ini
  pytest         python  -> pytest.ini
```

Take the type checker's configuration:

```sh
smith eject mypy
```

```text
Copied the default of mypy to mypy.ini: it is yours now, and smith passes no configuration to mypy.
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `✖ smith has no default for <name>` | The name is not one of the tools in the table above. |
| `✖ smith has no default for <tool> to copy: the repository already configures it, or no project uses it` | Nothing to eject: the file is yours already, or no project runs the tool. |
| `✖ <file> exists already and differs from smith's default` | Edit the file, or pass `--force` to replace it with the default. |

## Related

[`smith lint`](lint.md), [`smith typecheck`](typecheck.md), [`smith test`](test.md),
[Python](../plugins/using/catalogue/python.md#tool-configuration) and
[Go](../plugins/using/catalogue/go.md#lint-configuration) configuration,
[default tool configuration](../reference/tasks-and-cache.md#default-tool-configuration).
