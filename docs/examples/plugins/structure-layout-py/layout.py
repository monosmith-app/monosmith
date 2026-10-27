"""The structure plugin: no junk-drawer folders in widget projects, and a README for every project."""

from __future__ import annotations

import sys
from pathlib import Path
from typing import Any

from monosmith_plugin_sdk import contribute_to, task

JUNK_DRAWERS = frozenset({"util", "utils", "helper", "helpers", "common", "shared", "misc"})
SKIPPED = frozenset({"dist", "node_modules", ".git", "__pycache__"})

SELF = str(Path(__file__).resolve().with_name("plugin.py"))
"""The entry point, which is also the program the layout-check task runs."""


def find_junk_drawers(directory: str | Path, prefix: str = "") -> list[str]:
    """The folders below `directory` named for nothing, as slash-separated paths relative to it."""
    found: list[str] = []
    for entry in sorted(Path(directory, *prefix.split("/")).iterdir() if prefix else Path(directory).iterdir()):
        if not entry.is_dir() or entry.name in SKIPPED:
            continue
        relative = entry.name if prefix == "" else f"{prefix}/{entry.name}"
        if entry.name.lower() in JUNK_DRAWERS:
            found.append(relative)
        found.extend(find_junk_drawers(directory, relative))
    return found


class LayoutPlugin:
    """The plugin. It implements `contribute` (tasks) and `doctor` (findings): the SDK advertises exactly those."""

    def __init__(self) -> None:
        self.workspace_root = ""

    def describe(self, params: Any) -> Any:
        self.workspace_root = params["workspaceRoot"]
        return {
            "plugin": {"name": "widget-layout", "version": "0.1.0"},
            "kind": "convention",
            "roles": ["structure"],
            # The languages it applies to. A convention that names none applies to every language.
            "languages": ["widget"],
        }

    def contribute(self, params: Any) -> Any:
        """A layout-check task per project, and the project's lint (when it has one) waits for it."""
        return contribute_to(
            params["projects"],
            lambda project: {
                "tasks": {
                    # No inputs listed: every file of the project is one, so a new or renamed folder is a cache miss.
                    "layout-check": task(
                        [sys.executable, SELF, "layout-check"],
                        cache=True,
                        description="fail on folders named util, helpers, common, shared or misc",
                    ),
                },
                # dependsOn only touches a task the project has once every earlier plugin has contributed;
                # for a project without lint it is ignored. Do not decide from project["tasks"] here: it holds
                # only what the language plugin discovered, not what a toolchain plugin adds.
                "dependsOn": {"lint": ["layout-check"]},
            },
        )

    def doctor(self, params: Any) -> Any:
        """Every project needs a README.md."""
        findings = []
        for project in params["projects"]:
            check = f"{project['name']} has a README.md"
            if Path(self.workspace_root, *project["root"].split("/"), "README.md").exists():
                findings.append({"check": check, "ok": True})
                continue
            findings.append(
                {
                    "check": check,
                    "ok": False,
                    "detail": f"{project['root']}/README.md does not exist",
                    "remedy": f"write {project['root']}/README.md: what the project is for",
                    "warning": True,  # reported, but does not fail smith doctor
                }
            )
        return {"findings": findings}
