# language-widget

A smith **language plugin** in Go, on the plugin SDK. It teaches smith the widget project type: a
directory with a `widget.json` is a project, `uses` in it become graph edges, and each project gets a
cached `build` and `test` task. It also scaffolds widgets (`smith add lib core --lang widget`),
writes the release version into `widget.json`, and reports a doctor finding with a fix.

Guide: [Creating a language plugin](../../../plugins/creating/language-plugin.md).

## Files

| File | What it is |
|---|---|
| `main.go` | serves the protocol when smith started the program (`SMITH_PLUGIN=1`), runs the command line tool otherwise |
| `widget.go` | the plugin: `Describe`, `Discover`, `Generate`, `WriteVersion`, `Doctor` |
| `manifest.go` | reading and rewriting `widget.json` |
| `tool.go` | the command line the tasks call: `check`, `build <dir>` |
| `widget_test.go` | the conformance suite and behaviour tests, over an in-process connection |
| `workspace-test.sh` | an end-to-end test in a throwaway workspace |
| `plugin.json` | how to install it on the search path |

## Build, test, check

```sh
GOWORK=off go vet ./... && GOWORK=off go test ./...
GOWORK=off go build -o smith-plugin-widget .
smith plugin check ./smith-plugin-widget --generate
sh workspace-test.sh
```

## Use

```toml
[plugins.external.widget]
command = ["/path/to/smith-plugin-widget"]
```

```sh
smith add lib core --lang widget
smith add app shop --lang widget      # then add "uses": ["core"] to apps/shop/widget.json
smith graph
smith run build test
```
