"""The roles a plugin can play: known roles, their cardinality and the defaults of a plugin kind."""

from .plugin_role_policy import (
    ROLES,
    default_roles,
    effective_roles,
    is_additive_role,
    is_exclusive_role,
    is_known_role,
)

__all__ = ["ROLES", "default_roles", "effective_roles", "is_additive_role", "is_exclusive_role", "is_known_role"]
