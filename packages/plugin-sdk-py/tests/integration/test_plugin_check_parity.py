"""`smith plugin check` prints the same lines for the Go, TypeScript and Python implementations of a plugin."""

from __future__ import annotations

import subprocess
from pathlib import Path

import pytest

from tests.integration.conftest import clean_environment, require, runs, sides

pytestmark = pytest.mark.integration


@pytest.mark.parametrize(
    ("go", "ts", "py", "flags"),
    [
        ("structure-layout-go", "structure-layout-node", "structure-layout-py", []),
        ("structure-scan-go", "structure-scan-node", "structure-scan-py", []),
        ("language-widget", "language-widget-ts", "language-widget-py", ["--generate"]),
    ],
)
def test_prints_the_same_lines(scratch: Path, go: str, ts: str, py: str, flags: list[str]) -> None:
    require(runs("smith", "--version"), "smith")
    outputs: dict[str, list[str]] = {}
    for name, side in sides(scratch, go, ts, py).items():
        result = subprocess.run(
            ["smith", "plugin", "check", *side.command, *flags],
            capture_output=True,
            text=True,
            encoding="utf-8",
            env=clean_environment(),
            check=False,
            timeout=600,
        )
        assert result.returncode == 0, f"{name}: {result.stdout}{result.stderr}"
        outputs[name] = result.stdout.splitlines()
    assert outputs["py"] == outputs["go"] == outputs["ts"]
