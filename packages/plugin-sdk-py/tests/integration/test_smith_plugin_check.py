"""`smith plugin check` against the Python examples, started the way smith starts any plugin."""

from __future__ import annotations

import subprocess
import sys

import pytest

from tests.integration.conftest import EXAMPLES, clean_environment, require, runs

pytestmark = pytest.mark.integration


def check(example: str, script: str = "plugin.py", *flags: str) -> tuple[int, str]:
    """Run `smith plugin check <python> <script>`; the exit status and everything it printed."""
    require(runs("smith", "--version"), "smith")
    result = subprocess.run(
        ["smith", "plugin", "check", sys.executable, str(EXAMPLES / example / script), *flags],
        capture_output=True,
        text=True,
        encoding="utf-8",
        env=clean_environment(),
        check=False,
        timeout=300,
    )
    return result.returncode, result.stdout + result.stderr


def test_passes_the_structure_plugin_that_uses_the_sdk() -> None:
    status, out = check("structure-layout-py")
    assert "widget-layout 0.1.0" in out
    assert "✖" not in out
    assert "✔ a structure plugin does something" in out
    assert "✔ shuts down when asked" in out
    assert status == 0


def test_passes_the_language_plugin_generating_projects_and_discovering_them_again() -> None:
    status, out = check("language-widget-py", "plugin.py", "--generate")
    assert "widget 0.1.0" in out
    assert "✖" not in out
    for kind in ("app", "lib", "package"):
        assert f"✔ discovers the generated {kind}" in out
    assert status == 0


def test_passes_the_plugin_that_is_cancellable_and_declares_a_concurrency_and_checks_both() -> None:
    status, out = check("structure-scan-py")
    assert "widget-scan 0.1.0" in out
    assert "✖" not in out
    assert "✔ accepts 4 requests at the same time" in out
    assert "✔ answers a cancelled request with its result or the cancelled error" in out
    assert "✔ ignores a cancel for a request it does not know" in out
    assert status == 0


def test_still_passes_the_plugin_that_writes_the_protocol_by_hand_with_no_sdk() -> None:
    # examples/python is the protocol by hand, with a strictly serial loop that knows nothing of cancel or concurrency.
    status, out = check("python")
    assert "✖" not in out
    assert "✔ answers 3 requests sent together, one at a time" in out
    assert "cancel" not in out
    assert status == 0


def test_fails_a_plugin_that_breaks_its_promises() -> None:
    require(runs("smith", "--version"), "smith")
    result = subprocess.run(
        ["smith", "plugin", "check", sys.executable, "-c", "raise SystemExit(3)"],
        capture_output=True,
        env=clean_environment(),
        check=False,
        timeout=300,
    )
    assert result.returncode != 0
