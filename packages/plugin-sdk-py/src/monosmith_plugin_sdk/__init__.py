"""The Python SDK for writing smith plugins.

A plugin is an object with `describe` and one method for each thing it does. `serve` frames the
JSON-RPC messages, answers `initialize`, derives the capabilities from the methods the object has and
keeps standard output for the protocol:

    from monosmith_plugin_sdk import contribute_to, serve, task

    class WidgetLint:
        def describe(self, params):
            return {"plugin": {"name": "widget-lint", "version": "0.1.0"}, "kind": "convention",
                    "roles": ["structure"], "languages": ["widget"]}

        def contribute(self, params):
            return contribute_to(params["projects"], lambda project: {
                "tasks": {"layout-check": task(["python", "check.py"], cache=True)},
                "dependsOn": {"lint": ["layout-check"]},
            })

    if __name__ == "__main__":
        serve(WidgetLint())

Test it in-process, with the real handshake and framing, and run the checks of `smith plugin check`:

    from monosmith_plugin_sdk import HostOptions, check_plugin, connect_in_process

    with connect_in_process(WidgetLint(), HostOptions(workspace_root=str(tmp_path))) as client:
        report = check_plugin(client)
        assert report.ok, report.failures()

The messages of the protocol are `TypedDict`s generated from the protocol's JSON Schema; the wire
form is the one the Go and TypeScript SDKs write, byte for byte. No runtime dependencies.
"""

from . import protocol as _protocol
from .build_dir import BUILD_DIR, build_dir_from_project, in_build_dir, project_build_dir
from .cancellation import CancelToken, RequestCancelledError
from .ci import action_step, ci_step, credential
from .codec import (
    EncodeError,
    MessageParseError,
    ParamsError,
    decode_params,
    decode_result,
    encode_params,
    encode_result,
)
from .concurrency import (
    CONCURRENCY_UNLIMITED,
    UNLIMITED_IN_FLIGHT,
    UNLIMITED_WORKERS,
    in_flight_limit,
    is_valid_concurrency,
    worker_limit,
)
from .conformance import CONFORMANCE_STACK, CheckResult, ConformanceReport, check_plugin
from .framing import FrameReader, FramingError, frame_message
from .host import (
    DEFAULT_CALL_TIMEOUT,
    DEFAULT_CANCEL_GRACE,
    CallInterruptedError,
    HostOptions,
    PluginClient,
    PluginStreams,
    StartOptions,
    connect_in_process,
    connect_plugin,
    start_plugin,
)
from .jsondoc import JsonDocument, JsonDocumentError
from .plugin_refs import conflicts_with, describe_ref, is_empty_ref, requires_language
from .plugin_roles import (
    ROLES,
    default_roles,
    effective_roles,
    is_additive_role,
    is_exclusive_role,
    is_known_role,
)
from .projects import (
    audit_blocks,
    audit_enabled,
    is_manifest,
    project_kind_directory,
    publishes,
    resolve_project_kind,
    set_toolchain,
    stack_plugins,
)
from .protocol import *  # noqa: F403
from .scaffold import scaffold_data, scaffold_project
from .serve import (
    APISurfaceReader,
    CIContributor,
    ComponentLister,
    Contributor,
    DeclarationEditor,
    DependencyManager,
    Deriver,
    Discoverer,
    DoctorContributor,
    Generator,
    KindChanger,
    LatestReader,
    Plugin,
    Publisher,
    RegistryReader,
    Remover,
    ServeStreams,
    Versioner,
    capabilities_of,
    serve,
    serve_streams,
)
from .tasks import build_task, contribute_to, task
from .templating import TemplateError, TemplateFileExistsError, render_template, write_templates

__all__ = [
    *_protocol.__all__,
    # Writing a plugin.
    "serve",
    "Plugin",
    "Discoverer",
    "Generator",
    "Versioner",
    "Publisher",
    "RegistryReader",
    "CIContributor",
    "DoctorContributor",
    "Remover",
    "KindChanger",
    "DependencyManager",
    "Deriver",
    "APISurfaceReader",
    "ComponentLister",
    "DeclarationEditor",
    "LatestReader",
    "Contributor",
    "CancelToken",
    "RequestCancelledError",
    "CONCURRENCY_UNLIMITED",
    "task",
    "build_task",
    "contribute_to",
    "scaffold_project",
    "scaffold_data",
    "action_step",
    "ci_step",
    "credential",
    "BUILD_DIR",
    "project_build_dir",
    "build_dir_from_project",
    "in_build_dir",
    "ROLES",
    "default_roles",
    "effective_roles",
    "is_known_role",
    "is_exclusive_role",
    "is_additive_role",
    "requires_language",
    "conflicts_with",
    "describe_ref",
    "is_empty_ref",
    "audit_enabled",
    "audit_blocks",
    "is_manifest",
    "project_kind_directory",
    "publishes",
    "resolve_project_kind",
    "set_toolchain",
    "stack_plugins",
    "render_template",
    "write_templates",
    "TemplateError",
    "TemplateFileExistsError",
    "JsonDocument",
    "JsonDocumentError",
    # Testing a plugin.
    "HostOptions",
    "StartOptions",
    "PluginClient",
    "PluginStreams",
    "CallInterruptedError",
    "DEFAULT_CALL_TIMEOUT",
    "DEFAULT_CANCEL_GRACE",
    "start_plugin",
    "connect_plugin",
    "connect_in_process",
    "check_plugin",
    "ConformanceReport",
    "CheckResult",
    "CONFORMANCE_STACK",
    # Building blocks: the wire form of the messages, for tools that speak the protocol themselves.
    "serve_streams",
    "ServeStreams",
    "capabilities_of",
    "decode_params",
    "decode_result",
    "encode_params",
    "encode_result",
    "EncodeError",
    "MessageParseError",
    "ParamsError",
    "FrameReader",
    "FramingError",
    "frame_message",
    "UNLIMITED_IN_FLIGHT",
    "UNLIMITED_WORKERS",
    "in_flight_limit",
    "is_valid_concurrency",
    "worker_limit",
]
