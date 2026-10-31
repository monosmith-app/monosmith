# Creating a language plugin

A language plugin teaches smith a project type: how to find its projects, what they import, which
tasks they run, how to scaffold a new one, how to write its version and how to publish it. Once a
language plugin exists, the graph, the cache, `affected`, the pipelines and releases work for its
projects with nothing else. This guide builds a small one, `widget`, step by step, in Go. The same plugin
in TypeScript, on [the TypeScript SDK](typescript-sdk.md), is
[`language-widget-ts`](../../examples/plugins/language-widget-ts): it has the same methods, passes the
same 27 checks of `smith plugin check --generate` and answers with the same bytes. In Python, on
[the Python SDK](python-sdk.md), it is [`language-widget-py`](../../examples/plugins/language-widget-py), with
the same checks and the same bytes again, and in C#, on [the C# SDK](csharp-sdk.md), it is
[`language-widget-cs`](../../examples/plugins/language-widget-cs).

## When to write one

Write a language plugin when smith has no built-in plugin for a project type: a language it does
not know, an internal build system, a manifest of your own. If the language is known and you only
want another linter, test runner or rule, write a [toolchain](toolchain-plugin.md) or
[structure](structure-plugin.md) plugin instead; if you only want new projects to start a certain
way, a [scaffold](scaffold-plugin.md) or a [recipe](recipe-and-stack.md) is enough.

## What smith asks of a language plugin

| Method | Required | smith calls it to |
|---|---|---|
| `initialize` | yes | learn the plugin's name, kind, roles, languages and manifest file names |
| `discover` | yes | turn the manifests it found into projects, with tasks, imports and exports |
| `generate` | for `smith add --lang <you>` | write a new project's starter files |
| `version/write` | for releases | record the release version in the manifest |
| `publish`, `published` | for releases of packages | publish to a registry, and ask what the registry has |
| `ci` | for generated pipelines | learn what CI must install, cache and have credentials for |
| `doctor` | optional | collect health findings, with fixes |
| `remove` | optional | clean up outside a project's directory after `smith remove` |
| `project/kind` | optional | say what changing a project between app, lib and package edits, for `smith kind` |
| `sbom/components` | optional | list the components a project ships (lockfile, else manifest) as package URLs, for the SBOM a release attaches |
| `dependencies/*` | optional | `smith install`, `uninstall`, `sync`, `up` |

The handshake must declare `kind: "language"`, at least one entry in `manifests` (the file names, or
patterns such as `*.csproj`, whose presence may make a directory a project) and `capabilities.discover`.
The full shapes are in [the cheat sheet](protocol-cheat-sheet.md).

## The example

A **widget** is any directory with a `widget.json`:

```json
{
  "name": "shop",
  "version": "0.0.0",
  "kind": "app",
  "description": "A widget of the lw workspace.",
  "license": "MIT",
  "uses": ["core"]
}
```

`name` names the project, `kind` is `app`, `lib` (the default) or `package`, and `uses` lists the
widgets it imports. The plugin gives each widget a cached `build` task (writes `widget.txt` into
`dist/`) and a `test` task (checks the manifest). The source is in
[`docs/examples/plugins/language-widget`](../../examples/plugins/language-widget); read its
[README](../../examples/plugins/language-widget/README.md) for how to build it.

```text
docs/examples/plugins/language-widget/
  go.mod              requires the SDK (a replace line points at this repository's copy)
  main.go             serve the protocol when smith starts us, run the tool otherwise
  widget.go           the plugin: Describe, Discover, Generate, WriteVersion, Doctor
  manifest.go         reading and rewriting widget.json
  tool.go             the command line the tasks run: check, build
  widget_test.go      conformance suite + behaviour tests
  plugin.json         how to install it
```

The tasks run the plugin's own executable (`os.Executable()`), so the example needs no other tool.
A real plugin names the tool it wraps (`cargo`, `dotnet`, a script) and says in its documentation
that it must be installed; a task that names a path under a user's home makes cache keys differ
between machines.

