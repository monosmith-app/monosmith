# ADR 0069: The assets phase can run again for a tag, and a glob can be optional

- Status: accepted (the source of the code a rerun builds is changed by [ADR 0075](0075-an-assets-rerun-for-a-tag-asks-which-code-to-build.md))
- Date: 2026-11-01

## Context

A real user (a Tauri app and a Go terminal app, smith 2.0.0-beta.13) ran the assets phase for the
first time. Four faults showed.

1. The `assets` list of a desktop app has a glob per installer kind (`mvd_*.msi`, `mvd_*.deb`,
   `mvd_*.dmg`). A glob that matches nothing was an error, so Windows and Linux failed on the
   `.dmg` glob although they had built their files.
2. The terminal app is native (cgo). It is built by the native job, one archive per runner:
   linux/amd64, windows/amd64 and macos/arm64. `platforms` does not apply to it. The docs named
   the default of six platforms without this exception.
3. A later push changed only `smith.toml` and the workflow. It released nothing, but the assets
   jobs ran and uploaded nothing to the existing Releases, and no command could ask for them again.
4. The assets jobs ran after a push that released nothing. The ci job's `released` output was
   `steps.release.outputs.released || steps.released.outputs.done`, and the step "Mark the release
   done" set `done=true` on every push to the release branch, released or not.

## Decision

- A glob of `assets` may match nothing when at least one other glob of the list matches a file on
  that platform. The runner prints one `no file matches <glob> on <platform>: skipped` line per such
  glob. When no glob matches, the build fails with the old message. smith has no per-platform
  asset tables: the list stays one list, and a platform that builds nothing at all still fails.
- `smith ci assets --tag <project@version>` builds that project's release files and uploads them to
  the GitHub Release of the tag, with `gh release upload --clobber`, so files of the same name are
  replaced. It builds the **current checkout**, not the commit of the tag, and stamps the version of
  the tag. Reason: the usual cause of a rerun is a fix in `smith.toml`, the build command or the
  workflow, and the commit of the tag has the fault. smith prints a notice when HEAD is not the
  commit of the tag. The tag must exist in the checkout (the workflow fetches the whole history).
- The generated GitHub workflow has the `workflow_dispatch` input `assets_tag`. When it is set, the
  assets job runs on that dispatch for that tag and does not need the ci job's `released` output;
  the ci job's release step does not run on a dispatch.
- Without `--tag`, `smith ci assets` lists the release tags that point at HEAD. When it finds none
  it prints one line, `nothing released, skipping: no project was released by this commit`, and
  exits 0.
- The ci job's `released` output is the output of the release step only. smith writes
  `released=true` itself (`markReleased`) when a project reached the status `released`. The step
  "Mark the release done" is removed.
- Docs state that a native app gets one archive per native runner (`[ci] native_runners`), and
  that `platforms` is for non-native apps.

## Consequences

- A typo in one glob is no longer an error when another glob matches. The skipped line in the log
  is the only sign. We accept this because the only other way is a table per platform, which is
  more config for the same result.
- A rerun for a tag can ship code newer than the tag. The release notes of that Release will not
  match the files. The user chooses this by running the command.
- No release is created by the rerun. The Release of the tag must exist.
