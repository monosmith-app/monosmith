"""The serve loop: handshake, dispatch, concurrency, cancel and shutdown, over pipes, with latches and no sleeps."""

from __future__ import annotations

import io
import json
import threading
from collections.abc import Callable
from typing import TYPE_CHECKING, Any

import pytest

from monosmith_plugin_sdk import CancelToken, FramingError, ServeStreams, frame_message, serve_streams
from tests.conftest import Lab, Widget

if TYPE_CHECKING:
    from tests.conftest import Conversation

Converse = Callable[[Any], "Conversation"]
WAIT = 30.0


def discover(conversation: Conversation, request_id: int, manifest: str) -> None:
    conversation.request(request_id, "discover", {"manifests": [manifest]})


class TestHandshake:
    def test_answers_initialize_with_the_version_it_chose_and_the_capabilities_derived_from_its_methods(
        self, converse: Converse
    ) -> None:
        class Plugin(Widget):
            def discover(self, params: Any) -> Any:
                return {"projects": []}

            def write_version(self, params: Any) -> Any:
                return {"files": []}

            def install(self, params: Any) -> Any:
                return {}

            def uninstall(self, params: Any) -> Any:
                return {}

        conversation = converse(Plugin())
        conversation.request(
            1, "initialize", {"protocolVersions": [1], "host": {"name": "t", "version": "0"}, "workspaceRoot": "/ws"}
        )
        assert conversation.recv() == (
            '{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":1,"plugin":{"name":"widget","version":"0.1.0"},'
            '"kind":"language","capabilities":{"discover":true,"version":true,"dependencies":true},'
            '"manifests":["widget.json"]}}'
        )

    def test_writes_no_capabilities_for_a_plugin_without_methods_and_ignores_what_describe_says_about_them(
        self, converse: Converse
    ) -> None:
        class Meddling:
            def describe(self, params: Any) -> Any:
                return {
                    "plugin": {"name": "w", "version": "1"},
                    "kind": "language",
                    "protocolVersion": 9,
                    "capabilities": {"publish": True},
                }

        result = converse(Meddling()).handshake()
        assert result["capabilities"] == {}
        assert result["protocolVersion"] == 1

    def test_refuses_a_host_that_does_not_speak_version_1_with_the_go_sdk_message_and_code(
        self, converse: Converse
    ) -> None:
        conversation = converse(Widget())
        conversation.request(
            1, "initialize", {"protocolVersions": [2, 3], "host": {"name": "t", "version": "0"}, "workspaceRoot": "/"}
        )
        assert conversation.recv_json()["error"] == {
            "code": -32000,
            "message": "no common protocol version: host speaks [2 3], plugin speaks 1",
        }

    def test_refuses_every_request_before_initialize_shutdown_included_and_goes_on(self, converse: Converse) -> None:
        conversation = converse(Widget())
        conversation.request(1, "shutdown", {})
        conversation.request(2, "discover", {})
        first, second = conversation.recv_json(), conversation.recv_json()
        assert first["error"] == {"code": -32600, "message": "initialize must come first"}
        assert second["error"]["code"] == -32600
        assert conversation.handshake()["protocolVersion"] == 1

    def test_allows_a_second_initialize(self, converse: Converse) -> None:
        conversation = converse(Widget())
        conversation.handshake()
        conversation.request(
            2, "initialize", {"protocolVersions": [1], "host": {"name": "t", "version": "0"}, "workspaceRoot": "/"}
        )
        assert conversation.recv_json()["result"]["protocolVersion"] == 1


