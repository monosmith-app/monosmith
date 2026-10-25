"""The handshake."""

from __future__ import annotations

from typing import Any, cast

from ..protocol import PROTOCOL_VERSION, InitializeParams, InitializeResult
from .capabilities_of_algorithm import capabilities_of


def initialize_plugin(plugin: Any, params: InitializeParams) -> InitializeResult:
    """Answer `initialize`.

    It refuses a host that does not speak this SDK's protocol version (the message is the Go
    SDK's), asks the plugin to describe itself, and fills in the version it chose and the
    capabilities derived from the methods the plugin has, whatever `describe` returned for them.
    """
    versions = params["protocolVersions"]
    if PROTOCOL_VERSION not in versions:
        # The Go SDK formats the host's list with %v: [1 2].
        listed = " ".join(str(v) for v in versions)
        raise RuntimeError(f"no common protocol version: host speaks [{listed}], plugin speaks {PROTOCOL_VERSION}")
    description = plugin.describe(params)
    result: dict[str, Any] = {
        **description,
        "protocolVersion": PROTOCOL_VERSION,
        "capabilities": capabilities_of(plugin),
    }
    return cast(InitializeResult, result)
