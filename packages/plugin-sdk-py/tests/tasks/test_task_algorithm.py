"""Building tasks, contributions, CI steps, scaffolds and the small project functions."""

from __future__ import annotations

from monosmith_plugin_sdk import (
    action_step,
    audit_blocks,
    audit_enabled,
    build_dir_from_project,
    build_task,
    ci_step,
    conflicts_with,
    contribute_to,
    credential,
    default_roles,
    describe_ref,
    effective_roles,
    in_build_dir,
    is_manifest,
    project_build_dir,
    project_kind_directory,
    publishes,
    requires_language,
    resolve_project_kind,
    scaffold_data,
    set_toolchain,
    stack_plugins,
    task,
)


def test_task_leaves_out_what_is_not_given() -> None:
    assert task(["python", "check.py"]) == {"command": ["python", "check.py"]}
    assert task(["x"], cache=True, depends_on=["build"], serve_url="http://h", env={"A": "b"}) == {
        "command": ["x"],
        "cache": True,
        "dependsOn": ["build"],
        "serveUrl": "http://h",
        "env": {"A": "b"},
    }
    assert task(["x"], future="member") == {"command": ["x"], "future": "member"}


def test_build_task_builds_into_the_projects_directory_under_dist() -> None:
    built = build_task("libs/core", lambda out: ["tool", out], inputs=["widget.json"])
    assert built == {
        "command": ["tool", "../../dist/libs/core"],
        "outputs": ["../../dist/libs/core"],
        "dependsOn": ["^build"],
        "cache": True,
        "inputs": ["widget.json"],
    }


def test_contribute_to_keys_by_project_name_and_skips_none() -> None:
    projects = [
        {"name": "a", "root": "a", "language": "w", "kind": "lib"},
        {"name": "b", "root": "b", "language": "w", "kind": "lib"},
    ]
    result = contribute_to(projects, lambda p: {"tasks": {"x": task(["x"])}} if p["name"] == "a" else None)  # type: ignore[arg-type]
    assert list(result["projects"]) == ["a"]


def test_build_dir() -> None:
    assert project_build_dir("/apps/web/") == "dist/apps/web"
    assert build_dir_from_project("apps/web") == "../../dist/apps/web"
    assert in_build_dir("dist") and in_build_dir("dist/x") and not in_build_dir("distant")


def test_ci_builders() -> None:
    assert action_step("actions/setup-node@v6", {"node-version": "24"}) == {
        "uses": "actions/setup-node@v6",
        "inputs": [{"name": "node-version", "value": "24"}],
    }
    assert action_step("NodeTool@0") == {"uses": "NodeTool@0"}
    assert ci_step("audit", ["pip-audit"], blocking=False, cache_files=["a"]) == {
        "name": "audit",
        "command": ["pip-audit"],
        "blocking": False,
        "cacheFiles": ["a"],
    }
    assert credential("TOKEN", "a token", setup=True) == {
        "env": "TOKEN",
        "description": "a token",
        "required": True,
        "setup": True,
    }


def test_roles_and_refs() -> None:
    assert default_roles("language") == ["language", "toolchain", "scaffold"]
    assert default_roles("convention") == ["structure"]
    assert effective_roles({"kind": "convention", "roles": []}) == ["structure"]
    assert effective_roles({"kind": "convention", "roles": ["deploy"]}) == ["deploy"]
    assert requires_language("node") == {"level": "language", "name": "node"}
    assert conflicts_with("x") == {"name": "x"}
    assert describe_ref({"level": "language", "name": "node"}) == "language plugin node"
    assert describe_ref({"slot": "test"}) == "plugin providing test"


def test_project_functions() -> None:
    assert is_manifest(["package.json", "*.csproj"], "App.csproj")
    assert not is_manifest(["package.json", "*.csproj"], "App.vbproj")
    assert is_manifest(["[a-c].json"], "b.json") and not is_manifest(["[^a-c].json"], "b.json")
    assert not is_manifest(["[unclosed"], "x")
    assert resolve_project_kind("app", "libs/x", "package") == "app"
    assert resolve_project_kind("nonsense", "libs/x", "package") == "lib"
    assert resolve_project_kind(None, "other/x", "package") == "package"
    assert project_kind_directory("app") == "apps" and project_kind_directory("package") == "packages"
    assert publishes({"kind": "package"}) and not publishes({"kind": "lib"})
    assert not publishes({"kind": "package", "registries": {"noPublish": True}})
    assert audit_enabled({"policies": ["python"]}, "python") and not audit_enabled(None, "python")
    assert audit_blocks({"warn": ["python"]}, "python") is False and audit_blocks(None, "python")
    assert stack_plugins({"toolchain": ["a"], "structure": ["b"], "deploy": ["c"]}) == ["a", "b", "c"]


def test_set_toolchain_probes_once_and_only_with_projects() -> None:
    calls: list[int] = []

    def probe() -> str:
        calls.append(1)
        return "python 3.12"

    set_toolchain({"projects": []}, probe)
    assert calls == []
    result = set_toolchain(
        {
            "projects": [
                {"name": "a", "root": "a", "language": "l", "kind": "lib"},
                {"name": "b", "root": "b", "language": "l", "kind": "lib"},
            ]
        },
        probe,
    )
    assert calls == [1]
    assert [p["toolchain"] for p in result["projects"]] == ["python 3.12"] * 2


def test_scaffold_data() -> None:
    data = scaffold_data({"kind": "lib", "name": "core", "directory": "libs/core", "workspace": {"name": "acme"}})
    assert data == {
        "Dir": "libs/core",
        "Name": "core",
        "Kind": "lib",
        "Stack": "",
        "Workspace": "acme",
        "Scope": "",
        "ModulePrefix": "",
    }
