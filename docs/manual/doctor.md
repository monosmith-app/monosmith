# smith doctor

`smith doctor` checks the workspace's health: smith's own invariants and each plugin's
checks. Every problem names its remedy. With `--fix` it shows the fixes it can make (file
changes as diffs, commands it would run such as `npm install` or `git rm --cached`, and
questions it would ask) and applies the safe ones.

A **safe** fix only restores what smith or a plugin owns, such as `go.work` or a generated
pipeline with its slots kept; doctor applies those once you confirm (or with `--yes`). Any
other fix, and any fix with a question, is asked about in a terminal. Without a terminal it
is shown with the command that applies it. The same fix on many projects is one fix that
says which projects it applies to.

## Local toolchain versus CI

For Node, Python, Go, Rust, .NET and the JDK, doctor compares the tool on this machine with
the version the generated CI workflow and dev container install, so no setting is needed
(`Rust is 1.90, as in CI`). A newer local tool is a warning; an older one fails, because
the build can then fail with an error that reads like a dependency problem. A tool that is
absent gives no version finding (the plugin's "installed" check already says so). Where smith
can apply a fix it offers one: for Rust, a `rust-toolchain.toml` pinning the version plus
`rustup toolchain install`, which also makes CI and the dev container follow the pin. A
`[toolchains]` pin in `smith.toml` is a separate check, against the same probed version.

## Suggestions about a project's kind

Two warnings point at [`smith kind`](kind.md). They are suggestions, never failures, and they
say nothing when there is nothing to say:

- `the lib <name> is used by a project`: no project imports the lib, and it has been in the
  repository for 14 days or more. A lib is shared inside the repository; this one is shared
  with nobody. A lib added today, or in a workspace of one project, is left alone. The remedy
  offers `smith kind <name> app`, `smith kind <name> package` or `smith remove <name>`, and
  says to leave it if something outside the graph uses it.
- `the package <name> is published`: a package that its own settings (`registry = "none"`
  under `[projects.<name>]`) or its manifest (Dart's `publish_to: none`) never publish, while
  the workspace does publish. The remedy offers `smith kind <name> lib`. A workspace that
  releases everything by tag (`registry = "none"` under `[release]`) is not asked.

## When to use it

- After cloning, after a merge, after changing tools, and before a release.
- In CI, as a gate: it exits 1 while any check fails.
- With `--fix` to repair generated files.

## Synopsis

```sh
smith doctor [--verbose] [--json]
smith doctor --fix [--yes] [--only <check or fix>]... [--choose <n>,<n>]
```

## Arguments and flags

<!-- smith:begin doctor -->
Check the workspace's health; every problem names its remedy

Usage: `smith doctor [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--choose <choose>` | integers, comma-separated |  | with --fix and --only, answer the fix's questions by option number, in order (1,2) |
| `--fix` | boolean | off | show the fixes doctor can make as diffs, and apply the safe ones |
| `--json` | boolean | off | print the report as JSON |
| `--only <only>` | string, repeatable |  | with --fix, fix only the named check, as doctor prints it (repeatable) |
| `--verbose` | boolean | off | list every passed check, not one line for each check that repeats per project |
| `-y, --yes` | boolean | off | with --fix, apply without asking |

<!-- smith:end doctor -->

Interactions:

- `--only` and `--choose` pick what to fix, so they need `--fix` (`--only and --choose pick
  what to fix, so they go with --fix`).
- `--fix` alone shows the fixes as diffs and, in a terminal, asks. Without a terminal, or
  without `--yes`, it prints `Pass --yes to apply these fixes.` and exits 1.
- `--only "<name>" --yes` applies one fix by the name doctor prints (the check's or the
  fix's). Doctor exits non-zero while a named fix is not applied.
- `--choose 1,2` answers the fix's questions by option number, in order, for the fix `--only`
  names.
- Fix steps run in the workspace root and stop at the first that fails; the failing step's
  output is shown.
- `--verbose` lists every passed check instead of one line for each check that repeats per
  project.
- `--json` prints `findings` (each with `check`, `ok`, `detail`, `remedy`, `warning`,
  `source`), `passed`, `failed` and `warnings`. For a passing check `detail` and `remedy`
  describe the failure that would be reported.

A check is `✔` (passed), `✖` (failed, exit 1) or `!` (a warning, not counted as a failure).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | No check failed (warnings allowed). |
| 1 | A check failed, or `--fix` could not apply what was asked. |

## Files and environment

Reads the workspace and runs the plugins' checks (they may call `go`, `node`, `python`,
`git`). `--fix` rewrites generated files (pipelines, `go.work`, the dev container) and runs
the commands it displays. Variables that change checks: `SMITH_CACHE_REMOTE` and
`SMITH_CACHE_KEY` (the remote cache must be signed), `SMITH_NO_UPDATE_CHECK`, and the
release credentials ([`smith release`](release.md)).

## Examples

A failing workspace: the pipeline file is stale and `go.work` was deleted.

```sh
smith doctor
```

```text
✖ CI pipelines match the workspace
    .github/workflows/ci.yml is out of date
    → run `smith ci generate` and commit the result
✖ go.work exists
    the workspace has Go projects but no go.work
    → run `go work init` then `go work use` for each Go project
✔ the cache directory is ignored by git
✔ the repository has commits
...
✔ Node projects are checked for vertical slices (@monosmith/eslint-config)

25 passed, 2 failed, 0 warning(s)
1 fix(es) available: `smith doctor --fix`
```

The fix, shown as a diff, then applied:

```sh
smith doctor --fix
```

```text
...
● regenerate the pipelines, keeping your slot contents
--- a/.github/workflows/ci.yml
+++ b/.github/workflows/ci.yml
@@ -25,7 +25,7 @@
...
-
-# stale

Pass --yes to apply these fixes.
```

```sh
smith doctor --fix --yes
```

```text
...
✔ regenerate the pipelines, keeping your slot contents

after fixing: 26 passed, 1 failed, 0 warning(s)
```

Read the diff before applying: in this trial `go.work` was missing, and the regenerated
pipeline fell back to `go-version: '1.24'` where it had `'1.27'` (the version came from
`go.work`). Restoring `go.work` first (`go work init ./apps/api ./apps/web2`) and then
running `smith ci generate` put `1.27` back.

A repository adopted a moment ago, in a non-interactive run, names the command that applies
a fix that asks:

```text
● install the Node dependencies
  step 1: install the workspace's dependencies
          $ npm install --no-audit --no-fund

not applied: install the Node dependencies needs your decision. Run one of:
  smith doctor --fix --only "install the Node dependencies" --yes
```

## External plugins

For every external plugin `smith.toml` declares, doctor shows where it comes from, whether it is
trusted and pinned, and the environment variables it may see (names, never values):

```text
✔ external plugin widget is trusted (from .smith/plugins, pinned no, sees the baseline plus GITHUB_TOKEN, NPM_TOKEN)
```

An external plugin smith would refuse to start fails the check, and because no plugin starts until
it is trusted, doctor shows only the plugin checks and exits non-zero until you trust it:

```text
✖ external plugin widget is trusted
    the external plugin "widget" is not trusted yet, so smith did not start it (from .smith/plugins, pinned no, sees the baseline plus GITHUB_TOKEN, NPM_TOKEN)
    → run `smith plugin trust widget` to review it and trust it
    → run `smith plugin list` to see every external plugin and what it may see

0 passed, 1 failed; the other checks run once these plugins are trusted
```

A plugin installed on the search path but not declared in `smith.toml` is not needed by the
workspace and is not checked. See [Plugin security](../plugins/using/plugin-security.md).

## Common errors

| Message | Cause and remedy |
|---|---|
| `external plugin <name> is trusted` fails | smith will not start an external plugin you have not trusted (or that changed since you did). Run `smith plugin trust <name>`; in CI, `smith plugin trust <name> --pin` on your machine and commit the `sha256`. |
| `--only and --choose pick what to fix, so they go with --fix` | Add `--fix`. |
| `not applied: <fix> needs your decision. Run one of: smith doctor --fix --only "<fix>" --yes` | A fix with a question or a command is not applied without a terminal; run the command it prints. |
| Check-specific failures | Each prints its own remedy; the names are stable, for example `go.work exists`, `Node dependencies are installed`, `CI pipelines match the workspace`, `The MonoSmith agent files are installed and current`, `tracked files are not generated files`, `Jest helper files of an older convert are replaced by @monosmith/jest-preset`. |

## Related

[`smith ci generate`](ci-generate.md), [`smith upgrade`](upgrade.md),
[Troubleshooting](../guides/troubleshooting.md), [`smith explain`](explain.md).
