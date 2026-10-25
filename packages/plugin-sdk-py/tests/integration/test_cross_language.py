"""The compatibility rule of the SDKs: for the same plugin logic, a Python plugin, a TypeScript plugin and a Go plugin
give byte-identical responses.

Each trio below is the same plugin written three times. The one thing that cannot be equal is the command of the
tasks (the Go program's path, or an interpreter and a script), which is replaced by a name on every side. Answers to
requests that run together are compared by id: they may legitimately come out of order.
"""

from __future__ import annotations

import json
import tempfile
from pathlib import Path
from typing import Any

import pytest

from tests.integration.conftest import RawRequest, Side, by_id, converse, sides

pytestmark = pytest.mark.integration


def run_all(
    scratch: Path,
    trio: dict[str, Side],
    files: dict[str, str],
    requests: Any,
    env: dict[str, str] | None = None,
    ordered: bool = True,
) -> tuple[dict[str, list[str]], dict[str, Path]]:
    """The same requests against every side, each in its own copy of the workspace."""
    answers: dict[str, list[str]] = {}
    roots: dict[str, Path] = {}
    for name, side in trio.items():
        root = Path(tempfile.mkdtemp(prefix=f"ws-{name}-", dir=scratch))
        for relative, content in files.items():
            path = root.joinpath(*relative.split("/"))
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(content, encoding="utf-8", newline="")
        bodies = converse(side.command, root, requests(str(root)), env)
        normal = [side.normalise(body) for body in bodies]
        answers[name] = normal if ordered else by_id(normal)
        roots[name] = root
    return answers, roots


def snapshot(root: Path) -> dict[str, str]:
    return {
        path.relative_to(root).as_posix(): path.read_text(encoding="utf-8")
        for path in sorted(root.rglob("*"))
        if path.is_file()
    }


def initialize(root: str) -> RawRequest:
    return RawRequest(
        "initialize",
        {"protocolVersions": [1], "host": {"name": "smith", "version": "1.62.0"}, "workspaceRoot": root},
    )


def project(name: str, root: str, kind: str) -> dict[str, str]:
    return {"name": name, "root": root, "language": "widget", "kind": kind}


def assert_identical(answers: dict[str, list[str]]) -> None:
    reference = answers["go"]
    for name in ("ts", "py"):
        assert len(answers[name]) == len(reference), name
        for position, body in enumerate(reference, start=1):
            assert answers[name][position - 1] == body, f"{name}, response {position}"


class TestLanguageWidget:
    FILES = {
        "libs/core/widget.json": (
            '{\n  "name": "core",\n  "version": "1.0.0",\n'
            '  "description": "Core <widgets> & more",\n  "kind": "lib"\n}\n'
        ),
        "apps/shop/widget.json": (
            '{"name":"shop","version":"2.1.0","kind":"app","uses":["core","other"],'
            '"license":"MIT","stack":"widget-plain","description":"café"}'
        ),
        "libs/skip/widget.json": "{}",
    }

    def test_answer_the_handshake_discover_doctor_version_write_and_generate_with_the_same_bytes(
        self, scratch: Path
    ) -> None:
        trio = sides(scratch, "language-widget", "language-widget-ts", "language-widget-py")
        manifests = ["libs/core/widget.json", "apps/shop/widget.json", "libs/skip/widget.json"]
        projects = [project("core", "libs/core", "lib"), project("shop", "apps/shop", "app")]
        answers, roots = run_all(
            scratch,
            trio,
            self.FILES,
            lambda root: [
                initialize(root),
                RawRequest("discover", {"manifests": manifests}),
                RawRequest("discover", {"manifests": []}),
                RawRequest("doctor", {"projects": projects}),
                RawRequest("doctor", {"projects": []}),
                RawRequest("version/write", {"project": projects[0], "version": "1.1.0", "dryRun": True}),
                RawRequest("version/write", {"project": projects[0], "version": "1.1.0"}),
                RawRequest("version/write", {"project": projects[1], "version": "2.2.0"}),
                RawRequest(
                    "generate",
                    {
                        "kind": "package",
                        "name": "gen",
                        "directory": "packages/gen",
                        "workspace": {"name": "acme", "scope": "@acme"},
                        "stack": "widget-plain",
                    },
                ),
                RawRequest(
                    "generate",
                    {"kind": "app", "name": "plain", "directory": "apps/plain", "workspace": {"name": "acme"}},
                ),
                RawRequest("contribute", {"projects": projects}),
                RawRequest("nope", {}),
            ],
        )
        assert_identical(answers)
        reference = answers["go"]
        assert len(reference) > 12
        # The calls that write leave the same files, byte for byte.
        assert snapshot(roots["py"]) == snapshot(roots["go"])
        assert "packages/gen/widget.json" in snapshot(roots["py"])
        # Spot checks, so that equal does not mean equally empty.
        assert reference[0] == (
            '{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":1,"plugin":{"name":"widget","version":"0.1.0"},'
            '"kind":"language","capabilities":{"discover":true,"generate":true,"version":true,"doctor":true},'
            '"languages":["widget"],"manifests":["widget.json"],"roles":["language","scaffold"]}}'
        )
        assert "Core \\u003cwidgets\\u003e \\u0026 more" in reference[1]
        assert '"imports":["core","other"]' in reference[1]

    def test_refuse_a_request_before_initialize_a_version_they_do_not_speak_and_a_method_they_lack(
        self, scratch: Path
    ) -> None:
        trio = sides(scratch, "language-widget", "language-widget-ts", "language-widget-py")
        early, _ = run_all(scratch, trio, self.FILES, lambda root: [RawRequest("discover", {"manifests": []})])
        assert_identical(early)
        assert (
            early["go"][0] == '{"jsonrpc":"2.0","id":1,"error":{"code":-32600,"message":"initialize must come first"}}'
        )
        version, _ = run_all(
            scratch,
            trio,
            self.FILES,
            lambda root: [
                RawRequest(
                    "initialize",
                    {"protocolVersions": [2, 3], "host": {"name": "smith", "version": "9"}, "workspaceRoot": root},
                )
            ],
        )
        assert_identical(version)
        assert "no common protocol version: host speaks [2 3], plugin speaks 1" in version["go"][0]

    def test_ignore_a_cancel_notification_which_they_did_not_advertise(self, scratch: Path) -> None:
        trio = sides(scratch, "language-widget", "language-widget-ts", "language-widget-py")
        answers, _ = run_all(
            scratch,
            trio,
            self.FILES,
            lambda root: [
                initialize(root),
                RawRequest("cancel", {"id": 1}, notification=True),
                RawRequest("discover", {"manifests": []}),
            ],
        )
        assert_identical(answers)
        # initialize, discover and shutdown are answered; the notification is not, and no cancel or concurrency
        # is declared.
        assert len(answers["go"]) == 3
        assert '"cancel"' not in answers["go"][0]
        assert '"concurrency"' not in answers["go"][0]

    def test_answer_params_of_the_wrong_shape_with_the_same_code(self, scratch: Path) -> None:
        trio = sides(scratch, "language-widget", "language-widget-ts", "language-widget-py")
        answers, _ = run_all(
            scratch,
            trio,
            self.FILES,
            lambda root: [
                initialize(root),
                RawRequest("discover", {"manifests": "libs/core/widget.json"}),
                RawRequest("doctor", []),
            ],
        )
        codes = {
            name: [json.loads(body).get("error", {}).get("code") for body in bodies] for name, bodies in answers.items()
        }
        assert codes["py"] == codes["go"] == codes["ts"]
        assert codes["go"][1:3] == [-32602, -32602]


