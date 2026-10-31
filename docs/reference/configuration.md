# Configuration: `smith.toml`

> Terms such as project, kind, task and stack are defined in the [glossary](glossary.md). The full list of commands and flags is in the [CLI reference](cli.md).

One file at the workspace root, and the only one smith reads. Its presence marks
the workspace root. Every setting has a default; a typo in a key is an error,
never a silently ignored setting.

## Editor support

`smith.toml` has a JSON Schema: completion, hover docs and validation in any
editor that reads TOML schemas. `smith new` and `smith convert` write its
address as the first line,

```toml
#:schema https://raw.githubusercontent.com/monosmith-app/monosmith/main/apps/smith/workspace/smith.schema.json
```

which Even Better TOML (VS Code) and other Taplo-based editors pick up. Add
the line to an older workspace's `smith.toml` to get the same. `smith schema`
prints the schema of the smith you run, for tools that want it offline. The
schema is versioned with the file: its `schema = 1` is the version it
describes, and `smith upgrade` migrates older files.

## Changing a setting from the command line

`smith config` reads and changes the settings of the `[workspace]`, `[ci]`,
`[release]`, `[cache]` and `[plugins]` tables:

```sh
smith config list                          # every setting, its value, what it does
smith config get ci.providers
smith config set ci.providers github azure # a list takes several values
smith config set cache.max_size_mb 4096
smith config unset release.branch          # back to the default
```

Values are checked against the schema before anything is written. smith
changes only that setting's value, so comments and every other line stay as
they are, and it refuses to write a file that would not load. Edit stacks,
conventions, tasks and projects in the file itself.

