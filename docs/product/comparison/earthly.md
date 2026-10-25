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
| Cache, affected, release | not covered by the pages read | cache and affected runs, releases from commits |

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

- Repository (description, licence, status notice, fork link): <https://github.com/earthly/earthly>
