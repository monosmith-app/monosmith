<!-- lore-master
parent: ci.md
-->
# smith ci check

`smith ci check` lists what you must set up by hand before the generated pipelines can
work: repository secrets, and each private registry's permissions. It reads the workspace
and the plugins' CI contributions and changes nothing.

## When to use it

After `smith new`, `smith add` or a change to the release settings, before the first run
of the pipeline on a new repository.

## Synopsis

```sh
smith ci check
```

## Arguments and flags

<!-- smith:begin ci check -->
List what to set up by hand for the pipelines: secrets, and each private registry's permissions

Usage: `smith ci check`

No flags.

<!-- smith:end ci check -->

## Exit codes

0 always when the workspace loads; 1 outside a workspace.

## Files and environment

Reads `smith.toml` and the project graph. Writes nothing.

## Examples

```sh
smith ci check
```

```text
To set up by hand before the pipelines run:
  - GitHub: add the repository secrets PYPI_TOKEN (Settings → Secrets and variables → Actions)
```

With nothing to do it prints `nothing to set up by hand`.

## Related

[`smith ci`](ci.md), [`smith release`](release.md), [CI pipelines](../reference/ci.md).
