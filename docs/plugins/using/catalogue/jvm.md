# The jvm plugin

The jvm plugin makes JVM modules smith projects, built by one of two tools. A
Maven module (a `pom.xml`) is Java: the root `pom.xml` is the aggregator and
the parent of every module. A Gradle module (a `build.gradle.kts` or
`build.gradle`) is Kotlin, through the `kotlin` stack, or an Android app,
through the `android` stack. The plugin finds the modules, gives each one build,
test, lint and (for apps) start tasks run from the workspace root and aimed at
that one module, creates new modules, writes release versions, deploys packages
to a Maven repository, and tells CI which JDK, Maven, Gradle and Android SDK to
install.

## At a glance

| | |
|---|---|
| Plugin name | `jvm` (built in; runs as a separate process, `smith plugin serve jvm`) |
| Kind | language |
| Roles | `language`, `toolchain`, `scaffold` (the defaults of a language plugin; it declares none of its own) |
| `--lang` values | `jvm`, `java` |
| Manifests that make a directory a candidate project | `pom.xml`, `build.gradle.kts`, `build.gradle` |
| Requires, conflicts, provides (task slots) | none |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `remove`, `ci`, `doctor`, `published`, `declarations`, `latest`, `changeKind`, `components`, `channels` |
| Not advertised | `link`: `smith install -w <project> <library>` says "the jvm plugin does not link an internal library yet: add the dependency in `pom.xml` or `build.gradle` by hand". `declared`: a declared project is refused (see Declared projects). `dependencies`: `smith install` and `smith uninstall` say the plugin cannot change a module's dependencies (Maven and Gradle have no built-in command to add one; the build tool's plugins would have to be installed first) |
| Stacks it publishes | `kotlin` ("a Kotlin module built with Gradle, linted with ktlint"), `android` ("an Android app in Kotlin, built with Gradle (AGP), linted with ktlint and Android Lint") |
| Project kinds `smith add` can create | `app`, `lib`, `package` (`smith kinds` lists `jvm-app`, `jvm-lib`, `jvm-package`); the `android` stack makes apps only |
| Tasks | `build`, `test`, `lint`, `start` |

Roles and capabilities are what `smith plugin list --json` prints; languages,
manifests and stacks come from the plugin's `Describe` method
(`libs/plugin-jvm/plugin/jvm_plugin_handler.go`).

How the stacks map: `smith add ... --lang java` (or `jvm`) with no stack writes
a **Maven** module in Java. `--stack kotlin` writes a **Gradle** module in
Kotlin. `--stack android` writes a **Gradle** Android app. There is no Gradle
Java module and no Maven Kotlin module in the scaffold.

## What it detects

smith hands the plugin every `pom.xml`, `build.gradle.kts` and `build.gradle`
in the files git lists. The root manifest (directory `.`) and anything under a
`buildSrc` directory are skipped.

**Maven (`pom.xml`)**

- **Is it a project.** Any pom whose packaging is not `pom`. The root pom and
  other aggregators are not projects.
- **Project name.** The directory name (`libs/core` is `core`).
- **Exports.** `groupId:artifactId`, where a missing `groupId` or `version`
  comes from `<parent>`.
- **Imports (graph edges).** Every `<dependency>` as `groupId:artifactId`
  (`${project.groupId}` is expanded). Third-party dependencies match no project
  and make no edge.
- **Kind.** The `<smith.kind>` property (`app`, `lib`, `package`); else the top
  directory; else `lib`. **Stack:** `<smith.stack>`.
- **Version and description.** `<version>` and `<description>`.

**Gradle (`build.gradle.kts` or `build.gradle`)**

- **Project name.** The directory name.
- **Exports.** The Gradle path (`:apps:tool` for `apps/tool`); plus
  `group:artifactId` when the script has both `group = "..."` and
  `artifactId = "..."` (a package has them).
