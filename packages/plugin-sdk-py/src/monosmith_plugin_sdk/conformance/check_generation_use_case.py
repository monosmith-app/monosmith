"""The checks of a plugin that generates projects."""

from __future__ import annotations

import functools
from collections.abc import Callable
from pathlib import Path
from typing import Any

from ..host import PluginClient
from ..plugin_roles import effective_roles
from ..projects import project_kind_directory
from ..protocol import GenerateParams, Project, ProjectKind
from .check_kind_change_use_case import check_kind_change
from .conformance_report_model import ConformanceReport
from .find_manifests_algorithm import find_manifests
from .go_list_mapper import go_list, go_quote

CONFORMANCE_STACK = "conformance"
"""The stack name generated projects are asked to record."""

LONG = 20 * 60.0


def _article(kind: ProjectKind) -> str:
    return "an app" if kind == "app" else f"a {kind}"


def check_generation(client: PluginClient, workspace: str | Path, report: ConformanceReport) -> None:
    """Ask the plugin for one project of each kind and check the files it reports exist.

    For a language plugin, also that its own `discover` returns what it made, with the kind and the
    stack it was asked for, tasks with commands, and that the dry runs of `version/write`, `remove`,
    `publish` and `published` answer, and, for a plugin that changes kinds, what `project/kind` edits.
    """
    language = "language" in effective_roles(client.info)
    made: list[GenerateParams] = []
    declared = client.info.get("kinds") or []
    kinds: tuple[ProjectKind, ...] = tuple(k for k in ("app", "lib", "package") if not declared or k in declared)
    for kind in kinds:
        name = f"c{kind}"
        params: GenerateParams = {
            "kind": kind,
            "name": name,
            "directory": f"{project_kind_directory(kind)}/{name}",
            "stack": CONFORMANCE_STACK,
            "workspace": {"name": "conformance", "scope": "@conformance", "modulePrefix": "example.com/conformance"},
        }
        try:
            files = client.call("generate", params, timeout=LONG)["files"]
        except Exception as error:
            report.add(f"generates {_article(kind)}", False, str(error))
            continue
        missing = [file for file in files if not Path(workspace, *file.split("/")).exists()]
        report.add(
            f"generates {_article(kind)}",
            len(files) > 0 and not missing,
            f"files reported but not written: {go_list(missing)} (of {go_list(files)})",
        )
        made.append(params)
    if not language or not made:
        return

    try:
        manifests = find_manifests(workspace, client.info.get("manifests") or [])
        projects: list[Project] = client.call("discover", {"manifests": manifests}, timeout=LONG)["projects"]
    except Exception as error:
        report.add("discovers what it generated", False, str(error))
        return
    found = {project["root"]: project for project in projects}
    for params in made:
        kind = params["kind"]
        project = found.get(params["directory"])
        report.add(
            f"discovers the generated {kind}",
            project is not None,
            f"no project at {params['directory']} among {len(projects)} discovered",
        )
        if project is None:
            continue
        report.add(
            f"the generated {kind} keeps its kind",
            project["kind"] == kind,
            f"discovered as {go_quote(project['kind'])}",
        )
        report.add(
            f"the generated {kind} records its stack",
            project.get("stack") == CONFORMANCE_STACK,
            f"discovered with stack {go_quote(project.get('stack', ''))}; "
            "generate must record params.stack in the manifest",
        )
        empty = sorted(name for name, task in (project.get("tasks") or {}).items() if not task["command"])
        report.add(
            f"the generated {kind}'s tasks have commands", not empty, f"tasks without a command: {go_list(empty)}"
        )
        call = functools.partial(client.call, timeout=60.0)
        if client.supports("version/write"):
            _attempt(
                report,
                f"writes the {kind}'s version (dry run)",
                functools.partial(call, "version/write", {"project": project, "version": "1.2.3", "dryRun": True}),
            )
        if client.supports("remove"):
            _attempt(
                report,
                f"would remove the {kind} (dry run)",
                functools.partial(call, "remove", {"project": project, "dryRun": True}),
            )
        if client.supports("project/kind"):
            check_kind_change(client, workspace, project, projects, report)
        if kind == "package" and client.supports("publish"):
            publish = functools.partial(client.call, timeout=300.0)
            _attempt(
                report,
                "publishes the package (dry run)",
                functools.partial(publish, "publish", {"project": project, "version": "1.2.3", "dryRun": True}),
            )
        if kind == "package" and client.supports("published"):
            _attempt(
                report,
                "says what the registry has of the package",
                functools.partial(call, "published", {"project": project}),
                " (an unreachable registry is a result, not an error)",
            )


def _attempt(report: ConformanceReport, check: str, call: Callable[[], Any], hint: str = "") -> None:
    try:
        call()
        report.add(check, True, "")
    except Exception as error:
        report.add(check, False, f"{error}{hint}")
