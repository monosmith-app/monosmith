# ADR 0083: The dist-tags move in a last phase, after every project has published

- Status: accepted
- Date: 2026-11-02
- Amends: [0082](0082-a-dist-tag-moves-only-after-the-platform-packages-are-fetchable.md)

## Context

ADR 0082 made the publish of the `smith` project wait, inline, until npm served its platform packages.
Release 2.0.0-beta.25 was the first real use, and it showed two faults.

1. One package, `@monosmith/cli-linux-x64@2.0.0-beta.25`, never became fetchable in 15 minutes. `npm
   publish` had returned success for it, and smith had not checked that the registry kept it. The
   release waited the full time for a package that was not there.
2. The wait blocked the later projects. The next project, the VS Code extension, publishes with an
   Azure federated sign-in made at the start of the job. After 15 minutes the token had expired
   ("Can not acquire a Microsoft Entra ID access token") and the release failed. PyPI and NuGet
   trusted-publishing tokens, minted earlier, have the same short life.

## Decision

- A publish never waits. A plugin that publishes a package with same-release platform packages (and
  the host, for the npm packages of an app) does at most one quick probe. When the platform packages
  are not fetchable at that probe, the package is published under the dist-tag `pending` and the
  publish returns the tag to move as a `TagMove` (`pendingMoves`): package, version, tag, and the
  packages to await.
- After every project has published, the host runs a last phase, `FinalizeTags`. For each pending
  move it waits (the algorithm of ADR 0082 with its progress lines and `propagation_wait`) until the
  platform packages are fetchable, then runs `npm dist-tag add` and removes `pending`. A plugin does
  this through `publish` with the new parameter `finalize` (a list of moves): no new method, no new
  capability. The host does it itself for the npm packages of an app.
- The GitHub Releases and the assets come before this phase. They do not need the registry, so they
  never wait.
- When the time runs out, the tags stay under `pending`, the report names the exact commands, a CI
  warning is printed and the release exits with code 2, as in ADR 0082. A failed move exits with 1.
  If a later project fails, the last phase does not run: the tags stay `pending` and the re-run
  moves them.
- A re-run finds the work again. For a version whose tag exists, the host asks the registry for the
  npm packages of the project: it publishes the ones the registry lacks (it rebuilds their files),
  and it hands every wrapper whose dist-tag `pending` equals the version to the last phase. A
  finished release is left alone. The `pending` tag is removed when the real tag moves.
- Each platform package is read back from the registry right after `npm publish` returned 0: three
  quick checks, 2 s apart. A package that is not there yet is not an error. npm needs minutes to
  serve a new package, so the release goes on, the wrappers take `pending` and the last phase waits.
  Only when that wait ends (`propagation_wait`, default 15 min) and a package is still not
  fetchable, smith says so: the note names `pkg@version` ("npm said it published it but the
  registry still does not serve it after 15m"), tells the user to run the release again (it
  republishes what is missing), prints the `npm dist-tag add` command, and the release exits with
  code 2. The first version of this decision failed the release after 25 s (beta.26); this
  amendment replaces that.
- A re-run decides a package is missing with the same probe as the wait: `npm view` finds it and a
  HEAD request of its tarball answers. A package npm never stored fails both and is republished.

## Consequences

- The publish steps of all projects take the time they need and nothing more. The wait happens once,
  at the end, and a sign-in that expires during it costs nothing, because no step comes after it.
- With several pending tags the waits run one after the other, each up to `propagation_wait`.
- A node package (not an app) with same-release `optionalDependencies` is resumed only through the
  command the report prints: the host does not find its pending state on a re-run. This is a known
  gap; the npm packages of an app, the case of the incident, resume.
- A `publish` call with `finalize` set publishes nothing. This reuse of the method keeps the
  protocol small. A plugin that does not support it gets no pending moves, because it never returns
  one.
