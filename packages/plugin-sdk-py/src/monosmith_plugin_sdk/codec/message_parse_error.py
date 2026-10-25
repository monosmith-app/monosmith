"""The error of a frame body that is not a JSON-RPC 2.0 message."""

from __future__ import annotations


class MessageParseError(Exception):
    """A frame whose body is not a JSON-RPC 2.0 message of the shape the protocol uses.

    The serve loop answers it with the code `PARSE_ERROR`.
    """
