# smith affected

`smith affected` lists the projects affected by the changes since a base branch, one per
line. A project is affected when its own files changed or when a project it depends on is
affected; a change to a workspace-level file (the lockfile, `smith.toml`, a pipeline,
`.devcontainer`) affects every project, except for neutral files that never do:
`*.md`, `docs/**`, `LICENSE`, `.github/ISSUE_TEMPLATE/**`, `.vscode/**` and
`*.code-workspace`. When smith cannot be sure (the base cannot be resolved, a shallow
clone), it answers with every project and says why.

## When to use it

- To see what a pull request touches.
- To give a deployment platform the paths to watch (`--paths`, `--json`).
- The same selection is what `--affected` uses in [`smith run`](run.md) and
  [`smith graph`](graph.md).

## Synopsis

```sh
smith affected [--base <branch>] [--json] [--paths]
```

## Arguments and flags

<!-- smith:begin affected -->
List the projects affected by changes since a base branch

Usage: `smith affected [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--base <base>` | string | the pull request target, else the release branch | base branch (default: the PR target, else the release branch) |
| `--json` | boolean | off | print JSON, with the directories to watch for each app |
| `--paths` | boolean | off | print each app and the directories a deployment platform should watch |

<!-- smith:end affected -->

- `--base` is resolved against `origin` first (CI checkouts rarely have the local branch).
  The default is the pull request target from the environment (`SMITH_BASE`,
  `GITHUB_BASE_REF`, `SYSTEM_PULLREQUEST_TARGETBRANCH`, `CI_MERGE_REQUEST_TARGET_BRANCH_NAME`,
  in that order), else the release branch.
- `--json` prints `projects`, `all` (true when every project is affected), `reason`, `base`,
  `changed` (the files) and `watch`: for each app, the directories a deployment platform
  should watch, the app's own and those of everything it depends on, transitively.
- `--paths` prints just that: the app, a tab, then its directories, one app per line. Both
  list every app, affected or not.

## Exit codes

0 on success; 1 outside a workspace or when git cannot be read.

## Files and environment

Reads git history and the project graph. Writes nothing.

## Examples

After editing a file in `apps/web`, against `main`:

```sh
smith affected
```

```text
web
changes since main
```

When a workspace file changed, everything is affected and the reason is printed:

```text
api
tools
ui
web
workspace file .devcontainer/devcontainer.json changed
```

```sh
smith affected --paths
```

```text
api	apps/api
web	apps/web
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `... is not a revision` style errors from git | The base does not exist locally or on `origin`; fetch it, or give a reachable commit. |
| Every project listed with a reason about history | A shallow clone; fetch with full depth. |

## Related

[`smith run`](run.md) (`--affected`), [`smith graph`](graph.md) (`--affected`),
[`smith ci verify`](ci.md#verify), [Tasks, cache and affected](../reference/tasks-and-cache.md).