## Step 1: say who you are (`initialize`)

```go
func (w *widget) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	w.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: "widget", Version: "0.1.0"},
		Kind:   protocol.KindLanguage,
		// A language plugin that also writes new projects plays the scaffold
		// level too; it does not need to be chosen as a toolchain.
		Roles:     []protocol.Role{protocol.RoleLanguage, protocol.RoleScaffold},
		Languages: []string{"widget"},
		Manifests: []string{manifestFile},
	}, nil
}
```

| Field | Why |
|---|---|
| `Plugin.Name` | what users write in `smith.toml`, and in `--lang`-less stacks; `smith.toml` must declare the plugin under this name |
| `Plugin.Version` | goes into every task's cache key, so upgrading the plugin re-runs its tasks. Use semver |
| `Roles` | declare them. A language plugin that declares none is also toolchain **and** scaffold, and the check `a scaffold generates projects` fails if it cannot `generate` |
| `Languages` | the values of `smith add --lang`; also what other plugins' `languages` and `requires` match |
| `Manifests` | file names or patterns. smith lists files with `git ls-files` plus untracked, non-ignored files, so `.gitignore` is respected for you |

`serve.Main` fills in `protocolVersion` and the `capabilities`, from the methods the type has.
`p.WorkspaceRoot` is absolute with the host's separators; keep it, you need it to read files.

## Step 2: find projects (`discover`)

smith passes the manifests that match your names, workspace-relative with forward slashes. Return
one `Project` per manifest that is yours; leave out one that is not (the plugin must not guess).

```go
func (w *widget) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, rel := range p.Manifests {
		m, err := readManifest(filepath.Join(w.root, filepath.FromSlash(rel)))
		if err != nil {
			return result, err
		}
		if m.Name == "" {
			continue // not a widget manifest: leave it out
		}
		kind := protocol.ProjectKind(m.Kind)
		if kind == "" {
			kind = protocol.ProjectLib
		}
		root := path.Dir(rel)
		result.Projects = append(result.Projects, protocol.Project{
			Name: m.Name, Root: root, Language: "widget", Kind: kind,
			Version: m.Version, Description: m.Description, Stack: m.Stack,
			Exports: []string{m.Name},
			Imports: m.Uses, // the host keeps the ones another project exports
			Tasks:   map[string]protocol.Task{ /* build and test: step 3 */ },
		})
	}
	return result, nil
}
```

What each field is for:

| Field | Meaning |
|---|---|
| `Name` | unique across the workspace; what people type (`smith run build -p shop`) |
| `Root` | the project's directory, workspace-relative, forward slashes (`path.Dir`, not `filepath.Dir`) |
| `Kind` | `app` (deployed, never published), `lib` (internal) or `package` (published) |
| `Exports`, `Imports` | the graph. A project **exports** identifiers others import it by (an npm name, a Go module path); it **imports** the identifiers its sources use. smith keeps the imports that some project exports and draws those edges. The plugin never computes edges itself |
| `Version` | the version recorded in the manifest |
| `Stack` | the stack recorded in the manifest, so a project made with a stack keeps it |
| `Toolchain` | optional: the compiler or runtime version the plugin probed (`"node 20.11.0"`). It goes into every task's cache key, so a result from another version is a miss. `SetToolchain` on the result probes once, and only when there is a project |
| `TagPrefix` | optional: the release tag prefix; empty means `<name>@` |
| `PublishTo` | optional: where a package publishes, `none` for never |

A project whose imports resolve to nothing is fine: leave `Imports` empty. Real output with the
example (`shop` uses `core`):

```text
$ smith projects
core                 lib      widget   libs/core
shop                 app      widget   apps/shop
$ smith graph
core  lib · widget · libs/core
shop  app · widget · apps/shop
  └─ core
```

### Optional: report a shape so new projects follow their siblings

