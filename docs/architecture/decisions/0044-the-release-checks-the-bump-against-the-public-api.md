# ADR 0044: The release checks the bump against the public API

- Status: accepted
- Date: 2026-10-31

## Context

The bump of a release comes from commit types: `feat` is a minor, `fix` and `perf` are a patch,
a breaking marker is a major. A person can write `fix: tidy the exports` on a commit that deletes
an export. The release then ships a patch that breaks callers, and nobody sees it until a user
complains (epic [#819](https://github.com/monosmith-app/monosmith/issues/819), phase 5, issue
[#824](https://github.com/monosmith-app/monosmith/issues/824), part b).

The core of smith must stay free of any language. A plugin never calls another plugin. A check
that builds the package is too slow and too fragile for a pull request preview.

## Options considered

**Read `package.json` and TypeScript in the core.** Rejected: the core would learn module syntax.
It cannot do the same for Go and C#.

**Send the API through `Contribution.facts` (ADR 0042).** Rejected: facts describe the working
tree at one point. This check needs the same project at two revisions, the last tag and now.

**Run `tsc` or `api-extractor`.** Rejected: it needs a build and the dependencies installed, and
it fails on a clean checkout. A release preview runs on a clean checkout.

**Make the check part of `publish`.** Rejected: publish runs after the plan is shown and after
versions are written. The warning must come with the plan, in `--dry-run` and in the preview
comment.

## Decision

One optional method joins the protocol. The protocol stays at version 1.

- `api/surface`, capability `apiSurface`. Params: `project` and `ref` (a git revision; empty
  means the working tree). Result: `names`, a sorted list of strings, and `unreadable`, a reason
  when no list exists.

The host does the generic part. For each package in the release plan that has a last release, it
asks the project's plugin for the names at the tag of that release and at the working tree. It
compares the two lists as sets of strings. It reads no meaning from a name.

| Difference | Needs |
|---|---|
| a name is gone | major (a minor before 1.0, as `Version.Bump` does) |
| a name is new | minor |
| none | nothing |

When the bump the commits give is lower than the bump the difference needs, the plan item gets a
warning. The warning names the package, the commits' bump and version, the names, the bump needed
and the version it makes, and the commit marker to use (`feat!:`, a `BREAKING CHANGE:` footer or
`feat:`). The warning shows in `smith release` text output, in the markdown of `--format
markdown`, in the `--comment` preview and in the `apiWarnings` field of `--json`. The `ci release`
phase prints it too, as a CI warning.

A package that was never released, or a version that a `version(...)[x]` directive forced, is not
checked. A plugin without the capability, a plugin that fails, a tag that is not fetched and a
reply with `unreadable` all give no warning. The check is advice. It never fails because it could
not run.

**Setting.** `[release] api_check`:

| Value | Effect |
|---|---|
| `"warn"` (default) | print the warning and the bump to use; the release continues |
| `"error"` | the same warning, and the release stops before anything is published |
| `"off"` | the plugins are not asked |

**Node.** `libs/plugin-node/apisurface` reads the entry points of `package.json` (`exports`,
`types`, `typings`, `module`, `main`; a private package has none). For each entry point it takes
the first candidate file in the repository, `types` first. When the built file (`dist/index.d.ts`)
is not in the repository, it reads the source it comes from (`src/index.ts`). It does not build.
A scanner, not a compiler, reads the file: the top-level `export` declarations, `export { a as b }`,
`export * from` (followed to the files it names), `export * as ns`, `export default`,
`export =`, and the CommonJS forms `exports.x =`, `module.exports = { ... }` and
`Object.defineProperty(exports, "x")`. Git revisions are read with `git show <ref>:./<path>`, so
no checkout is needed.

**Not detected.** A change in a signature or a type with the same name. Wildcard subpath exports
(`"./*"`). A second declarator (`export const a = 1, b = 2`) and destructured exports. The
implicit exports of an ambient `.d.ts` file. A regular expression literal with an unbalanced
brace. A `export *` from another package counts as one name (`* from <package>`).

## Consequences

- The core gains one generic comparison and no language knowledge. Go (exported identifiers) and
  C# (public types and members) add `api/surface` in their plugins with no change to the host.
- Plugins that say nothing change nothing: the method and the capability are optional.
- Names, not signatures: the check finds the common mistake (a deleted export) and misses a
  changed signature. A plugin can refine this by putting the kind or the signature into the
  name; the host will then see a change as one removal and one addition.
- A rename counts as a removal and an addition, so it asks for a major. That is right for callers.
- Two calls per released package in every preview. For Node the cost is a few `git show` calls.
- `smith ci verify` does not run the check. It does not compute a release plan; the release
  preview of the pull request is where the plan, and so the check, is.

## Sources

- Issues [#819](https://github.com/monosmith-app/monosmith/issues/819) and
  [#824](https://github.com/monosmith-app/monosmith/issues/824)
- [ADR 0006](0006-conventional-commits-tag-only-releases.md): versions come from commits
- [ADR 0042](0042-plugins-publish-shared-facts-the-core-merges.md): shared facts, and why they do not fit here
- [The plugin protocol](../plugin-protocol.md)
