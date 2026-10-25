"""Splitting the text of a template action into tokens."""

from __future__ import annotations

import json
import re
from collections.abc import Callable
from dataclasses import dataclass
from typing import NoReturn

Fail = Callable[[str], NoReturn]
"""Reports what cannot be parsed, by raising."""

# The elements of an action, in the order the alternatives are tried: space, brackets and pipe, a
# quoted or raw string, a number, a field (`.A.B`, `$`, `$.A`, or a lone `.`), a word, and `:=`/`=`,
# which are variables. Anything else is the last alternative, so every character belongs to some match.
_ELEMENT = re.compile(
    "|".join(
        [
            r"\s+",
            r"(?P<open>\()",
            r"(?P<close>\))",
            r"(?P<pipe>\|)",
            r'(?P<string>"(?:[^"\\]|\\.)*")',
            r"(?P<raw>`[^`]*`)",
            r"(?P<number>-?\d+(?:\.\d+)?(?![\w.]))",
            r"(?P<field>\$(?:\.[A-Z_]\w*)*|(?:\.[A-Z_]\w*)+|\.)",
            r"(?P<word>[A-Z_]\w*)",
            r"(?P<variable>:?=)",
            r"(?P<other>.)",
        ]
    ),
    re.IGNORECASE | re.DOTALL | re.ASCII,
)


@dataclass(frozen=True)
class Token:
    """One element of an action."""

    kind: str
    """`open`, `close`, `pipe`, `field`, `string`, `number` or `word`."""
    value: str | int | float = ""
    root: bool = False
    path: tuple[str, ...] = ()


def tokenize_action(body: str, fail: Fail) -> list[Token]:
    """Split the text of an action into tokens. `fail` reports what cannot be a token."""
    tokens: list[Token] = []
    for match in _ELEMENT.finditer(body):
        found = {name: text for name, text in match.groupdict().items() if text is not None}
        if "open" in found or "close" in found or "pipe" in found:
            tokens.append(Token(next(iter(found))))
        elif "string" in found:
            tokens.append(Token("string", _unquote(found["string"], fail)))
        elif "raw" in found:
            tokens.append(Token("string", found["raw"][1:-1]))
        elif "number" in found:
            text = found["number"]
            tokens.append(Token("number", float(text) if "." in text else int(text)))
        elif "field" in found:
            text = found["field"]
            tokens.append(
                Token("field", root=text.startswith("$"), path=tuple(p for p in text.split(".") if p not in ("", "$")))
            )
        elif "word" in found:
            tokens.append(Token("word", found["word"]))
        elif "variable" in found:
            fail("variables are not supported by the SDK's templates")
        elif "other" in found:
            fail(f'unexpected character "{found["other"]}" in action')
    return tokens


def _unquote(literal: str, fail: Fail) -> str:
    try:
        value = json.loads(literal)
    except ValueError:
        return fail(f"invalid string {literal}")
    return str(value)
