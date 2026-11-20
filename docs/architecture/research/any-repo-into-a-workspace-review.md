# Design review: any repository into a smith workspace

Date: 2026-11-20. smith 2.57.4. Status: findings and proposal. The owner decided D1 to D5 on 2026-11-20.

## Goal

A person takes any repository (Node, PHP, Java, mixed) and makes it a smith workspace. Then the person adds more projects to that workspace from a git repository or from a local folder. After each step, the build, the tests, the lint and `smith doctor` pass. If something cannot pass, smith says what is left, why, and the one step that clears it. This is the "just works" goal.

## How the review was done

Five of the owner's own public repositories, cloned into a lab folder without their remote:

| Repository | Shape |
|---|---|
| font-color-contrast | one published TypeScript npm package at the root, Jest, legacy ESLint |
| faviconit-api | one Express API (JavaScript, ESM), its own lockfile, npm-init test placeholder |
| colorvalidator | one Composer package (PHP, Laravel validator) |
| font-color-contrast-demo | one create-react-app front end |
| segurancarussoedutarikturco | mixed: an Eclipse Java project (no Maven or Gradle), PHP files, an Xcode C project |

Journey: `smith adopt` on font-color-contrast, then bring faviconit-api and colorvalidator into the same workspace, then `smith install`, `build`, `test`, `doctor`, `doctor --fix`, `release --dry-run`. Then `smith adopt` on colorvalidator and on the mixed repository alone.

## What worked

- The adopt report is clear: it shows the folder choice, the git mv commands, what stays at the root and what moves.
- The move of a root project into `packages/<name>` was correct. Build and 91 tests passed after `smith install`.
- `smith projects` found the copied Node and PHP projects at once, with the right kind.
- The failed-task hint for a missing `nodemon` is good.

## Findings

| # | Severity | Finding | Evidence |
|---|---|---|---|
| F1 | high | There is no way to add a project from a git repository or a folder. `smith add --from` does not exist. The only path is a manual copy, which loses the git history. | `smith: unknown flag: --from` |
| F2 | high | A copied project is seen but not wired into its language's workspace. smith lists it, but npm `workspaces` has only `packages/*`, so its dependencies are not installed. | `apps/faviconit-api` listed by `smith projects`; `express` missing from `node_modules` |
| F3 | high | `smith install` says "installed 2 project(s)" when one of them was not installed. | same as F2 |
| F4 | high | The same gap for PHP: the PHP plugin runs `composer install` at the root, and there is no root `composer.json`. The message says "fix the cause above" but not what the cause is in smith terms. | `Composer could not find a composer.json file in <root>` |
| F5 | high | Doctor reports "not in the root package.json workspaces" with no fix. | `× faviconit-api: is not in the root package.json "workspaces"`, no fix listed |
| F6 | high | `adopt` knows fewer languages than the workspace does. A Composer repository gives "no project found", but the same folder inside a workspace is a PHP project. | adopt on colorvalidator: `smith knows a project by its manifest (package.json in a workspace, go.mod, pyproject.toml)` |
| F7 | high | A repository without a build manifest (Eclipse Java with `.classpath`, bare PHP, Xcode C) gives "no project found" and no next step except `smith add`, which makes a new empty project. | adopt on the mixed repository |
| F8 | high | The default (conservative) adoption is not green by design. The project's own ESLint setup is kept, so doctor fails on the missing root `eslint.config`. `--level full` is refused for a plain repository (issue #553). | doctor after adopt: 1 failed |
| F9 | med | smith does not see the repository's lint. The project has `.eslintrc.js` and ESLint in its devDependencies but no `lint` script, so it gets no lint task. | `! font-color-contrast: has no lint task` |
| F10 | med | The npm-init placeholder `echo "Error: no test specified" && exit 1` becomes a real test task and fails `smith test`. | `faviconit-api:test × exited with code 1` |
| F11 | med | adopt says "smith adopt wrote files that smith doctor rejects. That is a bug in smith adopt", then "✓ adopted", then repeats the same two items. The two items are expected steps (install, lint config), not a bug. | adopt output |
| F12 | med | adopt says "Agent files: skip", then `doctor --fix --yes` writes `.claude/` and the MCP config as a "safe fix". | doctor --fix output |
| F13 | low | The README of a published package stays at the root after the move. The package that npm publishes from `packages/<name>` then has no README. | `leave at the root: ... readme.md` |
| F14 | low | The copied project keeps its own `package-lock.json`, `.gitignore` and `.editorconfig` inside its folder. Nothing says that the nested lockfile is now dead. | `ls apps/faviconit-api` |
| F15 | low | `release --dry-run` says "never released" advice for a package whose tags (`v1.1.1`) and npm version (11.1.0) do not match. It does not point at `adopt --tags`. | release dry run |

## Root causes

1. **Two kinds of knowledge.** `adopt` has its own detection (npm, Go, Python). The workspace discovery uses every language plugin. So adopt refuses what the workspace accepts (F6, F7).
2. **"Seen" is not "wired".** Only `smith add` wires a project into its language's workspace (npm `workspaces`, root `composer.json`, `go.work`, Cargo members, uv members, Maven modules, a .NET solution). A project that arrives any other way is seen, but nothing joins it, nothing checks the join, and install reports success (F2 to F5).
3. **Conservative is the only level for a plain repository, and it cannot be green.** It keeps the repository's own lint and placeholders, and then doctor fails on them (F8 to F10, #553).
4. **Bringing a project in is not one operation.** Adopt has the move, the hoist of dev tools, the tsconfig split and the lint decision. None of that applies to a project that comes into an existing workspace (F1, F13, F14).
5. **The messages disagree with each other** (F11, F12, F15).