class TestDispatch:
    @pytest.fixture
    def conversation(self, converse: Converse) -> Conversation:
        class Plugin(Widget):
            def discover(self, params: Any) -> Any:
                return {
                    "projects": [
                        {"name": m, "root": m, "language": "widget", "kind": "lib"} for m in params["manifests"]
                    ]
                }

            def doctor(self, params: Any) -> Any:
                raise RuntimeError("the toolchain is missing: install python 3.12")

            def generate(self, params: Any) -> Any:
                return {"files": "x"}

            def publish(self, params: Any) -> Any:
                return {"published": True}

            def remove(self, params: Any) -> Any:
                return None  # a result with no member at all

        conversation = converse(Plugin())
        conversation.handshake()
        return conversation

    def test_calls_the_method_with_the_params_shaped_as_the_types_promise(self, conversation: Conversation) -> None:
        conversation.request(2, "discover", {"manifests": None})
        assert conversation.recv_json()["result"] == {"projects": []}
        conversation.request(3, "discover", {"manifests": ["a"]})
        assert conversation.recv_json()["result"]["projects"][0]["name"] == "a"
        project = {"name": "a", "root": "a", "language": "w", "kind": "lib"}
        conversation.request(4, "publish", {"project": project, "version": "1.0.0"})
        assert conversation.recv_json()["result"] == {"published": True}

    def test_answers_a_method_the_plugin_lacks_with_method_not_found_whether_or_not_the_protocol_knows_it(
        self, conversation: Conversation
    ) -> None:
        conversation.request(2, "contribute", {})
        assert conversation.recv_json()["error"] == {"code": -32601, "message": "method not supported: contribute"}
        conversation.request(3, "nope", {})
        assert conversation.recv_json()["error"] == {"code": -32601, "message": "method not supported: nope"}

    def test_turns_what_a_method_raises_into_plugin_failure_with_its_message(self, conversation: Conversation) -> None:
        conversation.request(2, "doctor", {"projects": []})
        assert conversation.recv_json()["error"] == {
            "code": -32000,
            "message": "the toolchain is missing: install python 3.12",
        }

    def test_answers_params_of_the_wrong_shape_with_invalid_params(self, conversation: Conversation) -> None:
        conversation.request(2, "discover", {"manifests": "a"})
        assert conversation.recv_json()["error"]["code"] == -32602

    def test_answers_a_result_that_cannot_be_written_with_internal_error(self, conversation: Conversation) -> None:
        conversation.request(2, "generate", {"kind": "app", "name": "n", "directory": "d", "workspace": {"name": "w"}})
        error = conversation.recv_json()["error"]
        assert error["code"] == -32603
        assert "result.files must be an array" in error["message"]

    def test_writes_an_empty_object_for_a_method_that_returned_nothing(self, conversation: Conversation) -> None:
        project = {"name": "a", "root": "a", "language": "w", "kind": "lib"}
        conversation.request(2, "remove", {"project": project})
        assert conversation.recv() == '{"jsonrpc":"2.0","id":2,"result":{}}'

    def test_answers_shutdown_with_an_empty_object_and_stops_reading(self, converse: Converse) -> None:
        conversation = converse(Widget())
        conversation.handshake()
        conversation.request(2, "shutdown", {})
        conversation.request(3, "discover", {})
        assert conversation.recv() == '{"jsonrpc":"2.0","id":2,"result":{}}'
        assert conversation.recv_all() == []

    def test_answers_no_notification_and_an_unreadable_body_with_parse_error_and_an_id_of_null(
        self, converse: Converse
    ) -> None:
        conversation = converse(Widget())
        conversation.handshake()
        conversation.notify("discover", {})
        conversation.send_raw("{not json")
        conversation.send_raw("[]")
        conversation.request(4, "shutdown")
        answers = [json.loads(body) for body in conversation.recv_all()]
        assert [a["id"] for a in answers] == [None, None, 4]
        assert answers[0]["error"]["code"] == -32700
        assert answers[1]["error"]["code"] == -32700

    def test_calls_a_method_on_the_plugin_object_so_it_can_use_self(self, converse: Converse) -> None:
        class Counting(Widget):
            count = 0

            def doctor(self, params: Any) -> Any:
                self.count += 1
                return {"findings": [{"check": f"call {self.count}", "ok": True}]}

        conversation = converse(Counting())
        conversation.handshake()
        conversation.request(2, "doctor", {})
        conversation.request(3, "doctor", {})
        assert conversation.recv_json()["result"]["findings"][0]["check"] == "call 1"
        assert conversation.recv_json()["result"]["findings"][0]["check"] == "call 2"

    def test_passes_the_cancel_token_only_to_a_method_that_takes_it(self, converse: Converse) -> None:
        seen: list[object] = []

        class Plugin(Widget):
            def doctor(self, params: Any, cancel: CancelToken) -> Any:
                seen.append(cancel)
                return {"findings": []}

            def ci(self, params: Any) -> Any:
                seen.append("one argument")
                return {}

            def contribute(self, *args: Any) -> Any:
                seen.append(len(args))
                return {}

        conversation = converse(Plugin())
        conversation.handshake()
        for number, method in enumerate(("doctor", "ci", "contribute"), start=2):
            conversation.request(number, method, {"projects": []})
            conversation.recv_json()
        assert isinstance(seen[0], CancelToken)
        assert seen[1:] == ["one argument", 2]


