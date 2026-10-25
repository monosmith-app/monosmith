# smith

The `smith` command: MonoSmith's language-agnostic core and the entry point
of every built-in plugin. One package per slice; [docs/contributing/development.md](../../docs/contributing/development.md)
lists them, and the [README](../../README.md) and [docs](../../docs/) describe
what the command does.

```sh
go build -o ../../dist/dev/smith .        # from this directory
go run . --help
```

It builds inside the repository's `go.work`, which ties it to `libs/plugin-*`
and `packages/plugin-sdk`; `go install` from outside the repository does not
work. Released as `smith@<version>`: see [installing](../../docs/getting-started/installing.md).
