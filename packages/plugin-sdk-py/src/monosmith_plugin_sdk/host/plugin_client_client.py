"""The host's side of a connection to one initialized plugin."""

from __future__ import annotations

import threading
import time
from collections import deque
from collections.abc import Callable
from dataclasses import dataclass, field
from typing import Any

from ..cancellation import CancelToken
from ..codec import (
    ReceivedResponse,
    decode_result,
    encode_params,
    parse_response,
    write_notification,
    write_request,
)
from ..concurrency import in_flight_limit
from ..framing import FrameReader, frame_message
from ..protocol import METHOD_CAPABILITIES, PROTOCOL_VERSION, InitializeResult, RpcError
from ..serve import ByteSource
from .call_interrupted_error import CallInterruptedError, InterruptionReason
from .host_options_contract import DEFAULT_CALL_TIMEOUT, DEFAULT_CANCEL_GRACE, HostOptions

READ_SIZE = 64 * 1024


@dataclass
class PluginConnection:
    """The two ends of the pipe to a plugin, and what to do with the plugin when the stream cannot be trusted."""

    input: ByteSource
    """The plugin's responses."""
    write: Callable[[bytes], None]
    """Writes bytes to the plugin's requests. Raises when the plugin is gone."""
    close_input: Callable[[], None]
    """Ends the plugin's input, which is how a plugin that does not shut down is told the host is gone."""
    kill: Callable[[], None]
    """Called when the plugin did not answer in time: ends the plugin, if it is a process."""
    finish: Callable[[float], str | None]
    """Waits up to the given seconds for the plugin to go; returns the reason it did not end well, or `None`."""
    dispose: Callable[[], None] = lambda: None
    """Releases what the connection holds, once nothing reads from it any more."""


@dataclass
class _Waiter:
    """A call waiting for its answer."""

    wake: threading.Event
    done: threading.Event = field(default_factory=threading.Event)
    body: str | None = None
    error: Exception | None = None

    def settle(self, body: str | None, error: Exception | None) -> None:
        self.body, self.error = body, error
        self.done.set()
        self.wake.set()


