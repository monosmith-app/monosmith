"""Launching a plugin process the way smith does."""

from __future__ import annotations

import os
import subprocess
import sys
from typing import IO, Any

from .host_options_contract import StartOptions
from .plugin_client_client import PluginClient, PluginConnection


def start_plugin(options: StartOptions) -> PluginClient:
    """Launch a plugin process and perform the handshake, as smith does.

    The plugin's argv, `SMITH_PLUGIN=1` in its environment, its standard input and output as the
    protocol's channel, its standard error forwarded to this process's. Close the client when
    done (or use it as a context manager): it asks the plugin to shut down and kills it if it does not go.
    """
    command = list(options.command)
    if not command:
        raise ValueError("plugin command is empty")
    env = dict(os.environ if options.env is None else options.env)
    env.setdefault("SMITH_PLUGIN", "1")
    extra: dict[str, Any] = {}
    if sys.platform == "win32":
        extra["creationflags"] = subprocess.CREATE_NO_WINDOW
    try:
        process = subprocess.Popen(
            command,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL if options.stderr == "ignore" else None,
            env=env,
            cwd=options.cwd,
            bufsize=0,
            **extra,
        )
    except OSError as error:
        raise RuntimeError(f"starting plugin {command!r}: {error}") from error
    assert process.stdin is not None and process.stdout is not None
    stdin: IO[bytes] = process.stdin
    stdout: IO[bytes] = process.stdout

    def write(data: bytes) -> None:
        stdin.write(data)
        stdin.flush()

    def close_input() -> None:
        try:
            stdin.close()
        except OSError:
            pass

    def finish(grace: float) -> str | None:
        try:
            code = process.wait(timeout=grace)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()
            return "plugin did not exit after shutdown; killed"
        return None if code == 0 else f"plugin exited with status {code}"

    connection = PluginConnection(
        input=stdout, write=write, close_input=close_input, kill=process.kill, finish=finish, dispose=stdout.close
    )
    try:
        return PluginClient.open(connection, options)
    except BaseException:
        process.kill()
        process.wait()
        raise
