"""The version a run uses: SMITH_VERSION, the workspace's pin, or the newest."""

from __future__ import annotations

import functools
import re
from collections.abc import Mapping
from pathlib import Path

VERSION = re.compile(r"^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$")
"""A version smith publishes: 2.50.0, 3.0.0-beta.1."""


def pin_in(text: str) -> str | None:
    """Return the ``[workspace] smith`` pin in smith.toml text.

    Args:
        text: The text of smith.toml.

    Returns:
        The version, or None for none or ``latest``.
    """
    table = ""
    for raw in text.splitlines():
        line = re.sub(r"#.*$", "", raw).strip()
        header = re.fullmatch(r"\[([^\]]+)\]", line)
        if header:
            table = header.group(1).strip()
            continue
        pin = re.fullmatch(r"smith\s*=\s*[\"']([^\"']+)[\"']", line)
        if table == "workspace" and pin:
            return pin.group(1) if VERSION.match(pin.group(1)) else None
    return None


def workspace_pin(start: Path) -> str | None:
    """Return the pin of the workspace a folder is in: the nearest smith.toml above it.

    Args:
        start: The folder to start from (the current folder).

    Returns:
        The pinned version, or None when no smith.toml pins one.
    """
    for folder in (start, *start.parents):
        file = folder / "smith.toml"
        if file.is_file():
            return pin_in(file.read_text(encoding="utf-8"))
    return None


def choose_version(env: Mapping[str, str], cwd: Path) -> str | None:
    """Return the version a run uses: SMITH_VERSION, else the workspace's pin, else None (the newest).

    Args:
        env: The environment.
        cwd: The current folder.

    Returns:
        An exact version, or None for the newest.

    Raises:
        ValueError: When SMITH_VERSION is set to something that is not a version.
    """
    wanted = env.get("SMITH_VERSION", "").strip()
    if wanted and wanted != "latest":
        if not VERSION.match(wanted):
            raise ValueError(f'SMITH_VERSION is "{wanted}": use a version such as 2.50.0, or latest')
        return wanted
    return None if wanted == "latest" else workspace_pin(cwd)


def compare_versions(a: str, b: str) -> int:
    """Compare two versions by their numbers; a prerelease sorts before its release.

    Args:
        a: A version.
        b: Another version.

    Returns:
        Negative when a is older, positive when newer, 0 when equal.
    """
    main_a, _, pre_a = a.partition("-")
    main_b, _, pre_b = b.partition("-")
    for x, y in zip(map(int, main_a.split(".")), map(int, main_b.split(".")), strict=True):
        if x != y:
            return x - y
    if pre_a == pre_b:
        return 0
    if not pre_a:
        return 1
    if not pre_b:
        return -1
    return -1 if pre_a < pre_b else 1


version_key = functools.cmp_to_key(compare_versions)
"""A sort key that orders versions by compare_versions."""
