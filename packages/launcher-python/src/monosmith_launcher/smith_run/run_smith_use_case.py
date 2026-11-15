"""The smith program a run starts, and the run itself."""

from __future__ import annotations

import subprocess
import sys
from collections.abc import Callable, Mapping, Sequence
from pathlib import Path

from ..release_download import install_version, newest_release, system_of
from ..signature_check import RELEASE_PUBLIC_KEY
from ..version_choice import choose_version
from ..version_store import newest_installed, program_path


def resolve_program(
    env: Mapping[str, str],
    cwd: Path,
    platform: str,
    machine: str,
    say: Callable[[str], None],
    public_key: str = RELEASE_PUBLIC_KEY,
) -> Path:
    """Return the smith program a run starts.

    A pinned version (SMITH_VERSION or the workspace's pin) is downloaded when it is missing.
    Without a pin, the newest installed version runs, and only a machine with none asks for the
    newest release: from then on ``smith update`` moves it.

    Args:
        env: The environment.
        cwd: The current folder.
        platform: sys.platform.
        machine: platform.machine().
        say: Where progress goes.
        public_key: The release key; default smith's (tests sign with their own).

    Returns:
        The program's path.

    Raises:
        OSError: When the program cannot be downloaded or does not pass the checks.
        ValueError: When SMITH_VERSION is not a version, or the machine is not one smith is built for.
    """
    system = system_of(platform) or ""
    pinned = choose_version(env, cwd)
    if pinned:
        program = program_path(env, pinned, system)
        return program if program.is_file() else install_version(env, pinned, platform, machine, say, public_key)
    installed = newest_installed(env, system)
    if installed:
        return program_path(env, installed, system)
    return install_version(env, newest_release(env), platform, machine, say, public_key)


def run_smith(program: Path, args: Sequence[str], env: Mapping[str, str], launcher: str) -> int:
    """Run smith with the launcher's arguments and return its exit code.

    SMITH_LAUNCHER tells smith it was started by this launcher, so ``smith update`` installs into
    the versions folder.

    Args:
        program: The smith program.
        args: The arguments.
        env: The environment.
        launcher: The launcher's name and version (pypi@3.0.0).

    Returns:
        smith's exit code (1 when it could not start).
    """
    try:
        return subprocess.call([str(program), *args], env={**env, "SMITH_LAUNCHER": launcher})
    except OSError as error:
        sys.stderr.write(f"smith could not start: {error}\n")
        return 1
