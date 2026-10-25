"""Answering one request."""

from __future__ import annotations

import inspect
from typing import Any

from ..cancellation import CancelToken
from ..codec import EncodeError, ParamsError, ReceivedRequest, decode_params, encode_result
from ..protocol import ERROR_CODES
from .capabilities_of_algorithm import implements_method
from .initialize_plugin_use_case import initialize_plugin
from .plugin_members_config import PLUGIN_MEMBERS
from .request_outcome_contract import RequestOutcome, failed, succeeded


def message_of(error: BaseException) -> str:
    """The text of an exception: its message, or its class when it has none."""
    return str(error) or type(error).__name__


def accepts_cancel(function: Any) -> bool:
    """Whether a method can be given the cancel token as its second positional argument."""
    try:
        parameters = inspect.signature(function).parameters.values()
    except (TypeError, ValueError):
        return False
    positional = 0
    for parameter in parameters:
        if parameter.kind is inspect.Parameter.VAR_POSITIONAL:
            return True
        if parameter.kind in (inspect.Parameter.POSITIONAL_ONLY, inspect.Parameter.POSITIONAL_OR_KEYWORD):
            positional += 1
    return positional >= 2


def dispatch_request(plugin: Any, request: ReceivedRequest, token: CancelToken | None = None) -> RequestOutcome:
    """Answer one request: `initialize` and `shutdown` here, every other method by the member of the
    plugin that implements it.

    The codes are the Go SDK's: `METHOD_NOT_FOUND` for a method the plugin does not implement,
    `INVALID_PARAMS` for params of the wrong shape, `PLUGIN_FAILURE` for an exception the method raised
    (its message is shown to the user), `INTERNAL_ERROR` for a result that cannot be written as the
    schema describes. `token` is passed to the method as its second argument when it accepts one; it
    is cancelled when the host cancels the request.
    """
    if request.method == "shutdown":
        return succeeded("{}")
    method = request.method
    if method != "initialize" and not implements_method(plugin, method):
        return failed(ERROR_CODES.METHOD_NOT_FOUND, f"method not supported: {method}")

    try:
        params = decode_params(method, request.params)
    except ParamsError as error:
        return failed(ERROR_CODES.INVALID_PARAMS, str(error))

    try:
        if method == "initialize":
            result: Any = initialize_plugin(plugin, params)
        else:
            function = getattr(plugin, PLUGIN_MEMBERS[method].member)
            result = function(params, token or CancelToken()) if accepts_cancel(function) else function(params)
    except Exception as error:
        return failed(ERROR_CODES.PLUGIN_FAILURE, message_of(error))

    try:
        return succeeded(encode_result(method, {} if result is None else result))
    except EncodeError as error:
        return failed(ERROR_CODES.INTERNAL_ERROR, f"{method}: {error}")
    except Exception as error:
        return failed(ERROR_CODES.INTERNAL_ERROR, f"{method}: {message_of(error)}")
