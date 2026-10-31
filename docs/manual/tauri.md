# smith tauri

`smith tauri` is a group for the Tauri desktop apps of a workspace. Its two subcommands are
[`smith tauri add`](tauri-add.md), which makes a Node app that already exists a Tauri app, and
[`smith tauri icon`](tauri-icon.md), which makes the app's icons from a logo. A new Tauri app
needs neither: `smith add app studio --stack react-tauri` makes one (see [the tauri
plugin](../plugins/using/catalogue/tauri.md)).

## Subcommands

<!-- smith:begin tauri -->
Add Tauri to a Node app, or make its icons

Usage: `smith tauri <command>`

| Subcommand | Meaning |
|---|---|
| `smith tauri add` | Make an existing Node app a Tauri desktop app |
| `smith tauri icon` | Make the app's icons from a square PNG |

No flags.

<!-- smith:end tauri -->

## Exit codes

0 for the help; see the subcommand for the rest.

## Files and environment

None of its own.

## Examples

```sh
smith tauri
```

prints the group's help with its two subcommands.

## Related

[`smith tauri add`](tauri-add.md), [`smith tauri icon`](tauri-icon.md), [`smith add`](add.md),
[the tauri plugin](../plugins/using/catalogue/tauri.md).
