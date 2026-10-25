# smith docker

`smith docker` is a group for Docker image commands. It has one subcommand today,
[`smith docker description`](docker-description.md), which sets the Docker Hub
descriptions of an app's images. This group does not build or push images; that is the docker deploy plugin's work during
a release (see [Releasing](../reference/releasing.md)).

## Subcommands

<!-- smith:begin docker -->
Docker image commands

Usage: `smith docker <command>`

| Subcommand | Meaning |
|---|---|
| `smith docker description` | Set the Docker Hub descriptions of the apps' images |

No flags.

<!-- smith:end docker -->

## Exit codes

0 for the help; see the subcommand for the rest.

## Files and environment

None of its own.

## Examples

```sh
smith docker
```

prints the group's help with its one subcommand.

## Related

[`smith docker description`](docker-description.md), [`smith release`](release.md).
