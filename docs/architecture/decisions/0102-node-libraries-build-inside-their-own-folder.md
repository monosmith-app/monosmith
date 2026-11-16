# ADR 0102: Node libraries build inside their own folder

- Status: accepted
- Date: 2026-11-15
- Supersedes: [ADR 0091](0091-doctor-moves-imported-libraries-the-nx-way.md)
- Amends: [ADR 0087](0087-doctor-moves-a-build-into-dist-per-project-and-refuses-an-unsafe-move.md) (for Node libraries and packages)

## Context

smith built every project into the workspace's `dist/<project dir>`. For an app this is right: the
build is a deploy output, and the pack phase, Docker contexts and `dist/drop` read it from there.

For a Node library it caused one defect after another. A library is resolved by name through the
link `node_modules/<name>`, and its `package.json` `main`, `types` and `exports` cannot lead outside
the package. With the build in `dist/libs/models`, they either pointed at `../../dist/libs/models`,
which worked only by chance (a scoped name adds one folder level, and a rename to `@quotes/models`
broke every import with TS2307), or ADR 0091 mapped the name to the sources in tsconfig `paths` and
Jest. That covered tsc and Jest only: Vitest, Node at run time, a bin run by an npm script and the
TypeScript program of each consumer (TS6059, TS6307) still failed, so OpenCraw reverted it.

## Decision

1. An app, and a VS Code extension (it ships a `.vsix` packed like an app), keeps building into
   `dist/<project dir>`.
2. A Node library or package builds inside its own folder, `dist` by default, and its `main`,
   `types` and `exports` lead to `./dist/...`. Every tool resolves it through `package.json`.
3. Its build task caches that folder (`outputs = ["dist"]`, or the folder its tsconfig or build
   script names), and doctor's output check accepts it.
4. Doctor fails a library that builds outside its folder, and offers a fix that moves the build
   back into `./dist`: the tsconfig `outDir` (or the script's flag), `main`, `types`, `exports` and
   scripts that named the old folder, and an `outputs` override in smith.toml for it.
5. The grouped fix of ADR 0091 that moved libraries into `dist/` and mapped their names to sources
   is removed.
6. `smith add`, `smith adopt` and `smith convert` leave a library's build in its folder.
7. A release packs a package from its own folder when its `dist` exists, so a `dist/<project dir>`
   left from the old layout is never published.

## Consequences

- A library behaves like any npm package, for every tool and through any rename of its scope.
- `smith clean` and the cache still find the build: it is the task's declared output.
- Other languages keep their build folders: none of them resolves a library through a manifest
  path, so the old rule caused no defect there.
- A CommonJS Jest consumer of a library built as ES modules still needs a `moduleNameMapper` to the
  library's sources, as with any ESM package.