```toml
schema = 1

[workspace]
name = "acme"                         # required
scope = "@acme"                       # npm scope for Node projects
module_prefix = "github.com/acme/acme"  # Go module path prefix
exclude = ["examples/*"]              # directories whose manifests are not projects (below)
registry = "libs"                     # the [registries] entry installs come from (below); default: the public registries
recipe = "recipes/team.recipe.toml"   # new projects are created from this recipe: a path or an https URL (docs/reference/recipes.md)

[recipes]                             # a different recipe for one project kind: app, lib or package
package = "https://example.com/recipes/package.recipe.toml"

[ci]
providers = ["github"]                # "github", "azure", "gitlab", "circleci", or several
devcontainer = true                   # write .devcontainer/devcontainer.json with CI's toolchains (docs/reference/ci.md)
branches = ["main", "dev"]             # branches the GitHub and Azure pipelines run on, pushes and pull requests; the release branch is always one; default: the release branch (docs/reference/ci.md)
# [[ci.sibling_repos]]: repo, ref, env, dir, pre_build, token_secret: other repositories the GitHub workflow checks out beside the workspace (docs/reference/ci.md)
azure_pool = "Managed-Windows"        # the agent pool an Azure Pipelines definition runs on; default: Microsoft's hosted ubuntu-latest
azure_variable_groups = ["Build"]     # variable groups the definition links, for the secrets its steps read; default: none
native_runners = ["ubuntu-latest", "windows-latest", "macos-latest"]  # where native projects build (docs/reference/ci.md)
native_linux_packages = []            # apt packages the native job installs on Linux
smith_version = "latest"              # the smith the pipelines install: "latest" or an exact version; default: the version that generates them (docs/reference/ci.md)
audit_policies = []                    # "node", "go", "python", "rust": policy checks smith ci audit runs beyond the vulnerability scan (docs/reference/ci.md)
audit_warn = []                        # audit policies that only report
audit_allow_install_scripts = []       # npm packages the node policy lets run install scripts
audit_max_unsafe = 0                   # unsafe blocks the rust policy accepts in all

[dependencies]
one_version = true                    # one version per production dependency; dev tools at the root (below); default: false, smith new writes true
allow = ["left-pad"]                  # dependencies that may differ between projects

[boundaries]
kinds = true                          # no project depends on an app: an error (unset: a warning; false: off)
# [[boundaries.rules]] from/allow/deny/reason, [[boundaries.exceptions]] from/to/reason (below)

[release]
branch = "main"                       # real releases run only here
remote = "origin"                     # release tags are pushed here
github_releases = false               # a GitHub Release per tag (needs GITHUB_TOKEN)
tag_format = "{name}@"                # the tag prefix of a project with no tag_prefix; "{scope}/{name}@" gives @acme/kit@ (releasing.md, Tags)
rubygems_trusted_publishing = false   # publish gems with the workflow's OIDC token, not GEM_HOST_API_KEY (GitHub only; docs/reference/releasing.md)
vsce_entra = false                    # publish VS Code extensions as a Microsoft Entra ID app, not with VSCE_PAT (GitHub only; docs/reference/releasing.md)
registry = "libs"                     # where packages publish; default: [workspace] registry, then the public registries

[cache]
dir = ".smith/cache"                   # workspace-relative, or absolute
max_size_mb = 2048                     # keep the local cache under this size in MB; 0 = no limit ("The cache size limit")
remote = "s3://acme-smith-cache/main"  # a shared remote cache: https://, s3://, azblob:// (docs/reference/tasks-and-cache.md)
remote_read_only = false               # read the remote without writing to it

[agent]
enabled = false                       # no MonoSmith agent files on purpose (smith agent install); default: true

[plugins]
enabled = ["node", "go", "python"]    # default: every built-in plugin
marketplace = "https://example.com/plugins/index.json"   # where `smith plugin search|install <name>` look; default: monosmith-app/plugins

[default_stacks]                      # a project's stack when it records none (docs/reference/stacks.md)
node = "node-jest"

[stacks.ts-func]                      # declare a stack: the plugins a project is built with
description = "Azure Function in TypeScript"
language = "node"
toolchain = ["eslint", "vitest"]
structure = []
scaffold = ""                         # default: the language plugin generates
deploy = []

[projects.web]
stack = "node-vitest"                 # replaces the stack web's manifest records

[projects.sdk]
registry = "public-npm"               # this package publishes here instead of [release] registry

[projects.java]
registry = "none"                     # tag it and create its GitHub Release, publish it to no registry (a registry that builds from the tag, such as JitPack); no credentials are asked for
bare_tags = true                      # tag releases 0.1.0 instead of java@0.1.0, because JitPack uses the tag name as the version; one project only, not with tag_prefix

[projects.api.settings.docker]        # a plugin's settings for this project, by plugin name; only the plugin that owns them reads them (docker here)
context = "services/api"

[projects.studio.settings.tauri]      # a Tauri app's sidecar programs and resources, copied for tauri-build (plugins/using/catalogue/tauri.md)
sidecars = [{ name = "videocraw", project = "cli", bin = "videocraw" }, { name = "tool", from = "tools/tool" }]
resources = ["assets/models", { from = "$VIDEOPLUGINS/dist", to = "plugins/" }]

[registries.libs]                     # a private registry (below)
auth = "azure-artifacts"              # or "github-packages", or "token" with token_env
npm = "https://pkgs.dev.azure.com/acme/_packaging/libs/npm/registry/"

[conventions.vertical-slices]         # apply a convention (docs/reference/conventions.md)
# a convention takes only the keys shown here; an external plugin's own settings go in [projects.<p>.settings.<plugin>]
languages = ["go"]                    # default: every language it supports
exclude = ["legacy"]                  # projects it leaves alone

[conventions.docker]                  # an image task for every app with a Dockerfile
registry = "ghcr.io/acme"             # where release pushes the images; unset: built, never pushed
# registries = ["ghcr.io/acme", "docker.io/acme"]   # or several (never with registry): the same image under the same tags in each
# registries = [{ url = "registry.acme.dev/team", username_env = "ACME_USER", password_env = "ACME_TOKEN" }]
description_from = "README.md"        # Docker Hub only: the repository description (see releasing.md)
short_description = "What it does"    # Docker Hub only: the one-line summary, 100 characters at most; default: the project's own description
tag_commit = false                    # also tag every image sha-<short commit> (releasing.md)
tag_branch = false                    # also tag every image with the release branch's name
push = true                           # default: true when a registry is set
verify = true                         # smith ci verify builds the affected apps' images; default: true

[projects.api.docker]
image = "acme-api"                    # the image's name under the registry; default: the project's
registries = ["docker.io/acme"]       # this app's registries (or registry), instead of the convention's
push = false                          # build this app's image, never push it
context = "apps/api"                  # the build context; default: ".", the workspace
verify = false                        # this app's image is not built on pull requests

[plugins.external.kiwi]
version = "^1"                        # installed on the plugin search path (docs/reference/plugins.md)

[plugins.external.rust]
command = ["smith-plugin-rust"]        # any executable speaking the protocol
env = ["CARGO_REGISTRY_TOKEN", "AWS_*"] # variables it may see beyond the baseline (names, or a prefix ending in *)
sha256 = "f6967352942e34f139256eafee1da615931c1f0e0888ddecead3d9a38c15a239"   # the plugin file's SHA-256; in CI it must match
```

