"""Standard output is the protocol's alone: whatever else writes to it, in or out of process, goes to stderr."""

from __future__ import annotations

import sys
from pathlib import Path

from monosmith_plugin_sdk import StartOptions, start_plugin

SCRIPT = """
import subprocess, sys
from monosmith_plugin_sdk import serve

class Noisy:
    def describe(self, params):
        print("a print in describe")
        sys.stdout.write("a write to sys.stdout\\n")
        subprocess.run([sys.executable, "-c", "print('a child process, stdout inherited')"], check=True)
        # A child that reads stdin must not take the host's requests.
        subprocess.run([sys.executable, "-c", "import sys; sys.stdin.read()"], check=True)
        return {"plugin": {"name": "noisy", "version": "0.1.0"}, "kind": "convention", "roles": ["structure"]}

    def doctor(self, params):
        print("a print in doctor")
        return {"findings": [{"check": "ok", "ok": True}]}

serve(Noisy())
"""


def test_prints_and_child_processes_do_not_corrupt_the_stream(tmp_path: Path) -> None:
    script = tmp_path / "plugin.py"
    script.write_text(SCRIPT, encoding="utf-8")
    options = StartOptions(workspace_root=str(tmp_path), command=[sys.executable, str(script)], stderr="ignore")
    with start_plugin(options) as client:
        assert client.info["plugin"]["name"] == "noisy"
        assert client.call("doctor", {"projects": []}) == {"findings": [{"check": "ok", "ok": True}]}