- **Imports.** Each `project(":libs:core")` (also `project(path = ":x")`).
- **Kind and stack.** The `smith.kind` and `smith.stack` properties in the
  module's `gradle.properties`. They are read nowhere else.
- **Version.** `version = "x"` in the script; failing that `version` in
  `gradle.properties`.
- **Lint tasks.** `ktlintCheck` if the script mentions
  `org.jlleitschuh.gradle.ktlint`; otherwise `checkstyleMain` and
  `checkstyleTest` if it mentions `checkstyle`; an Android module (the script
  mentions `com.android.application` or `com.android.library`) adds Android's
  `lint`.
- **Build directory.** If the script sets a path starting with `dist/gradle/`,
  `build` declares it as its output.

**Toolchain.** The first version in `java -version`, reported as `java 21.0.12.1`
(it prints to stderr; smith reads both). Maven's and Gradle's versions are not
probed.

A real run in the Maven workspace, after the `sdk` pom depended on `core`:

```text
$ smith graph
core  lib · jvm · libs/core
sdk  package · jvm · packages/sdk
  └─ core
tool  app · jvm · apps/tool
```

## Tasks

### Maven modules

Run from the workspace root (`cwd` is `.`), each aimed at one module with
`-pl groupId:artifactId` (a path selector is not used: a Windows 8.3 short temp
path defeats it). `-B -q` (batch, quiet) are on, so a successful Maven run
prints almost nothing.

| Task | Command | Inputs | Outputs | dependsOn | Cached | Appears |
|---|---|---|---|---|---|---|
| `build` | `mvn -B -q -pl <g:a> install -DskipTests` | `src/**`, `pom.xml`, `{workspaceRoot}/pom.xml`, `{workspaceRoot}/.mvn/**` | `../../dist/maven/<artifactId>` | `^build` (the modules it imports) | no | every module |
| `test` | `mvn -B -q -pl <g:a> test` | same | none | `build` | yes | every module |
| `lint` | `mvn -B -q -pl <g:a> checkstyle:check` | same | none | `build` | yes | every module |
| `start` | `mvn -B -q -pl <g:a> exec:java` | every file of the project | none | `build` | no | kind `app` |

`build` is `install`: it puts the module's jar and flattened pom into your local
Maven repository (`~/.m2`), so the modules that use it resolve it there. That is
a side effect outside the workspace. Checkstyle uses Google's rules
(`google_checks.xml`) from the root pom, with `violationSeverity` `warning`, so
a warning fails `lint`.

### Gradle modules

Run from the workspace root, aimed at the module by its project path (for
example `:apps:tool`), with `--console=plain -q`. `gradle` must be on `PATH`.

| Task | Command | Outputs | dependsOn | Cached | Appears |
|---|---|---|---|---|---|
| `build` | `gradle --console=plain -q :path:assemble` | `../../dist/gradle/<artifactId>` when the script sets it | `^build` | no | every module |
| `test` | `gradle ... :path:test` | none | `build` | yes | every module |
| `lint` | `gradle ... :path:ktlintCheck` (or `checkstyleMain checkstyleTest`; Android adds `:path:lint`) | none | `build`; for an Android module `test` | yes | when a lint plugin is detected |
| `start` | `gradle ... :path:run` | none | `build` | no | kind `app`, not Android |

Inputs: `src/**`, `build.gradle.kts`, `build.gradle`, `gradle.properties`, and
the workspace's `settings.gradle(.kts)`, root `build.gradle(.kts)`,
`gradle.properties` and `gradle/**`. An Android app has no `start`: it runs on a
device, not from Gradle. The Android module runs `build`, `test`, `lint` one
after another because Android's tasks compile the same intermediates and two
Gradle builds of one module at once overwrite each other.

With the `vertical-slices` convention each module also has `slice-check` and
`lint` depends on it. That task is the convention's.

