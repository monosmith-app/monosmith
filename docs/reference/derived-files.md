# Derived files

Reference for `smith derive`: the ownership modes, the ownership record `smith.lock`, the states of
a file, and the renderers. The task-oriented guide is [Derived files](../guides/derived-files.md);
the decision is [ADR 0046](../architecture/decisions/0046-derived-files-have-an-owner-and-a-record.md).

## Setting

| Key | Type | Default | Meaning |
|---|---|---|---|
| `[derive] enabled` | boolean | `false` | turns derived files on: `smith derive` works, `smith doctor` reports drift, `smith ci verify` runs `smith derive --check` |

## Ownership modes

A renderer chooses the mode of each file. The mode says how much of the file smith owns.

| Mode | Smith owns | The rest | Hash covers |
|---|---|---|---|
| `file` | the whole file. A generated header names the renderer when the format has comments | nothing | the whole file |
| `region` | the lines between `smith:begin <renderer>` and `smith:end <renderer>`. The markers use the comment leader of the file (`//`, `#`) | yours | the lines between the markers |
| `keys` | the listed top-level keys of a JSON file, or of one table of a TOML file | yours, in order, with its comments (TOML) | the owned keys in canonical form (sorted, compact) |

Line endings do not count in a hash. Smith keeps the line ending of the file when it writes.

## smith.lock

A TOML file next to `smith.toml`. Commit it. Do not edit it; use `smith derive --eject`.

```toml
schema = 1

[[file]]
path = 'packages/web/tsconfig.lib.json'
owner = 'smith'
renderer = 'node.tsconfig.references'
version = 1
mode = 'keys'
hash = 'sha256:…'
wanted = 'sha256:…'
```

| Field | Meaning |
|---|---|
| `path` | the file, relative to the workspace, with forward slashes |
| `owner` | `smith`: smith rewrites the file when its inputs change. `you`: smith never touches it |
| `renderer`, `version` | the renderer that wrote the file and its version |
| `mode` | `file`, `region` or `keys` |
| `hash` | the hash of the content smith owns, as it last wrote it, or as a person adopted it |
| `wanted` | the hash of what the renderer wanted at that time. It differs from `hash` after `--adopt` |

A file can have several entries when several renderers own different keys or regions of it. A
workspace with no derived file has no `smith.lock`.

## States

`smith derive` prints the files that need action. `smith derive --check` fails when there is one.

| State | Meaning | `smith derive` |
|---|---|---|
| `current` | the file is as the renderer wants it, and the record agrees | nothing |
| `stale` | the renderer's inputs changed, and nobody edited the file | rewrites it |
| `missing` | the file, or the owned part of it, is gone | writes it again |
| `unrecorded` | the file is as the renderer wants it, but the record has no matching entry | records it |
| `drift` | the content differs from what smith last wrote: a hand edit | prints the diff, writes nothing |
| `diverged` | the file exists, the record has no entry, and it differs from the render | prints the diff, writes nothing |
| `broken` | smith cannot edit it: no markers, or the file does not parse | prints the reason, writes nothing |
| `ejected` | the owner is `you` | nothing |

The exit code of `smith derive` is not zero while a file is in `drift`, `diverged` or `broken`.

## Choosing

| Flag | For | Effect |
|---|---|---|
| `--replace <file>` | `drift`, `diverged` | overwrites the file with the render |
| `--adopt <file>` | `drift` of a `keys` file that has every derived key | records the file as it is. Refused otherwise, with the diff |
| `--eject <file>` | any derived file | sets the owner to `you` |
| `--explain <file>` | any derived file | prints the owner, renderer, state and, for a file in `stale`, `drift` or `diverged`, the diff |

## Renderers

| Id | Plugin | File | Mode | Content |
|---|---|---|---|---|
| `node.tsconfig.references` | node | `tsconfig.lib.json` of each project that has it (or that the record owns) | `keys` | `references`: one `{ "path": … }` for each internal dependency of the project graph, sorted by project name |
| `node.jest.display-name` | node | `jest.config.ts`, `.js`, `.cjs` or `.mjs` that has the markers | `region` | `displayName: '<package name>',` |

| `typescript.tsconfig` | typescript | `tsconfig.json` of a project the plugin scaffolded | `file` | the whole file, rendered from the scaffold template of the project |
| `vite.tsconfig` | vite | `tsconfig.json` of an app made from choices | `file` | the whole file, rendered from the saved choices |

The last two renderers derive a file only when the record owns it or when it is, line for line, what
the template renders. A project whose `tsconfig.lib.json` has no dependency and no `references` key is skipped.

## Where it is checked

| Place | Check |
|---|---|
| `smith derive --check` | the whole plan |
| `smith doctor` | one finding for each file that needs action; the detail names the state and the owner |
| `smith why <file>` | the owner, renderer, state and the ways to change one file |
| `smith ci verify` | `smith derive --check`, before the tasks, when `[derive] enabled = true` |
