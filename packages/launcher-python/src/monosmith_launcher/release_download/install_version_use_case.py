"""Downloads one smith version, checks it, and puts its program in the versions folder."""

from __future__ import annotations

import io
import os
import shutil
import zipfile
from collections.abc import Callable, Mapping
from pathlib import Path

from ..signature_check import RELEASE_PUBLIC_KEY, check_archive
from ..version_store import program_name, program_path, versions_dir
from .archive_name_mapper import archive_name, system_of
from .releases_client import release_file


def install_version(
    env: Mapping[str, str],
    version: str,
    platform: str,
    machine: str,
    say: Callable[[str], None],
    public_key: str = RELEASE_PUBLIC_KEY,
) -> Path:
    """Download a smith version, check it against the signed checksums.txt, and install its program.

    Nothing is written unless every check passes, and the program appears in one rename, so a run
    that stops halfway leaves no half-written smith behind.

    Args:
        env: The environment.
        version: An exact version.
        platform: sys.platform.
        machine: platform.machine().
        say: Where progress goes (stderr).
        public_key: The key the checksums must be signed with; default smith's release key.

    Returns:
        The installed program's path.

    Raises:
        OSError: When a file cannot be downloaded or does not pass the checks.
    """
    name = archive_name(version, platform, machine)
    system = system_of(platform) or ""
    say(f"downloading smith {version} ({name})")
    archive = release_file(env, version, name)
    checksums = release_file(env, version, "checksums.txt")
    signature = release_file(env, version, "checksums.txt.minisig")
    problem = check_archive(archive, name, checksums, signature.decode(), public_key)
    if problem:
        raise OSError(f"smith {version} was not installed: {problem}")
    program = program_name(system)
    with zipfile.ZipFile(io.BytesIO(archive)) as zipped:
        entry = next((e for e in zipped.namelist() if e == program or e.endswith("/" + program)), None)
        if entry is None:
            raise OSError(f"{name} holds no {program}")
        data = zipped.read(entry)
    target = program_path(env, version, system)
    staging = versions_dir(env) / f".{version}-{os.getpid()}"
    shutil.rmtree(staging, ignore_errors=True)
    staging.mkdir(parents=True)
    (staging / program).write_bytes(data)
    (staging / program).chmod(0o755)
    shutil.rmtree(target.parent, ignore_errors=True)
    staging.rename(target.parent)
    return target
