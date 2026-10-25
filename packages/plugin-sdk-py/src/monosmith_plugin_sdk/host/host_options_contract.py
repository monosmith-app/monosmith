"""What the host tells the plugin, and how long it waits."""

from __future__ import annotations

from collections.abc import Mapping, Sequence
from dataclasses import dataclass
from typing import Literal

from ..protocol import Peer

DEFAULT_CANCEL_GRACE = 5.0
"""How long a plugin has to answer after it was asked to stop, in seconds, before the host kills it."""

DEFAULT_CALL_TIMEOUT = 2 * 60.0
"""The longest a call may take unless the caller says otherwise, in seconds (generate and publish run toolchains)."""

DEFAULT_INITIALIZE_TIMEOUT = 30.0
"""How long `initialize` may take, in seconds."""


@dataclass(kw_only=True)
class HostOptions:
    """What the host tells the plugin, and how long it waits."""

    workspace_root: str
    """The absolute workspace root sent in `initialize`. A plugin that reads manifests resolves them against it."""
    host: Peer | None = None
    """The host's name and version. Default: `smith`, `dev`."""
    initialize_timeout: float = DEFAULT_INITIALIZE_TIMEOUT
    """How long `initialize` may take, in seconds."""
    cancel_grace: float = DEFAULT_CANCEL_GRACE
    """How long a plugin has to answer after it was asked to stop (or, for one that cannot be asked,
    after the caller cancelled) before it is killed, in seconds."""


@dataclass(kw_only=True)
class StartOptions(HostOptions):
    """Options of `start_plugin`: a plugin process to launch."""

    command: Sequence[str]
    """The plugin's argv: the program and its arguments."""
    env: Mapping[str, str] | None = None
    """The environment. Default: the current one. `SMITH_PLUGIN=1` is set in it unless it has a value, as smith does."""
    cwd: str | None = None
    """The plugin's working directory. Default: the current one."""
    stderr: Literal["inherit", "ignore"] = "inherit"
    """What happens to the plugin's standard error, where it logs. Default: it goes to this process's."""
