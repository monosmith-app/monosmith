"""The token that carries a cancel to the work it concerns."""

from __future__ import annotations

import threading
from collections.abc import Callable

from .request_cancelled_error import RequestCancelledError


class CancelToken:
    """A flag one thread sets and another watches, to stop work cooperatively.

    The serve loop gives a method one for the request it is answering, and sets it when the
    host sends `cancel` for the request (or when the plugin shuts down). A host in a test
    passes one to `PluginClient.call` and sets it to cancel the call.

    The usual loop checks it between steps, and waits on it instead of sleeping:

        for path in files:
            cancel.raise_if_cancelled()
            scan(path)
        if cancel.wait(0.1):   # sleeps up to 0.1 s, and wakes at once on a cancel
            return partial_result
    """

    def __init__(self) -> None:
        self._event = threading.Event()
        self._lock = threading.Lock()
        self._callbacks: list[Callable[[], None]] = []

    @property
    def cancelled(self) -> bool:
        """Whether the work was asked to stop."""
        return self._event.is_set()

    def cancel(self) -> None:
        """Ask the work to stop. Calling it again does nothing."""
        with self._lock:
            if self._event.is_set():
                return
            self._event.set()
            callbacks, self._callbacks = self._callbacks, []
        for callback in callbacks:
            callback()

    def wait(self, timeout: float | None = None) -> bool:
        """Wait up to `timeout` seconds (forever for `None`); return whether it was cancelled.

        A cancellable sleep: it returns early, with `True`, the moment the token is cancelled.
        """
        return self._event.wait(timeout)

    def raise_if_cancelled(self, message: str = "the request was cancelled") -> None:
        """Raise `RequestCancelledError` when the work was asked to stop."""
        if self._event.is_set():
            raise RequestCancelledError(message)

    def on_cancel(self, callback: Callable[[], None]) -> Callable[[], None]:
        """Call `callback` (in the cancelling thread) when the token is cancelled, at once if it already is.

        Returns a function that removes the callback again.
        """
        with self._lock:
            if not self._event.is_set():
                self._callbacks.append(callback)

                def remove() -> None:
                    with self._lock:
                        if callback in self._callbacks:
                            self._callbacks.remove(callback)

                return remove
        callback()
        return lambda: None
