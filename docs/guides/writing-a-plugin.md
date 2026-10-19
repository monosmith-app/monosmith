# Writing a plugin

This guide takes you from nothing to a working smith plugin, in Go, Python or
Node.js, and shows how any other language does the same. Every example in it
was run, and every `smith plugin check` result below is real output.

You need `smith` (see [installing](installing.md)) and git. For the examples,
Go 1.26+, Python 3 or Node.js (they were run with Go 1.27, Python 3.13 and
Node 24, on Windows).

## What a plugin is for

smith's core knows no language. Everything that depends on a technology lives
in a **plugin**: finding the projects of a language, the tasks they run,
scaffolding a new project, writing a version, publishing, what CI must install,
what `smith doctor` checks. The built-in plugins (Node, Go, Python, Rust and the
others) are ordinary plugins that ship inside the `smith` binary.

A plugin is **any program that talks to smith over its standard input and
output**. That is all. It does not link against smith, and it can be written in
any language. smith starts it as a child process when a command begins, asks it
questions, and stops it when the command ends.

Write a plugin when you want smith to:

- understand a project type it does not know (an internal build system, a
  language smith has no plugin for, a manifest of your own);
- add tasks, CI steps or doctor checks to projects of other languages (your
  company's lint rules, a license check);
- scaffold projects your team's way. For that, a [recipe](../reference/recipes.md)
  is usually enough, and it needs no code.

The deep dive is [the plugin system](../design/plugin-system.md). The reference
for the built-in ones is [plugins](../reference/plugins.md). Words such as
role, slot and stack are in the [glossary](../reference/glossary.md).

## Roles: what a plugin can be

A plugin declares its **roles** (also called levels) when it starts. A project
is made of one plugin per level:

| Role | Does | A project has | Typical methods |
|---|---|---|---|
| `language` | finds projects from manifests, writes versions, publishes | exactly one | `discover`, `version/write`, `publish` |
| `scaffold` | writes the starter files of a new project | at most one | `generate` |
| `toolchain` | decides how a project builds, tests, lints | any number, one per task name (**slot**) | `contribute`, `generate` |
| `structure` | checks layout and code rules | any number | `contribute`, `doctor` |
| `deploy` | adds pipeline steps and deploy targets | any number | `contribute`, `ci` |

There are two **kinds** of plugin:

- `language`: understands a language. It must say which manifest files make a
  directory a project (`package.json`, `go.mod`, or your own).
- `convention`: adds to projects another plugin discovered. Toolchains,
  structure rules, scaffolds and deploy targets are all of this kind.

If you declare no roles, a `language` plugin is language, toolchain **and**
scaffold, and a `convention` plugin is a structure plugin. That default bites:
a language plugin that does not implement `generate` fails the check
"a scaffold generates projects". Declare `"roles": ["language"]` when you do not
scaffold, as all three examples below do.

## The protocol in five minutes

The protocol is **JSON-RPC 2.0** over stdio, framed like the Language Server
Protocol. Version 1 is the only version.

### Framing

Each message is a header, an empty line and a JSON body. The length is in
**bytes** of the body, not characters:

```
Content-Length: 56\r\n
\r\n
{"jsonrpc":"2.0","id":4,"method":"shutdown","params":{}}
```

`\r\n` is a carriage return and a line feed. Both directions use it. Read the header lines until
the empty line, read exactly that many bytes, parse. `Content-Length` is the
only header smith sends or needs.

### Rules for stdout

- **Only protocol messages go to stdout.** One stray `print` corrupts the
  stream and the call fails. Log to **stderr**: smith shows it to the user.
- Do not buffer: flush after every message.
- smith sets `SMITH_PLUGIN=1` in the plugin's environment. Use it to make a
  program that is also a command-line tool behave as a plugin.

### Conversation

smith sends requests (with an `id`) and the plugin answers each with a result
or an error carrying the same `id`. Calls to one plugin are sequential.

1. smith sends `initialize` first. The plugin answers who it is and which
   optional methods it implements (its **capabilities**).
2. smith calls only the methods the plugin announced.
3. smith sends `shutdown`; the plugin answers `{}` and exits. Closing stdin also
   ends it.

A request without an `id` is a notification: do not answer it.

Errors use the JSON-RPC shape `{"code": ..., "message": ...}`:

| Code | Meaning |
|---|---|
| `-32700` | the message was not valid JSON |
| `-32600` | invalid request, e.g. `initialize` was not first |
| `-32601` | method not supported |
| `-32602` | the params do not have the expected shape |
| `-32603` | internal error |
| `-32000` | the method ran and failed for a reason the user must act on (a missing tool, a bad manifest). Put the remedy in `message`: smith shows it. |

Every call has a timeout (two minutes by default; `generate` and `publish` get
longer ones because they run real toolchains). A plugin that does not answer in
time is **killed**, because its stream can no longer be trusted. Paths in
messages are workspace-relative and use forward slashes; only `workspaceRoot`
is absolute and uses the host's separators.

### A real conversation

This is the actual traffic with the Python example below (bodies only; each is
framed as above). `>>` is smith to the plugin, `<<` the answer.

```json
>> {"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"protocolVersions": [1], "host": {"name": "smith", "version": "0.0.0"}, "workspaceRoot": "C:\\work\\demo"}}
<< {"jsonrpc": "2.0", "id": 1, "result": {"protocolVersion": 1, "plugin": {"name": "widget", "version": "0.1.0"}, "kind": "language", "roles": ["language"], "languages": ["widget"], "manifests": ["widget.json"], "capabilities": {"discover": true}}}

>> {"jsonrpc": "2.0", "id": 2, "method": "discover", "params": {"manifests": ["libs/hello/widget.json"]}}
<< {"jsonrpc": "2.0", "id": 2, "result": {"projects": [{"name": "hello", "root": "libs/hello", "language": "widget", "kind": "lib", "version": "1.0.0", "exports": ["hello"], "tasks": {"hello": {"command": ["git", "--version"], "inputs": ["widget.json"], "cache": true, "description": "say hello, as a widget does"}}}]}}

>> {"jsonrpc": "2.0", "id": 3, "method": "nope", "params": {}}
<< {"jsonrpc": "2.0", "id": 3, "error": {"code": -32601, "message": "method not supported: nope"}}

>> {"jsonrpc": "2.0", "id": 4, "method": "shutdown", "params": {}}
<< {"jsonrpc": "2.0", "id": 4, "result": {}}
```

## The methods

The Go types behind every message are in
`packages/plugin-sdk/protocol/`. Optional fields can be left out. Fields smith
does not know are ignored, so a plugin written today keeps working when the
protocol grows (an addition is a new capability or optional field; only a
breaking change raises the version).

| Method | Capability | smith calls it to |
|---|---|---|
| `initialize` | always | learn who the plugin is |
| `shutdown` | always | end the session |
| `discover` | `discover` | turn manifests into projects |
| `generate` | `generate` | scaffold a new project |
| `contribute` | `contribute` | add or replace tasks of projects |
| `version/write` | `version` | record a release version in a manifest |
| `publish` | `publish` | publish a project to its registry |
| `published` | `published` | ask what version the registry has |
| `ci` | `ci` | learn what a pipeline needs |
| `doctor` | `doctor` | collect health findings |
| `remove` | `remove` | clean up after a project was deleted |
| `dependencies/install`, `/uninstall` | `dependencies` | `smith install` and `smith uninstall` |
| `dependencies/declared`, `/rewrite` | `declarations` | `smith sync` and `smith up` |
| `dependencies/latest` | `latest` | `smith up`: newest releases |

### `initialize`

The only required method. Params:

```json
{
  "protocolVersions": [1],
  "host": {"name": "smith", "version": "1.4.0"},
  "workspaceRoot": "/home/me/acme"
}
```

Result, with every field:

```json
{
  "protocolVersion": 1,
  "plugin": {"name": "widget", "version": "0.1.0"},
  "kind": "language",
  "roles": ["language"],
  "languages": ["widget"],
  "manifests": ["widget.json", "*.widgetspec"],
  "provides": [],
  "requires": [{"level": "language", "name": "node"}],
  "conflicts": [{"name": "karma"}],
  "stacks": [{"name": "widget-app", "language": "widget", "toolchain": ["widget-lint"]}],
  "capabilities": {"discover": true}
}
```

| Field | Meaning | Default |
|---|---|---|
| `protocolVersion` | the version you chose from `protocolVersions`; must be `1` | required |
| `plugin.name`, `plugin.version` | identity. The name is what users write in `smith.toml`. Both are required (the check fails without them) | required |
| `kind` | `language` or `convention` | required |
| `roles` | the levels the plugin can fill | `language`: language, toolchain, scaffold. `convention`: structure |
| `languages` | the values users pass to `--lang` to reach this plugin | none |
| `manifests` | file names, or patterns like `*.csproj`, whose presence may make a directory a project. A language plugin must list at least one | none |
| `provides` | the task names (slots) a toolchain owns, so two plugins cannot both provide `test`. Only a toolchain may set it | none |
| `requires`, `conflicts` | plugins that must, or must not, share the project. A reference sets `level`, `name` and/or `slot`, at least one | none |
| `stacks` | ready-made [stacks](../reference/stacks.md) the plugin publishes | none |
| `capabilities` | the optional methods you implement: `discover`, `generate`, `version`, `publish`, `ci`, `doctor`, `contribute`, `remove`, `published`, `dependencies`, `declarations`, `latest` (booleans) | all false |

smith never calls a method whose capability is false. If the version list does
not contain `1`, answer with an error.

### `discover`

Params: `{"manifests": ["libs/hello/widget.json", ...]}`: the workspace-relative
files smith found whose names match your `manifests`. smith lists files with
`git ls-files` plus untracked, non-ignored files, so `.gitignore` is respected
for you. Result: `{"projects": [...]}`. Leave out a manifest that is not yours.

A **project**:

| Field | Meaning |
|---|---|
| `name` | unique in the workspace; what users type. Required |
| `root` | the project's directory, workspace-relative. Required |
| `language` | usually your `languages[0]` |
| `kind` | `app`, `lib` or `package` |
| `exports` | identifiers other projects import this one by (an npm name, a Go module path) |
| `imports` | identifiers its sources import. smith keeps those another project exports and turns them into graph edges (`smith graph`, `smith affected`) |
| `version` | the version recorded in the manifest |
| `toolchain` | e.g. `"node 20.11.0"`. Goes into every task's cache key: a different compiler is a cache miss |
| `description` | one line from the manifest |
| `tagPrefix` | release tag prefix; default `<name>@` |
| `tasks` | a map from task name to task (below) |
| `tags`, `stack` | free tags; the stack recorded in the manifest |

smith adds `settings` (your plugin's tables from `smith.toml`) and `registries`
(private registries) when it passes projects back to you.

A **task** is a command smith can run, hash and cache:

| Field | Meaning | Default |
|---|---|---|
| `command` | an argv array, never run through a shell. Required | |
| `cwd` | workspace-relative directory | the project root |
| `env` | extra environment variables | none |
| `inputs` | globs relative to the project root | every tracked or untracked-but-not-ignored file of the project |
| `outputs` | directories relative to the project root, restored on a cache hit; build output belongs under `dist/` | none |
| `dependsOn` | task names that run first: `build` is this project's, `^build` is every imported project's | none |
| `cache` | cache the result | `false` |
| `description` | shown by `smith explain` and the TUI | none |
| `verify` | `smith ci verify` runs it too | `false` |
| `serve`, `serveUrl`, `serveTimeout` | start another task (`web:start`) before this one, wait for the URL (`E2E_BASE_URL`), stop it after | none, `60s` |

Good tasks are cached honestly: declare `outputs` under `dist/`, never put a
timestamp or an absolute path into a command, and set `cache` only when the
result depends on the inputs alone. The cache key also includes your plugin's
version, so upgrading the plugin re-runs its tasks. See
[tasks and cache](../reference/tasks-and-cache.md).

### `contribute`

For `convention` plugins. Params: `{"projects": [<project>, ...]}`, the projects
whose stack includes the plugin, or that a `[conventions.<name>]` table applies
it to. Result, by project name:

```json
{
  "projects": {
    "hello": {
      "tasks":     {"rules": {"command": ["git", "--version"], "cache": true}},
      "replace":   {"lint":  {"command": ["my-linter"]}},
      "dependsOn": {"lint": ["rules"]}
    }
  }
}
```

`tasks` are added (a task the project already has is kept), `replace` overrides
or adds, `dependsOn` adds dependencies to existing tasks.

### `generate`

Scaffold a project. Params:

```json
{
  "kind": "lib", "name": "hello", "directory": "libs/hello",
  "workspace": {"name": "acme", "scope": "@acme", "modulePrefix": "github.com/acme/acme"},
  "stack": "widget-app", "convention": false, "settings": {}
}
```

Write the files under `directory` (it does not exist yet), **record `stack` in
the manifest** so `discover` reports it back, and answer
`{"files": ["libs/hello/widget.json"], "notes": ["run ..."]}` with
workspace-relative paths. `convention: true` means you are asked as a workspace
convention rather than as part of a stack: add only what every matching project
needs. `settings` is your table from a recipe.

### `version/write` and `publish`

```json
{"project": {"name": "hello", "root": "libs/hello"}, "version": "1.3.0",
 "dependencies": {"core": "1.3.0"}, "dryRun": true}
```

`version/write` records the version, and the new versions of workspace
dependencies, in the manifest. It answers `{"files": ["libs/hello/widget.json"]}`;
in a dry run it lists what it would change and changes nothing.

`publish` takes `project`, `version`, `dryRun` and, for releases of apps,
`assets`, `platform`, `commit`, `branch`. Answer
`{"published": true, "registry": "https://..."}`, or, for a version already
there, `{"published": false, "skipped": "version 1.3.0 already published"}`.
**Skipping with a reason is a success**, and it is what makes a failed release
safe to re-run. `published` takes `{"project": ...}` and answers
`{"version": "1.2.0", "registry": "..."}`, `{}` when the registry has nothing,
or `{"unreachable": true, "detail": "offline"}`.

### `ci`, `doctor`, `remove`

`ci` takes `{"projects": [...]}` and answers what a pipeline needs: `toolchains`
(a runtime, with its GitHub and Azure setup step or an `install` command),
`setup` and `audit` steps (`{"name", "command", "dir", "blocking"}`),
`postRelease` steps, `credentials` (`{"env", "description", "required"}`),
`permissions` and `stores` (package-manager download caches keyed on lockfiles).
See [CI pipelines](../reference/ci.md).

`doctor` takes `{"projects": [...]}` and answers findings:

```json
{"findings": [{
  "check": "widget.json has a license", "ok": false,
  "detail": "libs/hello/widget.json names none",
  "remedy": "add a \"license\" field",
  "warning": false,
  "fix": {"description": "add MIT", "safe": true,
          "files": [{"path": "libs/hello/widget.json", "content": "{...}"}]}
}]}
```

Every failed finding must name its remedy. A `fix` carries whole new file
contents so smith can show a diff; `smith doctor --fix` applies only fixes
marked `safe`, which restore what the plugin owns without losing anything of
the user's.

`remove` takes `{"project": ..., "dryRun": false}` after smith deleted the
project's directory, and answers the files it cleaned elsewhere
(`{"files": [...], "notes": [...]}`).

### Dependencies

`dependencies/install` takes `{"project": ..., "packages": [...], "dev": false}`
(with no project: `{"projects": [...]}`, install everything) and answers
`{"files": [...], "notes": [...]}`; `dependencies/uninstall` takes `project` and
`packages`. `dependencies/declared` takes `{"projects": [...]}` and answers
`declarations` (`{"project", "file", "name", "spec", "section", "peer",
"rewritable"}`) and the `installed` versions; `dependencies/rewrite` takes
`{"changes": [{"declaration": ..., "spec": "^2.0"}]}`; `dependencies/latest`
takes `{"names": [...], "tools": false}` and answers `latest`, `unreachable` and
`tools`. The built-in plugins show how: see `libs/plugin-node` and
`libs/plugin-python`.

## Installing a plugin: `plugin.json` and the search path

smith finds a plugin in one of two ways.

**Declare it in `smith.toml`**, by its command:

```toml
[plugins.external.widget]
command = ["python", "tools/widget/plugin.py"]   # any executable, on PATH or a path
```

or **install it** and name it, optionally with a version constraint
(`"1.2.3"`, `"^1.2"`, `"~1.2"`, `">=1.2"`):

```toml
[plugins.external.widget]
version = "^0.1"
```

An installed plugin is a directory holding a `plugin.json`:

```json
{
  "name": "widget",
  "version": "0.1.0",
  "description": "projects described by a widget.json",
  "command": ["./smith-plugin-widget{exe}"]
}
```

| Field | Meaning |
|---|---|
| `name`, `command` | required |
| `version`, `description` | shown by `smith plugin list` |

In `command`, `{os}`, `{arch}` and `{exe}` (`.exe` on Windows, empty elsewhere)
are filled in, so one directory can carry a binary per platform. **Only the
first element** is made relative to the plugin's directory, and only when it
starts with `./` or `../`. Later elements are passed as they are, to a process
whose working directory is wherever the user ran `smith`. So
`["python", "plugin.py"]` does not work from an installed directory; use a
launcher as the first element (below).

smith looks for `<dir>/<name>/plugin.json` in these places, the first match
winning:

1. the workspace's `.smith/plugins`
2. each directory of `SMITH_PLUGIN_PATH` (`;` separated on Windows, `:` elsewhere)
3. `$SMITH_HOME/plugins` (`SMITH_HOME` defaults to `~/.smith`)

`smith plugin list` shows the built-in plugins, the search path and what is
installed. An external plugin with the name of a built-in one **replaces** it.
smith checks that a plugin answers to the name it is declared under, and that
its version satisfies the constraint.

How plugins will be published and installed by name from a registry is not
decided yet ([#164](https://github.com/monosmith-app/monosmith/issues/164)).

## Testing: `smith plugin check`

```sh
smith plugin check <name | path to executable>
smith plugin check <name> --generate     # also scaffold an app, a lib and a package
smith plugin check <name> --json         # the report as JSON
```

`check` starts the plugin the way a workspace would and verifies that it names
itself, that its roles match what it implements, that its stacks are complete,
that its requirements and conflicts are well-formed, that every advertised
method answers a well-formed request (`discover`, `contribute`, `ci` and
`doctor` are called with no projects), and that it shuts down. With
`--generate`, a plugin that implements `generate` must write an app, a lib and a
package in a temporary directory, and a language plugin must then discover them
with the kind, the stack and the task commands it was given. It exits non-zero
when a check fails. A failed check prints why, under the check's name.

### Starting the plugin: name or path

`check` decides by looking at the **first argument**:

- If it is an **existing file**, that file is run as the plugin, and further
  arguments are passed to it. A directory, or a file the operating system cannot
  run, fails.
- Otherwise it is a **name**, looked up as the workspace would: `smith.toml`
  (`[plugins.external.<name>]`), then the built-in plugins, then the search path.

Two consequences, both easy to trip over:

- `smith plugin check python ./plugin.py` does **not** run your script. `python`
  is not a file, so it is looked up as a name and finds the *built-in* Python
  plugin. Real output: `python dev` and a list of passing checks for the wrong
  plugin.
- On Windows, a `.py` or `.mjs` is not an executable. `smith plugin check
  plugin.py` fails with `%1 is not a valid Win32 application`. Use a `.cmd`
  launcher (path form) or the name form.

So, to check an interpreted plugin:

| Where | Command |
|---|---|
| Linux, macOS | `chmod +x plugin.py` (the file starts with `#!/usr/bin/env python3`), then `smith plugin check ./plugin.py` |
| Windows, path form | `smith plugin check python\plugin.cmd`, a launcher next to the script. **Use backslashes**: `cmd.exe` cannot run `python/plugin.cmd` |
| Anywhere, name form | declare it in `smith.toml`, or install it on the search path, and run `smith plugin check widget` |

The launchers in the examples are two lines:

```bat
@echo off
python "%~dp0plugin.py" %*
```

Environment: smith sets `SMITH_PLUGIN=1`. The working directory is the one you
ran `smith` from. In Go you can run the same checks in a unit test with
`conformance.Check` over an in-process connection (`host.Connect`); see
`packages/plugin-sdk/host/plugin_client_test.go`.

## The example: a language smith does not know

The three plugins below do the same thing, so you can compare them. They teach
smith one made-up project type, **widget**: any directory with a `widget.json`

```json
{ "name": "hello", "version": "1.0.0", "kind": "lib" }
```

is a project, named by `name`, of the kind `kind` (default `lib`), and gets a
cached task `hello` that runs `git --version` (a command every machine that
runs smith has). They are in
[`docs/examples/plugins`](../examples/plugins) and are complete:

```
docs/examples/plugins/
  go/       main.go  go.mod  plugin.json
  python/   plugin.py  plugin.cmd  plugin.json
  node/     plugin.mjs  plugin.cmd  plugin.json
```

Each declares `"roles": ["language"]` and one capability, `discover`: the
smallest useful language plugin. Adding `generate`, `version/write` and
`publish` is what turns it into a full one.

### Go, with the SDK

The SDK (`packages/plugin-sdk`) does the framing, the version negotiation and
the capabilities for you. A plugin is a type with a `Describe` method, plus one
method for each capability it wants. **Capabilities are derived from the methods
the type has**, so a plugin cannot announce what it does not implement.
`docs/examples/plugins/go/main.go`:

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// widgetManifest is the content of a widget.json.
type widgetManifest struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Kind    string `json:"kind"` // app, lib or package
}

type widget struct{ root string }

// Describe answers "initialize". serve fills in the protocol version and the
// capabilities from the interfaces this type implements.
func (w *widget) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	w.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: "widget", Version: "0.1.0"},
		Kind:      protocol.KindLanguage,
		Roles:     []protocol.Role{protocol.RoleLanguage}, // not also scaffold: this plugin does not generate
		Languages: []string{"widget"},
		Manifests: []string{"widget.json"},
	}, nil
}