```text
$ smith explain tool          # Maven app
tool has 5 tasks:
  build        mvn -B -q -pl dev.jvmdemo:jvm-demo-tool install -DskipTests  (after ^build)
  lint         mvn -B -q -pl dev.jvmdemo:jvm-demo-tool checkstyle:check  (cached, after build, slice-check)
  slice-check  smith slices check --lang jvm .  (cached)
  start        mvn -B -q -pl dev.jvmdemo:jvm-demo-tool exec:java  (after build)
  test         mvn -B -q -pl dev.jvmdemo:jvm-demo-tool test  (cached, after build)

$ smith explain droid         # Android app, Gradle
droid has 4 tasks:
  build        gradle --console=plain -q :apps:droid:assemble  (after ^build)
  lint         gradle --console=plain -q :apps:droid:ktlintCheck :apps:droid:lint  (cached, after test, slice-check)
  slice-check  smith slices check --lang jvm .  (cached)
  test         gradle --console=plain -q :apps:droid:test  (cached, after build)
```

Real runs (Maven only; Gradle was not installed, see Verified):

```text
$ smith run build               # three modules
✔ 3/3 tasks succeeded (0 from cache) in 5.883s

$ smith run test lint           # build, slice-check, test, lint for each
✔ 12/12 tasks succeeded (0 from cache) in 11.151s

$ smith run start -p tool
tool:start ▶ mvn -B -q -pl dev.jvmdemo:jvm-demo-tool exec:java
tool:start │ Hello, world!
```

Build output lands in `dist/maven/<artifactId>` (classes, surefire reports,
`<artifactId>-0.0.0.jar`, Checkstyle's `checkstyle-result.xml`, and a
`.flattened-pom.xml`).

## Creating a project

```sh
smith add lib core --lang java           # Maven, Java
smith add app tool --lang jvm
smith add package sdk --lang jvm
smith add lib core --stack kotlin        # Gradle, Kotlin
smith add app droid --stack android      # Gradle, Android
```

Real output for the Maven case:

```text
$ smith add lib core --lang java
Adding lib core (jvm) in libs/core
Wrote 7 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  .mvn/maven.config
  libs/core/pom.xml
  libs/core/src/main/java/dev/jvmdemo/core/greeting/GreetUseCase.java
  libs/core/src/test/java/dev/jvmdemo/core/greeting/GreetUseCaseTest.java
  pom.xml
Depend on it from another module with <dependency><groupId>dev.jvmdemo</groupId><artifactId>jvm-demo-core</artifactId><version>0.0.0</version></dependency>.
Try: smith test -p core

$ smith add app tool --lang jvm
Wrote 5 file(s):
  apps/tool/pom.xml
  apps/tool/src/main/java/dev/jvmdemo/tool/Main.java
  apps/tool/src/main/java/dev/jvmdemo/tool/greeting/GreetUseCase.java
  apps/tool/src/test/java/dev/jvmdemo/tool/greeting/GreetUseCaseTest.java
  pom.xml
```

The Kotlin and Android cases (generation ran; the Gradle build did not):

```text
$ smith add lib core --stack kotlin
Wrote 9 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  .gitignore
  build.gradle.kts
  libs/core/build.gradle.kts
  libs/core/gradle.properties
  libs/core/src/main/kotlin/dev/ktdemo/core/greeting/GreetUseCase.kt
  libs/core/src/test/kotlin/dev/ktdemo/core/greeting/GreetUseCaseTest.kt
  settings.gradle.kts
Depend on it from another module with implementation(project(":libs:core")).

$ smith add app droid --stack android
Wrote 9 file(s):
  .github/workflows/ci.yml
  apps/droid/build.gradle.kts
  apps/droid/gradle.properties
  apps/droid/src/main/AndroidManifest.xml
  apps/droid/src/main/kotlin/dev/ktdemo/droid/MainActivity.kt
  apps/droid/src/main/kotlin/dev/ktdemo/droid/greeting/GreetUseCase.kt
  apps/droid/src/test/kotlin/dev/ktdemo/droid/greeting/GreetUseCaseTest.kt
  build.gradle.kts
  settings.gradle.kts
```

