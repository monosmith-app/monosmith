"""JSON editing that keeps key order. The cases are the Go SDK's jsondoc tests: the same input gives the same bytes."""

from __future__ import annotations

import pytest

from monosmith_plugin_sdk import JsonDocument, JsonDocumentError


def test_keeps_the_order_of_the_keys_the_text_of_leaves_and_edits_in_place() -> None:
    doc = JsonDocument.parse("""{
  "name": "@acme/web",
  "version": "1.0.0",
  "nx": {
    "targets": {
      "build": { "executor": "x" }
    }
  },
  "files": ["dist", "!**/*.map"],
  "scripts": {
    "test": "jest",
    "e2e": "a && b <c>"
  },
  "count": 1.50
}
""")
    doc.delete("nx")
    scripts = doc.object("scripts")
    assert scripts is not None
    scripts.set_string("build", "tsc -p tsconfig.lib.json")
    scripts.set_string("test", "jest --ci")
    created = doc.object("smith", True)
    assert created is not None
    created.set_string("kind", "package")
    assert (
        doc.render()
        == """{
  "name": "@acme/web",
  "version": "1.0.0",
  "files": [
    "dist",
    "!**/*.map"
  ],
  "scripts": {
    "test": "jest --ci",
    "e2e": "a && b <c>",
    "build": "tsc -p tsconfig.lib.json"
  },
  "count": 1.50,
  "smith": {
    "kind": "package"
  }
}
"""
    )


def test_refuses_what_is_not_an_object_and_writes_an_empty_object_as_braces() -> None:
    with pytest.raises(JsonDocumentError, match="not a JSON object"):
        JsonDocument.parse("[1]")
    with pytest.raises(JsonDocumentError):
        JsonDocument.parse('{"a": ')
    with pytest.raises(JsonDocumentError, match="invalid JSON value"):
        JsonDocument.parse('{"a": nope}')
    with pytest.raises(JsonDocumentError, match="invalid JSON value"):
        JsonDocument.parse('{"a": NaN}')
    with pytest.raises(JsonDocumentError, match="after the object"):
        JsonDocument.parse('{"a": 1} x')
    assert JsonDocument.parse('{"a": {}}').render() == '{\n  "a": {}\n}\n'
    assert JsonDocument.parse("{}").render() == "{}\n"


def test_makes_an_object_set_with_set_json_editable_and_writes_other_values_as_they_are() -> None:
    doc = JsonDocument.parse('{"name": "x"}')
    doc.set_json("jest", {"preset": "ts-jest", "roots": ["<rootDir>/src"]})
    jest = doc.object("jest")
    assert jest is not None
    jest.set_string("testEnvironment", "node")
    doc.set_json("private", True)
    assert (
        doc.render()
        == """{
  "name": "x",
  "jest": {
    "preset": "ts-jest",
    "roots": [
      "<rootDir>/src"
    ],
    "testEnvironment": "node"
  },
  "private": true
}
"""
    )


def test_reads_strings_and_says_no_for_what_is_not_one() -> None:
    doc = JsonDocument.parse('{"a": "x", "n": 1, "o": {}}')
    assert doc.get_string("a") == "x"
    for key in ("n", "o", "missing"):
        assert doc.get_string(key) is None
    assert doc.get_value("n") == 1
    assert doc.get_value("o") == {}
    assert doc.keys() == ["a", "n", "o"]
    assert len(doc) == 3
    assert doc.has("a") and not doc.has("z")


def test_gives_a_duplicate_key_its_first_position_and_last_value_and_puts_a_new_key_last() -> None:
    doc = JsonDocument.parse('{"a": 1, "b": 2, "a": 3}')
    doc.set_string("c", "x")
    doc.set_string("a", "y")
    assert doc.render() == '{\n  "a": "y",\n  "b": 2,\n  "c": "x"\n}\n'


def test_formats_nested_arrays_and_escapes_as_go_does() -> None:
    doc = JsonDocument.parse(r'{"matrix": [[1,2],[3,[]],{"a":[ ],"b":"x\"y"}], "s": "<>& "}')
    assert (
        doc.render()
        == r"""{
  "matrix": [
    [
      1,
      2
    ],
    [
      3,
      []
    ],
    {
      "a": [],
      "b": "x\"y"
    }
  ],
  "s": "<>& "
}
"""
    )
    doc.set_string("t", " <")
    assert '"t": "\\u2028<"' in doc.render()
