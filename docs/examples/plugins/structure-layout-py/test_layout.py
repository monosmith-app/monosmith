"""Tests for the plugin: run them with `python -m unittest` from this folder."""

import shutil
import tempfile
import unittest
from pathlib import Path

from layout import LayoutPlugin, find_junk_drawers
from monosmith_plugin_sdk import HostOptions, RpcError, check_plugin, connect_in_process

CORE = {"name": "core", "root": "libs/core", "language": "widget", "kind": "lib"}


class LayoutPluginTest(unittest.TestCase):
    def setUp(self) -> None:
        self.workspace = Path(tempfile.mkdtemp(prefix="layout-"))
        self.addCleanup(shutil.rmtree, self.workspace, ignore_errors=True)
        self.client = connect_in_process(LayoutPlugin(), HostOptions(workspace_root=str(self.workspace)))
        self.addCleanup(self.client.close)

    def test_the_plugin_conforms(self) -> None:
        report = check_plugin(self.client)
        self.assertTrue(report.ok, report.failures())

    def test_the_handshake_advertises_exactly_what_is_implemented(self) -> None:
        self.assertEqual(self.client.info["roles"], ["structure"])
        self.assertEqual(self.client.info["capabilities"], {"contribute": True, "doctor": True})

    def test_lint_waits_for_the_layout_check(self) -> None:
        result = self.client.call("contribute", {"projects": [CORE]})
        core = result["projects"]["core"]
        self.assertIn("layout-check", core["tasks"])
        self.assertEqual(core["dependsOn"]["lint"], ["layout-check"])

    def test_a_method_the_plugin_lacks_is_32601(self) -> None:
        with self.assertRaises(RpcError) as caught:
            self.client.call("publish", {"project": CORE, "version": "1.0.0"})
        self.assertEqual(caught.exception.code, -32601)

    def test_doctor_warns_about_a_missing_readme(self) -> None:
        (self.workspace / "libs" / "core").mkdir(parents=True)
        finding = self.client.call("doctor", {"projects": [CORE]})["findings"][0]
        self.assertFalse(finding["ok"])
        self.assertTrue(finding["warning"])
        self.assertTrue(finding["remedy"])

    def test_find_junk_drawers_skips_build_output(self) -> None:
        for folder in ("src/orders", "src/Utils", "dist/helpers", "common"):
            (self.workspace / folder).mkdir(parents=True)
        self.assertEqual(sorted(find_junk_drawers(self.workspace)), ["common", "src/Utils"])


if __name__ == "__main__":
    unittest.main()
