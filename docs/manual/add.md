# smith add

`smith add` creates a project in the workspace: an app, an internal lib or a
publishable package, in the language and stack you choose. The new project lints,
type-checks, tests and builds immediately, and the CI pipelines are regenerated so
CI installs its toolchain.

| Kind | Meaning | Folder |
|---|---|---|
| `app` | deployed or run, never published | `apps/<name>` |
| `lib` | imported by other projects, never published | `libs/<name>` |
| `package` | published to its registry on release | `packages/<name>` |

A combined kind from [`smith kinds`](kinds.md) names the language too: `node-app` is
`app --lang node`.

## When to use it

- A new project in a workspace made with [`smith new`](new.md), [`smith adopt`](adopt.md)
  or [`smith convert`](convert.md).
- A project from the workspace's recipe, with its choices answered on the command line.

Do not use it to register a project that already exists on disk: smith discovers
manifests it finds (see [`smith projects`](projects.md)). To delete one, use
[`smith remove`](remove.md).

## Synopsis

```sh
smith add <kind> <name> [--lang <language>] [--stack <stack>] [--choice name=answer ...] [--verbose]
```

## Arguments and flags

<!-- smith:begin add -->
Add a project: an app, an internal lib or a publishable package (to add a dependency, use smith install)

Usage: `smith add <kind> <name> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `kind` | yes | `app`, `lib` or `package`, or a combined kind such as `go-lib` (`smith kinds`) |
| `name` | yes | the project's name, in kebab-case |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--choice <choice>` | name=answer, repeatable |  | answer a choice of the workspace's recipe, name=answer (repeatable); the rest are asked in a terminal |
| `-l, --lang <lang>` | string | the stack's, or the workspace's only language | language: node, go, python, rust, dotnet, ruby, php, cpp, dart, jvm, bicep, recipe (default: the stack's, or the workspace's only language) |
| `-s, --stack <stack>` | string | the recipe, else `[default_stacks]` in `smith.toml` | stack the project is built with, bypassing the workspace's recipe (default: the recipe, else [default_stacks] in smith.toml; smith stacks lists them; none: the language plugin alone) |
| `--verbose` | boolean | off | show the package manager's output as it runs (it is shown only if it fails otherwise) |

<!-- smith:end add -->

How the language and stack are chosen:

1. `--stack` wins and bypasses the workspace recipe. `--stack none` uses the language
   plugin alone. [`smith stacks`](stacks.md) lists the names.
2. Otherwise, when `[workspace] recipe` is set in `smith.toml`, the recipe creates the
   project. It asks its choices one at a time in a terminal; `--choice name=answer`,
   repeated, answers them without asking. See [Recipes](../reference/recipes.md).
3. Otherwise the language comes from `--lang`, or the workspace's only language, and the
   stack from `[default_stacks]` in `smith.toml`.

In a terminal (standard input and output are terminals and `CI` is unset), a missing kind
or name is asked for (`Kind (app is deployed, lib is internal, package is published):`
with numbered options, then `Name of the project:`, three tries each); so is the
language when `--lang` is missing and more than one could apply. Elsewhere the missing
argument is the error listed below.

The name must be kebab-case. smith add leaves the files it wrote uncommitted (`smith new`
commits what it wrote): review them with `git status`, then commit.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The project was created. |
| 1 | The kind, name, language or stack is invalid, the name is taken, or generation or the first install failed. |

## Files and environment

Writes the project's files under its folder, and workspace files the plugin owns: the
CI pipeline, `go.work` (Go), the root `package.json`, `package-lock.json`, `tsconfig.base.json`
and `eslint.config.mjs` (Node), `requirements-dev.txt` (Python), the
dev container. It runs the package manager (`npm install --no-audit --no-fund`,
`go mod init`, and so on); its output is shown only when it fails, unless `--verbose`.
Reads `[default_stacks]`, `[workspace] recipe` and `[plugins]` from `smith.toml`.

## Examples

A Node app (stack `node-jest`, the default for Node in this workspace):

```sh
smith add app web --lang node
```

```text
Creating app web in apps/web: node-jest
Adding linting (eslint)
Adding tests (jest)
Wrote 14 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  .npmrc
  apps/web/package.json
  apps/web/src/greeting/greet.use-case.spec.ts
  apps/web/src/greeting/greet.use-case.ts
  apps/web/src/greeting/greeting.contract.ts
  apps/web/src/greeting/index.ts
  apps/web/src/main.ts
  apps/web/tsconfig.build.json
  apps/web/tsconfig.json
  eslint.config.mjs
  package.json
  tsconfig.base.json
Import it from another Node project as '@acme/web'.
Try: smith test -p web
```

A Go library using the combined kind, no `--lang`:

```sh
smith add go-lib core
```

```text
Creating lib core in libs/core: go
Wrote 7 file(s):
  .github/workflows/ci.yml
  libs/core/doc.go
  libs/core/go.mod
  libs/core/greeting/doc.go
  libs/core/greeting/greet_use_case.go
  libs/core/greeting/greet_use_case_test.go
  libs/core/greeting/greeting_contract.go
Import it from another Go project as github.com/acme/acme/libs/core/<slice>.
Try: smith test -p core
```

A Python package:

```sh
smith add package tools --lang python
```

The files land in `packages/tools` with `src/acme_tools/greeting/` as the first slice, and
the same `Try: smith test -p tools` hint is printed.

## Common errors

| Message | Cause and remedy |
|---|---|
| `smith add needs a kind (app, lib or package) and a name` | Missing arguments. `smith add app web --lang node`; `smith kinds` lists them. |
| `kind "widget" is not app, lib or package` | Use one of the three kinds or a combined kind such as `go-lib`. |
| `no plugin handles language "nothing"; available: bicep, cpp, ...` | Pick a language from the list, or enable its plugin (`smith plugin enable`). |
| `no stack named "nope"; known stacks: android, angular, ...` | `smith stacks` lists the stacks. |
| `a project named "web" already exists (apps/web)` | Choose another name. |
| `app name "Bad_Name" must be kebab-case: ...` | Lowercase words joined by hyphens. |

## Related

[`smith remove`](remove.md), [`smith kinds`](kinds.md), [`smith stacks`](stacks.md),
[`smith plugin`](plugin.md), [Stacks](../reference/stacks.md), [Recipes](../reference/recipes.md).
