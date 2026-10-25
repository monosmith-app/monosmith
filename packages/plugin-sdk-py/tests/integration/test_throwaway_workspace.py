"""A real workspace that uses the Python examples as external plugins."""

from __future__ import annotations

import shutil
import subprocess
import sys
import tempfile
from collections.abc import Iterator
from pathlib import Path

import pytest

from monosmith_plugin_sdk import JsonDocument
from tests.integration.conftest import EXAMPLES, clean_environment, require, runs

pytestmark = pytest.mark.integration


def toml(path: str) -> str:
    """A path as a TOML string in smith.toml: forward slashes, so Windows paths need no escaping."""
    return '"' + path.replace("\\", "/") + '"'


class Workspace:
    """A `smith new` workspace with its own smith home, so the trust it records is not the user's."""

    def __init__(self, root: Path) -> None:
        self.root = root
        self.directory = root

    def run(self, *args: str) -> tuple[int, str]:
        result = subprocess.run(
            ["smith", *args],
            cwd=self.directory,
            capture_output=True,
            text=True,
            encoding="utf-8",
            env=clean_environment(SMITH_HOME=str(self.root / "smith-home")),
            check=False,
            timeout=600,
        )
        return result.returncode, result.stdout + result.stderr


@pytest.fixture(scope="module")
def workspace() -> Iterator[Workspace]:
    require(runs("smith", "--version"), "smith")
    root = Path(tempfile.mkdtemp(prefix="sdk-workspace-"))
    created = Workspace(root)
    assert created.run("new", "ws")[0] == 0
    created.directory = root / "ws"
    with (created.directory / "smith.toml").open("a", encoding="utf-8") as config:
        config.write(
            "\n"
            "[plugins.external.widget]\n"
            f"command = [{toml(sys.executable)}, {toml(str(EXAMPLES / 'language-widget-py' / 'plugin.py'))}]\n"
            "\n"
            "[plugins.external.widget-layout]\n"
            f"command = [{toml(sys.executable)}, {toml(str(EXAMPLES / 'structure-layout-py' / 'plugin.py'))}]\n"
            "\n"
            "[conventions.widget-layout]\n"
        )
    # External plugins run only once trusted; this is the user saying yes to both. --pin writes the digest to
    # smith.toml, which is what CI requires. A smith from before trust existed has no such command and needs no trust.
    for name in ("widget", "widget-layout"):
        status, out = created.run("plugin", "trust", name, "--yes", "--pin")
        if status != 0:
            assert "unknown command" in out, out
    yield created
    shutil.rmtree(root, ignore_errors=True)


def test_creates_widget_projects_with_the_plugin_and_discovers_them_with_their_graph_edge(workspace: Workspace) -> None:
    assert workspace.run("add", "lib", "core", "--lang", "widget")[0] == 0
    assert workspace.run("add", "app", "shop", "--lang", "widget")[0] == 0
    manifest = workspace.directory / "apps" / "shop" / "widget.json"
    document = JsonDocument.parse(manifest.read_text(encoding="utf-8"))
    document.set_json("uses", ["core"])
    manifest.write_text(document.render(), encoding="utf-8", newline="")

    projects = workspace.run("projects")[1]
    assert any(line.startswith("shop ") and " app " in line and "widget" in line for line in projects.splitlines())
    assert any(line.startswith("core ") and " lib " in line and "widget" in line for line in projects.splitlines())
    assert "core" in workspace.run("graph")[1]


def test_runs_the_build_and_test_tasks_and_serves_the_second_run_from_the_cache(workspace: Workspace) -> None:
    status, out = workspace.run("run", "build", "test")
    assert "tasks succeeded" in out, out
    assert status == 0
    built = workspace.directory / "dist" / "libs" / "core" / "widget.txt"
    assert built.read_text(encoding="utf-8") == "core 0.0.0\n"
    assert "from cache" in workspace.run("run", "build", "test")[1]


def test_adds_the_layout_check_task_to_every_widget_project_and_fails_on_a_junk_drawer(workspace: Workspace) -> None:
    status, out = workspace.run("run", "layout-check")
    assert "layout-check" in out
    assert status == 0
    # A file in it: the task is cached on the files of the project, and an empty folder is none.
    (workspace.directory / "libs" / "core" / "utils").mkdir()
    (workspace.directory / "libs" / "core" / "utils" / "index.txt").write_text("everything", encoding="utf-8")
    status, out = workspace.run("run", "layout-check")
    assert status != 0
    assert "folders named for nothing: utils" in out


def test_makes_the_structure_plugin_s_doctor_finding_a_warning_about_the_readme_it_asks_for(
    workspace: Workspace,
) -> None:
    assert "has a README.md" in workspace.run("doctor")[1]