// Discover answers "discover". Having this method is what advertises the
// discover capability.
func (w *widget) Discover(p protocol.DiscoverParams) (protocol.DiscoverResult, error) {
	result := protocol.DiscoverResult{Projects: []protocol.Project{}}
	for _, manifest := range p.Manifests { // workspace-relative, forward slashes
		raw, err := os.ReadFile(filepath.Join(w.root, filepath.FromSlash(manifest)))
		if err != nil {
			return result, err
		}
		var m widgetManifest
		if err := json.Unmarshal(raw, &m); err != nil {
			return result, fmt.Errorf("%s: %w", manifest, err)
		}
		if m.Name == "" {
			continue // not ours
		}
		kind := protocol.ProjectKind(m.Kind)
		if kind == "" {
			kind = protocol.ProjectLib
		}
		result.Projects = append(result.Projects, protocol.Project{
			Name:     m.Name,
			Root:     path.Dir(manifest),
			Language: "widget",
			Kind:     kind,
			Version:  m.Version,
			Exports:  []string{m.Name},
			Tasks: map[string]protocol.Task{
				"hello": {
					Command:     []string{"git", "--version"},
					Inputs:      []string{"widget.json"},
					Cache:       true,
					Description: "say hello, as a widget does",
				},
			},
		})
	}
	return result, nil
}

