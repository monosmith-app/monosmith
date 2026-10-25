# Gradle compared with MonoSmith

Gradle is an open-source build automation tool whose README lists Java, Kotlin,
Scala, Android, Groovy, C++ and Swift. A Gradle multi-project build organises
related modules under one build so they can be built, tested and released together.
Its documentation is about building those languages; it does not describe
running Node, Go and Python projects beside them.

## What Gradle (multi-project) is best at

- **JVM and Android builds.** Subprojects are declared in `settings.gradle(.kts)`
  with `include(...)`, depend on each other with `project(":shared")`, and Gradle
  builds the required project first.
- **A build cache.** Task outputs are reused when inputs are unchanged. The local
  cache is on by default under Gradle's user home; an HTTP remote cache shares
  outputs across machines, and a common setup lets CI push while developers only
  pull. Only cacheable tasks take part, and custom task types are not cacheable by
  default.
- **Version catalogs** that centralise dependency versions across subprojects.
- **Open source.** Apache-2.0.

## Where MonoSmith differs

| | Gradle | MonoSmith |
|---|---|---|
| Scope | JVM, Android and some native languages | eleven languages; Gradle and Maven are run by the `jvm` plugin |
| Model | a programmable build per project | tasks inferred from manifests, run by the language's own tool |
| Cache | Gradle's task outputs | content-addressed cache across all languages, including Gradle tasks it runs |
| Affected | up-to-date checks skip tasks whose inputs and outputs are unchanged; no selection by git range found | `smith affected` selects projects from a base branch; the cache skips unchanged tasks |
| Release | publishing plugins such as `maven-publish` ("lets you publish artifacts ... to a Maven repository"); choosing and writing versions is left to the build | versions and tags from commits, then publish |
| CI generation | a section with recipes for CircleCI, GitHub Actions, GitLab CI, Jenkins, TeamCity and Travis CI; no generator found | generated pipelines for three providers |
| Scaffolding | `gradle init` (Build Init plugin) creates new builds | `smith add`, recipes |
| Dependency alignment | a how-to for aligning versions of modules in one platform; version catalogs; no repository-wide lint command found | `smith sync`; `smith doctor` checks one version per production dependency |
| Editors | the user guide lists integration with Android Studio, IntelliJ IDEA, Eclipse (Buildship), NetBeans, Visual Studio, Xcode and CLion | a VS Code extension, a terminal interface and an MCP server |
| Remote execution | no mention in the user guide (searched "remote execution" and "distributed build") | none |
| Repository status | Apache-2.0, not archived, last push 2026-10-25 (GitHub API, 2026-10-25) | MIT, not archived, first commit 2026-10-15; release 1.62.0 |

smith does not replace Gradle inside a JVM project. Its `jvm` plugin runs Gradle
modules (`gradle assemble`, `test`, ktlint) and Maven modules (`mvn install`)
as tasks, and puts their output under `dist/gradle` or `dist/maven`.

## When to choose Gradle instead

- The repository is a JVM or Android codebase and you want Gradle's build
  programmability and its own caching.
- You use Gradle features no other tool reproduces; keep Gradle, and consider smith
  only on top, to run it with the repository's other languages.

## When MonoSmith is the better fit

- Gradle modules sit beside Node, Go or Python projects and you want one command,
  one pipeline and one release process.

## Moving between Gradle and MonoSmith

You do not move away from Gradle: a Gradle module stays as it is, and smith discovers
it through the root settings file. `smith adopt --dry-run` reports what adopting
the repository would involve. Tried on 2026-10-25 with smith 1.62.0 on a scratch
two-module Kotlin Gradle build: `smith adopt --dry-run` listed both modules as
`jvm` projects and reported that Gradle was not on `PATH`. Nothing else was tried.

## Sources

Retrieved 2026-10-25.

The user guide was read as AsciiDoc source files from the repository, under
`platforms/documentation/docs/src/docs/userguide` (the paths below are relative to
it), not through a summarising tool.

- Multi-project builds: <https://docs.gradle.org/current/userguide/multi_project_builds.html>
- Build cache: <https://docs.gradle.org/current/userguide/build_cache.html>; source `optimizing-builds/build-cache/build_cache.adoc` has a "Remote HTTP build cache" section and says "By default, the local build cache has push enabled".
- Version catalogs (platforms page): <https://docs.gradle.org/current/userguide/platforms.html>
- Aligning versions: `how-to/dependency-management/how_to_align_dependency_versions.adoc` ("Dependency version alignment ensures that different modules belonging to the same logical group").
- CI recipes: `on-ci/gradle_on_ci.adoc` ("recipes and best practices for configuring popular CI platforms"), with one page each for CircleCI, GitHub Actions, GitLab CI, Jenkins, TeamCity and Travis CI.
- IDEs: `integration/third_party_integration.adoc` (Android Studio, Eclipse Buildship, IntelliJ IDEA, NetBeans, Visual Studio, Xcode, CLion).
- `gradle init`: `best-practices/best_practices_general.adoc` ("Kotlin DSL is the default for new builds created with `gradle init`").
- Publishing: `fundamentals/developing-plugins/publishing_binary_plugin_advanced.adoc` (the `maven-publish` plugin).
- Repository (licence, languages): <https://github.com/gradle/gradle>. GitHub API on 2026-10-25: licence Apache-2.0, archived false, last push 2026-10-25.

### Added for the master table

Retrieved 2026-10-25, from the same AsciiDoc source files as above, for the cells of [Everything compared](everything-compared.md) that the other tables do not cover. Paths are relative to the documentation folder named above.

- `optimizing-builds/build-cache/build_cache_debugging.adoc` ("finding out why a cache miss happened"); `reference/runtime-configuration/command_line_interface.adoc` (continuous build re-executes tasks when inputs change); `integration/third_party_integration.adoc` (the Tooling API); `fundamentals/authoring-builds/gradle_directories_intermediate.adoc` ("JDKs downloaded by the toolchain support") and `reference/core-plugins/checkstyle_plugin.adoc` ("Gradle's toolchain support automatically provisions the specified JDK").
