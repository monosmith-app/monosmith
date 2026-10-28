# smith ci

`smith ci` runs one phase of the pipeline, or regenerates the pipeline files. The generated
pipelines (GitHub Actions, Azure Pipelines, GitLab CI, CircleCI) call these phases, so CI runs exactly
what you can run locally.

| Phase | What it does |
|---|---|
| `setup` | Installs every toolchain dependency the workspace's plugins need. |
| `audit` | Checks dependencies for known vulnerabilities, plus the policies in `ci.audit_policies`. |
| `verify` | Checks [`smith boundaries`](boundaries.md) first, then runs `lint`, `typecheck`, `test` and `build`, only for the affected projects on a pull request. |
| `pack` | Runs every project's `package` task into `dist/drop`, kept as the pipeline's `drop` artifact on pushes to the release branch. |
| `native` | Lints, tests, builds and packages the projects tagged `native` (cgo) on this OS; the pipeline runs it on each OS and, on a release, attaches the build. |
| `release` | Versions, publishes and tags every package with releasable commits (see [`smith release`](release.md)). |

## When to use it

- To reproduce a CI failure locally: `smith ci verify`.
- To regenerate the pipelines after a change that affects them (a new project, a new
  provider): [`smith ci generate`](ci-generate.md).
- To learn what CI needs configured by hand: [`smith ci check`](ci-check.md).

The phases are meant for CI. On a laptop prefer `smith build`, `smith test` and the other
shortcuts, which do not run installs or the tracked-changes guard.

## Synopsis

```sh
smith ci setup | audit | verify | pack | native | release
smith ci generate [--check]
smith ci check
```

## Subcommands

<!-- smith:begin ci -->
Run a CI phase, or regenerate the pipelines

Usage: `smith ci <command>`

| Subcommand | Meaning |
|---|---|
| `smith ci audit` | Run the audit phase |
| `smith ci check` | List what to set up by hand for the pipelines: secrets, and each private registry's permissions |
| `smith ci generate` | Rewrite the CI pipelines from the workspace's projects (--check only reports drift) |
| `smith ci native` | Run the native phase |
| `smith ci pack` | Run the pack phase |
| `smith ci release` | Run the release phase |
| `smith ci setup` | Run the setup phase |
| `smith ci verify` | Run the verify phase |

No flags.

<!-- smith:end ci -->

The phases take no flags. [`smith ci generate`](ci-generate.md) has `--check`;
[`smith ci check`](ci-check.md) has none.

### setup

<!-- smith:begin ci setup -->
Run the setup phase

Usage: `smith ci setup`

No flags.

<!-- smith:end ci setup -->

Collects the setup steps the plugins contribute (Go modules, Node dependencies, the Python
venv and toolchain, other tools) and runs them in order. With nothing to do it prints
`nothing to setup`.

### audit

<!-- smith:begin ci audit -->
Run the audit phase

Usage: `smith ci audit`

No flags.

<!-- smith:end ci audit -->

Runs the audit steps the plugins contribute (`govulncheck`, `npm audit` for runtime
dependencies, `pip-audit`, and so on). A step whose inputs did not change since it last
passed is skipped with `✓ <step>: no change in <files> since it passed on <date>`. A step
that is configured as warn-only (`ci.audit_warn`) reports without failing: `warning:
pip-audit (exit 1) — not blocking`.

### verify

<!-- smith:begin ci verify -->
Run the verify phase

Usage: `smith ci verify`

No flags.

<!-- smith:end ci verify -->

Runs `lint`, `typecheck`, `test` and `build` on every project except the ones tagged
`native`. When the environment names a pull request target (`SMITH_BASE`, `GITHUB_BASE_REF`,
`SYSTEM_PULLREQUEST_TARGETBRANCH`, `CI_MERGE_REQUEST_TARGET_BRANCH_NAME`), only the affected
projects run. Then it runs the tasks a plugin marked for verify (an app's container image,
say), skipping with a warning when the program is not installed.

