"""Finding a plugin's manifests below a workspace."""

from __future__ import annotations

import os
from collections.abc import Sequence
from pathlib import Path

from ..projects import is_manifest

_SKIPPED_FOLDERS = frozenset({"node_modules", "__pycache__"})


def find_manifests(workspace: str | Path, names: Sequence[str]) -> list[str]:
    """The manifests of a plugin below a workspace, workspace-relative with forward slashes.

    Files named like `names` (see `is_manifest`), skipping dot-folders and dependency folders, and
    files at the root itself, as the Go conformance suite does.
    """
    found: list[str] = []

    def walk(relative: str) -> None:
        folder = Path(workspace, *[part for part in relative.split("/") if part])
        with os.scandir(folder) as entries:
            listing = sorted(entries, key=lambda entry: entry.name)
        for entry in listing:
            path = entry.name if relative == "" else f"{relative}/{entry.name}"
            if entry.is_dir():
                if not entry.name.startswith(".") and entry.name not in _SKIPPED_FOLDERS:
                    walk(path)
            elif is_manifest(names, entry.name) and "/" in path:
                found.append(path)

    try:
        walk("")
    except OSError:
        pass  # An unreadable folder ends the walk; what was found stays.
    return found
