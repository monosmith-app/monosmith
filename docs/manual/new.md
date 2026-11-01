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
smith new <name> [--dir <dir>] [--scope <scope>] [--module-prefix <prefix>] [--ci <providers>]
                 [--language <language>] [--default-stack node=<stack>] [--architecture <layout>]
                 [--release-branch <branch>] [--release-remote <remote>] [--github-releases]
                 [--cache-dir <dir>] [--cache-max-mb <mb>] [--save-recipe <name>]
                 [--recipe <recipe> | --recipe-file <file>] [--yes]
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
| `--agent` | boolean | on | add the MonoSmith agent: .claude/agents/smith.md, .mcp.json and .vscode/mcp.json (default) |
| `--architecture <architecture>` | string | `vertical-slices` | the code layout smith checks: vertical-slices, hexagonal, clean, ddd, modular-monolith, mvc, package-by-layer, none (default vertical-slices) |
| `--cache-dir <cache-dir>` | string | `.smith/cache` | the task cache folder inside the workspace (default .smith/cache) |
| `--cache-max-mb <cache-max-mb>` | string | `2048` | the largest the task cache may grow, in MB; 0 for no limit (default 2048) |
| `--ci <ci>` | string, repeatable: `github`, `azure`, `gitlab` | `github` | pipelines to generate: github, azure, gitlab, circleci, bitbucket (default github) |
| `--default-stack <default-stack>` | string, repeatable | `node=node-jest` | the stack Node projects get, as node=<stack>: node-jest (default), node-vitest or node-ts |
| `--dir <dir>` | string | `./<name>` | directory to create the workspace in (default ./<name>) |
| `--github-releases` | boolean | off | create a GitHub Release for every release tag |
| `--here` | boolean | off | create the workspace in the current folder, which must be empty; the name defaults to the folder's |
| `--language <language>` | string | `later` | the main language: node, python, go, jvm, dotnet, rust, or later (default later) |
| `--module-prefix <module-prefix>` | string | from the `origin` remote, else `example.com/<name>` | Go module prefix (default from the origin remote, else example.com/<name>) |
| `--no-agent` | boolean | off | leave the MonoSmith agent out; smith agent install adds it later |
| `--recipe <recipe>` | string |  | the recipe smith add creates projects from: a path relative to smith.toml or an https URL (docs/reference/recipes.md) |
| `--recipe-file <recipe-file>` | string |  | a recipe file on this machine (one saved by another workspace): copied to recipes/ and used as the recipe; its answers are not asked |
| `--release-branch <release-branch>` | string | `main` | the branch releases run from (default main) |
| `--release-remote <release-remote>` | string | `origin` | the git remote that receives release tags (default origin) |
| `--save-recipe <save-recipe>` | string |  | save the answers as recipes/<name>.recipe.toml in the workspace, to reuse on others (needs a main language) |
| `--scope <scope>` | string | `@<name>` | npm scope for packages (default @<name>) |
| `-y, --yes` | boolean | off | take the default for every question that has no flag |

<!-- smith:end new -->

Notes on the flags:

- `<name>` must be kebab-case: lowercase letters and digits, words joined by single
  hyphens, starting with a letter. It becomes `[workspace] name` and the default
  npm scope.
- `--ci` takes a comma-separated list or the flag repeated: `github`, `azure`,
  `gitlab`, `circleci` or `bitbucket`.
- `--module-prefix` is read from the `origin` remote of the folder when it has one.
- `--recipe` writes `recipe = ...` under `[workspace]`. See
  [Recipes](../reference/recipes.md). It is not checked at creation time.
- `--dir .` creates the workspace in the current folder. Files smith does not own
  are kept.
- `--here` creates the workspace in the current folder when it is empty (a `.git`, `.vscode`
  or `.DS_Store` does not count) and names it after the folder unless you give a name. A folder
  with files is refused with a problem that points at `smith adopt`.

In a terminal (standard input and output are terminals and `CI` is unset), a missing
`<name>` is asked for: `Name of the workspace:`, up to three tries. Elsewhere it is the
error listed below.

## The questions

At a terminal, `smith new` then asks the questions below, one at a time. The first
answer listed is the default and Enter takes it; type a number or the answer's name to
choose another. The [TUI](tui.md) form (**Create a workspace**) asks the same questions
in the same order. A question whose flag you gave is not asked, `--yes` takes the
default of the rest, and with no terminal nothing is asked: the flags and the defaults
are the whole answer.

