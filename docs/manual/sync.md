# smith sync

`smith sync` makes every project declare a shared dependency at the same spec, language by
language. The target is the version installed, with the operator most declarations use;
when nothing says what is installed, the highest version declared. Peer dependencies and the
workspace's own packages are left alone, and so is a spec smith cannot rewrite safely (a
URL, a git ref, a compound range): `sync` reports those. After rewriting it installs, so
lockfiles and the venv follow.

`smith sync` is `smith up --align` with no package names.

## When to use it

- `smith doctor` reports `every production dependency has one version across the workspace`
  as failing.
- In CI, as `smith sync --check`, to fail a pull request that lets versions drift.

A dependency that may differ between projects is listed in `smith.toml` under
`[dependencies] allow = [...]`; sync leaves it alone.

## Synopsis

```sh
smith sync [--check] [--no-install] [--json]
```

## Arguments and flags

<!-- smith:begin sync -->
Make every project declare a shared dependency at the same version (same as up --align)

Usage: `smith sync [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--check` | boolean | off | report only; fail when declarations disagree |
| `--json` | boolean | off | print the report as JSON |
| `--no-install` | boolean | off | rewrite the manifests without installing |
| `-y, --yes` | boolean | off | also rewrite declarations whose major version differs from the target |

<!-- smith:end sync -->

With `commit_operations = true` under `[workspace]` and a clean git work tree, a run that rewrites
declarations is one commit (`chore: align dependency versions`); [`smith undo`](undo.md) reverts it.

## Major versions

A declaration on another major version than the target (zod 3 when 4 is installed) is left alone, and
the report says so. Pass `--yes` to rewrite it too. `smith up --yes` prints a warning line that names
the major updates it takes.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Everything agrees, or the declarations were rewritten. |
| 1 | `--check` found declarations that disagree, or a rewrite or install failed. |

## Files and environment

Rewrites manifests and runs the installer unless `--no-install`. Reads `[dependencies]`
in `smith.toml`. The manifests it rewrites, by language:

- Node: `package.json`. Python: `pyproject.toml`.
- Rust: `Cargo.toml` (crates and the root `[workspace.dependencies]`).
- Go: `go.mod`, through `go get` (so `go.sum` follows).
- .NET: the project files, `Directory.Build.props` and the central `Directory.Packages.props`.
- Ruby: gemspecs and the root `Gemfile`.
- PHP: `composer.json` (`require` and `require-dev`).
- Dart: `pubspec.yaml` (`dependencies` and `dev_dependencies`).
- JVM: `pom.xml` (dependencies, dependency management and their properties) and the string notation of Gradle build scripts.

## Examples

```sh
smith sync --check
```

```text
✔ Every project agrees on every shared dependency version.
⊘ cpp: not compared; its plugin does not list declarations
```

## Common errors

A disagreement is reported with the packages and their declarations; the output was not
reproduced here because the demonstration workspace never disagreed.

## Related

[`smith up`](up.md), [`smith install`](install.md), [`smith doctor`](doctor.md),
[Configuration](../reference/configuration.md).
