"""The requests in flight, and the bounded set of worker threads that runs them."""

from __future__ import annotations

import threading
from collections import deque
from collections.abc import Callable
from dataclasses import dataclass, field
from typing import Literal

from ..cancellation import CancelToken
from ..codec import ReceivedRequest

MAX_QUEUED = 1024
"""The requests waiting for a worker are limited to this many.

The host never has more in flight than the plugin declared, so more means a host that breaks the protocol.
"""

Admission = Literal["admitted", "duplicate", "full"]
"""What `RequestPool.submit` decided."""


@dataclass
class PooledRequest:
    """One request from the moment it is read until it is answered."""

    request: ReceivedRequest
    """The request; its id is never `None`."""
    token: CancelToken = field(default_factory=CancelToken)
    """Cancelled when the host cancels the request (or the plugin is shutting down)."""
    cancelled: bool = False
    """Whether the host asked for the request to stop."""

    @property
    def id(self) -> int:
        """The request id."""
        assert self.request.id is not None
        return self.request.id


class RequestPool:
    """The state of the serve loop: the requests in flight, the ones waiting for a worker, and a
    bounded number of worker threads running them.

    Requests start in the order they arrive and may finish in any order. It is the Go SDK's
    request pool, with the same rules. Threads are started when there is work and a free place,
    and end when there is none, so an idle plugin holds no thread.
    """

    def __init__(self, run: Callable[[PooledRequest], None]) -> None:
        self._run = run
        self._limit = 1
        self._running = 0
        self._queue: deque[PooledRequest] = deque()
        self._inflight: dict[int, PooledRequest] = {}
        self._condition = threading.Condition()
        self._failure: BaseException | None = None

    @property
    def failure(self) -> BaseException | None:
        """The first exception a run raised, if any."""
        return self._failure

    def set_limit(self, limit: int) -> None:
        """How many workers may run at once.

        Raising it starts waiting requests; lowering it lets running ones finish.
        """
        with self._condition:
            self._limit = max(1, limit)
            while self._running < self._limit and self._queue:
                self._start(self._queue.popleft())

    def submit(self, request: ReceivedRequest) -> Admission:
        """Take a request: it runs at once on a free worker, or waits its turn."""
        assert request.id is not None
        with self._condition:
            if request.id in self._inflight:
                return "duplicate"
            if self._running >= self._limit and len(self._queue) >= MAX_QUEUED:
                return "full"
            pooled = PooledRequest(request)
            self._inflight[request.id] = pooled
            if self._running >= self._limit:
                self._queue.append(pooled)
            else:
                self._start(pooled)
        return "admitted"

    def release(self, pooled: PooledRequest) -> None:
        """Forget an answered request; call it before the answer is written, so a host that reuses
        the id right away is not refused as a duplicate."""
        with self._condition:
            self._inflight.pop(pooled.id, None)

    def cancel(self, request_id: int) -> PooledRequest | None:
        """Stop the request with an id.

        One that had not started is taken out of the queue and returned, because no worker will
        ever answer it. An unknown id is ignored.
        """
        with self._condition:
            pooled = self._inflight.get(request_id)
            if pooled is None:
                return None
            pooled.cancelled = True
            pooled.token.cancel()
            if pooled not in self._queue:
                return None
            self._queue.remove(pooled)
            del self._inflight[request_id]
            return pooled

    def cancel_all(self) -> list[PooledRequest]:
        """Stop every request. The ones that had not started are returned, to be answered."""
        with self._condition:
            for pooled in self._inflight.values():
                pooled.cancelled = True
                pooled.token.cancel()
            waiting = list(self._queue)
            self._queue.clear()
            for pooled in waiting:
                del self._inflight[pooled.id]
            return waiting

    def idle(self) -> None:
        """Block until no worker is running."""
        with self._condition:
            self._condition.wait_for(lambda: self._running == 0)

    def _start(self, first: PooledRequest) -> None:
        # Called with the lock held.
        self._running += 1
        threading.Thread(target=self._work, args=(first,), name=f"smith-plugin-request-{first.id}", daemon=True).start()

    def _work(self, first: PooledRequest) -> None:
        pooled: PooledRequest | None = first
        while pooled is not None:
            try:
                self._run(pooled)
            except BaseException as error:  # noqa: BLE001 - kept for the loop to raise; the pool goes on
                if self._failure is None:
                    self._failure = error
            pooled = self._next()

    def _next(self) -> PooledRequest | None:
        with self._condition:
            if self._queue and self._running <= self._limit:
                return self._queue.popleft()
            self._running -= 1
            if self._running == 0:
                self._condition.notify_all()
            return None
