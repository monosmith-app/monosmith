<!-- lore-master
parent: deps.md
-->
# smith deps repair

`smith deps repair` repairs the installed dependencies of the workspace when they are broken.
You pick one of four repairs. smith prints what it deletes, runs and installs, tool by tool, and
asks before it starts. Each language plugin says how its tool does each repair
([ADR 0093](../architecture/decisions/0093-plugins-describe-dependency-repairs-and-smith-carries-them-out.md)).

## When to use it

- A tool reports a package that is installed but not declared, or a broken install tree.
- An install stopped half-way, or a package's files are damaged.
- You want the newest versions the manifests allow, with a new lockfile.

Use the lightest repair that can clear the problem:

| Kind | What it does | Versions |
|---|---|---|
| `prune` | Removes the installed packages that no manifest declares. | stay |
| `reinstall` | Deletes the installed dependencies, then installs them again from the lockfile. | stay |
| `lockfile` | Deletes the installed dependencies and the lockfile, then installs. | can change |
| `cache` | Clears the download cache of the tool, then installs. The cache is shared by every workspace on the machine. | stay |

A tool that has no honest equivalent of a kind does not offer it. A workspace without a lockfile
has no `lockfile` repair.

## Synopsis

```sh
smith deps repair [--kind prune|reinstall|lockfile|cache] [-p <project>...] [--dry-run] [--yes] [--json]
```

## Arguments and flags

<!-- smith:begin deps repair -->
Repair broken installed dependencies: prune, clean reinstall, rebuild the lockfile or clear the download cache

Usage: `smith deps repair [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--dry-run` | boolean | off | print the plan of each repair and change nothing |
| `--json` | boolean | off | print the plans, or the repair that ran, as JSON (needs --dry-run, or --kind with --yes) |
| `--kind <kind>` | string |  | the repair to run: prune\|reinstall\|lockfile\|cache (without it a terminal asks) |
| `-p, --project <project>` | string, repeatable |  | repair only these projects (a tool that installs the workspace as one still deletes its shared folders) (also -w, --workspace, --projects) |
| `-y, --yes` | boolean | off | do not ask before the repair (needs --kind without a terminal) |

<!-- smith:end deps repair -->

- Without `--kind`, a terminal shows the four repairs, one line each, with what each deletes.
  Type the number. A repair that no tool of the workspace has is marked and cannot be picked.
- smith always prints the plan before it asks `Repair now? [y/N]`. Only `y` goes on.
- `--dry-run` prints the plan of each repair (or of `--kind`) and changes nothing.
- Without a terminal (a script, CI) smith refuses unless `--kind` and `--yes` are given.
- `--json` needs `--dry-run`, or `--kind` with `--yes`. The progress then goes to stderr.
- `-p` limits the repair to some projects. A tool that installs the workspace as one tree (npm)
  still deletes the folders the projects share, such as the root `node_modules`. The plan shows
  them.

## Safety

- smith deletes a link as a link. A symbolic link or a Windows junction inside a deleted folder
  is removed itself, and the folder it points to keeps every file. npm links each workspace
  package into `node_modules`, so this keeps your source safe.
- smith refuses a path that is not inside the workspace, `.git`, `.smith`, `smith.toml`, the root
  of a project or a folder that holds one, and a path below a link. It checks each path again
  just before it deletes it.
- A `cache` repair says that other workspaces on the machine use the same cache.
- A `lockfile` repair says that versions can change. smith never commits: review the lockfile
  diff and commit it yourself.

## What each language does

| Language | prune | reinstall | lockfile | cache |
|---|---|---|---|---|
| Node (npm) | `npm prune` | delete each `node_modules`, then `npm install` | delete each `node_modules` and `package-lock.json`, then `npm install` | `npm cache clean --force`, then `npm install` |

The other languages follow. Until their plugin offers a repair, smith says that the tool has no
dependency repair, and the language's page in the plugin catalogue has the manual steps.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The repair ran, or the dry run printed the plans. |
| 1 | Not in a workspace, no tool offers a repair, the question was answered no or could not be asked, or a step failed. |

## Files and environment

Deletes only the paths the plan shows. The commands of the plan and the install write what the
tool writes (for Node: `node_modules`, `package-lock.json`, the npm cache).

## Examples

See every plan without a change:

```sh
smith deps repair --dry-run
```

Rebuild the lockfile in a script:

```sh
smith deps repair --kind lockfile --yes
```

Without a terminal and without `--yes`:

```text
× refusing to repair dependencies without a terminal to ask in
  → run `smith deps repair --kind <prune|reinstall|lockfile|cache> --yes`, or add --dry-run to see the plan
```

## Related

[`smith install`](install.md), [`smith deps`](deps.md), [`smith cache clear`](cache-clear.md)
(the task cache, not the package manager's), the Deps tab of the [TUI](tui.md).
