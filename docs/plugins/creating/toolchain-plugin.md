# Creating a toolchain plugin

A toolchain plugin decides how projects of a language build, test, lint or type-check: it fills a
**slot**, a task name such as `lint` or `test`. It does not find projects (the language plugin
does); it adds a task to the projects whose stack names it, and it can write the files and scripts
the tool needs when a project is created. This guide builds `widget-lint`, which gives widget
projects a `lint` task, takes its rules from a setting, and publishes a stack.

## When to write one

Write a toolchain plugin when a language smith knows should run a tool smith does not: another
linter, a different test runner, a code generator, an extra type checker. The built-in
`eslint`, `jest`, `vitest`, `typescript`, `vite`, `angular`, `astro`, `playwright` and `vscode`
plugins are toolchain plugins; see [the catalogue](../using/catalogue/README.md).

If the tool should apply to every project of a language with no choice and no slot, a
[structure plugin](structure-plugin.md) applied as a convention is simpler. A toolchain is the right
shape when projects **choose** among alternatives (jest *or* vitest) and smith must refuse a stack
that holds both.

## What smith asks of it

| Method | Required | Purpose |
|---|---|---|
| `initialize` | yes | `kind: "convention"`, `roles` including `toolchain`, `languages`, **`provides`**, `requires` and optionally `conflicts` and `stacks` |
| `contribute` | to add or change tasks | tasks to add, tasks to replace, dependencies to add to existing tasks, by project |
| `generate` | to set up new projects | files, scripts and dependencies a new project of the stack needs |
| `ci`, `doctor` | optional | CI needs and health checks for the projects the plugin serves |

A toolchain **must** list the slots it fills in `provides`, and only a toolchain may. Two plugins of
one project may not provide the same slot. The conformance suite checks both
(`a toolchain names the slots it fills`, `only a toolchain provides tasks`) and that the plugin
implements at least one of `contribute`, `generate`, `ci` and `doctor`
(`a toolchain plugin does something`).

### Two ways to provide a task

| Way | Use it when | Example |
|---|---|---|
| **`contribute`**: return the task | the language plugin does not build tasks from scripts, or you want the task defined in one place | `widget-lint` below; `playwright` |
| **`generate`**: write a script and files | the language plugin turns scripts into tasks (the Node plugin does: every `package.json` script is a task) | `eslint`, `jest`, `vitest` write a `lint` or `test` script |

The built-in Node toolchains `eslint`, `jest` and `vitest` implement only `generate`; the node
plugin creates the `lint` and `test` tasks from the scripts they write. A task in `contribute` is
merged into the project like this, in the stack's order:

| Field of `Contribution` | Effect |
|---|---|
| `Tasks` | added; **a task the project already has is kept** |
| `Replace` | replaces the project's task of the same name, or adds it when there is none |
| `DependsOn` | adds dependencies to a task the project has once earlier plugins have contributed (ignored otherwise) |

Toolchain plugins apply in the order the stack lists them, so a later one may replace a task an
earlier one added.

## The example

`widget-lint` fills the `lint` slot for [widget](language-plugin.md) projects. Its lint command
fails when `widget.json` leaves a field empty; which fields are required is a **setting**. Source:
[`docs/examples/plugins/toolchain-widget-lint`](../../examples/plugins/toolchain-widget-lint).

```text
toolchain-widget-lint/
  go.mod
  main.go           serve the protocol for smith; run the linter for the task
  plugin.go         the plugin: Describe, Contribute, Generate
  linter.go         the lint command line
  plugin_test.go
  plugin.json
```

### Step 1: declare the slot, the requirement and a stack

```go
func (l *lint) Describe(p protocol.InitializeParams) (protocol.InitializeResult, error) {
	l.root = p.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: pluginName, Version: "0.1.0"},
		Kind:      protocol.KindConvention,
		Roles:     []protocol.Role{protocol.RoleToolchain},
		Languages: []string{"widget"},                                        // the projects it applies to
		Provides:  []string{"lint"},                                          // the slot it fills
		Requires:  []protocol.PluginRef{protocol.RequiresLanguage("widget")}, // no use without the language
		Stacks: []protocol.Stack{{
			Name:        "widget-linted",
			Description: "widget projects that must describe and license themselves",
			Language:    "widget",
			Toolchain:   []string{pluginName},
		}},
	}, nil
}
```

| Field | Rule |
|---|---|
| `Languages` | the language plugins it applies to, by name or `--lang` value. smith refuses it in a stack of another language: `widget-lint applies to [widget], not to node projects` |
| `Requires` | a reference sets `Level`, `Name` and/or `Slot`, at least one. `RequiresLanguage("widget")` is `{level: language, name: widget}`. `Conflicts` is the same for plugins that must not share a project |
| `Stacks` | ready-made stacks; each needs a unique `Name` and a `Language` |

The published stack shows up in `smith stacks`:

```text
widget-linted  (plugin widget-lint)
  widget projects that must describe and license themselves
  language   widget
  toolchain  widget-lint
```

### Step 2: contribute the task, driven by settings

```go
func requiredFields(p protocol.Project) []string {
	raw, ok := p.Settings[pluginName]["require"].([]any)
	if !ok {
		return defaultRequire
	}
	// ... keep the non-empty strings ...
}

func (l *lint) Contribute(p protocol.ContributeParams) (protocol.ContributeResult, error) {
	exe, _ := os.Executable()
	out := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, project := range p.Projects {
		out.Projects[project.Name] = protocol.Contribution{
			Tasks: map[string]protocol.Task{
				"lint": {
					Command:     []string{exe, "lint", strings.Join(requiredFields(project), ",")},
					Inputs:      []string{"widget.json"},
					Cache:       true,
					Description: "check that widget.json sets " + strings.Join(requiredFields(project), ", "),
				},
			},
		}
	}
	return out, nil
}
```

