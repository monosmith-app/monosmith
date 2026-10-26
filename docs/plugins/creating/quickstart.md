# Writing a plugin: quickstart

A first plugin in about ten minutes, with the tools you already have: smith and Go. You scaffold a
working plugin, run it, change it, and see it act on a project. Nothing here needs a registry.

## 1. Choose what to write

| If smith should | Role | Kind | Implements | Guide |
|---|---|---|---|---|
| find projects of a new type | `language` | `language` | `discover` (+ `generate`, `version/write`, `publish`, `doctor`, ...) | [language](language-plugin.md) |
| provide a task such as `lint` or `test` for a language | `toolchain` | `convention` | `contribute` (+ `generate`) | [toolchain](toolchain-plugin.md) |
| add checks and tasks for team rules | `structure` | `convention` | `contribute`, `doctor` | [structure](structure-plugin.md) |
| write a new project's starter files | `scaffold` | `convention` | `generate` | [scaffold](scaffold-plugin.md) |
| release an app to a target | `deploy` | `convention` | `contribute`, `publish`, `ci` | [deploy](deploy-plugin.md) |

If you are unsure, start with a structure plugin: it needs no new language and shows the whole
loop.

## 2. Scaffold it

In a workspace with the Go language plugin enabled:

```sh
smith new demo && cd demo
smith add app my-plugin --stack go-smith-plugin
```

```text
Creating app my-plugin in apps/my-plugin: go-smith-plugin
go: creating new go.mod: module example.com/demo/apps/my-plugin
go: downloading github.com/monosmith-app/monosmith/packages/plugin-sdk v0.33.0
go: added github.com/monosmith-app/monosmith/packages/plugin-sdk v0.33.0
Wrote 10 file(s):
  .devcontainer/devcontainer.json
  .github/workflows/ci.yml
  apps/my-plugin/README.md
  apps/my-plugin/go.mod
  apps/my-plugin/go.sum
  apps/my-plugin/main.go
  apps/my-plugin/plugin/doc.go
  apps/my-plugin/plugin/myplugin_plugin_handler.go
  apps/my-plugin/plugin/myplugin_plugin_handler_test.go
  go.work
Declare it in smith.toml to use it: see apps/my-plugin/README.md.
Try: smith test -p my-plugin
```

The generated plugin is a structure plugin that gives Node, Go and Python projects a `hello` task.
Its whole protocol surface is two methods:

```go
// Describe implements serve.Plugin.
func (p *Plugin) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: Name, Version: p.version},
		Kind:      protocol.KindConvention,
		Roles:     []protocol.Role{protocol.RoleStructure},
		Languages: []string{"node", "go", "python"},
	}, nil
}

// Contribute implements serve.Contributor: a hello task for every project
// the plugin applies to.
func (p *Plugin) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, project := range params.Projects {
		res.Projects[project.Name] = protocol.Contribution{Tasks: map[string]protocol.Task{
			"hello": {Command: []string{"{smith}", "--version"}, Description: "says hello from my-plugin"},
		}}
	}
	return res, nil
}
```

Having a `Contribute` method is what advertises the `contribute` capability: the SDK derives the
capabilities from the methods a plugin type has, so it cannot advertise one it lacks.

## 3. Test it

```sh
smith test -p my-plugin
```

```text
my-plugin:test ▶ go test ./...
my-plugin:test │ ?   	example.com/demo/apps/my-plugin	[no test files]
my-plugin:test │ ok  	example.com/demo/apps/my-plugin/plugin	0.365s
my-plugin:test ✔ 1.148s
✔ 1/1 tasks succeeded (0 from cache) in 1.155s
```

The test in `plugin/` runs the conformance suite in process. To run the same suite against the
built program, which is what smith does:

```sh
cd apps/my-plugin && go build -o my-plugin.exe . && cd ../..
smith plugin check apps/my-plugin/my-plugin.exe
```

```text
my-plugin 0.0.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: structure
  ✔ a structure plugin does something
  ✔ declares a concurrency the protocol defines
  ✔ answers 3 requests sent together, one at a time
  ✔ contribute answers with no projects
  ✔ shuts down when asked
```

## 4. Use it

Declare the plugin, and apply it as a convention to every project of the languages it names. The
README the scaffold wrote has exactly this:

```toml
[plugins.external.my-plugin]
command = ["go", "run", "./apps/my-plugin"]

[conventions.my-plugin]
```

smith starts an external plugin only once you trust it, so the first command below asks (in a
terminal) whether to trust `my-plugin`; `smith plugin trust my-plugin` does it up front. See
[Plugin security](../using/plugin-security.md).

```sh
smith projects
smith run hello
smith explain my-plugin:hello
```

```text
my-plugin            app      go       apps/my-plugin
my-plugin:hello ▶ {smith} --version
my-plugin:hello │ smith version 1.62.0
my-plugin:hello ✔ 35ms
✔ 1/1 tasks succeeded (0 from cache) in 38ms
my-plugin:hello — says hello from my-plugin
  command    smith --version
  cwd        apps/my-plugin
  dependsOn  —
  inputs     **/* (every file of the project)
  outputs    —
  env        —
  cache      false
```

`smith explain` says where each field came from. A task command is an argv array, never a shell
line, and `{smith}` stands for the running smith.

## 5. Make it yours

Change the task to run your own tool, give it `inputs`, `outputs` and `cache: true` so smith caches
it, and decide which projects it applies to. Then read the guide for your kind of plugin, which
adds what the scaffold does not: checks for the workspace (`doctor`), CI needs, releases and new
project files.

| Next | Page |
|---|---|
| a worked example of your kind | [language](language-plugin.md), [toolchain](toolchain-plugin.md), [structure](structure-plugin.md), [scaffold](scaffold-plugin.md), [deploy](deploy-plugin.md) |
| every method and field | [the cheat sheet](protocol-cheat-sheet.md) |
| not writing Go | [The TypeScript SDK](typescript-sdk.md), [The Python SDK](python-sdk.md) or [The C# SDK](csharp-sdk.md) (the same plugin in TypeScript, Python or C#, on an SDK), or [Writing a plugin in Go, Node.js or Python](plugin-sdks.md) (Node.js and Python without the SDK) |
| share it | [Packaging and distributing a plugin](packaging-a-plugin.md) |

## Common mistakes

| You see | Cause |
|---|---|
| nothing: the convention adds no task | the plugin names no `languages`, or `[conventions.<name>]` is missing |
| `the plugin did not start and initialize: ... plugin stopped answering: malformed header ...` | a `print` to stdout. Log to stderr |
| `smith plugin check` passes but a task is missing in `smith explain` | `contribute` returned a task a project already has; use `Replace` to overwrite it |

More, with the exact messages, in [Troubleshooting plugins](../using/troubleshooting-plugins.md).
