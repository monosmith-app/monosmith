"""Reading a schema node: what it holds, with `$ref` followed and `null` set aside."""

from __future__ import annotations

from typing import Any, Literal

SchemaNode = dict[str, Any]
"""A node of the JSON Schema, as the JSON parser returns it."""

NodeKind = Literal["string", "integer", "number", "boolean", "array", "struct", "map", "any"]
"""What a schema node holds, as far as the codec cares."""

REF_PREFIX = "#/$defs/"
MAX_REF_HOPS = 16


def resolve_node(schema: SchemaNode, node: SchemaNode) -> SchemaNode:
    """Follow `$ref` to the definition it names."""
    current = node
    hops = 0
    while "$ref" in current:
        name = str(current["$ref"])[len(REF_PREFIX) :]
        target = schema["$defs"].get(name)
        if target is None or hops > MAX_REF_HOPS:
            raise ValueError(f"the schema has no definition {current['$ref']}")
        current = target
        hops += 1
    return current


def is_nullable(node: SchemaNode) -> bool:
    """Whether `null` is a value of the node: `["integer", "null"]`, or an `anyOf` with null."""
    kind = node.get("type")
    if isinstance(kind, list):
        return "null" in kind
    return any(branch.get("type") == "null" for branch in node.get("anyOf") or [])


def without_null(schema: SchemaNode, node: SchemaNode) -> SchemaNode:
    """The node without its `null` alternative."""
    kind = node.get("type")
    if isinstance(kind, list):
        types = [t for t in kind if t != "null"]
        return {**node, "type": types[0] if len(types) == 1 else types}
    branches = [branch for branch in node.get("anyOf") or [] if branch.get("type") != "null"]
    return resolve_node(schema, branches[0]) if len(branches) == 1 else node


def base_node(schema: SchemaNode, node: SchemaNode) -> SchemaNode:
    """The node as the codec walks it: nullability removed, references followed."""
    return resolve_node(schema, without_null(schema, node))


def kind_of(schema: SchemaNode, node: SchemaNode) -> NodeKind:
    """The kind of value a node describes, after `$ref` and null are set aside."""
    base = base_node(schema, node)
    kind = base.get("type")
    if kind in ("string", "integer", "number", "boolean", "array"):
        return kind  # type: ignore[no-any-return]
    if kind == "object":
        return "map" if "properties" not in base else "struct"
    return "any"
