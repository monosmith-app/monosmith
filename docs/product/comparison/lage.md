# Lage compared with MonoSmith

Lage is a task runner for JavaScript monorepos from Microsoft. Its repository's
description gives that one line; the README points to a documentation site for the
rest.

## What Lage is best at

According to its documentation site:

- **Caching.** It remembers completed work and skips what is not needed, with a local
  cache or your own cloud storage as a remote cache.
- **Skipping work based on changes**, and reusing work teammates have already done.
- **Parallel execution** instead of sequential builds.
- **Profiling.** A built-in profiler and build graph visualisations show bottlenecks.
- **Quick setup:** the site says about a minute.
- **Open source.** MIT.

Whether Lage orders tasks by package dependencies, how it selects changed packages
and what its remote cache providers are were not verified from the pages read.

## Where MonoSmith differs

| | Lage | MonoSmith |
|---|---|---|
| Scope | JavaScript monorepos | eleven languages |
| Beyond running tasks | not covered by the pages read | scaffolding, CI generation, release, health check |

## When to choose Lage instead

- You want a task runner for a JavaScript monorepo and nothing else.

## When MonoSmith is the better fit

- The repository has more than JavaScript, or you need CI and release handled too.

## Moving between Lage and MonoSmith

No automated path. A JavaScript monorepo using npm workspaces can be read by
`smith adopt --dry-run`. This was not tried.

## Sources

Retrieved 2026-10-25.

- Documentation site: <https://microsoft.github.io/lage/>
- Repository (licence, description): <https://github.com/microsoft/lage>
