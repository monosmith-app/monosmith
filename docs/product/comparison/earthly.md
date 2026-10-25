# Earthly compared with MonoSmith

Earthly is a build and CI/CD framework that runs builds in containers, with a
syntax that combines ideas from Dockerfiles and Makefiles. Its README says it is
"like Dockerfile and Makefile had a baby". Read the status notice before anything
else.

## Status

The Earthly repository's README says, as retrieved on 2026-10-25, that Earthly
Technologies has ended active maintenance of the open-source project, that Earthly
Cloud (including Satellites) was shut down on 16 July 2025, and that the source stays
available under its licence. It points to a community-maintained fork, EarthBuild,
and says the company is now building another product. If you are evaluating Earthly
for a new monorepo, read the current state of the repository and of the fork
yourself first.

What the GitHub API shows on 2026-10-25: the repository is not archived; the latest
release is v0.8.16 (2025-07-16); the commit before the newest is also from
2025-07-16, and the newest (2026-10-23, the repository's "last push") is the one
that made the notice prominent in the README. The fork, EarthBuild, is not archived,
published v0.8.19 on 2026-09-17 and a release candidate (v0.8.20-rc-0) on 2026-10-22,
and says it is "endorsed and supported by" Earthly. This page compares Earthly as
documented; it does not assess the fork.

## What Earthly was built for

- Reproducible builds that run in containers, so the same Earthfile runs the same way
  on a laptop and in CI.
- Wrapping existing build tools instead of replacing them.
- Licence: MPL-2.0.

## Where MonoSmith differs

| | Earthly | MonoSmith |
|---|---|---|
| How tasks run | in containers | as child processes of the language's own tools on the host (a Docker image is built for apps with a Dockerfile) |
| Language support | any, through containers | eleven languages through plugins |
| Task graph | targets with declared dependencies run in parallel, per the README ("complex build graphs") | projects and tasks in dependency order |
| Cache | layer caching and cache mounts locally; a registry-based remote cache (`--remote-cache`); "auto-skip" for monorepos needs an Earthly Cloud account, and Earthly Cloud was shut down | content-addressed local cache; HTTP, S3 or Azure Blob remote cache, signed |
| Remote work | remote runners ("Earthly is able to use remote runners for performing builds on remote machines"); the hosted Satellites were shut down | none: tasks run on one machine |
| Release, scaffolding, dependency alignment, health check | none found; these are not areas the documentation covers (searched the documentation source) | releases from commits, `smith add`, `smith sync`, `smith doctor` |
| CI | guides for integrating Earthly with CI systems; no generator found | generated pipelines for three providers |
| Editors | a VS Code extension for Earthfiles exists (its build is listed among the examples) | a VS Code extension, a terminal interface and an MCP server |
| Repository status | MPL-2.0, not archived, last push 2026-10-23 (README notice) (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

## When to choose Earthly (or its fork) instead

- Your priority is that builds run identically in containers everywhere, and you accept
  the maintenance position above.

## When MonoSmith is the better fit

- You want a monorepo tool with a project graph, release and generated CI, and you
  are content to run tools on the host.

## Moving between Earthly and MonoSmith

No automated path. Earthfile targets would become smith tasks by hand
(`[projects.<name>.tasks.<task>]` in `smith.toml`). This was not tried.

## Sources

Retrieved 2026-10-25.

The documentation was read as Markdown source files from the repository's `docs`
folder, not through a summarising tool. Paths are relative to `docs`.

- Repository (description, licence, status notice, fork link): <https://github.com/earthly/earthly>. README: "Earthly Technologies has ended active maintenance of the Earthly open-source project". GitHub API on 2026-10-25: licence MPL-2.0, archived false, last push 2026-10-23; releases and commits as described above.
- Fork: <https://github.com/earthbuild/earthbuild>. GitHub API on 2026-10-25: licence MPL-2.0, archived false, last push 2026-10-25.
- Caching: `caching/caching-in-earthfiles.md` ("Most commands in an Earthfile create a cache layer"; auto-skip "stored in a cloud database" needing "an Earthly Cloud account").
- Remote cache: `guides/best-practices.md` (a `--remote-cache=... --max-remote-cache` example).
- Remote runners: `basics/part-7-using-remote-runners.md`.
- CI: `ci-integration/overview.md` and its guides.
- Examples: `examples/examples.md` (the VS Code extension build).
