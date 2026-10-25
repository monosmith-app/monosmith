"""The streams the serve loop runs on."""

from __future__ import annotations

from collections.abc import Callable
from dataclasses import dataclass
from typing import Protocol


class ByteSource(Protocol):
    """Where the host's requests come from: a binary stream that returns what is available.

    `read1` is used when the stream has it (a `BufferedReader`); otherwise `read` must return as
    soon as some bytes are available, as an unbuffered file or a pipe does.
    """

    def read(self, size: int = ..., /) -> bytes: ...


class ByteSink(Protocol):
    """Where the plugin's frames go: a binary stream."""

    def write(self, data: bytes, /) -> int | None: ...


@dataclass
class ServeStreams:
    """The streams `serve_streams` runs on."""

    input: ByteSource
    """The host's requests, as bytes."""
    output: ByteSink
    """The plugin's responses."""
    log: Callable[[str], None] | None = None
    """Where a warning about the plugin goes. Default: standard error."""
