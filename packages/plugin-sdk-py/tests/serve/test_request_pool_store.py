"""The request pool: places, order, duplicates and cancel."""

from __future__ import annotations

import threading

from monosmith_plugin_sdk.codec import ReceivedRequest
from monosmith_plugin_sdk.serve.request_pool_store import MAX_QUEUED, PooledRequest, RequestPool

WAIT = 30.0


def request(request_id: int) -> ReceivedRequest:
    return ReceivedRequest(request_id, "discover", {})


class Gate:
    """Runs hold on a gate until the test opens it; records what ran, in start order."""

    def __init__(self) -> None:
        self.open = threading.Event()
        self.started: list[int] = []
        self.lock = threading.Lock()
        self.changed = threading.Condition(self.lock)
        self.finished: list[int] = []

    def run(self, pooled: PooledRequest) -> None:
        with self.changed:
            self.started.append(pooled.id)
            self.changed.notify_all()
        assert self.open.wait(WAIT)
        with self.changed:
            self.finished.append(pooled.id)
            self.changed.notify_all()

    def wait_started(self, count: int) -> None:
        with self.changed:
            assert self.changed.wait_for(lambda: len(self.started) >= count, WAIT)


def test_starts_requests_in_order_on_as_many_workers_as_the_limit() -> None:
    gate = Gate()
    pool = RequestPool(gate.run)
    pool.set_limit(2)
    assert [pool.submit(request(i)) for i in (1, 2, 3, 4)] == ["admitted"] * 4
    gate.wait_started(2)
    assert gate.started == [1, 2]
    gate.open.set()
    pool.idle()
    assert sorted(gate.finished) == [1, 2, 3, 4]
    assert gate.started[2:] == [3, 4]


def test_refuses_an_id_in_flight_until_it_is_released() -> None:
    gate = Gate()
    pool = RequestPool(gate.run)
    assert pool.submit(request(1)) == "admitted"
    assert pool.submit(request(1)) == "duplicate"
    gate.wait_started(1)
    gate.open.set()
    pool.idle()
    assert pool.submit(request(1)) == "duplicate"  # run does not release; the serve loop does, before it answers


def test_refuses_a_request_when_every_worker_is_busy_and_the_queue_is_full() -> None:
    gate = Gate()
    pool = RequestPool(gate.run)
    assert pool.submit(request(0)) == "admitted"
    gate.wait_started(1)
    for number in range(1, MAX_QUEUED + 1):
        assert pool.submit(request(number)) == "admitted"
    assert pool.submit(request(MAX_QUEUED + 1)) == "full"
    gate.open.set()
    pool.idle()


def test_cancel_takes_a_waiting_request_out_and_marks_a_running_one() -> None:
    gate = Gate()
    pool = RequestPool(gate.run)
    pool.submit(request(1))
    pool.submit(request(2))
    gate.wait_started(1)
    waiting = pool.cancel(2)
    assert waiting is not None
    assert waiting.cancelled and waiting.token.cancelled
    assert pool.cancel(2) is None  # gone
    assert pool.cancel(1) is None  # running: nobody answers it but its worker
    assert pool.cancel(99) is None
    gate.open.set()
    pool.idle()
    assert gate.started == [1]


def test_cancel_all_marks_everything_and_returns_the_waiting() -> None:
    gate = Gate()
    pool = RequestPool(gate.run)
    pool.submit(request(1))
    pool.submit(request(2))
    pool.submit(request(3))
    gate.wait_started(1)
    waiting = pool.cancel_all()
    assert [w.id for w in waiting] == [2, 3]
    gate.open.set()
    pool.idle()


def test_keeps_the_first_failure_and_goes_on() -> None:
    ran: list[int] = []

    def run(pooled: PooledRequest) -> None:
        ran.append(pooled.id)
        raise ValueError(f"boom {pooled.id}")

    pool = RequestPool(run)
    pool.submit(request(1))
    pool.idle()
    pool.submit(request(2))
    pool.idle()
    assert ran == [1, 2]
    assert str(pool.failure) == "boom 1"
