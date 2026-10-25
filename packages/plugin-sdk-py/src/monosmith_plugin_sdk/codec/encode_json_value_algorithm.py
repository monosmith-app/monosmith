"""Writing JSON the way Go's `encoding/json` writes it."""

from __future__ import annotations

import json
import math
import re
from collections.abc import Mapping
from decimal import Decimal
from typing import Any

from .encode_error import EncodeError

_LONE_SURROGATE = re.compile("[\ud800-\udfff]")
_GO_ESCAPES = {
    "<": "\\u003c",
    ">": "\\u003e",
    "&": "\\u0026",
    " ": "\\u2028",
    " ": "\\u2029",
}
_GO_ESCAPED = re.compile("[<>&  ]")


def encode_json_string(text: str) -> str:
    """A string as the Go SDK writes it.

    JSON escaping with the characters left as they are (never `\\uXXXX` for a non-ASCII
    letter), and `<`, `>`, `&`, U+2028 and U+2029 as `\\u` escapes, because Go's `json.Marshal`
    escapes HTML by default. Both forms decode to the same text; writing Go's makes the bytes
    of the two SDKs equal. A lone surrogate, which has no UTF-8 form, becomes U+FFFD as Go
    writes invalid UTF-8.
    """
    cleaned = _LONE_SURROGATE.sub("�", text)
    return _GO_ESCAPED.sub(lambda match: _GO_ESCAPES[match.group()], json.dumps(cleaned, ensure_ascii=False))


def format_number(value: int | float, path: str) -> str:
    """A number as Go writes it: an integer as digits, a float in its shortest form."""
    if isinstance(value, int):
        return str(value)
    if not math.isfinite(value):
        raise EncodeError(f"{path}: {value} cannot be written as JSON")
    if value == 0:
        return "-0" if math.copysign(1.0, value) < 0 else "0"
    magnitude = abs(value)
    shortest = Decimal(repr(value))
    if magnitude < 1e-6 or magnitude >= 1e21:
        digits = shortest.as_tuple().digits
        exponent = shortest.adjusted()
        mantissa = f"{digits[0]}" + (f".{''.join(map(str, digits[1:])).rstrip('0')}" if len(digits) > 1 else "")
        mantissa = mantissa.removesuffix(".")
        sign = "-" if value < 0 else ""
        # Go writes the exponent with at least two digits, except e-0N, which it cleans up to e-N.
        exponent_text = f"{abs(exponent):02d}"
        if exponent < 0 and exponent_text.startswith("0"):
            exponent_text = exponent_text[1:]
        return f"{sign}{mantissa}e{'-' if exponent < 0 else '+'}{exponent_text}"
    text = format(shortest, "f")
    return text.rstrip("0").rstrip(".") if "." in text else text


def encode_json_value(value: Any, path: str) -> str:
    """Any JSON value as the Go SDK writes a `map[string]any`: object keys sorted by their UTF-8 bytes."""
    if value is None:
        return "null"
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, str):
        return encode_json_string(value)
    if isinstance(value, int | float):
        return format_number(value, path)
    if isinstance(value, Mapping):
        members = [
            f"{encode_json_string(str(name))}:{encode_json_value(member, f'{path}.{name}')}"
            for name, member in sorted(value.items(), key=lambda item: str(item[0]))
        ]
        return "{" + ",".join(members) + "}"
    if isinstance(value, list | tuple):
        return "[" + ",".join(encode_json_value(item, f"{path}[{i}]") for i, item in enumerate(value)) + "]"
    raise EncodeError(f"{path}: a {type(value).__name__} cannot be written as JSON")
