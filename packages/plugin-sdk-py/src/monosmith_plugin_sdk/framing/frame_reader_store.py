"""Cutting a byte stream into the bodies of its frames."""

from __future__ import annotations

from ..protocol import MAX_MESSAGE_BYTES
from .framing_error import FramingError


class FrameReader:
    """Cuts a byte stream into the bodies of its frames.

    Bytes arrive in chunks of any size; `push` returns every body a chunk completed. Headers
    end at a blank line (CRLF or LF alone, as the Go SDK reads them), `Content-Length` is
    matched without regard to case, and a header that is not `Name: value`, a length that is
    not a number from 0 to 64 MiB, or a header block without a length is a `FramingError`.
    """

    def __init__(self) -> None:
        self._buffer = bytearray()

    @property
    def pending(self) -> int:
        """Bytes received that are not part of a complete frame yet."""
        return len(self._buffer)

    def _read_header(self) -> tuple[int, int] | None:
        """The body length and where the body starts, or `None` while the header is incomplete."""
        length = -1
        position = 0
        while True:
            newline = self._buffer.find(b"\n", position)
            if newline == -1:
                return None
            line = bytes(self._buffer[position:newline]).decode("latin-1").rstrip("\r\n")
            position = newline + 1
            if line == "":
                break
            name, colon, value = line.partition(":")
            if not colon:
                raise FramingError(f'malformed header "{line}"')
            if name.strip().lower() != "content-length":
                continue
            text = value.strip()
            if not (text.isascii() and text.isdigit()) or int(text) > MAX_MESSAGE_BYTES:
                raise FramingError(f'invalid Content-Length "{value}"')
            length = int(text)
        if length < 0:
            raise FramingError("message without Content-Length")
        return length, position

    def push(self, chunk: bytes) -> list[str]:
        """Add the bytes that arrived and return the body of every frame they completed, in order."""
        self._buffer += chunk
        bodies: list[str] = []
        while True:
            header = self._read_header()
            if header is None:
                return bodies
            length, start = header
            end = start + length
            if len(self._buffer) < end:
                return bodies
            bodies.append(bytes(self._buffer[start:end]).decode("utf-8", errors="replace"))
            del self._buffer[:end]
