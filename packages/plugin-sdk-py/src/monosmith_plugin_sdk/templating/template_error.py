"""The error of a template that cannot be parsed or rendered."""

from __future__ import annotations


class TemplateError(Exception):
    """A template that cannot be parsed or rendered.

    Bad syntax, a field the data does not have, a construct outside the supported subset.
    """

    def __init__(self, name: str, message: str) -> None:
        super().__init__(f"template {name}: {message}")
