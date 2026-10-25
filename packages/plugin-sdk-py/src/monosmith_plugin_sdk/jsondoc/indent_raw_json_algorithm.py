"""Formatting the text of a JSON array the way Go indents it."""

from __future__ import annotations

import re

# A string literal (escapes included), a bracket, a comma or colon, or a run of any other characters
# (a number, `true`, `null`).
_TOKEN = re.compile(r'"(?:[^"\\]|\\.)*"|[{}\[\],:]|[^\s{}\[\],:"]+')


def indent_raw_json(raw: str, prefix: str) -> str:
    """Format JSON text as Go's `json.Indent(dst, src, prefix, "  ")` does.

    The insignificant whitespace is dropped and every element goes on its own line, indented two
    spaces per level and starting with `prefix`; an empty array or object stays `[]` or `{}`. Text
    that is not an array is returned as it is. The text of leaves (numbers, strings) is kept exactly.
    """
    if not raw.startswith("["):
        return raw
    out: list[str] = []
    depth = 0
    need_indent = False  # Right after an opening bracket: held back until it is known whether the container is empty.

    def newline() -> None:
        out.append(f"\n{prefix}{'  ' * depth}")

    for match in _TOKEN.finditer(raw):
        token = match.group()
        if need_indent and token not in ("}", "]"):
            need_indent = False
            depth += 1
            newline()
        if token in ("{", "["):
            need_indent = True
            out.append(token)
        elif token in ("}", "]"):
            if need_indent:
                need_indent = False
            else:
                depth -= 1
                newline()
            out.append(token)
        elif token == ",":
            out.append(token)
            newline()
        elif token == ":":
            out.append(": ")
        else:
            out.append(token)
    return "".join(out)
