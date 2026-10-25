"""Reading the text of a JSON object into a tree that keeps what you do not touch."""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from typing import Any

from .json_document_error import JsonDocumentError

_WHITESPACE = frozenset(" \t\n\r")


@dataclass
class ObjectNode:
    """A JSON object edited member by member."""

    members: list[tuple[str, ObjectNode | RawNode]] = field(default_factory=list)


@dataclass
class RawNode:
    """Anything else, as the text it was written in."""

    text: str


JsonNode = ObjectNode | RawNode
"""A JSON value as the document keeps it: an object, or the text of anything else."""


def _reject_constant(name: str) -> Any:
    raise ValueError(f"{name} is not JSON")


def parse_json_text(text: str) -> Any:
    """`json.loads`, but `NaN` and `Infinity` are not JSON."""
    return json.loads(text, parse_constant=_reject_constant)


def scan_json_object(text: str) -> ObjectNode:
    """Read the text of a JSON object into a tree that keeps the order of its members and the exact
    text of every leaf and array (`1.0` stays `1.0`).

    A duplicate key keeps its first position and takes its last value, as the Go SDK's document does.
    """
    scanner = _Scanner(text)
    scanner.skip()
    if scanner.at() != "{":
        raise JsonDocumentError("not a JSON object")
    root = scanner.object()
    scanner.skip()
    if scanner.position < len(text):
        raise JsonDocumentError(f'unexpected "{text[scanner.position]}" after the object, at {scanner.position}')
    return root


class _Scanner:
    def __init__(self, text: str) -> None:
        self._text = text
        self.position = 0

    def at(self) -> str:
        return self._text[self.position] if self.position < len(self._text) else ""

    def skip(self) -> None:
        while self.position < len(self._text) and self._text[self.position] in _WHITESPACE:
            self.position += 1

    def object(self) -> ObjectNode:
        members: list[tuple[str, JsonNode]] = []
        index: dict[str, int] = {}
        self.position += 1  # {
        self.skip()
        if self.at() == "}":
            self.position += 1
            return ObjectNode(members)
        while True:
            self.skip()
            if self.at() != '"':
                raise self.fail("an object key")
            key = json.loads(self.string())
            self.skip()
            if self.at() != ":":
                raise self.fail('":"')
            self.position += 1
            value = self.value()
            at = index.get(key)
            if at is None:
                index[key] = len(members)
                members.append((key, value))
            else:
                members[at] = (key, value)
            self.skip()
            following = self.at()
            if following == "}":
                self.position += 1
                return ObjectNode(members)
            if following != ",":
                raise self.fail('"," or "}"')
            self.position += 1

    def value(self) -> JsonNode:
        self.skip()
        first = self.at()
        if first == "{":
            return self.object()
        start = self.position
        if first == "[":
            self.array()
        elif first == '"':
            self.string()
        else:
            self.literal()
        text = self._text[start : self.position]
        try:
            parse_json_text(text)
        except ValueError as error:
            raise JsonDocumentError(f'invalid JSON value "{text[:40]}" at {start}') from error
        return RawNode(text)

    def array(self) -> None:
        """Move past an array, whose elements stay text."""
        depth = 0
        while self.position < len(self._text):
            c = self._text[self.position]
            if c == '"':
                self.string()
                continue
            if c in "[{":
                depth += 1
            elif c in "]}":
                depth -= 1
            self.position += 1
            if depth == 0:
                return
        raise JsonDocumentError("unterminated array")

    def string(self) -> str:
        """Move past a string literal and return its text, quotes included."""
        start = self.position
        self.position += 1
        while self.position < len(self._text):
            c = self._text[self.position]
            self.position += 1
            if c == "\\":
                self.position += 1
            elif c == '"':
                return self._text[start : self.position]
        raise JsonDocumentError("unterminated string")

    def literal(self) -> None:
        start = self.position
        while (
            self.position < len(self._text)
            and self._text[self.position] not in _WHITESPACE
            and self._text[self.position] not in ",}]"
        ):
            self.position += 1
        if self.position == start:
            raise self.fail("a value")

    def fail(self, expected: str) -> JsonDocumentError:
        found = f'"{self._text[self.position]}"' if self.position < len(self._text) else "the end of the text"
        return JsonDocumentError(f"expected {expected}, found {found} at {self.position}")
