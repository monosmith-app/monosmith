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
  the output; the documentation says the cache can be shared with teammates and CI,
  through your own implementation or Nx Cloud.
- **Distributed task execution** through Nx Cloud: the documentation describes
  connecting to Nx Cloud and enabling it in the CI workflow.
- **Scaffolding and editors.** `lerna init` and `lerna create`; Nx Console for
  VS Code and JetBrains, "built and maintained by the Nx team".
- **Open source.** MIT.

## Where MonoSmith differs

| | Lerna | MonoSmith |
|---|---|---|
| Languages | JavaScript and TypeScript | eleven languages built in |
| Publishing | npm | npm, PyPI, crates.io, NuGet, RubyGems, pub.dev, Maven, Go module tags |
| Version source | `fixed` or `independent` versions; `lerna version --conventional-commits` derives them from commits | conventional commits |
| Task engine | the Nx task runner | smith's own, in Go |
| CI generation | CI examples for distributed execution; no generator found | generated pipelines |
| Health check | `lerna repair` updates configuration files after an upgrade; no repository check found | `smith doctor` |
| Dependency alignment | none found (the legacy `lerna bootstrap`, `add` and `link` commands are documented as moved to a separate package) | `smith sync` |
| Repository status | MIT, not archived, last push 2026-10-25 (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

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
named differently. Tried on 2026-10-25 with smith 1.62.0 on a scratch repository with npm workspaces, a `lerna.json` and two packages (one depending on the other): `smith adopt --dry-run` listed both as `node` packages and warned that `node_modules` was missing. Release tags were not tried.

## Sources

Retrieved 2026-10-25.

The documentation was read as Markdown source files from the repository's
`website/docs` folder, and the command READMEs under `libs/commands`, not through a
summarising tool.

- Introduction: <https://lerna.js.org/docs/introduction> ("original monorepo tool"; stewardship moved to Nx)
- Running tasks: <https://lerna.js.org/docs/features/run-tasks>
- Sharing the cache: <https://lerna.js.org/docs/features/share-your-cache> ("You can either build an implementation of the cache or use Nx Cloud")
- Distributed execution: <https://lerna.js.org/docs/features/distribute-tasks> ("you need to (1) connect to Nx Cloud")
- Commands: <https://lerna.js.org/docs/api-reference/commands>; `lerna version --conventional-commits` in the `lerna version` README (<https://github.com/lerna/lerna/tree/main/libs/commands/version>); `lerna repair`: "Update configuration files to match the currently installed version of lerna".
- Scaffolding: <https://lerna.js.org/docs/faq> ("use the `lerna create` command")
- Editors: <https://lerna.js.org/docs/features/editor-integrations>
- Repository: <https://github.com/lerna/lerna>. GitHub API on 2026-10-25: licence MIT, archived false, last push 2026-10-25.