The keys of `[plugins.external.<name>]`:

| Key | Meaning |
|---|---|
| `command` | the argv that starts the plugin; empty: find it installed on the plugin search path |
| `version` | a version constraint: `1.2.3`, `^1.2`, `~1.2` or `>=1.2` |
| `env` | environment variables the plugin may see beyond the baseline every external plugin gets: names, or prefixes ending in `*`. Added to what its `plugin.json` declares. It is the only way to give a plugin a variable smith uses for its own cache or releases (`SMITH_CACHE_KEY` and the like) |
| `sha256` | the SHA-256 of the plugin's file (the script, when the command runs an interpreter on one), 64 hexadecimal characters. In CI an external plugin starts only when this matches; write it with `smith plugin trust <name> --pin` |

An external plugin also has to be trusted on the machine that runs it before smith starts it:
`smith plugin trust <name>`. See [Plugin security](../plugins/using/plugin-security.md).

## Private registries

`[registries.<name>]` names a private package registry: an Azure Artifacts
feed, GitHub Packages, or any registry that takes a token. Give it the URL of
each ecosystem it serves, all optional, and how CI signs in:

| `auth` | Signs in with | Secrets to create |
|---|---|---|
| `azure-artifacts` | Azure Pipelines' own authenticate tasks (`npmAuthenticate@0`, `PipAuthenticate@1`, `TwineAuthenticate@1`, `NuGetAuthenticate@1`) and the pipeline's token | none |
| `github-packages` | the workflow's `GITHUB_TOKEN`, with `packages: read` (or `write` to publish) | none |
| `token` | the variable `token_env` names | that variable |

Then `[workspace] registry` makes installs resolve from it, and
`[release] registry` (default: the install registry) or `[projects.<name>]
registry` makes packages publish there. Without them nothing changes: the
public registries, as before.

An Azure Artifacts feed for npm, Python and .NET, end to end:

```toml
[workspace]
name = "libs"
registry = "acme-libs"                # install from the feed; its upstream sources serve npmjs.com, PyPI and nuget.org

[ci]
providers = ["azure"]

[registries.acme-libs]
auth = "azure-artifacts"
npm = "https://pkgs.dev.azure.com/acme/_packaging/libs/npm/registry/"
pypi = "https://pkgs.dev.azure.com/acme/_packaging/libs/pypi/simple/"   # pypi_upload defaults to .../pypi/upload/
nuget = "https://pkgs.dev.azure.com/acme/_packaging/libs/nuget/v3/index.json"
```

GitHub Packages' Maven repository, with the job's token and no secrets:

```toml
[registries.github]
auth = "github-packages"
maven = "github"          # or https://maven.pkg.github.com/<owner>/<repo>

[release]
registry = "github"
```

What smith does with it:

- **npm:** `smith doctor --fix` writes a block into the workspace's `.npmrc`
  (between `# smith:registry` markers; the rest of the file stays yours):
  `registry=` the feed, and for a token or GitHub Packages registry the
  `_authToken` line reading the variable. GitHub Packages serves only its own
  scopes, so it is set for the packages' scope (`@acme:registry=`).
