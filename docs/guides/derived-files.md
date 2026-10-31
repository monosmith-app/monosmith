# Derived files

Some config repeats in every project and follows from what smith already knows. A `tsconfig.lib.json`
differs between packages only by its `references`, which are the project's internal dependencies. A
Jest `displayName` is the package name. Smith can write these parts for you, keep a record of what it
wrote, and report a hand edit instead of overwriting it.

Derived files are opt-in. A workspace that does not turn them on sees no change.

## Turn it on

```toml
# smith.toml
[derive]
enabled = true
```

Then run:

```sh
smith derive
```

The first run is safe on an existing workspace. A file that already holds what the renderer would
write is only recorded. A file that differs is left alone and shown as a diff (state `diverged`).
Smith writes a file only when it is missing, or when smith wrote it before and nobody edited it since.

Commit the files and `smith.lock`. The lock file is the memory of smith: for each derived file it
holds the owner, the renderer, its version and a hash of what smith wrote.

## What smith derives today

Two kinds of Node renderer exist. The first two serve a converted workspace. The last two serve a
workspace made by `smith new` and `smith add`.

| Renderer | File | Mode | Derives |
|---|---|---|---|
| `node.tsconfig.references` | `<project>/tsconfig.lib.json` | keys | the `references` key: one entry for each internal dependency of the project graph, pointing at the `tsconfig.lib.json` (else `tsconfig.json`) of that project |
| `node.jest.display-name` | `<project>/jest.config.{ts,js,cjs,mjs}` | region | the `displayName` line: the package name |
| `typescript.tsconfig` | `<project>/tsconfig.json` of a lib, package or app that smith scaffolded | file | the whole file, as the scaffold wrote it |
| `vite.tsconfig` | `<project>/tsconfig.json` of an app made with `--with` choices | file | the whole file, from the saved choices |

The last two files are the scaffold's own. When derive is on, `smith add` records them in
`smith.lock`. Smith derives only a file that is exactly what the scaffold writes, or that the record
owns, so a `tsconfig.json` that you wrote is not touched. To change one, use `--eject`, or change
the stack. A deleted file comes back whole ([ADR 0049](../architecture/decisions/0049-a-scaffold-is-also-the-renderer-of-its-own-files.md)).

A Jest config needs markers around the line smith owns. Smith does not add them:

```js
module.exports = {
  // smith:begin node.jest.display-name
  displayName: '@acme/web',
  // smith:end node.jest.display-name
  testEnvironment: 'node',
}
```

The three modes are explained in [Derived files](../reference/derived-files.md).

## Day to day

- A dependency between two projects changes. Run `smith derive`. The `references` of the projects
  that changed are rewritten. Nobody edited them, so smith may write.
- `smith derive --check` writes nothing. It fails when a derived file or `smith.lock` is out of
  date. `smith ci verify` runs it when `[derive] enabled = true`.
- `smith doctor` reports each file that needs action, with its owner.
- You delete a derived file, or the lines between the markers. Smith treats that as a request to
  write it again, not as an edit. `smith derive` restores it. A `file` mode file comes back whole.
  A `keys` mode file (`tsconfig.lib.json`) comes back with the keys smith owns, and nothing else.

## Who owns a file

`smith why <file>` says whether smith or you own a file, which renderer wrote it, and how to change
it. It also answers for files that smith does not derive. See [`smith why`](../manual/why.md).

## A hand edit

You edit a file that smith owns. The next `smith derive` does not overwrite it. It prints the
state `drift` and the diff, and it exits with an error until you choose:

| You want | Run | Effect |
|---|---|---|
| the derived content back | `smith derive --replace <file>` | smith overwrites the file |
| to keep your edit as it is | `smith derive --adopt <file>` | smith takes the edit as the new baseline. Only for a file smith owns by keys, that still has every derived key |
| to own the file | `smith derive --eject <file>` | the record says owner `you`; smith does not touch the file again |

`smith derive --explain <file>` says who owns a file, which renderer wrote it, and why it is in its
state. An adopted edit stands while the renderer's inputs stay the same. If the inputs change, the
file is in drift again, because smith cannot merge your edit with the new content. A three-way merge
is planned ([ADR 0046](../architecture/decisions/0046-derived-files-have-an-owner-and-a-record.md)).

## Limits

- Smith edits JSON with its own writer. A file in `keys` mode is indented with two spaces when
  smith writes it. JSON with comments is reported as `broken`.
- A region needs its markers and its file. Smith never creates them.
- Other derivable parts (a shared tsconfig base, test setup, source locations) are not rendered yet.
  The release tag prefix needs no renderer: use `tag_format` in `[release]`.

Plugin authors: [Rendering derived files](../plugins/creating/derived-files.md).