func main() { serve.Main(&widget{}) }
```

Other methods are other interfaces in `serve` (`Generator`, `Versioner`,
`Publisher`, `Contributor`, `DoctorContributor`, `CIContributor`, `Remover`,
`RegistryReader`, `DependencyManager`, `DeclarationEditor`, `LatestReader`).
An error you return becomes a `-32000` response with your message.

Build and check it (here with the SDK as a neighbour; in your own module,
`go get github.com/monosmith-app/monosmith/packages/plugin-sdk` and drop the
`replace` line from `go.mod`):

```sh
cd docs/examples/plugins/go
GOWORK=off go build -o smith-plugin-widget .
smith plugin check ./smith-plugin-widget
```

(`GOWORK=off` only because this folder sits inside the MonoSmith repository,
whose `go.work` does not list it.) Real output, on Windows:

```
widget 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: language
  ✔ a language plugin discovers projects
  ✔ discover answers with no manifests
  ✔ shuts down when asked
```

`go build` appends `.exe` on Windows, so there `smith plugin check
.\smith-plugin-widget.exe`. The quickest start for a Go plugin of your own is
the scaffold, which writes a module that is a working structure plugin with a
conformance test and a README that says how to declare it:

```sh
$ smith add app my-plugin --stack go-smith-plugin
```

```text
Adding app my-plugin (stack go-smith-plugin) in apps/my-plugin
go: creating new go.mod: module github.com/acme/demo/apps/my-plugin
go: downloading github.com/monosmith-app/monosmith/packages/plugin-sdk v0.30.0
go: added github.com/monosmith-app/monosmith/packages/plugin-sdk v0.30.0
Wrote 8 file(s):
  .github/workflows/ci.yml
  apps/my-plugin/README.md
  apps/my-plugin/go.mod
  apps/my-plugin/go.sum
  apps/my-plugin/main.go
  apps/my-plugin/plugin/doc.go
  apps/my-plugin/plugin/myplugin_plugin_handler.go
  apps/my-plugin/plugin/myplugin_plugin_handler_test.go
