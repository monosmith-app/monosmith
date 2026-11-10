# ADR 0091: Doctor moves imported libraries into dist/ the Nx way

- Status: accepted
- Date: 2026-11-10

## Context

[0087](0087-doctor-moves-a-build-into-dist-per-project-and-refuses-an-unsafe-move.md) refused to move a
library that other projects import by name, because its `main`, `types` and `exports` lead into the old build
folder, and a `../../dist/...` path is not found through the link in `node_modules`. Its hint
("pass --outDir ../../dist/... and point main, types and start there") was also wrong for a package with
`exports`: Node forbids `..` in an `exports` target, and `npm publish` cannot pack files outside the package.
A real workspace (22 Nx packages, `"files": ["dist"]`, `exports` into `./dist/...`) got 22 warnings with no fix.

## Decision

1. The build goes to `dist/<project dir>`. The fix rewrites the build script (`--outfile`, `--outdir`,
   `--declarationDir`, `--tsBuildInfoFile`, the folder of `smith declaration-specifiers`) and the `outDir`,
   `declarationDir` and `tsBuildInfoFile` of every `tsconfig*.json` of the package. When the script gives tsc a
   `--declarationDir` and its tsconfig has none, the tsconfig gets the same one, so a project that references
   this one finds the declarations where the script writes them. A script smith cannot read with certainty
   (a command other than esbuild, tsc and smith, or shell syntax next to the old folder) is not rewritten: the
   package is not offered the fix and the finding says why.
2. `package.json` keeps `main`, `module`, `types` and `exports` as `./dist/...`: that is the published form.
   `Stage` makes them relative to the folder it packs (`./dist/index.js` becomes `./index.js`) when
   `dist/<project dir>` has no `dist` folder of its own, and keeps the exclusions of the `files` list
   (`!**/*.tsbuildinfo`) with `**` as its first entry.
3. Importers in the repository read the source. `tsconfig.base.json` gets one `paths` entry per package name
   (`"@auto/atlassian": ["./packages/atlassian/src/index.ts"]`), from the esbuild entry or the source of the
   types file. A name already mapped elsewhere, a `baseUrl` other than the file's folder, `exports` subpaths
   other than `.`, or a missing `tsconfig.base.json` keep the package where it is, with the reason. Jest needs
   no edit with `@monosmith/jest-preset`, whose resolver reads the tsconfig `paths`; with another CommonJS preset
   the fix asks whether to append a `moduleNameMapper`.
4. It is one fix for the workspace, carried by the finding of every package it covers, so the host folds them
   into one. It asks first (move, or leave), shows every changed file, is named for `--only` by
   `build each library into the workspace's dist folder and resolve it by name to its sources`, and is
   idempotent: a package that builds into `dist/<project dir>` is not a candidate. The group `outputs = ["dist"]`
   of the build task in `smith.toml` goes with it when no covered package stays behind.
5. It covers a library that other projects import by name, or that has `exports`. Other libraries and apps keep
   the per-project move of 0087.

## Consequences

- A leftover `<project dir>/dist` is read by Jest before the sources (its default resolver goes first):
  the fix says to delete it and deletes nothing.
- `depends_on = ["^build"]` of test tasks stays; it is not needed any more and costs time.
- `CheckEntryFiles` judges `./dist/...` entries in `dist/<project dir>` when the build writes there.

## Where to check

`libs/plugin-node/manifestfixes/library_build_fix_mapper.go`, `library_script_move_algorithm.go`,
`library_config_move_algorithm.go`, `libs/plugin-node/pipeline/check_build_dir_use_case.go`,
`libs/plugin-node/releasing/stage_package_use_case.go`.
