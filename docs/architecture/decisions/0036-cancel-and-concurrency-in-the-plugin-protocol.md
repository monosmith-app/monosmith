# ADR 0036: Cancel and a concurrency declaration in the plugin protocol

- Status: accepted
- Date: 2026-10-25

## Context

The plugin protocol had `shutdown` and nothing else for a call that must end early, and no way for
a plugin to say how many calls it can take at once. The decision to settle both before the Python
and C# SDKs are generated from the schema is on issue
[#564](https://github.com/monosmith-app/monosmith/issues/564). What the code did when this was
decided:

- **The host** (`packages/plugin-sdk/host.Client`) held a mutex around every call, so a plugin
  never had two requests in flight. A call that ran out of time killed the plugin process, the
  only way to stop it. A cancelled command (Ctrl+C, `--bail`) reached a plugin only because the
  terminal sends the interrupt to the whole process group; nothing in the protocol did.
- **The Go serve loop** (`serve.Serve`) read one message, ran its handler, wrote the answer and
  only then read the next. It could not have read a cancel while a handler ran, and the handlers
  take no context. A host that pipelined requests was answered in order.
- **`plugin.json`** only says how to start a plugin (`name`, `version`, `command`); the plugin
  describes itself in the `initialize` result. A declaration in the manifest would be a second
  source of truth that has to be read before the process starts.
- **`taskrun`** never calls a plugin; tasks are processes smith starts itself, and it already kills
  their trees on cancel. The plugin calls that run long are `generate`, `publish`, `version/write`
  and the dependency methods (20 to 30 minutes), made from `release`, `nativebuild`, `scaffold` and
  `dependencies`.
- Every plugin built so far, the built-in ones, the examples and the Python and Node plugins that
  write the framing by hand, is strictly serial and knows nothing of either feature.

## Options considered

**Raise `protocol.Version` to 2.** A host would have to speak `[1 2]` and a plugin `[1 2]`
(`serve.initialize` already takes a list), and every plugin written for version 1 would need a
release to be accepted. Nothing here breaks a version 1 peer: a receiver ignores members it does
not know, in both directions, which is the rule the protocol already states, and a request the
peer never sends cannot hurt it. A bump would cost every plugin author and buy nothing, so it was
not chosen.

**A `$/cancelRequest` notification as in the Language Server Protocol.** The framing is LSP's, but
the prefix marks protocol-level messages that a server may drop, and this protocol has no other
such message. A plain method name matches the others and the issue.

**Declare concurrency in `plugin.json`.** Rejected as above: the `initialize` result is where a
plugin already says what it is, it works for a plugin that is not installed through the search
path, and the conformance suite sees it.

**Cancel by closing the process.** What happens today on a timeout. It also ends every other call
in flight to that plugin, and a plugin killed in the middle of `publish` leaves a half-published
release; the new protocol makes it the last resort, not the first.

**An unbounded pool in the serve loop (a goroutine per request).** A host that breaks the
declared limit would then run the plugin out of memory; the pool is bounded.

## Decision

The protocol version stays **1**. Both features are additive and negotiated by what the plugin
says in its `initialize` result.

**Cancel.**

- A notification, `cancel`, with params `{"id": <id of an in-flight request>}`. It has no response.
  The host sends it only to a plugin that advertised the capability `cancel`.
- A plugin that supports it stops the work of that request and answers the original request
  either with its result, if the work was already done, or with an error, code **`-32800`**
  (`CodeRequestCancelled`, the Language Server Protocol's number). An id the plugin does not know,
  or whose request is already answered, is ignored. A plugin that did not advertise `cancel`
  ignores the notification.
- In the Go SDK a plugin advertises `cancel` by implementing `serve.Cancellable`, one method,
  `WithContext(ctx) serve.Plugin`, which returns the plugin as it serves one request; every handler
  interface stays as it was. `Serve` calls it for each request, cancels the context when the
  cancel arrives (and at shutdown or the end of the input), and, when a cancelled request fails with
  any error, answers `-32800` instead of `-32000`. In the TypeScript SDK a plugin says
  `cancellable: true`, and every method receives an `AbortSignal` as its second argument. Anything
  else keeps the old behaviour, signal included: it never aborts.
- A request that had not started when its cancel arrived is answered `-32800` at once and never
  runs.

**What the host does** (`host.Client.CallContext`; the TypeScript client takes `signal` in the call
options):

| Event | Plugin advertised `cancel` | Plugin did not |
|---|---|---|
| the context ends (Ctrl+C, a cancelled command) | send `cancel`, wait the grace period (5 s, `host.Options.CancelGrace`) for the answer; return it, and kill the process only if none comes | wait the grace period for the answer, then kill the process |
| the call's timeout runs out | the same | kill the process at once (what it always did) |
| the answer arrives within the grace period | a result is the call's result; an error is returned wrapping both the context error and the plugin's `-32800` | the same |

Killing a plugin fails every other call in flight to it. A cancelled call returns an error for
which `errors.Is(err, context.Canceled)` (or `context.DeadlineExceeded`) holds. In smith the
context is wired into the calls that run long: `release` (version, publish, deploy, the GitHub
release assets) and the native builds; the other calls keep `Call` and, as before, end with the
command. Built-in plugins stay non-cancellable.

**Concurrency.**

- A member of the `initialize` result, `concurrency`, an integer: absent or `0` means **1** (the
  default of every plugin that says nothing: the host sends the next request after the previous is
  answered), `N` above 1 means up to N requests in flight, `-1` means no limit declared. Any other
  negative value is invalid and read as 1; the conformance suite reports it.
- The host never has more requests in flight to a plugin than it declared. The Go client waits for a
  free place before it sends (a call that waits does not count its timeout, and is not sent if its
  context ends first); for `-1` the host keeps at most 32 in flight (`host.UnlimitedInFlight`).
  `shutdown` is the one request that does not wait, so it reaches a busy plugin. With the default,
  calls to one plugin are serialised as they always were, now by the declaration, not by a mutex.
- Responses may come in any order; a response carries the id of its request. A plugin starts
  requests in the order they arrive. The Go `Serve` runs them on a bounded pool of that many
  workers (1024 for `-1`), keeps up to 1024 more waiting in order and refuses the rest with `-32600`
  (as it refuses an id that is already in flight). Nothing is written on the stream but whole
  frames, so answers do not interleave.
- `shutdown` first cancels what is in flight when the plugin is cancellable, then waits for the
  workers, then answers.

**Conformance.** `conformance.Check` gains, in both SDKs, under the same names: a declared
concurrency is one the protocol defines; a plugin that declares nothing answers three requests
sent together while the host keeps one in flight; one that declares N accepts min(N, 4) at once; a
plugin that advertises cancel answers a cancelled request with its result or `-32800` within a bound
and ignores a cancel for a request it does not know. `smith plugin check` runs them.

## Consequences

- A plugin written before this change works with the new host: it declares no concurrency, so it
  is called one request at a time, and it has no `cancel` capability, so it is never sent one. This
  was checked against smith 1.49.0 as the plugin (`smith plugin serve go`) and against the Node and
  Python examples that write the protocol by hand. A new plugin works with an older host: the host
  ignores members it does not know and never sends the notification.
- The Go SDK's `serve.Serve` is no longer a loop that handles one message at a time. A plugin
  whose `Describe` or handlers kept state without locks keeps working with the default (one handler
  at a time), but the plugin that declares a concurrency above 1 owns its locking.
- The two examples `structure-scan-go` and `structure-scan-node` show both features and are
  compared by the cross-language test, which sends them the same cancel and gets the same bytes.
- The Python and C# SDKs are generated from a schema that now has the `cancel` method (a row with
  `notification: true` and no result), the `cancel` capability, `concurrency` and the error code.
- Not done: the built-in plugins do not support `cancel`, because their handlers run toolchains
  through helpers that take no context; they should, and each is a separate change. The other
  `host.Client.Call` users in smith (`dependencies`, `scaffold`, `doctor`, `ci`) still end with the
  command. A plugin cannot ask the host to cancel anything, and there is no progress message.

## Sources

- Issue [#564](https://github.com/monosmith-app/monosmith/issues/564), the quirks list in its comments
- [The plugin protocol](../plugin-protocol.md), [the protocol cheat sheet](../../plugins/creating/protocol-cheat-sheet.md)
- `packages/plugin-sdk/schema/protocol.schema.json`; `packages/plugin-sdk/serve`, `host`, `conformance`
- [ADR 0002](0002-out-of-process-plugins.md)
