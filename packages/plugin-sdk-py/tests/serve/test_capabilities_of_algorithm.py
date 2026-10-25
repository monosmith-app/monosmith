"""The capabilities a plugin advertises, and the table that maps methods to plugin members."""

from __future__ import annotations

from typing import Any

from monosmith_plugin_sdk import PROTOCOL_SCHEMA, capabilities_of
from monosmith_plugin_sdk.serve import PLUGIN_MEMBERS, implements_method


def plugin_with(*names: str, cancellable: bool = False) -> object:
    """An object with the named methods."""
    attributes: dict[str, Any] = {name: (lambda self, params: {}) for name in names}
    attributes["describe"] = lambda self, params: {}
    attributes["cancellable"] = cancellable
    return type("Plugin", (), attributes)()


def test_every_method_of_the_schema_has_its_place_in_the_plugin() -> None:
    methods = set(PROTOCOL_SCHEMA["x-methods"])
    assert set(PLUGIN_MEMBERS) == methods - {"initialize", "shutdown", "cancel"}


def test_capability_names_are_not_method_names() -> None:
    cases = {
        "version/write": ("write_version", "version"),
        "dependencies/install": ("install", "dependencies"),
        "dependencies/declared": ("declared", "declarations"),
        "dependencies/latest": ("latest", "latest"),
    }
    for method, (member, capability) in cases.items():
        assert PLUGIN_MEMBERS[method].member == member
        assert PROTOCOL_SCHEMA["x-methods"][method]["capability"] == capability


def test_derives_the_capabilities_from_the_methods_the_plugin_has() -> None:
    assert capabilities_of(plugin_with()) == {}
    assert capabilities_of(plugin_with("discover", "doctor", "write_version")) == {
        "discover": True,
        "doctor": True,
        "version": True,
    }


def test_a_pair_of_methods_needs_both_to_count() -> None:
    assert capabilities_of(plugin_with("install")) == {}
    assert capabilities_of(plugin_with("install", "uninstall")) == {"dependencies": True}
    assert capabilities_of(plugin_with("declared")) == {}
    assert capabilities_of(plugin_with("declared", "rewrite")) == {"declarations": True}
    assert implements_method(plugin_with("install"), "dependencies/install") is False


def test_cancel_comes_from_the_plugin_saying_it_is_cancellable() -> None:
    assert capabilities_of(plugin_with(cancellable=True)) == {"cancel": True}
    assert capabilities_of(plugin_with()) == {}