Declare it in smith.toml to use it: see apps/my-plugin/README.md.
Try: smith test -p my-plugin
```

### Python, standard library only

`docs/examples/plugins/python/plugin.py` frames the messages itself in about
a hundred lines. The core is the framing functions and a table from method name
to function:

```python
def read_message(stream):
    """Return the next message as a dict, or None when smith closed the pipe."""
    length = None
    while True:
        line = stream.readline()
        if not line:
            return None
        line = line.decode("ascii").strip()
        if line == "":
            break
        name, _, value = line.partition(":")
        if name.lower() == "content-length":
            length = int(value.strip())
    if length is None:
        return None
    return json.loads(stream.read(length))


def write_message(stream, message):
    body = json.dumps(message).encode("utf-8")
    stream.write(b"Content-Length: %d\r\n\r\n" % len(body))
    stream.write(body)
    stream.flush()
```

Use the **binary** streams (`sys.stdin.buffer`) so Windows does not translate
newlines. The handlers and the loop:

```python
def initialize(params):
    global workspace_root
    workspace_root = params["workspaceRoot"]
    if PROTOCOL_VERSION not in params["protocolVersions"]:
        raise RuntimeError("smith speaks %s, this plugin speaks %d"
                           % (params["protocolVersions"], PROTOCOL_VERSION))
    return {
        "protocolVersion": PROTOCOL_VERSION,
        "plugin": {"name": "widget", "version": "0.1.0"},
        "kind": "language",
        "roles": ["language"],  # not also scaffold: this plugin does not generate
        "languages": ["widget"],
        "manifests": ["widget.json"],
        # Only what the plugin implements; smith never calls anything else.
        "capabilities": {"discover": True},
    }