Finally the **tracked-changes guard** checks the working tree: if the tasks left tracked
files modified or deleted, verify fails, because a release refuses a dirty tree and the pull
request is where to find that out. A file that was already changed before verify started fails
it only when `CI` is set; on a laptop it is taken for your own edit.

### pack

<!-- smith:begin ci pack -->
Run the pack phase

Usage: `smith ci pack`

No flags.

<!-- smith:end ci pack -->

Creates `dist/drop` (always, so the artifact upload has a path), then runs every project's
`package` task except the native ones. With no `package` task anywhere it prints
`nothing to pack: no project has a package task`; on success, `packed into dist/drop`.

### native

<!-- smith:begin ci native -->
Run the native phase

Usage: `smith ci native`

No flags.

<!-- smith:end ci native -->

Builds the projects tagged `native` for the current OS. See [CI pipelines](../reference/ci.md).

### release

<!-- smith:begin ci release -->
Run the release phase

Usage: `smith ci release`

No flags.

<!-- smith:end ci release -->

Plans and executes the release with pushing of tags on, prints the report as markdown, then runs the post-release steps plugins want after every push to the
release branch, released or not.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The phase passed. |
| 1 | A step or task failed, or the guard found tracked changes. |

Every phase prints `smith ci <phase> passed in <duration>` or `... failed in <duration>`, and
when `GITHUB_STEP_SUMMARY` is set appends a line to the job summary.

## Files and environment

