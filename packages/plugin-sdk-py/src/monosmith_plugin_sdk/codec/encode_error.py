"""The error of a value that cannot be written as the message the schema describes."""

from __future__ import annotations


class EncodeError(Exception):
    """A value that cannot be written as the message the schema describes.

    A required member is missing, or a member has the wrong type. The serve loop answers it
    with the code `INTERNAL_ERROR`.
    """
