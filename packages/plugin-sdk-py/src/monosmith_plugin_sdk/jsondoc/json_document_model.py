"""A JSON object that keeps its key order and the text of everything you do not touch."""

from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any

from .indent_raw_json_algorithm import indent_raw_json
from .json_document_error import JsonDocumentError
from .scan_json_object_algorithm import ObjectNode, parse_json_text, scan_json_object


@dataclass
class _Raw:
    text: str


def _quote(text: str) -> str:
    """A string as Go's JSON encoder writes it without HTML escaping: JSON, with U+2028 and U+2029 escaped."""
    return json.dumps(text, ensure_ascii=False).replace("\u2028", "\\u2028").replace("\u2029", "\\u2029")


class JsonDocument:
    """A JSON object that keeps its key order and the text of everything you do not touch.

    So editing a `package.json` changes only what was edited: set a member and it keeps its place,
    add one and it goes last, delete one and the others stay where they were. It is a port of the
    Go SDK's `jsondoc.Document`.

        doc = JsonDocument.parse(text)
        doc.set_string("version", "1.2.3")
        path.write_text(doc.render())
    """

    def __init__(self) -> None:
        self._entries: dict[str, JsonDocument | _Raw] = {}

    @classmethod
    def parse(cls, text: str) -> JsonDocument:
        """Read a JSON object. Raises `JsonDocumentError` for anything else."""
        return cls._from_node(scan_json_object(text))

    @classmethod
    def _from_node(cls, node: ObjectNode) -> JsonDocument:
        document = cls()
        for key, value in node.members:
            document._entries[key] = cls._from_node(value) if isinstance(value, ObjectNode) else _Raw(value.text)
        return document

    def _write(self, indent: str) -> str:
        if not self._entries:
            return "{}"
        inner = f"{indent}  "
        lines = [
            f"{inner}{_quote(key)}: "
            + (entry._write(inner) if isinstance(entry, JsonDocument) else indent_raw_json(entry.text, inner))
            for key, entry in self._entries.items()
        ]
        return "{\n" + ",\n".join(lines) + f"\n{indent}}}"

    def has(self, key: str) -> bool:
        """Whether the key exists."""
        return key in self._entries

    def keys(self) -> list[str]:
        """The keys, in order."""
        return list(self._entries)

    def __len__(self) -> int:
        return len(self._entries)

    def object(self, key: str, create: bool = False) -> JsonDocument | None:
        """The object under a key, editable in place.

        `create` adds an empty one (last) when there is none. `None` for a key that holds something else.
        """
        entry = self._entries.get(key)
        if isinstance(entry, JsonDocument):
            return entry
        if not create:
            return None
        child = JsonDocument()
        self._entries[key] = child
        return child

    def get_string(self, key: str) -> str | None:
        """A string value; `None` when the key is missing or does not hold a string."""
        entry = self._entries.get(key)
        if entry is None or isinstance(entry, JsonDocument):
            return None
        value = parse_json_text(entry.text)
        return value if isinstance(value, str) else None

    def get_value(self, key: str) -> Any:
        """Any value under a key as Python sees it, or `None` for a missing key."""
        entry = self._entries.get(key)
        if entry is None:
            return None
        return parse_json_text(entry.render() if isinstance(entry, JsonDocument) else entry.text)

    def set_string(self, key: str, value: str) -> None:
        """Set a string, keeping the key's position."""
        self._entries[key] = _Raw(_quote(value))

    def set_json(self, key: str, value: Any) -> None:
        """Set any JSON-encodable value, keeping the key's position.

        A dict becomes an editable document, with its members in the order you wrote them.
        """
        try:
            text = json.dumps(value, ensure_ascii=False, separators=(",", ":"), allow_nan=False)
        except (TypeError, ValueError) as error:
            raise JsonDocumentError(f"{key}: {error}") from error
        if isinstance(value, dict):
            self._entries[key] = JsonDocument.parse(text)
        else:
            self._entries[key] = _Raw(text)

    def delete(self, key: str) -> None:
        """Remove a key."""
        self._entries.pop(key, None)

    def render(self) -> str:
        """The document as npm writes a `package.json`: two-space indentation, a final newline."""
        return f"{self._write('')}\n"

    def __str__(self) -> str:
        return self.render()
