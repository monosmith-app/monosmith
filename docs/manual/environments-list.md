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

Usage: `smith environments list [project]`

| Argument | Required | Meaning |
|---|---|---|
| `project` | no |  |

No flags.

<!-- smith:end environments list -->
