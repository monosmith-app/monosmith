"""Encoding and decoding the params and the result of each method of the protocol."""

from __future__ import annotations

from typing import Any

from ..protocol import PROTOCOL_SCHEMA
from .encode_by_schema_algorithm import encode_by_schema
from .encode_json_value_algorithm import encode_json_value
from .normalize_by_schema_algorithm import normalize_by_schema


def _side(method: str, side: str) -> dict[str, Any] | None:
    rows = PROTOCOL_SCHEMA["x-methods"]
    row = rows.get(method) if isinstance(method, str) else None
    node = None if row is None else row.get(side)
    return node if isinstance(node, dict) else None


def _encode(method: str, side: str, value: Any) -> str:
    node = _side(method, side)
    return encode_json_value(value, side) if node is None else encode_by_schema(PROTOCOL_SCHEMA, node, value, side)


def _decode(method: str, side: str, value: Any) -> Any:
    node = _side(method, side)
    return value if node is None else normalize_by_schema(PROTOCOL_SCHEMA, node, value, side)


def encode_params(method: str, params: Any) -> str:
    """The params of a request, written as the Go SDK writes them.

    A method the protocol does not know is written as plain JSON.
    """
    return _encode(method, "params", params)


def encode_result(method: str, result: Any) -> str:
    """The result of a method, written as the Go SDK writes it."""
    return _encode(method, "result", result)


def decode_params(method: str, params: Any) -> Any:
    """The params of a received request, shaped as the generated types promise.

    Absent params are an empty object. Raises `ParamsError` for params of the wrong shape.
    """
    return _decode(method, "params", {} if params is None else params)


def decode_result(method: str, result: Any) -> Any:
    """The result of a received response, shaped as the generated types promise."""
    return _decode(method, "result", {} if result is None else result)
