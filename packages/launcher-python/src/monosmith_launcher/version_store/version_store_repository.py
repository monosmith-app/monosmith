"""The folder that holds every smith the launchers and ``smith update`` installed."""

from __future__ import annotations

import re
from collections.abc import Mapping
from pathlib import Path

from ..version_choice import version_key


def versions_dir(env: Mapping[str, str]) -> Path:
    """Return the versions folder: ``<home>/versions``. SMITH_HOME moves it; default ~/.smith.

    Args:
        env: The environment.

    Returns:
        The versions folder.
    """
    home = env.get("SMITH_HOME")
    return (Path(home) if home else Path.home() / ".smith") / "versions"


def program_name(system: str) -> str:
    """Return the file name of the smith program on a system.

    Args:
        system: The archive's system name (windows, linux, darwin).

    Returns:
        smith.exe on Windows, smith elsewhere.
    """
    return "smith.exe" if system == "windows" else "smith"


def program_path(env: Mapping[str, str], version: str, system: str) -> Path:
    """Return where one version's program lives, whether it exists or not.

    Args:
        env: The environment.
        version: An exact version.
        system: The archive's system name.

    Returns:
        The program's path.
    """
    return versions_dir(env) / version / program_name(system)


def newest_installed(env: Mapping[str, str], system: str) -> str | None:
    """Return the newest installed version that has its program.

    Args:
        env: The environment.
        system: The archive's system name.

    Returns:
        The version, or None when none is installed.
    """
    folder = versions_dir(env)
    if not folder.is_dir():
        return None
    installed = [
        v.name
        for v in folder.iterdir()
        if re.match(r"^\d+\.\d+\.\d+", v.name) and program_path(env, v.name, system).is_file()
    ]
    return max(installed, key=version_key, default=None)
