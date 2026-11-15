"""The name of a release archive for one machine."""

from __future__ import annotations

_SYSTEMS = {"darwin": "darwin", "linux": "linux", "win32": "windows"}
_PROCESSORS = {"x86_64": "amd64", "amd64": "amd64", "x64": "amd64", "arm64": "arm64", "aarch64": "arm64"}


def system_of(platform: str) -> str | None:
    """Return the archive's name for a system.

    Args:
        platform: sys.platform.

    Returns:
        darwin, linux or windows, or None for a system smith is not built for.
    """
    return _SYSTEMS.get(platform)


def archive_name(version: str, platform: str, machine: str) -> str:
    """Return the release archive of a version for a machine: monosmith_<version>_<os>_<arch>.zip.

    Args:
        version: An exact version.
        platform: sys.platform.
        machine: platform.machine().

    Returns:
        The archive's file name.

    Raises:
        ValueError: For a system or processor smith is not built for.
    """
    system = system_of(platform)
    processor = _PROCESSORS.get(machine.lower())
    if not system or not processor:
        raise ValueError(
            f"smith is not built for {platform} on {machine}: it runs on macOS, Linux and Windows, on x64 and arm64"
        )
    return f"monosmith_{version}_{system}_{processor}.zip"
