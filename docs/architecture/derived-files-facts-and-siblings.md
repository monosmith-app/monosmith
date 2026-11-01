# Derived files, shared facts and sibling inference

Three features of smith 2.0 keep the config of many projects the same without a copy-and-paste
step: facts that plugins and `smith.toml` publish, siblings that a new project follows, and files
that smith derives from both and from the project graph. This page says how they fit together and
where each lives in the code. For the commands see the [reference](../reference/README.md); for
writing a plugin that takes part see [derived files](../plugins/creating/derived-files.md) and
[shared facts](../plugins/creating/shared-facts.md).

The rules behind all three are the same ones: the core knows no language, plugins never call each
other, and nothing rewrites a file a person edited without asking.

```mermaid
flowchart LR
  toml["smith.toml [facts]<br/>recipe, workspace"] --> merge
  plug["plugins: Contribution.facts"] --> merge["factmerge<br/>one value per key, with its origin"]
  merge --> gen["generate / contribute<br/>(plugins read the facts)"]
  disc["discover: Project.shape"] --> dialect["dialect<br/>what siblings agree on"]
  dialect --> gen
  graph["project graph"] --> derive
  merge --> derive["derive<br/>asks plugins to render"]
  derive --> lock[("smith.lock<br/>ownership record")]
  derive --> files["tsconfig.json, jest config, ..."]
```

## Shared facts

A fact is a typed value under a namespaced key such as `module.aliases`
([ADR 0042](decisions/0042-plugins-publish-shared-facts-the-core-merges.md)). The keys the protocol
knows are validated by `packages/plugin-sdk/protocol/fact_keys_contract.go` and
`fact_value_validator.go`: `module.aliases`, `source.roots`, `test.patterns`, `build.outputs` and
`runtime.node`. The same keys are the `[facts]` table of `smith.toml`, and `[projects.<name>.facts]` sets them for one project.

- A plugin publishes facts in its `contribute` answer (`Contribution.facts`). It reads the merged
  facts in the requests that carry them (`facts` in `GenerateParams` and in the contribution requests).
- `apps/smith/factmerge` merges the layers, narrowest scope first (`[projects.<name>.facts]`, then
  `[facts]`, then the defaults that plugins publish; a recipe scope exists in the code but no layer uses it yet), into one value per key that remembers where it came from. The merge
  is pure. Two plugin defaults that differ are an error that names both plugins. A `smith.toml`
  fact decides a conflict, because it is narrower.
- `smith explain --facts` (or `--fact <key>`) shows each value with its scope and the table or plugin it came from.

## Siblings: a new project follows what exists

`smith add` looks at the projects of the same kind and language before it generates
([ADR 0041](decisions/0041-smith-add-follows-the-dialect-of-sibling-projects.md)).

1. A language plugin reports the shape of each project it discovers (`Project.shape`, a map of
   strings such as `test.runner = "@swc/jest"`). A plugin declares the shape keys it applies when it
   generates in `InitializeResult.follows`.
2. `apps/smith/dialect` counts the values. A value is dominant when at least three projects have it
   and at least two thirds of the projects that report the key have it. With fewer than three
   siblings nothing is inferred.
3. The dominant values go to every plugin of the stack in `GenerateParams.shape`. A key with no
   dominant value is split: smith asks in a terminal, and uses the built-in default with a warning
   when there is no terminal or `CI` is set.
4. After the project is generated, discovery reads its shape again. A dominant key whose value
   differs is a warning that names the key and both values. `smith add --fresh` skips all of this.

The package knows no key and no language. The Node plugin reports shapes and the TypeScript, Jest and
Vitest toolchain plugins follow them.

## Derived files

Some config follows from what smith already holds: the `references` of a `tsconfig.lib.json` are the
internal dependencies of the graph, and a Jest `displayName` is the package name. `smith derive`
writes such files and records who owns them
([ADR 0046](decisions/0046-derived-files-have-an-owner-and-a-record.md)). It is opt-in:
`[derive] enabled = true`. When it is on, `smith doctor` reports drift and `smith ci verify` runs
`smith derive --check` first.

### Who does what

| Part | Owns | Code |
|---|---|---|
| Renderer | the content of one file, or of a part of one; a stable id and a version (`node.tsconfig.references`). Writes nothing | the plugin: protocol method `derive`, capability `derive`, Go SDK interface `Deriver` |
| `derive` | what to write now: asks the plugins, compares, decides | `apps/smith/derive` |
| `derivefile` | the three ownership modes and how to merge into a file | `apps/smith/derivefile` |
| `derivelock` | the ownership record `smith.lock` | `apps/smith/derivelock` |

