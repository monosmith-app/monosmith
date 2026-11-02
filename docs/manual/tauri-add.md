# smith tauri add

`smith tauri add` makes a Node app that already exists, such as a Vite React app, a Tauri
desktop app, as if it had been made with the `react-tauri` stack. It never overwrites a file that
exists, so running it again is safe.

## What it does

1. Checks the project is a Node app with an `index.html` (the web page Tauri wraps).
2. Gives the project a stack with `tauri` added last to its toolchain, so its own ESLint, test
   and web-build tooling stay as they are. A stack that already has that make-up is reused (`react`
   becomes `react-tauri`); otherwise `<stack>-tauri` is declared in `smith.toml` (`<project>-tauri`
   when that name is taken, `node-tauri` for a project with no stack) and the project's
   `[projects.<name>] stack` names it. A custom stack is the choice because `react-tauri` is only
   right for a project built with exactly the `react` toolchain; derived from the project's own
   stack, the other tools it uses (Jest, Playwright, a structure plugin) are not lost.
3. Has the tauri plugin write `src-tauri` as the scaffold does: `Cargo.toml` with its own empty
   `[workspace]`, `tauri.conf.json` whose `frontendDist` is the folder smith builds the web app
   into, `build.rs`, `src/main.rs`, `src/lib.rs`, `capabilities/default.json`, generated icons and a
   `.gitignore`.
4. Adds `@tauri-apps/cli` to the root `devDependencies` (where smith keeps dev tools, once for
   the workspace), `@tauri-apps/api` to the project's dependencies and a `tauri` script to its
   scripts.
5. Lists `src-tauri` under `exclude` in a root Cargo workspace whose `members` would take it in.
6. Installs the dependencies and regenerates the CI pipelines, which gain the Rust toolchain.

It prints each file it wrote or edited and the commands to run next. The files stay uncommitted.

## What smith asks

`smith tauri add` shows the plan (the Rust files, the two `package.json` edits, `smith.toml`, the
install and the pipelines) and asks the bundle identifier of the app, like
`com.yourcompany.yourapp`. The operating system and the app stores know the app by that name and
it is hard to change after you ship. Without a terminal smith needs `--identifier` or `--yes`
(which takes `com.<workspace>.<project>` and says so). `--dry-run` shows the plan.

## Synopsis

```sh
smith tauri add <project> [--identifier <id>] [--dry-run] [--yes] [--json]
```

## Arguments and flags

<!-- smith:begin tauri add -->
Make an existing Node app a Tauri desktop app

Usage: `smith tauri add <project> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `project` | yes |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `--dry-run` | boolean | off | show what would be written and change nothing |
| `--identifier <identifier>` | string |  | the bundle identifier of the app, like com.yourcompany.yourapp (smith asks when you leave it out) |
| `--json` | boolean | off | print what was done as JSON |
| `-y, --yes` | boolean | off | write without asking; the bundle identifier is com.<workspace>.<project> unless --identifier says otherwise |

<!-- smith:end tauri add -->

## Errors

| Message | Meaning |
|---|---|
| `the workspace has no project <name>` | `smith projects` lists the projects. |
| `<name> is a <language> project; Tauri wraps a Node app` | Only Node apps can be wrapped. |
| `<name> is a lib, not an app` | `smith kind <name> app` makes it one. |
| `<dir> has no index.html` | There is no web page to wrap. Start from a web app, or write `src-tauri` with `npx tauri init`. |

When it fails, `smith.toml` is put back as it was.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The project is a Tauri app. |
| 1 | A problem, named with the step that clears it. |

## Files and environment

Edits `smith.toml`, the project's `package.json`, the root `package.json` and its lockfile,
the root `Cargo.toml` when it has a workspace, and the CI pipelines; writes `<project>/src-tauri/`.

## Examples

```sh
smith tauri add studio
```

```text
Added Tauri to studio
  stack   react -> react-tauri (the stack with that make-up)
  writes  Cargo.toml
  writes  apps/studio/package.json
  writes  apps/studio/src-tauri/Cargo.toml
  ...
Next:
  smith doctor                             check the machine and the config
  smith run tauri-dev -p studio            run the app in development mode
  smith run tauri-build -p studio          build the installers
  smith tauri icon logo.png -p studio      replace the placeholder icons
```

## Related

[`smith tauri`](tauri.md), [`smith add`](add.md), [`smith kind`](kind.md),
[the tauri plugin](../plugins/using/catalogue/tauri.md).
