"""Whether a release publishes a project."""

from __future__ import annotations

from collections.abc import Mapping
from typing import Any


def publishes(project: Mapping[str, Any]) -> bool:
    """Whether a release publishes the project to a registry: a package, unless `smith.toml` says `registry = "none"`.

    Plugins list the credentials a release needs only for such projects.
    """
    registries = project.get("registries") or {}
    return project.get("kind") == "package" and registries.get("noPublish") is not True
