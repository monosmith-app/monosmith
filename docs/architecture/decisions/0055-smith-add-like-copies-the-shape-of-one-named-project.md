# ADR 0055: `smith add --like` copies the shape of one named project

- Status: accepted
- Date: 2026-11-01

## Context

`smith add` follows what most sibling projects share ([ADR 0041](0041-smith-add-follows-the-dialect-of-sibling-projects.md)).
When the workspace has two ways of doing a thing, or the majority is not the example the person
wants, the only escape was `--fresh`, which gives the stack's default. People asked to copy one
project they trust.

## Decision

`smith add ... --like <project>` gives the new project the shape of the named project instead of the
dominant shape of the siblings.

- The values of the named project's shape (build tool, test runner, layout, lint configuration) are
  sent to the plugins as the shape to follow, for the keys that a plugin of the new project's stack
  follows. This is the same channel as the sibling dialect (`GenerateParams.shape`).
- The command prints what it took, key by key, and says for each key that no plugin of the stack
  follows. `--dry-run` shows the same in the plan line of the project.
- After the project is made, a key that the named project has and the new project does not is a
  warning, as for siblings, in one line.
- `--like` and `--fresh` exclude each other. The named project must exist and report a shape; the
  error lists the projects and the closest names.
- No minimum of three siblings applies: the person chose the example.

## Consequences

- One project is enough to make a habit. The siblings' split is not asked when `--like` is given.
- A plugin needs no change: it already reads the shape.

## Sources

- UX review of smith 2.0 beta.2, idea B1.
- [ADR 0041](0041-smith-add-follows-the-dialect-of-sibling-projects.md).
