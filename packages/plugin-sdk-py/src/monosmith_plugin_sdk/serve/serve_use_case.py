"""The serve loop: a plugin on a pair of streams."""

from __future__ import annotations

import json
import sys
from dataclasses import dataclass
from typing import Any, NoReturn

from ..codec import (
    MessageParseError,
    ParamsError,
    ReceivedRequest,
    decode_params,
    parse_request,
    write_failure,
    write_result,
)
from ..concurrency import worker_limit
from ..framing import FrameReader
from ..protocol import ERROR_CODES, CancelParams
from .capabilities_of_algorithm import half_implemented, is_cancellable
from .claim_stdio_use_case import claim_stdio
from .dispatch_request_use_case import dispatch_request, message_of
from .frame_writer_store import FrameWriter
from .plugin_contract import Plugin
from .protocol_streams_contract import ByteSource, ServeStreams
from .request_outcome_contract import RequestOutcome
from .request_pool_store import PooledRequest, RequestPool

READ_SIZE = 64 * 1024


def serve(plugin: Plugin) -> NoReturn:
    """Run a plugin on the process's standard input and output, and exit when the host sends
    `shutdown` or closes the stream. Use it as the whole body of the plugin's entry point:

        if __name__ == "__main__":
            serve(MyPlugin())

    Standard output carries only the protocol's frames: whatever else the plugin or a library
    writes there (a `print`, a child process) goes to standard error instead. The process exits
    with status 0 after `shutdown`, and with 1, after the reason is printed, when the stream
    fails.
    """
    streams = claim_stdio()
    try:
        serve_streams(plugin, streams)
    except Exception as error:
        print(message_of(error), file=sys.stderr, flush=True)
        sys.exit(1)
    sys.exit(0)


@dataclass
class _Session:
    """What the loop remembers between requests."""

    plugin: Any
    writer: FrameWriter
    pool: RequestPool
    cancellable: bool
    initialized: bool = False

    def send(self, body: str) -> None:
        self.writer.send(body)

    def answer(self, request_id: int, outcome: RequestOutcome) -> None:
        if outcome.result_json is not None:
            self.send(write_result(request_id, outcome.result_json))
        else:
            self.send(write_failure(request_id, outcome.code, outcome.message))


def serve_streams(plugin: Any, streams: ServeStreams) -> None:
    """Answer requests from `streams.input` on `streams.output` until the host sends `shutdown` or the
    input ends, then return. This is the loop `serve` runs, on streams you choose: a test serves a
    plugin in-process over pipes with it.

    It has the semantics of the Go SDK's `serve.Serve`. Requests are answered by a bounded pool of
    worker threads, as many at a time as the plugin declared in `concurrency` (one, in order, when it
    said nothing), and the answers may be written in any order. A request before `initialize` is
    refused (`INVALID_REQUEST`); a notification (no `id`) gets no answer; an unreadable frame body is
    answered with `PARSE_ERROR` and the loop goes on; an unreadable header ends it with a
    `FramingError`. A `cancellable` plugin has the token of a request cancelled when the host sends
    `cancel` for it, and at the end of the input. `shutdown` is answered after everything read before it.

    Raises the first error writing a response (the host is gone) or a `FramingError`.
    """
    log = streams.log or (lambda message: print(message, file=sys.stderr, flush=True))
    for member in half_implemented(plugin):
        log(
            f"plugin: {member} needs its pair (install with uninstall, declared with rewrite) "
            "to be advertised; it will not be called"
        )
    writer = FrameWriter(streams.output)
    pool = RequestPool(lambda pooled: _run(session, pooled))
    session = _Session(plugin=plugin, writer=writer, pool=pool, cancellable=is_cancellable(plugin))
    reader = FrameReader()
    try:
        while True:
            chunk = _read_chunk(streams.input)
            if not chunk:
                break
            for body in reader.push(chunk):
                if pool.failure is not None:
                    raise pool.failure
                if _receive(session, body):
                    return
    finally:
        _drain(session)
    if pool.failure is not None:
        raise pool.failure


