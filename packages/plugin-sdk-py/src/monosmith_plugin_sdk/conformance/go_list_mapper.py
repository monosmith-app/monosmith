"""Writing values the way the details of the Go checks do."""

from __future__ import annotations

import json
from collections.abc import Iterable
from typing import Any


def go_list(items: Iterable[Any]) -> str:
    """A list as Go's `%v` writes it, `[a b]`, which the details of the Go checks use."""
    return "[" + " ".join(str(item) for item in items) + "]"


def go_quote(text: str) -> str:
    """A string as Go's `%q` writes it."""
    return json.dumps(text, ensure_ascii=False)
