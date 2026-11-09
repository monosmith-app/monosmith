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

Version 2.0 is the first stable release of this line. It replaces the 1.x command line,
configuration and plugin protocol.

## Written exceptions

Each package below holds more than 12 production files, the review trigger of
the vertical-slice rules. Rule waived: "split a slice that holds two outcomes".
Owner: the smith maintainers. None is permanent unless it says so.

Owned by this README's package set (kept as one slice, temporary until a split
with a clear one-way seam is made):

| Path | Constraint |
|---|---|
| `dependencies` (23 files) | install, uninstall, sync, align and up share one declaration model and one command set. The seams (updates, sync, root install) pass the same `Update` and `SyncReport` types, so a split would need a new contract slice first. |
| `projectgraph` (20) | the graph, its conventions, stacks and registries are read from one `Project` model; every file builds or reads it. |
| `adoption` (19) | one adoption flow: inspect, plan, move, settle, hand over. The steps share the plan and report types. |
| `toolinstall` (18) | detect, plan and install one tool set; the files share `Requirement` and the plan. |
| `operation` (14) | commit, undo and the operation log share the `Record`. |
| `taskrun` (13) | one run: schedule, process, cache, summary. |

Pending, owned by other clean-up work (owner-confirmed; the lead removes a row
when the package is split): `scaffold` (50), `cipipeline` (48), `doctor` (46),
`release` (38), `workspace` (37), `convertplan` (34), `convert` (18).

`goldenpath` has no production file: it is a test-only package of end-to-end
scenarios (`<scenario>_use_case_test.go`). Each test names its scenario, so the
rule "a test shares the basename of the file it tests" does not apply. This is
permanent.

`internallink`, `operation` and others keep `_client` files that call a
terminal, git or a plugin: each is an adapter for an outside system.

## Shared version ordering

`semver` holds the one comparison of versions with pre-releases
(`pluginmarket` uses it). `pluginhost/version_constraint_policy.go` keeps its own
triple comparison (it ignores pre-release and build suffixes and rejects text)
and `toolinstall/requirement_policy.go` keeps its own `CompareVersions` (it reads
a part that is not a number as 0): both give different answers than `semver`.
