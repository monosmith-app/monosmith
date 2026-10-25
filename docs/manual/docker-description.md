<!-- lore-master
parent: docker.md
-->
# smith docker description

`smith docker description` sets the Docker Hub repository descriptions of each app whose
registries include a `docker.io` entry: the full description from the app's
`description_from` file, and the short one (the line under the image's name) from
`short_description` or, when that is unset, the project's own description. It signs in with
the registry's credentials, `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN`, or the variables
the registry entry names. Editing a description needs a token with Read, Write & Delete
permission.

The call is an idempotent PATCH, so it can run on every push to the release branch. The
generated pipeline does so after the release, which lets a docs-only commit that changes
the file reach Docker Hub without a new version. A failure is a warning (the images are
already published); `--strict` makes it an error.

## When to use it

Rarely by hand: use `--dry-run` to see which repositories it would touch. The pipeline runs
it.

## Synopsis

```sh
smith docker description [--dry-run] [-p <app>]... [--strict]
```

## Arguments and flags

<!-- smith:begin docker description -->
Set the Docker Hub descriptions of the apps' images

Usage: `smith docker description [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--dry-run` | boolean | off | name the repositories without calling Docker Hub |
| `-p, --project <project>` | string, repeatable |  | only these apps |
| `--strict` | boolean | off | exit non-zero when a description could not be set |

<!-- smith:end docker description -->

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Every description was set, none was needed, or a failure occurred without `--strict` (a warning is printed). |
| 1 | With `--strict`, a description could not be set. |

## Files and environment

Reads the `description_from` file and the registry settings from `smith.toml`. Calls
Docker Hub's API. Variables: `DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN` (or the ones the entry
names).

## Examples

In a workspace with no `docker.io` registry configured:

```sh
smith docker description
```

```text
no app has a description_from file or a short description for a docker.io registry
```

## Related

[`smith docker`](docker.md), [`smith release`](release.md), [Releasing](../reference/releasing.md).
