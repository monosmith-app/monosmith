# Sharing facts between plugins

A fact is a value that several tools must agree on: an import alias, the source folder, the
Node version. Without facts, each tool's config repeats the value and a change means editing
every file. With facts, one plugin publishes a default, a person may override it once in
`smith.toml`, and every plugin that cares reads the same merged value. Plugins never call or
read each other: they only publish values and read the merged result.

Facts are part of protocol version 1. A plugin that ignores them keeps working.

## What a fact is

A fact is a JSON value under a key of the form `<namespace>.<name>`. These keys are well-known,
and smith checks the type of their value:

| Key | Type | Meaning |
|---|---|---|
| `module.aliases` | object of strings | import alias to project-relative directory, such as `{"@app": "src"}` |
| `source.roots` | array of strings | project-relative directories that hold source code |
| `test.patterns` | array of strings | globs that select test files |
| `build.outputs` | array of strings | directories a build writes |
| `runtime.node` | string | the Node.js version range the project runs on |

Any other key with a namespace is allowed and carries any JSON value. Use your plugin's
own namespace for it (`acme.region`).

## Who sets a fact, and which one wins

For each key the narrowest scope that sets it wins:

1. `[projects.<name>.facts]` in `smith.toml`
2. `[facts]` in `smith.toml`
3. a recipe (reserved; no recipe carries facts yet)
4. a default that a plugin publishes

Two plugins that publish different defaults for one key are an error. It names the project,
the key, both plugins and both values. A person decides it in `smith.toml`.

```toml
[facts]
"module.aliases" = { "@app" = "src" }

[projects.web.facts]
"source.roots" = ["src", "generated"]
```

`smith explain web --facts` prints each value with its scope and source;
`smith explain web --fact module.aliases` prints one.

## Publish a fact

Return `facts` in the `Contribution` of a project. Publish the value your tool uses by default,
so that nothing changes for a person who sets nothing.

```go
func (p *Plugin) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, project := range params.Projects {
		res.Projects[project.Name] = protocol.Contribution{
			Facts: map[string]any{protocol.FactSourceRoots: []string{"src"}},
		}
	}
	return res, nil
}
```

In TypeScript, Python and C# the field is `facts` (`Facts`) on the `Contribution` type. The types
come from the protocol schema, so they are in every SDK.

## Read a fact

Read it from a request you already get. Do not read it from a file or another plugin.

- `project.facts` is in every request that carries a project (`contribute`, `ci`, `doctor`).
  It holds the `smith.toml` facts and the defaults of the plugins asked before.
- `params.facts` is in `generate`. It holds the `smith.toml` facts only, because the project
  does not exist yet.

A key that is absent means no one set it. Use your own default then:

```go
aliases := map[string]string{}
if raw, ok := params.Facts[protocol.FactModuleAliases].(map[string]any); ok {
	for alias, dir := range raw {
		aliases[alias], _ = dir.(string)
	}
}
```

Order matters in one case. The `contribute` request of a plugin does not carry the defaults
published by the plugin itself or by the other plugins of the same stack in the same round.
If two such plugins must agree on a value, ask the person to set it in `smith.toml`.

## Check your plugin

`smith plugin check <plugin>` sends your plugin a `contribute` request for a probe project that
carries facts, and fails when you publish an invalid fact (a key without a namespace, or a
well-known key of the wrong type). The SDK test suites run the same check.

## A real use

The Node toolchains read `module.aliases` when they scaffold a project. One entry in
`smith.toml` becomes the TypeScript `paths`, the Jest `moduleNameMapper`, the Vite `resolve.alias`
and the Vitest alias:

```toml
[facts]
"module.aliases" = { "@app" = "src" }
```

Each alias maps to a directory relative to each project. The alias `@app` then resolves
`@app/greeting` to `src/greeting` in all four tools.
