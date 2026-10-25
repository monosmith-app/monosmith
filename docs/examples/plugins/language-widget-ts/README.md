# language-widget-ts

The language plugin of [`language-widget`](../language-widget), in **TypeScript on
[`@monosmith/plugin-sdk`](../../../../packages/plugin-sdk-ts)**. It teaches smith the widget project
type: a directory with a `widget.json` is a project, `uses` in it become graph edges, and each project
gets a cached `build` and `test` task. It also scaffolds widgets (`smith add lib core --lang widget`),
writes the release version into `widget.json` keeping its key order, and reports a doctor finding with
a fix.

For the same requests it gives the same bytes as the Go plugin, which
`packages/plugin-sdk-ts/integration` checks. The Python twin is [`language-widget-py`](../language-widget-py), compared
with both by `packages/plugin-sdk-py/tests/integration`.

Guide: [Creating a language plugin](../../../plugins/creating/language-plugin.md); the SDK:
[The TypeScript SDK](../../../plugins/creating/typescript-sdk.md).

## Files

| File | What it is |
|---|---|
| `plugin.ts` | the entry point: serves the protocol when smith started the program (`SMITH_PLUGIN=1`), runs the command line tool otherwise |
| `widget.ts` | the plugin: `describe`, `discover`, `generate`, `writeVersion`, `doctor` |
| `manifest.ts` | reading and rewriting `widget.json` |
| `tool.ts` | the command line the tasks call: `check`, `build <dir>` |
| `widget.test.ts` | the conformance suite and behaviour tests, over an in-process connection (`node --test`) |
| `workspace-test.sh` | an end-to-end test in a throwaway workspace |
| `plugin.cmd`, `plugin.json` | the Windows launcher, and how to install it on the search path |
| `package.json` | `"type": "module"`, the SDK as the one dependency |

Run with Node.js 24, which runs TypeScript files as they are (it only removes the types). Install the
SDK first outside this repository: `npm install`.

## Test, check

```sh
node --test
smith plugin check ./plugin.ts --generate     # Linux, macOS: chmod +x plugin.ts first (not run there)
sh workspace-test.sh
```

```text
PS> smith plugin check .\plugin.cmd --generate
```

## Use

```toml
[plugins.external.widget]
command = ["node", "/path/to/plugin.ts"]
```

smith starts an external plugin only after you trust it: the first command that needs it asks, or run `smith plugin trust widget` (`--yes` in a script, `--pin` to also write the digest CI checks). It starts with a minimal environment; see [Plugin security](../../../plugins/using/plugin-security.md).

```sh
smith add lib core --lang widget
smith add app shop --lang widget      # then add "uses": ["core"] to apps/shop/widget.json
smith graph
smith run build test
```
