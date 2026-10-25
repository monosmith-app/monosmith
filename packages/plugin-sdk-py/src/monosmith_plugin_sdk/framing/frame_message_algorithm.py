"""Writing one message as a frame."""

from __future__ import annotations


def frame_message(body: str) -> bytes:
    """One message as the protocol frames it: `Content-Length: <bytes>`, a blank line (CRLF CRLF), the body.

    The length counts bytes of UTF-8, not characters.
    """
    data = body.encode("utf-8", errors="replace")
    return b"Content-Length: " + str(len(data)).encode("ascii") + b"\r\n\r\n" + data
