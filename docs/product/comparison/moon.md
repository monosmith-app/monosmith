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
| Release | the FAQ answers "Can moon version/publish packages?" with "At this time, no", and suggests Yarn releases, Changesets or Lerna | conventional commits to npm, PyPI, crates.io, NuGet, RubyGems, pub.dev and Maven repositories, Go modules and Packagist by tag |
| CI | `moon ci` runs affected tasks; the CI guide shows examples; no pipeline files are generated | generates the pipeline files for three providers, and the steps call `smith ci` |
| Health check | `moon sync projects` repairs configuration "to help achieve a healthy repository state"; no read-only check command found | `smith doctor` with fixes |
| Dependency alignment | none found | `smith sync`; `smith doctor` checks one version per production dependency |
| Plugins | WASM plugins and extensions, in any language that compiles to WASM; the plugin page says the implementations "are still experimental" | out-of-process JSON-RPC plugins, in any language |
| Editors | a VS Code extension (documented as "in its early stages") and an MCP server | a VS Code extension, a terminal interface and an MCP server |
| Moving in | `moon ext migrate-nx` and `moon ext migrate-turborepo` | `smith convert` from Nx, Turborepo or mnci |
| Repository status | MIT, not archived, last push 2026-10-24 (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

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

The documentation was read as MDX source files from the repository's `website/docs`
folder, not through a summarising tool.

- Documentation home: <https://moonrepo.dev/docs>; supported languages (tiers): <https://moonrepo.dev/docs> (Introduction)
- `moon run` (`--affected`): <https://moonrepo.dev/docs/run-task>
- `moon ci`: <https://moonrepo.dev/docs/commands/ci> ("run all tasks that are affected by changed files")
- Remote cache: <https://moonrepo.dev/docs/guides/remote-cache> ("compatible with the Bazel Remote Execution v2 API", self-hosted; a cloud-hosted option is on the same page)
- Release: <https://moonrepo.dev/docs/faq> ("Can moon version/publish packages?")
- Sync: <https://moonrepo.dev/docs/commands/sync/projects>
- Editor: <https://moonrepo.dev/docs/editors/vscode> ("This extension is in its early stages"); MCP: <https://moonrepo.dev/docs/guides/mcp>
- Plugins: <https://moonrepo.dev/docs/guides/wasm-plugins> ("any language that compiles to WASM")
- Migration extensions: <https://moonrepo.dev/docs/guides/extensions> (`migrate-nx`, `migrate-turborepo`)
- `moon init`: <https://moonrepo.dev/docs/setup-workspace> ("Creates a `.moon` folder")
- Repository (licence, features): <https://github.com/moonrepo/moon>. The README lists "Action distribution" and adds "Not all features are currently supported". GitHub API on 2026-10-25: licence MIT, archived false, last push 2026-10-24.
