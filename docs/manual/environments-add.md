<!-- lore-master
parent: environments.md
-->
# smith environments add

`smith environments add <project> <name>...` gives a web app more environments. For each name it
creates `.env.<name>` (starter `VITE_` placeholders and the warning that these values end up in the
browser), adds the script `build:<name>` to `package.json` (`vite build --mode <name> --outDir
<folder>`, into the folder the app's settings say) and records the new list in `smith.toml`. A file
that exists is never replaced.

## When to use it

You need another environment, such as `qa`. Names are lowercase words. The app must have been made
with environments (`smith add ... --environments`); for an app that was not, see
[An app that exists](environments.md#an-app-that-exists).

## What it asks

It shows the plan first: the files it creates, the script it adds and the line in `smith.toml`.
Without a terminal it refuses and names `--yes` (apply) and `--dry-run` (see the plan).

## Synopsis

```sh
smith environments add <project> <name>... [--dry-run] [--yes] [--commit]
```

## Arguments and flags

<!-- smith:begin environments add -->
Add environments to a web app

Usage: `smith environments add <project> <name...> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `project` | yes |  |
| `name...` | yes |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `--dry-run` | boolean | off | show the plan and change nothing |
| `-y, --yes` | boolean | off | do not ask for confirmation; without a terminal the change is made only with --yes |

<!-- smith:end environments add -->
