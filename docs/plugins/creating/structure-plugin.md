# Creating a structure plugin

A structure plugin checks and enforces how a team lays out and writes code: no junk-drawer
folders, a README in every project, a naming rule, a license header. It adds tasks to the projects
it applies to, makes existing tasks (usually `lint`) wait for them, and reports health findings
with `smith doctor`. It needs no new language and no new tool, which makes it the best first plugin
to write. This guide builds `widget-layout`, in Go, TypeScript and Python.

## When to write one

Write a structure plugin to turn a team rule into a task CI runs. The built-in
[`vertical-slices`](../using/catalogue/vertical-slices.md) plugin is one: it adds `slice-check`,
runs before `lint`, and fails on a folder named `helpers`. If the rule is about *how a tool is run*
and projects choose between tools, write a [toolchain plugin](toolchain-plugin.md) instead.

## What smith asks of it

| Method | Required | Purpose |
|---|---|---|
| `initialize` | yes | `kind: "convention"`, `roles: ["structure"]`, and **`languages`** |
| `contribute` | to add tasks | tasks to add, tasks to replace, dependencies to add to existing tasks |
| `doctor` | optional | findings about the workspace's projects, each with a remedy and, if you can write it, a fix |
| `ci`, `generate` | optional | CI needs; files for a new project |

A structure plugin must implement at least one of `contribute`, `generate`, `ci` and `doctor`
(`a structure plugin does something`). A convention plugin that declares no roles *is* a structure
plugin by default.

### How a structure plugin is applied

| Way | Applies to | Configure |
|---|---|---|
| in a **stack** | the projects whose stack lists it under `structure` | `structure = ["widget-layout"]` in `[stacks.<name>]`, or a recipe's `use.structure` |
| as a **convention** | every project of the languages it names | `[conventions.widget-layout]` with optional `languages = [...]` (to narrow) and `exclude = [...]` (projects left alone) |

**`Languages` is what a convention applies to.** A plugin that names none applies to no project,
and nothing says so. Name the languages in the handshake.

## The example

`widget-layout` fails a project that contains a folder called `util`, `utils`, `helper`, `helpers`,
`common`, `shared` or `misc` (build output and `node_modules` are skipped), makes `lint` wait for
the check, and warns about a project without a `README.md`. The same plugin, three ways (a fourth, in
Node.js with the framing written by hand, is the smaller [`node`](../../examples/plugins/node)):

| Language | Directory | Files | How the protocol is spoken |
|---|---|---|---|
| Go | [`structure-layout-go`](../../examples/plugins/structure-layout-go) | `main.go`, `plugin.go`, `check.go`, `plugin_test.go` | the SDK (`serve.Main`) |
| TypeScript | [`structure-layout-node`](../../examples/plugins/structure-layout-node) | `plugin.ts`, `layout.ts`, `layout.test.ts`, `plugin.cmd`, `plugin.json` | the [TypeScript SDK](typescript-sdk.md) (`serve`) |
| Python | [`structure-layout-python`](../../examples/plugins/structure-layout-python) | `plugin.py`, `test_plugin.py`, `plugin.cmd`, `plugin.json` | by hand, standard library only |

### Step 1: the handshake

```go
func (l *layout) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	l.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: "widget-layout", Version: "0.1.0"},
		Kind:   protocol.KindConvention,
		Roles:  []protocol.Role{protocol.RoleStructure},
		// The languages it applies to. A convention that names none applies to
		// no project at all.
		Languages: []string{"widget"},
	}, nil
}
```

The TypeScript version returns the same object and the SDK adds the version and the capabilities; the
Python version, which has no SDK, returns the JSON with them spelled out:

```json
{
  "protocolVersion": 1,
  "plugin": { "name": "widget-layout", "version": "0.1.0" },
  "kind": "convention",
  "roles": ["structure"],
  "languages": ["widget"],
  "capabilities": { "contribute": true, "doctor": true }
}
```

### Step 2: add the task and order it before `lint`

