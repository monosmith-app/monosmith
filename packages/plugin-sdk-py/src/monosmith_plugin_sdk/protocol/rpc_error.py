"""The JSON-RPC 2.0 error object as an exception."""

from __future__ import annotations


class RpcError(Exception):
    """A JSON-RPC 2.0 error object as an exception.

    It is what a host's call raises when the plugin answers an error, and its `code` is one
    of `ERROR_CODES`. The serve loop answers any exception a plugin method raises with the
    code `PLUGIN_FAILURE`, whatever it is.
    """

    def __init__(self, code: int, message: str) -> None:
        super().__init__(message)
        self.code = code
        """The numeric error code."""
        self.message = message
        """The text of the error, as the plugin wrote it."""

    def __str__(self) -> str:
        return self.message
