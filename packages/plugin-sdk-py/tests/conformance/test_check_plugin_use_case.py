"""The conformance checks, run against plugins that keep and break their promises."""

from __future__ import annotations

from typing import Any

from monosmith_plugin_sdk import HostOptions, check_plugin, connect_in_process
from tests.conftest import Lab, Widget


def names(report: Any) -> list[str]:
    return [result.check for result in report.results]


def test_a_plugin_that_keeps_its_promises_passes() -> None:
    class Layout(Widget):
        def describe(self, params: Any) -> Any:
            return {"plugin": {"name": "layout", "version": "0.1.0"}, "kind": "convention", "roles": ["structure"]}

        def contribute(self, params: Any) -> Any:
            return {}

        def doctor(self, params: Any) -> Any:
            return {"findings": []}

    with connect_in_process(Layout(), HostOptions(workspace_root="/w")) as client:
        report = check_plugin(client)
    assert report.ok, report.failures()
    assert "names itself" in names(report)
    assert "a structure plugin does something" in names(report)
    assert "answers 3 requests sent together, one at a time" in names(report)
    assert "doctor answers with no projects" in names(report)


def test_a_cancellable_plugin_with_a_concurrency_is_checked_for_both() -> None:
    class Scan(Lab):
        def describe(self, params: Any) -> Any:
            return {**super().describe(params), "kind": "convention", "roles": ["structure"], "manifests": []}

        def doctor(self, params: Any) -> Any:
            return {"findings": []}

    lab = Scan(cancellable=True, concurrency=4)
    with connect_in_process(lab, HostOptions(workspace_root="/w")) as client:
        report = check_plugin(client)
    assert report.ok, report.failures()
    assert "accepts 4 requests at the same time" in names(report)
    assert "answers a cancelled request with its result or the cancelled error" in names(report)
    assert "ignores a cancel for a request it does not know" in names(report)


def test_reports_what_is_wrong_with_a_plugin_that_does_not_name_itself() -> None:
    class Nameless:
        def describe(self, params: Any) -> Any:
            return {"plugin": {"name": "", "version": ""}, "kind": "weird", "roles": ["nonsense"]}

    with connect_in_process(Nameless(), HostOptions(workspace_root="/w")) as client:
        report = check_plugin(client)
    assert not report.ok
    failed = [result.check for result in report.results if not result.ok]
    assert failed == ["names itself", "has a known kind", "plays a known role: nonsense"]
    assert report.to_dict()["results"][0]["detail"].startswith('initialize returned name ""')


def test_a_language_plugin_needs_to_discover_and_name_manifests() -> None:
    with connect_in_process(Widget(), HostOptions(workspace_root="/w")) as client:
        report = check_plugin(client)
    failed = [result.check for result in report.results if not result.ok]
    assert "a language plugin discovers projects" in failed
