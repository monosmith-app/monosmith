"""A project's kind, and where kinds conventionally live."""

from __future__ import annotations

from ..protocol import ProjectKind

_KINDS = frozenset({"app", "lib", "package"})
_DIRECTORY_KIND: dict[str, ProjectKind] = {"apps": "app", "libs": "lib", "packages": "package"}


def resolve_project_kind(declared: str | None, root: str, fallback: ProjectKind) -> ProjectKind:
    """A project's kind when its manifest may declare none.

    The declared kind when it is valid; otherwise the directory convention (`apps/` is an app,
    `libs/` a lib, `packages/` a package); and finally `fallback`.
    """
    kind = (declared or "").strip()
    if kind in _KINDS:
        return kind  # type: ignore[return-value]
    return _DIRECTORY_KIND.get(root.split("/", 1)[0], fallback)


def project_kind_directory(kind: ProjectKind) -> str:
    """The conventional parent directory of a kind: `apps`, `libs` or `packages`."""
    return {"app": "apps", "lib": "libs"}.get(kind, "packages")
