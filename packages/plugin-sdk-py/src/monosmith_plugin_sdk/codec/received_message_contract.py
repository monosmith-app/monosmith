"""The messages as they were read from a frame."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Any


@dataclass(frozen=True)
class ReceivedRequest:
    """A request or notification as it was received."""

    id: int | None
    """The request id; `None` for a notification, which gets no response."""
    method: str
    """The method name; empty when the message carried none."""
    params: Any
    """The params as the JSON parser returned them; `None` when absent."""


@dataclass(frozen=True)
class ReceivedError:
    """The error object of a response."""

    code: int
    message: str


@dataclass(frozen=True)
class ReceivedResponse:
    """A response as it was received."""

    id: int | None
    """The id of the request answered; `None` when the request could not be read."""
    result: Any
    """The result, when the response is not an error."""
    error: ReceivedError | None
    """The error, when the response is one."""