Reads `smith.toml`, the project graph and the plugins' CI contributions. `generate` writes
the pipeline files (see its page). Variables: `CI` (changes the guard), the pull request
target variables above, `GITHUB_STEP_SUMMARY`, and the release credentials of
[`smith release`](release.md#files-and-environment).

## Examples

A short phase:

```sh
smith ci audit
```

Every phase opens with a banner that says what it will do and on what, folds the tools' own
output in a collapsed log group (`::group::` on GitHub Actions, `##[group]` on Azure Pipelines),
and ends with a result table and one closing line. The banner, the table and the closing line
stay visible; the same table goes to the job summary (`$GITHUB_STEP_SUMMARY` on GitHub, an Azure
summary section). Status words and glyphs are coloured (green passed, yellow warning or dependency-caused,
red failed, dim unchanged) only where the CI renders it: GitHub Actions, Azure Pipelines and
GitLab, or a terminal. `NO_COLOR` and `TERM=dumb` switch it off, and a piped log has none, so the
words read the same without it; the job summary is always plain. `SMITH_ASCII=1` swaps the symbols for `[v]` `[x]` `[!]`.

```text
══ Audit ═══════════════════════════════════════════════════
Auditing 26 projects with 22 checks.
Rules: smith's default policies ([ci] audit_policies).
Check                             Result       Time  Note
govulncheck actions               ↷ unchanged     -  nothing changed since it passed on 2026-10-27
govulncheck plugin-sdk           ✓ done       1.5s
npm audit (runtime dependencies)  ✓ done       2.7s
pip-audit                         ▲ problems   0.0s  reported problems, not blocking
→ read each step's findings in the folded output above
✓ Audit passed in 33s: 22 checks over 26 projects: 5 ran, 4 unchanged since an earlier pass, 13 reported problems (not blocking)
```

```sh
smith ci setup
```

```text
══ Setup ═══════════════════════════════════════════════════
smith 1.102.2 on windows/amd64
Installing dependencies for 4 toolchains (dotnet 10.0, go 1.26, node 24, python 3.12) and preparing 26 projects: 8 steps.
Step                          Result  Time
Download Go modules           ✓ done  0.1s
Install Node dependencies     ✓ done   37s
Create the Python venv        ✓ done   15s
Install plugin-sdk-py         ✓ done  4.9s
✓ Setup passed in 1m 32s: 8 steps ran, 26 projects ready
```

`smith ci verify` ends with one row per project and one column per task, failures first; a failure
prints the end of its output and the command that repeats it, and the test column carries the
runner's own count when smith can read it (pytest, vitest, jest, dotnet) or the number of Go
packages:

```text
══ Verify ══════════════════════════════════════════════════
Checking 26 projects (lint, typecheck, test, build).
The dependency boundaries between projects are checked first.
Project            Lint      Types   Tests              Build     Time
jest-preset        ✗ failed  ✓       ✓ 19 passed        ○ none  1m 27s
smith              ✓         ○ none  ✓ 116 packages ok  ✓       2m 49s
eslint-config      ✓         ✓       ✓ 75 passed        ○ none  2m 37s
vscode             ✓         ✓       ✓ 220 passed       ✓          56s
last output of jest-preset:lint:
  Oops! Something went wrong! :(
→ repeat it with `smith run lint -p jest-preset`
✗ Verify failed after 2m 37s
```

`smith ci pack` runs the `package` task of the apps (here a Go app and a VS Code extension), then
lists what is in `dist/drop`, which the generated pipeline keeps as the artifact `drop` (the step
is named for it in the GitHub and Azure files):

```text
══ Pack ════════════════════════════════════════════════════
Packing 2 projects (task "package") into dist/drop.
Project  Package  Time
smith    ✓        2.0s
vscode   ✓        1.3s
Packed file                                Size
dist/drop/monosmith-vscode-0.0.0.vsix  113.3 KB
dist/drop/smith/smith.exe               28.5 MB
✓ Pack passed in 2.1s: 2 files packed into dist/drop (28.6 MB)
```

`smith ci release` says what it will release (the banner and a one-line plan), folds the whole
log (the version it resolved from each tag, the bump and why, the publishing) in the collapsed
group, then prints one table: a row per released package with the version it was and becomes, the
bump (major, minor or patch), the commit that decided it or the dependency or `version()`
directive that forced it, where it was published, its tag and how long it took. Failed packages
come first. Packages with nothing to release are not rows: they are listed on one line, or counted
when there are many. When nothing releases it says so, with the way to make something release. On
GitHub the closing result is also a `::notice::` annotation on the run:

```text
══ Release ═════════════════════════════════════════════════
Releasing 2 projects of 5 from main, in dependency order: each is versioned, built, published and tagged.
Plan: core 0.2.0, logger 0.1.1
Project  Was    Now      Bump   Why                                           Published             Tag                          Time
core     0.1.0  › 0.2.0  minor  feat(search): add search to the core (+1 more)  ✓ registry.npmjs.org  packages/core/v0.2.0 (pushed)    0.3s
logger   0.1.0  › 0.1.1  patch  dependency core 0.2.0                          ✓ registry.npmjs.org  packages/logger/v0.1.1 (pushed)  0.2s
Unchanged (3): docs, tool, util
✓ Release passed in 0.8s: 2 released (core 0.2.0, logger 0.1.1), 3 unchanged, 0 failed
```

```text
══ Release ═════════════════════════════════════════════════
Looking for commits that release (feat, fix, perf or a breaking change) since each project's last tag.
Nothing to release: none of the 2 projects has a releasing commit since its last release.
→ commit with a message like "feat: add search" or "fix: handle empty input"; docs, chore, test and ci commits do not release
✓ Release passed in 0.2s: nothing to release
```

## Common errors

| Message | Cause and remedy |
|---|---|
| A task failure in `verify` | See the task's output, then `smith why-miss <project>:<task>`. |
| The tracked-changes message listing files and the tasks that likely wrote them | A task rewrote a tracked file. Commit it with the change, stop tracking it (`git rm --cached <file>`), or fix the task. |
| `unknown phase "x" (phases: setup, audit, verify, pack, native, release)` | Only those six exist. |

## Related

[`smith ci generate`](ci-generate.md), [`smith ci check`](ci-check.md), [`smith release`](release.md),
[`smith pipeline`](pipeline.md), [CI pipelines](../reference/ci.md).
