# smith deps

`smith deps` is a group for dependency work that spans projects. Its only subcommand today
is `smith deps up`, which is [`smith up`](up.md) under another name; `smith deps up --align`
brings a dependency that projects declare at different versions to one, as
[`smith sync`](sync.md) does. Running `smith deps` alone prints its help.

## Subcommands

<!-- smith:begin deps -->
Group for dependency work across projects; deps up is the same as up

Usage: `smith deps <command>`

| Subcommand | Meaning |
|---|---|
| `smith deps up` | Show dependencies with newer releases and update the ones you pick (same as deps up; --align is sync) |

No flags.

<!-- smith:end deps -->

See [`smith deps up`](deps-up.md).

## Exit codes

0 for the help; see the subcommand for the rest.

## Files and environment

None of its own.

## Examples

```sh
smith deps
```

prints the help of the group:

```text
Dependency work that spans projects. smith deps up is smith up: it updates every project that declares a package in one change. ...

Usage:
  smith deps [command]

Available Commands:
  up          Show dependencies with newer releases and update the ones you pick (same as deps up; --align is sync)
```

## Related

[`smith up`](up.md), [`smith sync`](sync.md), [`smith install`](install.md).