The request (`protocol.DeriveParams`) carries the plugin's projects, the internal dependencies of
each from the graph and the paths the record says smith owns. The answer (`DeriveResult`) lists
`DerivedFile` values (project, path, mode, renderer id and version, content or keys) and, optionally,
`diverging` files: files that differ from the shape of the siblings and belong to the project.

### Three ownership modes

| Mode | Smith owns | The person owns |
|---|---|---|
| `file` | the whole file, with a generated header when the format has comments | nothing |
| `region` | the lines between `smith:begin <renderer>` and `smith:end <renderer>` | the rest; a file without the markers is reported as broken |
| `keys` | listed top-level keys of a JSON file or of a table of a TOML file | the other keys |

### The record and the states

`smith.lock`, next to `smith.toml`, has one entry for each pair of file and renderer: `path`,
`owner` (`smith` or `you`), `renderer`, `version`, `mode`, `hash` (of the content smith last wrote)
and `wanted` (of what the renderer wanted then). The core compares what the renderer wants now,
what is on disk and the record, and puts each file in one state (`apps/smith/derive/plan_item_contract.go`):

| Status | Meaning | `smith derive` does |
|---|---|---|
| `current` | disk equals wanted, the record agrees | nothing |
| `stale` | inputs changed, nobody edited the file | rewrites it |
| `missing` | smith owns the file and it is gone | writes it again |
| `unrecorded` | disk equals wanted, no entry | records it |
| `drift` | disk differs from the record: a hand edit | prints a diff, writes nothing |
| `diverged` | the file exists, is not recorded and differs | leaves it alone |
| `broken` | smith cannot edit it (no markers, does not parse) | reports it |
| `ejected` | the owner is `you` | nothing |

A drifted file needs a choice: `--replace <file>` overwrites it, `--eject <file>` makes the owner
`you`, and `--adopt <file>` takes a hand edit of a `keys` file as the new baseline. `--check` writes
nothing and fails when anything would change. `--explain <file>` and `smith why <file>` say who owns
a file and why it is in its state. A person commits `smith.lock`.

### Renderers that exist

- The Node plugin renders the `references` of `tsconfig.lib.json` (`node.tsconfig.references`,
  `keys`) and the `displayName` of a Jest config that has markers (`node.jest.display-name`,
  `region`) ([ADR 0046](decisions/0046-derived-files-have-an-owner-and-a-record.md)).
- A scaffold is also the renderer of its own files: the TypeScript and Vite plugins render
  `tsconfig.json` with the template `smith add` used, in `file` mode, so a workspace made by smith
  derives with no markers ([ADR 0049](decisions/0049-a-scaffold-is-also-the-renderer-of-its-own-files.md)).
- A renderer reads the shape of siblings: when at least three projects hold a config file at the
  same path and two thirds hold it in one shape, that shape is the convention and each project's
  copy is derived. A file that differs is `diverging`, not drift
  ([ADR 0050](decisions/0050-config-that-sibling-projects-share-is-derived-from-their-shape.md),
  `libs/plugin-node/siblingfiles`).

## Plan, confirm and undo

Every command that writes files shows what it will do first
([ADR 0047](decisions/0047-mutating-commands-show-a-plan-and-can-be-undone.md)).
`apps/smith/operationplan` holds the plan: groups of lines, each with an action (`create`, `change`,
`remove`), a path and the owner from `smith.lock`. A command asks to go ahead only in a terminal,
outside CI, and without `--yes`.

With `[workspace] commit_operations = true`, `apps/smith/operation` makes `add`, `remove`, `upgrade`,
`derive` and `sync` one conventional commit each, when the work tree is clean, and records it in
`.smith/operations.json` (local, git-ignored). `smith undo` reverts the newest recorded operation
with `git revert`, never rewrites history, and refuses when a later commit or an uncommitted change
touches the same files.

## Explaining the workspace to people and agents

Two commands read the same facts, graph and record
([ADR 0048](decisions/0048-why-report-and-agent-file-explain-the-workspace-to-people-and-agents.md)):

- `smith report` (`apps/smith/report`) writes one self-contained HTML page: the dependency graph as
  inline SVG, the doctor results, derived-file drift, the choices of each project and the release plan.
  The page makes no request and runs no script.
- `smith agent file` (`apps/smith/agentdoc`) writes `AGENTS.md`, and a `CLAUDE.md` that points to
  it, from the workspace's rules. The generated text sits between region markers, so the text a person
  writes in the same file survives. `smith agent install` (`apps/smith/agentfiles`) adds the Claude
  Code subagent and registers `smith mcp`.
