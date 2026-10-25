"""References to other plugins: what a plugin requires or conflicts with."""

from __future__ import annotations

from ..protocol import PluginRef


def requires_language(name: str) -> PluginRef:
    """A requirement on the project's language plugin: `requires_language("node")`."""
    return {"level": "language", "name": name}


def conflicts_with(name: str) -> PluginRef:
    """A plugin that must not be in the same project."""
    return {"name": name}


def is_empty_ref(ref: PluginRef) -> bool:
    """Whether the reference names nothing: every member of a reference is optional, and at least one must be set."""
    return not ref.get("level") and not ref.get("name") and not ref.get("slot")


def describe_ref(ref: PluginRef) -> str:
    """How a message names it: `language plugin node`, `plugin providing test`."""
    level = ref.get("level") or ""
    parts = ["plugin" if level == "" else f"{level} plugin"]
    if ref.get("name"):
        parts.append(ref["name"])
    if ref.get("slot"):
        parts.append(f"providing {ref['slot']}")
    return " ".join(parts)
