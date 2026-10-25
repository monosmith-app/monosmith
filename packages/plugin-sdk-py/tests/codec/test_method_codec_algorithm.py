"""The wire form: the bytes the Go SDK writes, and params shaped as the types promise."""

from __future__ import annotations

import pytest

from monosmith_plugin_sdk import (
    PROTOCOL_SCHEMA,
    EncodeError,
    ParamsError,
    decode_params,
    decode_result,
    encode_params,
    encode_result,
)


class TestEncode:
    def test_writes_the_members_in_the_order_of_the_schema_whatever_the_order_given(self) -> None:
        text = encode_result(
            "initialize",
            {
                "capabilities": {"doctor": True, "discover": True},
                "languages": ["widget"],
                "kind": "language",
                "plugin": {"version": "0.1.0", "name": "widget"},
                "protocolVersion": 1,
            },
        )
        assert text == (
            '{"protocolVersion":1,"plugin":{"name":"widget","version":"0.1.0"},"kind":"language",'
            '"capabilities":{"discover":true,"doctor":true},"languages":["widget"]}'
        )

    def test_leaves_out_an_optional_member_that_is_absent_or_zero_and_keeps_a_required_false(self) -> None:
        text = encode_result(
            "doctor",
            {
                "findings": [
                    {"check": "a", "ok": True, "detail": "", "warning": False, "remedy": None},
                    {"check": "b", "ok": False, "detail": "broken", "warning": True},
                ]
            },
        )
        assert (
            text == '{"findings":[{"check":"a","ok":true},{"check":"b","ok":false,"detail":"broken","warning":true}]}'
        )
        assert (
            encode_result("ci", {"setup": [{"name": "x", "command": [], "blocking": False}]})
            == '{"setup":[{"name":"x","command":[],"blocking":false}]}'
        )

    def test_sorts_map_keys_and_escapes_html_characters_as_go_does(self) -> None:
        text = encode_result(
            "contribute",
            {
                "projects": {
                    "zeta": {"tasks": {"b": {"command": ["x<y>&z"]}, "a": {"command": [" "]}}},
                    "alpha": {},
                }
            },
        )
        assert text == (
            '{"projects":{"alpha":{},"zeta":{"tasks":{"a":{"command":["\\u2028"]},'
            '"b":{"command":["x\\u003cy\\u003e\\u0026z"]}}}}}'
        )

    def test_writes_non_ascii_text_as_it_is(self) -> None:
        assert encode_result("generate", {"files": ["café/€.txt"]}) == '{"files":["café/€.txt"]}'

    def test_writes_a_lone_surrogate_as_the_replacement_character_go_writes(self) -> None:
        assert encode_result("generate", {"files": ["a\ud800b"]}) == '{"files":["a�b"]}'

    def test_writes_a_struct_member_even_when_empty_and_an_empty_collection_only_when_required(self) -> None:
        assert encode_params("ci", {"projects": [], "audit": {}}) == '{"projects":[],"audit":{}}'
        assert encode_params("ci", {"projects": []}) == '{"projects":[]}'
        project = {"name": "a", "root": "r", "language": "l", "kind": "lib", "tasks": {}, "exports": []}
        assert (
            encode_result("discover", {"projects": [project]})
            == '{"projects":[{"name":"a","root":"r","language":"l","kind":"lib"}]}'
        )

    def test_keeps_free_form_members_with_sorted_keys_and_members_the_schema_does_not_know(self) -> None:
        text = encode_params(
            "generate",
            {
                "kind": "app",
                "name": "n",
                "directory": "apps/n",
                "workspace": {"name": "w"},
                "settings": {"b": 1, "a": {"d": None, "c": [1, "x"]}},
                "future": True,
            },
        )
        assert text == (
            '{"kind":"app","name":"n","directory":"apps/n","workspace":{"name":"w"},'
            '"settings":{"a":{"c":[1,"x"],"d":null},"b":1},"future":true}'
        )

    def test_writes_shutdown_as_an_empty_object_and_a_method_it_does_not_know_as_plain_json(self) -> None:
        assert encode_result("shutdown", {}) == "{}"
        assert encode_params("nope", {"b": 1, "a": [None]}) == '{"a":[null],"b":1}'

    def test_writes_floats_in_go_s_shortest_form(self) -> None:
        text = encode_params("nope", [1.5, 100000.0, 1e21, 1e-7, 0.000001, -2.5, 3])
        assert text == "[1.5,100000,1e+21,1e-7,0.000001,-2.5,3]"

    def test_names_the_required_member_that_is_missing(self) -> None:
        with pytest.raises(EncodeError, match=r"^result\.projects is required$"):
            encode_result("discover", {})
        with pytest.raises(EncodeError, match=r"result\.projects\[0\]\.root is required"):
            encode_result("discover", {"projects": [{"name": "a"}]})

    def test_names_a_member_of_the_wrong_type(self) -> None:
        cases = [
            ("discover", {"projects": "x"}, "result.projects must be an array, not a string"),
            ("generate", {"files": [1]}, "result.files[0] must be a string, not a number"),
            ("publish", {"published": "yes"}, "result.published must be a boolean"),
            ("generate", {"files": [True]}, "must be a string, not a boolean"),
        ]
        for method, value, message in cases:
            with pytest.raises(EncodeError) as caught:
                encode_result(method, value)
            assert message in str(caught.value)

    def test_refuses_an_integer_that_is_a_boolean_or_a_float(self) -> None:
        with pytest.raises(EncodeError, match="must be an integer"):
            encode_params("cancel", {"id": True})
        with pytest.raises(EncodeError, match="must be an integer"):
            encode_params("cancel", {"id": 1.5})

    def test_refuses_a_null_for_a_member_that_is_not_nullable(self) -> None:
        with pytest.raises(EncodeError, match="cannot be null"):
            encode_result("discover", {"projects": None})