A project in `discover` may carry `shape`: a map of short facts about how it is made, for example
`"build.tool": "esbuild"`, `"test.runner": "@swc/jest"`, `"exports.target": "dist"`. Choose the keys
and values; the host knows none of them. `smith add` counts the shapes of the projects of the new
project's kind and language. When at least 3 of them, and at least two thirds, share a value, it
prints the reason and sends the value back in `generate` as `params.shape`. Apply the keys you can,
list them in `initialize.follows` (the host asks about a split only for those keys), and ignore the
rest. The host reads the shape of the new project after `generate` and warns for each key that
differs. See [ADR 0040](../../architecture/decisions/0040-smith-add-follows-the-dialect-of-sibling-projects.md).

## Step 3: tasks that cache honestly

A task is a command smith runs, hashes and caches. The plugin describes it; it never runs it.

```go
"build": {
	Command:     []string{self(), "build", protocol.BuildDirFromProject(root)},
	Inputs:      []string{manifestFile},
	Outputs:     []string{protocol.BuildDirFromProject(root)},
	DependsOn:   []string{"^build"},
	Cache:       true,
	Description: "write widget.txt into the build directory",
},
"test": {
	Command:     []string{self(), "check"},
	Inputs:      []string{manifestFile},
	Cache:       true,
	Description: "check the manifest: a name and an x.y.z version",
},
```

| Rule | Why |
|---|---|
| `Command` is an argv array | it is never passed through a shell. `{smith}`, `{workspaceRoot}` and `{projectRoot}` are replaced when the task runs |
| builds go to `protocol.ProjectBuildDir(root)` | `dist/<project dir>`; declare it in `Outputs` with `BuildDirFromProject(root)` (`../../dist/libs/core`, relative to the project). `smith doctor` fails an output elsewhere, and two tasks may not write the same directory |
| `DependsOn: ["^build"]` | `^build` runs the build of every project this one imports first; `build` alone is this project's own |
| `Cache: true` only when the result depends on the inputs alone | no timestamps, no absolute paths in the command, no reading outside `Inputs` |
| `Inputs` | globs relative to the project root; empty means every tracked or untracked-but-not-ignored file |
| `Description` | shown by `smith explain` and the TUI |

```text
$ smith run build test
core:build ▶ <dir>\smith-plugin-widget.exe build ../../dist/libs/core
core:test ▶ <dir>\smith-plugin-widget.exe check
shop:test ▶ <dir>\smith-plugin-widget.exe check
shop:test │ shop 0.0.0 is valid
core:test │ core 0.0.0 is valid
core:build │ wrote ../../dist/libs/core/widget.txt
...
✔ 4/4 tasks succeeded (0 from cache) in 48ms
$ smith run build test
shop:build │ wrote ../../dist/apps/shop/widget.txt
shop:build ✔ from cache
✔ 4/4 tasks succeeded (4 from cache, 65ms saved) in 6ms
$ smith explain core:build
core:build — write widget.txt into the build directory
  command    <dir>\smith-plugin-widget.exe build ../../dist/libs/core
  cwd        libs/core
  dependsOn  ^build (build of the projects it imports, first)
  inputs     widget.json
  outputs    ../../dist/libs/core
  env        —
  cache      true

Every value above comes from plugin widget, except where marked.
```

(`<dir>` stands for the directory of the example.) Task names are free, but a few have meaning
across smith: `build`, `test`, `lint`, `typecheck`, `start`, `package`; see
[Tasks, cache and affected](../../reference/tasks-and-cache.md).

## Step 4: scaffold (`generate`)

`smith add lib core --lang widget` calls `generate` with the kind, name and directory (which does
not exist yet), the workspace information and the stack:

```go
func (w *widget) Generate(p protocol.GenerateParams) (protocol.GenerateResult, error) {
	data := map[string]string{
		"Dir": p.Directory, "Name": p.Name, "Kind": string(p.Kind), "Stack": p.Stack,
		"Workspace": p.Workspace.Name,
	}
	files, err := templating.Write(w.root, map[string]string{
		"[[.Dir]]/widget.json": `{
  "name": "[[.Name]]",
  "version": "0.0.0",
  "kind": "[[.Kind]]",
  "description": "A widget of the [[.Workspace]] workspace.",
  "license": "MIT"[[if .Stack]],
  "stack": "[[.Stack]]"[[end]]
}
`,
		"[[.Dir]]/README.md": "# [[.Name]]\n\nA widget. Edit widget.json.\n",
	}, data)
	if err != nil {
		return protocol.GenerateResult{}, err
	}
	return protocol.GenerateResult{Files: files}, nil
}
```

| Rule | Why |
|---|---|
| **record `p.Stack` in the manifest** | `discover` must report it back (`Stack`); the conformance check `the generated lib records its stack` fails otherwise |
| report every file in `Files`, workspace-relative | smith lists them, and the conformance check verifies each exists |
| never overwrite | `templating.Write` writes nothing if any target exists and returns `file already exists` |
| use the SDK's `templating` package | templates use `[[ ]]` delimiters, so `{{ }}` in your output is untouched; target paths can be templates too. In TypeScript: `writeTemplates` or `scaffoldProject`; in Python: `write_templates` or `scaffold_project`; in C#: `Templates.Write` or `Scaffold.Project` |
| `p.Convention` | true when the plugin is asked as a workspace convention rather than as part of a stack; a language plugin ignores it |
| `Notes` | lines printed to the user after generation |

```text
$ smith add lib core --lang widget
Creating lib core in libs/core: widget
Wrote 2 file(s):
  libs/core/README.md
  libs/core/widget.json
Try: smith test -p core
```

## Step 5: write the release version (`version/write`)

`smith release` computes a version from conventional commits and asks the plugin to record it:

```go
func (w *widget) WriteVersion(p protocol.VersionParams) (protocol.VersionResult, error) {
	file := path.Join(p.Project.Root, manifestFile)
	if !p.DryRun {
		if err := setField(filepath.Join(w.root, filepath.FromSlash(file)), "version", p.Version); err != nil {
			return protocol.VersionResult{}, err
		}
	}
	return protocol.VersionResult{Files: []string{file}}, nil
}
```