- **Python:** CI setup points the venv's pip at the index (`pip config
  --site`), so every install through `.venv` uses it; twine uploads to
  `pypi_upload`.
- **.NET:** `smith doctor --fix` writes `nuget.config` with the feed (and
  nuget.org beside it, unless the feed is Azure Artifacts, whose upstream
  sources serve it). A `nuget.config` without smith's marker comment is yours:
  doctor shows the change and leaves it to you.
- **CI:** the pipeline signs in before setup, gives setup the token when
  installs need one, and drops `NODE_AUTH_TOKEN`, `PYPI_TOKEN` and
  `NUGET_API_KEY` for packages published to the feed.
- **JVM (Maven and Gradle):** `maven` is the repository packages deploy to.
  `maven = "github"` means GitHub Packages,
  `https://maven.pkg.github.com/<owner>/<repo>` of the repository the release
  runs in (`GITHUB_REPOSITORY`, else the `origin` remote); an explicit URL
  works too. With `auth = "github-packages"` the user is `GITHUB_ACTOR` and the
  password the job's `GITHUB_TOKEN`: the workflow gets `packages: write` and no
  Maven secret. `token` uses `MAVEN_USERNAME` and the variable `token_env`
  names; `azure-artifacts` signs in with `MavenAuthenticate@0` (Maven only:
  Gradle needs `SYSTEM_ACCESSTOKEN` mapped into the release step). With no
  registry, `MAVEN_REPOSITORY_URL`, `MAVEN_USERNAME` and `MAVEN_PASSWORD` work
  as before. GitHub Packages needs a token even to read a Maven package, so
  whoever depends on yours needs their own with `read:packages`.
- **What is left to you:** `smith ci check` lists it in the provider's terms.
  For Azure Artifacts: give the project's build service the Feed Publisher
  (Contributor) role on the feed, add the upstream sources, and on a laptop
  sign in once with vsts-npm-auth, artifacts-keyring or the Azure Artifacts
  Credential Provider.

`auth = "azure-artifacts"` signs in only on Azure Pipelines, and
`github-packages` only on GitHub Actions: on another provider (GitLab
included) use `auth = "token"` with a personal access token. With `token`,
pip reads its credentials from `~/.netrc` or a keyring; smith sets only the
index.

## Dependencies

The one-version policy keeps a workspace on one copy of each library, as pnpm catalogs and Gradle version catalogs do, without moving declarations out of the projects:

- **Production dependencies are declared in each project**, at the same version in every project of an ecosystem (Node with Node, Python with Python).
- **Dev tools are declared at the root, once** (ESLint, Jest, ...), as the Node stack does today.

```toml
[dependencies]
one_version = true        # default: false; `smith new` writes true
allow = ["left-pad"]      # may differ between projects
```

With `one_version = true`, `smith doctor` fails when projects of one ecosystem declare a production dependency at different versions (it names each project and spec) and warns about a dev tool declared in a project as well as at the root, or by two projects. Peer dependencies, the workspace's own packages and the names in `allow` are never reported. `allow` also keeps `smith sync` and `smith deps up --align` from touching those names.

To fix a fragmented dependency, `smith deps up --align [name...]` rewrites every declaration to one spec: the version installed, with the operator most projects use, else the highest version declared. `smith up <name>` and `smith deps up <name>` move a dependency to its latest release in every project that declares it, in one change. Both go through the plugins' `dependencies/declared` and `dependencies/rewrite` methods, so any language plugin that lists declarations takes part; Go does not need it, since a `go.work` builds every module with the highest version required. The policy is a workspace setting; a recipe cannot carry it yet (it has no workspace scope).

## Boundaries

Rules for which projects may depend on which, checked on the project graph (imports and shared schemas, in every language):

