# ADR 0090: Adopt splits the workspace from the project

- Status: accepted
- Date: 2026-11-10
- Amends: [0070](0070-adopt-asks-where-to-move-the-project-at-the-root.md)

## Context

ADR 0070 moved every top-level entry of a single-project repository into the project folder,
except a fixed list. A real repository showed the cost: the `.code-workspace` file moved into
`apps/<name>`, the project's `package.json` kept every tool in `devDependencies`, the root had
no ESLint config, and `smith doctor` failed on the result. The person had to repair what
`smith adopt` had written.

## Decision

- Each language declares `adoptsplit.Rules`: the top-level names of tool configuration that stay at
  the root (`Workspace`), the lockfiles that stay (`Lockfiles`), a `Sort` that reads the manifest
  before the move and sorts the dependencies, and a `Split` that edits the files after it. A
  language that declares none moves everything, as before. The rules are data in the core, not a
  plugin protocol method: that keeps the 24 plugins unchanged for now, and a protocol step can
  replace the table later without changing the callers.
- `adoptmove.Classify` asks the rules. The repository's own entries and every `*.code-workspace`
  file stay at the root. Documents (`*.md`, `*.rst`, `*.adoc`, `documentation`, `wiki`) go to the
  workspace's `doc/` folder; README, LICENSE and the other repository documents keep their names
  at the root, where hosts and licence tools look for them, and `docs/` stays where it is.
- Tools move to the workspace without a question. A development dependency smith does not know is
  asked about, with two answers (workspace, or project; the project is the default). `--hoist
  <name>` answers for it, and without a terminal or with `--yes` the dependency stays in the
  project and the output says so. `dependencies`, the ones the code runs on, always stay.
- Node: `Sort` names the tools (typescript, jest, prettier, `@types/*` and the like), the ESLint
  stack, and the unknown. `Split` moves the tools to a root `package.json` (written when the root
  has none, with `workspaces`), writes the compiler options every project shares to
  `tsconfig.base.json` and makes the project's `tsconfig.json` extend it (`outDir`, `rootDir`,
  `paths`, `include` and the other project paths stay), and points the build at `dist/<project>`
  with the same edit as `smith doctor --fix`. The legacy `.eslintrc*` is kept as `.pre-smith`, the
  root gets the shared flat config, and the ESLint packages leave the project, because the shared
  config brings its own ESLint and a second set would not install beside it. `jest.config.*`
  stays with the project, since its paths are relative to the project folder. The lockfile stays
  at the root and `npm install` rewrites it.
- A `tsconfig.json` that already extends another file, that has a trailing comma, or beside a root
  `tsconfig.base.json` is left whole, and the output says why.
- Go, Python and Rust declare the names of their tool configuration only (golangci, ruff and
  flake8, rustfmt, clippy and the toolchain pin). Splitting `go.mod` into a `go.work`, the
  `[tool.*]` tables of `pyproject.toml`, and a `[workspace]` `Cargo.toml` are not done.

## Consequences

- A Node project at the root is adopted in one run, `smith doctor` finds no failure or warning
  (with `--install` and the agent files), and the project builds and tests as before.
- Replacing the person's ESLint config is a decision made for them (kept as `.pre-smith`, listed in
  the output). A question with two answers (keep mine, use the shared one) is the next step.
- The adopt wizard and the VS Code extension do not ask about unknown development dependencies
  yet: they pass `--move` and the dependencies stay in the project.
- Every other language still needs its `Sort` and `Split`; the contract is in `adoptsplit`.