class TestDecode:
    def test_turns_the_null_a_go_host_writes_for_a_nil_slice_into_an_empty_list(self) -> None:
        assert decode_params("discover", {"manifests": None}) == {"manifests": []}
        assert decode_params("discover", {}) == {"manifests": []}
        assert decode_params("discover", None) == {"manifests": []}
        project = {"name": "a", "root": "r", "language": "l", "kind": "lib", "tasks": None, "tags": None}
        assert decode_params("contribute", {"projects": [project]}) == {
            "projects": [{"name": "a", "root": "r", "language": "l", "kind": "lib"}]
        }

    def test_gives_a_required_scalar_its_zero_value_and_removes_a_null_optional_member(self) -> None:
        assert decode_params("initialize", {"host": None, "workspaceRoot": None}) == {
            "protocolVersions": [],
            "host": {"name": "", "version": ""},
            "workspaceRoot": "",
        }
        assert decode_params("dependencies/install", {"packages": None, "dev": None}) == {}

    def test_keeps_members_it_does_not_know(self) -> None:
        assert decode_params("discover", {"manifests": ["a"], "next": {"generation": 2}}) == {
            "manifests": ["a"],
            "next": {"generation": 2},
        }

    def test_refuses_a_value_of_the_wrong_json_type_naming_the_member(self) -> None:
        cases = [
            ("discover", {"manifests": "a/b.json"}, "params.manifests must be an array, not a string"),
            ("initialize", {"protocolVersions": [1.5]}, "params.protocolVersions[0] must be an integer"),
            ("discover", [], "params must be an object, not an array"),
            ("version/write", {"project": {"name": 4}, "version": "1.0.0"}, "params.project.name must be a string"),
            ("cancel", {"id": True}, "params.id must be an integer, not a boolean"),
        ]
        for method, value, message in cases:
            with pytest.raises(ParamsError) as caught:
                decode_params(method, value)
            assert message in str(caught.value)

    def test_shapes_a_result_like_params(self) -> None:
        assert decode_result("doctor", {"findings": None}) == {"findings": []}
        assert decode_result("shutdown", None) == {}


def test_the_schema_the_codec_follows_is_the_one_of_protocol_version_1() -> None:
    assert PROTOCOL_SCHEMA["x-protocolVersion"] == 1
