# smith plugin draft

`smith plugin draft <sample-folder> --lang <name>` looks at a small working project in a language
smith does not know and writes `<name>.plugin-draft.toml`: the files that name a project and give it a
version, the line that holds the version, the tasks the project already runs, and the CI lines that
publish. Every value says where it came from. It then prints what a plugin author still writes.

The draft is not a plugin. smith has no declarative plugin format: a language plugin is a program that
answers the protocol. The draft is the list of facts the author, or an assistant writing the plugin,
starts from ([ADR 0061](../architecture/decisions/0061-teach-by-example-drafts-the-facts-not-the-plugin.md)).

## When to use it

Before you write a [language plugin](../plugins/creating/language-plugin.md) for a language or build
system smith does not know.

## Synopsis

```sh
smith plugin draft <sample-folder> --lang <name> [-o <file>] [--stdout]
```

## Arguments and flags

<!-- smith:begin plugin draft -->
Draft what a language plugin needs from a sample project: manifests, version line, tasks, publish step

Usage: `smith plugin draft <sample-folder> <name> [flags]`

| Argument | Required | Meaning |
|---|---|---|
| `sample-folder` | yes |  |
| `name` | yes |  |

| Flag | Type | Default | Effect |
|---|---|---|---|
| `--lang <lang>` | string |  | the language's name, lower case (required) |
| `-o, --output <output>` | string | <lang>.plugin-draft.toml | the draft file (default <lang>.plugin-draft.toml) |
| `--stdout` | boolean | off | print the draft instead of writing a file |

<!-- smith:end plugin draft -->

## What it finds

- **Manifests**: files at the sample's root with a name key and a version key (JSON, TOML, YAML, XML
  and Zig-style `.name =`). README, license and lock files are skipped.
- **Tasks**: targets of a `Makefile` or `justfile`, and the `scripts` of a JSON file, for the names
  smith knows (build, test, lint, format, check, clean, dev, run, typecheck, publish, release).
- **Publish**: lines of the root files and `.github/workflows/*` that contain publish, upload or
  push as a command.

It reads only the sample's root and its GitHub workflows; it runs nothing.

## Then

```sh
smith add lib lang-zig --stack go-smith-plugin
smith plugin check ./lang-zig --generate
```

The conformance run against the real tool needs the plugin; it is the second step, not part of the
draft.