def discover(params):
    projects = []
    for manifest in params["manifests"]:  # workspace-relative, forward slashes
        path = os.path.join(workspace_root, *manifest.split("/"))
        with open(path, encoding="utf-8") as f:
            data = json.load(f)
        if not data.get("name"):
            continue  # not ours
        projects.append({
            "name": data["name"],
            "root": os.path.dirname(manifest).replace("\\", "/"),
            "language": "widget",
            "kind": data.get("kind") or "lib",
            "version": data.get("version", ""),
            "exports": [data["name"]],
            "tasks": {
                "hello": {
                    "command": ["git", "--version"],
                    "inputs": ["widget.json"],
                    "cache": True,
                    "description": "say hello, as a widget does",
                }
            },
        })
    return {"projects": projects}


HANDLERS = {
    "initialize": initialize,
    "discover": discover,
    "shutdown": lambda params: {},
}


def main():
    stdin, stdout = sys.stdin.buffer, sys.stdout.buffer
    while True:
        request = read_message(stdin)
        if request is None:
            return
        if "id" not in request:
            continue  # a notification: no reply
        handler = HANDLERS.get(request["method"])
        response = {"jsonrpc": "2.0", "id": request["id"]}
        if handler is None:
            response["error"] = {"code": -32601, "message": "method not supported: " + request["method"]}
        else:
            try:
                response["result"] = handler(request.get("params") or {})
            except Exception as e:  # -32000: a failure the user must act on
                response["error"] = {"code": -32000, "message": str(e)}
        write_message(stdout, response)
        if request["method"] == "shutdown":
            return
