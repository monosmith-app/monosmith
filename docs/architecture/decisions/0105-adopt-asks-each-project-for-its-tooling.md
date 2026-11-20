# ADR 0105: Adopt asks each project which build and lint tools it uses

- Status: accepted
- Date: 2026-11-20
- Amends: [ADR 0034](0034-adopt-and-convert-depth-levels.md)

## Context

`smith adopt` kept the build and lint setup of a plain repository (the conservative level), and
`--level full` was refused for it (issue 553). The removal plan of `smith convert --level full`
needs smith's build and lint in place before it removes the old tools, and only a conversion put
them there. A plain repository with a legacy tool (an ESLint 8 `.eslintrc`, create-react-app, a
`tsc` build into `lib/`) could not become green: doctor failed on the lint, and a create-react-app
project did not build on the current Node. The owner decided (decision D3 of the any-repository
review) that smith asks for each project, and recommends its own tooling when the project's tool
is legacy.

## Decision

1. For each Node project, `adopt --init` asks: keep its own build and lint tools, or use smith's.
   The recommended answer is smith's when a tool is legacy, the project's own otherwise.
   `--tooling smith|keep` answers for every project, `--tooling <project>=smith|keep` for one.
   With `--yes` an open project keeps its tools and a line says how to choose smith's. Without a
   terminal and without an answer, adopt refuses and names the flag.
2. `--level full` on a plain repository is `--tooling smith` for every project without an answer.
3. smith's tooling puts smith's tools in place first: the root `eslint.config.mjs` with
   `@monosmith/eslint-config`, an `eslint` that it runs with, a `lint` script; a `tsc` build into
   `dist` (doctor's own fixes); Vite and Vitest for create-react-app. Only then the removal plan
   of the full conversion (`legacyinventory`, `removalplan`, `fullconversion`) retires the old
   tools, one asked group at a time, and only for a tool whose replacement is in place.
4. The step is the slice `projecttooling`, with a small input (workspace root, project folder,
   answer), so `smith import` calls the same step for one project. Its removal plan sees the
   project's folder and the files at the workspace root only.
5. The closing check lints without `--fix`: smith reports what its rules find and does not
   rewrite code it just took in. What smith cannot replace is an item of adopt's `Left to do`.

## Consequences

- A plain Node repository can end with smith as its only build and lint tool, and a
  create-react-app project builds again.
- smith's lint rules are strict, so a project that takes smith's tooling usually has lint items
  left. They are named, with `smith lint --fix` as the first step.
- A rollup, webpack or script build stays; the project keeps those tools and the outcome says so.
- Python and .NET projects are not asked yet. Their families in `legacyinventory` and their
  replacement steps are follow-up work of issue 553.
- The terminal wizard of adopt still offers the full level only for Nx and Turborepo workspaces.
