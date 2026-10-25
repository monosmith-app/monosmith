"""Reading, checking and rewriting a widget.json."""

from __future__ import annotations

import json
import re
from dataclasses import dataclass, field
from pathlib import Path

from monosmith_plugin_sdk import JsonDocument

MANIFEST_FILE = "widget.json"
"""The file that makes a directory a widget project."""

SEMVER = re.compile(r"^\d+\.\d+\.\d+$")


@dataclass
class Manifest:
    """The content of a widget.json. A member it does not have is empty."""

    name: str = ""
    version: str = ""
    kind: str = ""
    """app, lib or package; default lib"""
    description: str = ""
    license: str = ""
    uses: list[str] = field(default_factory=list)
    """names of other widgets this one imports"""
    stack: str = ""


def read_manifest(path: Path) -> Manifest:
    """Read a widget.json."""
    try:
        parsed = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, ValueError) as error:
        raise ValueError(f"{path.as_posix()}: {error}") from error
    return Manifest(
        name=parsed.get("name") or "",
        version=parsed.get("version") or "",
        kind=parsed.get("kind") or "",
        description=parsed.get("description") or "",
        license=parsed.get("license") or "",
        uses=parsed.get("uses") or [],
        stack=parsed.get("stack") or "",
    )


def validate(manifest: Manifest) -> None:
    """What the test task checks: a name and an x.y.z version."""
    if manifest.name == "":
        raise ValueError(f"{MANIFEST_FILE} has no name")
    if not SEMVER.match(manifest.version):
        raise ValueError(f'{MANIFEST_FILE}: version "{manifest.version}" is not x.y.z')


def with_field(path: Path, key: str, value: str) -> str:
    """The widget.json with one string member set, every other key keeping its order and layout."""
    document = JsonDocument.parse(path.read_text(encoding="utf-8"))
    document.set_string(key, value)
    return document.render()


def set_field(path: Path, key: str, value: str) -> None:
    """Rewrite one string member of a widget.json."""
    path.write_bytes(with_field(path, key, value).encode("utf-8"))
