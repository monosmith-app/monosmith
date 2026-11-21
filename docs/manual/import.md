# smith import

`smith import <git-url|folder>` brings a project that exists already into the smith
workspace you are in. The source is a git repository (smith clones it into a temporary
folder) or a folder on your computer. smith shows a plan first and changes nothing until
you say yes.

## When to use it

- You have a workspace, and you want to add a project that lives in another repository.
- You want to keep that project's git history in the workspace.
- To make a workspace from a repository, use [`smith adopt`](adopt.md) first. To make a
  new, empty project, use [`smith add`](add.md).

## What it does

1. **Fetch.** smith reads the folder as it is, or clones the git URL into a temporary
   folder. `--ref` takes a branch, tag or commit other than the default branch.
2. **Identify.** The language plugins find the project by its manifest at the top of the
   source: `package.json`, `pyproject.toml`, a `.csproj`, and every other manifest they
   know. A source with no manifest at its top is looked at further (see
   [A source without a manifest at its top](#a-source-without-a-manifest-at-its-top)).
3. **Place.** The project goes into `<kind folder>/<name>`: `apps/` for an app, `libs/`
   for a lib, `packages/` for a package. The name comes from the manifest, or from
   `--name`. In a terminal smith asks for the folder and suggests the kind the project
   looks like. Without a terminal, `--as` or `--into` must say it: smith never picks it
   for you. A folder that exists is refused. Right after the place step, smith makes sure
   git ignores the build output of the language's tools in the folder (`bin/` and `obj/`
   for .NET, `node_modules/` and `dist/` for Node, `__pycache__/` and `*.egg-info/` for
   Python). It adds a line to the workspace `.gitignore`, anchored to the folder
   (`/packages/astar/bin/`), for each pattern that git does not ignore yet. The plan shows
   these lines. So `--commit` never commits the output of the check step.
4. **Split.** The known development tools go to the workspace manifest. A tool that the
   workspace has at another major version (TypeScript 5 beside the workspace's TypeScript 4)
   stays in the project, and smith says so. smith asks about
   each development dependency it does not know (`--hoist <name>` sends one to the
   workspace without a question), and about a Node project's own ESLint setup
   (`--eslint-setup shared|keep`). Without a terminal the project keeps them, and smith
   says so. README and LICENSE stay with the project. The project's own lockfile
   (`package-lock.json`, `yarn.lock`, `pnpm-lock.yaml`, `uv.lock`) is removed, because
   the workspace has one lockfile for every project. Before that, smith keeps the
   versions it locked: the entries of an npm lockfile (`package-lock.json` or
   `npm-shrinkwrap.json`) go into the workspace `package-lock.json` under the project's
   folder, and smith names the direct dependencies it keeps. So the first install puts
   the same versions in place, not the newest version in each range. smith cannot carry
   the versions of `yarn.lock`, `pnpm-lock.yaml`, a bun lockfile or `uv.lock`. It says
   so in "Left to do", because the install can then move a package to a newer version.
5. **Join.** The project joins its language's workspace: the root `package.json`
   `"workspaces"`, the uv members of the root `pyproject.toml`, or the .NET solution
   (see [`smith doctor`](doctor.md)). A join that asks a question (which solution) is left
   to `smith doctor --fix`.
6. **Tooling.** A Node project keeps its own build and lint tools, or uses smith's. smith
   asks this before the plan, and recommends its own tooling when a tool is legacy (an
   ESLint 8 configuration, create-react-app, a `tsc` build into `lib/`, a bundler build).
   `--tooling smith|keep` answers. With `--yes` the project keeps its tools, and a line
   says how to choose smith's. Without a terminal and without either flag, smith refuses
   and names `--tooling`. With smith's tooling, smith's lint and build go in first, then
   each group of old tools is shown and removed. The step is the same as in
   [`smith adopt`](adopt.md#keep-the-projects-tools-or-use-smiths). It changes only the
   project's folder, the root `package.json` and the root `eslint.config.*`, so the other
   projects and the workspace's own lint files stay. It installs the whole workspace once,
   because smith's tools go into the root `package.json`. A project of another language
   keeps its tools.
7. **Check.** smith writes the pipelines again, runs `smith install` (the whole workspace,
   because the root manifest changed), then
   `smith run build test -p <name>`, then `smith run lint -p <name>` (without `--fix`), then
   `smith doctor`. When the lint fails, the "Left to do" list says what it found: the count of
   problems, the rules found most, and `smith lint --fix -p <name>` with the count it repairs.
   smith does not run the fix, because it changes your files. A file that breaks the
   vertical-slice layout gets three ways on: move it into a slice, write the exception in the
   project README, or add the project to `exclude` under `[conventions.vertical-slices]`. A task
   that stops the lint before it runs (a `deps-check`) is named with its cause.

It ends with `✓ imported <name> into <folder>`, then either "smith doctor finds nothing to
fix" or one "Left to do (N)" list: each item with the step that clears it.

## A source without a manifest at its top

**The project is in a sub-folder.** A .NET repository often keeps its library in
`src/<Name>/<Name>.csproj`, and a JavaScript repository its package in `packages/<name>`.
smith looks for manifests up to four folders below the top, without hidden folders,
installed dependencies and build output.

- One project: smith offers it. `--path <sub-folder>` says it without a question.
- Several projects: the source is a workspace. smith says so and offers one of them.
  `smith adopt`, run in a clone of the source, takes all of them.

A git source keeps the commits of the sub-folder: in a temporary clone,
`git subtree split --prefix=<sub-folder>` writes each commit that touches the folder again
with its files at the top, and `--history keep` or `squash` takes that branch.

A file the project uses outside its sub-folder does not come with it. Before the import,
smith reads the source and lists these files in the plan:

- a path that leaves the folder in a `.csproj`, `.fsproj`, `.vbproj`, `.props` or
  `.targets` file at its top: an item (`<None Include="..\LICENSE">`), an `<Import>`, or a
  path property (`AssemblyOriginatorKeyFile`, `ApplicationIcon`, `ApplicationManifest`,
  `CodeAnalysisRuleSet`, `Win32Resource`). A path with `$(`, `@(`, `%(` or a wildcard is
  not read, because only MSBuild knows its value.
- a `Directory.Build.props`, `Directory.Build.targets` or `Directory.Packages.props` in a
  folder above it, for a .NET project.
- a `"files"` entry with `../` in its `package.json`.

smith asks whether to copy each plain file into the project folder: the copy takes the
file's name, and the file that names it gets the new path (`..\LICENSE` becomes
`LICENSE`). `--outside-files copy|leave` answers. With `--yes` and without a terminal, the
files stay in the source. A file that stays, a missing file, a folder, a project reference
and an MSBuild file above the folder each get an item in "Left to do", with its path.

**A folder of sources and no manifest.** For loose `.ts` or `.js`, `.py` and `.cs` files
(Node, Python and .NET for now), smith offers to write the manifest of the language's default
stack: the stack in `[default_stacks]`, else the language plugin's own. The manifest comes
from the same templates as `smith add`, with the name of the folder and the kind you choose.
The sources stay where they are. `--manifest <language>` answers, `--manifest none` refuses,
and in a terminal smith asks. Without a terminal and without the flag smith refuses and
prints the command with the flag: it never writes a manifest you did not ask for.

The template may expect another layout. A Node template points `tsconfig.json` and
`package.json` at `src/`; when the folder has no `src/`, the "Left to do" list says so. A
Python manifest finds the packages at the top of a folder without `src/`. A `.csproj` takes
every `.cs` file in its folder.

## History

| `--history` | What happens | Commits git makes |
|---|---|---|
| `keep` (default for a git source) | every commit of the source is written again with its files under the project folder, then merged. `git log -- <folder>` and `git blame` follow them | 1 merge commit, and the source's commits |
| `squash` | `git subtree add --squash`: one commit holds the source, with its URL and commit in the message | 2 |
| `copy` (default, and the only mode, for a plain folder) | the files are copied, without `.git`, installed dependencies (`node_modules`, `.venv`) and caches | none |

`keep` and `squash` need a clean work tree, because git makes commits. smith refuses a
dirty tree and names the step: commit or stash, or use `--history copy`. The source
repository is never changed: smith rewrites the history in a temporary copy. When a local
source folder has changes that are not committed, the plan says that the import takes only
the committed state.

smith commits nothing else. `--commit` commits the rest of the import (the split, the join,
the pipelines) as one operation that `smith undo` reverts.

## Examples

```sh
# see the plan, change nothing
smith import https://github.com/owner/api --dry-run

# an app with its history, without questions
smith import https://github.com/owner/api --as app --yes

# a local Python library, copied, a test dependency sent to the workspace
smith import ../color-tools --as lib --hoist freezegun --yes

# a .NET library kept in src/MyLib, with the commits of that folder
smith import https://github.com/owner/dotnet-lib --path src/MyLib --as package --yes

# a folder of loose Python scripts, with a pyproject.toml written for it
smith import ../scripts --as app --manifest python --yes

# a .NET package as one squashed commit, and the rest committed too
smith import https://github.com/owner/codes --as package --history squash --commit --yes
```

## What it refuses

| Situation | Message | Step |
|---|---|---|
| Not in a workspace | this folder is not in a smith workspace | `smith adopt`, or cd into a workspace |
| The folder exists | `<folder>` exists already | `--into` another folder, or `--name` |
| No manifest and no source file smith knows | smith finds no project at the top of `<source>` | add the manifest first |
| Projects below the top, no terminal | the manifest of `<source>` is not at its top, or it holds N projects (a workspace) | `--path <sub-folder>`, or `smith adopt` |
| Sources without a manifest, no terminal | `<source>` has N source files and no manifest | `--manifest <language>`, or write the manifest |
| Uncommitted changes and a git history | the work tree has uncommitted changes | commit or stash, or `--history copy` |
| No terminal and no folder | smith does not choose the folder | `--as <kind>` or `--into <folder>` |

<!-- smith:begin import -->
Bring an existing project into this workspace, from a git repository or a folder

Usage: `smith import <git-url|folder> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `git-url\|folder` | yes |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--as <as>` | string |  what the project looks like, shown in the plan | the project's kind: app, lib or package (default: what the project looks like, shown in the plan) |
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `--dry-run` | boolean | off | show the plan and change nothing |
| `--eslint-setup <eslint-setup>` | string | without a terminal | what happens to a Node project's own ESLint setup: shared (the workspace's replaces it) or keep (default without a terminal) |
| `--history <history>` | string |  | keep (every commit, under the folder), squash (one commit) or copy (the files only); default keep for a git source, copy for a plain folder |
| `--hoist <hoist>` | string, repeatable |  | development dependencies to move to the workspace without asking (the known tools move anyway) |
| `--into <into>` | string |  <kind folder>/<name>, asked in a terminal | the project's folder, from the workspace root (default: <kind folder>/<name>, asked in a terminal) |
| `--manifest <manifest>` | string |  | a source with no manifest: the language whose default stack's manifest smith writes (node, python, dotnet), or none; asked in a terminal |
| `--name <name>` | string |  the name in its manifest | the project's name (default: the name in its manifest) |
| `--outside-files <outside-files>` | string | with --yes | with --path, the files the project uses outside its sub-folder (..\LICENSE, a signing key): copy (into the project folder, with the new path in the file that names it) or leave, and Left to do names each (default with --yes) |
| `--path <path>` | string |  | the sub-folder of the source that holds the project (src/MyLib); a git source keeps the commits of that folder |
| `--ref <ref>` | string |  its default branch | the branch, tag or commit of a git source to import (default: its default branch) |
| `--tooling <tooling>` | string, repeatable | with --yes | a Node project's build and lint tools: smith (smith's build and ESLint config, then the old tools go) or keep (default with --yes) |
| `-y, --yes` | boolean | off | do not ask: apply the plan (--as or --into must say where the project goes) |

<!-- smith:end import -->
