<!-- lore-master
parent: deps.md
-->
# smith deps up

`smith deps up` is `smith up`: it shows the dependencies with newer releases and updates the
ones you pick, in every project that declares them. The full description, the exit codes and
examples are on [`smith up`](up.md); the two commands share one implementation and one set of
flags.

## Synopsis

```sh
smith deps up [package]... [--yes] [--check] [--json] [--no-install] [--align]
```

## Arguments and flags

<!-- smith:begin deps up -->
Show dependencies with newer releases and update the ones you pick (same as deps up; --align is sync)

Usage: `smith deps up [packages...] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `packages...` | no | packages to update; none lists the updates |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--align` | boolean | off | bring dependencies declared at different versions to one (the named ones, or all); with --check, fail when any differ |
| `--check` | boolean | off | report only; never prompt, never write |
| `--json` | boolean | off | print the updates and tools as JSON; never write |
| `--no-install` | boolean | off | rewrite the manifests without installing |
| `-y, --yes` | boolean | off | take every update without prompting |

<!-- smith:end deps up -->

## Example

```sh
smith deps up --check
```

prints the same report as `smith up --check`; see [the example there](up.md#examples).

## Related

[`smith up`](up.md), [`smith sync`](sync.md).