def _read_chunk(source: ByteSource) -> bytes:
    read1 = getattr(source, "read1", None)
    return bytes((read1 or source.read)(READ_SIZE))


def _receive(session: _Session, body: str) -> bool:
    """Handle one frame. Returns whether it was `shutdown`, after which the loop ends."""
    try:
        request = parse_request(body)
    except MessageParseError as error:
        session.send(write_failure(None, ERROR_CODES.PARSE_ERROR, str(error)))
        return False
    if request.id is None:
        _notified(session, request)
        return False
    request_id = request.id
    if not session.initialized and request.method != "initialize":
        session.send(write_failure(request_id, ERROR_CODES.INVALID_REQUEST, "initialize must come first"))
        return False
    if request.method == "initialize":
        outcome = dispatch_request(session.plugin, request)
        if outcome.ok:
            session.initialized = True
            session.pool.set_limit(worker_limit(_declared_concurrency(outcome)))
        session.answer(request_id, outcome)
        return False
    if request.method == "shutdown":
        # Everything read before shutdown, started or still waiting, is answered with its result first,
        # as in the Go SDK; only a cancel message stops a request.
        session.pool.idle()
        session.send(write_result(request_id, "{}"))
        return True
    admission = session.pool.submit(request)
    if admission == "duplicate":
        session.send(
            write_failure(request_id, ERROR_CODES.INVALID_REQUEST, f"request id {request_id} is already in flight")
        )
    elif admission == "full":
        session.send(
            write_failure(
                request_id,
                ERROR_CODES.INVALID_REQUEST,
                "too many requests in flight: the host must not send more than the plugin's declared concurrency",
            )
        )
    return False


def _declared_concurrency(outcome: RequestOutcome) -> int | None:
    """The `concurrency` member of an `initialize` result that was already written as JSON."""
    value = json.loads(outcome.result_json or "{}").get("concurrency")
    return value if isinstance(value, int) and not isinstance(value, bool) else None


def _notified(session: _Session, request: ReceivedRequest) -> None:
    """Handle a message without an `id`.

    `cancel` is the only one the protocol defines for a plugin; every other is ignored, and so is a
    `cancel` to a plugin that is not `cancellable`, or one that names no request.
    """
    if request.method != "cancel" or not session.cancellable:
        return
    try:
        params: CancelParams = decode_params("cancel", request.params)
    except ParamsError:
        return
    waiting = session.pool.cancel(params["id"])
    if waiting is not None:
        _answer_cancelled(session, waiting)


def _run(session: _Session, pooled: PooledRequest) -> None:
    """A worker's job: handle one request and answer it."""
    try:
        outcome = dispatch_request(session.plugin, pooled.request, pooled.token)
    except BaseException as error:  # a SystemExit or the like in a method: answered, not lost
        outcome = RequestOutcome(code=ERROR_CODES.PLUGIN_FAILURE, message=message_of(error))
    if not outcome.ok and outcome.code == ERROR_CODES.PLUGIN_FAILURE and pooled.cancelled:
        outcome = RequestOutcome(code=ERROR_CODES.REQUEST_CANCELLED, message=f"request cancelled: {outcome.message}")
    session.pool.release(pooled)
    session.answer(pooled.id, outcome)


def _answer_cancelled(session: _Session, pooled: PooledRequest) -> None:
    session.send(write_failure(pooled.id, ERROR_CODES.REQUEST_CANCELLED, "request cancelled before it started"))


def _drain(session: _Session) -> None:
    """End the work in flight when the loop returns (the input ended, the host is gone).

    A cancellable plugin is asked to stop and what had not started is refused; any other finishes
    what it is doing.
    """
    if session.cancellable:
        for waiting in session.pool.cancel_all():
            try:
                _answer_cancelled(session, waiting)
            except OSError:
                break
    session.pool.idle()
