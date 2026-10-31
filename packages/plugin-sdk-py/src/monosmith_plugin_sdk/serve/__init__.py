"""Serving a plugin: the loop, the shape of a plugin, and the capabilities derived from its methods."""

from .capabilities_of_algorithm import capabilities_of, implements_method
from .plugin_contract import (
    APISurfaceReader,
    CIContributor,
    ComponentLister,
    Contributor,
    DeclarationEditor,
    DependencyManager,
    Discoverer,
    DoctorContributor,
    Generator,
    KindChanger,
    LatestReader,
    Plugin,
    Publisher,
    RegistryReader,
    Remover,
    Versioner,
)
from .plugin_members_config import PLUGIN_MEMBERS
from .protocol_streams_contract import ByteSink, ByteSource, ServeStreams
from .serve_use_case import serve, serve_streams

__all__ = [
    "PLUGIN_MEMBERS",
    "ByteSink",
    "ByteSource",
    "CIContributor",
    "Contributor",
    "APISurfaceReader",
    "ComponentLister",
    "DeclarationEditor",
    "DependencyManager",
    "Discoverer",
    "DoctorContributor",
    "Generator",
    "KindChanger",
    "LatestReader",
    "Plugin",
    "Publisher",
    "RegistryReader",
    "Remover",
    "ServeStreams",
    "Versioner",
    "capabilities_of",
    "implements_method",
    "serve",
    "serve_streams",
]
