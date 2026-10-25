"""The JSON-RPC 2.0 envelopes: written the way the Go SDK writes them, and read."""

from __future__ import annotations

import json
from typing import Any

from .encode_json_value_algorithm import encode_json_string
from .message_parse_error import MessageParseError
from .received_message_contract import ReceivedError, ReceivedRequest, ReceivedResponse

_INT64_MIN = -(2**63)
_INT64_MAX = 2**63 - 1


def write_request(request_id: int, method: str, params_json: str | None) -> str:
    """A request as the Go SDK writes it: `jsonrpc`, `id`, `method`, `params` (left out when there are none)."""
    params = "" if params_json is None else f',"params":{params_json}'
    return f'{{"jsonrpc":"2.0","id":{request_id},"method":{encode_json_string(method)}{params}}}'


def write_notification(method: str, params_json: str) -> str:
    """A notification as the Go SDK writes it: a request with no `id`, which gets no response."""
    return f'{{"jsonrpc":"2.0","method":{encode_json_string(method)},"params":{params_json}}}'


def write_result(request_id: int | None, result_json: str) -> str:
    """A successful response; `id` is `None` only when the request could not be read."""
    return f'{{"jsonrpc":"2.0","id":{_id(request_id)},"result":{result_json}}}'


def write_failure(request_id: int | None, code: int, message: str) -> str:
    """An error response."""
    return (
        f'{{"jsonrpc":"2.0","id":{_id(request_id)},"error":{{"code":{code},"message":{encode_json_string(message)}}}}}'
    )


def _id(request_id: int | None) -> str:
    return "null" if request_id is None else str(request_id)


def _reject_constant(name: str) -> Any:
    raise ValueError(f"{name} is not JSON")


def _parse_object(body: str) -> dict[str, Any]:
    try:
        value = json.loads(body, parse_constant=_reject_constant)
    except ValueError as error:
        raise MessageParseError(str(error)) from error
    if not isinstance(value, dict):
        raise MessageParseError("a message is a JSON object")
    return value


def _read_id(message: dict[str, Any]) -> int | None:
    value = message.get("id")
    if value is None:
        return None
    if isinstance(value, bool) or not isinstance(value, int) or not _INT64_MIN <= value <= _INT64_MAX:
        raise MessageParseError("id must be an integer")
    return value


def parse_request(body: str) -> ReceivedRequest:
    """Read a request frame. A `null` or missing id makes it a notification."""
    message = _parse_object(body)
    request_id = _read_id(message)
    method = message.get("method")
    if method is None:
        method = ""
    if not isinstance(method, str):
        raise MessageParseError("method must be a string")
    return ReceivedRequest(request_id, method, message.get("params"))


def parse_response(body: str) -> ReceivedResponse:
    """Read a response frame."""
    message = _parse_object(body)
    request_id = _read_id(message)
    error = message.get("error")
    if error is None:
        return ReceivedResponse(request_id, message.get("result"), None)
    if not isinstance(error, dict):
        raise MessageParseError("error must be an object")
    code, text = error.get("code"), error.get("message")
    if isinstance(code, bool) or not isinstance(code, int) or not isinstance(text, str):
        raise MessageParseError("error needs a numeric code and a message")
    return ReceivedResponse(request_id, None, ReceivedError(code, text))