```go
out.Projects[project.Name] = protocol.Contribution{
	Tasks: map[string]protocol.Task{
		"layout-check": {
			Command:     []string{exe, "layout-check"},
			Cache:       true,
			Description: "fail on folders named util, helpers, common, shared or misc",
		},
	},
	// DependsOn only touches a task the project has once every earlier plugin
	// has contributed; for a project without lint it is ignored.
	DependsOn: map[string][]string{"lint": {"layout-check"}},
}
```

| Choice | Why |
|---|---|
| no `Inputs` | every file of the project is an input, so a new or renamed folder is a cache miss. List `Inputs` when the rule reads only some files, to keep hits |
| `Cache: true` | the result depends on the inputs alone |
| `DependsOn` emitted for every project | `Contribute` is given the projects as the *language plugin* discovered them, before toolchain plugins add `lint`. Deciding from `project.Tasks["lint"]` would miss a lint that another plugin adds. smith ignores a dependency on a task the project lacks |
| `Tasks` rather than `Replace` | a task of that name the project already has is kept |

The check itself is ordinary code that runs when the task runs: the task command is the plugin's own
executable, which serves the protocol when smith sets `SMITH_PLUGIN=1` and is a checker when it is
run with `layout-check` and the variable is not set. Real output:

```text
$ smith explain core
core has 3 tasks:
  build         ...  (cached, after ^build)
  layout-check  ...  (cached)
  test          ...  (cached)
...
core:layout-check — fail on folders named util, helpers, common, shared or misc
  command    <examples>\structure-layout-go\smith-plugin-widget-layout.exe layout-check   (from convention widget-layout)
  cwd        libs/core   (from convention widget-layout)
  dependsOn  —   (from convention widget-layout)
  inputs     **/* (every file of the project)   (from convention widget-layout)
  outputs    —   (from convention widget-layout)
  env        —   (from convention widget-layout)
  cache      true   (from convention widget-layout)

$ mkdir libs/core/utils
$ smith run layout-check
core:layout-check ▶ <examples>\structure-layout-go\smith-plugin-widget-layout.exe layout-check
core:layout-check │ folders named for nothing: utils (name them for what they do)
core:layout-check ✖ exited with code 1
✖ 0/1 tasks succeeded (0 from cache) in 14ms
$ rmdir libs/core/utils
$ smith run layout-check
core:layout-check │ no junk-drawer folders
core:layout-check ✔ 14ms
```

In a stack the same task makes the project's linter wait. With the `widget-lint` toolchain in the
same stack, a violation fails the layout check and skips the lint:

```text
$ smith run lint -p strict
strict:layout-check │ folders named for nothing: utils (name them for what they do)
strict:layout-check ✖ exited with code 1
  ✖ strict:layout-check exited with code 1
  ↷ strict:lint because strict:layout-check did not succeed
✖ 0/2 tasks succeeded (0 from cache) in 16ms
$ smith explain strict:lint
  dependsOn  layout-check (this project's layout-check, first)   (from plugin widget-layout (stack widget-strict))
```

### Step 3: health findings (`doctor`)

```go
func (l *layout) Doctor(p protocol.DoctorParams) (protocol.DoctorResult, error) {
	result := protocol.DoctorResult{Findings: []protocol.Finding{}}
	for _, project := range p.Projects {
		readme := filepath.Join(l.root, filepath.FromSlash(project.Root), "README.md")
		_, err := os.Stat(readme)
		f := protocol.Finding{Check: project.Name + " has a README.md", OK: err == nil}
		if err != nil {
			f.Detail = project.Root + "/README.md does not exist"
			f.Remedy = "write " + project.Root + "/README.md: what the project is for"
			f.Warning = true // reported, but does not fail smith doctor
		}
		result.Findings = append(result.Findings, f)
	}
	return result, nil
}
```

Every failed finding names its remedy. `Warning: true` reports without failing `smith doctor`.
`doctor` is asked about the projects the plugin applies to, so a project excluded from the
convention is not checked:

```text
$ smith doctor
...
✔ core has a README.md
```

(`legacy`, listed in `exclude = ["legacy"]`, has no README and is not reported.)

