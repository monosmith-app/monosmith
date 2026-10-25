"""The answer to `generate` for a scaffold."""

from __future__ import annotations

from collections.abc import Mapping, Sequence
from pathlib import Path
from typing import Any

from ..protocol import GenerateParams, GenerateResult
from ..templating import write_templates
from .scaffold_data_mapper import scaffold_data


def scaffold_project(
    workspace_root: str | Path,
    params: GenerateParams,
    files: Mapping[str, str],
    *,
    data: Mapping[str, Any] | None = None,
    notes: Sequence[str] | None = None,
) -> GenerateResult:
    """The answer to `generate` for a scaffold: render the templates into the new project, never
    overwriting a file, and report what was written.

    `workspace_root` is the workspace root from `initialize`: the project's directory is relative to
    it. `files` maps a target path to template text, with `[[ ]]` actions in both
    (`"[[.Dir]]/widget.json"`). `data` adds fields for the templates, next to the ones
    `scaffold_data` lists; they win over them. `notes` are printed to the user after generation.

    A language plugin that scaffolds must record `params["stack"]` in the manifest it writes
    (`[[if .Stack]]` in a template), so discovery reports it back.
    """
    template_data = {**scaffold_data(params), **(data or {})}
    written = write_templates(workspace_root, files, template_data)
    return {"files": written, "notes": list(notes)} if notes is not None else {"files": written}
