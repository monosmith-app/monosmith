# Serve and the event stream

> Terms such as project, task and cache key are defined in the [glossary](glossary.md). Design: [a rich terminal interface and the IDE plugins](../architecture/tui-and-ide.md). For editors, see the [editors guide](../guides/editors.md#the-engine-over-json-rpc-smith-serve).

Everything that takes time is a **run**, and a run tells what it does as
**events**. The same events come out two ways, so a terminal interface, an
editor plugin and a script all show the same facts:

- `smith run <task> --json-events` prints them as JSON lines on stdout.
- `smith serve` sends them as notifications over JSON-RPC.

The existing `--json` outputs do not change.

## The event stream

One event is one JSON object. Every event has `event`, `run`, `seq` and `time`;
the other fields depend on the kind and are absent otherwise.

| Event | Fields | When |
|---|---|---|
| `run.started` | `command`, `tasks` | the run begins: the smith command line it stands for, and every task it will consider, in dependency order |
| `task.queued` | `task` | once per task, right after `run.started` |
| `task.started` | `task` | the task got a slot; it is about to run, or to be replayed from the cache |
| `task.output` | `task`, `stream`, `line` | one line the task wrote, without its newline; a cached task replays its log the same way |
| `task.finished` | `task`, `status`, `durationMs`, `cacheKey`, `reason` | the task ended |
| `run.finished` | `ok`, `summary` | the run ended; the last event |

- `run` identifies the run. `seq` counts that run's events from 1, so a client
  can order events and notice a gap. `time` is UTC, RFC 3339.
- `task` is the task id, `<project>:<task>` (`web:build`).
- `status` is `passed`, `failed`, `cached` or `skipped`. A skipped task has a
  `reason` (`because core:build did not succeed`, `cancelled`, `after an
  earlier failure (--bail)`); so does a failed one (`exited with code 3`). A
  task that is skipped is queued and finished, and never started.
- `durationMs` is how long the task ran (0 for a cached or skipped one).
  `cacheKey` is the task's cache key, the same one `smith why-miss` explains.
- `stream` is `stdout`, `stderr` or `combined`. A task's two streams arrive
  merged, as a terminal shows them, and are `combined`; a command run by
  `smith serve` (below) keeps them apart.
- `summary` counts the run: `passed`, `failed`, `cached`, `skipped`,
  `durationMs`. `ok` is true when nothing failed or was skipped.
- A run with nothing to do (no project has the task) is still a run:
  `run.started` and `run.finished`, no tasks.

```json
{"event":"run.started","run":"run-20261020T060558-1","seq":1,"time":"2026-10-20T06:05:58Z","command":"smith run typecheck -p eslint-config","tasks":["eslint-config:typecheck"]}
{"event":"task.queued","run":"run-20261020T060558-1","seq":2,"time":"2026-10-20T06:05:58Z","task":"eslint-config:typecheck"}
{"event":"task.started","run":"run-20261020T060558-1","seq":3,"time":"2026-10-20T06:05:58Z","task":"eslint-config:typecheck"}
{"event":"task.output","run":"run-20261020T060558-1","seq":4,"time":"2026-10-20T06:05:59Z","task":"eslint-config:typecheck","stream":"combined","line":"checking"}
{"event":"task.finished","run":"run-20261020T060558-1","seq":5,"time":"2026-10-20T06:05:59Z","task":"eslint-config:typecheck","status":"passed","durationMs":731,"cacheKey":"ebfd6c4d..."}
{"event":"run.finished","run":"run-20261020T060558-1","seq":6,"time":"2026-10-20T06:05:59Z","ok":true,"summary":{"passed":1,"failed":0,"cached":0,"skipped":0,"durationMs":776}}
```

With `--json-events` stdout carries events and nothing else; errors before the
run starts (no such project) go to stderr and the exit code. The exit code is
the run's: 0 when `ok`, 1 otherwise. Ctrl-C stops the run's processes.

**Compatibility.** The shape has a version (`eventVersion` in `initialize`,
currently 1). Fields are only ever added, as optional ones, so ignore what you
do not know.

## `smith serve`

```sh
smith serve
```

JSON-RPC 2.0 over stdin and stdout, one JSON message per line (the framing
`smith mcp` uses). It serves the workspace it is started in, opens no network
port, and ends when stdin closes or the client sends `shutdown`. Anything
still running is then cancelled and its processes stopped. smith's own logging
goes to stderr, never to stdout.

Requests are answered concurrently, so a `cancel` is heard while a `run` is in
flight; match answers to requests by `id`. A request without an `id` is a
notification and gets no answer. Errors use the JSON-RPC codes: `-32700` parse
error, `-32601` method not found, `-32602` invalid params, and `-32000` for a
function that failed (the message says why).

### Events

While a run goes, the server sends notifications with the method `event`, whose
`params` is the event above:

```json
{"jsonrpc":"2.0","method":"event","params":{"event":"task.started","run":"r1","seq":3,"time":"...","task":"web:build"}}
```

All of a run's events arrive before the answer to the request that started it.

### Methods

| Method | Params | Answers |
|---|---|---|
| `initialize` | none | `name`, `version`, `protocolVersion` (1), `eventVersion` (1), `workspace`, `methods` |
| `ping` | none | `{}` |
| `shutdown` | none | `{}`, then the server ends |
| `run` | `tasks` (required), `projects`, `affected`, `base`, `parallel`, `bail`, `noCache`, `run` (the id to give the run) | `run`, `ok`, `summary` (what `smith run --json` prints) |
| `cancel` | `run` | `{"cancelled": true}`, or false when no such run is in flight |

`run` runs in the server's process, from the cache, and answers when the run
ends; a failed run is `ok: false`, not an error. Cancelling stops the task
processes and their whole trees; the unfinished tasks end `failed` with the
reason `cancelled`, or `skipped` if they had not started.

The other methods stand for a smith command. Their params are the input of
the MCP tool they map to (see the [editors guide](../guides/editors.md#ai-assistants-mcp)
for each tool's inputs), so a value is one argument and one that looks like an
option is refused.

| Method | Command | Answers |
|---|---|---|
| `info` | `smith info --json` | that JSON |
| `projects` | `smith projects --json` | that JSON |
| `graph` | `smith graph --format json` (`focus`) | that JSON |
| `affected` | `smith affected --json` (`base`) | that JSON |
| `kinds` | `smith kinds --json` | that JSON |
| `doctor` | `smith doctor --json` | that JSON, also when it reports failures |
| `doctor.fixes` | `smith doctor --fix` | the proposed fixes |
| `release.preview` | `smith release --dry-run --json` | that JSON |
| `updates` | `smith up --json` | that JSON |
| `upgrade.preview` | `smith upgrade --dry-run --json` | that JSON |
| `explain` | `smith explain <target> --json` (`target`, `recipe`) | that JSON |
| `recipe.check` | `smith recipe check --json` | that JSON |
| `plugins` | `smith plugin list --json` | that JSON |
| `kind.preview` | `smith kind <project> <kind> --dry-run --json` (`project`, `kind`) | that JSON |

A query answers with the JSON its command prints, exactly as `--json` does, so
a client and a script read the same contract.

The methods that change the workspace are runs of their own:

| Method | Command |
|---|---|
| `install`, `uninstall` | `smith install`, `smith uninstall` |
| `add`, `remove` | `smith add`, `smith remove` |
| `kind` (`project`, `kind`) | `smith kind <project> <kind> --json`; a refusal is a failed run, and `--force` is not offered |
| `upgrade` | `smith upgrade` |
| `doctor.fix` | `smith doctor --fix --yes` |
| `sync` | `smith sync` |
| `release` (`noPush`) | `smith release --yes [--no-push]` |
| `config.set` (`setting`, `value` or `values`) | `smith config set <setting> <value...>` |
| `config.unset` (`setting`) | `smith config unset <setting>` |

`release` publishes, so it has no MCP tool: an assistant is not handed it. A
client that calls it has shown the plan (`release.preview`) and asked the
person, which is why the command is run with `--yes`. Its output, with the
registry checks and each publish, arrives as `task.output`, and a failed
release ends the run `failed` with `exitCode` 1 and the journal in
`dist/release/` as `smith release --status` reads it. `config.set` takes
`value` for a setting with one value and `values` for a list, and checks them
against the schema like the command line; a refused value is a failed run with
the reason in its output.

Such a run has one task, named after the method (`install`), and the events
`run.started` (with the command line in `command`), `task.queued`,
`task.started`, a `task.output` for each line of the command's stdout or stderr,
`task.finished` and `run.finished`. It answers `run`, `ok` and `exitCode`, and
`cancel` stops it. The `run` param names the run, as for `run`.

Settings and recipes are read by the server itself, in the workspace it serves,
and answered at once:

| Method | Params | Answers |
|---|---|---|
| `config.list` | none | what `smith config list --json` prints: every setting with its value and what it does |
| `config.get` | `setting` (`ci.devcontainer`) | `setting`, `value` (as written in `smith.toml`), `set` (false: it has its default), `type`, `description` |
| `recipe.list` | none | `recipes`: each `*.recipe.toml` of the workspace (not under `node_modules`, `dist` or a hidden directory) with `path`, `name`, `version`, `extends`, and an `error` when it is not a recipe |
| `recipe.describe` | `recipe` (a file in the workspace, or an https URL) | the recipe as written, before any answer: `chain`, every `choices` entry (with `when` when it is asked only for an answer, and its file and line), the roles it fills as `use` (`plugins` always, `alternatives` by answer, such as `language=csharp`) and the plugins it `configures` |
| `recipe.roles` | `recipe`, `choices` (`{"language": "csharp"}`) | what the recipe puts in each role once the choices are answered (the rest take their defaults): `choices` as answered, and `roles`, each a level (`language`, `deploy`) or a slot (`lint`) with its `plugins` and the file and line they came from; an unanswered choice without a default is an invalid-params error |

A recipe file must be inside the workspace. `recipe.check` still validates a
recipe against the installed plugins, and `explain` with `recipe` shows what
the recipe decided for a project.

Not yet served: recipe creation (`smith recipe new`), and `smith mcp` as a view
on this server: it still runs its own process, with the same functions behind
it. A `smith lsp` command (a language server for `smith.toml` and recipes,
issue #303) is planned and does not exist in 1.62.0.

### A client in a few lines

```sh
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize"}' \
  '{"jsonrpc":"2.0","id":2,"method":"run","params":{"tasks":["build"],"projects":["web"]}}' \
  | smith serve
```

(Closing stdin right away ends the server and cancels the run; a real client
keeps the pipe open until the answer to `run` arrives.)
