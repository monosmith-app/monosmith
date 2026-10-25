"""Telling whether a file is one of a plugin's manifests."""

from __future__ import annotations

import re
from collections.abc import Iterable

# One element of a pattern: an escaped character, a class, a wildcard, or any other character.
_PART = re.compile(r"\\(.)|\[([^\]]+)\]|([*?])|(.)", re.DOTALL)


def is_manifest(manifests: Iterable[str], base: str) -> bool:
    """Whether a file's base name is one of a plugin's manifests: equal to a name, or matching a pattern (`*.csproj`).

    A pattern is Go's `path.Match`: `*` matches any run of characters, `?` one, `[a-z]` and `[^a-z]`
    a class, `\\` escapes. A malformed pattern matches nothing.
    """
    for manifest in manifests:
        if manifest == base:
            return True
        expression = _to_regex(manifest)
        if expression is not None and expression.fullmatch(base) is not None:
            return True
    return False


def _to_regex(pattern: str) -> re.Pattern[str] | None:
    source = ""
    for match in _PART.finditer(pattern):
        escaped, characters, wildcard, literal = match.groups()
        if escaped is not None:
            source += re.escape(escaped)
        elif characters is not None:
            source += "[" + characters.replace("\\", "\\\\") + "]"
        elif wildcard == "*":
            source += "[^/]*"
        elif wildcard == "?":
            source += "[^/]"
        elif literal in ("[", "\\"):
            return None  # A bracket or backslash that did not make an escape or a class is malformed.
        else:
            source += re.escape(literal)
    try:
        return re.compile(source, re.DOTALL)
    except re.error:
        return None
