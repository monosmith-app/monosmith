"""The conformance checks, run against plugins that keep and break their promises."""

from __future__ import annotations

from pathlib import Path
from typing import Any

import pytest

from monosmith_plugin_sdk import HostOptions, check_plugin, connect_in_process
from tests.conftest import Lab, Widget


def names(report: Any) -> list[str]:
    return [result.check for result in report.results]


def test_a_plugin_that_keeps_its_promises_passes() -> None:
    class Layout(Widget):
        def describe(self, params: Any) -> Any:
            return {"plugin": {"name": "layout", "version": "0.1.0"}, "kind": "convention", "roles": ["structure"]}

        def contribute(self, params: Any) -> Any:
            return {}

        def doctor(self, params: Any) -> Any:
            return {"findings": []}

    with connect_in_process(Layout(), HostOptions(workspace_root="/w")) as client:
        report = check_plugin(client)
    assert report.ok, report.failures()
    assert "names itself" in names(report)
    assert "a structure plugin does something" in names(report)
    assert "answers 3 requests sent together, one at a time" in names(report)
    assert "doctor answers with no projects" in names(report)


def test_a_cancellable_plugin_with_a_concurrency_is_checked_for_both() -> None:
    class Scan(Lab):
        def describe(self, params: Any) -> Any:
            return {**super().describe(params), "kind": "convention", "roles": ["structure"], "manifests": []}

        def doctor(self, params: Any) -> Any:
            return {"findings": []}

    lab = Scan(cancellable=True, concurrency=4)
    with connect_in_process(lab, HostOptions(workspace_root="/w")) as client:
        report = check_plugin(client)
    assert report.ok, report.failures()
    assert "accepts 4 requests at the same time" in names(report)
    assert "answers a cancelled request with its result or the cancelled error" in names(report)
    assert "ignores a cancel for a request it does not know" in names(report)


def test_reports_what_is_wrong_with_a_plugin_that_does_not_name_itself() -> None:
    class Nameless:
        def describe(self, params: Any) -> Any:
            return {"plugin": {"name": "", "version": ""}, "kind": "weird", "roles": ["nonsense"]}

    with connect_in_process(Nameless(), HostOptions(workspace_root="/w")) as client:
        report = check_plugin(client)
    assert not report.ok
    failed = [result.check for result in report.results if not result.ok]
    assert failed == ["names itself", "has a known kind", "plays a known role: nonsense"]
    assert report.to_dict()["results"][0]["detail"].startswith('initialize returned name ""')


def test_a_language_plugin_needs_to_discover_and_name_manifests() -> None:
    with connect_in_process(Widget(), HostOptions(workspace_root="/w")) as client:
        report = check_plugin(client)
    failed = [result.check for result in report.results if not result.ok]
    assert "a language plugin discovers projects" in failed


class Fig:
    """A language whose manifest, fig.txt, holds "kind stack"; `how` says how its change_kind misbehaves, if it does."""

    def __init__(self, workspace: Path, how: str | None = None) -> None:
        self.workspace = workspace
        self.how = how

    def describe(self, params: Any) -> Any:
        return {"plugin": {"name": "fig", "version": "1.0.0"}, "kind": "language", "manifests": ["fig.txt"]}

    def generate(self, params: Any) -> Any:
        folder = self.workspace.joinpath(*params["directory"].split("/"))
        folder.mkdir(parents=True, exist_ok=True)
        (folder / "fig.txt").write_text(f"{params['kind']} {params.get('stack', '')}", encoding="utf-8")
        return {"files": [f"{params['directory']}/fig.txt"]}

    def discover(self, params: Any) -> Any:
        projects = []
        for manifest in params["manifests"]:
            kind, stack = self.workspace.joinpath(*manifest.split("/")).read_text(encoding="utf-8").split(" ", 1)
            root = manifest.rsplit("/", 1)[0]
            projects.append(
                {
                    "name": root.rsplit("/", 1)[1],
                    "root": root,
                    "language": "fig",
                    "kind": kind,
                    "stack": stack,
                    "tasks": {"build": {"command": ["fig", "build"]}},
                }
            )
        return {"projects": projects}


class FigChanger(Fig):
    def change_kind(self, params: Any) -> Any:
        manifest = f"{params['project']['root']}/fig.txt"
        if self.how == "nothing":
            return {}
        if self.how == "missing":
            return {"files": [{"path": f"{params['project']['root']}/nowhere.txt", "content": "x"}]}
        content = f"{params['to']} conformance"
        if self.how == "write":
            self.workspace.joinpath(*manifest.split("/")).write_text(content, encoding="utf-8")
        return {"files": [{"path": manifest, "content": content}]}


def checked(plugin: Any, workspace: Path) -> Any:
    with connect_in_process(plugin, HostOptions(workspace_root=str(workspace))) as client:
        return check_plugin(client, workspace=workspace)


def test_a_plugin_that_changes_kinds_is_asked_what_each_generated_kind_becomes(tmp_path: Path) -> None:
    report = checked(FigChanger(tmp_path), tmp_path)
    assert report.ok, report.failures()
    assert "answers what making the app a lib edits (dry run)" in names(report)
    assert "answers what making the lib a package edits (dry run)" in names(report)
    assert "answers what making the package a lib edits (dry run)" in names(report)


class FigAppOnly(Fig):
    def describe(self, params: Any) -> Any:
        return {**super().describe(params), "kinds": ["app"]}


def test_a_plugin_is_asked_to_generate_only_the_kinds_it_declares(tmp_path: Path) -> None:
    report = checked(FigAppOnly(tmp_path), tmp_path)
    assert report.ok, report.failures()
    assert "discovers the generated app" in names(report)
    assert "generates a lib" not in names(report)
    assert "generates a package" not in names(report)


def test_a_plugin_that_does_not_declare_change_kind_is_not_asked(tmp_path: Path) -> None:
    report = checked(Fig(tmp_path), tmp_path)
    assert not any("edits (dry run)" in check for check in names(report))


@pytest.mark.parametrize(
    ("how", "want"),
    [
        ("nothing", "neither a refusal nor an edit"),
        ("missing", "files that do not exist"),
        ("write", "it wrote to the workspace"),
    ],
)
def test_a_kind_change_that_answers_badly_is_reported(tmp_path: Path, how: str, want: str) -> None:
    assert want in checked(FigChanger(tmp_path, how), tmp_path).failures()
