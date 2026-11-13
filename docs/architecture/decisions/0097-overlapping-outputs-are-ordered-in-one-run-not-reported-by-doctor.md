# ADR 0097: Overlapping outputs are ordered in one run, not reported by doctor

- Status: accepted
- Date: 2026-11-12
- Supersedes: Part A of [ADR 0081](0081-a-web-app-builds-once-per-environment-and-overlap-fixes-are-presented-choices.md)

## Context

ADR 0081 Part A made `smith doctor` fail "no two tasks write the same output directory", with a
presented fix: give each task its own folder, or list the tasks under `[workspace]
alternative_outputs`. `smith ci generate` and the assets run printed the same warning.

The owner's case: a local build with source maps and a release build both write `apps/api/dist`.
They never run in the same run, so nothing goes wrong. Doctor still failed, and a converted
workspace (the JATO conversion) ended with this failure among its findings. The setting that
silenced it, `alternative_outputs`, existed only to say "this is not a problem".

What does go wrong happens inside one run: a task clears its output folders before it runs and a
cache hit replaces them, so two such tasks at once fail each other mid-write, and of two in a row
the folder keeps only the last one's files.

## Decision

- Doctor no longer checks overlapping outputs. `smith ci generate` and `smith release assets` no
  longer warn about them.
- `smith run` (every caller of `taskrun.Run`) compares the outputs of the tasks of the one run.
  For each pair that overlaps, the later task in the run's order waits for the earlier one, so
  they run one after the other, never at once; the cache restore is inside the wait too. Before
  the first task, the run prints one warning per pair:
  `▲ api:build-dev and api:build both write apps/api/dist: smith runs them one after the other, and the folder keeps only the files of api:build, the last one`,
  with the remedy on the next line.
- `[workspace] shared_outputs` keeps its meaning: such a folder is filled on purpose and never
  cleared, so its tasks are not compared and may run at once.
- A task that runs until it is stopped (`watch`, `serve`, `start`, `dev`) is not compared: waiting
  for it would never end.
- `[workspace] alternative_outputs` is removed (beta, no compatibility shim). A smith.toml that
  still has it fails to load as any unknown key does.

## Consequences

- A workspace with a dev build and a release build into one folder is clean in doctor.
- `smith run build build-dev` is slower than before by the time of the shorter build, and says why.
- The check needs the run's task graph, so a plan that is never run is not checked.
