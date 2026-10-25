"""Recording the toolchain on discovered projects."""

from __future__ import annotations

from collections.abc import Callable

from ..protocol import DiscoverResult


def set_toolchain(result: DiscoverResult, probe: Callable[[], str]) -> DiscoverResult:
    """Record the toolchain on every discovered project.

    The probe (`node --version` and the like) runs only when there is a project, and once. Returns the
    result it was given.
    """
    if not result["projects"]:
        return result
    toolchain = probe()
    for project in result["projects"]:
        project["toolchain"] = toolchain
    return result