What is written:

- **Maven.** The first module writes the root `pom.xml` (the aggregator and
  parent: `groupId` `dev.<workspace>`, artifact `<workspace>-parent`,
  `maven.compiler.release` 21, JUnit Jupiter 5.11.4 managed, plugin versions
  pinned, the Google Checkstyle configuration, the flatten plugin, and a build
  directory of `dist/maven/${project.artifactId}`) and `.mvn/maven.config`
  (file-lock settings so parallel Maven processes do not corrupt `~/.m2`). Each
  module is added to `<modules>`. The module's pom has the parent, artifact
  `<workspace>-<name>`, version `0.0.0`, `<smith.kind>`, and JUnit; an app also
  sets `exec.skip` to `false` and `exec.mainClass`.
- **The group is a placeholder**, `dev.<workspace>`. Maven Central verifies the
  owner of a group, so replace it before you publish there.
- **Sources.** The greeting slice: `greeting/GreetUseCase.java` (a final class,
  static `greet(String)` returning `Hello, <name>!`) and its JUnit test; an app
  adds `Main.java`, which prints the greeting.
- **Kotlin.** `settings.gradle.kts` (with `google()`, `mavenCentral()` and the
  plugin portal), a root `build.gradle.kts` pinning the Kotlin plugin 2.4.21
  and ktlint 14.2.0, and per module a `build.gradle.kts`, `gradle.properties`
  (`smith.kind`, `smith.stack`), the Kotlin greeting slice and test (JUnit 6.1.3
  BOM). The module's build directory is set to `dist/gradle/<artifactId>`.
  An app applies `application` with `<package>.MainKt`; a package applies
  `maven-publish` and a publication that deploys to the repository named by
  `MAVEN_REPOSITORY_URL`, `MAVEN_USERNAME`, `MAVEN_PASSWORD` when set.
  `.gradle/` and `.kotlin/` are added to `.gitignore`.
- **Android.** An `AndroidManifest.xml`, a `MainActivity` showing the
  greeting, and a script with AGP (version 9.4.1, pinned in the root script),
  `compileSdk` 36, `minSdk` 26, `targetSdk` 36 and ktlint.
- `smith add lib x --stack android` is refused: `smith: the android stack
  writes apps; a library shared with an Android app is a Kotlin module: smith
  add lib x --stack kotlin`.

`smith add` refuses a directory that exists, writes no Gradle wrapper, and runs
neither Maven nor Gradle.

## Settings

The plugin has no `[projects.<p>.settings.jvm]` keys. What can be configured:

| Where | Key | Effect | Read in |
|---|---|---|---|
| `pom.xml` | `<properties><smith.kind>`, `<smith.stack>` | kind, stack | `discovery/pom_mapper.go` |
| `gradle.properties` (module) | `smith.kind`, `smith.stack` | kind, stack (the only place for Gradle) | `discovery/gradle_build_mapper.go` |
| `pom.xml` / `build.gradle(.kts)` | `<version>` / `version = "..."` (or `version=` in `gradle.properties`) | the version a release starts from and rewrites | `releasing/` |
| root `pom.xml` | `<maven.compiler.release>` | the Java release CI installs when no `.java-version` | `pipeline/ci_contribution_use_case.go` |
| `.java-version` (workspace root) | first line; the part before the first `.` is used (`17.0.9` gives `17`) | the Java release CI installs; default `21` | same |
| `smith.toml` | `[toolchains] java = "21"` | doctor pin | host |
| `smith.toml` | `[registries.<name>] auth`, `maven`, `token_env` and `[release] registry` or `[projects.<p>] registry` | where packages deploy and how CI signs in (see Releasing) | `registries/` |
| `smith.toml` | `[projects.<p>] registry = "none"`, `bare_tags`, `tag_prefix` | package published by its tag (JitPack) | host |
| environment | `MAVEN_REPOSITORY_URL`, `MAVEN_USERNAME`, `MAVEN_PASSWORD` | the repository, with no `[registries]` entry | `releasing/release_module_use_case.go` |
| environment | `GITHUB_REPOSITORY`, `GITHUB_ACTOR`, `GITHUB_TOKEN` | GitHub Packages (`maven = "github"`) | `registries/maven_target_policy.go` |
| environment | `ANDROID_HOME`, `ANDROID_SDK_ROOT`; `local.properties` `sdk.dir` | where doctor looks for the Android SDK | `pipeline/android_sdk_config.go` |
| `smith.toml` | `[tasks.<task>]`, `[projects.<p>.tasks.<task>]` | override an inferred task | host |

