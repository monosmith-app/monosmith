"""Where every build writes its output."""

from __future__ import annotations

BUILD_DIR = "dist"
"""The directory at the workspace root every build writes into.

It is smith's opinion, so all output is in one place, outside every source tree, and cleaning is
deleting one folder. A plugin builds a project into `project_build_dir`; `smith doctor` fails a task
whose outputs fall outside it.
"""


def project_build_dir(root: str) -> str:
    """Where a project's build goes, workspace-relative: `dist/<project dir>` (`dist/apps/web`)."""
    return f"{BUILD_DIR}/{root.strip('/')}"


def build_dir_from_project(root: str) -> str:
    """The same directory relative to the project (`../../dist/apps/web`).

    It is what a tool configured from inside the project, or a task's `outputs`, name.
    """
    trimmed = root.strip("/")
    return "../" * len(trimmed.split("/")) + project_build_dir(trimmed)


def in_build_dir(path: str) -> bool:
    """Whether a workspace-relative path is inside the build directory."""
    return path == BUILD_DIR or path.startswith(f"{BUILD_DIR}/")