class PluginClient:
    """The host's side of a connection to one initialized plugin, for a plugin author's tests.

    smith itself is the host in real use. The client never has more requests in flight than the plugin
    declared in `concurrency`, as smith does: with the default of one, calls are serialised, each waiting
    for the one before it to be answered. Each call has a timeout and may be cancelled with a
    `CancelToken`. Calls may be made from several threads at once. Open one with `start_plugin` (a
    process), `connect_plugin` (streams) or `connect_in_process` (a plugin object served in this process).
    """

    @classmethod
    def open(cls, connection: PluginConnection, options: HostOptions) -> PluginClient:
        """Perform the handshake on a connection."""
        client = cls(connection, options.cancel_grace)
        try:
            info: InitializeResult = client.call(
                "initialize",
                {
                    "protocolVersions": [PROTOCOL_VERSION],
                    "host": options.host or {"name": "smith", "version": "dev"},
                    "workspaceRoot": options.workspace_root,
                },
                timeout=options.initialize_timeout,
            )
        except Exception as error:
            connection.kill()
            raise RuntimeError(f"initializing plugin: {error}") from error
        if info["protocolVersion"] != PROTOCOL_VERSION:
            connection.kill()
            raise RuntimeError(
                f"plugin {info['plugin']['name']} chose protocol version {info['protocolVersion']}, "
                f"host speaks {PROTOCOL_VERSION}"
            )
        client.info = info
        client._limit = in_flight_limit(info.get("concurrency"))
        return client

    def __init__(self, connection: PluginConnection, cancel_grace: float = DEFAULT_CANCEL_GRACE) -> None:
        self._connection = connection
        self._grace = cancel_grace
        self._lock = threading.RLock()
        self._places = threading.Condition(self._lock)
        self._waiters: dict[int, _Waiter] = {}
        self._queue: deque[object] = deque()
        self._write_lock = threading.Lock()
        self._lost: Exception | None = None
        self._next_id = 0
        self._limit = 1
        self._in_flight = 0
        self._peak = 0
        self.info: InitializeResult = {}  # type: ignore[typeddict-item]
        """What the plugin answered to `initialize`: identity, kind, roles, capabilities."""
        self._reader = threading.Thread(target=self._read_responses, name="smith-plugin-client", daemon=True)
        self._reader.start()

    # -- places among the requests that may be in flight --------------------------------------------

    def _acquire(self, method: str, cancel: CancelToken | None) -> None:
        """Wait for a place among the requests that may be in flight; raise if `cancel` is set first."""
        ticket = object()
        interrupted = CallInterruptedError(f"{method}: cancelled before it was sent", "cancelled", False)
        with self._places:
            if cancel is not None and cancel.cancelled:
                raise interrupted
            self._queue.append(ticket)

            def wake() -> None:
                with self._places:
                    self._places.notify_all()

            remove = cancel.on_cancel(wake) if cancel is not None else (lambda: None)
            try:
                while not (self._queue[0] is ticket and self._in_flight < self._limit):
                    if cancel is not None and cancel.cancelled:
                        raise interrupted
                    self._places.wait()
            except BaseException:
                self._queue.remove(ticket)
                self._places.notify_all()
                raise
            finally:
                remove()
            self._queue.popleft()
            self._in_flight += 1
            self._peak = max(self._peak, self._in_flight)

    def _release(self) -> None:
        with self._places:
            self._in_flight -= 1
            self._places.notify_all()

    # -- one request -------------------------------------------------------------------------------

    def _write(self, body: str) -> None:
        frame = frame_message(body)
        with self._write_lock:
            self._connection.write(frame)

    def _send(self, method: str, params: Any, timeout: float, cancel: CancelToken | None) -> Any:
        if self._lost is not None:
            raise RuntimeError(f"{method}: plugin stopped answering: {self._lost}") from self._lost
        wake = threading.Event()
        waiter = _Waiter(wake)
        with self._lock:
            self._next_id += 1
            request_id = self._next_id
            self._waiters[request_id] = waiter
        body = write_request(request_id, method, None if params is None else encode_params(method, params))
        try:
            self._write(body)
        except (OSError, ValueError) as error:
            self._waiters.pop(request_id, None)
            raise RuntimeError(f"{method}: plugin is not accepting requests: {error}") from error
        remove = cancel.on_cancel(wake.set) if cancel is not None else (lambda: None)
        try:
            wake.wait(timeout)
            if waiter.done.is_set():
                self._waiters.pop(request_id, None)
                return self._decode(method, waiter)
            reason: InterruptionReason = "cancelled" if cancel is not None and cancel.cancelled else "timeout"
            return self._give_up(method, request_id, waiter, reason, timeout)
        finally:
            remove()

    def _decode(self, method: str, waiter: _Waiter) -> Any:
        if waiter.error is not None:
            raise RuntimeError(f"{method}: plugin stopped answering: {waiter.error}") from waiter.error
        assert waiter.body is not None
        try:
            response = parse_response(waiter.body)
            if response.error is not None:
                raise RpcError(response.error.code, response.error.message)
            return decode_result(method, response.result)
        except RpcError:
            raise
        except Exception as error:
            raise RuntimeError(f"{method}: plugin stopped answering: {error}") from error

    def _give_up(
        self, method: str, request_id: int, waiter: _Waiter, reason: InterruptionReason, timeout: float
    ) -> Any:
        """The end of a call whose token was cancelled or whose time ran out.

        A plugin that advertised cancel is sent the notification and has the grace period to answer; one
        that cannot be asked is killed at once when the time ran out, and after the grace period when the
        caller cancelled. A plugin that does not answer in time is killed, and every other call in flight
        to it fails.
        """
        asked = self.supports("cancel")
        if asked:
            try:
                self.notify("cancel", {"id": request_id})
            except Exception:
                pass  # The plugin is gone; waiting for the answer below reports how it went.
        if (asked or reason == "cancelled") and waiter.done.wait(self._grace):
            self._waiters.pop(request_id, None)
            return self._answered_after_stopping(method, waiter, reason, timeout)
        self._waiters.pop(request_id, None)
        self._connection.kill()
        if reason == "timeout":
            raise CallInterruptedError(f"{method}: plugin did not answer within {timeout:g}s", reason, True)
        raise CallInterruptedError(
            f"{method}: cancelled; the plugin did not stop within {self._grace:g}s and was killed", reason, True
        )

    def _answered_after_stopping(self, method: str, waiter: _Waiter, reason: InterruptionReason, timeout: float) -> Any:
        """The plugin answered while the host was giving up: a result is the work finished in time."""
        try:
            return self._decode(method, waiter)
        except Exception as error:
            what = f"did not answer within {timeout:g}s" if reason == "timeout" else "cancelled"
            raise CallInterruptedError(
                f"{method}: {what}: {error}", reason, False, error if isinstance(error, RpcError) else None
            ) from error

    # -- the plugin's side of the pipe -----------------------------------------------------------------

    def _read_responses(self) -> None:
        reader = FrameReader()
        try:
            while True:
                source = self._connection.input
                read1 = getattr(source, "read1", None)
                chunk = bytes((read1 or source.read)(READ_SIZE))
                if not chunk:
                    break
                for body in reader.push(chunk):
                    self._deliver(body)
            self._lose(EOFError("EOF"))
        except Exception as error:
            self._lose(error)

    def _deliver(self, body: str) -> None:
        response: ReceivedResponse
        try:
            response = parse_response(body)
        except Exception as error:
            # A malformed response cannot be attributed to a call: the stream is lost.
            self._lose(RuntimeError(f"malformed response: {error}"))
            return
        if response.id is None:
            return
        with self._lock:
            waiter = self._waiters.pop(response.id, None)
        if waiter is not None:
            waiter.settle(body, None)

    def _lose(self, error: Exception) -> None:
        with self._lock:
            self._lost = error
            waiting, self._waiters = self._waiters, {}
        for waiter in waiting.values():
            waiter.settle(None, error)

    # -- the public surface --------------------------------------------------------------------------

    def call(
        self,
        method: str,
        params: Any = None,
        *,
        timeout: float | None = None,
        cancel: CancelToken | None = None,
    ) -> Any:
        """Send one request and return its decoded result.

        An error answer raises an `RpcError` carrying its code. The call waits for a place first, so
        the plugin never has more requests in flight than it declared; the timeout (seconds, default
        two minutes) counts from the moment the request is sent.

        When `cancel` is cancelled or the time runs out the host stops waiting. A plugin that
        advertised `cancel` is sent the notification and has the grace period to answer: with the result
        it had already (the call returns it), or an error (the call raises a `CallInterruptedError`
        carrying it). A plugin that does not answer in time is killed, because its stream can no longer
        be trusted, and the call raises; a plugin that cannot be asked to stop is killed at once on a
        timeout, and after the grace period when the caller cancelled.
        """
        seconds = DEFAULT_CALL_TIMEOUT if timeout is None else timeout
        if method == "shutdown":
            return self._send(method, params, seconds, cancel)  # it must reach a plugin that is busy
        self._acquire(method, cancel)
        try:
            return self._send(method, params, seconds, cancel)
        finally:
            self._release()

    @property
    def cancel_grace(self) -> float:
        """How long the plugin has to answer after it was asked to stop, in seconds."""
        return self._grace

    @property
    def peak_in_flight(self) -> int:
        """The most requests this client has had in flight at once."""
        return self._peak

    def notify(self, method: str, params: Any) -> None:
        """Send a message that gets no answer, such as `cancel`."""
        if self._lost is not None:
            raise RuntimeError(f"{method}: plugin stopped answering: {self._lost}") from self._lost
        try:
            self._write(write_notification(method, encode_params(method, params)))
        except (OSError, ValueError) as error:
            raise RuntimeError(f"{method}: plugin is not accepting requests: {error}") from error

    def supports(self, method: str) -> bool:
        """Whether the plugin advertised a method.

        `initialize` and `shutdown` always, the others by their capability.
        """
        if method in ("initialize", "shutdown"):
            return True
        capability = METHOD_CAPABILITIES.get(method)
        advertised = self.info.get("capabilities") or {}
        return capability is not None and advertised.get(capability) is True

    def close(self) -> None:
        """Ask the plugin to shut down and wait briefly for it to go. Raises if it did not go quietly."""
        try:
            self.call("shutdown", {}, timeout=5.0)
        except Exception:
            pass  # A plugin that is already gone cannot be asked; waiting for it below reports how it went.
        self._connection.close_input()
        trouble = self._connection.finish(5.0)
        deadline = time.monotonic() + 2.0
        self._reader.join(max(0.0, deadline - time.monotonic()))
        self._connection.dispose()
        if trouble is not None:
            raise RuntimeError(trouble)

    def __enter__(self) -> PluginClient:
        return self

    def __exit__(self, *exc: object) -> None:
        self.close()