The scaffold constants (Java 21, JUnit, Kotlin, AGP and SDK levels) are in
`scaffolding/*_config.go` and edited in the generated files afterwards. CI
installs Maven 3.9.11 and Gradle 9.8.1 (`pipeline/ci_contribution_use_case.go`);
neither is configurable.

## Toolchain

- **Probe.** `java -version`, as `java <version>`. Maven, Gradle and the Android
  SDK are not part of the cache key; only doctor checks them.
- **Pin.** `[toolchains] java = "17"` is checked by `smith doctor`. CI does not
  read it; use `.java-version`. With `17.0.9` in `.java-version`, the generated
  workflows said `java-version: '17'`, Azure `versionSpec: '17'`, GitLab
  `.../latest/17/ga/...`.
- **Needed on the machine.** A JDK; Maven for `pom.xml` modules; Gradle for
  Gradle modules (smith does not use a Gradle wrapper); the Android SDK for an
  Android module.
- **Where output goes.** `dist/maven/<artifactId>` for Maven (the root pom
  sets it) and `dist/gradle/<artifactId>` for Gradle (the generated scripts set
  it). Maven's `install` also writes to `~/.m2`. Gradle's caches
  (`.gradle/`, `.kotlin/`) stay in the workspace root and are ignored by git.

## Releasing

| | |
|---|---|
| Version in the manifest | Maven: the module's own `<version>`. Gradle: `version = "..."` in the script, else `version` in `gradle.properties` |
| Version write (Maven) | the module pom's own `<version>` (not the parent's) is replaced, or one is added after `<artifactId>`; every workspace pom that depends on the module with a `<version>` gets the new version too. Directories `.git`, `dist`, `node_modules`, `.smith` are not searched. A dry run lists the files and writes nothing |
| Version write (Gradle) | the script's top-level `version = ...` is replaced, or added after the `group` line or at the end; if the script has none and `gradle.properties` has `version`, that line is rewritten |
| Tag | `<project>@<version>`; `bare_tags = true` on one project tags `0.1.0` (for JitPack); `tag_prefix` overrides |
| Publish: not a package | "released by its tag" |
| Publish: app assets | "JVM release assets are not built yet: the app is released by its tag" |
| Publish: Maven package | `mvn -B -q -pl <g:a> deploy -DskipTests -s dist/maven/settings.xml -DaltDeploymentRepository=smith::<url>`; the settings define a server `smith` reading `MAVEN_USERNAME` and `MAVEN_PASSWORD` from the environment |
| Publish: Gradle package | `gradle --console=plain -q :path:publishAllPublicationsToSmithRepository` with the same variables set |
| Repository | `[registries.<name>] maven` named by `[release] registry` or `[projects.<p>] registry`, else `MAVEN_REPOSITORY_URL`. With neither the release stops: `set MAVEN_REPOSITORY_URL to the repository <name> deploys to` |
| Release channel | the plugin declares `channels`: Maven and Gradle order `2.0.0-beta.1` below `2.0.0` and take that spelling as it is, so a fixed version never picks it, and a repository has no channel tag. Maven sets `release` in `maven-metadata.xml` to the last version deployed, a prerelease too, so the registry query then reads the newest version without a `-` from the `versions` list. A build that asks for the `RELEASE` or `LATEST` keyword, or for an open version range, can still see a prerelease. JitPack builds each tag on its own, so a channel tag is a separate version there. |
| Registry query | `GET <repository>/<group path>/<artifactId>/maven-metadata.xml`, field `versioning/release` (the newest version of `versioning/versions` without a `-` when `release` is a prerelease); the repository is the named registry (signed in when credentials exist), else `MAVEN_REPOSITORY_URL`, else `https://repo1.maven.org/maven2` |
| Dry run | `smith release --dry-run` plans without deploying; the result says "dry run" and names the repository |

