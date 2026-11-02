<!-- lore-master
parent: environments.md
-->
# smith environments remove

`smith environments remove <project> <name>...` takes environments away from a web app: it removes
the script `build:<name>` from `package.json` and the name from the list in `smith.toml`. It leaves
`.env.<name>` and the folders of old builds: the file holds your values, and smith deletes nothing
it was not asked to delete. The plan says so, and the command repeats it after it ran.

## When to use it

You do not build an environment any more. The default environment cannot be removed, because it is
the task `build`: make another environment the default first (`default_environment` in
`[projects.<name>.settings.vite]`, and the script of `build`).

## Synopsis

```sh
smith environments remove <project> <name>... [--dry-run] [--yes] [--commit]
```

## Arguments and flags

<!-- smith:begin environments remove -->
Remove environments from a web app (its .env files stay)

Usage: `smith environments remove <project> <name...> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `project` | yes |  |
| `name...` | yes |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--commit` | boolean | off | commit the change as one operation so smith undo can revert it (nothing is committed without it) |
| `--dry-run` | boolean | off | show the plan and change nothing |
| `-y, --yes` | boolean | off | do not ask for confirmation |

<!-- smith:end environments remove -->
