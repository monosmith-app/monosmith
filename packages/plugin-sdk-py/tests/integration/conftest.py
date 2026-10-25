"""Helpers of the integration tests: they start real processes (smith, go, node, the example plugins)."""

from __future__ import annotations

import json
import os
import shutil
import subprocess
import sys
import tempfile
import threading
from collections.abc import Iterator
from dataclasses import dataclass
from pathlib import Path
from typing import Any

import pytest

from monosmith_plugin_sdk import FrameReader, frame_message

EXAMPLES = Path(__file__).resolve().parents[4] / "docs" / "examples" / "plugins"
"""docs/examples/plugins"""

WAIT = 120.0

IN_CI = any(os.environ.get(name) for name in ("CI", "GITHUB_ACTIONS", "TF_BUILD"))
"""CI has smith, go and node: a missing one there is a failure, not a reason to skip."""


def require(available: bool, what: str) -> None:
    """Skip a test locally when a tool is missing; fail it in CI, where nothing may be missing."""
    if available:
        return
    if IN_CI:
        pytest.fail(f"{what} is required in CI and was not found")
    pytest.skip(f"{what} is not available")


def runs(program: str, *args: str) -> bool:
    """Whether a program starts and succeeds."""
    try:
        return subprocess.run([program, *args], capture_output=True, check=False, timeout=60).returncode == 0
    except (OSError, subprocess.SubprocessError):
        return False


def clean_environment(**extra: str) -> dict[str, str]:
    """The environment of a subprocess: without the variable that would send a session to another workspace."""
    env = {**os.environ, "NO_COLOR": "1", **extra}
    env.pop("SMITH_WORKSPACE_ROOT", None)
    return env


@dataclass
class RawRequest:
    """A message to send as it is: the id is the position in the conversation, from 1."""

    method: str
    params: Any = None
    notification: bool = False
    """Send it as a notification (no id). The ids of the requests after it are still their positions."""
    wait_for: int | None = None
    """Wait until the plugin has written this many frames before the next message is sent."""


def converse(command: list[str], cwd: Path, requests: list[RawRequest], env: dict[str, str] | None = None) -> list[str]:
    """Start a plugin the way smith does (`SMITH_PLUGIN=1`), write every request as a frame and return the body of
    every frame it answered with, exactly as the plugin wrote it.

    The conversation ends when the plugin answers `shutdown`, which this adds to the end. Nothing here waits for a
    time: a message that must come after an answer says how many frames to wait for.
    """
    child = subprocess.Popen(
        command,
        cwd=cwd,
        env={**os.environ, "SMITH_PLUGIN": "1", **(env or {})},
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        bufsize=0,
    )
    assert child.stdin is not None and child.stdout is not None
    bodies: list[str] = []
    changed = threading.Condition()

    def read() -> None:
        assert child.stdout is not None
        reader = FrameReader()
        while chunk := child.stdout.read(65536):
            with changed:
                bodies.extend(reader.push(chunk))
                changed.notify_all()

    thread = threading.Thread(target=read, daemon=True)
    thread.start()
    messages = [*requests, RawRequest("shutdown", {})]
    try:
        for position, message in enumerate(messages, start=1):
            body: dict[str, Any] = {"jsonrpc": "2.0", "method": message.method}
            if not message.notification:
                body = {"jsonrpc": "2.0", "id": position, "method": message.method}
            if message.params is not None:
                body["params"] = message.params
            child.stdin.write(frame_message(json.dumps(body)))
            if message.wait_for is not None:
                with changed:
                    assert changed.wait_for(lambda m=message: len(bodies) >= (m.wait_for or 0), WAIT), bodies  # type: ignore[misc]
        # A plugin that refused every request (none came after initialize) does not stop at shutdown: closing its
        # input does.
        child.stdin.close()
        child.wait(WAIT)
    finally:
        if child.poll() is None:
            child.kill()
            child.wait()
    thread.join(WAIT)
    return bodies


def by_id(bodies: list[str]) -> list[str]:
    """The bodies in the order of their ids: a plugin that runs requests together may answer in any order."""
    return sorted(bodies, key=lambda body: json.loads(body)["id"])


def python_command(script: Path) -> list[str]:
    """The command that starts a Python example with the interpreter of the tests (which has the SDK)."""
    return [sys.executable, str(script)]


@dataclass
class Side:
    """One implementation of a plugin: how to start it, and how it writes the command of its own tasks."""

    command: list[str]
    self_text: str
    """The text its tasks' command is written as in the answers, replaced by one name."""

    def normalise(self, body: str) -> str:
        return body.replace(self.self_text, '"<self>"')


@pytest.fixture(scope="module")
def scratch() -> Iterator[Path]:
    path = Path(tempfile.mkdtemp(prefix="sdk-cross-"))
    yield path
    shutil.rmtree(path, ignore_errors=True)


def go_side(scratch: Path, example: str) -> Side:
    require(runs("go", "version"), "go")
    binary = scratch / f"{example}{'.exe' if sys.platform == 'win32' else ''}"
    built = subprocess.run(
        ["go", "build", "-o", str(binary), "."],
        cwd=EXAMPLES / example,
        env={**os.environ, "GOWORK": "off"},
        capture_output=True,
        text=True,
        check=False,
    )
    assert built.returncode == 0, f"go build {example}: {built.stderr}"
    return Side([str(binary)], json.dumps(str(binary)))


def ts_side(example: str) -> Side:
    require(runs("node", "--version"), "node")
    require(
        (EXAMPLES.parents[2] / "node_modules" / "@monosmith" / "plugin-sdk").exists(), "the installed npm workspace"
    )
    plugin = str(EXAMPLES / example / "plugin.ts")
    node = shutil.which("node") or "node"
    return Side([node, plugin], f'"node",{json.dumps(plugin)}')


def py_side(example: str) -> Side:
    plugin = EXAMPLES / example / "plugin.py"
    return Side(python_command(plugin), f"{json.dumps(sys.executable)},{json.dumps(str(plugin))}")


def sides(scratch: Path, go: str, ts: str, py: str) -> dict[str, Side]:
    return {"go": go_side(scratch, go), "ts": ts_side(ts), "py": py_side(py)}
