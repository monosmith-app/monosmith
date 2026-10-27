<!-- lore-master
parent: ci.md
-->
# smith ci generate

`smith ci generate` rewrites the CI pipeline files from the workspace's projects and
settings: `.github/workflows/ci.yml` for GitHub, `azure-pipelines.yml` for Azure,
`.gitlab-ci.yml` for GitLab, `.circleci/config.yml` for CircleCI (per `ci.providers`), and `.devcontainer/devcontainer.json`
when `ci.devcontainer` is on. The pipelines install smith with `npm`, `pip` or `dotnet tool` when the workspace has a
Node, Python or .NET toolchain (in that order of preference). For any other workspace it also writes
`tools/smith/install.sh`, the installer the pipelines run (`sh tools/smith/install.sh "<version>"`) instead of
fetching a script from GitHub at every run; commit it, and do not edit it (`smith doctor --fix` restores it).
A generated copy in a workspace that no longer needs it is deleted. Steps you added in `smith:slot` markers are kept; a pipeline
for a provider the workspace no longer asks for is deleted.

`smith add` and `smith remove` run it for you. Run it yourself after editing `smith.toml`
by hand or changing `ci.*` with [`smith config set`](config-set.md). `smith upgrade`
regenerates these files too, together with the rest of what smith owns.

## Synopsis

```sh
smith ci generate [--check]
```

## Arguments and flags

<!-- smith:begin ci generate -->
Rewrite the CI pipelines from the workspace's projects (--check only reports drift)

Usage: `smith ci generate [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--check` | boolean | off | report drift without writing |

<!-- smith:end ci generate -->

`--check` writes nothing; it compares what would be generated with what is on disk and
fails when they differ. [`smith doctor`](doctor.md) runs the same comparison as its
`CI pipelines match the workspace` check.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Files were written, or they were already current. |
| 1 | `--check` found drift, or generation failed. |

## Files and environment

Writes the files above. A slot that no longer exists in the new pipeline cannot keep its
content: `warning: slot no longer exists, its content was not kept: <slot>` goes to
standard error. When a configured provider cannot express a setting (CircleCI builds native
projects on Linux only and ignores `ci.branches`), `note: <what is left out>` goes to standard
error as well, and [`smith ci check`](ci-check.md) lists it. Reads no smith-specific environment variable; the generated files call
`smith` and set `SMITH_VERSION`.

## Examples

```sh
smith ci generate --check
```

```text
pipelines are up to date
```

After `ci.devcontainer` was removed (so the dev container is no longer configured) the check
fails with exit code 1:

```text
smith: pipelines are out of date: .devcontainer/devcontainer.json is no longer configured — run `smith ci generate`
```

```sh
smith ci generate
```

```text
wrote .devcontainer/devcontainer.json
```

(In that trial `wrote` was printed for a file that was an obsolete dev container; the
message names the file that changed, not what was done to it.)

## Common errors

| Message | Cause and remedy |
|---|---|
| `pipelines are out of date: <files> — run smith ci generate` | Run it and commit the result. |

## Related

[`smith ci`](ci.md), [`smith pipeline`](pipeline.md), [`smith upgrade`](upgrade.md),
[CI pipelines](../reference/ci.md).
