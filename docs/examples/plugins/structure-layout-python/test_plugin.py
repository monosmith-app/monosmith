"""Tests for plugin.py: run them with `python -m unittest` from this folder."""
import json
import os
import subprocess
import sys
import tempfile
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
PLUGIN = os.path.join(HERE, "plugin.py")
sys.path.insert(0, HERE)
import plugin  # noqa: E402


def talk(workspace_root, requests):
    """Start the plugin the way smith does and play a conversation: each
    request is a frame on stdin, each answer a frame on stdout."""
    frames = [{"method": "initialize", "params": {"protocolVersions": [1], "host": {"name": "test", "version": "0"},
                                                  "workspaceRoot": workspace_root}}]
    frames += requests + [{"method": "shutdown", "params": {}}]
    stdin = b""
    for i, frame in enumerate(frames):
        body = json.dumps({"jsonrpc": "2.0", "id": i + 1, **frame}).encode()
        stdin += b"Content-Length: %d\r\n\r\n" % len(body) + body
    out = subprocess.run([sys.executable, PLUGIN], input=stdin, capture_output=True, check=True,
                         env={**os.environ, "SMITH_PLUGIN": "1"}).stdout
    answers = []
    while out:
        header, _, out = out.partition(b"\r\n\r\n")
        length = int(header.split(b":")[1])
        answers.append(json.loads(out[:length]))
        out = out[length:]
    return answers


class PluginTest(unittest.TestCase):
    def test_handshake_advertises_what_is_implemented(self):
        init = talk(tempfile.gettempdir(), [])[0]["result"]
        self.assertEqual(init["protocolVersion"], 1)
        self.assertEqual(init["roles"], ["structure"])
        self.assertEqual(init["capabilities"], {"contribute": True, "doctor": True})

    def test_contribute_adds_layout_check_and_makes_lint_wait(self):
        answer = talk(tempfile.gettempdir(), [{"method": "contribute", "params": {"projects": [{"name": "core", "root": "libs/core"}]}}])[1]
        core = answer["result"]["projects"]["core"]
        self.assertEqual(core["dependsOn"], {"lint": ["layout-check"]})
        self.assertEqual(core["tasks"]["layout-check"]["command"][-1], "layout-check")

    def test_unknown_method_is_32601(self):
        answer = talk(tempfile.gettempdir(), [{"method": "publish", "params": {}}])[1]
        self.assertEqual(answer["error"]["code"], -32601)

    def test_doctor_warns_about_a_missing_readme(self):
        with tempfile.TemporaryDirectory() as ws:
            os.makedirs(os.path.join(ws, "libs", "core"))
            answer = talk(ws, [{"method": "doctor", "params": {"projects": [{"name": "core", "root": "libs/core"}]}}])[1]
            finding = answer["result"]["findings"][0]
            self.assertFalse(finding["ok"])
            self.assertTrue(finding["warning"])
            self.assertTrue(finding["remedy"])

    def test_find_junk_drawers_skips_build_output(self):
        with tempfile.TemporaryDirectory() as d:
            for sub in ("src/orders", "src/Utils", "dist/helpers", "common"):
                os.makedirs(os.path.join(d, *sub.split("/")))
            self.assertEqual(sorted(plugin.find_junk_drawers(d)), ["common", "src/Utils"])


if __name__ == "__main__":
    unittest.main()
