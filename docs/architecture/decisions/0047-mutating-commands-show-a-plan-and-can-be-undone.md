# ADR 0047: Mutating commands show a plan, make one commit and can be undone

- Status: accepted
- Date: 2026-10-31

## Context

`smith add`, `remove`, `upgrade`, `derive` and `sync` change many files. A person sees the result
only after the write. Some commands ask a question (`remove` asks for the directory), some have
`--dry-run`, and some have neither. Nothing links an operation to a commit, so to undo one you read
`git status` and guess (epic [#819](https://github.com/monosmith-app/monosmith/issues/819), phase 5,
issue [#824](https://github.com/monosmith-app/monosmith/issues/824), part c).

The rules of the epic apply: the core stays language-free, nothing silently rewrites a file a person
edited, there are no shims, and a run without a person (a script, CI, an agent) must never block.

## Options considered

**A file diff of the whole work tree as the plan.** Rejected: a plan is shown before the write. A
diff exists only after it, unless the command runs twice.

**Run every command twice, once in a scratch copy.** Rejected: it doubles the time of `smith add`,
which starts plugins and installs packages.

**Commit always.** Rejected: it changes the behaviour for every user. The option is opt-in.

**Commit only the files the operation touches.** Rejected: the files are known after the write. The
check "clean for the files it touches" cannot run before it. The work tree as a whole is the check.

**Store the log in `smith.lock`.** Rejected: `smith.lock` is the ownership record of derived files
and is committed. A log of commit hashes belongs to one clone: another clone may hold the same
history under other hashes after a rebase.

**Undo with `git reset`.** Rejected: it rewrites history and loses work.

## Decision

**Plan.** The package `operationplan` holds the plan: a title and groups. Each group has a heading
(what a feature adds, for example "Added by the eslint plugin") and lines. A line has an action
(`create`, `change`, `remove`), a path, a note and an owner (`smith` or `you`). The owner comes from
`smith.lock` for the paths it records, and it is empty for the others.

Each command builds its plan with the code it already has, and does not duplicate it:

- `add`: the directory, the stack's plugins and the workspace files, from the stack that
  `smith add` resolves. When the stack depends on a later answer (the workspace recipe, `--with`, a
  language chosen from several), the plan names the directory and the workspace files, and says why
  the plugin files are missing.
- `remove`: the dry run of `RemoveProject`, which the command already has.
- `upgrade`: the dry run of `Upgrade` (`--check`).
- `derive`: the derive plan, for the files that smith writes.

**Confirmation.** A command asks "Go ahead? [y/N]" only when stdin and stdout are terminals, the
`CI` variables are not set, and `--yes` was not given. Every other run prints the plan and goes on.
`--dry-run` prints the plan and writes nothing (`add` and `undo`; `remove` and `upgrade` had it).
`upgrade` and `derive` show their plan in a terminal, or in any run for `derive`; `remove` keeps its
rule that a run without a terminal needs `--yes`, because it deletes a directory. `convert` keeps
its report (`--dry-run`) and its install question; `deps up` is not changed.

**One commit.** With `[workspace] commit_operations = true` (default `false`), the commands `add`,
`remove`, `upgrade`, `derive` and `sync` commit what they changed:

- Before the command, smith checks that the workspace is in a git work tree with one commit and that
  `git status` is clean, apart from `.smith/`. If it is not, smith prints why and the command runs
  as before, without a commit. A commit would mix the operation with other work.
- After a successful command, smith stages everything, unstages `.smith/`, and makes one commit
  `chore: <command>` (`chore: add package billing`). The body names the command and says that
  `smith undo` reverts it. The repository's hooks run; smith never skips them. If git refuses, the
  changes stay in the work tree and smith says so; the command does not fail.
- `--dry-run` and `--check` runs, failed runs and runs that changed no file make no commit.
- `convert` makes no commit: `smith.toml` does not exist when it starts, so the setting cannot be read.

**Log.** `.smith/operations.json` (git-ignored, local to the clone) lists the commits that smith
made: id, command, commit hash, changed files, time and, once undone, the revert commit.

**Undo.** `smith undo` takes the newest operation that is not undone. It refuses, with the reason
and the next step, when:

- no operation is recorded;
- the commit is not in the repository or not in the history of the current branch (rewritten);
- a later commit changed one of its files (`git diff <commit> HEAD` names it);
- one of its files has an uncommitted change.

Otherwise it prints the plan (the inverse of each change: an added file is removed, a changed file
is restored, a deleted file is created) and asks like the other commands. It runs
`git revert --no-commit <commit>` and commits `chore: undo <command>`. A revert that does not apply
is aborted. History is never rewritten. Operations undo in reverse order, one at a time, because
the check for later commits refuses an operation that later work builds on.

## Consequences

- `smith add` can show the plugin files only when the stack is known before the write. The plan for
  a recipe-driven project is shorter than the work. A later step can ask the plugin for a dry-run
  `generate`; the protocol has no such method yet and this ADR does not add one.
- A person with local changes gets no commit and no undo for that run. The note says why.
- The undo of an operation after other work is refused even when the other work is unrelated but
  touches the same lockfile. The person reverts it with git.
- The setting is a new key in `smith.toml`; `smith.schema.json` and the configuration reference carry it.