```

(`import json, os, sys` and `PROTOCOL_VERSION = 1` at the top; the file in the
repository is the whole thing.) Check it. Real output, on Windows, using the
`.cmd` launcher:

```
> smith plugin check docs\examples\plugins\python\plugin.cmd
widget 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: language
  ✔ a language plugin discovers projects
  ✔ discover answers with no manifests
  ✔ shuts down when asked
```

On Linux and macOS: `chmod +x docs/examples/plugins/python/plugin.py`, then
`smith plugin check docs/examples/plugins/python/plugin.py`. (This path was
not run on Linux or macOS when this guide was written; it relies on the
operating system honouring the shebang.)

### Node.js, no dependencies

`docs/examples/plugins/node/plugin.mjs` is the same program. Messages arrive in
arbitrary chunks, so it buffers bytes and cuts them at the frame boundaries;
count **bytes** (`Buffer`), not string length:

```js
function send (message) {
  const body = Buffer.from(JSON.stringify(message), 'utf8')
  process.stdout.write(Buffer.concat([Buffer.from(`Content-Length: ${body.length}\r\n\r\n`, 'ascii'), body]))
}

// Cut the byte stream into framed messages.
let buffer = Buffer.alloc(0)
process.stdin.on('data', (chunk) => {
  buffer = Buffer.concat([buffer, chunk])
  for (;;) {
    const end = buffer.indexOf('\r\n\r\n')
    if (end < 0) return
    const match = /content-length:\s*(\d+)/i.exec(buffer.subarray(0, end).toString('ascii'))
    if (!match) throw new Error('message without Content-Length')
    const length = Number(match[1])
    if (buffer.length < end + 4 + length) return
    const body = buffer.subarray(end + 4, end + 4 + length).toString('utf8')
    buffer = buffer.subarray(end + 4 + length)
    handle(JSON.parse(body))
  }
})
```

`handle` dispatches to the `handlers` object (`initialize`, `discover`,
`shutdown`) exactly as the Python one does, answers errors with `-32601` or
`-32000`, and exits after `shutdown`. Real output, on Windows:

```
> smith plugin check docs\examples\plugins\node\plugin.cmd
widget 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: language
  ✔ a language plugin discovers projects
  ✔ discover answers with no manifests
  ✔ shuts down when asked
