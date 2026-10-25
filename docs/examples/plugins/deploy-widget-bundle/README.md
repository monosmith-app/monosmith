# deploy-widget-bundle

A smith **deploy plugin** in Go. For every widget app it adds a `bundle` task, and on release it
deploys the app's build as `<name>-<version>.txt` into a folder (`dist/deploy` by default, or the
`target` setting). A deploy that finds the version already there is skipped; `--dry-run` writes
nothing. It asks the pipeline for an optional `WIDGET_DEPLOY_TOKEN` and its `doctor` rejects a target
outside the workspace.

Guide: [Creating a deploy plugin](../../../plugins/creating/deploy-plugin.md).

## Files

| File | What it is |
|---|---|
| `main.go` | serves the protocol for smith; runs the `bundle` command for the task |
| `plugin.go` | `Describe`, `Contribute`, `Publish`, `CI`, `Doctor` |
| `plugin_test.go` | conformance, idempotence, dry run, error message and doctor tests |
| `plugin.json` | how to install it |

## Build, test, check

```sh
GOWORK=off go vet ./... && GOWORK=off go test ./...
GOWORK=off go build -o smith-plugin-widget-bundle .
smith plugin check ./smith-plugin-widget-bundle
```

## Use

```toml
[plugins.external.widget-bundle]
command = ["/path/to/smith-plugin-widget-bundle"]

[conventions.widget-bundle]

[projects.shop.settings.widget-bundle]
target = "dist/deploy"
```

smith starts an external plugin only after you trust it: the first command that needs it asks, or run `smith plugin trust widget-bundle` (`--yes` in a script, `--pin` to also write the digest CI checks). It starts with a minimal environment; see [Plugin security](../../../plugins/using/plugin-security.md).

```sh
smith release --dry-run
smith release --no-push --yes       # deploys to dist/deploy and tags locally
```
