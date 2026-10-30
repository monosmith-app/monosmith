# smith release

`smith release` versions, publishes and tags every package that has releasable
conventional commits since its last tag. The bump follows the commits: `feat` is a minor
bump, `fix` and `perf` a patch, a breaking change a major (a minor before 1.0). Packages
release in dependency order. Each is stamped with its new version, built, published and
tagged; manifests are restored afterwards, because the tag is the record. A re-run after
a failure resumes where it stopped.

Before the first publish, every npm, PyPI, NuGet and RubyGems registry is asked whether
it accepts the credential and the package names, so a bad key stops the release before
anything is published. What was published where is kept in
`dist/release/<project>/journal.json`; `smith release --status` prints it.

CI runs the release through [`smith ci release`](ci.md#release) on a push to the release
branch. Running `smith release` by hand is the same thing from your machine.

## When to use it

- `smith release --dry-run` locally or on a pull request, to see what would be released.
- `smith release` from the release branch when you publish by hand.
- `smith release --status` to see what a failed release already published.

Do not use it to cut a release from a feature branch: `releases run from main, not
<branch>` (see the environment variables). Versions come from commit messages; there is no
flag to set one.

## Synopsis

```sh
smith release [--dry-run [--comment]] [--notes] [--format text|markdown] [--json] [--verbose] [--no-push] [--status] [-y]
smith release pub --project <app> --version <version>
```

## Arguments and flags

<!-- smith:begin release -->
Version, publish and tag every package with releasable commits

Usage: `smith release <command> [flags]`

| Subcommand | Meaning |
|---|---|
| `smith release assets` | Build a release's files with its build command and upload them to its GitHub Release |
| `smith release pub` | Publish a released app to pub.dev from its GitHub Release (the generated pub workflow runs it) |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--comment` | boolean | off | with --dry-run in a pull request build, post the preview on the pull request (GitHub, Azure Repos), updating smith's earlier comment |
| `--dry-run` | boolean | off | show what would be released; change nothing |
| `--format <format>` | `text` or `markdown` | `text` | how the result is printed: text, or markdown (what the CI job summary and --comment use) |
| `--json` | boolean | off | print the release report as JSON |
| `--no-push` | boolean | off | create tags locally only |
| `--notes` | boolean | off | also print each release's notes (what its GitHub Release says) |
| `--status` | boolean | off | show the release journal: what was published where (dist/release/<project>/journal.json) |
| `--verbose` | boolean | off | name every unchanged package |
| `-y, --yes` | boolean | off | do not ask for confirmation |

<!-- smith:end release -->

The subcommand [`smith release pub`](release-pub.md) publishes a Dart app to pub.dev;
[`smith release assets`](release-assets.md) builds the files of an app whose `[release]` has a
`build` command and uploads them to its GitHub Release.

Interactions:

- `--dry-run` computes the plan and prints it; it publishes nothing, tags nothing and does
  not need credentials. `--notes` adds each release's notes (what its GitHub Release will
  say). `--comment` needs `--dry-run` (`--comment posts a preview: use it with --dry-run`)
  and posts the markdown preview on the pull request, updating smith's earlier comment
  (GitHub and Azure Repos).
- `--format markdown` is what the CI job summary and `--comment` use. In the trial the
  plan line was printed first in text and the markdown table after it.
- `--json` prints the report as JSON (`dryRun`, `started`, `finished`, `results` with
  `project`, `next`, `level`, `tag`, `reason`, `commits`, `changes`, `status`, `files`,
  `published`, `registry`, `pushed`).
- `--status` prints the release journal and exits; it ignores the other flags except
  `--json`.
- `--no-push` creates the tags locally only.
- Without a terminal a real release must be confirmed with `-y`/`--yes` (CI passes it).
- `--verbose` names the unchanged packages.

A real release refuses to start (`cannot release:` followed by one line per problem) when:
the clone is shallow; the current branch is not `[release] branch` (default `main`);
tracked files have uncommitted changes; a publish credential for a package in the plan is
missing (`NODE_AUTH_TOKEN` or `NPM_TOKEN`, `PYPI_TOKEN`, `NUGET_API_KEY`, `GEM_HOST_API_KEY`,
`VSCE_PAT`, `GITHUB_TOKEN` when `[release] github_releases` is on, and what a deploy plugin
such as docker needs).

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Released, or the dry run printed, or there was nothing to release. |
| 1 | A preflight problem, a wrong release tag, a failed publish, a declined confirmation, or `--comment` without `--dry-run`. |

## Files and environment

Reads git history and tags, project manifests and `[release]` in `smith.toml`. Writes
version stamps temporarily, `dist/release/<project>/` (journal, built assets, and a
CycloneDX SBOM for apps and packages with release assets, Go and npm only), git tags, and
pushes them to `[release] remote` (default `origin`). With `github_releases` it creates a
GitHub Release per tag and attaches the assets and the SBOM. The generated GitHub workflow
adds a build-provenance attestation step for `dist/release/*/*` and npm publishes with
provenance when an OIDC token is present.

| Variable | Used for |
|---|---|
| `SMITH_RELEASE_ANY_BRANCH` | Any value lets a real release run from a branch other than the release branch. |
| `NODE_AUTH_TOKEN`, `NPM_TOKEN`, `SMITH_NPM_ALREADY_AUTHENTICATED` | npm credential (the last says the run signed in already). |
| `PYPI_TOKEN`, `TWINE_PASSWORD` | PyPI. |
| `NUGET_API_KEY`, `GEM_HOST_API_KEY` | NuGet, RubyGems. |
| `GITHUB_TOKEN` | GitHub Releases and the pull request comment. |
| `GITHUB_STEP_SUMMARY` | When set, CI phases append their outcome to the job summary. |

The full list and the registry settings are in [Releasing](../reference/releasing.md).

## Examples

A package with a `feat` commit since its last release:

```sh
smith release --dry-run
```

```text
Dry run: nothing is published or tagged.
  tools                unreleased → 0.1.0  (1 releasing commit(s))


→ Push to the release branch and CI releases these; or run smith release here to publish them yourself.
```

With the notes:

```sh
smith release --dry-run --notes
```

```text
Dry run: nothing is published or tagged.
  tools                unreleased → 0.1.0  (1 releasing commit(s))

# tools 0.1.0

## Features

- **tools:** add a greeting (dfd8640)

## Thanks

Thanks to Ada Example.


→ Push to the release branch and CI releases these; or run smith release here to publish them yourself.
```

The markdown form, as it appears in a pull request comment:

```sh
smith release --dry-run --format markdown
```

```text
### Release preview

| Package | Version | Why | Status |
|---|---|---|---|
| tools | new → **0.1.0** | 1 releasing commit(s) | planned |
```

When no commit releases anything:

```text
Nothing to release: no package has a feat: or fix: commit since its last release (or since the start, if it was never released).
  → Releases come from conventional commits: commit with a message like "feat: add search" or "fix: handle empty input" ...
Unchanged: 1 packages (--verbose names them)
```

`smith release --status` before anything was published:

```text
No release journal yet: nothing has been published from here.
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `refusing to release without a terminal; pass --yes (CI does)` | Add `-y`. |
| `release cancelled` | You did not answer `y` to `Release these packages? [y/N]`. |
| `cannot release:` then `releases run from main, not <branch>` | Switch to the release branch, or set `[release] branch`. |
| `... the clone is shallow, so versions cannot be computed from history: fetch with full depth (actions/checkout fetch-depth: 0)` | Fetch the whole history. |
| `... the working tree has uncommitted changes (...); commit or stash them first` | Commit or stash; generated files that a build rewrites make this happen (see `tracked files are not generated files` in `smith doctor`). |
| `<ENV> is not set: <what it must be>` | Set the secret; `smith ci check` lists them for the pipelines. |
| `cannot release: the release tags would give wrong versions, so no plan is shown` | A tag is behind the registry or under the wrong prefix. `smith doctor` names the check. |
| `--comment posts a preview: use it with --dry-run` | Add `--dry-run`. |
| `SBOM for <tag>: ... (re-run the release to finish it)` / `GitHub Release <tag>: ... (re-run the release to finish it)` | A late step failed; run `smith release` again, it resumes. |

## Related

[`smith release pub`](release-pub.md), [`smith ci`](ci.md), [`smith doctor`](doctor.md),
[`smith pipeline`](pipeline.md), [Releasing](../reference/releasing.md).
