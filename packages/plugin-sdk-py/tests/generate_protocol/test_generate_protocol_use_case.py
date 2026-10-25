"""The generator: the committed protocol modules are what the schema produces (the drift check)."""

from __future__ import annotations

import json
from pathlib import Path

from tools.generate_protocol import SCHEMA_PATH, generate_protocol, generated_files
from tools.generate_protocol.docstring_text_algorithm import docstring, escape_docstring
from tools.generate_protocol.python_types_mapper import declaration, type_of

PACKAGE_ROOT = Path(__file__).resolve().parents[2]


def test_the_committed_files_are_current() -> None:
    stale = generate_protocol(PACKAGE_ROOT, check=True)
    assert stale == [], f"run `smith run generate -p plugin-sdk-py` and commit the result: {stale}"


def test_generates_the_same_files_every_time() -> None:
    text = (PACKAGE_ROOT / SCHEMA_PATH).read_text(encoding="utf-8")
    assert [f.content for f in generated_files(text)] == [f.content for f in generated_files(text)]


def test_the_generated_types_cover_every_definition_of_the_schema() -> None:
    schema = json.loads((PACKAGE_ROOT / SCHEMA_PATH).read_text(encoding="utf-8"))
    types = generated_files(json.dumps(schema))[0].content
    for name in schema["$defs"]:
        assert f"\n{name} = " in types or f"\nclass {name}(" in types, name


def test_maps_schema_nodes_to_annotations() -> None:
    assert type_of({"type": "string"}) == "str"
    assert type_of({"type": ["integer", "null"]}) == "int | None"
    assert type_of({"type": "array", "items": {"$ref": "#/$defs/Task"}}) == "list[Task]"
    assert type_of({"type": "object", "additionalProperties": {"type": "string"}}) == "dict[str, str]"
    assert type_of({}) == "Any"


def test_splits_required_from_optional_members_and_keeps_enums_open_to_the_reader() -> None:
    node = {
        "type": "object",
        "properties": {"a": {"type": "string"}, "b": {"type": "boolean", "description": "B."}},
        "required": ["a"],
    }
    text = declaration("Thing", node)
    assert "class _ThingRequired(TypedDict):\n    a: str" in text
    assert "class Thing(_ThingRequired, total=False):" in text
    assert declaration("Kind", {"type": "string", "enum": ["x", "y"]}).startswith('Kind = Literal["x", "y"]')


def test_escapes_what_would_end_a_docstring() -> None:
    assert escape_docstring('a\\b """c" ') == 'a\\\\b \\"\\"\\"c" '
    assert docstring('ends with a quote"') == '"""ends with a quote\\""""\n'
