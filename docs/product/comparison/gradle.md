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
| Release, CI generation | not covered by the pages read | both |

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
the repository would involve. This was not tried on a Gradle repository.

## Sources

Retrieved 2026-10-25.

- Multi-project builds: <https://docs.gradle.org/current/userguide/multi_project_builds.html>
- Build cache: <https://docs.gradle.org/current/userguide/build_cache.html>
- Version catalogs (platforms page): <https://docs.gradle.org/current/userguide/platforms.html>
- Repository (licence, languages): <https://github.com/gradle/gradle>
