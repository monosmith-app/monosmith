# ADR 0065: `smith swap` changes one choice of a project, and migrates where the move is mechanical

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
- Without `--dry-run` the command migrates the axes where the move is mechanical, and refuses, with the
  manual steps, on the others. Today the mechanical axis is the test runner (Jest and Vitest). The
  Node toolchain plugins already describe what a choice adds as a setup: dev dependencies, scripts,
  fields and files. The swap computes the setup of the old and of the new plugin for the project and:
  - removes the old plugin's root dev dependencies that the new one does not need (kept when another
    project still uses the old plugin), its scripts and fields, and its config files, each only where
    the project still has exactly what the plugin wrote; an edited script, field or file stays and is
    listed as left to do by hand;
  - has the new plugin write its own setup, which adds only what is absent;
  - moves test imports of `describe`, `it`, `expect` and the hooks to the new module; a file that
    uses `jest.fn`, `jest.mock` or `vi.*` is left and listed;
  - records the new stack in the manifest (the stack whose toolchain is the current one with the
    plugin replaced), installs the packages once, and runs the project's build and tests.
  The lint axis has one choice (ESLint), so there is nothing to swap. A style swap edits entry files
  and the Vite config through template slots, and is not mechanical yet: the command names the
  change to make by hand.
- A project made from a stack that a plugin publishes by name (`node-jest`) lists no choices. The
  command reads its choices from its plugins: a plugin that offers one choice contributes it.
- The solver check for a swap does not ask a choice to write the project, because the project exists.
- The command takes the shared `--yes`, `--dry-run` and `--commit` flags of the mutating commands
  ([ADR 0047](0047-mutating-commands-show-a-plan-and-can-be-undone.md)), so a later plugin
  migration needs no change of the command line.

## Consequences

- People get an early, checked answer to "can this project use that runner, and what comes with it".
- A migration lives in the host, over the setup a plugin already gives, so no protocol method is
  added. A plugin outside smith that owns a choice does not migrate yet; a later record can add a
  plugin method for that.
- A later step can cover the style axis once a style choice can say which lines of the entry file and
  the Vite config it owns.

## Sources

- UX review of smith 2.0 beta.2 and the ideas batch brief (exemplar and swap).
- [ADR 0045](0045-frameworks-tests-and-styles-are-choices-shipped-as-data.md).
