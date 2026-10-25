"""The error of text that is not a JSON object."""

from __future__ import annotations


class JsonDocumentError(Exception):
    """The text is not a JSON object, or not JSON at all."""
