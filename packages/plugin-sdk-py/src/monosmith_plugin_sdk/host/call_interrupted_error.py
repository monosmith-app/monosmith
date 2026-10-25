"""The error of a call the host gave up on."""

from __future__ import annotations

from typing import Literal

from ..protocol import RpcError

InterruptionReason = Literal["cancelled", "timeout"]
"""Why the host stopped waiting for a call."""


class CallInterruptedError(Exception):
    """A call the host gave up on, because its token was cancelled (`cancelled`) or its time ran out
    (`timeout`).

    `killed` says the plugin did not stop in time and was killed, which also fails every other call in
    flight to it. When the plugin answered while being stopped, `response` is its error (the
    request-cancelled code, usually).
    """

    def __init__(
        self, message: str, reason: InterruptionReason, killed: bool, response: RpcError | None = None
    ) -> None:
        super().__init__(message)
        self.reason: InterruptionReason = reason
        self.killed = killed
        self.response = response
        if response is not None:
            self.__cause__ = response
