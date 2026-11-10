# ADR 0082: A dist-tag moves only after the platform packages are fetchable

- Status: accepted
- Date: 2026-11-02
- Amended by: [0083](0083-the-dist-tags-move-in-a-last-phase-after-every-project-has-published.md): the wait moves to a last phase after every project has published

## Context

On three releases in a row (beta.19, beta.22, beta.24) `npm view monosmith dist-tags.next` pointed to
the new version while one to six of the six platform packages (`@monosmith/cli-win32-x64` and the
others, pinned as `optionalDependencies` of the main package) answered 404 from other machines for
five to ten minutes. This is npm edge propagation. A person who ran `npm install -g monosmith@next`
in that window got a broken install: npm skips a missing optional package without an error, and
`smith` then failed with "is not installed". CI jobs of other repositories failed the same way.

The check added in beta.23 reads each package back from the CI runner with `npm view` for 60 seconds.
The runner asks the same edge that took the publish, so it passes while other edges still say 404.

## Decision

- A release publishes all platform packages first, then waits until each one is fetchable the way a
  person fetches it: `npm view <pkg>@<version>` finds it and a HEAD request of its `dist.tarball`
  URL succeeds. Only then does it publish the wrapper packages under their dist-tag.
- The wait polls with exponential backoff (2 s, doubling, at most 30 s). After every package is
  fetchable, three consecutive probes 30 s apart must all succeed, so that more than one edge
  has answered. One failed probe restarts the count. A progress line names the missing packages
  every 60 s.
- The longest wait is `propagation_wait`: `[release]` for the workspace, `[projects.x.release]` for
  one project, a duration such as `15m`. Default `15m`. `0` does not wait: one probe, and the tag
  moves only if it passes.
- When the wait ends before the packages are fetchable, no dist-tag moves. The wrapper packages are
  published under the dist-tag `pending`, the tag and the release are otherwise complete, the report
  names the exact `npm dist-tag add <name>@<version> <tag>` commands, a CI warning annotation is
  printed, and `smith release` exits with code 2. A failure keeps code 1.
- The rule is generic: the npm packages of an app (`[projects.x.release] npm`) and any node package
  whose `optionalDependencies` pin the version being published use the same wait. A stable release
  without a channel also takes the `pending` tag instead of `latest`.
- The launcher of the npm packages does not only print "is not installed". It asks npm, with a 5 s
  limit, whether the platform package exists at its own version, and says which of three
  situations it is: the package exists (a transient install miss: it offers the exact install
  command and asks "Install it now? [Y/n]" in a terminal), the package does not exist yet (the
  release may still be spreading: it offers to wait and retry every 30 s for up to 10 minutes,
  default no) or npm could not be asked. It never installs without a yes. `SMITH_AUTO_FIX=1`
  accepts the install of the first situation, for CI.

## Consequences

- A release takes at least one minute longer for a package with platform packages (three probes
  30 s apart). The user can set `propagation_wait = "0"` to skip it.
- A slow edge no longer produces a broken `next`. The cost is that `next` can stay on the previous
  version until a person runs the printed command. Nothing is changed without that person's say.
- The probes use `--prefer-online` so the npm cache of the runner cannot answer for the registry.
  Rationale for three probes and 30 s: not recorded beyond the three incidents above.