```

### Use it

In a workspace, declare the plugin (here the Python one, by its command) and add
a `widget.json`:

```toml
# smith.toml
[plugins.external.widget]
command = ["python", "docs/examples/plugins/python/plugin.py"]
```

```sh
$ smith projects
hello                lib      widget  libs/hello
$ smith run hello
hello:hello ▶ git --version
hello:hello │ git version 2.41.0.windows.1
hello:hello ✔ 25ms
✔ 1/1 tasks succeeded (0 from cache) in 31ms
$ smith run hello
hello:hello │ git version 2.41.0.windows.1
hello:hello ✔ from cache
✔ 1/1 tasks succeeded (1 from cache, 25ms saved) in 2ms
$ smith explain hello
hello:hello — say hello, as a widget does
  command    git --version
               from plugin widget
  cwd        libs/hello
               from plugin widget
  ...
```

The project is in the graph, its task is cached, and `smith explain` says every
value came from your plugin. To install instead of declaring: copy the
directory to `.smith/plugins/widget/` (or any directory of `SMITH_PLUGIN_PATH`),
and write `[plugins.external.widget]` with no command, or only a `version`.
With the Python plugin installed that way on Windows, with a `plugin.json`
whose command is `["./plugin.cmd"]`, `smith plugin list` shows (and
`smith plugin check widget` passes):

```
Installed (search path: ...\searchpath, C:\Users\russo\.smith\plugins):
  widget         0.1.0      ...\searchpath\widget
```

For a plugin installed this way, the `command` in `plugin.json` must start with
a launcher inside the directory: `./plugin.py` (executable, with a shebang) on
Linux and macOS, `./plugin.cmd` on Windows, `./smith-plugin-widget{exe}` for a
compiled one. The `plugin.json` files in the examples use these forms: the
Python and Node ones the Unix launcher, the Go one the compiled binary.

## A second plugin: adding tasks to projects

A plugin of kind `convention` adds to projects other plugins discovered. This
one adds a `rules` task to the widget projects and makes `lint` wait for it.
In Go (checked with `smith plugin check`, and run in a workspace with the widget
plugin above):

```go
type rules struct{}

