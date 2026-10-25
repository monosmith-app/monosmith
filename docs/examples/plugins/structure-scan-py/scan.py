"""A structure plugin whose doctor walks the files of widget projects: it stops when smith cancels the request
and it takes four scans at the same time."""

from __future__ import annotations

import os
from pathlib import Path
from typing import Any

from monosmith_plugin_sdk import CancelToken

BIG_FILE = 1 << 20
"""The size from which a file is worth a warning."""

STOPPED = "the scan was stopped"
"""What a scan raises when it is cancelled. The SDK answers a cancelled request that fails with the request-cancelled
code, whatever the message; a fixed one keeps this plugin and its twins byte-equal."""


def big_files(directory: Path, delay: float, cancel: CancelToken) -> list[str]:
    """The files below `directory` over a mebibyte, relative to it with slashes.

    It waits `delay` seconds before each file (which stands in for a slow disk) on the cancel token, so a cancel
    ends the walk within one file's time instead of after the sleep.
    """
    found: list[str] = []

    def walk(relative: str) -> None:
        with os.scandir(directory / relative if relative else directory) as entries:
            listing = sorted(entries, key=lambda entry: entry.name)
        for entry in listing:
            path = f"{relative}/{entry.name}" if relative else entry.name
            if entry.is_dir():
                walk(path)
                continue
            if cancel.wait(delay):
                cancel.raise_if_cancelled(STOPPED)
            if entry.stat().st_size > BIG_FILE:
                found.append(path)

    walk("")
    return found


class ScanPlugin:
    """The plugin: a structure plugin whose doctor walks every file of a project.

    A walk can take a while, so it shows the two things a protocol client expects of such a method:

    - it is `cancellable`: the SDK passes each method a `CancelToken` that is cancelled when smith cancels the
      request, the walk watches it, and the plugin advertises the cancel capability;
    - it declares a `concurrency` of 4: smith may send up to four requests at once, and the SDK runs them together
      on worker threads. Scanning shares no state, so this costs nothing.
    """

    cancellable = True

    def __init__(self, delay: float = 0.0) -> None:
        self.delay = delay
        self.workspace_root = ""

    def describe(self, params: Any) -> Any:
        self.workspace_root = params["workspaceRoot"]
        return {
            "plugin": {"name": "widget-scan", "version": "0.1.0"},
            "kind": "convention",
            "roles": ["structure"],
            "languages": ["widget"],
            "concurrency": 4,
        }

    def doctor(self, params: Any, cancel: CancelToken) -> Any:
        """No file of a project is bigger than a mebibyte."""
        findings = []
        for project in params["projects"]:
            directory = Path(self.workspace_root, *project["root"].split("/"))
            check = f"{project['name']} has no file over 1 MiB"
            try:
                big = big_files(directory, self.delay, cancel)
            except OSError as error:
                raise OSError(f"scanning {directory.name}: {error}") from error
            if not big:
                findings.append({"check": check, "ok": True})
                continue
            findings.append(
                {
                    "check": check,
                    "ok": False,
                    "detail": ", ".join(big),
                    "remedy": f"keep large files out of the repository: {big[0]}",
                    "warning": True,  # reported, but does not fail smith doctor
                }
            )
        return {"findings": findings}
