# ADR 0107: smith offers a manifest for a source without one

- Status: accepted
- Date: 2026-11-20
- Issue: #1154 (epic #1156, finding F7 and owner decisions D4 and D5 of
  [the review](../research/any-repo-into-a-workspace-review.md)). Amends
  [ADR 0106](0106-smith-import-brings-a-project-into-a-workspace.md).

## Context

smith knows a project by its manifest at the top of its folder. Two kinds of source had no
next step:

- A folder of sources with no manifest (loose `.ts`, `.py` or `.cs` files). `smith adopt` said
  "no project found" and `smith import` refused, with the list of the manifests smith knows.
- A source whose manifest is below its top: a .NET repository with `src/<Name>/<Name>.csproj`,
  or a JavaScript repository with its package in `packages/<name>`. `smith import` refused it,
  and the import of the sub-folder lost the history.

The owner decided (D4) that smith offers to write the manifest of the language's default stack
and keeps the sources where they are, and (D5) that the first release covers Node, Python and
.NET.

## Decision

1. The plugin protocol gets two fields and one capability.
   - `InitializeResult.sources`: patterns of source file names (`*.ts`, `*.py`, `*.cs`) the
     plugin recognises in a folder without its manifest. The host reads them with the manifest
     claims (`adoptmove.Claim.Sources`).
   - `GenerateParams.manifestOnly`: generate into a folder that exists and holds the person's
     sources. The plugin writes only the files at the top of the folder that are not sources and
     not documents (`templating.ManifestOnly` keeps them from the plugin's own templates): the
     manifest and its configuration (`package.json` and the tsconfig files, `pyproject.toml`, the
     `.csproj`). It writes no starter file and never writes over a file. It may add the workspace
     files its tooling needs, as generate does, except for a project at the root itself
     (`directory` is `.`). It installs nothing itself: `installPending` asks the host.
   - The capability `manifestOnly` (a modifier of `generate`, no method of its own). The host
     never sends `manifestOnly` to a plugin without it, because that plugin would write its
     starter files among the sources.
2. The Node, Python and .NET plugins name their sources and declare the capability, and so does
   the typescript plugin (the scaffold of the Node stacks). Python writes
   `[tool.setuptools.packages.find] where = ["."]` for a folder without `src/`, so the sources
   stay where they are. Other languages follow with the same two fields.
3. The manifest comes from the scaffold of the workspace's default stack for the language
   (`[default_stacks]`), else from the language plugin: the same templates as `smith add`.
4. Ask, never impose. `--manifest <language>` answers, `--manifest none` refuses. In a terminal
   smith asks, with the language that has the most source files as the default and "none" as an
   option. Without a terminal and without the flag smith refuses and prints the command with the
   flag. `--yes` does not answer the question.
5. `smith import`: a source with no manifest at its top is looked at in this order.
   - Projects below the top (up to four folders deep, without hidden folders, installed
     dependencies and build output). One project: smith offers it. Several: smith says the
     source is a workspace (`smith adopt` takes it whole) and offers one of them. `--path
     <sub-folder>` answers. A git source keeps the commits of that folder: in a temporary clone
     `git subtree split --prefix=<sub-folder>` writes each commit that touches it again with its
     files at the top, and `--history keep` or `squash` takes that branch.
   - Source files: the manifest offer. The folder and the kind are settled as before, the files
     are placed, then the plugin writes the manifest into the project folder. A Node template
     points at `src/`: when the folder has no `src/`, the import ends with that item in its
     "Left to do" list.
6. `smith adopt`: a root with sources and no manifest is reported with the offer. `--init` asks
   (or reads `--manifest`), then asks where the project goes (or reads `--move`). The plugin
   writes the manifest at the root, named after the destination folder and with its kind (apps/,
   libs/, packages/), `git add` stages it, and the move goes on as for a root project with a
   manifest. `git reset --hard` takes it back with the move.
7. The import installs the whole workspace (`smith install`), not only the new project: the split
   and a written manifest change the root manifest, and the install of one project does not
   install the root's tools.

## Consequences

- A folder of loose sources and a repository with its project in a sub-folder come into a
  workspace without hand edits, with the history of the sub-folder.
- The manifest is the template's. The sources may need more (a `src/` layout, references to the
  rest of the repository they came from, files outside the sub-folder such as a signing key):
  the import names what fails in its "Left to do" list.
- PHP, Java and the other languages need only `sources` and `manifestOnly` in their plugins.
