"""The cancel token."""

from __future__ import annotations

import threading

import pytest

from monosmith_plugin_sdk import CancelToken, RequestCancelledError


def test_starts_uncancelled_and_cancels_once() -> None:
    token = CancelToken()
    assert not token.cancelled
    token.cancel()
    token.cancel()
    assert token.cancelled


def test_wait_returns_at_once_when_cancelled_and_false_when_the_time_runs_out() -> None:
    token = CancelToken()
    assert token.wait(0) is False
    threading.Timer(0, token.cancel).start()
    assert token.wait(30) is True


def test_raise_if_cancelled() -> None:
    token = CancelToken()
    token.raise_if_cancelled()
    token.cancel()
    with pytest.raises(RequestCancelledError, match="was cancelled"):
        token.raise_if_cancelled()


def test_callbacks_run_once_at_cancel_or_at_once_if_already_cancelled_and_can_be_removed() -> None:
    token = CancelToken()
    calls: list[str] = []
    token.on_cancel(lambda: calls.append("a"))
    remove = token.on_cancel(lambda: calls.append("b"))
    remove()
    token.cancel()
    token.cancel()
    assert calls == ["a"]
    token.on_cancel(lambda: calls.append("late"))
    assert calls == ["a", "late"]
