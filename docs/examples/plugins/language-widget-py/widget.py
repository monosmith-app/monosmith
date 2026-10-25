"""The widget language plugin."""

from __future__ import annotations

import posixpath
import sys
from pathlib import Path
from typing import Any

from manifest import MANIFEST_FILE, read_manifest, set_field
from monosmith_plugin_sdk import JsonDocument, build_task, scaffold_project, task

SELF = [sys.executable, str(Path(__file__).resolve().with_name("plugin.py"))]
"""The command the tasks run: this program. A plugin that wraps a real tool would name the tool here instead."""

GENERATED_FILES = {
    "[[.Dir]]/widget.json": """{
  "name": "[[.Name]]",
  "version": "0.0.0",
  "kind": "[[.Kind]]",
  "description": "A widget of the [[.Workspace]] workspace.",
  "license": "MIT"[[if .Stack]],
  "stack": "[[.Stack]]"[[end]]
}
""",
    "[[.Dir]]/README.md": "# [[.Name]]\n\nA widget. Edit widget.json.\n",
}
"""A new project's starter files. The templates use [[ ]], so a body can contain {{ }} untouched."""


class WidgetPlugin:
    """The plugin. Every method it has beyond `describe` is a capability the SDK advertises:
    `discover`, `generate`, `write_version` and `doctor`."""

    def __init__(self) -> None:
        self.workspace_root = ""

    def _manifest_path(self, relative: str) -> Path:
        return Path(self.workspace_root, *relative.split("/"))

    def describe(self, params: Any) -> Any:
        self.workspace_root = params["workspaceRoot"]
        return {
            "plugin": {"name": "widget", "version": "0.1.0"},
            "kind": "language",
            # A language plugin that also writes new projects plays the scaffold level too;
            # it does not need to be chosen as a toolchain.
            "roles": ["language", "scaffold"],
            "languages": ["widget"],
            "manifests": [MANIFEST_FILE],
        }

    def discover(self, params: Any) -> Any:
        """One project per widget.json that is ours."""
        projects = []
        for relative in params["manifests"]:  # workspace-relative, forward slashes
            manifest = read_manifest(self._manifest_path(relative))
            if manifest.name == "":
                continue  # not a widget manifest: leave it out
            root = posixpath.dirname(relative)
            projects.append(
                {
                    "name": manifest.name,
                    "root": root,
                    "language": "widget",
                    "kind": manifest.kind or "lib",
                    "version": manifest.version,
                    "description": manifest.description,
                    "stack": manifest.stack,
                    "exports": [manifest.name],
                    "imports": manifest.uses,  # the host keeps the ones another project exports
                    "tasks": {
                        "build": build_task(
                            root,
                            lambda out: [*SELF, "build", out],
                            inputs=[MANIFEST_FILE],
                            description="write widget.txt into the build directory",
                        ),
                        "test": task(
                            [*SELF, "check"],
                            inputs=[MANIFEST_FILE],
                            cache=True,
                            description="check the manifest: a name and an x.y.z version",
                        ),
                    },
                }
            )
        return {"projects": projects}

    def generate(self, params: Any) -> Any:
        """A new project's starter files."""
        return scaffold_project(self.workspace_root, params, GENERATED_FILES)

    def write_version(self, params: Any) -> Any:
        """Record the release version."""
        file = posixpath.join(params["project"]["root"], MANIFEST_FILE)
        if params.get("dryRun") is not True:
            set_field(self._manifest_path(file), "version", params["version"])
        return {"files": [file]}

    def doctor(self, params: Any) -> Any:
        """Every failed finding names its remedy, and a fix when the plugin can write one."""
        findings = []
        for project in params["projects"]:
            file = posixpath.join(project["root"], MANIFEST_FILE)
            manifest = read_manifest(self._manifest_path(file))
            check = f"{project['name']} names a license"
            if manifest.license:
                findings.append({"check": check, "ok": True})
                continue
            document = JsonDocument.parse(self._manifest_path(file).read_text(encoding="utf-8"))
            document.set_string("license", "MIT")
            findings.append(
                {
                    "check": check,
                    "ok": False,
                    "detail": f"{file} has no license field",
                    "remedy": f'add "license": "MIT" (or your license) to {file}',
                    # Not safe: it picks a license for the team, so smith asks first.
                    "fix": {
                        "description": "add an MIT license to widget.json",
                        "safe": False,
                        "applies_to": [project["name"]],
                        "files": [{"path": file, "content": document.render()}],
                    },
                }
            )
        return {"findings": findings}
