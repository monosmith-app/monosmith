"""Rendering `[[ ]]` templates."""

from __future__ import annotations

import pytest

from monosmith_plugin_sdk import TemplateError, render_template

DATA = {
    "Dir": "apps/web",
    "Name": "web",
    "Kind": "app",
    "Stack": "",
    "Port": 8080,
    "Tags": ["a", "b"],
    "Workspace": {"Name": "acme", "Scope": "@acme"},
    "Empty": [],
}


def test_prints_fields_dotted_fields_and_literals() -> None:
    template = '[[.Name]] in [[.Dir]] on [[.Port]] of [[.Workspace.Scope]] [[.Tags]] [[ "[[" ]] [[ 3 ]] [[ true ]]'
    assert render_template(template, DATA) == "web in apps/web on 8080 of @acme [a b] [[ 3 true"


def test_chooses_with_if_else_if_and_else() -> None:
    template = '[[if eq .Kind "app"]]app[[else if eq .Kind "lib"]]lib[[else]]other[[end]]'
    assert render_template(template, DATA) == "app"
    assert render_template(template, {**DATA, "Kind": "lib"}) == "lib"
    assert render_template(template, {**DATA, "Kind": "package"}) == "other"
    assert render_template("[[if .Stack]]stack: [[.Stack]][[end]]|[[if not .Stack]]none[[end]]", DATA) == "|none"
    assert (
        render_template(
            '[[if ne .Kind "package"]]x[[end]][[if and .Name .Port]]y[[end]][[if or .Stack .Empty]]z[[end]]', DATA
        )
        == "xy"
    )


def test_writes_the_optional_member_of_a_manifest_the_way_the_sdk_s_own_templates_do() -> None:
    template = '{\n  "license": "MIT"[[if .Stack]],\n  "stack": "[[.Stack]]"[[end]]\n}\n'
    assert render_template(template, DATA) == '{\n  "license": "MIT"\n}\n'
    assert render_template(template, {**DATA, "Stack": "widget"}) == '{\n  "license": "MIT",\n  "stack": "widget"\n}\n'


def test_narrows_with_with_loops_with_range_and_reaches_the_root_with_a_dollar() -> None:
    assert render_template("[[with .Workspace]][[.Name]]/[[$.Name]][[end]]", DATA) == "acme/web"
    assert render_template("[[range .Tags]]<[[.]]>[[end]]", DATA) == "<a><b>"
    assert render_template("[[range .Empty]]x[[else]]none[[end]]", DATA) == "none"


def test_calls_the_functions_of_text_template() -> None:
    template = (
        '[[len .Tags]] [[len .Name]] [[index .Tags 1]] [[lt .Port 9000]] [[ge .Name "web"]] '
        '[[printf "%s:%d" .Name .Port]] [[print "a" 1 2 "b"]] [[.Name | printf "%q"]]'
    )
    assert render_template(template, DATA) == '2 3 b true true web:8080 a1 2b "web"'


def test_trims_whitespace_with_the_dash_markers_and_skips_comments() -> None:
    assert render_template("a  \n[[- .Name -]]\n  b[[/* a comment */]]", DATA) == "awebb"
    assert render_template("x[[ -1 ]]", DATA) == "x-1"


@pytest.mark.parametrize(
    ("template", "message"),
    [
        ("[[.Missing]]", 'map has no entry for key "Missing"'),
        ("[[.Name.Deeper]]", "can't evaluate field Deeper in string"),
        ("[[if .Name]]x", "unclosed if: missing [[end]]"),
        ("[[end]]", "unexpected end"),
        ("[[else]]", "unexpected else"),
        ("[[.Name", "unclosed action"),
        ("[[nope .Name]]", 'function "nope" not defined'),
        ("[[eq .Name 3]]", "incompatible types for comparison"),
        ("[[$x := 1]]", "variables are not supported"),
        ('[[template "x"]]', "is not supported"),
        ("[[(.Name]]", "unclosed ("),
        ("[[.Name .Port]]", "can't give an argument"),
    ],
)
def test_fails_with_the_name_of_the_template_and_what_is_wrong(template: str, message: str) -> None:
    with pytest.raises(TemplateError) as caught:
        render_template(template, DATA, "main.go")
    assert str(caught.value).startswith("template main.go: ")
    assert message in str(caught.value)


def test_treats_nothing_as_an_empty_data_object_only_when_it_has_no_fields() -> None:
    assert render_template("plain text", None) == "plain text"
    with pytest.raises(TemplateError, match="can't evaluate field X in nil"):
        render_template("[[.X]]", None)
