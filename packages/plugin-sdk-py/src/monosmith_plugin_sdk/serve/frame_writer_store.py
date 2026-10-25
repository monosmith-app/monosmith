"""The one writer of frames, so answers never interleave."""

from __future__ import annotations

import threading

from ..framing import frame_message
from .protocol_streams_contract import ByteSink


class FrameWriter:
    """Writes whole frames to the output, one at a time, from any thread."""

    def __init__(self, output: ByteSink) -> None:
        self._output = output
        self._lock = threading.Lock()

    def send(self, body: str) -> None:
        """Write one message as a frame and flush it. Raises what the stream raises (a broken pipe)."""
        data = frame_message(body)
        with self._lock:
            while data:
                written = self._output.write(data)
                data = data[len(data) if written is None else written :]
            flush = getattr(self._output, "flush", None)
            if flush is not None:
                flush()
