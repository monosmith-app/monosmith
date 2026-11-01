# The ruby plugin

The ruby plugin makes every `*.gemspec` below the workspace root a smith project (a gem), and resolves all of them with one root `Gemfile` and one `Gemfile.lock`. It runs RuboCop and RSpec through Bundler, starts an app gem from its `exe/` executable, installs Ruby in CI, and publishes package gems to RubyGems with `gem build` and `gem push`. It is a built-in plugin: no installation needed.

## At a glance

| | |
|---|---|
| Plugin name | `ruby` (built-in, version follows smith, 1.62.0 here) |
| Roles | `language`, `toolchain`, `scaffold` (all three by default for a language plugin) |
| `--lang` values | `ruby` |
| Manifest that makes a directory a project | `*.gemspec` (a gemspec at the workspace root is skipped; one gem per directory) |
| Requires / conflicts / provides | none / none / no task slots |
| Capabilities advertised | `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `remove`, `published`, `dependencies`, `declarations`, `latest`, `changeKind`, `components`. |
| Stacks it publishes | none |
| Project kinds it can create | `app`, `lib`, `package` (`ruby-app`, `ruby-lib`, `ruby-package`) |

Capabilities are from `smith plugin list --json`.

## What it detects

(`libs/plugin-ruby/discovery/`) The gemspec is read with regular expressions, not by running Ruby, so a value must be a quoted string literal on one line.

| Field | How it is read |
|---|---|
| Project | the first gemspec in each subdirectory (a second gemspec in the same directory is ignored) |
| Name | the last element of the directory (`packages/kit` is `kit`) |
| Kind | `spec.metadata['smith_kind'] = 'app' \| 'lib' \| 'package'`; else the top directory (`apps/`, `libs/`, `packages/`); else `lib` |
| Stack | `spec.metadata['smith_stack']` |
| Exports | the gem name, `spec.name = '...'` (the directory name when absent) |
| Imports (graph edges) | the first argument of every `add_dependency`, `add_runtime_dependency` and `add_development_dependency` call, sorted. They become edges where a name matches another gem's export. |
| Version | `spec.version = '...'` (a computed version, such as `Foo::VERSION`, reads as empty) |
| Executable | the first entry of `spec.executables = [...]`; it decides the `start` task |
| Specs | whether a `spec/` directory exists; it decides the `test` task |
| Toolchain | `ruby <version>` from `ruby --version` |

Real output after `smith install -w cli rb_demo_core`:

```
$ smith graph
cli  app · ruby · apps/cli
  └─ core
core  lib · ruby · libs/core
kit  package · ruby · packages/kit
```

## Tasks

All tasks run in the gem directory through Bundler (`bundle exec`), so they use the root `Gemfile.lock`. Cached tasks' inputs: `**/*`, `{workspaceRoot}/Gemfile`, `{workspaceRoot}/Gemfile.lock`, `{workspaceRoot}/.rubocop.yml`.

| Task | Command | Appears for | Cached |
|---|---|---|---|
| `lint` | `bundle exec rubocop` | every gem | yes |
| `test` | `bundle exec rspec` | a gem with a `spec/` directory | yes |
| `start` | `bundle exec ruby exe/<executable>` | an `app` whose gemspec lists an executable | no |

That is all. There is no `build`, `package`, `typecheck` or `format` task: `smith run build` answers `no selected project has a [build] task`. The gem is built only at release time, by the publish step. No task has `dependsOn` or declared outputs.

With `[conventions.vertical-slices]` each gem also gets `slice-check` (inputs `lib/**/*.rb`, `spec/**/*.rb`, `test/**/*.rb`) and `lint` runs after it; see [vertical-slices](vertical-slices.md).

Real output:

```
$ smith explain cli
cli has 4 tasks:
  lint         bundle exec rubocop  (cached, after slice-check)
  slice-check  smith slices check --lang ruby .  (cached)
  start        bundle exec ruby exe/rb_demo_cli
  test         bundle exec rspec  (cached)

