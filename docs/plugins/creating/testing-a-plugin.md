# Testing a plugin

Three layers, from cheap to real: `smith plugin check`, unit tests that talk to the plugin in
process, and a workspace test that runs it under smith. The examples in
[`docs/examples/plugins`](../../examples/plugins/README.md) use all three.

## What `smith plugin check` verifies

```sh
smith plugin check <name | path to executable> [args...]
smith plugin check <name> --generate     # also scaffold an app, a lib and a package
smith plugin check <name> --json         # the report as JSON
smith plugin check <name> --env NPM_TOKEN --env 'AWS_*'   # variables the plugin may see
```

It starts the plugin the way a workspace would (with `SMITH_PLUGIN=1`), runs the checks below, asks
it to shut down, and exits non-zero (`smith: the plugin does not conform`) if any failed. A failed
check prints why under its name. The checks live in `packages/plugin-sdk/conformance` and are the
same ones `conformance.Check` runs in your tests.

The plugin is yours, so it needs no trust to be checked. It starts with the [minimal
environment](../using/plugin-security.md#the-environment-an-external-plugin-gets) every external plugin
gets: if it reads a variable (a token, a registry URL), pass `--env NAME` (repeatable; `--env 'AWS_*'`
for a prefix) for the run, and declare the same names in the `env` of your `plugin.json`
([Declaring the environment a plugin needs](packaging-a-plugin.md#declaring-the-environment-a-plugin-needs)).
A plugin found by name also gets the variables its `plugin.json` and `smith.toml` declare.

| Check | When | Passes when |
|---|---|---|
| `names itself` | always | `initialize` returned a name and a version |
| `has a known kind` | always | `kind` is `language` or `convention` |
| `plays a known role: <role>` | for each effective role | the role is one of `language`, `toolchain`, `scaffold`, `structure`, `deploy` |
| `a language plugin discovers projects` | the `language` role | `discover` is advertised and `manifests` is not empty |
| `a scaffold generates projects` | the `scaffold` role | `generate` is advertised |
| `a toolchain plugin does something`, `a structure plugin ...`, `a deploy plugin ...` | that role, without the `language` role | at least one of `contribute`, `generate`, `ci`, `doctor` is advertised |
| `only a toolchain provides tasks` | `provides` is set | the `toolchain` role is declared |
| `a toolchain names the slots it fills` | a convention with the `toolchain` role | `provides` is not empty |
| `requires <ref>: a known reference` | each `requires` | the reference sets something and its `level`, if any, is a known role |
| `conflicts with <ref>: a known reference` | each `conflicts` | the same, and it does not name the plugin itself |
| `stack <name> is complete` | each published stack | it has a unique name and a language |
| `discover answers with no manifests` | `discover` advertised | `discover` with an empty list returns |
| `contribute answers with no projects` | `contribute` advertised | the same for `contribute` |
| `ci answers with no projects` | `ci` advertised | the same for `ci` |
| `doctor answers with no projects` | `doctor` advertised | the same for `doctor` |
| `shuts down when asked` | always | `shutdown` is answered and the process exits within five seconds |

The roles are the *effective* ones: a plugin that declares none gets its kind's default (a language
plugin: `language`, `toolchain`, `scaffold`).

With `--generate`, a plugin that advertises `generate` also has to work:

| Check | Passes when |
|---|---|
| `generates an app`, `generates a lib`, `generates a package` | `generate` returns files and every one exists under the temporary workspace. Each of the three kinds must succeed |
| `discovers the generated <kind>` | (language plugins) `discover` over the files found returns a project at that directory |
| `the generated <kind> keeps its kind` | the discovered kind is the requested one |
| `the generated <kind> records its stack` | the discovered `stack` is the one `generate` was asked to record (`conformance`) |
| `the generated <kind>'s tasks have commands` | no task has an empty `command` |
| `writes the <kind>'s version (dry run)` | `version/write` in dry run answers, if advertised |
| `would remove the <kind> (dry run)` | `remove` in dry run answers, if advertised |
| `publishes the package (dry run)` | `publish` in dry run answers for the package, if advertised |
| `says what the registry has of the package` | `published` answers (an unreachable registry is a result, not an error) |

`--generate` may install dependencies (the built-in plugins run their package managers), so it is
opt-in. A scaffold that supports only one kind of project fails the other two; check such a plugin
without `--generate`.

Real output for a structure plugin:

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

and a failure, from the scaffold example run with `--generate`:

```text
  ✔ generates an app
  ✖ generates a lib
      a widget service is an app: smith add app clib --stack widget-service
smith: the plugin does not conform
```

### Name or path

`check` decides by looking at the first argument:

- if **something exists at that path** (a file, or even a directory), it is run as the plugin, with
  the remaining arguments passed to it;
- otherwise it is a **name**, looked up as the workspace would: `[plugins.external.<name>]` in
  `smith.toml`, then the built-in plugins, then the plugin search path.

So `smith plugin check python plugin.py` runs your script only if a file or folder called `python`
exists in the current directory; otherwise `python` is looked up as a name and finds the *built-in*
Python plugin, and you are checking the wrong plugin. Use the path form with a launcher, or the name
form.

### Checking an interpreted plugin

| Where | Command |
|---|---|
| Linux, macOS | `chmod +x plugin.py` (the file starts with `#!/usr/bin/env python3`), then `smith plugin check ./plugin.py`. Not run on Linux or macOS for this page |
| Windows, path form | `smith plugin check .\plugin.cmd`, a launcher next to the script. Write the leading `.\` and use backslashes (Git Bash drops them: use PowerShell or `cmd`) |
| anywhere, name form | declare it in `smith.toml`, or install it on the search path, and run `smith plugin check widget` |

On Windows a `.py`, `.mjs` or `.ts` is not an executable: `smith plugin check .\plugin.py` fails with
`fork/exec .\plugin.py: %1 is not a valid Win32 application.` The launchers in the examples are
two lines:

```bat
@echo off
node "%~dp0plugin.ts" %*
```

Without a launcher, the path form also takes the absolute path of the interpreter followed by the
script (`smith plugin check C:\...\node.exe C:\...\plugin.ts`): the first argument exists, so it is run
as the plugin with the rest as its arguments. The TypeScript SDK's own integration tests check the
TypeScript examples that way.

## Unit tests that talk to the plugin

In Go, serve the plugin in your test process over pipes, with the real handshake, and call it as
smith would. The pattern is the same in every Go example:

```go
func connect(t *testing.T, workspace string) *host.Client {
	t.Helper()
	hostR, pluginW := io.Pipe()
	pluginR, hostW := io.Pipe()
	go func() { _ = serve.Serve(&layout{}, pluginR, pluginW); _ = pluginW.Close() }()
	c, err := host.Connect(hostR, hostW, host.Options{WorkspaceRoot: workspace})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestConforms(t *testing.T) {
	report := conformance.Check(connect(t, t.TempDir()), conformance.Options{})
	if !report.OK() {
		t.Fatal(report.Failures())
	}
}
```

Then test behaviour through the same client: `c.Call(protocol.MethodContribute, params, &result, 0)`.
Give the test a temporary workspace (`t.TempDir()`), write the manifests the plugin reads, and assert
on the typed result. What is worth asserting:

| Assert | Because |
|---|---|
| the capabilities and `provides` of `c.Info` | smith relies on them |
| the graph: `Exports` of one project equal the `Imports` of another | the edge is made by smith from these |
| task `Outputs` are under `dist/` and `Command` is non-empty | `smith doctor` and the conformance suite check them |
| a dry run changes no file | `smith release --dry-run` calls you |
| a second `publish` is skipped with a reason | a failed release is re-run |
| `generate` never overwrites | a toolchain adds, only where absent |
| an error message names the remedy | the user sees it |

With `conformance.Options{Workspace: ws}` the suite also runs the `--generate` checks in process, as
`language-widget`'s `TestConforms` does.

In TypeScript, the SDK has the same two tools: `connectInProcess(plugin, { workspaceRoot })` serves a
plugin object over pipes with the real handshake and returns a client (`client.call('contribute',
params)` is typed by method), and `checkPlugin(client)` runs the conformance checks. `startPlugin({
command, workspaceRoot })` starts the real process instead. See [The TypeScript SDK](typescript-sdk.md#3-test-it),
and `structure-layout-node/layout.test.ts` (`node --test`).

In Python, or in Node.js without the SDK, start the plugin as a child process with `SMITH_PLUGIN=1` and
speak the protocol to it, as `structure-layout-python/test_plugin.py` (`python -m unittest`) does: write
the framed requests, close stdin, parse the framed answers.

```text
$ node --test
✔ the plugin conforms (5.9366ms)
✔ the handshake advertises exactly what is implemented (1.8395ms)
✔ lint waits for the layout check (1.5453ms)
✔ an unknown method is -32601 (1.5749ms)
✔ doctor warns about a missing README (2.0646ms)
✔ findJunkDrawers skips build output (2.474ms)
ℹ tests 6
ℹ pass 6
ℹ fail 0
$ python -m unittest
.....
Ran 5 tests in 0.246s
OK
```

If a Go and a TypeScript plugin implement the same logic, the strongest test is that they give the same
bytes: start both with the same requests and compare the answers. The SDK's
`integration/cross-language.integration.spec.ts` does that for two of the examples.

### Golden files for a scaffold

A scaffold or a language plugin's `generate` writes files; compare them with a reviewed copy.
`scaffold-widget-service/scaffold_test.go` generates a project into a temporary directory and
compares every file with `testdata/golden` (each stored with a `.golden` suffix, so that no tool takes a golden `README.md` for a page); `go test -update` rewrites the golden files and a diff in
review shows what a template change did:

```text
$ GOWORK=off go test -update ./... && GOWORK=off go test -run Golden -v ./...
=== RUN   TestGolden
--- PASS: TestGolden (0.01s)
```

Fix the inputs of the test (name, directory, workspace, settings): nothing may vary between runs, or
the golden files will.

## A workspace test

The last layer is smith itself. `language-widget/workspace-test.sh` builds the plugin, checks it,
creates a throwaway workspace, declares the plugin, adds projects with it and runs their tasks twice,
expecting the second run to come from the cache:

```text
$ sh workspace-test.sh
ok
```

Run something like it in your plugin's CI, with the smith version you support. It catches what unit
tests cannot: a command that cannot be started from smith's working directory, an output outside
`dist/`, a cache key that changes every run.

## Common mistakes

| You see | Cause and remedy |
|---|---|
| the test passes but `smith plugin check` fails | the test connected to a different type, or never built the binary; check the program smith starts |
| `conformance` passes but the real workspace shows no task | the convention names no `Languages`, or is not configured with `[conventions.<name>]` |
| a golden test fails on another machine | the output contains a path, a time or an unsorted map |
| the Windows test cannot start the plugin | path form without `.\`, or a script without a launcher |
| the child-process test hangs | the test did not close the plugin's stdin, or did not send `shutdown` |

## See also

- [Troubleshooting plugins](../using/troubleshooting-plugins.md)
- [Packaging and distributing a plugin](packaging-a-plugin.md)
- `packages/plugin-sdk/host/plugin_client_test.go` and `packages/plugin-sdk/conformance`: the SDK's own tests
