"""The error of a byte stream that is not a sequence of frames."""

from __future__ import annotations


class FramingError(Exception):
    """The bytes on the stream are not a sequence of frames.

    A malformed header, or a missing or invalid `Content-Length`. The stream cannot be trusted
    after it.
    """
