"""Which member of a plugin answers each method of the protocol."""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class PluginMember:
    """The member that answers a method, and the members a plugin must have for it to count as implemented."""

    member: str
    requires: tuple[str, ...]


PLUGIN_MEMBERS: dict[str, PluginMember] = {
    "discover": PluginMember("discover", ("discover",)),
    "generate": PluginMember("generate", ("generate",)),
    "version/write": PluginMember("write_version", ("write_version",)),
    "publish": PluginMember("publish", ("publish",)),
    "ci": PluginMember("ci", ("ci",)),
    "doctor": PluginMember("doctor", ("doctor",)),
    "contribute": PluginMember("contribute", ("contribute",)),
    "remove": PluginMember("remove", ("remove",)),
    "project/kind": PluginMember("change_kind", ("change_kind",)),
    "published": PluginMember("published", ("published",)),
    "dependencies/install": PluginMember("install", ("install", "uninstall")),
    "dependencies/uninstall": PluginMember("uninstall", ("install", "uninstall")),
    "dependencies/declared": PluginMember("declared", ("declared", "rewrite")),
    "dependencies/rewrite": PluginMember("rewrite", ("declared", "rewrite")),
    "dependencies/latest": PluginMember("latest", ("latest",)),
}
"""Every method of the protocol other than `initialize`, `shutdown` and `cancel`, which the serve loop
answers itself. A plugin with `install` and no `uninstall` does not implement the `dependencies`
capability, as in the Go SDK, where the interface needs both methods. A test checks the table against
the methods of the schema, so a method added to the protocol cannot be left out."""