## Proposal

### One pipeline, two entry points

Make "bring a project in" one operation, with the steps of adopt today:

1. **Fetch**: a folder as it is, or a git repository cloned into a temporary folder.
2. **Identify**: the language plugins say what the project is. This is the same discovery that the workspace uses, so adopt and the workspace never disagree again (root cause 1).
3. **Place**: the folder by kind (`apps/`, `libs/`, `packages/`), asked, never picked.
4. **Split**: what goes to the workspace (dev tools, shared tsconfig options, lint) and what stays with the project. Files of a published package (README, LICENSE) stay with the project. A nested lockfile is dropped and merged by the next install.
5. **Join**: the language plugin joins the project to its workspace. This is a new plugin operation (`JoinWorkspace`) for each language: npm `workspaces`, a Composer path repository, `go.work use`, Cargo `members`, uv `members`, Maven `<modules>`, `dotnet sln add`.
6. **Tooling level**: keep the project's own tools, or use smith's (see D3).
7. **Check**: install, then doctor. The result is green, or it names each item that is left and its step.

The two entry points are:

- `smith adopt`: a new workspace around the repository, then the pipeline for each project in it.
- A new command (see D1) that runs the pipeline into the workspace that already exists.

### Join is also a doctor fix

Doctor gets one check, "a project is not joined to its language's workspace", with the fix of step 5. So a project that someone copied by hand is repaired with `smith doctor --fix` (F5). `smith install` counts the projects that it really installed, and names a project it could not reach (F3).

### Sources without a manifest

When no plugin finds a manifest, the plugins look at the source files (`*.java` under `src/`, `.classpath`, `*.php`, `*.c`). smith then offers to give the project a manifest of the language's default stack (Maven or Gradle, `composer.json`, CMake), with the sources kept where they are (F7, see D4).

### Placeholders and inferred tasks

- A script that is a known placeholder (the npm-init test line) gives no task. Doctor then says "no tests" as a warning, not as a failed run (F10).
- A project with a lint tool configured but no lint script gets a lint task from the plugin (F9).

### Message fixes

- adopt's final block becomes one list, "left to do", with no "bug" wording when the items are expected steps (F11).
- Agent files follow the adopt answer. Doctor offers them only when the workspace did not decline them (F12).
- `release --dry-run` points at `smith adopt --tags` when tags exist under another pattern (F15).

## Decisions for the owner

**D1. The command name**

1. `smith import <git-url|folder>` (recommended): a verb of its own, because it brings in an existing project, which is different from making one. It takes `--as app|lib|package`, `--into <folder>` and `--name`.
2. `smith add <kind> <name> --from <git-url|folder>`: one command for every project, but `add` already has many flags for new projects (stack, `--with`, environments) that do not apply here.
3. `smith adopt --into <workspace>`: it reuses the word, but adopt means "make a workspace" today.

**D2. Git history of an imported repository**

1. Keep it (recommended): `git subtree add --prefix <folder>`. `git log` and `blame` work in the new folder, and old tags can be mapped by `adopt --tags`.
2. Squash it into one commit, with the source URL and commit in the message.
3. Copy the files only. This is the only option for a plain folder that is not a git repository.

**D3. The default tooling level when a project comes in**

1. Ask for each project, and recommend smith's tooling when the project's tool is legacy (ESLint 8 `.eslintrc`, create-react-app) (recommended). This needs #553 for plain repositories.
2. Always keep the project's own tools (today's conservative level), and make doctor accept them as a choice, not a failure.
3. Always use smith's tooling.

**D4. A source folder without a manifest**

1. Offer to write the manifest of the language's default stack, and keep the sources in place (recommended).
2. Refuse, and say which manifest to write.

**D5. The scope of the first release of this work**

1. Node and PHP first (the owner's repositories), then Java (Maven and Gradle), then the other languages through the same `JoinWorkspace` operation (recommended).
2. Every language before the first release.

## Owner decisions (2026-11-20)

- D1: `smith import <git-url|folder>`.
- D2: keep the history with `git subtree add --prefix`. A folder that is not a git repository is copied.
- D3: ask for each project, and recommend smith's tooling when the project's tool is legacy.
- D4: offer to write the manifest of the language's default stack, and keep the sources in place.
- D5: the first release covers Node (JavaScript and TypeScript), Python and C# (.NET). PHP, Java and the other languages follow through the same `JoinWorkspace` operation.

## Proposed order of work

Tracked by the epic #1156: #1150 (step 1), #1151 (step 2), #1152 (step 3), #1153 (step 4), #553 (step 5), #1154 (step 6), #1155 (step 7).

1. One identification engine for adopt and the workspace (root cause 1). This fixes F6.
2. `JoinWorkspace` in the plugin protocol, for Node, Python and .NET. Add the doctor check and fix, and make the install count honest (F2 to F5).
3. The new command (D1), with fetch, history (D2), place, split and join.
4. Placeholders and the inferred lint task (F9, F10), and the message fixes (F11 to F15).
5. The tooling level for plain repositories (D3, #553).
6. Sources without a manifest (D4) for Node, Python and .NET, then PHP, Java and the other languages (D5).
7. A lab test: a fixed list of public sample repositories that CI adopts and imports, and that must end green. Today the "just works" goal has no test that holds it.
