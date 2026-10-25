"""Runs a Python module with the workspace's virtual environment, on any operating system.

    python -m tools.run_in_venv.run_in_venv_handler -m ruff check .

smith.toml has one `lint` command for every project (golangci-lint, for the Go ones), so this
project's own override cannot name `.venv/Scripts/python.exe` (Windows) and `.venv/bin/python` (the rest)
at once. Any `python` can start this file; it hands over to the venv's.
"""

from __future__ import annotations

import subprocess
import sys
from pathlib import Path

from .venv_interpreter_algorithm import find_venv_interpreter


def main(argv: list[str]) -> int:
    """Run the venv's interpreter with `argv`; its exit status is ours."""
    interpreter = find_venv_interpreter(Path.cwd())
    if interpreter is None:
        print("no .venv above the current folder: run `smith ci setup` first", file=sys.stderr)
        return 1
    return subprocess.run([str(interpreter), *argv], check=False).returncode


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
