"""The checks of how a plugin is called: its declared concurrency and its cancel."""

from __future__ import annotations

import threading
import time
from collections.abc import Callable
from concurrent.futures import ThreadPoolExecutor
from typing import Any

from ..cancellation import CancelToken
from ..concurrency import in_flight_limit, is_valid_concurrency
from ..host import CallInterruptedError, PluginClient
from ..protocol import ERROR_CODES
from .conformance_report_model import ConformanceReport

CANCEL_BOUND = 10.0
"""How long a plugin that advertised cancel has to answer a request the host cancelled, in seconds."""

UNKNOWN_REQUEST = 2**40
"""An id no request ever had."""

CALL_TIMEOUT = 2 * 60.0

Probe = Callable[[CancelToken | None], Any]


def _probe_of(client: PluginClient) -> Probe | None:
    """The cheapest call the plugin answers without doing work.

    A method that takes a list of projects or manifests, asked about none.
    """
    if client.supports("discover"):
        return lambda cancel: client.call("discover", {"manifests": []}, timeout=CALL_TIMEOUT, cancel=cancel)
    for method in ("contribute", "ci", "doctor"):
        if client.supports(method):
            return lambda cancel, method=method: client.call(
                method, {"projects": []}, timeout=CALL_TIMEOUT, cancel=cancel
            )  # type: ignore[misc]
    return None


def check_scheduling(client: PluginClient, report: ConformanceReport, cancel_bound: float = CANCEL_BOUND) -> None:
    """Check what a plugin says about how it is called.

    The concurrency it declares must be one the protocol defines and be accepted, and the cancel it
    advertises must stop a request within a bound and leave the plugin serving. The same checks, under
    the same names, as the Go SDK's.
    """
    info = client.info
    declared = info.get("concurrency") or 0
    report.add(
        "declares a concurrency the protocol defines",
        is_valid_concurrency(declared),
        f"concurrency {declared} is below -1; it must be absent, 1 or more, or -1 for unlimited",
    )
    probe = _probe_of(client)
    if probe is None:
        return
    _check_concurrency(client, probe, report)
    if info.get("capabilities", {}).get("cancel") is True:
        _check_cancel(client, probe, report, cancel_bound)


def _check_concurrency(client: PluginClient, probe: Probe, report: ConformanceReport) -> None:
    """Send requests together.

    A plugin that declares nothing is sent them one at a time and must answer them all; one that
    declares a concurrency is sent that many at once (at most four, to keep the check cheap) and must
    answer them all.
    """
    limit = in_flight_limit(client.info.get("concurrency"))
    together = min(limit, 4)
    name = f"accepts {together} requests at the same time"
    if limit == 1:
        together = 3
        name = "answers 3 requests sent together, one at a time"
    failure: BaseException | None = None
    with ThreadPoolExecutor(max_workers=together) as pool:
        futures = [pool.submit(probe, None) for _ in range(together)]
        for future in futures:
            error = future.exception()
            if error is not None and failure is None:
                failure = error
    peak = client.peak_in_flight
    detail = "<nil>" if failure is None else str(failure)
    report.add(
        name,
        failure is None and peak <= limit,
        f"{detail} (the host had {peak} in flight, the plugin declared {limit})",
    )


def _check_cancel(client: PluginClient, probe: Probe, report: ConformanceReport, cancel_bound: float) -> None:
    """Cancel a request, then a request that does not exist."""
    token = CancelToken()
    outcome: dict[str, BaseException | None] = {}
    finished = threading.Event()

    def run() -> None:
        try:
            probe(token)
            outcome["error"] = None
        except BaseException as error:
            outcome["error"] = error
        finished.set()

    began = time.monotonic()
    threading.Thread(target=run, daemon=True).start()
    time.sleep(0.005)  # let the request leave; a cancel that wins the race is just as good an answer
    token.cancel()
    late = not finished.wait(cancel_bound + client.cancel_grace)
    error: BaseException | None = outcome.get("error")
    if late:
        error = RuntimeError(f"no answer within {cancel_bound:g}s of the cancel")
    stopped = not late and (error is None or (isinstance(error, CallInterruptedError) and not error.killed))
    detail = (
        f"{'<nil>' if error is None else error} (a plugin that advertised cancel answers the request, with "
        f"{ERROR_CODES.REQUEST_CANCELLED} if it stopped, within {cancel_bound:g}s)"
    )
    report.add(
        "answers a cancelled request with its result or the cancelled error",
        stopped and time.monotonic() - began < cancel_bound,
        detail,
    )

    notify_error: Exception | None = None
    follow_up: Exception | None = None
    try:
        client.notify("cancel", {"id": UNKNOWN_REQUEST})
    except Exception as caught:
        notify_error = caught
    try:
        probe(None)
    except Exception as caught:
        follow_up = caught
    report.add(
        "ignores a cancel for a request it does not know",
        notify_error is None and follow_up is None,
        f"{'<nil>' if notify_error is None else notify_error}, then {'<nil>' if follow_up is None else follow_up}",
    )