class TestStreams:
    def test_warns_about_half_of_a_pair_of_methods_which_is_not_advertised_and_not_called(
        self, converse: Converse
    ) -> None:
        class Plugin(Widget):
            def install(self, params: Any) -> Any:
                return {}

        conversation = converse(Plugin())
        assert conversation.handshake()["capabilities"] == {}
        conversation.request(2, "dependencies/install", {})
        assert conversation.recv_json()["error"]["code"] == -32601
        assert len(conversation.logs) == 1
        assert "install needs its pair" in conversation.logs[0]

    def test_ends_without_an_answer_when_the_host_closes_the_stream(self, converse: Converse) -> None:
        conversation = converse(Widget())
        conversation.close_input()
        assert conversation.recv_all() == []

    def test_raises_when_the_header_is_not_a_frame_header(self) -> None:
        streams = ServeStreams(io.BytesIO(b"hello\r\n\r\n"), io.BytesIO(), lambda message: None)
        with pytest.raises(FramingError, match="malformed header"):
            serve_streams(Widget(), streams)

    def test_ends_when_the_host_is_gone_and_answers_what_is_still_running(self, converse: Converse) -> None:
        lab = Lab()
        conversation = converse(lab)
        conversation.handshake()
        discover(conversation, 2, "stubborn")
        lab.wait_started(1)
        conversation.close_input()  # the host is gone, but a plugin that cannot be cancelled finishes what it does
        lab.release.set()
        assert json.loads(conversation.recv_all()[0])["id"] == 2

    def test_reads_a_request_in_pieces(self, converse: Converse) -> None:
        conversation = converse(Widget())
        data = frame_message('{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersions":[1]}}')
        for start in range(0, len(data), 7):
            conversation.send_bytes(data[start : start + 7])
        assert conversation.recv_json()["result"]["protocolVersion"] == 1


class TestConcurrency:
    def test_handles_one_request_at_a_time_in_order_for_a_plugin_that_declares_nothing(
        self, converse: Converse
    ) -> None:
        lab = Lab()
        conversation = converse(lab)
        result = conversation.handshake()
        assert "concurrency" not in result
        assert "cancel" not in result["capabilities"]
        for request_id in range(2, 8):
            discover(conversation, request_id, f"quick-{request_id}")
        assert [conversation.recv_json()["id"] for _ in range(6)] == [2, 3, 4, 5, 6, 7]
        assert lab.peak == 1

    def test_runs_as_many_handlers_together_as_the_plugin_declared_and_no_more(self, converse: Converse) -> None:
        lab = Lab(concurrency=3)
        conversation = converse(lab)
        assert conversation.handshake()["concurrency"] == 3
        for request_id in range(2, 10):
            discover(conversation, request_id, f"block-{request_id}")
        lab.wait_started(3)
        assert lab.peak == 3
        assert len(lab.started) == 3  # the other five wait for a place; the pool never starts a fourth
        lab.release.set()
        answered = {conversation.recv_json()["id"] for _ in range(8)}
        assert answered == set(range(2, 10))
        assert lab.peak == 3

    def test_may_answer_out_of_order_each_answer_carrying_the_id_of_its_request(self, converse: Converse) -> None:
        lab = Lab(concurrency=2)
        conversation = converse(lab)
        conversation.handshake()
        discover(conversation, 2, "stubborn-first")
        lab.wait_started(1)
        discover(conversation, 3, "quick")
        answer = conversation.recv_json()
        assert answer["id"] == 3
        assert answer["result"]["projects"][0]["name"] == "quick"
        lab.release.set()
        assert conversation.recv_json()["id"] == 2

    def test_refuses_a_request_id_that_is_already_in_flight_and_a_host_that_floods_it(self, converse: Converse) -> None:
        lab = Lab(concurrency=1)
        conversation = converse(lab)
        conversation.handshake()
        discover(conversation, 2, "stubborn-a")
        lab.wait_started(1)
        discover(conversation, 2, "stubborn-again")
        assert conversation.recv_json()["error"]["code"] == -32600
        for request_id in range(3, 3 + 1024 + 5):
            discover(conversation, request_id, "stubborn-queued")
        refused = [conversation.recv_json() for _ in range(5)]
        for answer in refused:
            assert answer["error"]["code"] == -32600
            assert "too many requests" in answer["error"]["message"]
        lab.release.set()
        assert len({conversation.recv_json()["id"] for _ in range(1 + 1024)}) == 1025


