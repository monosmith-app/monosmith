"""Writing templates into a project."""

from __future__ import annotations

from pathlib import Path

import pytest

from monosmith_plugin_sdk import TemplateError, TemplateFileExistsError, write_templates


def test_renders_paths_and_bodies_and_leaves_github_syntax_alone(tmp_path: Path) -> None:
    written = write_templates(
        tmp_path,
        {
            "[[.Dir]]/main.go": "package main // [[.Name]]",
            "[[.Dir]]/ci.yml": "run: ${{ secrets.TOKEN }}",
            "[[.Dir]]/sub/doc.go": "package sub",
        },
        {"Dir": "apps/web", "Name": "web"},
    )
    assert written == ["apps/web/ci.yml", "apps/web/main.go", "apps/web/sub/doc.go"]
    assert (tmp_path / "apps" / "web" / "main.go").read_text() == "package main // web"
    assert (tmp_path / "apps" / "web" / "ci.yml").read_text() == "run: ${{ secrets.TOKEN }}"


def test_refuses_to_overwrite_a_file_and_writes_nothing(tmp_path: Path) -> None:
    (tmp_path / "b.txt").write_text("mine")
    with pytest.raises(TemplateFileExistsError, match=r"b\.txt: file already exists"):
        write_templates(tmp_path, {"a.txt": "x", "b.txt": "y"})
    assert not (tmp_path / "a.txt").exists()
    assert (tmp_path / "b.txt").read_text() == "mine"


def test_fails_on_a_missing_key_writing_nothing(tmp_path: Path) -> None:
    with pytest.raises(TemplateError, match='map has no entry for key "Missing"'):
        write_templates(tmp_path, {"ok": "fine", "a": "[[.Missing]]"}, {})
    assert not (tmp_path / "ok").exists()


def test_refuses_a_target_outside_the_root(tmp_path: Path) -> None:
    (tmp_path / "inside").mkdir()
    with pytest.raises(TemplateError, match="outside the project root"):
        write_templates(tmp_path / "inside", {"[[.Up]]/x": "y"}, {"Up": ".."})
    with pytest.raises(TemplateError, match="outside the project root"):
        write_templates(tmp_path, {"../escape": "y"})
