"""Fixtures shared by the tests: a plugin served over pipes that a test talks to frame by frame."""

from __future__ import annotations

import json
import os
import shutil
import subprocess
import threading
from collections.abc import Callable, Iterator
from pathlib import Path
from typing import Any

import pytest

from monosmith_plugin_sdk import CancelToken, FrameReader, ServeStreams, frame_message, serve_streams

WAIT = 30.0
"""How long a test waits for a frame before it fails, in seconds. Never a pause the test depends on."""


class Widget:
    """The smallest language plugin; subclasses add the methods under test."""

    def describe(self, params: Any) -> Any:
        return {"plugin": {"name": "widget", "version": "0.1.0"}, "kind": "language", "manifests": ["widget.json"]}


class Lab:
    """A plugin whose `discover` does what the manifest name says.

    `block*` waits for `release` (or for the cancel), `stubborn*` waits for `release` and ignores the cancel, `fail*`
    raises, anything else answers at once. Every wait is on an event, never a sleep.
    """

    def __init__(self, concurrency: int | None = None, cancellable: bool = False) -> None:
        self.cancellable = cancellable
        self._concurrency = concurrency
        self.release = threading.Event()
        self.lock = threading.Lock()
        self.active = 0
        self.peak = 0
        self.started: list[str] = []
        self.started_changed = threading.Condition(self.lock)
        self.aborted = 0

    def describe(self, params: Any) -> Any:
        description: dict[str, Any] = {
            "plugin": {"name": "lab", "version": "1.0.0"},
            "kind": "language",
            "manifests": ["lab.json"],
        }
        if self._concurrency is not None:
            description["concurrency"] = self._concurrency
        return description

    def discover(self, params: Any, cancel: CancelToken) -> Any:
        name = params["manifests"][0] if params["manifests"] else ""
        with self.started_changed:
            self.active += 1
            self.peak = max(self.peak, self.active)
            self.started.append(name)
            self.started_changed.notify_all()
        try:
            if name.startswith("fail"):
                raise RuntimeError("the manifest is broken")
            if name.startswith("block"):
                while not self.release.is_set():
                    if cancel.wait(0.005):
                        with self.lock:
                            self.aborted += 1
                        raise RuntimeError("aborted")
            if name.startswith("stubborn"):
                assert self.release.wait(WAIT)
        finally:
            with self.lock:
                self.active -= 1
        return {"projects": [{"name": name, "root": name, "language": "lab", "kind": "lib"}]}

    def wait_started(self, count: int) -> None:
        with self.started_changed:
            assert self.started_changed.wait_for(lambda: len(self.started) >= count, WAIT), self.started


class Conversation:
    """A plugin served in this process over pipes, spoken to with raw frames.

    The test sends bodies exactly as it writes them and reads the bodies the plugin wrote, so it can
    check the bytes. Nothing here sleeps: `recv` blocks until a frame arrives.
    """

    def __init__(self, plugin: Any) -> None:
        to_plugin_read, to_plugin_write = os.pipe()
        from_plugin_read, from_plugin_write = os.pipe()
        self._plugin_input = os.fdopen(to_plugin_read, "rb", buffering=0)
        self._plugin_output = os.fdopen(from_plugin_write, "wb", buffering=0)
        self._input = os.fdopen(to_plugin_write, "wb", buffering=0)
        self._output = os.fdopen(from_plugin_read, "rb", buffering=0)
        self._reader = FrameReader()
        self._pending: list[str] = []
        self.logs: list[str] = []
        self.error: BaseException | None = None
        self._thread = threading.Thread(target=self._serve, args=(plugin,), daemon=True)
        self._thread.start()

    def _serve(self, plugin: Any) -> None:
        try:
            serve_streams(plugin, ServeStreams(self._plugin_input, self._plugin_output, self.logs.append))
        except BaseException as error:
            self.error = error
        finally:
            self._plugin_output.close()

    def send(self, message: dict[str, Any]) -> None:
        """Write a JSON-RPC message as a frame."""
        self.send_raw(json.dumps({"jsonrpc": "2.0", **message}))

    def send_raw(self, body: str) -> None:
        """Write a body as a frame, as it is."""
        self._input.write(frame_message(body))

    def send_bytes(self, data: bytes) -> None:
        """Write bytes as they are, frame or not."""
        self._input.write(data)

    def request(self, request_id: int, method: str, params: Any = None) -> None:
        """Write a request."""
        message: dict[str, Any] = {"id": request_id, "method": method}
        if params is not None:
            message["params"] = params
        self.send(message)

    def notify(self, method: str, params: Any = None) -> None:
        """Write a notification."""
        self.send({"method": method, "params": params})

    def recv(self) -> str:
        """The body of the next frame the plugin wrote; fails the test if none comes."""
        while not self._pending:
            chunk = self._output.read(65536)
            assert chunk, "the plugin closed its output before the expected frame"
            self._pending.extend(self._reader.push(chunk))
        return self._pending.pop(0)

    def recv_json(self) -> dict[str, Any]:
        """The next frame, parsed."""
        parsed: dict[str, Any] = json.loads(self.recv())
        return parsed

    def recv_all(self) -> list[str]:
        """Every remaining frame, until the plugin closes its output."""
        bodies = list(self._pending)
        self._pending.clear()
        while chunk := self._output.read(65536):
            bodies.extend(self._reader.push(chunk))
        return bodies

    def close_input(self) -> None:
        """End the plugin's input, as a host that is gone does."""
        self._input.close()

    def finish(self) -> None:
        """Wait for the serve loop to end; re-raise what it raised."""
        self._thread.join(WAIT)
        assert not self._thread.is_alive(), "the serve loop did not end"
        for stream in (self._input, self._output, self._plugin_input):
            stream.close()
        if self.error is not None:
            raise self.error

    def handshake(self, version_list: list[int] | None = None) -> dict[str, Any]:
        """Send `initialize` with id 1 and return the result."""
        self.request(
            1,
            "initialize",
            {"protocolVersions": version_list or [1], "host": {"name": "smith", "version": "1"}, "workspaceRoot": "/w"},
        )
        answer = self.recv_json()
        assert "result" in answer, answer
        result: dict[str, Any] = answer["result"]
        return result


@pytest.fixture
def converse() -> Iterator[Callable[[Any], Conversation]]:
    """Serve a plugin over pipes: `conversation = converse(plugin)`."""
    started: list[Conversation] = []

    def start(plugin: Any) -> Conversation:
        conversation = Conversation(plugin)
        started.append(conversation)
        return conversation

    yield start
    for conversation in started:
        conversation.close_input()
        conversation.finish()


def tool_available(program: str, *args: str) -> bool:
    """Whether a program runs: the probes of tests that need a tool on the path."""
    if shutil.which(program) is None:
        return False
    return subprocess.run([program, *args], capture_output=True, check=False).returncode == 0


@pytest.fixture
def workspace(tmp_path: Path) -> Path:
    """An empty directory to act as the workspace root."""
    return tmp_path
