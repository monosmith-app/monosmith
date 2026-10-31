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
smith add <kind> <name> [--lang <language>] [--stack <stack>] [--with <choice,...>] [--choice name=answer ...] [--fresh] [--verbose]
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
| `--fresh` | boolean | off | give the project the built-in default of its stack, without following the sibling projects |
| `-l, --lang <lang>` | string | the stack's, or the workspace's only language | language: node, go, python, rust, dotnet, ruby, php, cpp, dart, jvm, bicep, lua, applescript, vba, swift, objc, bundle, recipe (default: the stack's, or the workspace's only language) |
| `-s, --stack <stack>` | string | the recipe, else `[default_stacks]` in `smith.toml` | stack the project is built with, bypassing the workspace's recipe (default: the recipe, else [default_stacks] in smith.toml; smith stacks lists them; none: the language plugin alone) |
| `--verbose` | boolean | off | show the package manager's output as it runs (it is shown only if it fails otherwise) |
| `--with <with>` | string, repeatable |  | the choices the project is built from, comma separated: smith choices lists them (react,vitest,bootstrap); an axis left out takes its default |

<!-- smith:end add -->

How the language and stack are chosen:

0. `--with react,vitest,bootstrap` builds the project from choices. See
   [Building from choices](#building-from-choices).
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

## Building from choices

`--with` names the choices a project is built from instead of a ready-made stack: a framework, a
test runner, a style. [`smith choices`](choices.md) lists them.

```sh
smith add app web --with react,vitest,bootstrap
```

```text
Choices: react, vitest, bootstrap, eslint (eslint by default) as the stack react-vitest-bootstrap
```

smith checks the selection first (one choice per axis, what each requires, what conflicts) and
gives one message that names the choices and the fix, for example `chakra requires react, which is
not in the selection: add react to --with, or drop chakra`. An axis that a choice needs and the
selection leaves out takes its default. The result is saved in `smith.toml` as a stack with
`choices`, and the project records it. `--with` cannot be used with `--stack` or `--choice`. See
[Stacks](../reference/stacks.md#choices-a-stack-as-a-saved-selection).

## Following the siblings

A workspace has habits: its libs build with esbuild, its packages test with
`@swc/jest`, their `main` points into `dist/`. `smith add` reads the **siblings**, the
projects of the same kind and language, and gives the new project the same shape. It
prints what they agree on and why:

```text
The 3 sibling packages agree on:
  build.tool = esbuild (3 of 3 packages)
  exports.target = dist (3 of 3 packages)
  test.runner = @swc/jest (3 of 3 packages)
```

The rules, from [ADR 0041](../architecture/decisions/0041-smith-add-follows-the-dialect-of-sibling-projects.md):

- A value is followed when **at least 3** siblings have it and **at least two thirds** of
  the siblings that report the key. With fewer than 3 siblings nothing is inferred and
  nothing is printed.
- When the siblings **split** (no value has two thirds), `smith add` asks in a terminal and
  lists the values with their counts and the default of the stack. Without a terminal, or
  with `CI` set, it never waits: it uses the default and prints a warning that names the
  split.
- After the project is made, each key the siblings agree on that the new project does not
  have is a warning (`Warning: 3 of 3 sibling packages have tsconfig.files = ...; the new
  package has ...`). The warning means the stack cannot follow that key.
- `--fresh` skips all of it: the project gets the default of its stack.

The facts for a Node project are `build.tool` (`tsc`, `esbuild`, ...), `test.runner`
(`jest`, `ts-jest`, `@swc/jest`, `vitest`, ...), `exports.target` (`dist`, `src`),
`tsconfig.files`, `lint.config`, `scripts` and `layout`. The stacks follow
`build.tool` (`tsc`, `esbuild`), `exports.target` and `test.runner` (`node-jest` follows
`ts-jest` and `@swc/jest`; `node-vitest` follows `vitest`); the others are reported and
warned about. A stack does not switch plugins for a sibling shape: if the siblings use
Vitest and the stack uses Jest, the warning is your cue to pass `--stack node-vitest`.
Languages whose plugin does not report a shape are not affected.

For a Node project, the folder follows the root `package.json` `workspaces`. When they
cover `libs/*` (or the kind's own folder), the project goes there. When they do not (a
workspace converted from Nx lists only `packages/*` and `apps/*`), a lib goes into the
first folder the globs name that exists, `packages/` first, so npm links it at once.

After the first build and test, `smith add` runs the language plugins' `smith doctor`
checks for the new project (npm workspaces, scripts, entry files). If one fails, it
prints the check with its remedy, exits with 1 and does not print "Done"; follow the
remedy, then run `smith doctor`.

The name must be kebab-case. smith add leaves the files it wrote uncommitted (`smith new`
commits what it wrote): review them with `git status`, then commit.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The project was created. |
| 1 | The kind, name, language or stack is invalid, the name is taken, generation or the first install failed, or the new project fails its `smith doctor` checks. |

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
