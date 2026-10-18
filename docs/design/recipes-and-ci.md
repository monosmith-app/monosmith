# Design: recipes, and CI as part of them

Status: **proposal**, for discussion. Decisions already taken are marked
**Decided**; the rest are **Open**, each with a recommendation. Nothing here
is built yet unless it says so.

Built so far (epic #312): plugins declare their level, requirements and
conflicts (#307); the recipe file, its loading with URL pinning, resolving and
validating, and `smith recipe check` (#306, see
[recipes](../reference/recipes.md)); named private registries (#310); creating
projects from a recipe with `smith add` and `smith new --recipe` (#308),
which records the recipe, its version and the answers in `smith.toml`. To
come: drift checks and upgrades of the projects a recipe made (#309), the
creation wizard and versioned recipes (#311).

## The goal

MonoSmith exists so that a team can keep its internal libraries, in many
languages, in one monorepo that is easy to set up and that releases
consistently. Four things make that hard today, in every tool:

1. **Many languages, one workspace.** Each language has its own build, test,
   lint and publish story.
2. **Configuration.** Getting a workspace to build *and release* takes far
   more knowledge than it should. Nx is flexible, but its releases are
   unpredictable, `affected` is hard to understand, and dependency
   installation is inconsistent.
3. **Conventions change.** A scaffold captures today's conventions; the
   projects made last year are stuck with last year's.
4. **Teams differ.** One opinionated setup (the old mnci) does not fit
   everyone, but a hundred options are as bad as none.

The answer this document proposes: **plugins** do the work, **recipes**
describe which plugins a project uses and how, a recipe can extend another and
offer choices, and CI is part of the recipe, not a separate system.

## What exists today

smith already has most of the parts, under other names:

| Today | Is |
|---|---|
| Plugins (out-of-process, any language), with **roles**: language, toolchain, structure, scaffold, deploy | the code that does the work |
| A toolchain plugin's `provides` (`test`, `lint`, `e2e`): two plugins cannot provide the same task | slots, enforced |
| **Stacks** (`node-vitest`, `node-vitest-playwright`, …) | a fixed combination of plugins, but only a plugin can publish one |
| **Conventions** (`[conventions.vertical-slices]`, `[conventions.docker]`) | a plugin applied to every project, with settings |
| The stack recorded in each project's manifest | which combination made the project |
| `smith doctor` with safe fixes for files smith owns; `smith upgrade` | keeping projects current, for owned files |
| `smith explain <project>:<task>` | where every setting came from |
| The generated CI pipeline: thin YAML calling `smith ci <phase>`; plugins contribute toolchains, setup, audit and credentials | CI from the workspace's own description |

What is missing is a way for a **team** to write its own combination, share
it, extend it, offer choices in it, and have CI follow from it.

## Recipes

### What a recipe is

**Decided:** a recipe is **pure data**: a description of which plugin fills
each role, and with what settings. It contains no code, so it cannot break
anything by itself; it can only choose and configure. Plugins stay the only
thing that runs. (ESLint plugin vs shareable config, Babel plugin vs preset,
Renovate presets.)

### One parent

**Decided:** a recipe has **at most one parent** (`extends`), like a
Dockerfile's single `FROM`. A team recipe extends a company recipe, which
extends nothing. One parent keeps the result easy to reason about: the order is
always parent first, then the child, and later wins.

### Choices, including the root

**Decided:** a recipe offers the choices its author wants, and any setting can
depend on them, including the most basic one, the language. One "Team Recipe"
can cover Python and C#: the developer picks the language, and the rest
follows from it.

That needs some logic in a recipe. To keep recipes data, the logic is
**declarative**: named choices, and `when` blocks that apply only for certain
answers. No expressions, no scripts. (Copier's `when`, Backstage template
parameters, GitHub Actions' `if` on a fixed vocabulary.)

```toml
# team.recipe.toml
name = "team"
version = "1.2.0"
extends = "https://example.com/recipes/company/v3.recipe.toml"  # optional, one parent

[choices.language]
ask = "Language"
options = ["python", "csharp"]

[choices.kind]
ask = "What are you creating?"
options = ["library", "service"]
default = "library"

# Always, whatever the answers.
[use]
structure = "vertical-slices"

# Only for Python.
[when.language.python.use]
language = "python"
test = "pytest"
lint = "ruff"

# Only for C#.
[when.language.csharp.use]
language = "dotnet"
test = "xunit"

[when.language.csharp.settings.dotnet]
target_framework = "net10.0"

# Only for a C# service.
[when.language.csharp.when.kind.service.use]
deploy = "docker"
```

- `use` fills roles (slots) with plugins; `settings.<plugin>` configures one.
- `when.<choice>.<answer>` applies a block only for that answer, and nests
  for combinations.
- A choice can be offered only when it matters: a `when` block can hold its
  own `[choices.x]`.

**Open: how far the logic goes.** Recommendation: only "this answer selects
this block", nested. No comparisons, no computed values. If a team needs more,
that is a plugin's job, and the recipe sets the plugin.

### Levels: what a recipe chooses

**Decided.** Every plugin plays at a level. Some levels are **exclusive**:
a project has exactly one plugin there. Others are **additive**: a project can
have any number, and they only add, never replace.

| Level | Kind | Examples | Does |
|---|---|---|---|
| Language | exclusive, one per project | `node`, `python`, `dotnet`, `go` | finds the project, builds, versions and publishes it |
| Toolchain | exclusive, one per task slot (`lint`, `test`, `typecheck`, `format`, `e2e`) | `eslint` (lint), `jest` (test), `ruff` (lint), `pytest` (test) | runs one tool for one task |
| Scaffold | exclusive, one per project kind | the TypeScript starter, the Go plugin starter | writes a new project's starter files |
| Structure | additive | `vertical-slices` | an architecture rule, checked across the project |
| Deploy and conventions | additive | `docker`, `playwright` | extra tasks and release steps for the projects they match |

**Rules and configuration are not plugins.** A lint rule set (an ESLint
shareable config, a Ruff configuration, Checkstyle rules) is a versioned
package that a toolchain plugin reads, chosen by a setting:

```toml
[settings.eslint]
config = "@acme/eslint-config"   # extends @monosmith/eslint-config, published like any package
```

The toolchain plugin knows how to run the tool; the package says what it
checks. When the team's rules change, a new package version and `smith deps
up` bring every project along, with no plugin change and no scaffold edit.

**One concern can span levels without owning a slot.** A structure plugin
such as `vertical-slices` adds:
- its own check task (`slice-check`)
- its rules to the linter's configuration, through the linter's setting
- its folder layout to the files a scaffold writes

It never replaces the linter or the test runner, so it composes with any of
them.

**Where a new thing belongs.** Starter files are written once and then belong
to the project, and upgrades touch only files smith owns. So anything that
must keep evolving with the team must be either:
- **owned**: generated by smith and regenerated on upgrade, such as CI, base
  configs and editor settings
- **a dependency**: a versioned package, such as lint rules or shared
  tsconfig bases

Never a starter file. That rule decides the level of every new feature.

**Plugins declare their level** in the protocol's handshake:
- the level and, for a toolchain, the slots it fills (`provides` today)
- what it requires, e.g. "a Node language plugin"
- what it conflicts with

Resolution checks all three before anything is written.

### Resolving a recipe

When a project is created (or checked), smith resolves its recipe:

1. Load the parent chain, oldest first.
2. Apply each recipe's unconditional blocks, then its matching `when` blocks
   in file order. A later value replaces an earlier one, key by key.
3. Check the result before anything is written:
   - every role has at most one plugin (as `provides` does today)
   - every plugin's requirements are met, e.g. "needs a Node language plugin"
   - no two chosen plugins conflict
   - every plugin named exists, at a version in range
4. Record the recipe, its version and the answers in the project's manifest,
   where the stack is recorded today.

A conflict is an error at resolve time, naming both sides and the recipe
lines that chose them. It never surfaces half-way through creating a project.
`smith explain` shows, for each setting, the recipe and line it came from.

### Using a recipe

```toml
# smith.toml
[workspace]
name = "libs"
recipe = "./recipes/team.recipe.toml"   # or a URL
```

- `smith new` and `smith add` ask the recipe's choices, and nothing else.
- `smith add --choice language=python --choice kind=library` answers them
  without prompting (CI, scripts).
- A recipe can also be named for one project kind only.

### Keeping projects current

**Decided (for now):** `smith upgrade` and `doctor` update **files smith owns**
only: generated config, CI, files marked as generated. Starter files that
people edit are not merged. (Projen's model. Copier's three-way merge of
starter files is a later option.)

With the recipe and its version recorded in each project:
- `doctor` reports drift: "orders is on team 1.2.0; 1.3.0 changes the lint
  plugin's settings".
- `smith upgrade` re-resolves with the same answers and rewrites what smith
  owns.

### Where recipes live

**Decided (for now):** a recipe is a **file or a URL**. A team keeps its
recipes in a repository and points at a raw URL, pinned by tag. No registry
until one is needed (#164 stays open for plugins).

Open: whether a URL recipe is cached and pinned in a lockfile, so a build
does not change because a URL's content did. Recommendation: yes. Record the
content's hash beside the URL, as a lockfile does, and refuse a changed hash
until `smith upgrade`.

### Creating a recipe

**Decided (later, after recipes work):** recipes get a creation wizard, not
only the choice prompt. `smith recipe new` builds one interactively:

1. It asks for the parent, if any.
2. It goes through the roles one by one, listing the plugins that can fill
   each: the built-in ones and those on the plugin search path.
3. For each role it asks whether the plugin is fixed, or a choice the
   recipe's users make.

It validates as it goes (one plugin per role, requirements, conflicts), so a
recipe the wizard writes always resolves. It writes the same TOML a person
would, so the result can be edited and reviewed like any file. `smith recipe
check <file>` runs the same validation on a recipe written by hand.

### Versioning a recipe

**Decided (later):** recipes are versioned the way code is, from commit
messages. A recipe is a project kind (`recipe`) in a smith workspace, usually
the team's recipes repository:
- `feat(team-recipe): …` releases a minor, `fix` a patch, and a breaking
  change a major, as for any package
- `smith release` tags it `team-recipe@1.3.0` and attaches the file to the
  release
- a recipe URL points at a tag, so a workspace pins a version and `smith
  upgrade` moves it forward
- `doctor` reports a newer version of the recipe a project was made with

So the recipes repository is itself a smith workspace, and its releases need
nothing beyond what smith already does for packages.

## CI

### What is already right

- **The pipeline is generated and thin.** Every step is `smith ci <phase>`, so
  what CI runs is what a laptop runs, and no manifest carries CI tricks. The
  old mnci's "install Python from package.json" problem cannot happen.
- **Plugins contribute their own part:** toolchains, setup, audit and the
  credentials they need. A new language plugin brings its CI with it.
- **Only what is needed runs:**
  - a pull request verifies the affected projects
  - pack runs only projects with a `package` task
  - release publishes only what the commits since each project's last tag
    say changed, never everything

### What is missing: private registries

Installing from and publishing to an internal feed is the hardest part of
monorepo CI today (Azure Artifacts with npm, PyPI and NuGet feeds is the
common case). smith knows public registries and a few tokens, and nothing about
"our feed". Proposal:

```toml
# in a recipe, or smith.toml
[registries.internal]
auth = "azure-artifacts"                 # or "github-packages", "token"
npm = "https://pkgs.dev.azure.com/acme/_packaging/libs/npm/registry/"
pypi = "https://pkgs.dev.azure.com/acme/_packaging/libs/pypi/upload/"
nuget = "https://pkgs.dev.azure.com/acme/_packaging/libs/nuget/v3/index.json"

[release]
registry = "internal"                    # where packages publish by default
```

- smith writes the client configuration itself (`.npmrc`, `pip.conf`,
  `nuget.config`), so local installs and CI resolve the same feed.
- The generated pipeline authenticates the provider's own way:
  - on Azure Pipelines, its authenticate tasks (`npmAuthenticate@0`,
    `PipAuthenticate@1`, `NuGetAuthenticate@1`) with the pipeline's own token,
    so an Azure Artifacts feed needs no secret at all
  - on GitHub, `GITHUB_TOKEN` with `packages: write` for GitHub Packages
  - elsewhere, a named token
- A team recipe sets its registries once, and every repository created from it
  installs and publishes there with no setup.
- `smith ci check` and `doctor` say what is missing, in the provider's terms:
  "the feed `libs` must give the build service Contributor".

### Wizard or plugins?

**Recommendation: neither as a separate system.** CI settings are recipe
settings: providers, the release branch, registries and environments. They are
chosen like the language, through the recipe's choices, with the same prompt
as `smith add`. A separate CI wizard that writes YAML would be one more thing
to drift. What a person must still do by hand (create a secret, grant a feed
permission) is listed by `smith ci check`, with the exact steps.

**Open: CI providers as a plugin role.** GitHub Actions, Azure Pipelines and
GitLab are rendered by the core today. Recommendation: keep those three in the
core, and add a `ci-provider` plugin role later only if someone needs
another (Jenkins, Bitbucket). The provider-neutral model already exists, so
that role would be a renderer.

**Open: deploy environments.** Releases publish packages and images today.
Deploying apps to environments (dev, test, prod), with approvals, is the next
CI question. Recommendation: a separate document once recipes exist, since
environments belong in the recipe too.

## Dependencies across the workspace

Not a plugin question, but the same philosophy, and a pain point:

- **Dev tools live at the root, once** (ESLint, Jest, …), as smith does today.
- **Production dependencies live in each project**, with **one version per
  dependency** across the workspace (pnpm catalogs, Gradle version catalogs).
- `smith deps up axios` updates it in every project that uses it, in one
  change, and `doctor` flags two projects on different versions.

**Open:** whether the version policy is on by default. Recommendation: on, with
a per-dependency exception list. The plugin protocol already has the
`dependencies/*` and `declarations` methods it needs.

## Summary of decisions

| | Decision | Status |
|---|---|---|
| 1 | Recipes are pure data describing plugins and their settings | Decided |
| 2 | One parent per recipe | Decided |
| 3 | Recipes offer any choices, including the language; declarative `when` blocks | Decided; how far the logic goes is Open |
| 4 | Upgrades update files smith owns only | Decided, for now |
| 5 | Recipes are a file or a URL | Decided, for now; URL pinning is Open |
| 5b | A recipe creation wizard (`smith recipe new`) and `smith recipe check` | Decided, later |
| 5c | Recipes versioned from commit messages, as a `recipe` project kind | Decided, later |
| 5d | Plugin levels: exclusive (language, each task slot, scaffold) and additive (structure, deploy, conventions); rules and configs are versioned packages, not plugins | Decided |
| 6 | CI is part of the recipe, no separate wizard | Open, recommended |
| 7 | Named private registries, with provider-native auth | Open, recommended |
| 8 | CI providers stay in the core; a renderer role later | Open, recommended |
| 9 | One version per production dependency across the workspace | Open, recommended |
| 10 | Deploy environments | Open, separate document |

## Prior art

| Tool | What it taught |
|---|---|
| Docker | one parent (`FROM`), layers applied in order, history to see where a layer came from |
| ESLint, Babel, Renovate | plugins (code) separate from configs and presets (data); presets extend presets; shared from a URL or a repo |
| Debian, Nix | packages declare what they provide, require and conflict with; resolved before installing |
| Projen | files the tool owns are regenerated, never hand-edited |
| Copier, Cruft | a project records the template version that made it, so it can be brought forward |
| Backstage templates, Yeoman | a scaffold asks a few named questions its author chose |
| pnpm catalogs, Gradle version catalogs | one version of each dependency for the whole workspace |
