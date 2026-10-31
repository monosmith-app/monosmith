# ADR 0049: A scaffold is also the renderer of its own files

- Status: accepted
- Date: 2026-10-31

## Context

[ADR 0046](0046-derived-files-have-an-owner-and-a-record.md) added `smith derive`. Its two Node
renderers target the shape of a converted Nx workspace: a `tsconfig.lib.json` with `references`,
and a `jest.config.*` with `smith:begin` markers. A workspace made by `smith new` and `smith add`
has neither file, so `smith derive` derived nothing there (issue
[#840](https://github.com/monosmith-app/monosmith/issues/840)). A `keys` file that a person deleted
also came back as a stub that held only the owned key.

Smith writes a repeated file in every smith-made project: `tsconfig.json`. Its content follows
from data that smith holds: the directory, the stack, the choices ([ADR
0045](0045-frameworks-tests-and-styles-are-choices-shipped-as-data.md)) and the facts.

## Decision

Choose between two ways to make `derive` useful here:

- (a) The scaffold writes files with the markers and keys that a derive renderer owns.
- (b) A renderer learns the shape that the scaffold writes.

**We choose (b), with one change: the renderer is the scaffold's own template.**

A plugin that scaffolds a file declares a `scaffolding.FileRenderer` for it. The renderer builds
the template data from the project (kind, directory, stack, shape, facts and the saved choices),
renders the same template that `smith add` used, and returns the whole file in the `file` mode.
The typescript plugin renders `tsconfig.json` of libs, packages and plain apps
(`typescript.tsconfig`). The vite plugin renders `tsconfig.json` of the apps it made from choices
(`vite.tsconfig`).

Reasons:

- A deleted file is restored whole and as the scaffold wrote it, because the template is the
  specification. In `keys` mode the renderer owns only some keys, so it cannot rebuild the file.
- (a) needs the scaffold to know that derive is on. That is a new field in the generate request,
  and the generated types of four SDKs change. It also adds markers that no tool needs when derive
  is off.
- With (b), `smith add` writes the same bytes as before. Derive stays opt-in through `[derive]
  enabled`.

**Which files are derived.** A renderer proposes a file only when the record owns it, or when the
file is, line for line, what the template renders. A file that differs and is not in the record
was written by a person: it is not proposed. So a converted workspace that turns derive on does not
get a list of diverged files.

**Record from day one.** When `[derive] enabled` is on, `smith add` records the new files in
`smith.lock` (`derive.Sync`). `Sync` writes and records only what needs no decision. A file that
needs a choice stays for `smith derive` to report.

**Whole-file ownership.** The renderers use the `file` mode. A hand edit of `tsconfig.json` is
drift. The choices are `--replace`, or `--eject` to own the file. `--adopt` applies to `keys`
files only. A person who wants to change the shape of every `tsconfig.json` changes the stack or
its choices, not the file.

## Consequences

- A smith-made Node workspace gets derived files without a hand-written marker.
- A change to a template needs a new renderer `Version`, so unedited files are rewritten.
- The `keys` stub is not changed. A deleted `tsconfig.lib.json` of a converted workspace still comes
  back with the owned `references` only. The guide says so. Only a renderer that can rebuild the
  whole file may use the `file` mode.
- Not done: `tsconfig.build.json` (it follows the test file suffix of the workspace, which can
  change after the project exists), the `jest` block of `package.json`, and `vite.config.ts`.
  Each needs the same step: declare a `FileRenderer` for it.

## Sources

- Issue [#840](https://github.com/monosmith-app/monosmith/issues/840), epic
  [#819](https://github.com/monosmith-app/monosmith/issues/819)
- [ADR 0045](0045-frameworks-tests-and-styles-are-choices-shipped-as-data.md), [ADR
  0046](0046-derived-files-have-an-owner-and-a-record.md)
- `libs/plugin-node/scaffolding`, `libs/plugin-node-toolchains/typescript`,
  `libs/plugin-node-toolchains/vite`, `apps/smith/derive`
