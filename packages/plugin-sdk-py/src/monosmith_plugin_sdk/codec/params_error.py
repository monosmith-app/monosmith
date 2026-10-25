"""The error of received params, or a received result, of the wrong shape."""

from __future__ import annotations


class ParamsError(Exception):
    """A received message whose params (or result) do not have the shape the schema describes.

    The serve loop answers it with the code `INVALID_PARAMS`.
    """
