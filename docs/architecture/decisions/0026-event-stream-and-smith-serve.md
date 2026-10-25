# ADR 0026: A documented event stream and `smith serve`

- Status: accepted
- Date: 2026-10-20

## Context

The terminal interface and the IDE plugins both need the engine to do things and to stream
results. The design proposal ([#359](https://github.com/monosmith-app/monosmith/pull/359))
wanted them to call the same functions so that what they show and do cannot drift.

## Options considered

From the proposal: the terminal interface calls the functions in process while IDE plugins
start the engine as a child process; `smith mcp` and a planned `smith lsp` become views on
the same server where they can. An embedded engine in each plugin was rejected in
[ADR 0025](0025-ide-plugins-run-the-installed-smith.md).

## Decision

Everything that takes time is a run, and a run tells what it does as events: `run.started`,
`task.queued`, `task.started`, `task.output`, `task.finished`, `run.finished`, each with a
run id, sequence number and time. `smith run --json-events` prints them as JSON lines.
`smith serve` is a JSON-RPC 2.0 server over stdio, with no network port. Queries answer with
the JSON of their `--json` command, so existing contracts do not change; `run`, `install`,
`add`, `doctor.fix` and the like are runs that stream events as `event` notifications;
`cancel` stops a run and its processes. Fields are only ever added, as optional ones.

## Consequences

Release became a streamed run, with config and recipe methods, in
[#387](https://github.com/monosmith-app/monosmith/pull/387). Not built: `smith lsp`
(planned, [#303](https://github.com/monosmith-app/monosmith/issues/303)), and `smith mcp`
as a view on the server (it still runs its own process).

## Sources

- Issue [#360](https://github.com/monosmith-app/monosmith/issues/360); pull requests
  [#385](https://github.com/monosmith-app/monosmith/pull/385),
  [#387](https://github.com/monosmith-app/monosmith/pull/387)
- [Serve and the event stream](../../reference/serve.md), [`smith serve`](../../manual/serve.md)
