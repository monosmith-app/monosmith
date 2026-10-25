# The recipe plugin

The recipe plugin makes a smith recipe (a `*.recipe.toml`) a package project of
its own, so a team can keep its recipes in a smith workspace and version,
test and release them like any package. A recipe is the file that tells
`smith add` which plugins to build a new project with; this plugin is not what
interprets that file (smith's core and `smith recipe` do), it only treats the
file as something to check on every pull request, version from commits, tag and
attach to a GitHub Release.

## At a glance

| | |
|---|---|
| Plugin name | `recipe` (built in; runs as a separate process, `smith plugin serve recipe`) |
| Kind | language |
| Roles | `language`, `toolchain`, `scaffold` (the defaults of a language plugin; it declares none of its own) |
| `--lang` value | `recipe` |
| Manifest that makes a directory a candidate project | `*.recipe.toml` |
| Requires, conflicts, provides (task slots) | none |
| Capabilities advertised | `discover`, `generate`, `version`, `publish` |
| Not advertised | `ci`, `doctor`, `remove`, `published`, `dependencies`, `declarations`, `latest`: no CI step, no doctor check |
| Stacks it publishes | none |
| Project kinds | always `package`, whatever `smith add` is asked for |
| Tasks | `test` |

Roles and capabilities are what `smith plugin list --json` prints; the language,
manifest pattern and absence of requires, conflicts, provides and stacks come
from the plugin's `Describe` method
(`libs/plugin-recipe/plugin/recipe_plugin_handler.go`). For what a recipe
contains and how `smith add` uses one, see [recipes](../../../reference/recipes.md),
in particular [versioned recipes](../../../reference/recipes.md#versioned-recipes).

## What it detects

smith hands the plugin every `*.recipe.toml` in the files git lists. A file is a
project only if all of these hold (`discovery/discover_recipes_use_case.go`):

- **It is in a folder of its own.** The folder holds exactly one recipe file,
  and besides it only documentation (`README`, `CHANGELOG`, `LICENSE`,
  `LICENCE`, `NOTICE`, with any extension) and hidden files (names starting with
  `.`). Any other file, or any sub-folder, disqualifies the folder. Two recipes
  in one folder would be released together, so neither is a project.
- **It is not at the workspace root.** A root recipe is an input of
  `[workspace] recipe` in `smith.toml`, not a project.
- **It is TOML and has a `name`.** One that does not parse, or has no `name`, is
  skipped silently.

For a recipe that qualifies:

| Property | From |
|---|---|
| Project name | the recipe's `name` (not the folder or file name) |
| Kind | always `package` |
| Exports | the folder (`packages/team`) |
| Imports (graph edges) | if `extends` is a path to a folder inside the workspace, that folder (the parent recipe's project). A URL, an absolute path, a path outside the workspace, or the same folder gives none |
| Version | the recipe's `version` |
| Stack, description, tags, toolchain | none |

A real run, after `packages/team` got `extends = "../../apps/dev/dev.recipe.toml"`:

```text
$ smith graph
dev  package · recipe · apps/dev
team  package · recipe · packages/team
  └─ dev
```

A stray file removes the project. With `packages/team/notes.txt` added,
`smith projects` listed only `dev` and `shared`; with only a `README.md` added,
`team` stayed.

## Tasks

One task, `test`, for every recipe project:

| Task | Command | Inputs | Outputs | dependsOn | Cached | Appears |
|---|---|---|---|---|---|---|
| `test` | `smith recipe check <file>.recipe.toml --all` (run in the recipe's folder) | `**/*` | none | none | yes | every recipe project |

`--all` checks the recipe and its parents for every combination of answers (up to
64), so a pull request that breaks one branch of a recipe fails. Because a child
imports its parent's folder, a change to the parent is a change upstream of the
child and invalidates its cached `test`. No `lint`, `build`, `start` or other
task exists; `smith doctor` therefore warns, per recipe project, `dev can be
verified / dev has no lint task, so CI never checks it`.

```text
$ smith explain team:test
team:test — Check the recipe for every combination of answers (smith recipe check --all)
  command    smith recipe check team.recipe.toml --all
  cwd        packages/team
  dependsOn  —
  inputs     **/*
  outputs    —
  env        —
  cache      true

$ smith run test
team:test │ Choices
team:test │   language     node           (default; options node, python) Language?
team:test │ Plugins
team:test │   language     node                     team.recipe.toml:15 (recipe team)
team:test │ The recipe is valid.
team:test ✔ 55ms
```

## Creating a project

```sh
smith add package team --lang recipe
```

Real output:

```text
$ smith add package team --lang recipe
Adding package team (recipe) in packages/team
Wrote 1 file(s):
  packages/team/team.recipe.toml
Check it with smith recipe check packages/team/team.recipe.toml --all, then use it with [workspace] recipe in smith.toml.
Try: smith test -p team
```

The one file is a recipe that resolves: `name`, `version = "0.0.0"`, a
`[choices.language]` choice (`node` or `python`, default `node`), and one
`[when.language.<answer>.use]` block per answer that sets the `language`
level. Edit it, or replace it with the output of `smith recipe new`
([recipe new](../../../manual/recipe-new.md)).

- **The kind is always `package`.** `smith add app dev --lang recipe` created
  `apps/dev/dev.recipe.toml`, and `smith projects` listed it as `package`; so
  does `smith add lib shared --lang recipe` for `libs/shared`. The directory only
  follows the kind you typed.
- **The name must be a release-tag name**: lowercase words joined by dashes,
  starting with a letter. A different one is refused by the plugin:
  `"Team" is not a recipe name: use lowercase words joined by dashes (team,
  team-node), since it starts the release tags` (smith's own check refuses
  `Team` before that, with `lib name "Team" must be kebab-case ...`).
- It refuses an existing directory (`<dir> already exists`).
- The generated hint `Try: smith test -p team` works: `smith test` runs `test`.

## Settings

The plugin has no `[projects.<p>.settings.recipe]` keys and reads no
environment variables. The settings are the recipe file's own top-level keys:

| Where | Key | Effect | Read in |
|---|---|---|---|
| `<name>.recipe.toml` | `name` | the project's name; also starts the release tag (`team@0.1.0`) | `discovery/recipe_file_mapper.go` |
| `<name>.recipe.toml` | `version` | the version a release starts from and rewrites | same, `releasing/recipe_version_mapper.go` |
| `<name>.recipe.toml` | `extends` | a workspace path makes a graph edge to the parent's folder | `discovery/discover_recipes_use_case.go` |
| `smith.toml` | `[release] github_releases = true` | the recipe is attached to the GitHub Release (see below) | host |
| `smith.toml` | `[projects.<name>] tag_prefix` | the release tag | host |
| `smith.toml` | `[tasks.test]`, `[projects.<name>.tasks.test]` | override the check | host |

Only `name`, `version` and `extends` are read by the plugin; the recipe loader
checks every other key (`smith recipe check`).

## Toolchain

None. The plugin runs only `smith` itself (the `{smith}` placeholder in the task
is the running binary), reports no toolchain version, and needs nothing else on
the machine. A `[toolchains]` pin has nothing to match.

## Releasing

A recipe is released like any package: `feat:` commits that touch its folder
release a minor, `fix:` a patch, and so on. `version(team)[2.0.0]: ...` forces a
version.

| | |
|---|---|
| Version in the manifest | the top-level `version = "x.y.z"` of the recipe file |
| Version write | the first `version` line before the first table is replaced in place (single or double quotes; a trailing comment is kept; the value becomes double-quoted); else `version = "x"` is added after `name`, keeping the file's line endings; else at the top. Keys under tables (`[choices.x]`) are never touched. A dry run lists the file and writes nothing |
| Tag | `<name>@<version>` (for example `team@0.1.0`) |
| Publish | nothing is pushed to a registry: "released by its tag". The release has one **asset**, `dist/release/<name>/<file>.recipe.toml`, a copy with the release's version written into it |
| Credentials, private registry | none |
| Dry run | lists the asset it would attach |

The repository's recipe keeps the version it was committed with (a release
commits nothing); only the copy says which release it is.

A real dry run and a real local release (`--no-push`, in the scratch workspace):

```text
$ smith release --dry-run
Dry run: nothing is published or tagged.
  dev                  unreleased → 0.1.0  (1 releasing commit(s))
  team                 unreleased → 0.1.0  (1 releasing commit(s))
dev: would attach dist/release/dev/dev.recipe.toml
team: would attach dist/release/team/team.recipe.toml

$ smith release --no-push --yes
dev: tagged dev@0.1.0
team: tagged team@0.1.0
  dev                  new → 0.1.0  released
  team                 new → 0.1.0  released
```

Afterwards `git tag` listed `dev@0.1.0` and `team@0.1.0`;
`dist/release/team/team.recipe.toml` had `version = "0.1.0"`, and
`packages/team/team.recipe.toml` still said `version = "0.0.0"`.

Consumers pin the released copy by URL:

```toml
[workspace]
recipe = "https://github.com/acme/recipes/releases/download/team@1.3.0/team.recipe.toml"
```

Whether the file is really attached to a GitHub Release depends on
`[release] github_releases = true` and a real GitHub repository; I could not
run that part (see Verified). The copy keeps `extends` as written (the plugin
rewrites only `version`), so a parent given as a relative path is not
rewritten into something that resolves from a release URL; I did not test that
with a downloaded asset.

## CI

Nothing. The plugin has no `ci` capability: no toolchain step, store, setup,
credential or permission. `smith ci generate` printed `pipelines are up to date`
after the recipe projects were added. The recipe `test` runs inside the
workspace's normal `smith ci verify`, with the `smith` binary the pipeline
installs.

## Doctor, dependencies and removal

- **Doctor.** No checks of its own. The host warns that a recipe project has no
  `lint` task (three warnings in the demo: `14 passed, 0 failed, 3 warning(s)`).
- **Dependencies.** None: `smith install`, `smith sync` and `smith up` do nothing
  for recipes.
- **Remove.** `smith remove <project>` deletes the folder and regenerates the
  pipelines; the plugin cleans nothing else. A recipe that another recipe
  extends is refused: `smith: dev is imported by team: remove those imports
  first, or pass --force to remove it anyway` (from `smith remove dev --dry-run`).

## Requirements and limits

- A recipe project must be alone in its folder (see What it detects). A folder
  that holds anything else, or a recipe at the workspace root, is not a project
  and is not released. There is no message; the project is just absent from
  `smith projects`.
- A recipe is always a package: it cannot be an app or a lib.
- Only the recipes in the workspace are versioned. `extends` to a URL or to
  another repository is not a graph edge, so a change there does not invalidate
  the child's `test`.
- `test` checks at most 64 combinations of answers.

## Verified

Windows 11 Pro, smith 1.62.0. No other tool is involved.

Run: `smith new`, `smith add package|app|lib --lang recipe`, the name refusal,
`smith projects` (with a stray file, with a README), `smith graph` (with a real
`extends`), `smith explain`, `smith run test` (including the cached run after a
child changed), `smith recipe check --all`, `smith doctor`, `smith ci generate`,
`smith ci check`, `smith release --dry-run` (plain and `--json`),
`smith release --no-push --yes` (local tags and the stamped asset),
`smith release --status`, `smith remove --dry-run`.

Not verified: attaching the asset to a real GitHub Release (no remote); a
consumer resolving a recipe by that release URL; a recipe that extends another
by a relative path after it has been downloaded alone.

## See also

- [Plugin catalogue](README.md)
- [What is a plugin](../what-is-a-plugin.md) and [managing plugins](../managing-plugins.md)
- [Recipes](../../../reference/recipes.md) and [versioned recipes](../../../reference/recipes.md#versioned-recipes)
- [smith recipe](../../../manual/recipe.md), [recipe check](../../../manual/recipe-check.md), [recipe new](../../../manual/recipe-new.md)
- [Releasing](../../../reference/releasing.md)
- [Writing a language plugin](../../creating/language-plugin.md)