class TestStructureScan:
    FILES = {
        **{f"libs/core/{name}.txt": name for name in "abcdefgh"},
        "libs/core/data/big.bin": "x" * ((1 << 20) + 1),
    }
    PROJECTS = [project("core", "libs/core", "lib")]

    def test_answer_the_handshake_and_doctor_with_the_same_bytes_concurrency_and_cancel_included(
        self, scratch: Path
    ) -> None:
        trio = sides(scratch, "structure-scan-go", "structure-scan-node", "structure-scan-py")
        answers, _ = run_all(
            scratch,
            trio,
            self.FILES,
            lambda root: [
                initialize(root),
                RawRequest("doctor", {"projects": self.PROJECTS}),
                RawRequest("doctor", {"projects": []}),
                RawRequest("cancel", {"id": 424242}, notification=True),
                RawRequest("doctor", {"projects": self.PROJECTS}),
            ],
            ordered=False,  # the scans run together, so they are compared by id
        )
        assert_identical(answers)
        reference = answers["go"]
        assert len(reference) == 5  # the notification gets no answer
        assert reference[0] == (
            '{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":1,"plugin":{"name":"widget-scan","version":"0.1.0"},'
            '"kind":"convention","capabilities":{"doctor":true,"cancel":true},"concurrency":4,'
            '"languages":["widget"],"roles":["structure"]}}'
        )
        assert '"check":"core has no file over 1 MiB","ok":false,"detail":"data/big.bin"' in reference[1]

    def test_stop_a_scan_that_is_cancelled_and_answer_it_with_the_cancelled_error_and_the_same_bytes(
        self, scratch: Path
    ) -> None:
        trio = sides(scratch, "structure-scan-go", "structure-scan-node", "structure-scan-py")
        # A scan that would take 8 files times a minute: only the cancel ends it, so the answer is no matter of timing.
        answers, _ = run_all(
            scratch,
            trio,
            self.FILES,
            lambda root: [
                initialize(root),
                RawRequest("doctor", {"projects": self.PROJECTS}),
                RawRequest("cancel", {"id": 2}, notification=True),
                RawRequest("doctor", {"projects": []}),
            ],
            env={"WIDGET_SCAN_DELAY_MS": "60000"},
            ordered=False,
        )
        assert_identical(answers)
        assert len(answers["go"]) == 4
        assert answers["go"][1] == (
            '{"jsonrpc":"2.0","id":2,"error":{"code":-32800,"message":"request cancelled: the scan was stopped"}}'
        )


class TestStructureLayout:
    FILES = {"libs/core/README.md": "# core\n"}

    def test_answer_the_handshake_contribute_and_doctor_with_the_same_bytes(self, scratch: Path) -> None:
        trio = sides(scratch, "structure-layout-go", "structure-layout-node", "structure-layout-py")
        projects = [project("core", "libs/core", "lib"), project("web", "apps/web", "app")]
        answers, _ = run_all(
            scratch,
            trio,
            self.FILES,
            lambda root: [
                initialize(root),
                RawRequest("contribute", {"projects": projects}),
                RawRequest("contribute", {"projects": []}),
                RawRequest("doctor", {"projects": projects}),
                RawRequest("doctor", {"projects": []}),
                RawRequest("discover", {"manifests": []}),
            ],
        )
        assert_identical(answers)
        reference = answers["go"]
        assert reference[0] == (
            '{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":1,"plugin":{"name":"widget-layout","version":"0.1.0"},'
            '"kind":"convention","capabilities":{"doctor":true,"contribute":true},"languages":["widget"],'
            '"roles":["structure"]}}'
        )
        assert '"dependsOn":{"lint":["layout-check"]}' in reference[1]
        assert '"warning":true' in reference[3]