```toml
[boundaries]
kinds = true                        # nothing depends on an app: an error. Unset: a warning; false: off

[[boundaries.rules]]
from = "tag:ui"                     # a name, a glob, kind:app, lang:go or tag:ui
deny = ["tag:server"]               # may not depend on these
allow = ["kind:lib"]                # or: may depend only on these
reason = "the browser must not pull in server code"

[[boundaries.exceptions]]           # one dependency that may stay
from = "web"
to = "admin"
reason = "the shared login page, until it moves into a lib"

[projects.web]
tags = ["ui"]                       # added to the tags the project's plugin reports
```

`smith doctor` fails on each dependency that breaks one of your rules, naming the two projects (`web -> admin`), the rule and what to do; `smith boundaries` prints the same list, and `smith ci verify` fails on it. The built-in rule (no project depends on an app) is only a warning until you set `kinds = true`, so an existing workspace does not turn red. An exception needs a `reason`. See [`smith boundaries`](../manual/boundaries.md). A recipe cannot carry the rules yet (they are workspace settings).

## Toolchains

Pin the runtimes the workspace builds with, by the label the plugins report (`node`, `go`, `python`, `rustc`, `java`, `dotnet`, `ruby`, `php`, `dart`, `cmake`):

```toml
[toolchains]
node = "22"      # any 22.x
go = "1.23"      # any 1.23.x
```

A version is a prefix on dot boundaries: `22` accepts 22.3.0 and not 220.0.0. `smith doctor` fails when a machine runs another version, naming the remedy (`mise use node@22`), and `smith doctor --fix` writes the pins to `.tool-versions`, which mise and asdf read and install from; the fix is applied automatically only when that file is new, otherwise it is shown as a diff. smith does not download runtimes itself ([ADR 0020](../architecture/decisions/0020-toolchain-pins-checked-not-downloaded.md)). The installed version is already part of every task's cache key, so a machine on another version never reuses a result. A pin for a tool no project uses, or that is not installed, gives no finding.

## Leaving directories out

smith makes a project of every manifest it finds (`package.json`, `go.mod`,
`pyproject.toml`, `Cargo.toml`, …). `[workspace] exclude` lists the
directories that are not projects, as workspace-relative globs with forward
slashes:

```toml
[workspace]
exclude = ["examples/*", "tools/scratch"]
```

A manifest is left out when its directory, or a directory above it, matches a
glob: `examples/*` covers `examples/movies/package.json` and everything under
`examples/movies`. No plugin sees those manifests, so they have no tasks, and
the graph, `smith run`, `smith doctor` and CI ignore them. A change inside an
excluded directory affects no project. Files directly in `examples/` (a
README) are not covered by `examples/*`; use `examples/**` to cover them too.

Use it for folders a repository keeps standalone on purpose, such as
examples users copy and install on their own. A Node package outside the root
`package.json` `"workspaces"` is a typical case: npm does not install it with
the workspace, so `smith doctor` asks you to either add it to `"workspaces"`
or exclude it. `smith convert` excludes such packages itself when Nx knew
them only through their `package.json`.

## Shared schemas

Projects in different languages are linked by the schema they share. Each
`[projects.<name>]` table takes two lists of workspace-relative globs:

```toml
[projects.orders]
schemas = ["contracts/*.proto"]        # owned besides the OpenAPI, AsyncAPI, Protocol Buffers
                                       # and GraphQL files inside the project's directory

[projects.orders-go]
consumes_schemas = ["services/orders/openapi.yaml"]
```

