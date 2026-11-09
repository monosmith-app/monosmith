# ADR 0088: Doctor checks and moves the build folder of every language where it can read the setting

- Status: accepted
- Date: 2026-11-05
- Amends: [0087](0087-doctor-moves-a-build-into-dist-per-project-and-refuses-an-unsafe-move.md)

## Context

ADR 0087 checked only Node. A project that smith made builds into the central `dist/` in every
language, because smith writes the setting (Rust `.cargo/config.toml` `target-dir`, .NET
`Directory.Build.props` `ArtifactsPath`, Maven `<build><directory>`, Gradle `layout.buildDirectory`,
CMake `-B`) or passes the output option (`go build -o`, `python -m build --outdir`, `dart compile -o`).
An adopted, converted or hand-made project keeps the folder its tool picks: `target/`, `bin/` and
`obj/`, `build/`. Nothing told the owner, and Maven showed a real fault: its build task caches
`dist/maven/<artifactId>` while the pom writes `target/`, so a cache hit restores nothing.

## Options considered

- Check every language by running its tool (`cargo metadata`, `dotnet msbuild -getProperty`): not
  chosen. It needs the toolchain and a restore, and doctor must work without them.
- Read only the setting smith writes, as text: chosen. Each check says what it reads, and reads
  nothing it cannot read reliably.
- Rewrite the command line of a hand-made task: not chosen. A command is the owner's; the finding
  names the option to pass.

## Decision

1. Rust, .NET, Maven, Gradle, CMake, Go, Python and Dart each report where a project builds and where
   it should, from the setting that can be read as text (see the manual for the exact list). A setting
   in a form that cannot be read reliably (a property, an include, an inline table) is not checked and
   not guessed.
2. Rust, .NET, Maven, Gradle and CMake offer the 0087 fix: a question (move, or leave it as it is),
   a review of every changed line, no change on `--yes` alone, `--commit` and `smith undo`. The fix
   edits only the file that holds the setting and keeps its other content. The old folder is not
   deleted. Go, Python and Dart report the option to pass; the owner edits `smith.toml`.
3. The move is refused, with the reason and the manual step, when a file names the old folder
   (Dockerfile, script, pipeline, `smith.toml`), another project reads the build by its path
   (.NET `HintPath`, a pom, a script, a CMake file), the project file names the folder itself, a
   .NET project sets `OutputPath`, `CARGO_TARGET_DIR` is set, or the file to edit cannot be edited
   without guessing.
4. A refused move is a warning, not a failure: the build still works. It is a failure only when
   something is broken now (the build task caches a folder the build never writes).
5. The answer "leave it as it is" ends with "Left as it is.": no empty review, no question.
6. `smith convert --level full` uses the refusal of Node doctor: a package that another project
   imports by name stays unmoved, and the review says why.

## Consequences

- A hand-made or adopted project in any of these languages gets a path to the central `dist/`, and
  the projects that cannot move say why without turning doctor red.
- Rust, .NET and Maven fixes edit shared or per-module files; two projects that share a file are one
  finding, so fixes never write the same file twice.
- A published package is not a reason to refuse here: `mvn deploy`, `gradle publish`, `dotnet pack
  -o`, `cargo package --target-dir` and the wheel build write where smith says.
- Gradle and Dart were not run on the machine of the author; their checks rest on fixture tests.

## Where to check

`libs/plugin-{rust,dotnet,jvm,cpp}/pipeline/check_build_dir_use_case.go`,
`libs/plugin-{go,python,dart}/pipeline/check_build_dir_use_case.go`,
`packages/plugin-sdk/{buildrefs,taskoutput}`, `packages/plugin-sdk/xmltext/edit_path_mapper.go`,
`apps/smith/doctorfix/walk_plan_use_case.go`, `apps/smith/removalplan/build_output_group_use_case.go`,
`docs/manual/doctor.md`.
