<!-- lore-master
parent: release.md
-->
# smith release pub

`smith release pub` builds a Dart app's pub.dev package from the archives on its GitHub
Release and publishes it with `dart pub publish`, unless pub.dev already has that
version. The release starts the generated `.github/workflows/pub.yml` workflow on the app's
pub tag, and that workflow runs this command: pub.dev's automated publishing accepts only
such a run.

## When to use it

Almost never by hand. It is the step of the generated pub workflow; run it locally only
to rehearse the failure messages. See [Releasing](../reference/releasing.md) for how a Dart
app is released.

## Synopsis

```sh
smith release pub --project <app> --version <version>
```

## Arguments and flags

<!-- smith:begin release pub -->
Publish a released app to pub.dev from its GitHub Release (the generated pub workflow runs it)

Usage: `smith release pub [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--project <project>` | string |  | the released app |
| `--version <version>` | string |  | its released version |

<!-- smith:end release pub -->

Both flags are required.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The package was published, or pub.dev already had the version. |
| 1 | A flag is missing, the GitHub Release or its archives cannot be found, or `dart pub publish` failed. |

## Files and environment

Downloads the release archives, builds the package in a temporary folder, and publishes
through `dart pub publish`. It needs the GitHub token of the workflow and pub.dev's
automated publishing set up for the repository. Not exercised for this page beyond the
missing-flag error.

## Examples

```sh
smith release pub
```

```text
smith: --project and --version are required
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `--project and --version are required` | Pass both, as the workflow does. |

## Related

[`smith release`](release.md), [Releasing](../reference/releasing.md).
