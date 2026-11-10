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
- Go: `go.mod`, `go.sum` and the code stay with the module. `.golangci.*` stays at the root, and the
  `go.work` that makes the module part of the workspace is written by adopt's workspace step.
- Python: `SortPython` reads the dev, test, tests and lint lists of `pyproject.toml` (the extras and
  the PEP 735 groups). Tools (pytest, ruff, mypy, black, isort, flake8, build, type stubs and the
  like) move to the root `requirements-dev.txt`, the workspace's toolchain, which smith creates
  with its own pins when the root has none (smith's pin wins for a tool it lists already). An
  unknown one is asked about, as for Node. The `[tool.ruff]` tables become a root `ruff.toml`.
  `[project]`, the runtime dependencies, the build system and the other `[tool.*]` tables (mypy,
  pytest, setuptools: their paths are relative to the project) stay. `--install` also creates
  the `.venv`. Poetry's dev group is not read.
- Rust: `SplitRust` writes a root `Cargo.toml` with `[workspace]` and the crate as its member,
  moves the crate's `Cargo.lock` to the root, and writes a root `.cargo/config.toml` that builds
  into `dist/cargo`. The crate's `Cargo.toml` stays whole: cargo scopes `[dependencies]` and
  `[dev-dependencies]` to the crate, so there is no dependency to move. rustfmt, clippy and the
  toolchain pin stay at the root.

## Consequences

- A Node project at the root is adopted in one run, `smith doctor` finds no failure or warning
  (with `--install` and the agent files), and the project builds and tests as before.
- The project's own ESLint setup is a question, never a silent replacement: use the shared config
  (old file kept as `.pre-smith`, ESLint packages removed from the project) or keep their own
  (moved into the project). `--eslint-setup shared|keep` answers; without a terminal or with
  `--yes` their own is kept and the output says so.
- The adopt wizard and the VS Code extension ask both questions after "where does it go?": the
  wizard reads the plan through its flow (`MovePlan`), the extension reads `unsure` and `lint` from
  `smith adopt --dry-run --json --move <folder>`. Each question puts the answer that changes the
  least first and passes the answer as `--hoist` and `--eslint-setup`.
- Every other language (.NET, JVM, Dart, Ruby, PHP, C++, Swift) still needs its rules; adopt
  does not move them at all yet (ADR 0070 detects Node, Go, Python and Rust only). The contract is
  in `adoptsplit`.
