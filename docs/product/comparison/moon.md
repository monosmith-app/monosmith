# moon compared with MonoSmith

moon is a repository management, organisation, orchestration and notification
tool, written in Rust. Of the tools compared here it is the closest in shape to
smith: a project graph, a hash-based cache, code generation and several
languages.

## What moon is best at

- **Languages.** Bun, Deno, Go, Node.js (npm, pnpm, yarn) and Rust are fully
  supported; Python and Ruby are partial (experimental); PHP, Bash and others are
  limited, and anything else can run through a "system" toolchain.
- **Integrated toolchain.** moon downloads and installs explicit versions of Node.js
  and other tools, workspace-wide or per project. A separate product, proto, is a
  multi-language version manager.
- **Smart hashing and remote caching** between teammates and CI.
- **Affected runs.** `moon run --affected` runs a task only when changed files
  intersect its inputs; `--affected remote` compares HEAD with the default branch.
  `moon ci` runs the tasks affected by changed files that have `runInCI` enabled.
- **Code generation** to scaffold applications, libraries and tooling.
- **Other features** the README lists (and notes are not all supported yet): code
  ownership with CODEOWNERS generation, flakiness detection with retries, webhook
  events and Git hook management.
- **Open source.** MIT.

## Where MonoSmith differs

| | moon | MonoSmith |
|---|---|---|
| Languages fully supported | Bun, Deno, Go, Node, Rust (others partial) | eleven languages built in |
| Toolchain | moon can install tool versions | smith does not install compilers; `smith doctor` reports what is missing |
| Release | not covered by the pages read | conventional commits to eight registries |
| CI | `moon ci` runs affected tasks | generates the pipeline files for three providers, and the steps call `smith ci` |
| Health check | not covered by the pages read | `smith doctor` with fixes |

## When to choose moon instead

- You want the tool to install and pin the Node, Bun or Deno version for everyone.
- You use Bun or Deno, which smith has no plugin for.
- You want its code ownership and Git hook features.

## When MonoSmith is the better fit

- You want release and generated CI in the same tool, or languages such as .NET, Ruby,
  PHP, Dart, the JVM or Bicep.

## Moving between moon and MonoSmith

No automated path in either direction. `smith adopt --dry-run` is the way to see
what adopting a repository would involve. This was not tried on a moon workspace.

## Sources

Retrieved 2026-10-25.

- Documentation home: <https://moonrepo.dev/docs>
- `moon run` (`--affected`): <https://moonrepo.dev/docs/run-task>
- `moon ci`: <https://moonrepo.dev/docs/commands/ci>
- Repository (licence, features): <https://github.com/moonrepo/moon>
