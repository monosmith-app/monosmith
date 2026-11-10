"""Scaffolding a new project from templates."""

from .scaffold_project_use_case import scaffold_project
from .template_fields_mapper import scaffold_data

__all__ = ["scaffold_data", "scaffold_project"]
