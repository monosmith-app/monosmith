"""The error of a target file that already exists."""

from __future__ import annotations


class TemplateFileExistsError(FileExistsError):
    """A target file already exists: templates are written for a new project and never overwrite what is there."""

    def __init__(self, path: str) -> None:
        super().__init__(f"{path}: file already exists")
        self.path = path
        """The existing file, relative to the root."""

    def __str__(self) -> str:
        return f"{self.path}: file already exists"
