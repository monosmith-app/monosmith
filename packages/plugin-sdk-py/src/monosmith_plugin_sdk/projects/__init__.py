"""Small functions about projects: kinds, manifests, publishing, audit settings, stacks."""

from .audit_settings_policy import audit_blocks, audit_enabled
from .manifest_match_algorithm import is_manifest
from .project_kind_policy import project_kind_directory, resolve_project_kind
from .publishes_policy import publishes
from .set_toolchain_algorithm import set_toolchain
from .stack_plugins_algorithm import stack_plugins

__all__ = [
    "audit_blocks",
    "audit_enabled",
    "is_manifest",
    "project_kind_directory",
    "publishes",
    "resolve_project_kind",
    "set_toolchain",
    "stack_plugins",
]