| Question | Flag | Default | What it writes |
|---|---|---|---|
| Which CI should smith write pipelines for? | `--ci` | `github` | `[ci] providers` |
| What is the main language? | `--language` | `later` | nothing in `smith.toml`; it decides whether the Node question is asked and the `--lang` of the next step |
| Which tests and lint for Node projects? (asked for Node only) | `--default-stack node=<stack>` | `node-jest` (or `node-vitest`, `node-ts`) | `[default_stacks] node` |
| Which code layout should smith check? | `--architecture` | `vertical-slices` | the `[conventions.<layout>]` table (none for `none`); see [`smith architecture`](architecture.md) |
| Add the MonoSmith agent, so Claude Code and VS Code can work in this workspace through smith? | `--agent` / `--no-agent` | yes | the three [agent files](../guides/agent.md); `no` writes `[agent] enabled = false` |
| Change anything else? | | keep the defaults | "More options" asks the five below |
| Which branch do releases run from? | `--release-branch` | `main` | `[release] branch` |
| Which remote receives the release tags? | `--release-remote` | `origin` | `[release] remote` |
| Create a GitHub Release for every release tag? | `--github-releases` | no | `[release] github_releases` |
| Where should the task cache live? | `--cache-dir` | `.smith/cache` | `[cache] dir` |
| How large may the cache grow, in MB? | `--cache-max-mb` | `2048` (`0` for no limit) | `[cache] max_size_mb` |
| Save these answers as a recipe? (last, asked only with a main language) | `--save-recipe <name>` | no | `recipes/<name>.recipe.toml` |

After the summary, and before any file is written, one last question: **Save these
answers as a recipe to reuse on other workspaces?** (default no; `--save-recipe <name>`
answers it). It is asked only when a main language was chosen, since a recipe must fix
one. On yes it asks the recipe's name (Enter uses the workspace name) and writes
`recipes/<name>.recipe.toml` in the new workspace, in the first commit. The file is an
ordinary [recipe](../reference/recipes.md): the language, for Node its scaffold and
toolchain, and the code layout as the structure plugin. The CI, the release settings and
the cache are workspace settings, not part of a recipe, so they are asked again for
the next workspace. The recipe is not set as the new workspace's own recipe, so you can
still add projects in other languages there.

To create the next workspace from it, give the file: `smith new other --recipe-file
../acme/recipes/team.recipe.toml`. smith copies it to `recipes/` in the new workspace,
sets `[workspace] recipe`, and does not ask what the recipe answers (the language, the
Node tests and lint, the code layout), saying so before the first question. A flag still
wins over the recipe. To share one without copying files, publish it at an https URL and
use `--recipe <url>`, which skips the same questions.

A setting left at its default stays the commented line in `smith.toml`, so the file
still shows what could be set. Giving one of the last five flags settles "More
options": the other four then take their defaults. Task overrides are not asked,
since they depend on projects that do not exist yet; `smith.toml` keeps a commented
example.

A session, answering `node`, `2` (Vitest), the defaults, then "More options" with a
different branch and GitHub Releases:

```text
══ New workspace: acme ═════════════════════════════════════
A few questions. Enter takes the default shown; smith new --yes skips them.

Which CI should smith write pipelines for?
  1) GitHub Actions   default
  2) Azure Pipelines
  3) GitLab CI
  4) CircleCI
  5) Bitbucket Pipelines
Choice [1]:
What is the main language of this workspace?
  1) Decide later             add projects in any language with smith add (default)
  2) TypeScript / JavaScript
  ...
Choice [1]: node
Which tests and lint for Node projects?
  1) Jest             TypeScript, ESLint and Jest (default)
  2) Vitest           TypeScript, ESLint and Vitest
  3) TypeScript only  tsc, no lint and no tests
Choice [1]: 2
...
Your choices
  CI                   github
  Main language        node
  Node tests and lint  node-vitest
  Code layout          vertical-slices
  Releases             from trunk, tags pushed to origin, with a GitHub Release
  Cache                .smith/cache, up to 2048 MB

Save these answers as a recipe to reuse on other workspaces? [y/N]: y
Name of the recipe [acme]: team

  wrote smith.toml
  wrote recipes/team.recipe.toml
  ...
  committed

Created acme. Next:
  cd acme
  smith add app web --lang node

Recipe saved: /work/acme/recipes/team.recipe.toml
For another workspace: smith new <name> --recipe-file /work/acme/recipes/team.recipe.toml
```

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The workspace was created and committed. |
| 1 | The name is invalid, the folder already holds a workspace, or a step failed. |

## Files and environment

Writes, in the new folder: `.gitignore`, `.gitattributes`, `smith.toml`, `README.md`,
`.github/workflows/ci.yml` (and `azure-pipelines.yml` or `.gitlab-ci.yml` for the other
providers), `.devcontainer/devcontainer.json`, `tools/smith/install.sh` (the installer the pipelines run when the
workspace has no npm, pip or `dotnet tool` to install smith with), and the MonoSmith agent: `.claude/agents/smith.md`,
`.mcp.json` and `.vscode/mcp.json` (see [the agent guide](../guides/agent.md)). It then runs `git init -b main`, stages
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
  wrote tools/smith/install.sh
  wrote .claude/agents/smith.md
  wrote .mcp.json
  wrote .vscode/mcp.json
  committed

Created demo2. Next:
  cd demo2
  smith add app web --lang <language>   # smith kinds lists the languages
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
  wrote tools/smith/install.sh
  wrote .claude/agents/smith.md
  wrote .mcp.json
  wrote .vscode/mcp.json
  committed

Created notes. Next:
  smith add app web --lang <language>   # smith kinds lists the languages
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
