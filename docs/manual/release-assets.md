<!-- lore-master
parent: release.md
-->
# smith release assets

`smith release assets` runs the `build` command of every app released by the current commit
whose `[projects.<name>.release]` has one, on this machine, and uploads the files matching
its `assets` globs to the app's GitHub Release with `gh release upload --clobber`. It is the
step the generated GitHub Actions workflow runs on one runner per platform (as
`smith ci assets`); run it by hand to build the files of the platform you are on.

## When to use it

To rehearse the build, or to add this machine's platform to a release whose runners cannot
build it. The release itself (`smith release`) creates the tag and the GitHub Release without
these files. See [Releasing](../reference/releasing.md) for the `release` table.

## Synopsis

```sh
smith release assets [--platform <platform>] [--project <app>]
```

## Arguments and flags

<!-- smith:begin release assets -->
Build a release's files with its build command and upload them to its GitHub Release

Usage: `smith release assets [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--platform <platform>` | string |  detected | the platform of this machine: windows-x64, macos-arm64, macos-x64, linux-x64 or linux-arm64 (default: detected) |
| `--project <project>` | string |  | build only this project |

<!-- smith:end release assets -->

The platform is detected unless given: `windows-x64`, `macos-arm64`, `macos-x64`, `linux-x64`
or `linux-arm64`. A project whose `platforms` leave it out is skipped.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The files were built and uploaded, or no project needed building on this platform. |
| 1 | The platform is unknown, the build failed, a glob in `assets` matched no file, or `gh release upload` failed (gh missing, not signed in, or no release for the tag). |

## Files and environment

Runs the command with `sh -c` (`bash -c` on Windows, else `cmd /C`) from the workspace root,
with `SMITH_PLATFORM`, `SMITH_VERSION` and `SMITH_TAG` added to the environment. `gh` needs
`GH_TOKEN` or `GITHUB_TOKEN` with `contents: write`, or a signed-in session.

## Examples

```sh
smith release assets --platform linux-x64
```

```text
building music-fixer 1.2.0 for linux-x64: vcw app build examples/apps/music-fixer/app.json --out out --zip
uploaded 2 file(s) to music-fixer@1.2.0
```

## Common errors

| Message | Cause and remedy |
|---|---|
| `"x" is not a platform smith builds release files for` | Use one of the five platform names. |
| `the build left no file matching assets "out/*.zip"` | Make the command leave the file, or fix the glob. |
| `<app> was not released by this commit; nothing to build` | HEAD carries no release tag of the app; check out the tagged commit. |

## Related

[`smith release`](release.md), [`smith ci assets`](ci.md), [Releasing](../reference/releasing.md).
