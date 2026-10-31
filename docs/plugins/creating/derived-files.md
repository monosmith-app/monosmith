# Rendering derived files

A language plugin can render files that follow from facts and the project graph. This is an optional
capability. A plugin that does not implement it is not asked, and the protocol stays at version 1.

| | |
|---|---|
| Method | `derive` |
| Capability | `derive` |
| Go interface | `Deriver` (`derive` in the TypeScript and Python SDKs, `IDeriver` in C#) |
| Request | `DeriveParams{projects, graph, owned}` |
| Answer | `DeriveResult{files}`, each a `DerivedFile` |
| Called by | `smith derive`, `smith doctor`, `smith ci verify` (only when `[derive] enabled = true`) |

The plugin writes nothing. It answers with content, and the host decides, from `smith.lock`,
whether to write it. Read [Derived files](../../reference/derived-files.md) for the modes and states.

## The request

- `projects`: the projects the plugin serves.
- `graph`: for each project name, the names of the projects it depends on inside the workspace,
  sorted. The host resolved these edges. Do not compute them again.
- `owned`: the paths that `smith.lock` says smith owns. Propose such a file again even when it is
  missing on disk. This is how a deleted derived file comes back. Do not propose a file that is
  neither on disk nor in `owned`: the person did not ask for it.

## The answer

| `DerivedFile` field | Meaning |
|---|---|
| `project`, `path` | the project, and the file relative to the workspace, with forward slashes |
| `mode` | `file`, `region` or `keys` |
| `renderer`, `version` | a stable id (`<plugin>.<what>`), and a version you raise when the same inputs give new output |
| `content` | the whole file (`file`) or the lines of the region without the markers (`region`) |
| `comment` | the comment leader of the file (`//`, `#`). Needed for `region`; for `file` it makes the host write a generated header |
| `format`, `table`, `keys` | for `keys`: `json` or `toml`; the TOML table that holds the keys; the owned keys and their values |

Make the answer deterministic and independent of the order of the files. A reference to another
derived file must not depend on whether that file exists yet: count a path in `owned` as existing.
The Node plugin does this in `libs/plugin-node/deriving`.

## Test it

`smith plugin check` calls `derive` on a plugin that advertises it. Unit-test the renderer with a
temporary directory: write the files, call the function, compare the `DerivedFile` values. Run the
host side with an end-to-end test of `smith derive`, as `apps/smith/goldenpath` does.
