"""The plugin protocol: every message as a type, generated from the JSON Schema, and the schema as data."""

from .protocol_schema_contract import PROTOCOL_SCHEMA
from .protocol_types_contract import *  # noqa: F403
from .protocol_types_contract import __all__ as _types
from .rpc_error import RpcError

__all__ = [*_types, "PROTOCOL_SCHEMA", "RpcError"]
