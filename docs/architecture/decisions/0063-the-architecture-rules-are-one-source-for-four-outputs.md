# ADR 0063: The architecture rules are one source for four outputs

- Status: accepted
- Date: 2026-11-01

## Context

Idea C1 of the design review: define the architecture once and derive the cheat sheet, the lint and
boundary checks, the scaffolds for new files and the instructions for AI agents. This ADR records
what was already true, the gaps found, and the gaps closed.

## Where the rules live

- Vertical slices: `libs/plugin-vertical-slices/languages` (`Roles`, `ForbiddenFolders`, one rules
  file per language) and `layout` (the slice size).
- Layered architectures (hexagonal, clean, ddd, modular-monolith, mvc, package-by-layer):
  `libs/plugin-architectures/rules` (a `Ruleset` with `Layers`, `Summary`, and flags).
- The catalog of the choices a person picks: `apps/smith/architecture` (name, one-line meaning, the
  `smith.toml` table).

## What was already true

- `smith lint`, `smith slices check` and `smith architecture` read the rules from those packages:
  the check is the rules.
- `smith boundaries` is separate by design: it checks the dependency rules the person writes in
  `[boundaries]`, not an architecture.

## Gaps found and closed

1. The agent file (`smith agent file`) kept its own copy of the file roles and of five of the ten
   forbidden folder names. It now reads `languages.Roles` and `languages.ForbiddenFolders`. A test
   fails when the agent file lacks a role or a folder the check knows.
2. For a layered architecture the agent file gave only the one-line meaning. It now lists the layers
   (name, folder spellings, meaning) and the rule summary from the `Ruleset` that `smith lint`
   enforces.
3. Nothing proved that the files `smith add` writes follow the rules. A test now generates the Go app
   and lib scaffolds and runs the slice check on them.
4. `docs/reference/conventions.md` listed the roles by hand. A test fails when it lacks a role.

## Not closed

- The scaffolds of the other languages are not under the same test; each needs its generator run
  without its installer. Add one test per language when the language's runner can be stubbed as the
  Go one is.
- The catalog `Meaning` and the `Ruleset.Summary` say related things in two places. They answer
  different questions (what the choice is; what the checker enforces) and are both shown now, so
  they are left as they are.
- The conventions page is checked for the roles, not generated. Generating it would make the page a
  build artifact with no gain a test does not give.

## Sources

- `apps/smith/agentdoc/`, `apps/smith/slicecheck/check_slices_handler_test.go`
