# Please compared with MonoSmith

Please is a cross-language build system, in its own words "with an emphasis on high
performance, extensibility and reproducibility". It is written in Go and supports
Linux, macOS and FreeBSD.

## What Please is best at

- **Hash-based caching.** It caches artifacts by hashes of each rule's inputs, not
  file timestamps. A directory cache is on by default (it makes switching between
  branches fast), and an HTTP cache can run on a central machine to share artifacts.
- **Selecting what changed.** `plz query changes` "queries changed targets versus a
  revision or from a set of files" (for example `--since master`), and its output can
  be piped to `plz test`.
- **Remote execution (experimental).** From v15 Please can submit build actions to
  remote executors through the Remote Execution API; the page opens with "This
  feature is still experimental".
- **Cross-language builds**, with "a number of popular languages" (Go, Python and
  Java are mentioned in passing) and plugins that add rules for further languages.
- **Editors.** Extensions for Visual Studio Code, IntelliJ and Neovim, and a language
  server for BUILD files.
- **Open source.** Apache-2.0.

No release or versioning command, dependency-alignment check or health check was
found in the documentation source (searched), and CI is covered by a codelab for
GitHub Actions rather than a generator.

## Where MonoSmith differs

| | Please | MonoSmith |
|---|---|---|
| Platforms | Linux, macOS, FreeBSD; the FAQ says "Windows is unfortunately not supported natively" (WSL is described) | Linux, macOS, Windows (x64 and arm64) |
| What you write | BUILD files; `plz init` writes `.plzconfig` and a `pleasew` wrapper | `smith.toml`; the rest inferred |
| Remote execution | experimental, from v15 | none |
| Release, CI generation, health check | none found (searched); a GitHub Actions codelab | all three |
| Repository status | Apache-2.0, not archived, last push 2026-10-24 (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

## When to choose Please instead

- You want a BUILD-file based, reproducibility-focused build system written in Go, and
  you run on Linux, macOS or FreeBSD.

## When MonoSmith is the better fit

- You work on Windows, or want a tool that infers projects from manifests and also
  generates CI and releases.

## Moving between Please and MonoSmith

No automated path in either direction, and neither was tried.

## Sources

Retrieved 2026-10-25.

The Please website is built from the `docs` folder of the repository, whose HTML
and Markdown files were read here as source, not through a summarising tool. Paths
are relative to `docs`.

- Repository (description, platforms, licence): <https://github.com/thought-machine/please>. README: "Currently Linux (tested on Ubuntu), macOS and FreeBSD are actively supported." GitHub API on 2026-10-25: licence Apache-2.0, archived false, last push 2026-10-24.
- Caches: `cache.html` ("The directory cache"; "The HTTP cache").
- Remote execution: `remote_builds.html` ("From v15 onwards, Please has the ability to submit build actions to remote executors").
- Commands: `commands.html` (`plz query changes`; `plz init`).
- Editors: `tooling.html` (extensions for VS Code, IntelliJ and Neovim; "Please ships with a language server for BUILD files").
- Plugins: `plugins.html` ("Plugins are a way to extend Please with build rules for additional languages").
- Windows: `faq.html` ("Windows is unfortunately not supported natively").
- CI: `codelabs/github_actions.md` ("Running Please on GitHub Actions").

### Added for the master table

Retrieved 2026-10-25, from the same HTML and Markdown source files as above, for the cells of [Everything compared](everything-compared.md) that the other tables do not cover. Paths are relative to the documentation folder named above.

- `commands.html` (`plz watch`, `plz clean`); `codelabs/genrule.md` ("builds are hermetic"); `tooling.html` (a language server for BUILD files); `faq.html` ("Windows is unfortunately not supported natively"); `codelabs/puku.md` (diagnosing missing imports).
