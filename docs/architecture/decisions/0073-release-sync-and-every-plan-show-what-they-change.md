# ADR 0073: release, sync and every plan show what they change

- Status: accepted
- Date: 2026-11-02

## Context

[ADR 0072](0072-smith-presents-every-decision-that-changes-something.md) fixed the first wave of
silent choices. An audit of the rest found more. `smith release` chose the bump, the tags and the
publish from the commits and asked one y/N. `smith sync` and `up --align` picked a version for the
person, rewrote every project and installed. `smith kind` moved a folder and edited many files with
no question. `smith ci generate` rewrote and deleted files with no diff. `smith up --yes` took major
updates. `eject --force`, `report -o`, `graph -o`, `plugin draft -o` and `derive --replace`
overwrote a file without a diff. `smith tauri add` wrote a placeholder bundle identifier. The
confirmation dialogs of the TUI showed one sentence and passed `--yes`.

## Decision

- **Release.** smith shows each package with the reason it releases and lets the person hold a
  package back or change its bump: in a terminal by a choice, anywhere by `--hold <project>` and
  `--bump <project>=patch|minor|major`. A package that releases only because a held one does is held
  with it. A first release, a `version()` directive and a release channel decide their own version,
  so `--bump` refuses them. The last question says what cannot be undone: the tags, the push and the
  registry publish.
- **Sync and `up --align`.** smith prints a plan with one line per project, then asks which version
  wins: smith's pick, the highest declared, the lowest declared, or one dependency at a time with
  every version it can use. Flags: `--dry-run`, `--yes`, `--prefer highest|lowest|installed`,
  `--to <name>=<version>`. A declaration on another major version stays unless `--major` is given.
  Without a terminal smith refuses unless a flag answers.
- **Major updates.** `up --yes` takes patch and minor updates. A major update is a choice of its own:
  smith lists it and `--major` takes it. A named package with a major update asks in a terminal.
- **Kind.** `smith kind` needs a clean work tree (`--allow-dirty` overrides), shows the plan with a
  diff of each edit, asks, and commits only with `--commit`.
- **Files smith owns.** `ci generate` and `bootstrap` take `--dry-run` and show a diff of every
  file they create, rewrite or delete. `eject --force`, `report -o`, `graph -o` and `plugin draft -o`
  refuse to replace a file that differs unless the person gives `--force`, and print the diff of what
  is lost. `derive --replace` prints the diff of each hand edit it throws away.
- **Tauri.** `smith tauri add` shows the plan, takes `--dry-run`, and asks the bundle identifier
  (`--identifier`; `--yes` takes `com.<workspace>.<project>` and says so).
- **Interfaces.** A command that changes the workspace and has a plan flag declares it in the action
  catalog (`Plan`). The TUI runs that flag first and shows its text in the confirmation, so the person
  answers yes to a plan. The release confirmation reads the plan itself. The VS Code launcher prompt
  does the same with `bootstrap --dry-run`.
- **Wave 1 leftovers.** The TUI New form asks the first commit of a folder that has files and the Go
  module prefix. The TUI remove form asks keep a copy or delete. The TUI adopt wizard asks line
  endings, the CI provider and a pipeline of the person's own, then shows the plan of `adopt --init`.
  The agent files keep their own step. The interactive remove plan prints once.

## Turning a prompt off

`[workspace] confirm = "always" | "risky"`. `"always"` is the default and asks for every plan.
`"risky"` asks only for a plan that deletes a file, or edits a file that exists and shows a diff of
it. A plan that only creates files is printed and applied, with a line that says so, even without a
terminal. A flag always wins. Release, the keep-a-copy question of `remove`, and the choices that
name a different result are questions, not confirmations, and the setting does not skip them.

## Consequences

- Scripts that ran `smith sync`, `up --align`, `kind`, `ci generate`, `bootstrap` or `tauri add`
  with no terminal and no `--yes` now stop with a message that names the flags.
- `smith up --yes` and the Deps view no longer take a major update by accident. The Deps view passes
  `--major` for a major update the person marked.
- `smith sync --yes` no longer crosses major versions. Use `--major`.
- `smith report`, `graph`, `plugin draft` and `eject` no longer replace a different file without
  `--force`.
- Not done: the VS Code extension has no commands for release, sync, kind, ci generate or tauri add,
  so it has nothing to mirror. The adopt offer runs `smith adopt --init` in a terminal, which asks.
  JetBrains is unchanged.

## Sources

- [ADR 0072](0072-smith-presents-every-decision-that-changes-something.md)
- [ADR 0047](0047-mutating-commands-show-a-plan-and-can-be-undone.md)
- `docs/reference/feature-parity.md`
