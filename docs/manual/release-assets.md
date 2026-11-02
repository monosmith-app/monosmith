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
smith release assets [--platform <platform>] [--project <app>] [--tag <project@version>] [--source current|tag]
```

## Arguments and flags

<!-- smith:begin release assets -->
Build a release's files with its build command and upload them to its GitHub Release

Usage: `smith release assets [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--platform <platform>` | string |  detected | the platform of this machine: windows-x64, macos-arm64, macos-x64, linux-x64 or linux-arm64 (default: detected) |
| `--project <project>` | string |  | build only this project |
| `--source <source>` | string |  | with --tag, the code to build when the tag is not at HEAD: current (the code as it is now, named with the tag's version) or tag (the tag's own commit); default: ask in a terminal, refuse without one |
| `--tag <tag>` | string |  | build for an existing release tag (project@version), replacing its files |

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

`--tag <project@version>` builds that project's files and uploads them to the existing GitHub
Release of the tag, replacing files of the same name. The version of the tag is stamped into the
files. The tag must be in the checkout (`git fetch --tags`) and its Release must exist.

When the tag is not at HEAD, you choose which code to build with `--source`:

| `--source` | What it builds | Consequence |
|---|---|---|
| `current` | the code as it is now, named with the tag's version | a fix to `smith.toml`, the build command or the workflow reaches the files, but they carry newer code under an old number |
| `tag` | the tag's own commit, in a temporary `git worktree add --detach` that smith removes afterwards | the exact code that was released, with the `smith.toml` of that commit; the setup phase installs its dependencies first |

Without `--source`, a terminal is asked, with these consequences. Any other run (a script, CI) is
refused, and the message names both flags. If the tag's commit has no assets config for the
project (the config did not exist yet), smith says so with the commit and tells you to use
`--source current`.

In the generated GitHub workflow, run it from the Actions tab with the `assets_tag` input
(`mvd@0.11.1`) and the `assets_source` input (`choose`, `current` or `tag`; `choose` stops the run
and names the flags): only the assets job runs. See
[ADR 0075](../architecture/decisions/0075-an-assets-rerun-for-a-tag-asks-which-code-to-build.md).

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
