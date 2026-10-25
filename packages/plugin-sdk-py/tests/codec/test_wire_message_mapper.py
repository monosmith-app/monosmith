"""The JSON-RPC envelopes."""

from __future__ import annotations

import pytest

from monosmith_plugin_sdk import MessageParseError
from monosmith_plugin_sdk.codec import (
    parse_request,
    parse_response,
    write_failure,
    write_notification,
    write_request,
    write_result,
)


def test_writes_the_envelopes_in_the_go_sdk_member_order() -> None:
    assert write_request(3, "discover", '{"manifests":[]}') == (
        '{"jsonrpc":"2.0","id":3,"method":"discover","params":{"manifests":[]}}'
    )
    assert write_request(1, "shutdown", None) == '{"jsonrpc":"2.0","id":1,"method":"shutdown"}'
    assert write_notification("cancel", '{"id":2}') == '{"jsonrpc":"2.0","method":"cancel","params":{"id":2}}'
    assert write_result(3, "{}") == '{"jsonrpc":"2.0","id":3,"result":{}}'
    assert write_failure(None, -32700, "bad <json>") == (
        '{"jsonrpc":"2.0","id":null,"error":{"code":-32700,"message":"bad \\u003cjson\\u003e"}}'
    )


def test_reads_a_request_and_a_notification_as_a_request_without_an_id() -> None:
    request = parse_request('{"jsonrpc":"2.0","id":7,"method":"doctor","params":{"projects":[]}}')
    assert (request.id, request.method, request.params) == (7, "doctor", {"projects": []})
    assert parse_request('{"jsonrpc":"2.0","method":"note"}').id is None
    assert parse_request('{"id":null,"method":"note"}').id is None
    assert parse_request('{"id":2}').method == ""


@pytest.mark.parametrize(
    "body",
    ["", "{", "[]", '"x"', '{"id":"7","method":"a"}', '{"id":1.5,"method":"a"}', '{"id":true,"method":"a"}',
     '{"id":1,"method":4}', '{"id":NaN,"method":"a"}'],
)  # fmt: skip
def test_refuses_what_is_not_a_request_object(body: str) -> None:
    with pytest.raises(MessageParseError):
        parse_request(body)


def test_reads_a_result_and_an_error() -> None:
    response = parse_response('{"jsonrpc":"2.0","id":1,"result":{"a":1}}')
    assert (response.id, response.result, response.error) == (1, {"a": 1}, None)
    failed = parse_response('{"id":null,"error":{"code":-32700,"message":"x"}}')
    assert failed.id is None
    assert failed.error is not None
    assert (failed.error.code, failed.error.message) == (-32700, "x")
    with pytest.raises(MessageParseError):
        parse_response('{"id":1,"error":{"code":"x"}}')
