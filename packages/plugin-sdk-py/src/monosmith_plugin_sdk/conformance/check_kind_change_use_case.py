"""The check of a plugin that changes kinds."""

from __future__ import annotations

import posixpath
from collections.abc import Sequence
from pathlib import Path

from ..host import PluginClient
from ..projects import project_kind_directory
from ..protocol import Project, ProjectKind
from .conformance_report_model import ConformanceReport
from .find_manifests_algorithm import find_manifests
from .go_list_mapper import go_list


def _kind_change_target(kind: ProjectKind) -> ProjectKind:
    """The kind a generated project is asked to become.

    A lib is published as a package, an app becomes a lib, and a package goes back to a lib.
    """
    if kind == "app":
        return "lib"
    return "package" if kind == "lib" else "lib"


def _article(kind: ProjectKind) -> str:
    return "an app" if kind == "app" else f"a {kind}"


def _snapshot(workspace: str | Path, names: Sequence[str]) -> dict[str, str]:
    """The content of every manifest of the workspace, by path."""
    paths = find_manifests(workspace, names)
    return {path: Path(workspace, *path.split("/")).read_text(encoding="utf-8") for path in paths}


def check_kind_change(
    client: PluginClient, workspace: str | Path, project: Project, projects: list[Project], report: ConformanceReport
) -> None:
    """Ask what making a generated project another kind edits.

    The plugin must answer with a refusal or with edits to files that exist, and write nothing itself:
    the host applies the edits.
    """
    to = _kind_change_target(project["kind"])
    new_root = f"{project_kind_directory(to)}/{posixpath.basename(project['root'])}"
    check = f"answers what making the {project['kind']} {_article(to)} edits (dry run)"
    names = client.info.get("manifests") or []
    before = _snapshot(workspace, names)
    try:
        answer = client.call(
            "project/kind",
            {"project": project, "to": to, "newRoot": new_root, "projects": projects},
            timeout=300.0,
        )
    except Exception as error:
        report.add(check, False, str(error))
        return
    files = answer.get("files") or []
    missing = [f["path"] for f in files if not Path(workspace, *f["path"].split("/")).exists()]
    if not answer.get("refused") and not files:
        report.add(check, False, "it answered with neither a refusal nor an edit: a changed kind is recorded in a file")
    elif missing:
        report.add(check, False, f"edits files that do not exist (paths are those before the move): {go_list(missing)}")
    else:
        report.add(
            check,
            before == _snapshot(workspace, names),
            "it wrote to the workspace: it must only answer with the new content, and let the host write it",
        )
