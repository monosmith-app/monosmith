"""Writes the generated protocol modules, or checks that they are current."""

from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path

from .python_types_mapper import map_schema_to_types
from .schema_module_mapper import map_schema_to_module

SCHEMA_PATH = "../plugin-sdk/schema/protocol.schema.json"
"""The schema, relative to the package root: the Go SDK's, next to the types it comes from."""

PROTOCOL_DIRECTORY = "src/monosmith_plugin_sdk/protocol"


@dataclass(frozen=True)
class GeneratedFile:
    """A file the generator owns."""

    path: str
    """Relative to the package root, forward slashes."""
    content: str


def generated_files(schema_text: str) -> list[GeneratedFile]:
    """What the generator writes, from the schema's text."""
    schema = json.loads(schema_text)
    return [
        GeneratedFile(f"{PROTOCOL_DIRECTORY}/protocol_types_contract.py", map_schema_to_types(schema)),
        GeneratedFile(f"{PROTOCOL_DIRECTORY}/protocol_schema_contract.py", map_schema_to_module(schema)),
    ]


def generate_protocol(package_root: Path, check: bool = False) -> list[str]:
    """Write the generated files, or, with `check`, return the ones that are not what the schema
    produces. An empty result from a check means they are current."""
    schema_text = (package_root / SCHEMA_PATH).read_text(encoding="utf-8")
    stale: list[str] = []
    for file in generated_files(schema_text):
        target = package_root / file.path
        current = target.read_text(encoding="utf-8").replace("\r\n", "\n") if target.exists() else None
        if current == file.content:
            continue
        stale.append(file.path)
        if not check:
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text(file.content, encoding="utf-8", newline="\n")
    return stale
