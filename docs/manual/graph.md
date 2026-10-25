# smith graph

`smith graph` shows how the projects depend on each other, as a tree in the terminal
(`text`), JSON, Graphviz (`dot`), a static page (`html`) or an interactive browser (`tui`).
Every format shows the same projects. It can also answer two questions about change: what a
change would reach (`--blast-radius`), and how the graph differs from an earlier revision
(`--diff`).

An edge means "imports": a project depends on another when its code or manifest imports it,
in the same language (npm workspaces, Go modules in `go.work`, Python and so on) or through a
shared schema file (OpenAPI and the like) that `--blast-radius` follows across languages.

## When to use it

- To see what depends on a project before changing it.
- To produce a diagram (`--format dot | dot -Tsvg`) or a page (`--format html -o graph.html`).
- In review: `--diff main` for what a pull request changed in the structure.

## Synopsis

```sh
smith graph [--format text|json|dot|html|tui] [-o <file>] [--focus <projects>] [--affected [--base <branch>]]
smith graph --blast-radius <project or file> [--format text|json]
smith graph --diff <revision> [--format text|json]
```

## Arguments and flags

<!-- smith:begin graph -->
Show how projects depend on each other

Usage: `smith graph [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--affected` | boolean | off | keep the projects affected by changes since --base; mark the changed ones |
| `--base <base>` | string | the pull request target, else the release branch | base branch for --affected (default: the PR target, else the release branch) |
| `--blast-radius <blast-radius>` | project or file path |  | list everything a change to a project or a file reaches, across languages and through shared schemas; text or json |
| `--diff <diff>` | revision |  | show how the graph changed since a revision (a branch, tag or commit): projects and imports added and removed, kinds and stacks changed; text or json |
| `--focus <focus>` | string, repeatable |  | keep these projects and everything connected to them (names, globs, tag:x) |
| `--format <format>` | `text`, `json`, `dot`, `html` or `tui` | `text` | text, json, dot, html or tui |
| `-o, --output <output>` | string |  | write to a file |

<!-- smith:end graph -->

Interactions:

- `--focus` keeps the named projects (names, globs, `tag:x`) and everything connected to
  them, in both directions; the named projects are marked `*`.
- `--affected` keeps the projects the changes since `--base` affect and marks the changed
  ones with `*`. `--base` defaults to the pull request target, else the release branch.
- `--blast-radius` and `--diff` are queries of their own: they print text or JSON
  (`--format json`) and not the other formats.
- `-o` writes to a file; with `html` it is the usual choice (`wrote graph.html`).
- `--format tui` opens an interactive browser and needs a terminal.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Printed. |
| 1 | Unknown format, a revision that does not exist, or no workspace. |

## Files and environment

Reads the project graph and, for `--diff` and `--affected`, git history. Writes only the
file `-o` names.

## Examples

```sh
smith graph
```

```text
api  app · go · apps/api
tools  package · python · packages/tools
ui  lib · node · libs/ui
web  app · node · apps/web
```

With no imports between projects, each is a line. When `web` declares `@acme/ui` as a
dependency, `web` shows what it uses beneath it:

```text
api  app · go · apps/api
tools  package · python · packages/tools
ui  lib · node · libs/ui
web  app · node · apps/web
  └─ ui
web2  app · go · apps/web2
```

`--focus ui` keeps `ui` (marked `*`) and everything connected to it:

```sh
smith graph --focus ui
```

```text
ui *  lib · node · libs/ui
web  app · node · apps/web
  └─ ui
```

After a change in `apps/web` (here the edge above is absent again):

```sh
smith graph --affected
```

```text
web *  app · node · apps/web
affected: changes since main
```

```sh
smith graph --format dot
```

```text
digraph smith {
  rankdir=BT;
  node [shape=box, style=rounded];
  "api" [label="api\napp · go"];
  "tools" [label="tools\npackage · python"];
  "ui" [label="ui\nlib · node"];
  "web" [label="web\napp · node"];
}
```

What a change to a project, or to a file in it, reaches:

```sh
smith graph --blast-radius libs/ui/src/index.ts
```

```text
libs/ui/src/index.ts starts at ui
  web                  via import of ui
```

and for a project nothing depends on:

```text
web2 starts at web2
nothing else depends on it
```

How the graph changed since an earlier commit:

```sh
smith graph --diff HEAD
```

```text
project graph from HEAD to the working tree
+ api (go app, apps/api)
+ tools (python package, packages/tools)
+ ui (node lib, libs/ui)
+ web (node app, apps/web)
```

(In that trial the workspace was still uncommitted, so every project was new.)

## Common errors

| Message | Cause and remedy |
|---|---|
| `unknown format "bogus" (text, json, dot, html, tui)` | Use one of the five. |
| `nonexistent is not a revision: git rev-parse --verify ...` | `--diff` needs a branch, tag or commit that exists. |

## Related

[`smith affected`](affected.md), [`smith projects`](projects.md), [`smith explain`](explain.md),
[How smith finds projects and edges](../architecture/README.md).
