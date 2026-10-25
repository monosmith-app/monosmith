# smith pipeline

`smith pipeline` adds a ready-made pipeline file beside the generated CI. It starts the way
every generated job does (the toolchains, smith, `smith ci setup`) and then does one thing.
The file is written once and is yours from then on: [`smith upgrade`](upgrade.md) never
rewrites it, and an existing file is kept unless `--force`. With no template it lists the
templates.

| Template | Providers | Does |
|---|---|---|
| `e2e` | github, azure | Runs every project's `e2e` task, one at a time, on a push to the release branch, on pull requests (GitHub) and on demand. |
| `package-zip` | github, azure | Packs the projects (`smith ci pack`) and keeps `dist/drop` as a build artifact, the input of a classic release pipeline. |
| `deploy-pages` (`--project <app>`) | github | Builds an app and deploys its build output to GitHub Pages. |
| `deploy-azure-function` (`--project <app>`) | github, azure | Builds a function app and deploys it to an Azure Function App. |

## When to use it

- You need a pipeline the generated CI does not cover (end-to-end tests, a deploy).

If the pipeline should be regenerated with the workspace, it is not this command's job;
see [`smith ci generate`](ci-generate.md).

## Synopsis

```sh
smith pipeline [template] [--ci github|azure|both] [--project <app>] [--force]
```

## Arguments and flags

<!-- smith:begin pipeline -->
Add a ready-made pipeline: e2e, package-zip, deploy-pages, deploy-azure-function

Usage: `smith pipeline [template] [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `template` | no | the template to add; none lists them |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--ci <ci>` | `github`, `azure` or `both` | the workspace's providers | github, azure or both (default: the workspace's providers) |
| `--force` | boolean | off | replace a pipeline file that exists |
| `--project <project>` | string |  | the app a deploy template builds |

<!-- smith:end pipeline -->

- `--project` is required by the two deploy templates and names the app that is built.
- `--ci` picks the provider (`github`, `azure` or `both`); the default is the workspace's
  `ci.providers`. The table above shows which providers each template supports.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The template was written (or listed). |
| 1 | Unknown template, a missing `--project`, or a file exists (without `--force`). |

## Files and environment

Writes `.github/workflows/<template>.yml` and/or the Azure file. The generated jobs read
the secrets the template names; [`smith ci check`](ci-check.md) lists them.

## Examples

```sh
smith pipeline
```

```text
e2e                                      github, azure  run every project's e2e task, one at a time, on a push to the release branch, on pull requests (GitHub) and on demand
package-zip                              github, azure  pack the projects (smith ci pack) and keep dist/drop as a build artifact, the input of a classic release pipeline
deploy-pages (--project <app>)           github         build an app and deploy its build output to GitHub Pages
deploy-azure-function (--project <app>)  github, azure  build a function app and deploy it to an Azure Function App
```

```sh
smith pipeline e2e
```

```text
✔ wrote .github/workflows/e2e.yml
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `deploy-pages builds one app: pass --project <app>` | Add `--project web`. |
| A refusal naming an existing file | Pass `--force` to replace it, or edit the file. (Not captured.) |

## Related

[`smith ci`](ci.md), [`smith ci generate`](ci-generate.md), [CI pipelines](../reference/ci.md).
