# Lerna compared with MonoSmith

Lerna is a tool for managing and publishing several JavaScript and TypeScript
packages from one repository. Its documentation calls it the original monorepo
tool for JavaScript and TypeScript. Stewardship has moved to Nx, and from Lerna
v6 task scheduling is delegated to the Nx task runner.

## What Lerna is best at

- **Versioning and publishing to npm**, with independent or shared versions
  across packages.
- **Running scripts in dependency order**: `lerna run build`, several targets at
  once, `--scope` for one project.
- **Affected runs:** `lerna run test --since=origin/main`.
- **Caching:** it detects when a task has already run, restores files and replays
  the output; the documentation says the cache can be shared with teammates and CI.
- **Open source.** MIT.

## Where MonoSmith differs

| | Lerna | MonoSmith |
|---|---|---|
| Languages | JavaScript and TypeScript | eleven languages built in |
| Publishing | npm | npm, PyPI, crates.io, NuGet, RubyGems, pub.dev, Maven, Go module tags |
| Version source | options for different workflows | conventional commits |
| Task engine | the Nx task runner | smith's own, in Go |
| CI generation, health check | not covered by the pages read | generated pipelines, `smith doctor` |

## When to choose Lerna instead

- You publish a set of npm packages, the repository is JavaScript or TypeScript only,
  and you want the most direct tool for versioning and publishing them.
- You already use Nx and want Lerna's publishing commands on top.

## When MonoSmith is the better fit

- Other languages are in the repository, or you publish to registries other than npm.

## Moving between Lerna and MonoSmith

smith has no Lerna importer. A Lerna repository that uses npm workspaces can be read by
`smith adopt --dry-run`; the report names any blocker. If the repository
uses Nx beneath Lerna, `smith convert` applies. Release tags would need checking:
`smith adopt --tags` creates, locally, the tags smith needs where existing tags are
named differently. This was not tried on a Lerna repository.

## Sources

Retrieved 2026-10-25.

- Introduction: <https://lerna.js.org/docs/introduction>
- Running tasks: <https://lerna.js.org/docs/features/run-tasks>
- Repository (licence): <https://github.com/lerna/lerna>
