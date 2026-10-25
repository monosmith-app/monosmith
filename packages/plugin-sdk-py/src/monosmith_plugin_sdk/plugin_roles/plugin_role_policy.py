"""The roles a plugin can play in a project's make-up."""

from __future__ import annotations

from collections.abc import Mapping
from typing import Any

from ..protocol import PROTOCOL_SCHEMA, PluginKind, Role

_ROWS: list[dict[str, Any]] = PROTOCOL_SCHEMA["x-roles"]

ROLES: tuple[str, ...] = tuple(row["name"] for row in _ROWS)
"""Every known role, in the order a project's make-up lists them."""


def is_known_role(role: str) -> bool:
    """Whether the role is one of `ROLES`."""
    return role in ROLES


def _cardinality(role: str) -> str | None:
    return next((row["cardinality"] for row in _ROWS if row["name"] == role), None)


def is_exclusive_role(role: str) -> bool:
    """Whether a project has at most one plugin at this level (language, scaffold).

    A toolchain is exclusive per slot instead.
    """
    return _cardinality(role) == "exclusive"


def is_additive_role(role: str) -> bool:
    """Whether any number of plugins can share the level, each only adding (structure, deploy)."""
    return _cardinality(role) == "additive"


def default_roles(kind: PluginKind | str) -> list[Role]:
    """The roles of a plugin that declares none.

    A language plugin brings its own default toolchain and scaffold; any other kind of plugin is a
    structure rule.
    """
    wanted = "language" if kind == "language" else "convention"
    return [row["name"] for row in _ROWS if wanted in row["defaultFor"]]


def effective_roles(info: Mapping[str, Any]) -> list[Role]:
    """The plugin's roles (`roles` of its `initialize` result), or the default of its kind."""
    roles = info.get("roles")
    return list(roles) if roles else default_roles(info.get("kind", ""))
