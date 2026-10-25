"""The data a project template can use."""

from __future__ import annotations

from ..protocol import GenerateParams


def scaffold_data(params: GenerateParams) -> dict[str, str]:
    """What a project template can say, from the `generate` params.

    | Field | Value |
    |---|---|
    | `Dir` | the project's directory, workspace-relative |
    | `Name` | the project's name |
    | `Kind` | `app`, `lib` or `package` |
    | `Stack` | the stack to record in the manifest (may be empty) |
    | `Workspace` | the workspace's name |
    | `Scope` | the npm scope, with its `@` (may be empty) |
    | `ModulePrefix` | the Go module path prefix (may be empty) |
    """
    workspace = params["workspace"]
    return {
        "Dir": params["directory"],
        "Name": params["name"],
        "Kind": params["kind"],
        "Stack": params.get("stack", ""),
        "Workspace": workspace["name"],
        "Scope": workspace.get("scope", ""),
        "ModulePrefix": workspace.get("modulePrefix", ""),
    }
