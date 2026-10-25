"""Writing a value as the JSON the Go SDK writes for the same message."""

from __future__ import annotations

import math
from collections.abc import Mapping
from typing import Any

from .encode_error import EncodeError
from .encode_json_value_algorithm import encode_json_string, encode_json_value, format_number
from .schema_lookup_algorithm import SchemaNode, base_node, is_nullable, kind_of


def encode_by_schema(schema: SchemaNode, node: SchemaNode, value: Any, path: str) -> str:
    """Write a value as the JSON the Go SDK would write for the same message.

    The members of a message come in the order the schema lists them (the order of the Go
    fields); an optional member is left out when it is absent or holds the zero value Go's
    `omitempty` drops (`""`, `0`, `false`, an empty list or dict); the keys of a map are sorted;
    HTML characters are escaped. So the bytes a Python plugin writes equal the bytes a Go
    plugin writes for the same answer.

    Raises `EncodeError` for a required member that is missing or a value of the wrong type.
    """
    kind = kind_of(schema, node)
    if value is None:
        if is_nullable(node) or kind == "any":
            return "null"
        raise EncodeError(f"{path} is required and cannot be null")
    base = base_node(schema, node)
    if kind == "string":
        if not isinstance(value, str):
            raise _wrong_type(path, "a string", value)
        return encode_json_string(value)
    if kind == "integer":
        if not _is_integer(value):
            raise _wrong_type(path, "an integer", value)
        return str(value)
    if kind == "number":
        if isinstance(value, bool) or not isinstance(value, int | float) or not math.isfinite(value):
            raise _wrong_type(path, "a number", value)
        return format_number(value, path)
    if kind == "boolean":
        if not isinstance(value, bool):
            raise _wrong_type(path, "a boolean", value)
        return "true" if value else "false"
    if kind == "array":
        if not isinstance(value, list | tuple):
            raise _wrong_type(path, "an array", value)
        items = base.get("items") or {}
        return (
            "[" + ",".join(encode_by_schema(schema, items, item, f"{path}[{i}]") for i, item in enumerate(value)) + "]"
        )
    if kind == "map":
        if not isinstance(value, Mapping):
            raise _wrong_type(path, "an object", value)
        values = base.get("additionalProperties")
        member_schema: SchemaNode = values if isinstance(values, dict) else {}
        members = [
            f"{encode_json_string(str(name))}:{encode_by_schema(schema, member_schema, member, f'{path}.{name}')}"
            for name, member in sorted(value.items(), key=lambda item: str(item[0]))
        ]
        return "{" + ",".join(members) + "}"
    if kind == "struct":
        if not isinstance(value, Mapping):
            raise _wrong_type(path, "an object", value)
        return _encode_struct(schema, base, value, path)
    return encode_json_value(value, path)


def _encode_struct(schema: SchemaNode, base: SchemaNode, value: Mapping[str, Any], path: str) -> str:
    properties: dict[str, SchemaNode] = base.get("properties") or {}
    required = set(base.get("required") or [])
    members: list[str] = []
    for name, property_schema in properties.items():
        member = value.get(name)
        if name not in required and _is_omitted(schema, property_schema, member):
            continue
        if name in required and name not in value:
            raise EncodeError(f"{path}.{name} is required")
        members.append(
            f"{encode_json_string(name)}:{encode_by_schema(schema, property_schema, member, f'{path}.{name}')}"
        )
    # Members the schema does not know (a protocol that grew past this SDK) are kept, after the known ones.
    for name in sorted(str(n) for n in value if n not in properties and value[n] is not None):
        members.append(f"{encode_json_string(name)}:{encode_json_value(value[name], f'{path}.{name}')}")
    return "{" + ",".join(members) + "}"


def _is_omitted(schema: SchemaNode, property_schema: SchemaNode, member: Any) -> bool:
    """Go's `omitempty`: absent, null, or the zero value of a scalar, array or map. A struct is never empty."""
    if member is None:
        return True
    kind = kind_of(schema, property_schema)
    if kind == "string":
        return bool(member == "")
    if kind in ("integer", "number"):
        return not isinstance(member, bool) and member == 0
    if kind == "boolean":
        return member is False
    if kind == "array":
        return isinstance(member, list | tuple) and len(member) == 0
    return bool(kind == "map" and isinstance(member, Mapping) and len(member) == 0)


def _is_integer(value: Any) -> bool:
    return isinstance(value, int) and not isinstance(value, bool)


def _wrong_type(path: str, expected: str, value: Any) -> EncodeError:
    return EncodeError(f"{path} must be {expected}, not {describe_type(value)}")


def describe_type(value: Any) -> str:
    """How an error names the type of a JSON value."""
    if isinstance(value, bool):
        return "a boolean"
    if isinstance(value, str):
        return "a string"
    if isinstance(value, int | float):
        return "a number"
    if isinstance(value, list | tuple):
        return "an array"
    if isinstance(value, Mapping):
        return "an object"
    return f"a {type(value).__name__}"
