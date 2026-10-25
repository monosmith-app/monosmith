"""Connecting to a plugin that is reachable through streams, or served in this process."""

from __future__ import annotations

import os
import threading
from dataclasses import dataclass
from typing import Any, BinaryIO

from ..serve import ByteSource, ServeStreams, serve_streams
from .host_options_contract import HostOptions
from .plugin_client_client import PluginClient, PluginConnection


@dataclass
class PluginStreams:
    """The streams of a plugin that is somewhere else: its responses to read, its requests to write."""

    input: ByteSource
    """What the plugin writes (its standard output)."""
    output: BinaryIO
    """What the plugin reads (its standard input)."""


def connect_plugin(streams: PluginStreams, options: HostOptions) -> PluginClient:
    """Perform the handshake with a plugin reachable through a pair of streams.

    Closing the client shuts the plugin down and ends the streams; there is no process to kill.
    """

    def write(data: bytes) -> None:
        streams.output.write(data)
        streams.output.flush()

    def close_input() -> None:
        try:
            streams.output.close()
        except OSError:
            pass

    return PluginClient.open(
        PluginConnection(
            input=streams.input,
            write=write,
            close_input=close_input,
            kill=close_input,
            finish=lambda _grace: None,
        ),
        options,
    )


def connect_in_process(plugin: Any, options: HostOptions) -> PluginClient:
    """Serve a plugin object in this process over pipes and connect to it.

    It is the way the Go SDK's tests serve a plugin with `serve.Serve` over `io.Pipe`: the real
    framing, handshake and dispatch, on worker threads, with no process. Closing the client waits for
    the serve loop to end.
    """
    to_plugin_read, to_plugin_write = os.pipe()
    from_plugin_read, from_plugin_write = os.pipe()
    plugin_input = os.fdopen(to_plugin_read, "rb", buffering=0)
    plugin_output = os.fdopen(from_plugin_write, "wb", buffering=0)
    host_input = os.fdopen(from_plugin_read, "rb", buffering=0)
    host_output = os.fdopen(to_plugin_write, "wb", buffering=0)
    outcome: list[str | None] = [None]

    def run() -> None:
        try:
            serve_streams(plugin, ServeStreams(input=plugin_input, output=plugin_output, log=lambda _message: None))
        except Exception as error:
            outcome[0] = str(error) or type(error).__name__
        finally:
            plugin_output.close()
            plugin_input.close()

    thread = threading.Thread(target=run, name="smith-plugin-in-process", daemon=True)
    thread.start()

    def close_input() -> None:
        try:
            host_output.close()
        except OSError:
            pass

    def finish(grace: float) -> str | None:
        thread.join(grace)
        return outcome[0] if not thread.is_alive() else "the plugin did not stop after shutdown"

    def write(data: bytes) -> None:
        host_output.write(data)

    return PluginClient.open(
        PluginConnection(
            input=host_input,
            write=write,
            close_input=close_input,
            kill=close_input,
            finish=finish,
            dispose=host_input.close,
        ),
        options,
    )