### Step 4: TypeScript and Python

The protocol is the same. The TypeScript version is on the SDK, which frames the messages and derives
the capabilities from the two methods the object has (`layout.ts` is the plugin, `plugin.ts` the entry
point that is both the plugin and the checker). The core of it:

```ts
export function createLayoutPlugin (): SmithPlugin {
  return {
    describe: (params) => { /* the object above */ },
    contribute: ({ projects }) => contributeTo(projects, () => ({
      tasks: { 'layout-check': task(['node', SELF, 'layout-check'], { cache: true, description: '...' }) },
      dependsOn: { lint: ['layout-check'] },
    })),
    doctor: ({ projects }) => ({ findings: /* one per project */ }),
  }
}

if (process.env.SMITH_PLUGIN === '1') await serve(createLayoutPlugin())
else if (process.argv[2] === 'layout-check') { /* run the check */ }
```

[The TypeScript SDK](typescript-sdk.md) walks through it. The Python version has no SDK: the framing is
yours, and the file is one program that is both the plugin and the checker. Its core:

```python
def contribute(params):
    command = [sys.executable, os.path.abspath(__file__), "layout-check"]
    ...

def main():
    if os.environ.get("SMITH_PLUGIN") == "1":
        serve()
    elif sys.argv[1:] == ["layout-check"]:
        ...
```

[Writing a plugin in Go, Node.js or Python](plugin-sdks.md) explains the framing loop a plugin without
an SDK writes.

## Run and check

```text
$ smith plugin check ./smith-plugin-widget-layout
widget-layout 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: structure
  ✔ a structure plugin does something
  ✔ contribute answers with no projects
  ✔ doctor answers with no projects
  ✔ shuts down when asked
```

The TypeScript and Python versions give the same seven lines (`smith plugin check .\plugin.cmd` on
Windows; see [Testing a plugin](testing-a-plugin.md#checking-an-interpreted-plugin)). Each ships its
own tests: `go test ./...`, `node --test` and `python -m unittest`, all passing. For the same
requests the TypeScript plugin and the Go plugin answer with the same bytes.

## Use it

As a convention, for every widget project:

```toml
[plugins.external.widget-layout]
command = ["/path/to/smith-plugin-widget-layout"]

[conventions.widget-layout]
exclude = ["legacy"]
```

Or in a stack, for the projects that choose it:

```toml
[stacks.widget-strict]
language = "widget"
toolchain = ["widget-lint"]
structure = ["widget-layout"]
```

## Common mistakes

| You see | Cause and remedy |
|---|---|
| the plugin is configured and no task appears | `Languages` is empty, or it does not include the project's language |
| `✖ a structure plugin does something` | the plugin implements none of `contribute`, `generate`, `ci`, `doctor` |
| `smith.toml [conventions.x]: x is not a convention plugin` (from the code) | `Kind` is `language`; a structure plugin is kind `convention` |
| `smith.toml [conventions.x]: it does not apply to "go"; it supports [widget]` (from the code) | `languages = [...]` in the convention names a language the plugin does not list |
| `lint` does not wait for the check | `DependsOn` named a task that does not exist yet; emit it unconditionally and let smith ignore it |
| `plugin.cmd` fails to start on Windows | a bare name is looked up on `PATH`; write `.\plugin.cmd` |
| a plugin that frames the messages itself prints a banner and smith stops answering | stdout is the protocol; use `console.error`. (The TypeScript SDK sends whatever else is written to stdout to stderr, so a stray `console.log` cannot corrupt it) |
| the check passes locally and fails in CI | the task command contains the plugin's absolute path (as this example's does); in a shared plugin put the tool on `PATH` and name it in the command |

## See also

- [Creating a toolchain plugin](toolchain-plugin.md), [Creating a deploy plugin](deploy-plugin.md)
- [The vertical-slices plugin](../using/catalogue/vertical-slices.md)
- [Conventions](../../reference/conventions.md)
- [The plugin protocol cheat sheet](protocol-cheat-sheet.md)
