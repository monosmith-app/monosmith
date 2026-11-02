<!-- lore-master
parent: environments.md
-->
# smith environments list

`smith environments list` shows, for each web app that is built once per environment, its
environments, the task that builds each and which one is the default (the task `build`, which CI
and releases use).

## When to use it

To see which `smith run` command builds an environment, or whether an app has environments. To
change them, use [`smith environments add`](environments-add.md) or
[`remove`](environments-remove.md).

## Synopsis

```sh
smith environments list [project]
```

## Arguments and flags

<!-- smith:begin environments list -->
List the environments of the web apps and the task that builds each

Usage: `smith environments list [project] [--json]`

| Argument | Required | Meaning |
|---|---|---|
| `project` | no |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print the web apps and their environments as JSON |

`--json` prints a list with one entry for each web app: `project`, `envFiles`, `output`, `default`
and `environments`, a list of `name` and `task`. The VS Code extension reads it to offer the web
apps and the names that `environments remove` accepts.

<!-- smith:end environments list -->
