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
smith release assets [--platform <platform>] [--project <app>] [--tag <project@version>]
```

## Arguments and flags

<!-- smith:begin release assets -->
Build a release's files with its build command and upload them to its GitHub Release

Usage: `smith release assets [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--platform <platform>` | string |  detected | the platform of this machine: windows-x64, macos-arm64, macos-x64, linux-x64 or linux-arm64 (default: detected) |
| `--project <project>` | string |  | build only this project |
| `--tag <tag>` | string |  | build for an existing release tag (project@version) from the current checkout, replacing its files |

<!-- smith:end release assets -->

The platform is detected unless given: `windows-x64`, `macos-arm64`, `macos-x64`, `linux-x64`
or `linux-arm64`. A project whose `platforms` leave it out is skipped.

A glob in `assets` may match nothing on a platform (a `.dmg` glob on Windows) when another glob
of the list matches. The run prints `no file matches <glob> on <platform>: skipped` and uploads
the rest. When no glob matches, it fails.

## System packages on Linux

The generated assets job installs the apt packages of `[ci] native_linux_packages` on its
Linux runners before it runs the build, so the build command does not need to.

## Several tasks into one folder

When the build command runs several smith tasks that write the same folder (a zip
task, then a `package-dmg` task, both into `dist/drop`), list that folder under
`[workspace] shared_outputs` in `smith.toml`. A task clears its output folders before it
runs, so without the setting the second task deletes the file the first one made. With it,
smith keeps what is in the folder. Without it, the assets run, `smith ci generate` and
`smith doctor` warn, name both tasks and say how to fix it.

## Run it again for a tag

`--tag <project@version>` builds that project's files from the **current checkout** and uploads
them to the existing GitHub Release of the tag, replacing files of the same name. It builds
the checkout, not the commit of the tag, because the usual reason to run again is a fix to
`smith.toml`, the build command or the workflow, and the commit of the tag still has the fault.
The version of the tag is stamped into the files. smith says so when HEAD is not the tag's
commit. The tag must be in the checkout (`git fetch --tags`) and its Release must exist. In
the generated GitHub workflow, run it from the Actions tab with the `assets_tag` input
(`mvd@0.11.1`): only the assets job runs.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | The files were built and uploaded, or the run was skipped: nothing was released by this commit, or no project needed building on this platform. |
| 1 | The platform is unknown, the build failed, no glob in `assets` matched a file, the `--tag` is unknown, or `gh release upload` failed (gh missing, not signed in, or no release for the tag). |

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