$ smith explain kit:test
kit:test — RSpec
  command    bundle exec rspec
  cwd        packages/kit
  dependsOn  —
  inputs     **/*, {workspaceRoot}/Gemfile, {workspaceRoot}/Gemfile.lock, {workspaceRoot}/.rubocop.yml
  outputs    —
  env        —
  cache      true
```

`smith run test` (1 example, 0 failures per gem), `smith run lint` (`5 files inspected, no offenses detected` per gem) and `smith run start -p cli` (`Hello, world!`) succeeded.

## Creating a project

```
smith add app|lib|package <name> --lang ruby
```

The first gem writes the root `Gemfile` and `.rubocop.yml`. Each gem is added to the `Gemfile` as `gemspec path: '<dir>'`, then `bundle install` runs (it needs the network the first time), and `Gemfile.lock` is written. The gem name is the workspace name and the project name in snake_case joined with `_` (`rb-demo` + `core` = `rb_demo_core`); the Ruby module is its CamelCase form (`RbDemoCore`).

Real output:

```
$ smith add lib core --lang ruby
Adding lib core (ruby) in libs/core
Fetching gem metadata from https://rubygems.org/...........
Resolving dependencies...
...
Bundle complete! 4 Gemfile dependencies, 24 gems now installed.
Wrote 10 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  .rubocop.yml
  Gemfile
  Gemfile.lock
  libs/core/lib/rb_demo_core.rb
  libs/core/lib/rb_demo_core/greeting/greet_use_case.rb
  libs/core/lib/rb_demo_core/greeting/greeting.rb
  libs/core/rb_demo_core.gemspec
  libs/core/spec/rb_demo_core/greeting/greet_use_case_spec.rb
Depend on it from another gem with spec.add_dependency 'rb_demo_core' (the workspace's Gemfile resolves it locally).
Try: smith test -p core

$ smith add app cli --lang ruby      # adds apps/cli/exe/rb_demo_cli, no .rubocop.yml again
```

The `Gemfile` after three gems:

```ruby
source 'https://rubygems.org'

gemspec path: 'libs/core'
gemspec path: 'apps/cli'
gemspec path: 'packages/kit'

group :development, :test do
  gem 'rake', '~> 13.0'
  gem 'rspec', '~> 3.13'
  gem 'rubocop', '~> 1.70', require: false
end
```

| Generated | What it is |
|---|---|
| `<dir>/<gem>.gemspec` | name, `version = '0.0.0'`, summary, authors (the workspace name), MIT, `required_ruby_version = '>= 3.2'`, `rubygems_mfa_required`, `smith_kind` (and `smith_stack`). Apps and libs also get `allowed_push_host = 'none'` ("released by its tag, never pushed to a gem server"). An app lists `exe/*` and `executables`. |
| `<dir>/lib/<gem>.rb`, `lib/<gem>/greeting/` | a worked slice: `greeting.rb` (requires), `greet_use_case.rb` with `Greeting.greet(name)` returning `Hello, <name>!` |
| `<dir>/spec/<gem>/greeting/greet_use_case_spec.rb` | one RSpec example |
| `<dir>/exe/<gem>` (app) | prints the greeting |
| `Gemfile`, `.rubocop.yml` (once) | the shared `Gemfile` (rake, rspec, rubocop pinned `~> 13.0`, `~> 3.13`, `~> 1.70`) and RuboCop settings (Ruby 3.2 target, `NewCops: enable`, LF line endings, `dist`, `node_modules`, `vendor` excluded) |

The plugin publishes no stack.

## Settings

The plugin has no `[projects.<p>.settings.ruby]` keys.

| Where | Key | Effect | Read in |
|---|---|---|---|
| gemspec | `spec.metadata['smith_kind']`, `['smith_stack']` | kind and stack | `discovery/gemspec_mapper.go` |
| gemspec | `spec.name`, `spec.version`, `spec.executables`, `add_*dependency` | export, version, `start`, graph edges | same |
| `Gemfile` | `gemspec path: '<dir>'` lines | which gems the shared bundle resolves; smith keeps the list in step | `scaffolding/gemfile_members_mapper.go` |
| `.rubocop.yml` | whole file | the RuboCop configuration of every gem (and a cache input) | RuboCop |
| `.ruby-version` | one line, `ruby-` prefix optional | the Ruby CI installs (default `3.4`) | `pipeline/ci_contribution_use_case.go` |
| `smith.toml` | `[toolchains] ruby = "3.4"` | version pin checked by doctor | host |
| `smith.toml` | `[release] rubygems_trusted_publishing = true` | publish with the workflow's OIDC token instead of `GEM_HOST_API_KEY` (GitHub only) | host / CI generator |
| env | `GEM_HOST_API_KEY` | RubyGems API key (read by `gem push`) | `releasing/release_gem_use_case.go` |
| env | `GEM_HOST` | push to another gem server instead of `https://rubygems.org` | same |

## Toolchain

The pin label is `ruby`, probed with `ruby --version`. Real output with `[toolchains] ruby = "3.3"` on a machine with 3.4.11:

```
✖ ruby matches the pinned version 3.3
    [toolchains] pins ruby 3.3, but this machine runs 3.4.11: tasks cached on another version are misses, and results can differ from CI
    → install ruby 3.3 (`mise use ruby@3.3`, asdf, or your installer; `smith doctor --fix` writes .tool-versions for mise and asdf), or change [toolchains] ruby
```

The machine needs `ruby` and `bundle` (Bundler). Generated gems require Ruby 3.2 or newer. There is no output directory for tasks: nothing is built. The gem file that a release builds goes to `dist/<gem dir>/<gem>-<version>.gem`.

## Releasing

| Aspect | Behaviour |
|---|---|
| Version | `spec.version = '...'` in the gemspec. The release rewrites that assignment (keeping the quote style), builds and pushes, then restores the file. A gemspec that computes its version is left alone (no change is made). |
| Tag | `<name>@<version>`, name being the directory name: `kit@0.1.0` (real). |
| What is published | `package` gems only: `gem build <gemspec> --output dist/<dir>/<gem>-<version>.gem`, then `gem push <file> --host <host>`. Apps and libs are released by their tag; an app with `[projects.<app>.release]` is tagged and gets a GitHub Release, without assets (`Ruby release assets are not built yet`). |
| Registry | `https://rubygems.org`, or `GEM_HOST`. `published` reads `<host>/api/v1/versions/<gem>/latest.json` (RubyGems answers `unknown` for a gem it has not got). |
| Credentials | `GEM_HOST_API_KEY` (required; listed when a package exists). With `[release] rubygems_trusted_publishing = true`, GitHub only: the workflow gets `id-token: write` and a `rubygems/configure-rubygems-credentials@v2.1.0` step before the release, and the secret is no longer listed. |
| Private gem servers | only through `GEM_HOST`; there is no `[registries]` support for gems. |
| Dry run | the plugin returns "dry run" and builds nothing. Real output: |

```
$ smith release --dry-run
Dry run: nothing is published or tagged.
  kit                  unreleased → 0.1.0  (1 releasing commit(s))

$ smith release --dry-run --json      # excerpt
  "tag": "kit@0.1.0",  "publish": "dry run",  "registry": "https://rubygems.org"
```

The release's SBOM lists the gems of `Gemfile.lock`, falling back to the `Gemfile` and the gemspec, through the `sbom/components` method of the plugin protocol ([releasing](../../../reference/releasing.md#sbom-and-build-provenance)).

## CI

Real output of `smith ci generate` (GitHub):

```yaml
      - uses: ruby/setup-ruby@v1
        with:
          ruby-version: '3.4'
      ...
        env:
          GEM_HOST_API_KEY: ${{ secrets.GEM_HOST_API_KEY }}
```

| Contribution | What |
|---|---|
| Toolchain | Ruby from `.ruby-version`, else `3.4`: `ruby/setup-ruby@v1` (GitHub), `UseRubyVersion@0` (Azure), `apt-get install ruby-full build-essential` and `gem install bundler` elsewhere, the `ruby` dev container feature |
| Package store | none: the plugin declares no download store and does not set `bundler-cache` |
| Setup (`smith ci setup`) | `bundle install`, blocking. Real: `▶ Install the bundle`, `Bundle complete! ...`, `smith ci setup passed in 600ms`. |
| Audit | `bundle-audit check --update` (bundler-audit) against `Gemfile.lock`, reported and not blocking, remembered per content of the `Gemfile`, `Gemfile.lock` and the gemspecs so it is skipped while they are unchanged. On a CI runner (`CI` or `TF_BUILD` set) a missing bundler-audit is installed first (`gem install bundler-audit`); on a developer machine the step says `bundler-audit is not installed: run gem install bundler-audit` and fails as a warning. |
| Credentials | `GEM_HOST_API_KEY` (required, when a package exists). `smith ci check`: `GitHub: add the repository secrets GEM_HOST_API_KEY`. Trusted publishing replaces it (above). |

## Doctor, dependencies and removal

Doctor checks (`pipeline/ci_contribution_use_case.go`), shown only when the workspace has Ruby gems:

| Check | Fails when |
|---|---|
| `Ruby is installed` | `ruby --version` fails; remedy: install Ruby (the `.ruby-version` or `3.4`) |
| `Bundler is installed` | `bundle --version` fails; remedy: `gem install bundler` |

Both passed in the test run (`16 passed, 0 failed`). Doctor does not compare the local Ruby with the CI version, and there is no check on the `Gemfile` listing every gem.

Dependencies (`dependencies/`): the plugin edits the gemspec, then runs `bundle install`.

| Command | What it does |
|---|---|
| `smith install` | `bundle install` for the workspace |
| `smith install -w <gem> <name>...` | adds `spec.add_dependency '<name>'` before the gemspec's closing `end` (`add_development_dependency` with `-D`; a gem already declared is left as is), then `bundle install`. No version constraint is written. |
| `smith uninstall -w <gem> <name>...` | removes those lines, then `bundle install` |

Real: `smith install -w cli thor` printed `✔ cli: added thor (apps/cli/rb_demo_cli.gemspec, Gemfile.lock)`; `smith uninstall -w cli thor` removed it again.

**Versions: `smith sync` and `smith up`.** The plugin lists the gems with a version constraint that each gemspec declares (`add_dependency`, `add_runtime_dependency`, `add_development_dependency`) and the workspace `Gemfile` declares (`gem 'rake', '~> 13.0'`; gems inside a `development` or `test` group count as development dependencies), so `smith sync`, `smith up` and the one-version check in `smith doctor` cover Ruby.

- A gem declared without a constraint has nothing to align and is left out. Several constraints on one gem (`'>= 2.0', '< 4'`) and pre-releases are reported and left alone.
- `smith sync` brings every declaration of a gem to the highest constraint declared, changing only the quoted constraint. A pessimistic constraint keeps its precision: moving `~> 1.2` to 1.4.1 writes `~> 1.4`, not `~> 1.4.1`.
- `smith up` asks rubygems.org (`/api/v1/versions/<gem>/latest.json`) for the newest release and runs `bundle install` afterwards, so `Gemfile.lock` follows. A gem rubygems.org does not have (a private server's) is skipped.

Removal: `smith remove <gem>` deletes the directory and takes its `gemspec path:` line out of the root `Gemfile`. Real dry run: `Would remove core (libs/core)`, `changes Gemfile`, `the pipelines would be regenerated`.

## Requirements and limits

- `ruby` and `bundle` on PATH; network for the first `bundle install`. Generated gems need Ruby 3.2 or newer; the test machine runs 3.4.11, CI defaults to 3.4.
- One bundle for the workspace: all gems share the same resolution, so two gems cannot pin conflicting versions of a dependency.
- The gemspec is parsed with regular expressions. Single-line, quoted-string values only; `spec.version = Foo::VERSION` or a dependency built in a loop is not seen.
- No `build`, `package`, `typecheck` or `format` task; no release assets; newer gems come from rubygems.org only, and the audit does not tell fixable findings from the rest, so it never blocks.
- Test and lint use the root `.rubocop.yml`. A gem without a `spec/` directory has no `test` task.

## Verified

Windows 11 Pro 10.0.26300, smith 1.62.0, ruby 3.4.11 (x64-mingw-ucrt), gem 3.6.9, Bundler 2.6.9; gems installed by the run: rubocop 1.92.1, rspec ~> 3.13, rake ~> 13.0. Throw-away workspace `rb-demo` from `smith new rb-demo`. Run: `smith add lib|app|package ... --lang ruby`, `smith projects`, `smith graph`, `smith explain`, `smith run build` (refusal), `test`, `lint`, `start`, `smith doctor` (clean, and with a wrong `[toolchains] ruby` pin), `smith ci generate` (and with `rubygems_trusted_publishing`), `smith ci setup`, `smith ci audit`, `smith ci check`, `smith release --dry-run [--json]`, `smith install -w` / `smith uninstall -w`, `smith up --check`, `smith remove core --dry-run --force`.

Not verified:

- A real `gem build` / `gem push` (needs a RubyGems key and would publish); read from `releasing/release_gem_use_case.go`.
- `GEM_HOST` against another gem server, and trusted publishing on a real GitHub runner (only the generated YAML was read).
- Azure and GitLab pipelines for Ruby; Linux and macOS.

## See also

- [The plugin catalogue](README.md)
- [vertical-slices](vertical-slices.md)
- [Releasing](../../../reference/releasing.md), [CI](../../../reference/ci.md)
- [Writing a language plugin](../../creating/language-plugin.md)
