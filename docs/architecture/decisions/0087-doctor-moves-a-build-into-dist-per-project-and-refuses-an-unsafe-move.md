# ADR 0087: Doctor moves a build into dist/ per project, and refuses a move that is not safe

- Status: accepted
- Date: 2026-11-05

## Context

smith builds every project into `dist/<project dir>`. A project that `smith convert` or `smith adopt` keeps
with its own build scripts (an Nx package that builds into `packages/x/dist`) builds elsewhere, and the
sibling-shape rule makes a new project copy it. The Node doctor check `compiles into dist/<project dir>` saw
only `--outDir` and the tsconfig `outDir`, and its fix was one folded plan for every project. A real run
on a converted workspace showed two further facts:

- the build task can cache `dist/` of the package (a kept Turborepo `outputs`) or the central folder, while the
  script writes the other one, so a cache hit restores nothing;
- a package that other projects import by name cannot move. Node and TypeScript read its `main` and `types`
  through the link in `node_modules`, where `../../dist/packages/x` (the form a staged package needs) resolves
  from the link and is not found. The importers stop compiling.

## Options considered

- One folded fix for all projects: not chosen. The owner called the package.json rewriting complicated and wants
  a preview and an answer per project.
- Rewrite `smith.toml` in the fix of each project: not chosen. Whole-file edits of several projects undo each
  other, and the host marks the later ones stale.
- Move every project and let the build break: not chosen.

## Decision

1. Each project that builds outside `dist/` is its own finding, with its own fix that asks: move the build, or
   leave the project as it is. The review shows every changed line before anything is written. `--yes` alone
   never applies it. `smith doctor --fix --commit` commits it as one operation, and `smith undo` reverts it.
2. The move rewrites, in place and keeping other lines: the build script flags (`--outDir`, `--outdir`,
   `--out-dir`, `--outfile`, `--declarationDir`), the tsconfig options (`outDir`, `declarationDir`,
   `tsBuildInfoFile`, `outFile`), `main`, `types`, `typings`, `module`, `browser`, `bin` and every `exports` target,
   scripts that name the old folder as a path, and a `files` list that names only the old folder. The outputs
   override of the build task in `smith.toml` goes through `smith config unset`, a step, not a rewrite.
3. The move is refused, with the reason in the finding, when it is not safe: another project imports the project
   by name and its entry points lead into the old folder; a published package needs a file that the staged build
   folder does not hold (a `files` entry, a `bin` source); the build writes a second folder outside the first; or
   `smith.toml` gives the task an output that is not the build folder.
4. Only Node is checked. For the other languages smith passes the output folder to the tool, so the task writes
   to `dist/` whatever the project says. The tool's own settings (`target-dir`, `ArtifactsPath`, Maven
   `<directory>`, Gradle `buildDirectory`, CMake `-B`) are not read.

## Consequences

- A converted workspace gets a per-project path to the central `dist/`; the projects that cannot move stay
  red with a reason, not silently.
- Moving the build of a package that others import needs the importers to read its sources: a later decision.
- `smith convert --level full` still moved every tsc project without the importer check; [0088](0088-doctor-checks-and-moves-the-build-folder-of-every-language-it-can-read.md) gives it the same refusal and checks the other languages.

## Where to check

`libs/plugin-node/manifestfixes/migrate_build_fix_mapper.go`, `libs/plugin-node/pipeline/check_build_dir_use_case.go`,
`apps/smith/configedit/set_config_value_use_case.go`, `docs/manual/doctor.md`.
