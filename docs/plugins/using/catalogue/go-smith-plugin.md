# The go-smith-plugin plugin

The `go-smith-plugin` plugin is a scaffold: it writes a smith plugin in Go, on the plugin SDK, with
a conformance test and a README that says how to declare it. It is the quickest way to start a
plugin of your own.

## At a glance

| | |
|---|---|
| Kind | `convention` |
| Roles | `scaffold` |
| Requires | language plugin `go` |
| Applies to | `go` projects |
| Capabilities | `generate` |
| Stack it publishes | `go-smith-plugin`: language `go`, scaffold `go-smith-plugin`, no toolchain |
| Project kinds | `app` only: a plugin is an executable |

## What it does

```sh
smith add app my-plugin --stack go-smith-plugin
```

```text
Adding app my-plugin (stack go-smith-plugin) in apps/my-plugin
go: creating new go.mod: module example.com/gsp/apps/my-plugin
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

The generated plugin is a **structure** plugin for Node, Go and Python projects that gives each a
`hello` task running `{smith} --version`: the smallest working plugin, to be replaced with your
own. Its handshake:

```go
return protocol.InitializeResult{
	Plugin:    protocol.Peer{Name: Name, Version: p.version},
	Kind:      protocol.KindConvention,
	Roles:     []protocol.Role{protocol.RoleStructure},
	Languages: []string{"node", "go", "python"},
}, nil
```

The README it writes declares the plugin and applies it:

```toml
[plugins.external.my-plugin]
command = ["go", "run", "./apps/my-plugin"]

[conventions.my-plugin]
```

The module depends on the released SDK (here v0.33.0, fetched from the network).

Asking for a library or a package fails: `a smith plugin is an executable: add it as an app (smith
add app x --stack go-smith-plugin)`.

## Settings, releasing, CI

No settings. The scaffold has no `publish` or `ci` method; the project is an ordinary Go app that
releases like one (see [go](go.md#releasing)).

## Verified

smith 1.62.0, Windows 11, Go 1.27.0. The command above; `smith test -p my-plugin` passed (the
conformance test); building the app and running `smith plugin check apps/my-plugin/my-plugin.exe`
passed (six checks).

## See also

- [Creating plugins](../../creating/README.md): the guides that take it from here
- [The go plugin](go.md)
