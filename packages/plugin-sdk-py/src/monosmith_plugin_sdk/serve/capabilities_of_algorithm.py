"""The capabilities a plugin advertises, derived from the methods it has."""

from __future__ import annotations

from typing import Any

from ..protocol import METHOD_CAPABILITIES, Capabilities
from .plugin_members_config import PLUGIN_MEMBERS


def _has(plugin: object, member: str) -> bool:
    return callable(getattr(plugin, member, None))


def implements_method(plugin: object, method: str) -> bool:
    """Whether the plugin implements a method of the protocol: it has every member the method needs."""
    entry = PLUGIN_MEMBERS.get(method)
    return entry is not None and all(_has(plugin, member) for member in entry.requires)


def is_cancellable(plugin: object) -> bool:
    """Whether the plugin says its methods stop when the host cancels a request (`cancellable = True`)."""
    return getattr(plugin, "cancellable", False) is True


def capabilities_of(plugin: object) -> Capabilities:
    """The capabilities a plugin advertises, derived from the methods it implements.

    A plugin cannot advertise a method it lacks, or implement one it does not advertise. The
    capability of each method is read from the schema's methods table; `cancel` comes from the
    plugin saying it is `cancellable`.
    """
    capabilities: dict[str, Any] = {}
    for method, capability in METHOD_CAPABILITIES.items():
        if capability is not None and implements_method(plugin, method):
            capabilities[capability] = True
    if is_cancellable(plugin):
        capabilities["cancel"] = True
    return capabilities  # type: ignore[return-value]


def half_implemented(plugin: object) -> list[str]:
    """Members of a pair (`install` without `uninstall`) that are not advertised for lack of the other."""
    return [
        entry.member
        for method, entry in PLUGIN_MEMBERS.items()
        if _has(plugin, entry.member) and not implements_method(plugin, method)
    ]