A dry run lists the files it would change and changes none. `p.Dependencies` maps the identifiers of
other workspace projects to their new versions, for a manifest that records dependency ranges; the
example has none. The SDK's `jsondoc` package edits JSON in place and keeps the key order and layout,
so a release does not reformat the file or turn CRLF lines into LF (in TypeScript and Python, `JsonDocument`; in C#, `EditableJsonDocument`).

## Step 6: health checks with fixes (`doctor`)

```go
finding := protocol.Finding{Check: project.Name + " names a license", OK: m.License != ""}
if !finding.OK {
	finding.Detail = file + " has no license field"
	finding.Remedy = `add "license": "MIT" (or your license) to ` + file
	// ... build the new file content with jsondoc ...
	finding.Fix = &protocol.Fix{
		Description: "add an MIT license to widget.json",
		Safe:        false, // it picks a license for the team, so smith asks first
		AppliesTo:   []string{project.Name},
		Files:       []protocol.FileEdit{{Path: file, Content: string(doc.Bytes())}},
	}
}
```

Every failed finding must name its remedy. A fix carries the whole new content of each file, so smith
can show a diff. Mark a fix `Safe` only when it restores something the plugin owns and loses nothing
of the user's; `smith doctor --fix` applies only safe fixes unasked. A fix may also carry `Steps`
(commands, no shell) and `Choices` (questions): see below.

### Doctor fixes: files, steps and choices

A fix is `{description, safe, files, steps, choices, applies_to}` and every part but `description` is
optional. A host that predates `steps` and `choices` ignores them, so mark a fix `safe` only when its
`files` alone are sound.

| Part | Does |
|---|---|
| `files` | written first, as whole new contents (`"delete": true` removes one) |
| `steps` | run after the files, in order, from the workspace root (or the step's `dir`): `{"description": "install the missing packages", "argv": ["npm", "install", "-D", "oxfmt"]}`. There is no shell; the first step that fails stops the fix and smith shows its output |
| `choices` | questions the user answers before anything applies, in order: `{"question": "oxfmt is not installed", "options": [{"label": "install it", "fix": {"steps": [...]}}, {"label": "remove the script", "fix": {"files": [...]}}]}`. The picked option's files and steps are added after the fix's own; the first option is the suggested one; an option's fix may carry choices of its own. A fix with choices is never applied unasked, whatever `safe` says. Because options write files, give an option the whole new content of every file it changes, and do not let two options of different questions edit the same file |
| `applies_to` | the projects the fix is for. Fixes without choices that share a `description` and name projects are folded into one, which smith presents as "applies to 22 projects": give the same description to the same kind of fix on every project. A fix that edits one shared file for many projects lists them all here and may be returned by each project's finding, since identical fixes are one |

```json
{"description": "declare the packages each project imports in its package.json",
 "safe": false, "applies_to": ["ocr"],
 "files": [{"path": "packages/ocr/package.json", "content": "{...}"}],
 "steps": [{"description": "install what package.json declares", "argv": ["npm", "install"]}]}
```

Mark a fix `safe` when it only restores what the plugin owns and installs nothing; a fix that runs a
package manager, edits a file the team wrote, or changes what git tracks is asked about. See
[ADR 0033](../../architecture/decisions/0033-doctor-fixes-with-steps-and-choices.md).

The example's finding, run:

```text
$ smith doctor
✖ core names a license
    libs/core/widget.json has no license field
    → add "license": "MIT" (or your license) to libs/core/widget.json
...
$ smith doctor --fix --only "core names a license" --yes
$ cat libs/core/widget.json
{
  "name": "core",
  "version": "0.0.0",
  "kind": "lib",
  "description": "A widget of the wd workspace.",
  "license": "MIT"
}
```

## Methods the example leaves out

A complete language plugin usually also implements these. The built-in plugins in `libs/plugin-*`
show how; the shapes are in [the cheat sheet](protocol-cheat-sheet.md).

| Method | What to return, and the rules |
|---|---|
| `publish` | `Published: true` with the `Registry`, or `Skipped` with a reason for a version already there. **Skipping with a reason is a success**, and it is what makes a re-run of a failed release safe. Honour `DryRun`. For an app with `Assets`, build the release files instead of publishing to a registry |
| `published` | the newest version the registry has of the project (`Version` empty when it has none), or `Unreachable: true` with a `Detail` when it could not be asked. smith checks release tags against it |
| `ci` | `Toolchains` (a runtime to install, with GitHub and Azure steps or an `Install` command), package `Stores` to cache, `Setup` and `Audit` steps, `Credentials` the release needs, `Permissions`. Only list credentials for projects that publish (`Project.Publishes()`) |
| `remove` | after smith deleted the project's directory, clean what you own outside it (a workspace file, a lockfile) and list the files you changed |
| `project/kind` | `smith kind`: answer, without writing anything, the whole new content of every file the change edits (the manifest's kind and publish guard, the references other projects and the workspace make to the folder `newRoot`), as paths before the move; the host applies them and moves the folder. Say `refused` when the kind makes no sense for the language, and `warnings` for what the change loses. The SDK helpers `pathrefs` (retarget, rebase and follow relative paths), `tomltext`, `xmltext` and `jsondoc` (edit one key or element and keep the file as written) and `fileedits` (the set of edited files) do the rewriting |
| `dependencies/install`, `/uninstall` | add and remove packages with the ecosystem's own tool |
| `dependencies/declared`, `/rewrite`, `/latest` | list declared dependency specs, rewrite them in place, and report the newest releases: what `smith sync` and `smith up` need |

A plugin that reads `Project.Registries` points its package manager at the private registry smith.toml
names, and publishes there. `Project.Settings[<plugin name>]` holds the plugin's own settings for the
project.

## Run and check it

```sh
cd docs/examples/plugins/language-widget
GOWORK=off go build -o smith-plugin-widget .
smith plugin check ./smith-plugin-widget --generate
```

(`GOWORK=off` only because the folder sits inside the MonoSmith repository, whose `go.work` does not
list it; on Windows the file is `smith-plugin-widget.exe`.) Real output:

```text
widget 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: language
  ✔ plays a known role: scaffold
  ✔ a language plugin discovers projects
  ✔ a scaffold generates projects
  ✔ declares a concurrency the protocol defines
  ✔ answers 3 requests sent together, one at a time
  ✔ discover answers with no manifests
  ✔ doctor answers with no projects
  ✔ generates an app
  ✔ generates a lib
  ✔ generates a package
  ✔ discovers the generated app
  ✔ the generated app keeps its kind
  ✔ the generated app records its stack
  ✔ the generated app's tasks have commands
  ✔ writes the app's version (dry run)
  ✔ discovers the generated lib
  ...
  ✔ writes the package's version (dry run)
  ✔ shuts down when asked
```

With `--generate` the plugin writes an app, a lib and a package in a temporary directory and a
language plugin must discover them back, with their kinds, the stack it was asked to record, and
tasks that have commands. It also calls `version/write`, `remove` and `publish` in dry run when the
plugin has them.

## Use it

```toml
# smith.toml
[plugins.external.widget]
command = ["/path/to/smith-plugin-widget"]
```

Trust it once (`smith plugin trust widget`; smith starts an external plugin only after you do, see
[Plugin security](../using/plugin-security.md)). Then `smith add lib core --lang widget`,
`smith add app shop --lang widget`, and the commands above.
To install it by name instead, put it on the search path with a `plugin.json`; see
[Packaging and distributing a plugin](packaging-a-plugin.md).

## Test it

The example's `widget_test.go` connects to the plugin in process with `host.Connect`, runs the
SDK's conformance suite, and then tests behaviour: the graph edge `shop → core`, a dry-run
`version/write` that changes nothing, and a doctor finding with a fix.

```text
$ GOWORK=off go test ./...
ok  	example.com/smith-plugin-widget	0.519s
```

See [Testing a plugin](testing-a-plugin.md) for the approach and the golden-file option.

## Common mistakes

| You see | Cause and remedy |
|---|---|
| `✖ a scaffold generates projects` | the plugin declares no `Roles`, so it is also a scaffold, but has no `Generate`. Declare `Roles: [language]`, or implement `Generate` |
| `✖ a language plugin discovers projects` | `Manifests` is empty or `Discover` is not implemented |
| `✖ the generated lib records its stack` / `discovered with stack ""; generate must record params.stack in the manifest` | `Generate` ignores `p.Stack`, or `Discover` does not return the manifest's stack in `Project.Stack` |
| `✖ the generated app's tasks have commands` | a task has an empty `Command` |
| `✖ generates a lib`: `files reported but not written` | `Files` names a path that does not exist under the workspace |
| the project appears but has the wrong root on Windows | `Root` used `filepath.Dir`; use `path.Dir` on the forward-slash path smith gave you |
| `smith doctor`: an output outside `dist/` | a build writes somewhere else; use `protocol.ProjectBuildDir` |
| every call fails after a good start | something wrote to stdout. Log to stderr |
| `smith.toml [plugins.external.widget]: the plugin calls itself something else; declare it under its own name` | the name in `smith.toml` differs from `Plugin.Name` |
| tasks re-run after every build of the plugin | the task command contains a path or timestamp that changes |

## See also

- [The plugin protocol cheat sheet](protocol-cheat-sheet.md), [The TypeScript SDK](typescript-sdk.md), [The Python SDK](python-sdk.md), [The C# SDK](csharp-sdk.md)
- [Testing a plugin](testing-a-plugin.md), [Packaging and distributing a plugin](packaging-a-plugin.md)
- The built-in language plugins: [the catalogue](../using/catalogue/README.md), and their source in `libs/plugin-*`
- [Creating a toolchain plugin](toolchain-plugin.md), the next level up
