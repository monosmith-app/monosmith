"""Tests for the plugin: run them with `python -m unittest` from this folder.

Nothing here depends on how long a scan takes: the scans wait on events the test controls.
"""

import shutil
import tempfile
import threading
import unittest
from pathlib import Path
from typing import Any

from monosmith_plugin_sdk import CallInterruptedError, CancelToken, HostOptions, check_plugin, connect_in_process
from scan import ScanPlugin


class SlowScan(ScanPlugin):
    """A scan that tells the test when it has started its walk."""

    def __init__(self) -> None:
        super().__init__(delay=60.0)  # a scan would take a minute a file: only a cancel ends it
        self.started = threading.Semaphore(0)

    def doctor(self, params: Any, cancel: CancelToken) -> Any:
        self.started.release()
        return super().doctor(params, cancel)


class ScanPluginTest(unittest.TestCase):
    def setUp(self) -> None:
        self.workspace = Path(tempfile.mkdtemp(prefix="scan-"))
        self.addCleanup(shutil.rmtree, self.workspace, ignore_errors=True)

    def core(self, files: dict[str, int]) -> dict[str, str]:
        for name, size in files.items():
            path = self.workspace.joinpath("libs", "core", *name.split("/"))
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_bytes(bytes(size))
        return {"name": "core", "root": "libs/core", "language": "widget", "kind": "lib"}

    def connect(self, plugin: ScanPlugin):  # noqa: ANN202
        client = connect_in_process(plugin, HostOptions(workspace_root=str(self.workspace)))
        self.addCleanup(client.close)
        return client

    def test_the_plugin_conforms(self) -> None:
        report = check_plugin(self.connect(ScanPlugin()))
        self.assertTrue(report.ok, report.failures())

    def test_it_advertises_cancel_and_a_concurrency_of_four(self) -> None:
        client = self.connect(ScanPlugin())
        self.assertTrue(client.info["capabilities"]["cancel"])
        self.assertEqual(client.info["concurrency"], 4)

    def test_doctor_warns_about_a_big_file(self) -> None:
        project = self.core({"src/small.txt": 10, "data/huge.bin": (1 << 20) + 1})
        finding = self.connect(ScanPlugin()).call("doctor", {"projects": [project]})["findings"][0]
        self.assertFalse(finding["ok"])
        self.assertTrue(finding["warning"])
        self.assertEqual(finding["detail"], "data/huge.bin")

    def test_a_cancelled_scan_stops_midway_and_the_answer_is_the_cancelled_error(self) -> None:
        project = self.core({f"{name}.txt": 1 for name in "abcdefgh"})
        plugin = SlowScan()
        client = self.connect(plugin)
        token = CancelToken()
        outcome: list[BaseException] = []

        def scan() -> None:
            try:
                client.call("doctor", {"projects": [project]}, cancel=token)
            except BaseException as error:  # noqa: BLE001
                outcome.append(error)

        thread = threading.Thread(target=scan)
        thread.start()
        self.assertTrue(plugin.started.acquire(timeout=30))
        token.cancel()
        thread.join(30)
        error = outcome[0]
        self.assertIsInstance(error, CallInterruptedError)
        assert isinstance(error, CallInterruptedError)
        self.assertEqual((error.reason, error.killed), ("cancelled", False))
        self.assertIsNotNone(error.response)
        assert error.response is not None
        self.assertEqual(error.response.code, -32800)

    def test_four_scans_run_at_the_same_time(self) -> None:
        project = self.core({"a.txt": 1, "b.txt": 1})
        plugin = SlowScan()
        client = self.connect(plugin)
        token = CancelToken()
        threads = [
            threading.Thread(target=lambda: self._ignore(client, project, token)) for _ in range(4)
        ]
        for thread in threads:
            thread.start()
        for _ in range(4):  # all four reached doctor before any of them could finish
            self.assertTrue(plugin.started.acquire(timeout=30))
        token.cancel()
        for thread in threads:
            thread.join(30)
        self.assertEqual(client.peak_in_flight, 4)

    @staticmethod
    def _ignore(client: Any, project: dict[str, str], token: CancelToken) -> None:
        try:
            client.call("doctor", {"projects": [project]}, cancel=token)
        except CallInterruptedError:
            pass


if __name__ == "__main__":
    unittest.main()
