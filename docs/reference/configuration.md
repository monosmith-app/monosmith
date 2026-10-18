# Configuration: `smith.toml`

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

[ci]
providers = ["github"]                # "github", "azure", "gitlab", or several
devcontainer = true                   # write .devcontainer/devcontainer.json with CI's toolchains (docs/reference/ci.md)
native_runners = ["ubuntu-latest", "windows-latest", "macos-latest"]  # where native projects build (docs/reference/ci.md)
native_linux_packages = []            # apt packages the native job installs on Linux

[release]
branch = "main"                       # real releases run only here
remote = "origin"                     # release tags are pushed here
github_releases = false               # a GitHub Release per tag (needs GITHUB_TOKEN)

[cache]
dir = ".smith/cache"                   # workspace-relative, or absolute
max_size_mb = 2048
remote = "s3://acme-smith-cache/main"  # a shared remote cache: https://, s3://, azblob:// (docs/reference/tasks-and-cache.md)
remote_read_only = false               # read the remote without writing to it

[plugins]
enabled = ["node", "go", "python"]    # default: every built-in plugin

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

[conventions.vertical-slices]         # apply a convention (docs/reference/conventions.md)
languages = ["go"]                    # default: every language it supports
exclude = ["legacy"]                  # projects it leaves alone

[conventions.docker]                  # an image task for every app with a Dockerfile
registry = "ghcr.io/acme"             # where release pushes the images; unset: built, never pushed
push = true                           # default: true when a registry is set
verify = true                         # smith ci verify builds the affected apps' images; default: true

[projects.api.docker]
image = "acme-api"                    # the image's name under the registry; default: the project's
push = false                          # build this app's image, never push it
context = "apps/api"                  # the build context; default: ".", the workspace
verify = false                        # this app's image is not built on pull requests

[plugins.external.kiwi]
version = "^1"                        # installed on the plugin search path (docs/reference/plugins.md)

[plugins.external.rust]
command = ["smith-plugin-rust"]        # any executable speaking the protocol
```

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
`depends_on`, `env`, `cache`.

There is exactly one precedence rule: **plugin → `[tasks.<task>]` →
`[projects.<name>.tasks.<task>]`**, later wins, field by field. To see the
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

## A project's release settings

```toml
# keep existing tags that do not follow <name>@<version>
[projects.kit]
tag_prefix = "@acme/kit@"

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
package_managers = ["homebrew", "scoop", "winget"]    # manifests attached to the GitHub Release
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
| `SMITH_HOME` | smith's own directory, `~/.smith` by default; installed plugins go in its `plugins/` |
| `SMITH_NO_UPDATE_CHECK` | stop `smith info` asking GitHub for the newest release |
| `TWINE_REPOSITORY_URL` | upload Python packages and an app's wheels to another index |
| `NUGET_PUSH_URL` | the NuGet v2 push endpoint for an app's .NET tool packages (default nuget.org) |
| `RUBYGEMS_HOST` | the gem server for an app's platform gems (default rubygems.org) |
