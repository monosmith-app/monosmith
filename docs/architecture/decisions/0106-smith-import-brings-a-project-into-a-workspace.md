# 0106. smith import brings a project into a workspace

- Status: accepted
- Date: 2026-11-20
- Issue: #1152 (epic #1156, finding F1 and decisions D1 and D2 of
  [the review](../research/any-repo-into-a-workspace-review.md))

## Context

A workspace had no way to take in a project that exists already. The only path was a copy by
hand: the git history was lost, the project was not joined to its language's workspace, and
nothing of adopt's split (the tools to the workspace, the nested lockfile) applied. The owner
chose the command `smith import <git-url|folder>` (D1), and chose to keep the history of a git
repository and copy a plain folder (D2).

## Decision

1. `smith import <git-url|folder>` runs in a smith workspace and refuses outside one, with
   `smith adopt` as the step. The steps are adopt's, in this order:
   - **fetch**: a local folder as it is, a git URL cloned into a temporary folder;
   - **identify**: `adoptmove.Detect` with the language plugins' manifest claims (#1150), so import
     knows every language the workspace knows. No manifest at the top of the source is refused
     with the list of manifests smith knows;
   - **place**: the folder is `<kind folder>/<name>`. `--as` and `--into` say it. In a terminal
     smith asks, with the kind the project looks like as the default. Without a terminal and
     without a flag, smith refuses: it never picks the folder. A folder that exists is refused;
   - **split**: `adoptsplit` with `InPlace`: the known tools go to the workspace manifest, an
     unknown development dependency and the project's ESLint setup are asked about (`--hoist`,
     `--eslint-setup`), and the project's own lockfile is dropped;
   - **join**: the `workspace/join` method (ADR 0105), with its fix applied when it is safe and
     asks nothing;
   - **check**: the pipelines are written again, `smith install -p <name>` runs in the process,
     then `smith run build test -p <name>` and `smith doctor`. The command ends with
     "✓ imported <name>" and one "Left to do (N)" list.
2. History (`--history`). `keep` is the default for a git source. smith clones the source as a
   bare repository, writes every commit again with its files under the project folder
   (`git fast-import`, authors, dates and messages kept), fetches that branch into the workspace
   and merges it with `--allow-unrelated-histories` as one merge commit. `squash` runs
   `git subtree add --squash`. `copy` copies the files: it is the default, and the only mode, for a
   folder that is not a git repository. A copy leaves out `.git`, installed dependencies and caches.
3. `keep` and `squash` make commits, so they need a clean work tree and a confirmation (`y` in a
   terminal, or `--yes`). The plan says how many commits git makes. smith commits nothing else
   unless `--commit`.
4. The tooling step (D3, #553) is a seam: the project keeps its own tools, and the output says that
   a later version asks.

## Why not git subtree add for keep

D2 named `git subtree add --prefix`. It keeps the commits, but their trees hold the files at the
root of the source, not under the folder. So `git log -- <folder>` shows only the merge commit, and
`git blame` stops there. That defeats the reason for keeping the history. Writing the commits again
under the folder gives the same merge, and `git log -- <folder>` and `git blame` follow every
commit. The hashes of the imported commits change, which is also true of `git subtree add
--squash`. The source repository is never written: the rewrite runs in a bare clone.

## Consequences

- `smith doctor --fix --yes` says what a safe fix is: one that loses nothing of the person's. A join
  fix edits the person's package.json, pyproject.toml or solution, so "only restore files smith
  owns" was not true.
- A source whose manifest is not at its top (a .NET repository with `src/<Name>/<Name>.csproj`)
  is refused. The person imports the sub-folder, which copies the files without the history.
  #1154 offers a manifest for sources without one.
- The project's build and tests decide the end state. A legacy tool (create-react-app in a hoisted
  npm workspace) fails there and is named in the list, until the tooling step (#553) offers smith's
  tooling.
