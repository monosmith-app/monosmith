"""The smith command pip installs: a launcher, not smith itself (ADR 0101).

It runs the smith the workspace pins in smith.toml (``[workspace] smith``), SMITH_VERSION, or the
newest installed, and downloads and checks that version the first time.
"""

from __future__ import annotations

import os
import platform
import sys
from importlib.metadata import PackageNotFoundError, version
from pathlib import Path

from .smith_run import resolve_program, run_smith


def main() -> int:
    """Run smith with the command line's arguments.

    Returns:
        smith's exit code, or 1 when it could not be found or started.
    """

    def say(line: str) -> None:
        sys.stderr.write(f"smith launcher: {line}\n")

    try:
        own = version("monosmith")
    except PackageNotFoundError:
        own = "0.0.0"
    try:
        program = resolve_program(os.environ, Path.cwd(), sys.platform, platform.machine(), say)
    except (OSError, ValueError) as error:
        say(str(error))
        return 1
    return run_smith(program, sys.argv[1:], os.environ, f"pypi@{own}")


if __name__ == "__main__":
    sys.exit(main())
