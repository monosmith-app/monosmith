"""The host's client: handshake, concurrency, cancel and timeouts, over an in-process connection."""

from __future__ import annotations

import threading
from typing import Any

import pytest

from monosmith_plugin_sdk import (
    CallInterruptedError,
    CancelToken,
    HostOptions,
    PluginClient,
    RpcError,
    connect_in_process,
)
from tests.conftest import Lab, Widget


def connect(plugin: Any, grace: float = 5.0) -> PluginClient:
    return connect_in_process(plugin, HostOptions(workspace_root="/w", cancel_grace=grace))


def call_in_thread(
    client: PluginClient, manifest: str, results: list[Any], cancel: CancelToken | None = None
) -> threading.Thread:
    def run() -> None:
        try:
            results.append(client.call("discover", {"manifests": [manifest]}, cancel=cancel))
        except BaseException as error:
            results.append(error)

    thread = threading.Thread(target=run)
    thread.start()
    return thread


def test_the_handshake_gives_the_plugin_info_and_supports_follows_the_capabilities() -> None:
    with connect(Lab(cancellable=True, concurrency=2)) as client:
        assert client.info["plugin"]["name"] == "lab"
        assert client.supports("discover") and client.supports("cancel") and client.supports("shutdown")
        assert not client.supports("doctor")


def test_an_error_answer_raises_an_rpc_error_with_its_code() -> None:
    with connect(Widget()) as client:
        with pytest.raises(RpcError) as caught:
            client.call("doctor", {"projects": []})
        assert caught.value.code == -32601


def test_never_has_more_requests_in_flight_than_the_plugin_declared() -> None:
    lab = Lab(concurrency=1)
    with connect(lab) as client:
        results: list[Any] = []
        threads = [call_in_thread(client, f"quick-{i}", results) for i in range(5)]
        for thread in threads:
            thread.join()
        assert len(results) == 5
        assert client.peak_in_flight == 1
        assert lab.peak == 1


def test_runs_as_many_calls_together_as_the_plugin_declared() -> None:
    lab = Lab(concurrency=4)
    with connect(lab) as client:
        results: list[Any] = []
        threads = [call_in_thread(client, f"stubborn-{i}", results) for i in range(4)]
        lab.wait_started(4)
        lab.release.set()
        for thread in threads:
            thread.join()
        assert client.peak_in_flight == 4
        assert all(isinstance(result, dict) for result in results)


def test_a_cancelled_call_to_a_cancellable_plugin_raises_with_the_cancelled_error_and_the_plugin_survives() -> None:
    lab = Lab(cancellable=True)
    with connect(lab) as client:
        token = CancelToken()
        results: list[Any] = []
        thread = call_in_thread(client, "block-a", results, token)
        lab.wait_started(1)
        token.cancel()
        thread.join()
        error = results[0]
        assert isinstance(error, CallInterruptedError)
        assert error.reason == "cancelled"
        assert error.killed is False
        assert error.response is not None and error.response.code == -32800
        assert client.call("discover", {"manifests": ["quick"]})["projects"][0]["name"] == "quick"


def test_a_cancelled_call_returns_the_result_when_the_plugin_had_finished_the_work_anyway() -> None:
    lab = Lab(cancellable=True)
    with connect(lab) as client:
        token = CancelToken()
        results: list[Any] = []
        thread = call_in_thread(client, "stubborn", results, token)
        lab.wait_started(1)
        token.cancel()
        lab.release.set()  # the plugin ignores the cancel and finishes
        thread.join()
        assert results[0]["projects"][0]["name"] == "stubborn"


def test_a_plugin_that_cannot_be_cancelled_is_killed_after_the_grace_period() -> None:
    lab = Lab()
    client = connect(lab, grace=0.2)
    token = CancelToken()
    results: list[Any] = []
    thread = call_in_thread(client, "stubborn", results, token)
    lab.wait_started(1)
    token.cancel()
    thread.join()
    lab.release.set()
    error = results[0]
    assert isinstance(error, CallInterruptedError)
    assert error.killed is True
    with pytest.raises(RuntimeError, match="plugin stopped answering|not accepting"):
        client.call("discover", {"manifests": []})


def test_a_timeout_of_a_cancellable_plugin_sends_cancel_and_does_not_kill() -> None:
    lab = Lab(cancellable=True)
    with connect(lab) as client:
        with pytest.raises(CallInterruptedError) as caught:
            client.call("discover", {"manifests": ["block-a"]}, timeout=0.1)
        assert caught.value.reason == "timeout"
        assert caught.value.killed is False
        assert lab.aborted == 1


def test_a_call_cancelled_before_it_is_sent_raises_and_sends_nothing() -> None:
    lab = Lab()
    with connect(lab) as client:
        token = CancelToken()
        token.cancel()
        with pytest.raises(CallInterruptedError, match="cancelled before it was sent"):
            client.call("discover", {"manifests": ["quick"]}, cancel=token)
        assert lab.started == []


def test_a_call_waiting_for_a_place_can_be_cancelled() -> None:
    lab = Lab(concurrency=1)
    with connect(lab) as client:
        results: list[Any] = []
        first = call_in_thread(client, "stubborn-first", results)
        lab.wait_started(1)
        token = CancelToken()
        waiting_results: list[Any] = []
        second = call_in_thread(client, "quick", waiting_results, token)
        token.cancel()
        second.join()
        assert isinstance(waiting_results[0], CallInterruptedError)
        lab.release.set()
        first.join()
        assert lab.started == ["stubborn-first"]