`p.Projects` are the projects whose stack names the plugin (or the ones a `[conventions.<name>]`
applies it to), each with the settings smith.toml holds for the plugin in
`Settings["widget-lint"]`. The settings are the table

```toml
[projects.extra.settings.widget-lint]
require = ["description", "license", "homepage"]
```

and they reach the plugin as a map; numbers arrive as `float64`, lists as `[]any`.

### Step 3: files for a new project (`generate`)

```go
func (l *lint) Generate(p protocol.GenerateParams) (protocol.GenerateResult, error) {
	if p.Convention {
		return protocol.GenerateResult{Files: []string{}}, nil // a convention adds only what every project needs
	}
	// write libs/<name>/WIDGET-LINT.md only when it is absent ...
}
```

`generate` runs after the language plugin (or the scaffold) wrote the project. A toolchain adds,
never replaces: write a file, script or dependency only when it is absent. `p.Convention` is true
when smith asks the plugin as a workspace convention rather than as one of the stack's plugins; add
only what every matching project needs then, and leave suites and starter files to the stack that
names the plugin.

## Run and check it

```sh
cd docs/examples/plugins/toolchain-widget-lint
GOWORK=off go build -o smith-plugin-widget-lint .
smith plugin check ./smith-plugin-widget-lint
```

```text
widget-lint 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: toolchain
  ✔ a toolchain plugin does something
  ✔ only a toolchain provides tasks
  ✔ a toolchain names the slots it fills
  ✔ requires language plugin widget: a known reference
  ✔ stack widget-linted is complete
  ✔ contribute answers with no projects
  ✔ shuts down when asked
```

## Use it in a workspace

A workspace that already has the `widget` language plugin declares this one and creates a project
with its stack:

```toml
[plugins.external.widget-lint]
command = ["/path/to/smith-plugin-widget-lint"]
```

```text
$ smith add lib extra --stack widget-linted
Adding lib extra (stack widget-linted) in libs/extra
Adding widget-lint
Wrote 3 file(s):
  libs/extra/README.md
  libs/extra/WIDGET-LINT.md
  libs/extra/widget.json
$ smith run lint
extra:lint ▶ .../smith-plugin-widget-lint.exe lint description,license
extra:lint │ widget.json sets description,license
extra:lint ✔ 15ms
✔ 1/1 tasks succeeded (0 from cache) in 20ms
$ smith explain extra:lint
extra:lint — check that widget.json sets description, license
  command    .../smith-plugin-widget-lint.exe lint description,license
  cwd        libs/extra
  ...
  cache      true
```

`Adding widget-lint` is the plugin's `generate`. With the setting above, the same task fails and says
why:

```text
extra:lint ▶ .../smith-plugin-widget-lint.exe lint description,license,homepage
extra:lint │ widget.json must set homepage
extra:lint ✖ exited with code 1
✖ 0/1 tasks succeeded (0 from cache) in 26ms
```

Before the plugin, `smith doctor` said `core has no lint task, so CI never checks it`; a slot filled
is a task CI verifies.

## Test it

`plugin_test.go` runs the conformance suite, checks that the slot is `lint`, that settings change the
command (`description,license` by default, `description,license,version` with a `require` list), that
`generate` writes a file once and nothing the second time, and that the lint command names the empty
fields.

```text
$ GOWORK=off go test ./...
ok  	example.com/smith-plugin-widget-lint	0.369s
```

## Common mistakes

| You see | Cause and remedy |
|---|---|
| `✖ a toolchain names the slots it fills` — `a toolchain plugin must list the tasks it provides` (from the code) | `Provides` is empty. List every task the plugin owns |
| `✖ only a toolchain provides tasks` | `Provides` is set without the toolchain role |
| `jest and vitest both provide the test task; keep one` | two plugins in one stack fill the same slot. Keep one, or give them different slots |
| `widget-lint applies to [widget], not to node projects` | `Languages` does not include the stack's language |
| `plugin "widget-lint" is not loaded` | the stack names a plugin that is not declared under `[plugins.external]` |
| `eslint is listed as structure, but it plays [toolchain]` | the stack lists the plugin at a level it does not declare in `Roles` |
| the task is missing | the project already had a `lint`: `Tasks` keeps it. Use `Replace` to take the slot over |
| `dependsOn` has no effect | it only touches a task the project has once earlier plugins contributed; contribute the task in the same plugin, or put the plugin after the one that adds it |
| the task re-runs every time | the command contains something that changes (a path, a time), or `Inputs` is missing and the whole project is an input |
| `smith.toml: ... unknown field` for `[conventions.widget-lint] require = ...` | `[conventions.<name>]` takes only `languages`, `exclude` and the docker keys. Use `[projects.<name>.settings.<plugin>]` |

## See also

- [Creating a structure plugin](structure-plugin.md), [Creating a language plugin](language-plugin.md)
- [Creating a recipe and a stack](recipe-and-stack.md): choosing `widget-lint` at the `lint` slot
- The built-in toolchains: [eslint](../using/catalogue/eslint.md), [vitest](../using/catalogue/vitest.md), [playwright](../using/catalogue/playwright.md)
- [The plugin protocol cheat sheet](protocol-cheat-sheet.md), [The TypeScript SDK](typescript-sdk.md) (a toolchain plugin in TypeScript implements `contribute` and, to scaffold, `generate`)
