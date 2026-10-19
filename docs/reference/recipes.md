# Recipes

> Terms such as project, kind, task and stack are defined in the [glossary](glossary.md). The design behind recipes: [recipes and CI](../design/recipes-and-ci.md).

A recipe is a team's description of which plugin fills each level of a project
and with what settings. It is data: a TOML file with no code, so it can only
choose and configure plugins, never run anything. Recipes can extend one
another and ask a few choices, and the answers decide the rest.

Recipes are being built in steps (epic #312). What exists today is the file
format, loading, resolving, validating, `smith recipe check`, creating
projects from a recipe with `smith add` (below), and keeping those projects
current: `smith doctor` reports drift, `smith upgrade` brings them up to date
and `smith explain` shows where each value came from (see
[Keeping projects current](#keeping-projects-current)), writing a recipe with
the wizard (`smith recipe new`) and releasing recipes like packages
([Versioned recipes](#versioned-recipes)). The whole design is in
[recipes and CI](../design/recipes-and-ci.md).

## Using a recipe

```toml
# smith.toml
[workspace]
recipe = "recipes/team.recipe.toml"   # a path relative to smith.toml, or an https URL

[recipes]                             # optional: another recipe for one project kind
package = "https://example.com/recipes/package.recipe.toml"
```

`smith new acme --recipe recipes/team.recipe.toml` writes the setting into a
new workspace's `smith.toml`. A URL must be reachable and a recipe, or `new`
refuses before creating anything; a path that is not there yet is only said,
since a recipe kept in the same repository may be added next.

With a recipe configured, `smith add <kind> <name>` creates the project from it:

1. **It asks the recipe's choices.** In a terminal, one question at a time, in
   the order the recipe declares them, with the default pre-selected (Enter
   takes it). A choice that a `when` block adds is asked only after the answer
   that makes the block apply. An answer is an option's text or its number.
   `--choice name=answer`, as often as needed, answers without asking. With no
   terminal (or `CI` set) a choice not given takes its default, and one with
   neither fails, naming the choice and the flag that answers it. Nothing is
   written before every choice is answered.
2. **It shows what the recipe will use**, for each level or slot, with the
   recipe line that chose it, and the settings it carries:

   ```text
   Recipe team 1.2.0 (language=python, kind=library) will use:
     language   python           team.recipe.toml:21 (recipe team)
     structure  vertical-slices  team.recipe.toml:22 (recipe team)
   ```

3. **It makes the project exactly as a stack's project is made.** The resolved
   plugins become a stack: `use.language` and `use.scaffold` are its language
   and scaffold, `use.structure` and `use.deploy` its structure and deploy
   plugins, `use.toolchain` its toolchain in the order given, and the plugin
   of any other key (a slot such as `lint`) follows it. The scaffold (or the
   language plugin) writes the files, each other plugin adds its part, and the
   pipelines are regenerated. The recipe's `settings.<plugin>` tables are given
   to each plugin they name, when it generates the project and afterwards in
   the project's `settings`; those of a plugin the project does not use are
   left out, and said.
4. **It records what it did in `smith.toml`:** the stack
   (`[stacks.<recipe>-<answers>]`, shared by every project made with the same
   answers; a different stack that already has the name gets a short hash
   after it), the project's recipe (name, version, source and answers) and its
   settings. See [configuration](configuration.md#recipes). If creating the
   project fails, `smith.toml` is put back as it was.

```sh
smith add package billing --choice language=python            # the other choices take their defaults
smith add app basket --choice language=csharp --choice kind=service
smith add lib core --stack none --lang go                     # bypasses the recipe
```

`--stack` bypasses the recipe: the project is made as it was before the
workspace had one, with `[default_stacks]` and `--lang`. `--choice` with
`--stack` is an error, and so is `--choice` where no recipe is configured.
`--lang` is accepted when it names the language plugin the recipe picks for
these answers, and an error when it does not.

Existing stacks, `--stack` and `[default_stacks]` keep working unchanged: a
workspace with no recipe behaves as before.

## Keeping projects current

A recipe changes: a new lint plugin, a changed setting, a new version. The
projects it made before do not follow by themselves. Each one recorded the
recipe, its version and its answers when `smith add` made it, so smith can
resolve the recipe as it is now with the same answers and say what differs.

**`smith doctor`** does that for every project made from a recipe. A project
that has drifted is a warning, in the recipe's terms:

```text
! orders follows its recipe team
    orders is on team 1.2.0; the recipe is 1.3.0, which changes the settings of
    the lint plugin eslint (config: "@acme/eslint-config" → "@acme/eslint-config/strict")
    → run `smith upgrade` (`smith upgrade --dry-run` shows what it changes first)
```

What it compares is the project's stack (the plugins at each level: a plugin
the recipe adds, drops or replaces), the settings in `smith.toml` for its
plugins, and the answers recorded. A project the recipe has not changed passes.
Some things are reported as warnings, never as failures:

- the recipe cannot be loaded (a path that moved, a URL with the network down
  and no pinned copy); a URL recipe fetched once is read from its pinned copy
- the recipe now asks a choice the project has no answer for and that has no
  default: the warning names the choice, and the answer to add under
  `[projects.<name>.recipe.answers]` (a new choice with a default takes the
  default, and upgrade records it)
- the recorded answers no longer fit the recipe (it dropped a choice or an
  option)
- `[workspace] recipe` now names a different recipe from the one the project
  was made with, so there is no way to say how it differs

**`smith upgrade`** brings the projects up to date. For each one that has
drifted it resolves the recipe again with the recorded answers and:

1. rewrites what `smith.toml` holds for the project: the recorded recipe
   version and answers, its settings and, when the recipe now picks other
   plugins, the stack declaration, with everything else in the file kept as it
   was
2. has the project's plugins restore the files they own from the new settings,
   through the safe fixes their doctor checks offer (a fix that only restores
   what a plugin owns, such as a generated configuration file)
3. regenerates the CI pipelines, as upgrade always does

```sh
smith upgrade --dry-run   # the plan, and the change to smith.toml as a diff; changes nothing
smith upgrade --check     # the same, and fails when something would change (CI)
smith upgrade             # does it, and prints what it changed
```

**Only files smith owns change.** The files a plugin wrote once when the
project was created (a starter test, a config you then edited) are yours and
are never rewritten. When the recipe adds a plugin, a new project would get its
starter files, but an existing project does not: upgrade says so, per project
and plugin, and `smith doctor` lists what is missing. A fix that is not safe
(it could change a file you edited) is named and left; `smith doctor --fix`
shows its diff.

A project is left as it is, with the reason, when:

- its stack is provided by a plugin, not declared in `smith.toml` (smith cannot
  change the plugins a plugin's stack holds)
- its stack is shared with a project that the recipe does not change the same
  way, or that was not made from a recipe: rewriting the stack would change
  that project too
- the recipe's plugins cannot work together, or the language plugin is not
  enabled in this workspace

A stack declared by a recipe is named after the recipe and the answers
(`team-python-library`) and its description does not carry the recipe's
version: the version is recorded for each project.

**`smith explain <project>`** also lists what the project's recipe decided for
it: its choices, the plugin at each level and the plugins' settings, each with
the recipe file and line it came from. `--recipe` prints only that, and with
`--json` prints it for tools:

```text
$ smith explain orders
recipe team 1.3.0 — recipes/team.recipe.toml
  choices.language       python
                           from recipes/team.recipe.toml:6 (recipe team)
  use.lint               eslint
                           from recipes/team.recipe.toml:14 (recipe team)
  settings.eslint.config "@acme/eslint-config/strict"
                           from recipes/team.recipe.toml:21 (recipe team)
```

## The file

```toml
# team.recipe.toml
name = "team"
version = "1.2.0"
extends = "company/base.recipe.toml"   # optional, one parent: a path or an https URL

[choices.language]
ask = "Language"
options = ["python", "csharp", "node"]

[choices.kind]
ask = "What are you creating?"
options = ["library", "service"]
default = "library"

[when.language.python.use]
language = "python"
structure = "vertical-slices"

[when.language.csharp.use]
language = "dotnet"
structure = "vertical-slices"

[when.language.csharp.settings.dotnet]
target_framework = "net10.0"

[when.language.csharp.when.kind.service.use]
deploy = "docker"

[when.language.node.use]
language = "node"
lint = "eslint"
test = "vitest"
```

| Key | Does |
|---|---|
| `name` | required; names the recipe in messages |
| `version` | the recipe's version |
| `extends` | the parent recipe: a path relative to this file, or an `https://` URL. At most one |
| `[choices.<name>]` | a question: `ask` (the text), `options` (the answers allowed) and `default` |
| `[use]` | fills a level or a slot with a plugin (below) |
| `[settings.<plugin>]` | configures a plugin; any TOML, tables merge |
| `[when.<choice>.<answer>]` | a block (`choices`, `use`, `settings`, more `when`) that applies only for that answer |

An unknown key is an error with its file, line and column, and the keys that
would have been accepted there. So is a `default` that is not one of the
`options`, a missing `name`, and a `use` value that is not a plugin name.

### Use: levels and slots

Each key of `use` is a **level** or a **slot**, and its value is a plugin name
(`language = "python"`) or, for the levels that take several, a list
(`structure = ["vertical-slices"]`).

| Key | Is | Takes |
|---|---|---|
| `language` | exclusive level | one plugin |
| `scaffold` | exclusive level | one plugin |
| `structure` | additive level | any number |
| `deploy` | additive level | any number |
| `toolchain` | toolchain plugins, without naming their slots | any number |
| anything else (`lint`, `test`, `typecheck`, `e2e`, ...) | a **slot**: the task a toolchain plugin provides | one plugin, which must provide that slot |

A plugin may carry a version constraint after `@`: `lint = "eslint@^1.2"`
(`1.2.3`, `^1.2`, `~1.2` and `>=1.2` work, as in `[plugins.external]`).

Which plugins can be named, and which level each plays, is what
`smith plugin list` shows. See [plugins](plugins.md).

## Choices and `when` blocks

A choice is answered by `--choice name=answer`, else by its `default`. One
with neither is an error that names it and says how to answer. A given answer
must be one of the `options`.

A `when.<choice>.<answer>` block applies only when the choice has that answer.
There is no logic beyond that: no comparisons and no computed values. A block
can hold its own `choices`, asked only when the block applies, and its own
`when` blocks, so combinations nest:

```toml
[when.language.csharp.when.kind.service.use]
deploy = "docker"          # only for a C# service
```

A `when` block must name a choice a recipe in the chain declares, and an
answer among its options; a typo is an error, not a block that never applies.

## Resolving

`Resolve` takes the chain of recipes and the answers and produces one result:

1. The choices of every recipe in the chain are merged key by key, the
   child's over the parent's: a child can change only the `default` of a
   choice and keep its `options`.
2. The recipes are applied **oldest first**. Within one recipe the
   unconditional `use` and `settings` come first, then the `when` blocks that
   match, **in file order**. A block's own `use` and `settings` come before
   its nested `when` blocks.
3. A later value replaces an earlier one, key by key. Settings tables merge
   recursively; a value that is not a table, including a list, replaces.
4. Every value records the recipe, file and line it came from.

## Validating

Before anything is written, the result is checked with the same rules a stack
is checked with:

- every plugin named exists, as a built-in plugin, one declared in
  `smith.toml` or one on the plugin search path, and at a version its
  constraint accepts
- a slot is filled by a plugin that provides it
- the language and the scaffold are one plugin each, and a language is chosen
- one plugin per exclusive level and per slot, every requirement met, no
  conflicts, and every plugin applies to the chosen language

An error names both sides and the recipe lines that chose them:

```
error: eslint applies to [node], not to python projects
error:   chosen at:
error:     language = "python"  broken.recipe.toml:4 (recipe broken)
error:     lint = "eslint"  broken.recipe.toml:5 (recipe broken)
```

Settings for a plugin the recipe does not use are carried, not checked.

## Recipes from URLs

`extends` (and the recipe you check) may be an `https://` URL. Plain `http://`
is refused. A recipe loaded from a URL can extend only other URLs, relative to
its own.

In a workspace a URL recipe is **pinned**: its content is cached under
`.smith/recipes/` (build output, ignored by git) and its SHA-256 recorded in
`smith.recipes.lock`, which is committed:

```toml
# Written by smith. It pins the content of every recipe loaded from a URL, so
# a build does not change because a URL's content did. Review a change, then
# accept it with smith upgrade.

[[recipe]]
url = "https://example.com/recipes/team.recipe.toml"
sha256 = "9f2c..."
```

- The first load fetches the recipe and pins it.
- Later loads use the cached copy without the network, so a URL changing
  changes nothing.
- On a machine without the cache (a fresh clone) the recipe is fetched and
  must match the pinned hash. A recipe that changed is refused, with a message
  pointing at `smith upgrade`.
- Accepting the new content is `Loader.Accept(url)` (all pinned URLs:
  `AcceptAll`). It fetches again and replaces the cached copy and the pin, and
  reports the old and new hash. `smith upgrade` calls it for every recipe
  `smith.toml` names (`smith upgrade --check` only reports a recipe that
  changed, and sees it only where the cache is missing, such as a fresh clone
  or CI; `smith upgrade` always asks the URL).

Outside a workspace nothing is cached or pinned; the URL is fetched each time.

## Versioned recipes

A team keeps its recipes in a repository of their own, usually a smith
workspace, and releases them the way it releases packages. The built-in
`recipe` plugin makes each recipe a project:

```
recipes/
  company/
    company.recipe.toml
  team/
    team.recipe.toml        extends = "../company/company.recipe.toml"
    README.md
```

```sh
smith add package team --lang recipe      # a recipe that resolves, in packages/team/
smith recipe new recipes/team/team.recipe.toml   # or write it with the wizard
smith run test -p team                    # smith recipe check team.recipe.toml --all
```

- **A project is a recipe file in a folder of its own**: one `*.recipe.toml`,
  and beside it only a README, CHANGELOG or LICENSE. The project is named by
  the recipe's `name`, has kind `package` and takes its version from the
  recipe's `version`. A recipe at the workspace root (an input of
  `[workspace] recipe`), two in one folder, one beside another project's
  files, or one with no name or that is not TOML is not a project.
- **Its `test` task checks every answer** (`smith recipe check --all`), so a
  pull request that breaks one branch of a recipe fails CI. A recipe that
  extends another recipe in the workspace depends on it: a change to the
  parent affects the child and its tests.
- **Versions come from commits**, like any package: `feat(team): ...` that
  touch the recipe's folder release a minor, `fix` a patch, a breaking change
  a major (`version(team)[2.0.0]` forces one). `smith release` tags
  `team@1.3.0` and, with `[release] github_releases = true`, attaches the
  recipe to the GitHub Release.
- **The attached file carries the release's version.** A release commits
  nothing, so the recipe in the repository keeps the version it was committed
  with; the copy attached to the release (built in
  `dist/release/<name>/`) has `version` set to the release's. Pin that copy:

  ```toml
  [workspace]
  recipe = "https://github.com/acme/recipes/releases/download/team@1.3.0/team.recipe.toml"
  ```

  A URL that points at the tag's source (`raw.githubusercontent.com/.../team@1.3.0/...`) also works
  and is pinned by content like any URL, but its file shows the version it
  was committed with.
- **`smith doctor` reports a newer release** of a recipe the workspace pins
  this way ("team 1.4.0 is released, and smith.toml pins 1.3.0") with the URL
  to change it to. It asks GitHub's API for the repository's releases of that
  recipe, says nothing when it cannot (offline, rate limited), and is off with
  `SMITH_NO_UPDATE_CHECK`. `smith upgrade` then accepts the new content and
  brings the projects up to date.

## `smith recipe check`

```sh
smith recipe check team.recipe.toml --choice language=csharp --choice kind=service
smith recipe check https://example.com/recipes/team.recipe.toml --json
```

Loads the recipe and its parents, answers the choices, resolves and validates
it. It prints the recipes in the chain, the choices asked and how each was
answered, the plugin for each level and the merged settings, each with its
recipe, file and line, or why the recipe is not valid. The exit code is
non-zero when it is not. A choice with no answer is listed as needing one.

`--json` prints the same for tools: `name`, `version`, `chain`, `choices`,
`use`, `settings`, `origins`, `pinned` and `errors` (empty when valid).

`--all` checks every combination of answers (up to 64), not only the one given
or defaulted, so a recipe whose python branch is fine and whose csharp branch
is broken is caught. The report still shows the single answer given, when
there is one; a recipe whose choices have no default is checked through its
options. This is what a recipe project's `test` task runs.

## Writing a recipe: `smith recipe new`

```sh
smith recipe new                                  # asks everything; writes <name>.recipe.toml
smith recipe new recipes/team/team.recipe.toml    # writes there
smith recipe new --from-answers answers.toml      # no terminal: the same answers, from a file
```

The wizard asks:

1. the recipe's **name** (lowercase words joined by dashes: it names the
   recipe's releases), its **version**, and the recipe it **extends**, if any
   (a path, from the new file, or an https URL). The parent is loaded at
   once, so a wrong path is caught while you type it, and the choices it asks
   can be used by the levels below.
2. each **level** a plugin can fill, in the order a recipe is written: the
   language, the scaffold, each task slot (`lint`, `test`, ...), structure and
   deploy. For a level you say whether its plugin is **fixed** (`f`), a
   **choice** the recipe's users make (`c`), or **left unset** (enter).
3. for a choice: its name, the question, the answers (at least two) and the
   default; and the plugin each answer picks. A later level can use the same
   choice, so one `language` choice picks the linter and the test runner too.

Only plugins that **fit** are offered: those that can fill the level, apply to
the language chosen so far, and agree with every plugin already chosen (the
rules of [Validating](#validating)), in each branch of the choices. With a
`language` choice of python and node, the linter that needs Node is offered
for the node branch and not for the python one, and a plugin that fits only
one branch cannot be fixed. A requirement on something other than the
language may still be met by a plugin chosen later, so it does not rule a
plugin out. Built-in plugins, those `smith.toml` declares and those installed
on the search path are all offered.

Before anything is written the finished recipe is rendered, parsed, and
checked with its parents for **every combination of answers**, as
`smith recipe check --all` does. A recipe that does not resolve is not
written, and the error says for which answers and which plugins. The file is
the TOML you would write by hand, with the choices, the plugins that are
fixed under `[use]`, and a `[when.<choice>.<answer>.use]` block per answer.
An existing file is replaced only after you confirm, or with `--force`.

The wizard does not write `settings` or nested `when` blocks; add them by
hand, and run `smith recipe check --all`.

### Answers file

`--from-answers` reads what the questions would learn, as TOML, so a script, a
test or CI gets the same recipe (the questions and the file go through the
same code after that):

```toml
name = "team"
version = "0.1.0"
extends = "../company.recipe.toml"      # optional

[choices.language]
ask = "Language"
options = ["python", "node"]
default = "python"

[levels.language]
choice = "language"                      # a plugin per answer
plugins = { python = "python", node = "node" }

[levels.lint]
choice = "language"
plugins = { node = "eslint" }            # python: unset

[levels.structure]
fixed = ["vertical-slices"]              # any number for structure and deploy
```

A level is `fixed` plugins or a `choice` with `plugins` by answer (a name, or
a list for structure and deploy), never both. A choice the parent declares may
be used without declaring it again. Unknown keys and a name or version a
release cannot use are refused.
