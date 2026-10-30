# smith kinds

`smith kinds` lists the project kinds [`smith add`](add.md) accepts: a role (`app`, `lib`,
`package`) combined with a language, such as `go-lib` or `node-package`. With `--json` it
prints them as one document for editors. A plugin that scaffolds only some kinds (Bicep: no
package) is listed for those only.

## When to use it

- To see which languages this smith can create, including those a plugin added.

It needs no workspace.

## Synopsis

```sh
smith kinds [--json]
```

## Arguments and flags

<!-- smith:begin kinds -->
List the project kinds smith add accepts (--json for editors)

Usage: `smith kinds [flags]`

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--json` | boolean | off | print the kinds as one JSON document |

<!-- smith:end kinds -->

`--json` prints an array of `{ "kind", "language", "label", "description", "flags" }`, where
`flags` lists the `smith add` flags that apply (such as `stack`).

## Exit codes

0.

## Files and environment

None. It reads no file of the workspace and works outside one.

## Examples

```sh
smith kinds
```

```text
bicep-app        Bicep: an app: deployed or run, never published. Built with the workspace's default stack, or --stack.
bicep-lib        Bicep: an internal library: imported by other projects, never published. ...
cpp-app          Cpp: an app: deployed or run, never published. ...
...
rust-package     Rust: a package: published to its registry on release. ...
```

Thirty-nine rows: applescript, bicep, cpp, dart, dotnet, go, jvm, lua, node, php, python, recipe, ruby and rust,
each as app, lib and package (applescript, bicep and lua: app and lib only).
Thirty-seven rows: applescript, bicep, cpp, dart, dotnet, go, jvm, node, php, python, recipe, ruby and rust,
each as app, lib and package (applescript and bicep: app and lib only).

## Related

[`smith add`](add.md), [`smith stacks`](stacks.md), [`smith plugin list`](plugin-list.md).
