# ADR 0046: Derived files have an owner and an ownership record

- Status: accepted
- Date: 2026-10-31

## Context

Some config repeats in every project and follows from what smith already knows. In a measured
converted workspace, 22 packages carry 2,584 lines of per-package config. `tsconfig.lib.json`
differs between packages only by its `references`, the project's internal dependencies, which the
project graph holds. The Jest `displayName` is the package name. `tsconfig.json` and `.spec.swcrc`
are identical copies, and `tag_prefix` repeats. All of this is derivable (epic
[#819](https://github.com/monosmith-app/monosmith/issues/819), phase 4, issue
[#823](https://github.com/monosmith-app/monosmith/issues/823)).

A generator that rewrites such files is dangerous: a person edits them, and a silent overwrite
destroys the edit. The rules of the epic apply: the core stays language-free, plugins never call
each other, nothing silently rewrites a file a person edited, the protocol stays at version 1, and
there are no shims.

## Options considered

**Generate once, like a scaffold, and forget.** Rejected: the file then goes stale when the graph
changes, and nothing says which parts follow the graph.

**Regenerate always, with a header that says "do not edit".** Rejected: it loses hand edits
without warning, and it forces the whole file to be generated, even when only one key is derivable.

**The core renders the files.** Rejected: a tsconfig is Node knowledge. The plugin that knows the
language renders; the core owns the ownership rules.

**Reuse the name `smith sync`.** Rejected: `smith sync` already aligns dependency versions.

**Reuse `smith eject`.** Rejected: `smith eject <tool>` copies a tool default into the repository.
A derived file is ejected per file, and it stays where it is.

## Decision

**Renderer.** A renderer is a part of a plugin. It maps facts and graph data to the content of one
file, or to a part of one. It has a stable id (`node.tsconfig.references`) and a version. A plugin
that has renderers advertises the new optional capability `derive` and answers the new method
`derive`. The request carries the plugin's projects, the internal dependencies of each project
from the project graph, and the paths the ownership record says smith owns. The answer lists
`DerivedFile` values: project, path, mode, renderer id and version, and the content. The plugin
writes nothing. The change is additive, so the protocol stays at version 1.

**Three ownership modes.** The mode says how much of the file smith owns.

- `file`: smith owns the whole file. If the format has comments, the file starts with a generated
  header that names the renderer and its version.
- `region`: smith owns the lines between `smith:begin <renderer>` and `smith:end <renderer>`
  markers. The person owns the rest. A file without markers is reported, not edited.
- `keys`: smith owns the listed top-level keys of a JSON file, or of a table of a TOML file. The
  person owns the other keys. The core edits the file with the SDK helpers `jsondoc` and
  `tomltext`, which keep the order of keys and, for TOML, the comments. JSON that has comments is
  not read: the file is reported as broken.

**Ownership record.** `smith.lock` is a TOML file next to `smith.toml`. It has one entry for each
pair of file and renderer: `path`, `owner` (`smith` or `you`), `renderer`, `version`, `mode`, `hash`
and `wanted`. `hash` is the hash of the content that smith owns, as smith last wrote it. For a
file it is the whole content, for a region the lines between the markers, and for keys the
canonical form of the owned keys (the rest of the file does not count). Line endings do not count.
`wanted` is the hash of what the renderer wanted at that time. It tells a change of the inputs
apart from an edit of the file. A person commits `smith.lock`.

**Behaviour.** The core compares three things: what the renderer wants now, what is on disk, and
the record.

| State | Condition | `smith derive` does |
|---|---|---|
| current | disk equals wanted, record agrees | nothing |
| stale | inputs changed, disk equals the record | rewrites the file |
| missing | the file, or the owned part of it, is gone | writes it again |
| unrecorded | disk equals wanted, no matching entry | records it |
| drift | disk differs from the record (a hand edit) | reports the diff, writes nothing |
| diverged | file exists, no entry, differs from wanted | reports the diff, writes nothing |
| broken | no markers, or the file does not parse | reports it, writes nothing |
| ejected | the owner is `you` | nothing |

Smith never overwrites a file in the states drift, diverged or broken. A person chooses, for each
file, with one of three flags:

- `--replace <file>` lets the derived content overwrite the file. Smith prints the diff in the
  run that reported the state, so a person sees it before the choice.
- `--adopt <file>` takes the hand-edited content as the new baseline. It applies only to a `keys`
  file that still has every derived key. For other files smith shows the diff and refuses. An
  adopted edit stands while the inputs stay the same. If the inputs change later, the file is in
  drift again, because smith cannot merge the two changes.
- `--eject <file>` records the owner `you`. Smith does not touch the file again.

A hand edit that equals what the renderer wants is not drift: the record is refreshed.

**Commands.** The new command is `smith derive`. `smith derive --check` writes nothing and fails
when a file or the record is out of date, or when a file waits for a choice. `smith derive
--explain <file>` states the owner, the renderer, the state and the diff. `smith eject` and `smith
sync` keep their meaning.

**Opt-in.** `[derive] enabled = true` in `smith.toml` turns the feature on. The default is
`false`. With the default, `smith derive` stops with a message, `smith doctor` and `smith ci
verify` do not look at derived files, and nothing changes for a workspace that does not opt in.
When it is on, `smith doctor` gives one finding for each file that needs action, with the owner
in the detail, and `smith ci verify` runs `smith derive --check` before the tasks.

**First renderers.** The Node plugin renders two files:

- `node.tsconfig.references` (mode `keys`): the `references` key of `tsconfig.lib.json`, from the
  internal dependencies of the graph. A project is rendered when its file exists or is owned.
- `node.jest.display-name` (mode `region`): the `displayName` line of a `jest.config.*` file that
  has the markers. The value is the package name.

`tag_prefix` is not rendered: the workspace-level `tag_format` already covers it.

## Consequences

- A workspace that does not opt in sees no change. The protocol grows by one optional method, and
  the generated types of the Go, TypeScript, Python and C# SDKs carry it.
- A person who edits a derived file gets a diff and a choice, not a lost edit. Deleting a derived
  file or its owned part is not an edit: `smith derive` writes it again.
- Because `keys` mode hashes only the owned keys, an edit of another key is never drift.
- JSON that `jsondoc` writes is indented with two spaces. A file in `keys` mode is reformatted
  when smith writes it. A file that smith does not write keeps its format.
- Not done: the other seed intents of the issue (shared tsconfig base, test setup, source
  locations, source resolution condition), `file` mode renderers (no renderer uses it yet), and
  `smith why <file>` (phase 6).

## Future: three-way merge

Smith keeps one hash, not the old content, so it cannot merge a hand edit with a change of the
inputs. A later step can store the last written content (in `.smith/`) and do a three-way merge:
base is the last written content, ours the disk, theirs the new render. Until then, drift after a
changed input needs a choice by hand.

## Sources

- Epic [#819](https://github.com/monosmith-app/monosmith/issues/819), issue
  [#823](https://github.com/monosmith-app/monosmith/issues/823)
- [ADR 0042](0042-plugins-publish-shared-facts-the-core-merges.md), [The plugin
  protocol](../plugin-protocol.md)
- `apps/smith/derive`, `apps/smith/derivefile`, `apps/smith/derivelock`,
  `libs/plugin-node/deriving`, `packages/plugin-sdk/protocol/derive_contract.go`