class TestCancel:
    def test_is_advertised_by_a_plugin_that_says_it_is_cancellable(self, converse: Converse) -> None:
        assert converse(Lab(cancellable=True)).handshake()["capabilities"]["cancel"] is True

    def test_sets_the_token_of_a_running_request_and_answers_it_with_the_cancelled_error(
        self, converse: Converse
    ) -> None:
        lab = Lab(cancellable=True)
        conversation = converse(lab)
        conversation.handshake()
        discover(conversation, 2, "block-a")
        lab.wait_started(1)
        conversation.notify("cancel", {"id": 2})
        assert conversation.recv() == (
            '{"jsonrpc":"2.0","id":2,"error":{"code":-32800,"message":"request cancelled: aborted"}}'
        )
        discover(conversation, 3, "quick")
        assert conversation.recv_json()["result"]["projects"][0]["name"] == "quick"

    def test_ignores_a_cancel_for_an_unknown_or_finished_request_or_with_bad_params(self, converse: Converse) -> None:
        lab = Lab(cancellable=True)
        conversation = converse(lab)
        conversation.handshake()
        discover(conversation, 2, "quick")
        conversation.recv_json()
        conversation.notify("cancel", {"id": 2})
        conversation.notify("cancel", {"id": 424242})
        conversation.notify("cancel", "not params")
        discover(conversation, 3, "quick")
        assert conversation.recv_json()["id"] == 3  # nothing else was written: the next frame is this answer
        assert conversation.logs == []

    def test_answers_with_the_result_a_request_produced_despite_the_cancel(self, converse: Converse) -> None:
        lab = Lab(cancellable=True)
        conversation = converse(lab)
        conversation.handshake()
        discover(conversation, 2, "stubborn")
        lab.wait_started(1)
        conversation.notify("cancel", {"id": 2})
        lab.release.set()
        assert conversation.recv_json()["result"]["projects"][0]["name"] == "stubborn"

    def test_keeps_the_plugin_failure_code_for_an_error_nobody_asked_for(self, converse: Converse) -> None:
        conversation = converse(Lab(cancellable=True))
        conversation.handshake()
        discover(conversation, 2, "fail-by-itself")
        assert conversation.recv_json()["error"] == {"code": -32000, "message": "the manifest is broken"}

    def test_answers_a_request_that_had_not_started_at_once_and_never_starts_it(self, converse: Converse) -> None:
        lab = Lab(cancellable=True, concurrency=1)
        conversation = converse(lab)
        conversation.handshake()
        discover(conversation, 2, "stubborn-running")
        discover(conversation, 3, "stubborn-waiting")
        lab.wait_started(1)
        conversation.notify("cancel", {"id": 3})
        waiting = conversation.recv_json()
        assert waiting["id"] == 3
        assert waiting["error"] == {"code": -32800, "message": "request cancelled before it started"}
        lab.release.set()
        assert conversation.recv_json()["id"] == 2
        assert lab.started == ["stubborn-running"]

    def test_is_ignored_by_a_plugin_that_did_not_say_it_is_cancellable_and_its_token_never_cancels(
        self, converse: Converse
    ) -> None:
        lab = Lab()
        conversation = converse(lab)
        conversation.handshake()
        discover(conversation, 2, "stubborn")
        lab.wait_started(1)
        conversation.notify("cancel", {"id": 2})
        lab.release.set()
        assert conversation.recv_json()["result"]["projects"][0]["name"] == "stubborn"

    @pytest.mark.parametrize("cancellable", [True, False])
    def test_answers_what_is_in_flight_and_what_is_waiting_with_their_results_before_shutdown(
        self, converse: Converse, cancellable: bool
    ) -> None:
        lab = Lab(cancellable=cancellable, concurrency=1)
        conversation = converse(lab)
        conversation.handshake()
        discover(conversation, 2, "stubborn-running")
        discover(conversation, 3, "stubborn-waiting")
        conversation.request(4, "shutdown")
        lab.wait_started(1)
        lab.release.set()
        answers = [conversation.recv_json() for _ in range(3)]
        assert [a["id"] for a in answers] == [2, 3, 4]
        assert answers[0]["result"]["projects"][0]["name"] == "stubborn-running"
        assert answers[1]["result"]["projects"][0]["name"] == "stubborn-waiting"

    def test_cancels_what_is_in_flight_when_the_input_ends_the_host_being_gone(self, converse: Converse) -> None:
        lab = Lab(cancellable=True)
        conversation = converse(lab)
        conversation.handshake()
        discover(conversation, 2, "block-a")
        lab.wait_started(1)
        conversation.close_input()
        conversation.finish()
        assert lab.aborted == 1

    def test_answers_each_of_a_stress_of_concurrent_and_cancelled_requests_exactly_once(
        self, converse: Converse
    ) -> None:
        total = 300
        lab = Lab(cancellable=True, concurrency=8)
        conversation = converse(lab)
        conversation.handshake()

        def send_all() -> None:
            for request_id in range(2, total + 2):
                discover(conversation, request_id, "block-x" if request_id % 3 == 0 else "quick-x")
                if request_id % 6 == 0:
                    conversation.notify("cancel", {"id": request_id})
            lab.release.set()  # whatever was not cancelled finishes

        # The answers are read while the requests are written: a full pipe in either direction would stall both ends.
        sender = threading.Thread(target=send_all)
        sender.start()
        answers = [conversation.recv_json() for _ in range(total)]
        counts: dict[int, int] = {}
        for answer in answers:
            counts[answer["id"]] = counts.get(answer["id"], 0) + 1
            if "error" in answer:
                assert answer["error"]["code"] == -32800
        assert set(counts) == set(range(2, total + 2))
        assert set(counts.values()) == {1}
        sender.join()
        assert lab.peak <= 8
