# ADR 0072: smith presents every decision that changes something

- Status: accepted
- Date: 2026-11-02

## Context

An audit of smith found places where it chose for the person without saying so: `smith new` committed
every file of a non-empty folder (a `.env` was committed), `smith remove` deleted untracked files for
good, `smith adopt --init` wrote `.gitattributes`, the agent files and a CI provider in silence, and
`add`, `upgrade`, `derive`, `undo` and `convert` applied a plan with no terminal and no `--yes`.
The person has read no manual, so a default that smith picks and does not name is a decision taken
for them.

## Decision

- A decision that changes something is a real choice. Each answer says what it does in plain words
  and the safe answer is the default. In a terminal smith asks. Without a terminal smith refuses and
  names the answers as flags. `--yes` takes the safe default of every open choice and prints what it
  took and the flag that changes it. `--yes` never picks the answer that loses data.
- A silent default is allowed only when nothing is lost and the output says what was chosen and how
  to change it.
- A plan command (`add`, `upgrade`, `derive`, `undo`, `adopt --init`, `remove`, `convert`) does not
  apply without a terminal unless `--yes` is given. `--dry-run` shows the plan. The plan names every
  file smith may rewrite (including the generated pipeline files) and shows a diff for a file of the
  person that it edits.
- `smith new` in a folder that holds files and no git repository lists the files and asks what the
  first commit holds: only smith's files, nothing, or everything. Files whose names look like secrets
  (`.env*`, keys, credentials) need a second, explicit yes (`--include-secrets`). Flags:
  `--first-commit smith|all|none`, `--no-commit`.
- `smith remove` lists the files git cannot bring back (untracked, uncommitted) and asks: move the
  folder to `.smith/removed/<name>-<time>/` (default), delete it for good, or cancel. Flags:
  `--keep-copy`, `--delete`; `--yes` alone does not choose.
- `smith adopt --init` asks four separate questions with their consequences: `.gitattributes`
  (`--gitattributes add|skip`), the agent files (`--agent add|skip|never`), the CI provider
  (`--ci <provider>|none`, detecting GitHub, Azure, GitLab, CircleCI and Bitbucket), and a pipeline
  of the person's own (`--ci-existing replace|keep`). It then shows the plan with diffs before it
  writes. `--init --dry-run` shows the same plan.
- `smith new` asks the Go module prefix when there is no git remote (`--module-prefix`) and the dev
  container (`--devcontainer`, `--no-devcontainer`). Under `--yes` it prints "Chosen for you" with
  the flag of each default.
- `smith convert` shows the files it deletes and whether a pnpm or yarn workspace switches to npm,
  and offers apply, show the report and stop, or cancel.

## Turning a prompt off

Every prompt has a flag, and the flag is the way to silence it. A workspace setting
(`[workspace] confirm = "always" | "risky"`) that asks less is proposed, not built: it needs the
owner to pick which prompts count as risky. Until then the smallest switch is the flag, or `--yes`.

## Consequences

- Scripts that ran `smith add`, `upgrade`, `derive`, `undo` or `adopt --init` with no terminal and no
  `--yes` now stop with a message that names `--yes`.
- `adopt --init --yes` no longer writes `.gitattributes` or the agent files. It used to write both.
- The TUI adopt wizard runs `adopt --init --yes`, so it takes the safe defaults and prints them;
  asking the questions inside the wizard is open work (see `docs/reference/feature-parity.md`).
