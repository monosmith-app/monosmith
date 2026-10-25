"""Tests for the plugin: run them with `python -m unittest` from this folder.

The plugin is served in this process over pipes, with the real handshake and framing: the way smith talks to it.
"""

import shutil
import tempfile
import unittest
from pathlib import Path

from monosmith_plugin_sdk import HostOptions, check_plugin, connect_in_process
from widget import WidgetPlugin

CORE = {"name": "core", "root": "libs/core", "language": "widget", "kind": "lib"}


class WidgetPluginTest(unittest.TestCase):
    def setUp(self) -> None:
        self.workspace = Path(tempfile.mkdtemp(prefix="widget-"))
        self.addCleanup(shutil.rmtree, self.workspace, ignore_errors=True)

    def write(self, relative: str, content: str) -> None:
        path = self.workspace.joinpath(*relative.split("/"))
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")

    def connect(self):  # noqa: ANN202
        client = connect_in_process(WidgetPlugin(), HostOptions(workspace_root=str(self.workspace)))
        self.addCleanup(client.close)
        return client

    def test_the_plugin_conforms_generating_an_app_a_lib_and_a_package_that_it_discovers_again(self) -> None:
        report = check_plugin(self.connect(), self.workspace)
        self.assertTrue(report.ok, report.failures())

    def test_discover_turns_uses_into_imports(self) -> None:
        self.write("libs/core/widget.json", '{"name":"core","version":"1.0.0"}')
        self.write("apps/shop/widget.json", '{"name":"shop","version":"1.0.0","kind":"app","uses":["core"]}')
        projects = self.connect().call(
            "discover", {"manifests": ["libs/core/widget.json", "apps/shop/widget.json"]}
        )["projects"]
        core, shop = projects
        self.assertEqual((core["kind"], shop["kind"]), ("lib", "app"))
        self.assertEqual(shop["imports"], ["core"])
        self.assertEqual(core["exports"], ["core"])
        self.assertEqual(core["tasks"]["build"]["outputs"], ["../../dist/libs/core"])

    def test_a_dry_run_of_version_write_changes_nothing(self) -> None:
        self.write("libs/core/widget.json", '{\n  "name": "core",\n  "version": "1.0.0"\n}\n')
        client = self.connect()
        path = self.workspace / "libs" / "core" / "widget.json"
        dry = client.call("version/write", {"project": CORE, "version": "1.1.0", "dryRun": True})
        self.assertEqual(dry["files"], ["libs/core/widget.json"])
        self.assertIn('"1.0.0"', path.read_text(encoding="utf-8"))
        client.call("version/write", {"project": CORE, "version": "1.1.0"})
        self.assertIn('"version": "1.1.0"', path.read_text(encoding="utf-8"))

    def test_doctor_offers_a_fix_for_a_missing_license(self) -> None:
        self.write("libs/core/widget.json", '{"name":"core","version":"1.0.0"}')
        finding = self.connect().call("doctor", {"projects": [CORE]})["findings"][0]
        self.assertFalse(finding["ok"])
        self.assertTrue(finding["remedy"])
        self.assertFalse(finding["fix"]["safe"])
        self.assertIn('"license": "MIT"', finding["fix"]["files"][0]["content"])
        self.assertEqual(finding["fix"]["applies_to"], ["core"])


if __name__ == "__main__":
    unittest.main()
