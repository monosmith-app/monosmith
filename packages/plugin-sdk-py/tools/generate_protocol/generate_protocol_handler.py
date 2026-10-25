"""The command line of the generator.

    python -m tools.generate_protocol.generate_protocol_handler            write the files
    python -m tools.generate_protocol.generate_protocol_handler --check    fail when they are stale

Run it with `smith run generate -p plugin-sdk-py` after the Go SDK's schema changes
(`smith run schema -p plugin-sdk`).
"""

from __future__ import annotations

import sys
from pathlib import Path

from .generate_protocol_use_case import generate_protocol


def main(argv: list[str]) -> int:
    """Run the generator; the exit status is 1 when `--check` finds a stale file."""
    package_root = Path(__file__).resolve().parents[2]
    check = "--check" in argv
    changed = generate_protocol(package_root, check)
    if check and changed:
        print(
            f"stale: {', '.join(changed)}\nrun `smith run generate -p plugin-sdk-py` and commit the result",
            file=sys.stderr,
        )
        return 1
    print("the generated files are current" if not changed else f"wrote {', '.join(changed)}")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
