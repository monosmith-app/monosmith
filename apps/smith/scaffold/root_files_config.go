package scaffold

import "github.com/monosmith-app/monosmith/apps/smith/workspace"

const configTemplate = `#:schema ` + workspace.SchemaURL + `
# smith workspace configuration. Every setting below has a default; most
# workspaces need nothing more than a name. See https://github.com/monosmith-app/monosmith/blob/main/docs/reference/configuration.md.
schema = 1

[workspace]
name = %q
# npm scope for Node packages
scope = %q
# Go module path prefix for Go projects
module_prefix = %q

[ci]
# Pipelines to generate: "github", "azure", "gitlab", or several.
providers = [%s]
# A dev container (.devcontainer/devcontainer.json) with the toolchains CI
# installs, at the same versions.
devcontainer = true

[dependencies]
# One version per production dependency across the workspace, and dev tools at
# the root, once: smith doctor reports a dependency two projects declare at
# different versions, and smith deps up --align brings it to one.
one_version = true
# Dependencies that may differ between projects: allow = ["left-pad"]

# A stack is the set of plugins a project is built with: its language plugin
# and the toolchain, structure, scaffold and deploy plugins on top. Each
# project records its own, so one monorepo can mix them; ` + "`smith stacks`" + ` lists
# them and ` + "`smith add app web --stack node-vitest`" + ` picks one. Projects that record
# none use the default of their language. See https://github.com/monosmith-app/monosmith/blob/main/docs/reference/stacks.md.
[default_stacks]
node = "node-jest"     # ESLint with @monosmith/eslint-config, and Jest

# [stacks.node-quick]
# description = "Vitest only, no lint"
# language = "node"
# toolchain = ["vitest"]

# Conventions apply a plugin to every project of the languages it supports.
# vertical-slices gives Go and Python projects a slice-check task, run before
# lint: role-named files, no junk-drawer folders. See https://github.com/monosmith-app/monosmith/blob/main/docs/reference/conventions.md.
[conventions.vertical-slices]

# [release]
# branch = "main"     # releases run only from this branch
# remote = "origin"   # release tags are pushed here

# [cache]
# dir = ".smith/cache"
# max_size_mb = 2048

# Override any task a plugin infers, for every project or for one:
# [tasks.test]
# cache = false
# [projects.web.tasks.build]
# depends_on = ["^build", "codegen"]
`

const gitignore = `# smith
.smith/

# build output
dist/
coverage/

# Node
node_modules/

# Python
.venv/
__pycache__/
*.egg-info/
.pytest_cache/
.ruff_cache/

# editors and OS
.DS_Store
.idea/
`

const gitattributes = `* text=auto eol=lf
*.png binary
*.jpg binary
*.zip binary
`

const readme = `# %s

A monorepo managed by [smith](https://github.com/monosmith-app/monosmith).

## Everyday commands

` + "```sh" + `
smith add app web --lang node        # an app: deployed, never published
smith add lib core --lang go         # an internal library: imported, never published
smith add package sdk --lang python  # a package: published on release

smith build | test | lint            # every project, in dependency order, cached
smith test --affected                # only what changed since main
smith graph                          # how projects depend on each other
smith release --dry-run              # what the next release would publish
smith doctor                         # check the workspace is healthy
` + "```" + `

Commits follow [Conventional Commits](https://www.conventionalcommits.org): ` + "`feat:`" + ` releases a
minor version, ` + "`fix:`" + ` a patch, and a breaking change a major one.
`
