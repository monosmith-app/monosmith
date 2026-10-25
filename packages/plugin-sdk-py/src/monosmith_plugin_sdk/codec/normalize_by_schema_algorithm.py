"""Shaping a received value the way the Go SDK's decoder does."""

from __future__ import annotations

from typing import Any

from .encode_by_schema_algorithm import describe_type
from .params_error import ParamsError
from .schema_lookup_algorithm import SchemaNode, base_node, is_nullable, kind_of

MAX_DEPTH = 64

_SCALARS = {"string": "a string", "integer": "an integer", "number": "a number", "boolean": "a boolean"}


def _is_scalar(kind: str, value: Any) -> bool:
    if kind == "string":
        return isinstance(value, str)
    if kind == "integer":
        return isinstance(value, int) and not isinstance(value, bool)
    if kind == "number":
        return isinstance(value, int | float) and not isinstance(value, bool)
    return isinstance(value, bool)


def normalize_by_schema(schema: SchemaNode, node: SchemaNode, value: Any, path: str, depth: int = 0) -> Any:
    """Shape a received value so that the types the generated `TypedDict`s promise hold.

    - a required member that is absent or `null` is its zero value (`[]`, `{}`, `""`, `0`,
      `False`): the Go host writes a nil slice as `null`;
    - an optional member that is `null` is removed;
    - members the schema does not know stay as they are.

    Raises `ParamsError` where a value has the wrong JSON type. The value is edited in place
    and returned.
    """
    if depth > MAX_DEPTH:
        raise ParamsError(f"{path} is nested too deeply")
    if value is None:
        zero = _zero_of(schema, node, path, depth)
        return None if zero is _NO_ZERO else zero
    kind = kind_of(schema, node)
    base = base_node(schema, node)
    if kind in _SCALARS:
        if not _is_scalar(kind, value):
            raise _wrong_type(path, _SCALARS[kind], value)
        return value
    if kind == "array":
        if not isinstance(value, list):
            raise _wrong_type(path, "an array", value)
        items = base.get("items") or {}
        return [normalize_by_schema(schema, items, item, f"{path}[{i}]", depth + 1) for i, item in enumerate(value)]
    if kind == "map":
        return _normalize_map(schema, base, value, path, depth)
    if kind == "struct":
        return _normalize_struct(schema, base, value, path, depth)
    return value


def _normalize_map(schema: SchemaNode, base: SchemaNode, value: Any, path: str, depth: int) -> Any:
    if not isinstance(value, dict):
        raise _wrong_type(path, "an object", value)
    values = base.get("additionalProperties")
    member_schema: SchemaNode = values if isinstance(values, dict) else {}
    keeps_null = kind_of(schema, member_schema) == "any"
    for name in list(value):
        member = value[name]
        if member is not None:
            value[name] = normalize_by_schema(schema, member_schema, member, f"{path}.{name}", depth + 1)
        elif not keeps_null:
            del value[name]
    return value


def _normalize_struct(schema: SchemaNode, base: SchemaNode, value: Any, path: str, depth: int) -> Any:
    if not isinstance(value, dict):
        raise _wrong_type(path, "an object", value)
    required = set(base.get("required") or [])
    for name, property_schema in (base.get("properties") or {}).items():
        member = value.get(name)
        where = f"{path}.{name}"
        if member is not None:
            value[name] = normalize_by_schema(schema, property_schema, member, where, depth + 1)
        elif name in required:
            zero = _zero_of(schema, property_schema, where, depth + 1)
            if zero is not _NO_ZERO:
                value[name] = zero
        else:
            value.pop(name, None)
    return value


_NO_ZERO: Any = object()


def _zero_of(schema: SchemaNode, node: SchemaNode, path: str, depth: int) -> Any:
    """The value Go's decoder leaves in a member the message did not carry."""
    if is_nullable(node):
        return None
    kind = kind_of(schema, node)
    zeros: dict[str, Any] = {"string": "", "integer": 0, "number": 0, "boolean": False}
    if kind in zeros:
        return zeros[kind]
    if kind == "array":
        return []
    if kind == "map":
        return {}
    if kind == "struct":
        return normalize_by_schema(schema, node, {}, path, depth + 1)
    return _NO_ZERO


def _wrong_type(path: str, expected: str, value: Any) -> ParamsError:
    return ParamsError(f"{path} must be {expected}, not {describe_type(value)}")
