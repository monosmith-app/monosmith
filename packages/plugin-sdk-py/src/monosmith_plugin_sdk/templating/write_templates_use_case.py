"""Writing rendered templates into a project."""

from __future__ import annotations

import posixpath
from collections.abc import Mapping
from pathlib import Path
from typing import Any

from .file_exists_error import TemplateFileExistsError
from .render_template_algorithm import render_template
from .template_error import TemplateError
from .template_functions_algorithm import by_key


def write_templates(root: str | Path, files: Mapping[str, str], data: Any = None) -> list[str]:
    """Render every template in `files` (target path to template text) under `root` with `data`,
    creating directories as needed.

    A target path may itself contain template actions (`[[.Dir]]/main.py`). Returns the written paths,
    forward-slashed and relative to `root`, in the order of the sorted templates.

    Nothing is written if any template fails to render or any target exists
    (`TemplateFileExistsError`): it never overwrites a file. A target that would land outside `root`
    is refused.
    """
    base = Path(root).resolve()
    rendered: list[tuple[str, str]] = []
    for target in sorted(files, key=by_key):
        path = posixpath.normpath(render_template(target, data, "path"))
        full = (base / Path(*path.split("/"))).resolve()
        if (
            path == ".."
            or path.startswith("../")
            or posixpath.isabs(path)
            or not full.is_relative_to(base)
            or full == base
        ):
            raise TemplateError(path, "the target is outside the project root")
        if full.exists():
            raise TemplateFileExistsError(path)
        rendered.append((path, render_template(files[target], data, path)))
    for path, content in rendered:
        full = base / Path(*path.split("/"))
        full.parent.mkdir(parents=True, exist_ok=True)
        full.write_bytes(content.encode("utf-8"))
    return [path for path, _content in rendered]
