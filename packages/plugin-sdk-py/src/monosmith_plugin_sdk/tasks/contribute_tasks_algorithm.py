"""The answer to `contribute` when every project gets the same kind of addition."""

from __future__ import annotations

from collections.abc import Callable, Sequence

from ..protocol import ContributeResult, Contribution, Project


def contribute_to(
    projects: Sequence[Project], contribution: Callable[[Project], Contribution | None]
) -> ContributeResult:
    """Call `contribution` for each project and key the result by project name.

    Return `None` for a project the plugin has nothing to add to; it is left out.

        def contribute(self, params):
            return contribute_to(params["projects"], lambda project: {
                "tasks": {"layout-check": task(["python", CHECKER], cache=True)},
                "dependsOn": {"lint": ["layout-check"]},
            })
    """
    result: dict[str, Contribution] = {}
    for project in projects:
        added = contribution(project)
        if added is not None:
            result[project["name"]] = added
    return {"projects": result}
