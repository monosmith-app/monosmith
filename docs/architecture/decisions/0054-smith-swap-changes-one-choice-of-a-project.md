# ADR 0054: `smith swap` changes one choice of a project, and stops where the migration is not automatic

- Status: accepted
- Date: 2026-11-01

## Context

A project is made from choices ([ADR 0045](0045-frameworks-tests-and-styles-are-choices-shipped-as-data.md)):
one per axis, such as the test runner, the linter and the style framework. People change their mind
later. Today they edit dependencies, config and imports by hand, and nothing checks that the new set
of choices fits together. The UX review of beta.2 asked for a command that changes a decision on an
existing project.

## Decision

`smith swap <project> <axis> <choice>` changes the choice of a project on one axis.

- The command reads the choices of the project's stack, replaces its choice on the axis, and checks
  the result with the choice solver of `smith add --with` (`choiceselect.Solve`). Nothing new is
  decided about what fits: the solver decides.
- The command prints the plan: the choice that leaves, the choice that joins, and the choices of the
  result. `--dry-run` stops there.
- Without `--dry-run` the command refuses, with the manual steps, because moving dependencies,
  config and imports needs a migration that the plugin owning the choice must provide. No plugin
  provides one yet. A swap that rewrote only the stack record would leave the project half changed,
  so smith does not do that.
- The command takes the shared `--yes`, `--dry-run` and `--commit` flags of the mutating commands
  ([ADR 0047](0047-mutating-commands-show-a-plan-and-can-be-undone.md)), so a later plugin
  migration needs no change of the command line.

## Consequences

- People get an early, checked answer to "can this project use that runner, and what comes with it".
- No file changes yet. A later record defines the plugin method that migrates one axis and says what
  is reported (dependencies, config files, imports and tests that name the old choice).
- The first axes to migrate are the ones smith already offers as choices: test runner, linter, style
  framework.

## Sources

- UX review of smith 2.0 beta.2 and the ideas batch brief (exemplar and swap).
- [ADR 0045](0045-frameworks-tests-and-styles-are-choices-shipped-as-data.md).