Private registries (`[registries.<name>]`, see
[configuration: private registries](../../../reference/configuration.md#private-registries)):

| `auth` | `maven` | Maven signs in with | Credentials |
|---|---|---|---|
| `github-packages` | `"github"` (the URL `https://maven.pkg.github.com/<owner>/<repo>` from `GITHUB_REPOSITORY`, else the `origin` remote) or an explicit URL | user `GITHUB_ACTOR` (else `x-access-token`), password `GITHUB_TOKEN`; the job needs `packages: write` | none to create |
| `token` | a URL | `MAVEN_USERNAME` and the variable named by `token_env` | those two secrets |
| `azure-artifacts` | the feed URL | `MavenAuthenticate@0` in the pipeline; the settings server id is the feed name | none (Maven only: Gradle needs `SYSTEM_ACCESSTOKEN` mapped into the release step) |

A real run with `[registries.github] auth = "github-packages", maven = "github"`
and `[release] registry = "github"`, `GITHUB_REPOSITORY=acme/tools`:

```text
$ smith release --dry-run --json      # the relevant fields
      "publish": "dry run",
      "registry": "https://maven.pkg.github.com/acme/tools",
warning: could not ask https://maven.pkg.github.com/acme/tools about sdk (401 Unauthorized), so its tags were not checked against what is published. ...
```

The 401 is GitHub Packages refusing an unauthenticated read; with no
registry the plan has no `registry` field. A real deploy was not run.

The release's SBOM lists the artifacts of `gradle.lockfile`, falling back to `pom.xml` and the Gradle build script, through the `sbom/components` method of the plugin protocol ([releasing](../../../reference/releasing.md#sbom-and-build-provenance)).

## CI

| What | Where it comes from |
|---|---|
| JDK | GitHub: `actions/setup-java@v5` with `distribution: 'temurin'`, the Java release and `cache: 'maven'` (`'gradle'` when no Maven module). Azure: `JavaToolInstaller@0` (`PreInstalled`). GitLab: Temurin from `api.adoptium.net` into `/opt/jdk`, and Maven 3.9.11 from `archive.apache.org` into `/opt/maven` |
| Maven | not installed on GitHub and Azure (the hosted images carry it); installed on GitLab as above |
| Gradle | when any module has a Gradle script: a step `Install gradle` that downloads `gradle-9.8.1-bin.zip` from `services.gradle.org`, unzips it into `$HOME` and puts its `bin` first on the `PATH` (also `$GITHUB_PATH` and the Azure logging command) on every provider |
| Android SDK | when a script mentions `com.android.`: a step `Install android-sdk` that, only if `ANDROID_HOME` is unset or missing, downloads the command-line tools (revision 16111833), accepts the licences and exports `ANDROID_HOME` |
| Setup step | Maven only: `Resolve Maven dependencies`: `mvn -B -q dependency:go-offline`, blocking. Gradle has none |
| Store | none beyond `setup-java`'s own cache |
| Credentials | for packages that publish: `MAVEN_REPOSITORY_URL`, `MAVEN_USERNAME`, `MAVEN_PASSWORD` with no registry (real output below); `GITHUB_TOKEN` and the permission `packages: write` for GitHub Packages; `MAVEN_USERNAME` plus the token variable for `token`; `MavenAuthenticate@0` for Azure Artifacts |
| Dev container | `ghcr.io/devcontainers/features/java:1` with `version`, `installMaven: "true"` (always) and `installGradle` (true when a Gradle module exists) |

A real GitHub workflow for a workspace with a package and no registry carries,
on the release step:

```yaml
          MAVEN_PASSWORD: ${{ secrets.MAVEN_PASSWORD }}
          MAVEN_REPOSITORY_URL: ${{ secrets.MAVEN_REPOSITORY_URL }}
          MAVEN_USERNAME: ${{ secrets.MAVEN_USERNAME }}
```

With the GitHub Packages registry the workflow had `packages: write` under
`permissions` and `GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}` instead, and
`smith ci check` explained that readers of the Maven package need their own
token with `read:packages`.

The Gradle and Android steps came from a real `smith ci generate` (GitHub) in a
Kotlin plus Android workspace:

```yaml
      - name: Install gradle
        run: |
          curl -fsSLo /tmp/gradle.zip https://services.gradle.org/distributions/gradle-9.8.1-bin.zip
          unzip -q -o -d "$HOME" /tmp/gradle.zip && rm /tmp/gradle.zip
          export PATH="$HOME/gradle-9.8.1/bin:$PATH"
          ...
      - uses: actions/setup-java@v5
        with:
          distribution: 'temurin'
          java-version: '21'
          cache: 'gradle'
```

## Doctor, dependencies and removal

**Doctor.** Run only when the workspace has JVM projects.

| Check | Fails when |
|---|---|
| `a JDK is installed` | `java -version` fails |
| `JDK is <release>, as in CI` | the JDK's major is older than the release CI installs (`.java-version`, the pom's `maven.compiler.release`, else `21`); a newer JDK is a warning |
| `Maven is installed` | (only with a `pom.xml` module) `mvn -v` fails |
| `Gradle is installed` | (only with a Gradle script) `gradle --version` fails |
| `the Android SDK is installed` | (only with `com.android.` in a script) neither `ANDROID_HOME`, `ANDROID_SDK_ROOT` nor `local.properties`' `sdk.dir` names an existing directory |

In the Maven workspace all passed. In the Kotlin plus Android workspace on this
machine (no Gradle, no Android SDK) the real output was:

```text
✖ Gradle is installed
      gradle is not on PATH
      → install Gradle 9.8.1 (https://gradle.org/install)
✖ the Android SDK is installed
      neither ANDROID_HOME nor local.properties' sdk.dir names an SDK
      → install the Android SDK command-line tools (https://developer.android.com/tools) and set ANDROID_HOME
```

**Dependencies.** `smith sync` and `smith up` read and rewrite the versions in
the build files themselves, so they need no Maven or Gradle plugin and work the
same with or without `mvn` or `gradle` on the machine.

- Maven: every `<dependency>` of a `pom.xml` with a `<version>`, directly and in
  `<dependencyManagement>`. A version that is a `${property}` the same pom
  defines is followed, and the property's value is what changes (so every
  dependency using it moves together). A dependency without a version takes the
  managed one and declares nothing. Ranges (`[1.7,2.0)`) and properties the pom
  does not define (the parent's) are reported and never rewritten.
- Gradle: the string notation of `build.gradle` and `build.gradle.kts`
  (`implementation("group:artifact:1.2.3")`, in either quote style). Version
  catalogs (`libs.versions.toml`), the map notation and variables are not read
  yet; a build that uses them keeps those versions to Gradle.
- `smith up` asks Maven Central (`maven-metadata.xml`) for the highest stable
  release (milestones, release candidates and snapshots are skipped; a backport
  shipped after a newer line does not win). An artifact Maven Central does not
  have (an internal repository's) is skipped. Nothing is installed afterwards:
  Maven and Gradle resolve the new version on the next build.
- Not offered: `smith install -w <module> <dependency>` and `smith uninstall`
  (edit the pom or build script), and a vulnerability audit (the usual tools,
  OWASP dependency-check for Maven and a Gradle plugin, have to be installed
  into the build first, so `smith ci audit` has nothing to run for JVM).

**Remove.** `smith remove <project>` deletes the directory, then the plugin takes
the module's `<module>` line out of the root `pom.xml` and its `include(...)` line
out of `settings.gradle.kts` (`--dry-run` lists the files it would change), and
smith regenerates the pipelines.

## Declared projects

The plugin does not declare `declared`. A Maven or Gradle module is one build unit, and the packages inside it cannot be built, tested or linted on their own, so a folder is not a project. smith refuses a `[projects.<name>]` declaration for `jvm` with "finds a project by its manifest only: give the folder a manifest of its own" (a `pom.xml` or `build.gradle(.kts)`, or a module of the build above it).

## Requirements and limits

- JDK 21 for new modules; Maven for Java; Gradle on `PATH` for Kotlin and
  Android (no wrapper is generated or used).
- Gradle modules are found by regular expressions over the script
  (`version =`, `group =`, `artifactId =`, `project("...")`, plugin ids): a
  version catalog alias or a script that builds these values indirectly is not
  seen.
- Maven `build` installs into `~/.m2`.
- Release assets for apps are not built.
- No `format` or `typecheck` task.
- Android and Kotlin: the scaffolds name AGP 9.4.1, Kotlin 2.4.21, ktlint
  plugin 14.2.0 and JUnit BOM 6.1.3; I could not check that these resolve,
  because Gradle was not available.

## Verified

Windows 11 Pro, smith 1.62.0, OpenJDK 21.0.12.1 (Temurin), Apache Maven 3.10.0.
The plugin's CI installs Maven 3.9.11; I ran 3.10.0.

Run for **Maven** (real): `smith new`, `smith add lib|app|package --lang java|jvm`,
`smith projects`, `smith graph` (with a real pom dependency), `smith explain`,
`smith run build`, `smith run test lint`, `smith run start -p tool`,
`smith doctor`, `smith ci generate` for GitHub, Azure and GitLab, with and
without `.java-version`, with a GitHub Packages registry, `smith ci setup`
(`mvn dependency:go-offline`), `smith ci check`, `smith release --dry-run`
(plain and `--json`), `smith remove`.

Run for **Gradle, Kotlin and Android** (generation and smith's side only):
`smith add --stack kotlin` (lib, app, package), `smith add app --stack android`,
the refusal of `--stack android` for a lib, `smith projects`, `smith explain`,
`smith doctor` (failing as expected), `smith ci generate` (Gradle and Android SDK
steps), `smith release --dry-run --json`, `smith remove`.

Not verified, and why: any `gradle` command (`gradle` is not installed. The
Flutter SDK cache holds a Gradle wrapper, but running it downloads a Gradle
distribution, which I did not do), so the Kotlin and Android builds, ktlint,
Android Lint, `gradle run`, the Gradle publish, and `settings.gradle.kts`
handling after a removal are code-derived; the Azure Artifacts and token
registries; a real `mvn deploy`; the live Maven Central query beyond the
doctor check passing.

## See also

- [Plugin catalogue](README.md)
- [What is a plugin](../what-is-a-plugin.md) and [managing plugins](../managing-plugins.md)
- [Plugins reference](../../../reference/plugins.md) and [stacks](../../../reference/stacks.md)
- [Releasing](../../../reference/releasing.md) and [CI pipelines](../../../reference/ci.md)
- [Configuration: private registries](../../../reference/configuration.md#private-registries)
- [Writing a language plugin](../../creating/language-plugin.md)
