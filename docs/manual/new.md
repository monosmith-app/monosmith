# smith new

`smith new` creates a workspace: it writes `smith.toml`, the CI pipelines and a few
repository files, runs `git init` on `main` and makes the first commit. It is the
start for an empty folder. A repository that already holds code is brought in with
[`smith adopt`](adopt.md) instead.

## When to use it

- You are starting a monorepo from nothing, or want an empty workspace to add
  projects to with [`smith add`](add.md).
- You want the workspace created in a folder that exists but holds no `smith.toml`
  (`--dir .`).

Do not use it to put smith on a repository that already has packages, modules or
an Nx or Turborepo setup; `smith adopt` and [`smith convert`](convert.md) read what is
there and keep it.

## Synopsis

```sh
smith new <name> [--dir <dir>] [--scope <scope>] [--module-prefix <prefix>] [--ci <providers>] [--recipe <recipe>]
```

## Arguments and flags

<!-- smith:begin new -->
Create a monorepo: configuration, git, CI pipelines and a first commit

Usage: `smith new <name> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `name` | yes | the workspace name, in kebab-case |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--ci <ci>` | string, repeatable: `github`, `azure`, `gitlab` | `github` | pipelines to generate: github, azure (default github) |
| `--dir <dir>` | string | `./<name>` | directory to create the workspace in (default ./<name>) |
| `--module-prefix <module-prefix>` | string | from the `origin` remote, else `example.com/<name>` | Go module prefix (default from the origin remote, else example.com/<name>) |
| `--recipe <recipe>` | string |  | the recipe smith add creates projects from: a path relative to smith.toml or an https URL (docs/reference/recipes.md) |
| `--scope <scope>` | string | `@<name>` | npm scope for packages (default @<name>) |

<!-- smith:end new -->

Notes on the flags:

- `<name>` must be kebab-case: lowercase letters and digits, words joined by single
  hyphens, starting with a letter. It becomes `[workspace] name` and the default
  npm scope.
- `--ci` takes a comma-separated list or the flag repeated. The help text lists
  `github` and `azure`; `gitlab` is also accepted (the schema of `smith.toml` lists
  all three under `ci.providers`).
- `--module-prefix` is read from the `origin` remote of the folder when it has one.
- `--recipe` writes `recipe = ...` under `[workspace]`. See
  [Recipes](../reference/recipes.md). It is not checked at creation time.
- `--dir .` creates the workspace in the current folder. Files smith does not own
  are kept.

In a terminal (standard input and output are terminals and `CI` is unset), a missing
`<name>` is asked for: `Name of the workspace:`, up to three tries. Elsewhere it is the
error listed below.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The workspace was created and committed. |
| 1 | The name is invalid, the folder already holds a workspace, or a step failed. |

## Files and environment

Writes, in the new folder: `.gitignore`, `.gitattributes`, `smith.toml`, `README.md`,
`.github/workflows/ci.yml` (and `azure-pipelines.yml` or `.gitlab-ci.yml` for the other
providers), `.devcontainer/devcontainer.json`. It then runs `git init -b main`, stages
what is in the folder and commits it as `chore: create the <name> workspace with smith`.

`smith.toml` starts with a `#:schema` line pointing at the published JSON Schema, so
editors that read TOML schemas validate it (see [`smith schema`](schema.md)).

Reads no smith-specific environment variable. It does read the `origin` remote of an
existing repository for the Go module prefix.

## Examples

A workspace with both CI providers:

```sh
smith new demo2 --ci github,azure --scope @demo --module-prefix github.com/demo/demo2
```

```text
  wrote .gitignore
  wrote .gitattributes
  wrote smith.toml
  wrote README.md
  wrote .github/workflows/ci.yml
  wrote azure-pipelines.yml
  wrote .devcontainer/devcontainer.json
  committed

Created demo2. Next:
  cd demo2
  smith add app web --lang node
```

In a folder that already has a file in it, with `--dir .`:

```sh
smith new notes --dir .
```

```text
  wrote .gitattributes
  wrote smith.toml
  wrote README.md
  wrote .gitignore
  wrote .github/workflows/ci.yml
  wrote .devcontainer/devcontainer.json
  committed

Created notes. Next:
  cd .
  smith add app web --lang node
```

When the folder was not a git repository, `git init` ran and the first commit held
every file in it, `notes.txt` included (checked with `git show --stat`). In a folder
that is already a repository, this was not exercised for this page.

## Common errors

| Message | Cause and remedy |
|---|---|
| `smith new needs a name for the workspace` | No `<name>`. Run `smith new acme`. To bring in a repository you already have, run `smith adopt --dry-run`. |
| `workspace name "bad/name" must be kebab-case: ...` | Use lowercase words joined by hyphens, such as `web-app`. |
| `<dir> already holds a smith workspace` | The folder has a `smith.toml`. Pick another `--dir`, or use [`smith upgrade`](upgrade.md) to refresh it. |

## Related

[`smith add`](add.md), [`smith adopt`](adopt.md), [`smith config`](config.md),
[Getting started](../getting-started/first-workspace.md), [CI pipelines](../reference/ci.md).
