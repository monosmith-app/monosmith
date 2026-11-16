# ADR 0103: Built-in plugins run in the smith process

- Status: accepted
- Date: 2026-11-16
- Amends: [ADR 0002](0002-out-of-process-plugins.md) (for the built-in plugins)

## Context

Every smith session started each built-in plugin as its own process, `smith plugin serve <name>`:
39 processes per session, whatever languages the workspace holds. A machine with `smith mcp`, the
VS Code extension's `smith serve` and one command running had about 120. Each one held the smith
program open, so on Windows `smith update` could not replace an installed version while any of
them ran.

## Decision

1. A built-in plugin (`pluginhost.Builtins` and `pluginhost.BuiltinAddons`) is served inside the
   smith process, over two in-memory pipes, with the same protocol (`serve.Serve` on one side,
   `host.Connect` on the other). A session is one process.
2. An external plugin, and a built-in replaced by one in `[plugins.external.<name>]`, still runs
   as its own process. `smith plugin serve <name>` stays for `smith plugin check` and for tests.
3. A panic while a plugin handles a request answers that request with an internal error instead
   of ending the plugin, so a built-in cannot end smith. Out-of-process plugins get the same.
4. The tools a built-in plugin runs read the host's colour decision (`SMITH_TOOL_COLOR`, set in the
   process when the person has not set it) and write their output to the session's log
   (`toolrun.Log`).

## Consequences

- One process per session, a faster start and nothing for `smith update` to wait for.
- Built-in plugins share the process: they must not change the working folder or exit the
  process, and none does. A request that hangs can no longer be ended by killing its process; the
  host still stops waiting for it after the call's timeout.