A consumer depends on the schema's owner: a change to the schema affects both
(`smith affected`), and `smith graph --blast-radius` lists what it reaches.
See [Affected](tasks-and-cache.md#shared-schemas).

A consumer can have its client generated from the schema:

```toml
[projects.orders-go.generate]
schema = "services/orders/openapi.yaml"   # consumed, so no consumes_schemas needed
generator = "go"                          # openapi-generator -g, or the protoc plugin
version = "7.8.0"                         # required: pins the generator
out = "generated"                         # default; the task's output
# command = ["buf", "generate", "{schema}"]   # replaces the default command
```

This adds a cached `generate` task; see
[Generated code](tasks-and-cache.md#generated-code-from-schemas).

## Recipes

`[workspace] recipe` names the recipe `smith add` creates projects from: a
path relative to `smith.toml`, or an https URL. `[recipes]` names a different
one for a project kind (`app`, `lib` or `package`). `smith new --recipe` writes
the setting. [Recipes](recipes.md) explains them.

When a recipe creates a project, smith adds three kinds of table to
`smith.toml`, under a comment that says where they came from. They are ordinary
settings: edit or remove them like any other.

```toml
# Written by smith from the recipe team for the project billing.
[stacks.team-python-library]          # the plugins the recipe picked, as a stack other projects with the same answers share
description = "From the recipe team (language=python, kind=library); written by smith"
language = "python"
structure = ["vertical-slices"]

[projects.billing.recipe]             # which recipe, version and answers made the project
name = "team"
version = "1.2.0"
source = "recipes/team.recipe.toml"

[projects.billing.recipe.answers]
kind = "library"
language = "python"

[projects.billing.settings.python]    # the recipe's settings for a plugin the project uses
layout = "src"
```

The project's manifest records the stack's name, as for any stack. A plugin
reads `[projects.<name>.settings.<plugin>]` as its entry in the project's
`settings` (and gets it as `settings` when it generates the project). `smith
remove` drops a project's `recipe` and `settings` tables with the rest of its
tables; the stack stays, since other projects may use it.

## The cache size limit

`[cache] max_size_mb` is the size the local cache is kept under (default 2048). After a `smith
run` (or `build`, `test`, `lint`, `typecheck`, a `smith ci` phase, `smith release` and the other
commands that run tasks) that stored new results, smith removes the least recently used entries past the limit,
then the blobs no entry uses. It does not walk the store every time: it looks at most about once an
hour, or sooner when one run stored 4 MB or more and the store is probably over. It leaves
anything used or written in the last five minutes, and a problem while pruning is a warning,
never a failed run.

- `max_size_mb = 0` (or a negative number) means no limit: smith never prunes by itself.
- The limit applies to the local cache only. The remote cache is never pruned by smith.
- It applies in CI too, to that machine's local cache.
- `smith cache status` shows the size against the limit, `smith cache prune` trims now (and to
  another size with `--max-size-mb`), and `smith doctor` warns when the cache is over. See
  [`smith cache`](../manual/cache.md).

## Overriding a task

Plugins infer every task. To change one, say only what differs:

```toml
# every project's test task
[tasks.test]
cache = false

# one project's build
[projects.web.tasks.build]
depends_on = ["^build", "codegen"]
inputs = ["src/**", "!src/**/*.spec.ts", "{workspaceRoot}/tsconfig.base.json"]

# a task the plugin does not know about needs a command
[projects.web.tasks.e2e]
command = ["npx", "playwright", "test"]
depends_on = ["build"]
```

Fields: `command` (argv, never a shell string), `cwd` (workspace-relative,
`"."` for the root; default: the project root), `inputs`, `outputs`,
`depends_on`, `env`, `cache`, `verify` (run the task in `smith ci verify` on
pull requests; an app's `package` task does by default, `verify = false` turns
it off), and `serve`, `serve_url` and `serve_timeout`, which start an app around
the task ([Serving an app for a task](tasks-and-cache.md#serving-an-app-for-a-task)).

There is exactly one precedence rule: **plugin → `[tasks.<task>]` → matching
`[[task_groups]]` (in file order) → `[projects.<name>.tasks.<task>]`**, later
wins, field by field. See [Task groups](#task-groups) for the third layer. To see the
result and where each value came from:

```sh
smith explain web:build
```

```
web:build — Compile to dist/
  command    npm run build --silent
               from plugin node
  dependsOn  ^build, codegen
               from smith.toml [projects.web.tasks.build]
  ...
```

A `depends_on` override replaces the list, with one exception: a task that a
plugin added to the project for the overridden one stays, whether the plugin
is in the project's stack or applied as a convention (`[conventions.<name>]`).
For example, the `e2e-browsers` task the playwright plugin runs before `e2e`
stays, because `e2e` cannot run without it; `smith explain` shows it as
`from smith.toml …, with e2e-browsers from convention playwright`. Any other
dependency a plugin gave the task and the override leaves out (`build`,
`^build`) is gone. `smith doctor` names each one as a warning, so the drop is
deliberate rather than silent.

## Task groups

When several projects need the same overrides (every Node library builds into
`dist`, tests need the same variable), state them once, for the group, instead
of once per project:

```toml
# every Node library
[[task_groups]]
select = ["lang:node", "kind:lib"]

[task_groups.tasks.build]
outputs = ["dist"]

[task_groups.tasks.test]
depends_on = ["^build", "logger:build"]
env = { TS_NODE_COMPILER_OPTIONS = '{"module":"commonjs"}' }

[task_groups.tasks.typecheck]
depends_on = ["^build", "build", "^typecheck"]

# one library uses all of that except for its test task
[projects.parser.tasks.test]
depends_on = ["^build"]
```

`select` lists selectors, written the way `--projects` and `[boundaries]` write
them: a project name, a glob over names or folders (`libs/*`), `kind:app`,
`kind:lib`, `kind:package`, `lang:<language>` or `tag:<tag>`. A project is in
the group when it matches **every** selector (add a second group for "or").
A group sets the same fields as `[tasks.<task>]`, and changes only tasks the
project already has.

- Groups apply in the order of the file. When two groups match a project and
  set the same field of the same task, the later one wins; fields only one of
  them sets are kept.
- A project's own `[projects.<name>.tasks.<task>]` wins over every group,
  field by field: `parser` above keeps the group's `build` and `typecheck`, and
  its `test` takes its own `depends_on` and the group's `env`.
- `tag:` selectors see the tags set in `[projects.<name>] tags`.
- `smith explain parser:test` says where each value came from, for example
  `from smith.toml [[task_groups]] select = ["lang:node", "kind:lib"]`.
- Cache keys follow the resolved task, so a change in a group re-runs exactly
  the tasks it changes.
- `smith.toml` is rejected when a group has no `select`, an empty selector or
  an unknown `kind:`, or sets no task. `smith doctor` warns about a group that
  matches no project.
- The `depends_on` rules above apply unchanged: a group's list replaces the
  plugin's, and the dependencies it leaves out are named by `smith doctor`.

`smith convert` writes the settings that projects of one language and kind
share as a group. For a workspace that already repeats them, `smith doctor`
warns ("task settings shared by projects are written once") and `smith doctor
--fix` rewrites `smith.toml`: the repeated tables move into `[[task_groups]]`
entries added at the end of the file, a project that differs keeps its own
table (with the group's other fields added, so it runs as before), and every
other line, comments included, stays. The fix is only offered when the
rewritten file resolves to exactly the same tasks, and it never lets a group
reach a project that did not repeat the settings.

## A project's release settings

```toml
# keep existing tags that do not follow <name>@<version>
[projects.kit]
tag_prefix = "@acme/kit@"

# or name the prefix of every project once, and set tag_prefix only on the exceptions
[workspace]
scope = "@acme"
[release]
tag_format = "{scope}/{name}@"       # kit tags @acme/kit@1.0.0; {name} is required

# mirror each release to a repository of its own (see releasing.md, Split repositories)
[projects.client-php]
split_repo = "acme/client-php"

# make an app release: versioned by its tag, with per-platform binaries
[projects.cli.release]
asset_name = "{product}_{version}_{os}_{arch}.{ext}"   # the default
product = "acme"                                      # default: the project name
platforms = ["linux/amd64", "darwin/arm64", "windows/amd64"]  # default: six
os_alias = { darwin = "macos" }                       # names in file names only
arch_alias = { amd64 = "x64" }                        # the same for architectures
version_variable = "main.version"                     # Go: stamped with -X
package_managers = ["homebrew", "scoop", "winget", "aur", "nix"]    # manifests attached to the GitHub Release
linux_packages = ["deb", "rpm"]                       # Linux packages attached to the GitHub Release
npm = ["@acme/cli", "acme"]                           # npm packages; the first names the platform ones
pypi = "acme"                                         # PyPI wheels
nuget = "acme"                                        # .NET tool packages (dotnet tool install -g acme)
rubygems = "acme"                                     # platform gems
pub = "acme"                                          # pub.dev package, published by the pub workflow
pub_tag = "{name}-v{version}"                         # the default: the tag pub.dev's automated publishing matches
description = "What the app does"                     # shown by the registries and package managers
license = "MIT"
homepage = "https://cli.acme.dev"                     # default: the GitHub repository

# or make the release files with a command, in any language (see releasing.md, Release files from a command)
[projects.music-fixer.release]
build = "vcw app build examples/apps/music-fixer/app.json --out out --zip"  # run at the workspace root, once per platform
assets = ["out/*.zip", "out/*.zip.sha256"]            # globs of the files it leaves, uploaded to the GitHub Release
platforms = ["windows-x64", "macos-arm64", "macos-x64", "linux-x64", "linux-arm64"]  # a runner each
asset_name = "{product}-{version}-{platform}.{ext}"   # optional: rename the files on upload ({product}, {version}, {platform}, {ext})
checksums = true                                      # optional: attach <product>-<version>-<platform>.sha256 for what a runner uploads

# a Tauri app needs none of that: its plugin supplies build, assets, platforms, asset_name and checksums
[projects.studio.release]
```

See [Releasing](releasing.md#apps).

## Environment variables

| Variable | Effect |
|---|---|
| `SMITH_WORKSPACE_ROOT` | use this workspace instead of searching upward |
| `SMITH_CACHE_DIR` | cache directory, overriding `[cache] dir` |
| `SMITH_CACHE_REMOTE` | remote cache, overriding `[cache] remote` |
| `SMITH_CACHE_KEY` | the secret remote cache objects are signed with; without it the remote is off |
| `SMITH_CACHE_READ_ONLY` | read the remote cache without writing to it |
| `SMITH_CACHE_TOKEN` | bearer token for an `https://` remote cache |
| `SMITH_BASE` | base branch for `--affected` and `smith ci verify` |
| `SMITH_RELEASE_ANY_BRANCH` | allow a real release from another branch |
| `SMITH_PYPI_SIMPLE_URL` | the simple index to check before uploading to a custom Python index |
| `SMITH_PLUGIN_PATH` | directories to look for installed plugins in, after the workspace's `.smith/plugins` (docs/reference/plugins.md) |
| `SMITH_MARKETPLACE` | the plugin index `smith plugin search` and `smith plugin install <name>` read: an https URL or a file path; wins over `[plugins] marketplace` |
| `SMITH_HOME` | smith's own directory, `~/.smith` by default; installed plugins go in its `plugins/`, and the external plugins you trust are recorded in its `trust.json` |
| `SMITH_VERBOSE` | stream the output of the toolchain commands plugins run quietly while `smith add` creates a project (anything but empty or `0`; `smith add --verbose` sets it) |
| `SMITH_TASK`, `SMITH_TASK_KEY`, `SMITH_WORKSPACE_ROOT` | set by smith in the environment of every task it runs: the task id (`web:build`), its cache key and the workspace root |
| `SMITH_NO_UPDATE_CHECK` | stop `smith info` asking GitHub for the newest release |
| `SMITH_NPM_ALREADY_AUTHENTICATED` | tell the release's pre-flight check that npm is already authenticated, so the check does not require `NODE_AUTH_TOKEN` or `NPM_TOKEN` (set it when npm is logged in some other way) |
| `TWINE_REPOSITORY_URL` | upload Python packages and an app's wheels to another index |
| `NUGET_PUSH_URL` | the NuGet v2 push endpoint for an app's .NET tool packages (default nuget.org) |
| `RUBYGEMS_HOST` | the gem server for an app's platform gems (default rubygems.org) |
