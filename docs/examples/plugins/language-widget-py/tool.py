"""The command line the plugin's tasks call."""

from __future__ import annotations

from pathlib import Path

from manifest import MANIFEST_FILE, read_manifest, validate


def run_tool(args: list[str]) -> None:
    """The command line the tasks call: the tasks run this same program, so the example needs no other tool.

        plugin.py check            validate ./widget.json
        plugin.py build <out-dir>  write <out-dir>/widget.txt
    """
    if not args:
        raise ValueError("usage: plugin.py check | build <out-dir> (smith runs this program as a plugin)")
    command = args[0]
    manifest = read_manifest(Path(MANIFEST_FILE))
    if command == "check":
        validate(manifest)
        print(f"{manifest.name} {manifest.version} is valid")
    elif command == "build":
        if len(args) != 2:
            raise ValueError("build takes the output directory")
        directory = Path(args[1])
        directory.mkdir(parents=True, exist_ok=True)
        out = directory / "widget.txt"
        out.write_bytes(f"{manifest.name} {manifest.version}\n".encode())
        print("wrote", out.as_posix())
    else:
        raise ValueError(f'unknown command "{command}"')
