"""The functions of `text/template` that the supported subset has, and the way it prints values."""

from __future__ import annotations

import json
import re
from collections.abc import Callable, Mapping
from typing import Any

TemplateFunction = Callable[[list[Any]], Any]
"""A template function receives its evaluated arguments (the piped value last)."""


def is_object(value: Any) -> bool:
    """Whether the value is a map."""
    return isinstance(value, Mapping)


def by_key(key: str) -> bytes:
    """The sort key for the order Go sorts the keys of a map in (by their bytes)."""
    return key.encode("utf-8")


def is_truthy(value: Any) -> bool:
    """Go's truth: false, 0, nil, and an empty string, list or map are false; everything else is true."""
    if value is None or isinstance(value, bool):
        return bool(value)
    if isinstance(value, int | float):
        return value != 0
    if isinstance(value, str | list | tuple | Mapping):
        return len(value) > 0
    return True


def format_value(value: Any) -> str:
    """A value as `text/template` prints it."""
    if value is None:
        return "<no value>"
    if isinstance(value, bool):
        return "true" if value else "false"
    if isinstance(value, list | tuple):
        return "[" + " ".join(format_value(item) for item in value) + "]"
    if isinstance(value, Mapping):
        entries = [f"{key}:{format_value(value[key])}" for key in sorted(value, key=by_key)]
        return "map[" + " ".join(entries) + "]"
    return str(value)


def _comparable(a: Any, b: Any) -> str:
    kind = _scalar_kind(a)
    if kind is None or _scalar_kind(b) != kind:
        raise TypeError("incompatible types for comparison")
    return kind


def _scalar_kind(value: Any) -> str | None:
    if isinstance(value, bool):
        return "boolean"
    if isinstance(value, str):
        return "string"
    if isinstance(value, int | float):
        return "number"
    return None


def _ordered(test: Callable[[int], bool]) -> TemplateFunction:
    def compare(args: list[Any]) -> bool:
        a, b = args
        if _comparable(a, b) == "boolean":
            raise TypeError("invalid type for comparison")
        return test(0 if a == b else (1 if a > b else -1))

    return compare


def _index_into(current: Any, key: Any) -> Any:
    if isinstance(current, list | tuple):
        if isinstance(key, bool) or not isinstance(key, int) or key < 0 or key >= len(current):
            raise IndexError(f"index out of range: {format_value(key)}")
        return current[key]
    if is_object(current):
        return current.get(str(key))
    raise TypeError(f"can't index {format_value(current)}")


def _sprint(args: list[Any]) -> str:
    """What Sprint puts between operands: a space when neither side is a string."""
    out = ""
    for i, arg in enumerate(args):
        if i > 0 and not isinstance(arg, str) and not isinstance(args[i - 1], str):
            out += " "
        out += format_value(arg)
    return out


def _eq(args: list[Any]) -> bool:
    first, *others = args
    if not others:
        raise TypeError("missing argument for comparison")
    for other in others:
        if first is None or other is None:
            if first is other:
                return True
            continue
        _comparable(first, other)
        if first == other:
            return True
    return False


def _ne(args: list[Any]) -> bool:
    a, b = args
    if a is None or b is None:
        return a is not b
    _comparable(a, b)
    return bool(a != b)


def _len(args: list[Any]) -> int:
    value = args[0]
    if isinstance(value, str):
        return len(value.encode("utf-8"))
    if isinstance(value, list | tuple | Mapping):
        return len(value)
    raise TypeError(f"len of {format_value(value)}")


def _index(args: list[Any]) -> Any:
    value, *keys = args
    for key in keys:
        value = _index_into(value, key)
    return value


def _printf(args: list[Any]) -> str:
    pattern, *rest = args
    if not isinstance(pattern, str):
        raise TypeError("printf needs a format string")
    remaining = iter(rest)

    def replace(match: re.Match[str]) -> str:
        verb = match.group(1)
        if verb == "%":
            return "%"
        try:
            arg = next(remaining)
        except StopIteration:
            raise IndexError(f"missing argument for %{verb}") from None
        return json.dumps(str(arg), ensure_ascii=False) if verb == "q" else format_value(arg)

    return re.sub(r"%([%sdvqt])", replace, pattern)


def _and(args: list[Any]) -> Any:
    return next((arg for arg in args if not is_truthy(arg)), args[-1])


def _or(args: list[Any]) -> Any:
    return next((arg for arg in args if is_truthy(arg)), args[-1])


TEMPLATE_FUNCTIONS: dict[str, TemplateFunction] = {
    "eq": _eq,
    "ne": _ne,
    "lt": _ordered(lambda difference: difference < 0),
    "le": _ordered(lambda difference: difference <= 0),
    "gt": _ordered(lambda difference: difference > 0),
    "ge": _ordered(lambda difference: difference >= 0),
    "not": lambda args: not is_truthy(args[0]),
    "and": _and,
    "or": _or,
    "len": _len,
    "index": _index,
    "print": _sprint,
    "printf": _printf,
}
"""The functions of `text/template` the subset has."""
