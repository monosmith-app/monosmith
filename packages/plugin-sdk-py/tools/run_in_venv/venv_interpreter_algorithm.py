"""Finding the workspace's virtual environment interpreter."""

from __future__ import annotations

from pathlib import Path


def find_venv_interpreter(start: Path) -> Path | None:
    """The `.venv` interpreter of the nearest workspace above `start`, or `None`.

    It is `.venv/Scripts/python.exe` on Windows and `.venv/bin/python` elsewhere, which is where
    smith's Python plugin puts it.
    """
    for folder in (start, *start.parents):
        for relative in (Path("Scripts") / "python.exe", Path("bin") / "python"):
            candidate = folder / ".venv" / relative
            if candidate.is_file():
                return candidate
    return None
