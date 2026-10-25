"""The plugins of a stack."""

from __future__ import annotations

from collections.abc import Mapping
from typing import Any


def stack_plugins(stack: Mapping[str, Any]) -> list[str]:
    """The stack's plugins other than its language and scaffold, in the order they apply.

    That is toolchain, structure, deploy.
    """
    return [*(stack.get("toolchain") or []), *(stack.get("structure") or []), *(stack.get("deploy") or [])]
