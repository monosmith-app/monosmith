# smith convert

`smith convert` turns an Nx or Turborepo workspace (or an mnci one) in the current
folder into a smith workspace. Projects keep their directories; Nx targets and
`turbo.json` tasks become plugin tasks, `package.json` scripts or `smith.toml` tasks;
release settings and existing tags carry over; the old tool is removed and the CI
pipelines are regenerated. Nothing is dropped silently: the report lists what maps,
what changes behaviour and what needs you.

There are two depths. `--level conservative` (the default) keeps what works and lists
the rest. `--level full` makes smith the only tool: after the conversion it works out
which files and devDependencies of the old tooling nothing uses any more (ESLint
configs and plugins, the test runner you did not choose, rollup or webpack, Nx, mnci)
and asks about each group before deleting it.

## When to use it

You can also start with `smith adopt`: in an Nx or Turborepo workspace it hands over to this
command, and the terminal interface offers one entry, "Bring this repository into MonoSmith".

- The folder is an Nx or Turborepo workspace and you want it to run on smith.
- A Turborepo workspace on pnpm or yarn: it becomes an npm workspace, the one
  [`smith install`](install.md) installs (`pnpm-workspace.yaml` moves into
  `package.json` `workspaces`, `workspace:` and `catalog:` ranges become what npm
  accepts, the pnpm and yarn lockfiles and settings go, and scripts that call
  `turbo`, `pnpm` or `yarn` call `smith` and `npm`).

Do not use it on a plain repository (use [`smith adopt`](adopt.md), which also hands an Nx or Turborepo workspace to convert) or on a folder that
already has a `smith.toml` (use [`smith upgrade`](upgrade.md)).

## Synopsis

```sh
smith convert [--dry-run] [--level conservative|full] [--lint smith|keep] [--test jest|vitest|keep] [--build smith|keep] [--ci smith|keep] [--yes] [--strict] [--allow-dirty] [--skip-install] [--skip-verify] [--json] [--verbose]
```

## Arguments and flags

<!-- smith:begin convert -->
Convert an Nx, Turborepo or mnci workspace to smith (adopt sends you here when it recognises one)

