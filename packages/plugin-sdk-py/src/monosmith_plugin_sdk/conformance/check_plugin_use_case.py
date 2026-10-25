"""The checks of `smith plugin check`, run against a connected plugin."""

from __future__ import annotations

from pathlib import Path
from typing import Any

from ..host import PluginClient
from ..plugin_refs import describe_ref, is_empty_ref
from ..plugin_roles import ROLES, effective_roles, is_known_role
from ..protocol import PluginRef
from .check_generation_use_case import check_generation
from .check_scheduling_use_case import check_scheduling
from .conformance_report_model import ConformanceReport
from .go_list_mapper import go_list, go_quote

CALL_TIMEOUT = 20 * 60.0


def check_plugin(client: PluginClient, workspace: str | Path | None = None) -> ConformanceReport:
    """Run the checks of `smith plugin check` against a connected plugin.

    It names itself, has a known kind and known roles, its roles match what it implements, it names
    the slots it fills, its references and stacks are well-formed, every advertised method answers a
    request with nothing in it, requests sent together are answered as the declared `concurrency`
    allows, a plugin that advertises `cancel` stops a cancelled request, and, with `workspace`, the
    projects it generates are found again by its own `discover`. A port of the Go SDK's
    `conformance.Check`, with the same checks under the same names.

    `workspace` is an empty directory the plugin was started in (its `workspaceRoot`). When set, a
    plugin that generates projects is asked to generate an app, a lib and a package there, which a
    language plugin must then discover as generated. Generation may install dependencies, so it is
    opt-in.

    (`smith plugin check` adds one more, "shuts down when asked", after closing the plugin: call
    `PluginClient.close` and add it to the report if you want it.)
    """
    info = client.info
    plugin = info["plugin"]
    report = ConformanceReport(plugin)
    report.add(
        "names itself",
        plugin["name"] != "" and plugin["version"] != "",
        f"initialize returned name {go_quote(plugin['name'])}, "
        f"version {go_quote(plugin['version'])}; both are required",
    )
    report.add(
        "has a known kind",
        info["kind"] in ("language", "convention"),
        f"kind {go_quote(info['kind'])} is neither language nor convention",
    )
    roles = effective_roles(info)
    for role in roles:
        report.add(f"plays a known role: {role}", is_known_role(role), f"roles are {go_list(ROLES)}")
    caps = info.get("capabilities") or {}
    if "language" in roles:
        report.add(
            "a language plugin discovers projects",
            caps.get("discover") is True and len(info.get("manifests") or []) > 0,
            "it must implement discover and name its manifest files",
        )
    if "scaffold" in roles:
        report.add("a scaffold generates projects", caps.get("generate") is True, "it must implement generate")
    for role in ("toolchain", "structure", "deploy"):
        if role in roles and "language" not in roles:
            does_something = any(caps.get(name) is True for name in ("contribute", "generate", "ci", "doctor"))
            report.add(
                f"a {role} plugin does something",
                does_something,
                "it implements none of contribute, generate, ci and doctor",
            )
    provides = info.get("provides") or []
    if provides:
        report.add(
            "only a toolchain provides tasks",
            "toolchain" in roles,
            f"it provides {go_list(provides)} without the toolchain role",
        )
    if info["kind"] == "convention" and "toolchain" in roles:
        report.add(
            "a toolchain names the slots it fills",
            len(provides) > 0,
            "a toolchain plugin must list the tasks it provides",
        )
    for ref in info.get("requires") or []:
        report.add(
            f"requires {describe_ref(ref)}: a known reference",
            _is_valid_ref(ref),
            "a reference sets a known level, a name or a slot",
        )
    for ref in info.get("conflicts") or []:
        report.add(
            f"conflicts with {describe_ref(ref)}: a known reference",
            _is_valid_ref(ref) and ref.get("name") != plugin["name"],
            "a reference sets a known level, a name or a slot, and is not the plugin itself",
        )
    names: set[str] = set()
    for stack in info.get("stacks") or []:
        report.add(
            f"stack {stack['name']} is complete",
            stack["name"] != "" and stack["language"] != "" and stack["name"] not in names,
            "a stack needs a unique name and a language plugin",
        )
        names.add(stack["name"])

    check_scheduling(client, report)

    _answers(client, report, "discover answers with no manifests", "discover", {"manifests": []})
    _answers(client, report, "contribute answers with no projects", "contribute", {"projects": []})
    _answers(client, report, "ci answers with no projects", "ci", {"projects": []})
    _answers(client, report, "doctor answers with no projects", "doctor", {"projects": []})

    if workspace is not None and caps.get("generate") is True:
        check_generation(client, workspace, report)
    return report


def _answers(client: PluginClient, report: ConformanceReport, check: str, method: str, params: dict[str, Any]) -> None:
    if not client.supports(method):
        return
    try:
        client.call(method, params, timeout=CALL_TIMEOUT)
        report.add(check, True, "")
    except Exception as error:
        report.add(check, False, str(error))


def _is_valid_ref(ref: PluginRef) -> bool:
    """Whether a reference names something: a known level, a name or a slot."""
    level = ref.get("level") or ""
    return not is_empty_ref(ref) and (level == "" or is_known_role(level))
