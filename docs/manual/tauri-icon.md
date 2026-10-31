# smith tauri icon

`smith tauri icon` makes every icon of a Tauri app from one square PNG (1024x1024 is best). It
is the Tauri CLI's `icon` command, the project's own `@tauri-apps/cli`, run in the app's folder
with the logo's absolute path. It overwrites `src-tauri/icons/`, replacing the placeholders the
scaffold writes so the first build runs.

## Synopsis

```sh
smith tauri icon <logo.png> [-p <project>]
```

## Arguments and flags

<!-- smith:begin tauri icon -->
Make the app's icons from a square PNG

Usage: `smith tauri icon <logo.png> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `logo.png` | yes |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `-p, --project <project>` | string |  | the app, when more than one project has a src-tauri |

<!-- smith:end tauri icon -->

The app is the only project with a `src-tauri`, or the one `-p` names.

## Errors

| Message | Meaning |
|---|---|
| `<logo> is not a file` | Give the path of a square PNG. |
| `no project has a src-tauri` | `smith tauri add <project>` makes one. |
| `several projects have a src-tauri` | Name one with `-p`. |
| `the Tauri CLI could not make the icons` | `@tauri-apps/cli` is not installed: `smith install`. |

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The icons were made. |
| 1 | A problem, named with the step that clears it. |

## Files and environment

Reads the logo; writes `<project>/src-tauri/icons/`.

## Examples

```sh
smith tauri icon logo.png -p studio
```

## Related

[`smith tauri`](tauri.md), [`smith tauri add`](tauri-add.md),
[the tauri plugin](../plugins/using/catalogue/tauri.md#icons).