Usage: `smith convert [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--allow-dirty` | boolean | off | convert even with uncommitted changes (not with --level full) |
| `--build <build>` | `smith` or `keep` | `smith` | full level: build with smith's stack build (smith) or keep the current build setup (keep) |
| `--ci <ci>` | `smith` or `keep` | `smith` | full level: the pipeline smith generates (smith) or keep the current one (keep) |
| `--dry-run` | boolean | off | report the plan without changing anything |
| `--json` | boolean | off | print the plan as JSON |
| `--level <level>` | `conservative` or `full` | `conservative` | how far to go: conservative (keep what works, list the rest) or full (make smith the only tool, deleting the old tooling after asking) |
| `--lint <lint>` | `smith` or `keep` | `smith` when the ESLint config imports a shared config smith replaces, else `keep` | full level: lint with smith's ESLint config (smith) or keep the current setup (keep) |
| `--skip-install` | boolean | off | do not run npm install afterwards |
| `--skip-verify` | boolean | off | full level: install, but do not build, test, lint and run doctor afterwards |
| `--strict` | boolean | off | refuse to convert while anything needs manual work or is unsupported |
| `--test <test>` | `jest`, `vitest` or `keep` | the runner the workspace uses, else `keep` | full level: the test runner to keep, jest or vitest (the other's configs and packages go), or keep |
| `--verbose` | boolean | off | name every project a finding applies to, not just a few |
| `--yes` | boolean | off | apply without asking, and install an Nx workspace's packages first (a terminal asks before changing anything) |

<!-- smith:end convert -->

Interactions:

- Run it with `--dry-run` first. An Nx workspace's packages must be installed, because Nx
  itself describes the projects: in a terminal smith asks "Install them now?" and does it
  (about two minutes), `--yes` installs without asking, and elsewhere it stops and says so.
  A Turborepo workspace is read from its files, so nothing needs to be installed.
- `--lint`, `--test`, `--build` and `--ci` are the choices behind `--level full`. In a
  terminal the ones you do not give are asked; elsewhere the plan is printed with
  the default for each (`lint: keep   tests: keep   build: smith   CI: smith`) and the
  output ends with the exact command that applies them. Giving any of them with
  the conservative level is an error (`--lint only apply with --level full, ...`).
- `--build smith` (the default) also moves every project that compiles with `tsc` to the
  workspace's one `dist/<project dir>`: the `outDir` leaves its tsconfig, the `build` script
  passes `--outDir`, and `main`, `types` and the `start` script follow. It is asked as its own
  group (nothing is deleted, files are edited) and is skipped with `--build keep`.
- `--allow-dirty` converts with uncommitted changes. It is refused with
  `--level full`, which needs a clean git tree so that git can undo everything.
- `--strict` turns any item that needs manual work or is unsupported into a failure.
- `--yes` applies without asking; a terminal otherwise asks before changing anything.
- `--skip-install` skips the `npm install` afterwards. `--skip-verify` (full level)
  installs but does not build, test, lint and run `smith doctor` afterwards.
- `--json` prints the plan: `root`, `source`, `projects`, `findings` (each with a
  `severity`: `mapped`, `dropped`, `manual`), `rootScripts`, `removeDependencies`,
  `removeFields` and `delete`.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The plan was printed (`--dry-run`) or the conversion and its checks succeeded. |
| 1 | The folder is already a smith workspace, the tree is dirty, `--strict` found a blocker, a flag value is invalid, or a step of the full conversion failed. |

## Files and environment

Reads `nx.json` and `project.json` files (through Nx), or `turbo.json`, the root
`package.json` and the package manifests. Writes `smith.toml`, edits the root and
project `package.json` files, deletes the files the old tool needed, regenerates
`.github/workflows/ci.yml` (and the other providers' files), and at the full level
may delete configuration files after asking. The full report of an applied
conversion is saved to `.smith/convert-report.txt`. Runs `npm install` unless
`--skip-install`. No smith-specific environment variable is read.

### Shared task settings

When two or more projects of one language and kind (every Node library, say)
come out with the same settings for a task, `smith.toml` states them once, as a
`[[task_groups]]` entry that selects them (`select = ["lang:node", "kind:lib"]`),
instead of once per project. A project whose settings differ keeps its own
`[projects.<name>.tasks.<task>]`, completed with the group's other fields, so
every task still runs as before. The report lists each group it wrote. A group is only
written when no project it selects would start taking settings it did not have.
See [Task groups](../reference/configuration.md#task-groups). A workspace
converted earlier can fold its repeated settings with `smith doctor --fix`.

## Examples

The plan for a small Turborepo workspace, changing nothing:

```sh
smith convert --dry-run
```

```text
Conversion plan for Turborepo workspace "turbo-demo"

Ready to convert: 2 projects (1 app, 1 package). Nothing blocks the conversion.

2 things to check before you apply:
  1. web build: the build writes to dist inside the package, and smith doctor fails a build that does not write to the workspace's dist/apps/web/: point the tool there (tsconfig outDir, --outdir) and update main, types and exports, then set outputs to match
  2. ui build: the build writes to dist inside the package, and smith doctor fails a build that does not write to the workspace's dist/packages/ui/: point the tool there (tsconfig outDir, --outdir) and update main, types and exports, then set outputs to match

Converted, nothing to do (9):
  - [default_stacks]: node = "node-jest": Node projects that record no stack, and new ones, get ESLint and Jest tasks; a script a project already has wins
  - build: turbo.json task kept in smith.toml: outputs ["dist"]
      2 projects: ui, web
  - install: dependencies are installed with smith install (smith i), which runs npm install for the workspace's Node projects, so a second package manager is not needed
  - package.json: removes turbo
  - script build: smith run build --bail
  - scripts: package.json scripts stay scripts: smith's Node plugin runs each as a task (npm run <script>), so smith run build runs every package's build script
  - web test: turbo.json task kept in smith.toml: depends_on ["^build", "build"]
  - turbo.json: removed: its tasks are the package.json scripts, with their dependencies, inputs, outputs and cache settings in smith.toml

Dropped (Turborepo, pnpm and yarn only) (1):
  - packageManager: npm@10.0.0: smith installs with npm, so the field no longer pins anything

Removes (1 file or folder):
  - turbo.json

Nothing changed (dry run). To apply:
  → smith convert

Then:
  1. git diff: review what changed (git status lists the new files; commit them with the rest)
  2. smith doctor: check the workspace
  3. smith run build: check that everything still builds
  4. smith i: install dependencies this way from now on, not with pnpm or yarn
  5. commit
```

The full-level plan adds the choices and the steps it will run (the first part of the
report, identical to the above, is left out):

```sh
smith convert --dry-run --level full
```

```text
Full conversion: smith becomes the only tool
  lint: keep   tests: keep   build: smith   CI: smith

Nothing of the old tooling is left to remove.

● Install the packages and refresh package-lock.json, then check the workspace
  step 1: install the packages and refresh package-lock.json
          $ smith npm-lockfile
  step 2: build every project
          $ smith run build
  step 3: run every project's tests
          $ smith run test
  step 4: lint every project
          $ smith run lint
  step 5: check the workspace with doctor
          $ smith doctor

Nothing changed (dry run). To apply, with these choices:
  → smith convert --level full --lint keep --test keep --build smith --ci smith --yes
```

Strict mode on the same workspace, when something needs manual work, stops after the
report:

```sh
smith convert --strict --dry-run
```

```text
Conversion plan for Turborepo workspace "turbo-demo"

Cannot convert in strict mode: something needs manual work or is not supported.
...
smith: the conversion is blocked; see the report
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `this is already a smith workspace (smith.toml exists)` | A conversion already wrote `smith.toml` (a half-applied one included). `smith upgrade` brings a workspace up to date; to start over, undo the working tree with `git restore --staged --worktree . && git clean -fd`. |
| `--level "sideways" is not one of conservative, full` | Use one of the two levels. |
| `this folder is not a git repository: convert makes its changes through git` | Run `git init && git add -A && git commit -m "before smith"`, then convert again. |
| `the working tree has uncommitted changes, so the conversion would not be one reviewable change` | Commit or stash, or pass `--allow-dirty` (not with `--level full`, whose message says the tree must be clean). |
| `convert cancelled: nothing changed` | You answered anything but `y` to the confirmation prompt. |
| `--test "mocha" is not one of jest, vitest, keep` | `--test` takes `jest`, `vitest` or `keep`. |
| `smith: the full conversion stopped: ... step 4 of 7 failed (smith run build): exited with code 1` | A check after the conversion failed (in the trial, `tsc` was not installed in the demo workspace). The report ends with `Everything above is in the working tree and was clean before, so git undoes it: git restore --staged --worktree . && git clean -fd`. Fix the cause, or undo and retry. |
| `the build writes to dist inside the package ...` (a finding, not an error) | Point the tool at the workspace's `dist/<project dir>` and set `outputs` to match, or `smith doctor` will fail the build. |

## Related

[`smith adopt`](adopt.md), [`smith upgrade`](upgrade.md), [`smith doctor`](doctor.md),
[`smith install`](install.md), [Converting from Nx, Turborepo or mnci](../getting-started/converting.md).
