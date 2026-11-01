# ADR 0050: Config that sibling projects share is derived from their shape

- Status: accepted
- Date: 2026-10-31

## Context

[ADR 0046](0046-derived-files-have-an-owner-and-a-record.md) added `smith derive` with two Node
renderers: the `references` of `tsconfig.lib.json` and the `displayName` of a Jest config. The
second one needs `smith:begin` markers that a converted workspace does not have. In a measured
converted Nx workspace with 22 packages (issue
[#848](https://github.com/monosmith-app/monosmith/issues/848)), derive covered 59 reference
entries, about 7 percent of 3,200 lines of per-package config. A `jest.config.cts` got nothing, and
files that are byte-identical in every package (`.spec.swcrc`, `tsconfig.json`) were not derived.
The help named a Jest `displayName` as derived, which was true only for a config with markers.

## Decision

A renderer reads the shape of the sibling projects. When at least three projects of one kind hold a
config file at the same relative path, and at least two thirds of the projects that have it hold it
in the same shape, that shape is the workspace's convention (`node.sibling-shape`).

- A file is "in the shape" when it equals the majority form after the project name, the package
  name and the project's internal references are substituted.
- Smith derives the file for each project that has it, or that the record owns and that lost it. A
  file that is in the shape is a whole file. A JSON file that matches in some top-level keys only is
  derived key by key.
- A file, or a key, that differs is the project's own. Smith never proposes it and never overwrites
  it. The renderer lists it as `diverging` in the derive answer; `smith derive --explain` and
  `smith doctor` say so. It is not drift and needs no action.
- Only config files are candidates: `tsconfig*.json`, `.swcrc`, `.babelrc` and the configs of Jest,
  Vitest, ESLint, Prettier and Babel. A manifest, a test setup file and a source file are not.
- When two renderers propose the same whole file with the same content, the first one keeps it.
  When the content differs and the record lists exactly one of them, that one wins. Otherwise the
  dispute is an error, as before.

The `DeriveResult` of the plugin protocol gains an optional `diverging` list. The change is
additive and the protocol stays at version 1.

Alternative not chosen: correct the help to say that only tsconfig references are derived. It is
honest, but it leaves most of the repeated config as hand-written lines, which is the problem that
derive exists to solve.

## Consequences

- A converted workspace that turns derive on can delete the repeated config and keep what differs.
- A threshold decides what counts as a convention. A workspace with fewer than three siblings
  derives nothing from this renderer.
- The shape is read from the files on disk, so a hand edit of a shared file in one project makes
  that project diverge instead of changing every project.
- Not done: a shared base file that the projects extend (the files stay one copy per project).

## Sources

- Issue [#848](https://github.com/monosmith-app/monosmith/issues/848), issue
  [#847](https://github.com/monosmith-app/monosmith/issues/847)
- [ADR 0046](0046-derived-files-have-an-owner-and-a-record.md)
- `libs/plugin-node/siblingfiles`, `apps/smith/derive`