func (rules) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: "rules", Version: "0.1.0"},
		Kind:      protocol.KindConvention,
		Roles:     []protocol.Role{protocol.RoleStructure},
		Languages: []string{"widget"}, // the languages it applies to
	}, nil
}

func (rules) Contribute(p protocol.ContributeParams) (protocol.ContributeResult, error) {
	out := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, project := range p.Projects {
		out.Projects[project.Name] = protocol.Contribution{
			Tasks: map[string]protocol.Task{
				"rules": {Command: []string{"git", "--version"}, Cache: true, Description: "check the house rules"},
			},
			DependsOn: map[string][]string{"lint": {"rules"}},
		}
	}
	return out, nil
}

func main() { serve.Main(rules{}) }
```

```toml
[plugins.external.rules]
command = ["tools/rules/rules"]

[conventions.rules]        # apply it to every project of the languages it names
```

```
rules 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: structure
  ✔ a structure plugin does something
  ✔ contribute answers with no projects
  ✔ shuts down when asked
```

```sh
$ smith run rules hello
hello:rules ▶ git --version
hello:hello ▶ git --version
...
✔ 2/2 tasks succeeded (0 from cache) in 31ms
```

Note `Languages`: a convention that names none applies to **no** project, and
nothing says so. A convention can instead be named in a project's
[stack](../reference/stacks.md) (as `structure`, or `toolchain` with `provides`
for the slots it owns).

## Other languages

Any language that can read and write bytes on stdin and stdout can be a plugin.
The recipe is the same as in the Python and Node examples:

1. Read the header lines up to the empty line; take `Content-Length`; read that
   many **bytes**; parse the JSON.
2. Dispatch on `method`. Answer `initialize` first, with `protocolVersion: 1`,
   your name and version, `kind`, `roles` and `capabilities`.
3. Write `{"jsonrpc":"2.0","id":<same id>,"result":...}` or `"error"`, framed,
   and flush.
4. On `shutdown`, answer `{}` and exit. On end of input, exit.
5. Send nothing to stdout except frames. Log to stderr.

Per language:

- **Rust, Ruby, Java, C#, Swift**: use the standard I/O in binary mode and
  `serde_json`, `JSON`, Jackson or `System.Text.Json`. Ship a binary (or a jar
  behind a launcher script) and point `plugin.json` at it with
  `{os}`, `{arch}` and `{exe}` for one directory per platform.
- **Shell**: possible for a plugin that only runs commands, using `jq`.
  `read` the header, `head -c "$length"` the body, build the answer with `jq -n`.
  Keep the byte count in mind: `${#body}` counts characters, not bytes, in
  UTF-8 locales.
- **Go with the SDK** is the least work. There is no published SDK for other
  languages; the Go types in `packages/plugin-sdk/protocol` are the schema.

Test with `smith plugin check`, then with a real workspace and
`smith explain <project>`, which says which plugin each task came from.

## Troubleshooting a plugin

| You see | Cause and remedy |
|---|---|
| `the plugin did not start and initialize: ... %1 is not a valid Win32 application` | Windows cannot run a script. Use a `.cmd` launcher or the name form |
| `plugin stopped answering: EOF` | the process exited before answering `initialize`: read the lines above it (stderr is shown). With a `.cmd`, a forward-slash path is the usual cause |
| `plugin "x" is not installed: no x/plugin.json in ...` | the name is not on the search path; the message lists the directories searched |
| `a scaffold generates projects` fails | your language plugin has no `roles`, so it is also a scaffold. Set `"roles": ["language"]` or implement `generate` |
| `a language plugin discovers projects` fails | it needs `capabilities.discover` and at least one `manifests` entry |
| every call fails after the first | something wrote to stdout that was not a frame |
| a method times out and the plugin dies | it took more than the timeout (two minutes by default); do slow work in tasks, not in `discover` |
| `smith.toml [plugins.external.x]: the plugin calls itself something else` | the name in `smith.toml` must be the name in `initialize` |
| a convention changes nothing | it names no `languages`, or is not enabled with `[conventions.<name>]` |

## See also

- [The plugin system](../design/plugin-system.md): lifecycle, failure modes, why
  out-of-process.
- [Plugins](../reference/plugins.md): the built-in plugins and roles.
- [Stacks](../reference/stacks.md) and [recipes](../reference/recipes.md): how
  plugins combine into a project.
- `packages/plugin-sdk`: the Go SDK and the protocol types.
