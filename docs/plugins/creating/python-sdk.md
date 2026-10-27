# The Python SDK

`monosmith-plugin-sdk` is the SDK for writing smith plugins in Python. It does what the [Go
SDK](plugin-sdks.md#go-with-the-sdk) and the [TypeScript SDK](typescript-sdk.md) do: it frames the
messages, answers `initialize`, derives the capabilities from the methods your plugin has, gives you a
host to test it with and the conformance checks `smith plugin check` runs, and ports the Go
`templating` and `jsondoc` packages. It has no runtime dependencies and is fully type-annotated. This
page walks through a plugin, then lists the whole API.

```sh
pip install monosmith-plugin-sdk
```

```python
import monosmith_plugin_sdk
```

It needs Python 3.10 or later. The package is built from
[`packages/plugin-sdk-py`](../../../packages/plugin-sdk-py/README.md) in this repository and released
by smith like the other packages. The distribution is `monosmith-plugin-sdk`; the import name is
`monosmith_plugin_sdk`.

## Where the types come from

The Go types in `packages/plugin-sdk/protocol` are the source of truth of the protocol. Their JSON
Schema is committed as
[`packages/plugin-sdk/schema/protocol.schema.json`](../../../packages/plugin-sdk/schema/protocol.schema.json):
one definition per message, a methods table (each of the 18 methods with its params, its result and the
capability it needs), the error codes, the roles and the protocol version. The Python types are
generated from that file by `tools/generate_protocol`, so they cannot drift from the Go ones: a Go test
fails when the schema is stale, a Python test fails when the generated modules are, and a test fails
when the schema has a method the SDK cannot serve.

Every message is a `TypedDict`: the JSON object as a plain `dict`. A plugin returns
`{"projects": [...]}`, and a type checker such as mypy knows the members. A member is required in a
type exactly when the Go SDK writes it without `omitempty`, so `Finding` needs `check` and `ok` and
nothing else. An enumeration (`Role`, `ProjectKind`, `PluginKind`) is a `Literal` alias and is **not
validated at run time**: a value the schema does not list, from a newer host, passes through.

The same file is the contract for a language without an SDK. [The protocol cheat
sheet](protocol-cheat-sheet.md) shows it as tables.

## A plugin, step by step

The structure plugin `widget-layout` of the [structure plugin guide](structure-plugin.md), in Python:
it adds a `layout-check` task to widget projects, makes their `lint` wait for it, and warns about a
project without a README. The finished program is
[`structure-layout-py`](../../examples/plugins/structure-layout-py).

### 1. Describe it

A plugin is an object. `describe` is the one method every plugin has; it answers `initialize` with the
plugin's identity. The SDK adds the protocol version and the capabilities.

```python
from monosmith_plugin_sdk import InitializeParams, PluginDescription, serve


class LayoutPlugin:
    def __init__(self) -> None:
        self.workspace_root = ""

    def describe(self, params: InitializeParams) -> PluginDescription:
        self.workspace_root = params["workspaceRoot"]  # absolute; the paths in later requests are relative to it
        return {
            "plugin": {"name": "widget-layout", "version": "0.1.0"},
            "kind": "convention",
            "roles": ["structure"],
            "languages": ["widget"],  # a convention that names no language applies to every language
        }


if __name__ == "__main__":
    serve(LayoutPlugin())
```

`serve` runs the protocol on standard input and output and ends the process when smith sends
`shutdown`. Standard output carries only the protocol's frames, and the SDK enforces that at the level of
the file descriptors: from the moment `serve` starts, file descriptor 1 and `sys.stdout` are standard
error, so a `print` anywhere in the program, a library's write, a C extension or a child process that
inherits standard output reaches smith's log and cannot corrupt the stream. File descriptor 0 is the
null device, so a child process that reads standard input cannot take smith's requests. Log with
`print(..., file=sys.stderr)` all the same.

### 2. Add the methods the plugin has

Each method the plugin object has is a capability the SDK advertises, and each it lacks is answered with
`-32601`, `method not supported`, as in Go. The methods are plain functions and may block (see
[Threads](#threads-and-why-not-asyncio)). Their names are snake_case: the protocol's `version/write` is
`write_version`.

```python
from pathlib import Path
from typing import Any

from monosmith_plugin_sdk import contribute_to, task


class LayoutPlugin:
    ...

    def contribute(self, params: Any) -> Any:
        return contribute_to(
            params["projects"],
            lambda project: {
                "tasks": {"layout-check": task([sys.executable, CHECKER, "layout-check"], cache=True)},
                "dependsOn": {"lint": ["layout-check"]},
            },
        )

    def doctor(self, params: Any) -> Any:
        return {
            "findings": [
                {
                    "check": f"{project['name']} has a README.md",
                    "ok": Path(self.workspace_root, *project["root"].split("/"), "README.md").exists(),
                }
                for project in params["projects"]
            ]
        }
```

The params and results are typed from the schema (`DoctorParams`, `DoctorResult`, ...), so with
annotations `project["root"]` and the shape of a finding are checked by mypy as you type. The handshake of
this plugin advertises exactly what is implemented, `{"doctor": true, "contribute": true}`, and nothing
else.

What happens to what you return and raise:

| You | The host sees |
|---|---|
| return a result | the result, written with the members in the order of the schema and every empty optional member left out, exactly as the Go SDK writes it |
| return nothing (`None`) | `{}`: right for a result with no required member (`remove`), an error naming the member for the others |
| return a result with a required member missing, or of the wrong type | `-32603` (internal error) and a message that names the member: `doctor: result.findings[0].check is required` |
| raise an exception | `-32000` (plugin failure) and its message, which smith shows the user: put the remedy in it |
| are sent params of the wrong JSON type | `-32602` (invalid params) before your method is called |

Params reach your method shaped as the types promise: where the Go host writes a nil slice as `null`,
you get `[]`, a required member that is absent is its zero value, and members the SDK does not know are
kept. An optional member is simply absent when it was not sent: read it with `params.get("stack", "")`.

### Stop work when the host cancels it

A method that can run for long, `publish` or `doctor` over a big tree, should stop when smith gives up
on it (Ctrl+C, a timeout). Say the plugin is `cancellable`: the SDK then passes a `CancelToken` as the
**second argument** of every method that takes one, sets it when the host sends the `cancel`
notification for that request, and advertises the `cancel` capability. A method that stops raises (the
host sees the request-cancelled error `-32800`, whatever the message) or returns the result it already
has.

```python
from monosmith_plugin_sdk import CancelToken


class ScanPlugin:
    cancellable = True

    def describe(self, params):
        return {"plugin": {"name": "widget-scan", "version": "0.1.0"}, "kind": "convention",
                "roles": ["structure"], "languages": ["widget"], "concurrency": 4}

    def doctor(self, params, cancel: CancelToken):
        for project in params["projects"]:
            cancel.raise_if_cancelled()          # between units of work
            ...                                   # scan the project
        return {"findings": []}
```

A token is a flag one thread sets and another watches. `token.cancelled` says whether the work was asked
to stop, `token.raise_if_cancelled(message)` raises `RequestCancelledError`, and `token.wait(seconds)` is
a **cancellable sleep**: it returns `True` the moment the token is cancelled and `False` when the time
ran out, so `if cancel.wait(0.1): ...` pauses without the delay a `time.sleep` would add to a cancel. A
method that takes only the params (`def doctor(self, params)`) is called with only the params; the
token is passed to a method that accepts a second positional parameter (or `*args`).

`concurrency` in what `describe` returns says how many requests the plugin can work on at the same time.
Leave it out for one at a time, in order, which is what a plugin that says nothing gets; a number above
1 runs that many requests together, started in the order they arrive, with the answers written in the
order they finish. `-1` leaves the number to the host. A plugin without `cancellable = True` is never
sent a cancel and its token never fires. [`structure-scan-py`](../../examples/plugins/structure-scan-py)
is the whole example; its Go and TypeScript twins answer the same requests with the same bytes.

#### Threads, and why not asyncio

The serve loop reads requests on the main thread and runs them on a **bounded pool of worker threads**,
as many as `concurrency` allows, one writer for the frames. Threads were chosen over `asyncio` because
a plugin's methods mostly call blocking things (files, subprocesses, registries over `urllib`) that
`asyncio` would need `run_in_executor` for anyway, because a plain `def` is the simplest plugin to write
and test, and because cancellation is cooperative in both: a thread cannot be interrupted, so the
`CancelToken` is the mechanism, and the SDK needs no event loop of its own. If a method needs `asyncio`,
call `asyncio.run(...)` inside it. A plugin that declares a concurrency above 1 owns its locking: shared
state in the plugin object needs a `threading.Lock`.

### 3. Test it

`connect_in_process` serves the plugin object in your test process over pipes, on worker threads, with
the real framing, handshake and dispatch, and gives you the host side. `check_plugin` runs the checks of
`smith plugin check`.

```python
from monosmith_plugin_sdk import HostOptions, check_plugin, connect_in_process

with connect_in_process(LayoutPlugin(), HostOptions(workspace_root=str(tmp_path))) as client:
    report = check_plugin(client)
    assert report.ok, report.failures()

    result = client.call("contribute", {"projects": [{"name": "core", "root": "libs/core", "language": "widget", "kind": "lib"}]})
    assert result["projects"]["core"]["dependsOn"]["lint"] == ["layout-check"]
```

The example's tests, run with `python -m unittest -v` from its folder:

```text
test_a_method_the_plugin_lacks_is_32601 (test_layout.LayoutPluginTest.test_a_method_the_plugin_lacks_is_32601) ... ok
test_doctor_warns_about_a_missing_readme (test_layout.LayoutPluginTest.test_doctor_warns_about_a_missing_readme) ... ok
test_find_junk_drawers_skips_build_output (test_layout.LayoutPluginTest.test_find_junk_drawers_skips_build_output) ... ok
test_lint_waits_for_the_layout_check (test_layout.LayoutPluginTest.test_lint_waits_for_the_layout_check) ... ok
test_the_handshake_advertises_exactly_what_is_implemented (test_layout.LayoutPluginTest.test_the_handshake_advertises_exactly_what_is_implemented) ... ok
test_the_plugin_conforms (test_layout.LayoutPluginTest.test_the_plugin_conforms) ... ok
```

`client.call(method, params)` returns the decoded result. An answer that is an error raises an
`RpcError` that has the `code`. A call that is not answered within its timeout (two minutes unless
`timeout=` says otherwise, in seconds) kills the plugin process, because its stream can no longer be
trusted, and raises a `CallInterruptedError`. Calls may be made from several threads; the client never
has more in flight than the plugin declared, as smith does. To cancel one, pass a `CancelToken` and set
it from another thread:

```python
token = CancelToken()
threading.Timer(0.5, token.cancel).start()
try:
    client.call("doctor", {"projects": projects}, cancel=token)
except CallInterruptedError as error:
    assert error.reason == "cancelled" and not error.killed and error.response.code == -32800
```

To test the real process, with the launcher smith would use, start it instead:

```python
client = start_plugin(StartOptions(command=[sys.executable, "plugin.py"], workspace_root=str(tmp_path)))
```

Do not make a test depend on how long something takes. The SDK's own tests wait on events the test
controls (a `threading.Event` the plugin's method waits for, a latch the plugin sets when it has started),
never on a pause: a Linux runner under load is slower than your machine.

### 4. Run it under smith

```text
PS> smith plugin check .\plugin.cmd
widget-layout 0.1.0
  ✔ names itself
  ✔ has a known kind
  ✔ plays a known role: structure
  ✔ a structure plugin does something
  ✔ declares a concurrency the protocol defines
  ✔ answers 3 requests sent together, one at a time
  ✔ contribute answers with no projects
  ✔ doctor answers with no projects
  ✔ shuts down when asked
```

The lines are the ones `structure-layout-node` and `structure-layout-go` print. `plugin.cmd` is the
two-line launcher of [Starting an interpreted plugin](plugin-sdks.md#starting-an-interpreted-plugin)
(`python "%~dp0plugin.py" %*`). On Linux and macOS the script starts with `#!/usr/bin/env python3` and is
made executable: `smith plugin check ./plugin.py` (not run there). To check a script without a launcher,
give `smith plugin check` the absolute path of the interpreter that has the SDK and then the script, as the
SDK's own integration tests do. Declare it in `smith.toml`:

```toml
[plugins.external.widget-layout]
command = ["/path/to/venv/bin/python", "/path/to/plugin.py"]

[conventions.widget-layout]
```

The interpreter must be able to `import monosmith_plugin_sdk`: name the interpreter of a virtual
environment that has it installed, by its full path. smith starts an external plugin with a minimal
environment; see [Plugin security](../using/plugin-security.md).

### 5. A language plugin

[`language-widget-py`](../../examples/plugins/language-widget-py) is the language plugin of the
[language plugin guide](language-plugin.md): `discover`, `generate`, `write_version` and `doctor`. It
uses the helpers below: `build_task` for a cached `build` task that writes into `dist/`,
`scaffold_project` to write a new project from templates, and `JsonDocument` to change a `widget.json`
without reordering it. Its conformance run, with `smith plugin check python plugin.py --generate`, has the
same 27 checks as the Go and TypeScript plugins': it generates an app, a lib and a package and discovers
them again.

For the same requests a Python plugin, a TypeScript plugin and a Go plugin with the same logic give
**byte-identical answers**: the SDK writes the members in the order of the Go structs, leaves out what
`omitempty` leaves out, sorts the keys of maps and escapes `<`, `>`, `&` and U+2028/U+2029 as Go does,
while writing every other character as it is. The SDK's integration tests check it for
`language-widget-py`, `structure-layout-py` and `structure-scan-py` against their Go and TypeScript twins,
cancel included, and compare answers that may come out of order by their id.

## The plugin interfaces

`serve` takes any object with `describe` and any of these members, each `(params)` or `(params, cancel)`
and returning the result. The names follow the Go interfaces; `Plugin`, `Discoverer`, `Generator` and the
rest are `Protocol` classes that document the signatures, and a plugin does not have to inherit from them.

| Member | Protocol method | Capability | Go interface |
|---|---|---|---|
| `describe` | `initialize` | always | `Plugin.Describe` |
| `discover` | `discover` | `discover` | `Discoverer` |
| `generate` | `generate` | `generate` | `Generator` |
| `write_version` | `version/write` | `version` | `Versioner` |
| `publish` | `publish` | `publish` | `Publisher` |
| `published` | `published` | `published` | `RegistryReader` |
| `ci` | `ci` | `ci` | `CIContributor` |
| `doctor` | `doctor` | `doctor` | `DoctorContributor` |
| `remove` | `remove` | `remove` | `Remover` |
| `change_kind` | `project/kind` | `changeKind` | `KindChanger` |
| `contribute` | `contribute` | `contribute` | `Contributor` |
| `install` and `uninstall` | `dependencies/install`, `dependencies/uninstall` | `dependencies` | `DependencyManager` |
| `declared` and `rewrite` | `dependencies/declared`, `dependencies/rewrite` | `declarations` | `DeclarationEditor` |
| `latest` | `dependencies/latest` | `latest` | `LatestReader` |
| `cancellable = True` | the `cancel` notification | `cancel` | `Cancellable` |

A pair is advertised only when the plugin has both members, as in Go, where the interface needs both
methods; with one of them the SDK warns on standard error and does not call it. `initialize` and
`shutdown` are answered by the SDK. A request before `initialize` is refused with `-32600`, a
notification gets no answer, and an unreadable frame body is answered with `-32700`. `shutdown` is
answered after every request read before it has been answered with its result; when the input ends instead
(the host is gone), a cancellable plugin has what is in flight cancelled and what is waiting refused with
`-32800`, and any other finishes what it is doing.

## Reference

Everything is importable from `monosmith_plugin_sdk`. The messages (`Project`, `Task`, `DiscoverParams`,
`Finding`, and the rest) are `TypedDict`s with the descriptions of the Go doc comments; `PluginDescription`
is the `initialize` result without the two members the SDK fills in.

### Serving

| Name | What it is |
|---|---|
| `serve(plugin)` | runs the plugin on standard input and output and exits the process; the whole body of an entry point |
| `serve_streams(plugin, ServeStreams(input, output, log=None))` | the same loop on streams you give it (binary file objects); returns at `shutdown` or the end of the input |
| `capabilities_of(plugin)` | the `Capabilities` the plugin would advertise |
| `Plugin`, `Discoverer`, `Generator`, `Versioner`, `Publisher`, `RegistryReader`, `CIContributor`, `DoctorContributor`, `Remover`, `KindChanger`, `Contributor`, `DependencyManager`, `DeclarationEditor`, `LatestReader` | the shapes of the plugin object |
| `CancelToken` | `cancelled`, `cancel()`, `wait(timeout=None)`, `raise_if_cancelled(message?)`, `on_cancel(callback)` |
| `RequestCancelledError` | what `raise_if_cancelled` raises |

### Testing a plugin

| Name | What it is |
|---|---|
| `connect_in_process(plugin, HostOptions(...))` | serves a plugin object in this process over pipes and connects to it |
| `start_plugin(StartOptions(command=[...], workspace_root=...))` | starts a plugin process with `SMITH_PLUGIN=1`, performs the handshake, forwards its standard error (`stderr="ignore"` drops it) |
| `connect_plugin(PluginStreams(input, output), HostOptions(...))` | the handshake over streams you have |
| `PluginClient` | `info` (the `initialize` result), `call(method, params, timeout=None, cancel=None)`, `notify(method, params)`, `supports(method)`, `peak_in_flight`, `cancel_grace`, `close()`; a context manager |
| `CallInterruptedError` | a call given up on because its token was cancelled or its time ran out: `reason` (`"cancelled"` or `"timeout"`), `killed`, and `response` (the plugin's `-32800`, when it answered) |
| `check_plugin(client, workspace=None)` | the checks of `smith plugin check`; with `workspace` (the directory the plugin was started in) also `--generate` |
| `ConformanceReport` | `plugin`, `results`, `ok`, `failures()`, `to_dict()` (the shape of `smith plugin check --json`) |
| `RpcError` | an error answer, with its `code` and `message` |

`HostOptions` are `workspace_root` (required, absolute), `host` (name and version, default `smith` `dev`),
`initialize_timeout` (30 seconds) and `cancel_grace` (5 seconds: how long a plugin has to answer after it was
asked to stop); `StartOptions` adds `command`, `env`, `cwd` and `stderr`. All timeouts are seconds. A client
never has more calls in flight than the plugin declared in `concurrency`, as smith does: with the default
of one, calls are serialised. Cancel a call and the plugin is sent `cancel` if it advertised it; the call
ends with a `CallInterruptedError`, or with the result if the plugin had finished. A plugin that does not
answer within the grace period is killed. `check_plugin` is a port of the Go `conformance.Check` with the
same check names; `smith plugin check` adds one more, "shuts down when asked", which is `client.close()`
not raising. To run the real command from a test, spawn `smith plugin check` as
`tests/integration/test_smith_plugin_check.py` does.

### Helpers for the answers

| Name | What it is |
|---|---|
| `task(command, *, cwd, env, inputs, outputs, depends_on, cache, description, verify, serve, serve_url, serve_timeout)` | a `Task`: an argv, never a shell string; arguments left out are not written; `**extra` passes members the SDK does not know yet |
| `build_task(root, command, **options)` | a cached build task: `outputs` and the directory `command(output_dir)` is told to write into are `../../dist/<root>`, and it depends on `^build` |
| `contribute_to(projects, contribution)` | the answer to `contribute` when each project gets the same kind of addition; a project for which `contribution` returns `None` is left out |
| `scaffold_project(workspace_root, params, files, data=None, notes=None)` | the answer to `generate` for a scaffold: renders `files` (target path to template) with `scaffold_data(params)` (`Dir`, `Name`, `Kind`, `Stack`, `Workspace`, `Scope`, `ModulePrefix`) and `data`, never overwriting |
| `action_step(uses, inputs=None)`, `ci_step(name, command, *, dir, blocking, cache_files)`, `credential(env, description, *, required, setup, tasks)` | the pieces of a `ci` answer: a provider-native step with ordered inputs, a command a phase runs (blocking unless told not to), a secret the release needs |
| `project_build_dir(root)`, `build_dir_from_project(root)`, `in_build_dir(path)`, `BUILD_DIR` | where a project builds: `dist/apps/web`, and `../../dist/apps/web` from inside the project |
| `ROLES`, `is_known_role`, `is_exclusive_role`, `is_additive_role`, `default_roles(kind)`, `effective_roles(info)` | the roles, as the schema lists them |
| `requires_language(name)`, `conflicts_with(name)`, `is_empty_ref(ref)`, `describe_ref(ref)` | plugin references |
| `resolve_project_kind(declared, root, fallback)`, `project_kind_directory(kind)`, `publishes(project)`, `is_manifest(manifests, base)`, `set_toolchain(result, probe)`, `stack_plugins(stack)`, `audit_enabled(audit, policy)`, `audit_blocks(audit, policy)` | the small functions of the Go `protocol` and `projectkind` packages |

### Templating

`write_templates(root, files, data=None)` renders every template of `files` (target path to text) under
`root` with the `[[ ]]` delimiters and returns the paths it wrote. Target paths may contain actions
(`[[.Dir]]/main.py`). It checks every target before it writes anything: an existing file is a
`TemplateFileExistsError` (a `FileExistsError`) and nothing is written, a missing field is a `TemplateError`
(the Go `missingkey=error`), and a target outside `root` is refused. `render_template(text, data, name)`
renders one string. The data are plain `dict`s and lists.

The language is the subset of Go's `text/template` that project templates use. Anything else is reported
with the name of the construct, not rendered wrongly.

| Supported | Not supported |
|---|---|
| `[[.Name]]`, `[[.A.B]]`, `[[$.Name]]`, `[[.]]` | variables (`[[$x := ...]]`), `define`, `template`, `block`, `break`, `continue` |
| `[[if]]`, `[[else if]]`, `[[else]]`, `[[with]]`, `[[range]]`, `[[end]]` | methods on values |
| `eq ne lt le gt ge not and or len index print printf` (`%s %d %v %q %t %%`), pipes, parentheses | other functions |
| string, raw string, number, `true`, `false`, `nil` | |
| `[[- x -]]` trim markers, `[[/* comments */]]` | |

Text such as `${{ secrets.TOKEN }}` in a GitHub workflow passes through untouched, as in Go.

### Editing JSON

`JsonDocument` is the Go `jsondoc.Document`: a JSON object that keeps its key order and the text of
everything you do not touch (`1.50` stays `1.50`), so editing a `package.json` changes only what you
edited. `JsonDocument.parse(text)`, `has`, `keys()`, `len(document)`, `object(key, create=False)` (an object
you can edit in place), `get_string`, `get_value`, `set_string`, `set_json` (a `dict` becomes an editable
document), `delete`, and `render()`, which writes two-space indentation and a final newline as npm does. The
output is the same bytes as the Go SDK's for the same input. Write the result with
`path.write_bytes(document.render().encode())` or `newline=""`, so that Windows does not turn the line ends
into CRLF.

### Building blocks

For a tool that speaks the protocol itself: `encode_params`, `encode_result`, `decode_params`,
`decode_result` (the wire form of a message by method name), `FrameReader` and `frame_message` (the
`Content-Length` framing), `PROTOCOL_SCHEMA` (the schema as data), `PROTOCOL_VERSION`, `ERROR_CODES`
(`PARSE_ERROR`, `INVALID_REQUEST`, `METHOD_NOT_FOUND`, `INVALID_PARAMS`, `INTERNAL_ERROR`, `PLUGIN_FAILURE`,
`REQUEST_CANCELLED`) and `MAX_MESSAGE_BYTES`.

## What differs from the Go and TypeScript SDKs

| | Go | TypeScript | Python |
|---|---|---|---|
| a plugin | a struct with methods; `serve.Main(&plugin{})` | an object with members; `await serve(plugin)` | an object with methods; `serve(plugin)` |
| messages | structs | interfaces | `TypedDict`s: plain dicts |
| method names | `WriteVersion` | `writeVersion` | `write_version` |
| concurrency model | a goroutine pool | a promise pool on one thread | a pool of worker threads |
| cancel | `serve.Cancellable` (`WithContext(ctx)`) | `cancellable: true`, an `AbortSignal` as the second argument | `cancellable = True`, a `CancelToken` as the optional second argument |
| a long wait | `select` on `ctx.Done()` | `timers/promises` with the signal | `cancel.wait(seconds)` |
| standard output | the program's job | `process.stdout` is redirected | file descriptors 0 and 1 are redirected |
| error codes | an error becomes `-32000` | the same, plus `-32602` and `-32603` | the same as TypeScript |
| `null` in params | a nil slice or map | `[]` or `{}` for a required member, removed for an optional one | the same as TypeScript |
| the host's timeouts | `time.Duration` | milliseconds (`timeoutMs`) | seconds (`timeout`) |
| `toolrun` | `toolrun.Run` | not ported | not ported: use `subprocess.run(..., stdout=sys.stderr)` |

## Common mistakes

| You see | Cause and remedy |
|---|---|
| `malformed header "..."` in smith | something wrote to standard output before `serve` took it over, for example a `print` at module level or an import that prints: remove it, or print to standard error |
| `ModuleNotFoundError: monosmith_plugin_sdk` when smith starts the plugin | smith runs the `command` you declared: name the interpreter of the environment that has the SDK installed, not a bare `python` |
| `method not supported: x` for a method you wrote | a pair needs both members (`install` and `uninstall`; `declared` and `rewrite`), and the member names are the ones in the table above (`write_version`, not `version`) |
| `-32603` with `result.x is required` | a required member of the result is missing: the type says which |
| `plugin stopped answering: EOF` in a test | the plugin process died: its standard error is on yours, unless you passed `stderr="ignore"` |
| a test hangs | a plugin that never answers is killed after the call's timeout (two minutes); pass `timeout=` to fail sooner |
| the token never fires | the plugin is not `cancellable`, or the host did not send `cancel`: only a call made with a token that is cancelled (or whose time runs out) sends it |
| a cancelled method keeps running | a thread cannot be interrupted: the method must look at the token (`wait`, `raise_if_cancelled`) between steps, and pass it on to what it calls |
| data corrupted when two requests run together | with `concurrency` above 1 the methods run on several threads: guard state shared through the plugin object with a lock |
| CRLF in a file the plugin wrote on Windows | `Path.write_text` translates line ends; write bytes, or open the file with `newline=""` |

## See also

- [Writing a plugin in Go, Node.js or Python](plugin-sdks.md): what exists for each language, and the protocol by hand
- [The TypeScript SDK](typescript-sdk.md): the same API in TypeScript
- [The plugin protocol cheat sheet](protocol-cheat-sheet.md) and [the architecture of the protocol](../../architecture/plugin-protocol.md)
- [Testing a plugin](testing-a-plugin.md)
- [`packages/plugin-sdk-py`](../../../packages/plugin-sdk-py/README.md), the package
