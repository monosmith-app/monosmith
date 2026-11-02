# ADR 0075: An assets rerun for a tag asks which code to build

- Status: accepted
- Date: 2026-11-02

## Context

A first real release with the assets config (MVD, 2.0.0-beta.18) worked and showed three
problems.

1. A `build` that ran two smith tasks into one folder lost a file. Each task clears its outputs
   before it runs, so the second task (`package-dmg`, output `dist/drop`) deleted the zip the first
   task made.
2. The assets job did not install `[ci] native_linux_packages`. The native job does.
3. `smith ci assets --tag <project@version>` (ADR 0069) builds the current checkout under the
   version of the tag. The files then carry newer code under an old number. The owner does not
   publish that without choosing it.

## Decision

- A folder listed in `[workspace] shared_outputs`, or inside a listed folder, is not cleared before
  a task runs and is not cleared when a task restores its outputs from the cache. smith warns, in
  `smith ci generate`, `smith doctor` and the assets run, when two tasks declare the same output
  folder and it is not listed. The warning names both tasks and the two fixes.
- The generated assets job runs an `apt-get install` step for `[ci] native_linux_packages` on its
  Linux runners.
- `--source current|tag` is a real choice, with `smith ci assets --tag` and `smith release assets
  --tag`. It matters only when the tag is not at HEAD.
  - `current`: the code as it is now, named with the tag's version.
  - `tag`: the tag's own commit, built in a temporary `git worktree add --detach`. smith opens that
    checkout as a workspace, installs its dependencies (the assets setup), runs the build and
    removes the worktree, also after a failure. It uses the smith.toml of that commit. When that
    commit has no smith.toml, or its smith.toml gives the project no release build, smith says so
    with the commit and tells the user to use `--source current`.
  - Without `--source`, a terminal is asked, with the result of each option in plain words. A run
    without a terminal is refused, and the message names both flags.
- The generated workflow has the input `assets_source` (choice `choose`, `current`, `tag`; required;
  the descriptions say what each means). It reaches the command through the environment variable
  `ASSETS_SOURCE`. The value `choose` means no source was given, so the run is refused and names
  both flags. GitHub always sends a value for a required choice, so `choose` is the first and the
  default option: nobody gets a source they did not pick.

## Consequences

- A rerun built from the current checkout is now an explicit choice. A rerun of a tag can build the
  exact code that was released, when the build is the same kind of build as before.
- The `tag` source has no dependencies installed in the new worktree. The setup phase installs
  them before the build. A project whose build needs more than the setup installs must run it in
  its own build command.
- The overlap warning covers every pair of tasks, not only the tasks of one build command. smith
  cannot see which tasks a free-form build command runs.

ADR 0069 keeps its other decisions. Only the sentence that the rerun builds the current checkout
is replaced by this record.
