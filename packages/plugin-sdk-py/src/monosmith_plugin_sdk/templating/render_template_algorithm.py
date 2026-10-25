"""Rendering a parsed template with data."""

from __future__ import annotations

from collections.abc import Mapping
from dataclasses import dataclass
from typing import Any, NoReturn

from .parse_template_algorithm import parse_template
from .template_error import TemplateError
from .template_functions_algorithm import TEMPLATE_FUNCTIONS, by_key, format_value, is_object, is_truthy
from .template_syntax_contract import (
    ActionNode,
    BlockNode,
    Command,
    FieldOperand,
    LiteralOperand,
    NameOperand,
    NilOperand,
    Operand,
    Pipeline,
    PipelineOperand,
    TemplateNode,
    TextNode,
)
from .tokenize_action_algorithm import Fail

_NONE: Any = object()
"""Stands for "nothing piped"."""


@dataclass(frozen=True)
class _Scope:
    dot: Any
    root: Any


def render_template(text: str, data: Any = None, name: str = "template") -> str:
    """Render a `[[ ]]` template with the data, the way the Go SDK's `templating` does with `text/template`
    and `missingkey=error`: a field the data does not have is an error, not an empty string.

    See `parse_template` for the language. `name` is only used in error messages.
    """

    def fail(message: str) -> NoReturn:
        raise TemplateError(name, message)

    return _render(parse_template(text, name), _Scope(data, data), fail)


def _render(nodes: list[TemplateNode], scope: _Scope, fail: Fail) -> str:
    return "".join(_render_node(node, scope, fail) for node in nodes)


def _render_node(node: TemplateNode, scope: _Scope, fail: Fail) -> str:
    if isinstance(node, TextNode):
        return node.text
    value = _evaluate_pipeline(node.pipeline, scope, fail)
    if isinstance(node, ActionNode):
        return format_value(value)
    assert isinstance(node, BlockNode)
    if node.type == "if":
        return _render(node.body if is_truthy(value) else node.otherwise, scope, fail)
    if node.type == "with":
        if is_truthy(value):
            return _render(node.body, _Scope(value, scope.root), fail)
        return _render(node.otherwise, scope, fail)
    items = _range_items(value, fail)
    if not items:
        return _render(node.otherwise, scope, fail)
    return "".join(_render(node.body, _Scope(item, scope.root), fail) for item in items)


def _range_items(value: Any, fail: Fail) -> list[Any]:
    """What `range` loops over: the elements of a list, or the values of a map in key order."""
    if isinstance(value, list | tuple):
        return list(value)
    if value is None:
        return []
    if is_object(value):
        return [value[key] for key in sorted(value, key=by_key)]
    return fail(f"range can't iterate over {format_value(value)}")


def _evaluate_pipeline(pipeline: Pipeline, scope: _Scope, fail: Fail) -> Any:
    piped: Any = _NONE
    for command in pipeline:
        piped = _evaluate_command(command, scope, piped, fail)
    return piped


def _evaluate_command(command: Command, scope: _Scope, piped: Any, fail: Fail) -> Any:
    head, *rest = command
    if isinstance(head, NameOperand):
        return _call_function(head.name, rest, scope, piped, fail)
    if rest:
        return fail("can't give an argument to a value that is not a function")
    if piped is _NONE:
        return _evaluate_operand(head, scope, fail)
    return fail("can't pipe into a value that is not a function")


def _call_function(name: str, operands: list[Operand], scope: _Scope, piped: Any, fail: Fail) -> Any:
    function = TEMPLATE_FUNCTIONS.get(name)
    if function is None:
        return fail(f'function "{name}" not defined')
    args = [_evaluate_operand(operand, scope, fail) for operand in operands]
    if piped is not _NONE:
        args.append(piped)
    try:
        return function(args)
    except Exception as error:
        return fail(f"error calling {name}: {error}")


def _evaluate_operand(operand: Operand, scope: _Scope, fail: Fail) -> Any:
    if isinstance(operand, NilOperand):
        return None
    if isinstance(operand, LiteralOperand):
        return operand.value
    if isinstance(operand, PipelineOperand):
        return _evaluate_pipeline(operand.pipeline, scope, fail)
    if isinstance(operand, NameOperand):
        return _evaluate_command([operand], scope, _NONE, fail)
    assert isinstance(operand, FieldOperand)
    value = scope.root if operand.root else scope.dot
    for key in operand.path:
        if not isinstance(value, Mapping):
            return fail(f"can't evaluate field {key} in {'nil' if value is None else _type_name(value)}")
        if key not in value:
            return fail(f'map has no entry for key "{key}"')
        value = value[key]
    return value


def _type_name(value: Any) -> str:
    if isinstance(value, bool):
        return "boolean"
    if isinstance(value, str):
        return "string"
    if isinstance(value, int | float):
        return "number"
    return "object"
