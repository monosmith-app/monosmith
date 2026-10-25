"""Parsing a template with `[[ ]]` delimiters."""

from __future__ import annotations

import re
from collections.abc import Callable
from dataclasses import dataclass, field
from typing import Literal, NoReturn

from .template_error import TemplateError
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
from .tokenize_action_algorithm import Fail, Token, tokenize_action

OPEN = "[["
CLOSE = "]]"
SPACE = " \t\r\n"
UNSUPPORTED = frozenset({"define", "template", "block", "break", "continue"})


@dataclass
class _Frame:
    """An `if`, `with` or `range` whose `[[end]]` has not come yet."""

    type: Literal["if", "with", "range"]
    pipeline: Pipeline
    body: list[TemplateNode] = field(default_factory=list)
    otherwise: list[TemplateNode] = field(default_factory=list)
    in_else: bool = False
    chained: bool = False
    """Created by `else if`: closed together with the frame that holds it."""


def parse_template(text: str, name: str) -> list[TemplateNode]:
    """Parse a template with `[[ ]]` delimiters.

    The language is the subset of Go's `text/template` that project templates use: `[[.Field]]` and
    `[[.A.B]]`, `[[if]]`, `[[else if]]`, `[[else]]`, `[[with]]`, `[[range]]` and `[[end]]`, the
    functions `eq ne lt le gt ge not and or len index print printf`, string, number and boolean
    literals, parentheses, pipes, comments, and the `-` trim markers (`[[- x -]]`). Variables,
    `define`, `template` and `block` are not supported and are reported as such.
    """

    def fail(message: str) -> NoReturn:
        raise TemplateError(name, message)

    root: list[TemplateNode] = []
    stack: list[_Frame] = []

    def target() -> list[TemplateNode]:
        if not stack:
            return root
        top = stack[-1]
        return top.otherwise if top.in_else else top.body

    for literal, action in _split_pieces(text, fail):
        if action is None:
            if literal != "":
                target().append(TextNode(literal))
        elif action.startswith("/*"):
            if not action.endswith("*/"):
                fail("unclosed comment")
        else:
            _apply_action(action, stack, target, fail)
    if stack:
        fail(f"unclosed {stack[-1].type}: missing [[end]]")
    return root


def _split_pieces(text: str, fail: Fail) -> list[tuple[str, str | None]]:
    """The literal text and the actions of a template, in order, with the `-` trim markers applied."""
    pieces: list[tuple[str, str | None]] = []
    position = 0
    trim_next = False
    while position < len(text):
        open_at = text.find(OPEN, position)
        literal = text[position : len(text) if open_at == -1 else open_at]
        if trim_next:
            literal = literal.lstrip(SPACE)
        trim_next = False
        if open_at == -1:
            pieces.append((literal, None))
            break
        close = _find_close(text, open_at + len(OPEN))
        if close == -1:
            fail(f"unclosed action at offset {open_at}")
        action = text[open_at + len(OPEN) : close]
        if re.match(r"-[ \t\r\n]", action):
            literal = literal.rstrip(SPACE)
            action = action[1:]
        if re.search(r"[ \t\r\n]-$", action):
            trim_next = True
            action = action[:-1]
        pieces.append((literal, None))
        pieces.append(("", action.strip(SPACE)))
        position = close + len(CLOSE)
    return pieces


def _find_close(text: str, start: int) -> int:
    """The offset of the `]]` that ends the action, skipping string literals."""
    i = start
    while i < len(text):
        c = text[i]
        if c == '"':
            i += 1
            while i < len(text) and text[i] != '"':
                i += 2 if text[i] == "\\" else 1
        elif c == "`":
            i = text.find("`", i + 1)
            if i == -1:
                return -1
        elif text.startswith(CLOSE, i):
            return i
        i += 1
    return -1


def _apply_action(action: str, stack: list[_Frame], target_of: Callable[[], list[TemplateNode]], fail: Fail) -> None:
    tokens = tokenize_action(action, fail)
    first = tokens[0] if tokens else None
    keyword = first.value if first is not None and first.kind == "word" else ""
    if keyword in ("if", "with", "range"):
        stack.append(_Frame(keyword, _parse_pipeline(tokens[1:], fail)))  # type: ignore[arg-type]
    elif keyword == "else":
        _start_else(tokens, stack, fail)
    elif keyword == "end":
        if len(tokens) > 1:
            fail("unexpected tokens after end")
        _close_frames(stack, target_of, fail)
    elif keyword in UNSUPPORTED:
        fail(
            f'"{keyword}" is not supported by the SDK\'s templates (the supported subset: fields, if, else, '
            "with, range, end, and the functions eq ne lt le gt ge not and or len index print printf)"
        )
    else:
        if not tokens:
            fail("empty action")
        target_of().append(ActionNode(_parse_pipeline(tokens, fail)))


def _start_else(tokens: list[Token], stack: list[_Frame], fail: Fail) -> None:
    if not stack:
        fail("unexpected else")
    top = stack[-1]
    if top.in_else:
        fail("multiple else in one block")
    top.in_else = True
    following = tokens[1] if len(tokens) > 1 else None
    if following is not None and following.kind == "word" and following.value == "if":
        stack.append(_Frame("if", _parse_pipeline(tokens[2:], fail), chained=True))
    elif len(tokens) > 1:
        fail("unexpected tokens after else")


def _close_frames(stack: list[_Frame], target_of: Callable[[], list[TemplateNode]], fail: Fail) -> None:
    """End the innermost block, and the `else if` blocks chained to it."""
    while True:
        if not stack:
            fail("unexpected end")
        frame = stack.pop()
        target_of().append(BlockNode(frame.type, frame.pipeline, frame.body, frame.otherwise))
        if not frame.chained:
            return


def _parse_pipeline(tokens: list[Token], fail: Fail) -> Pipeline:
    pipeline, used = _pipeline_from(tokens, 0, fail)
    if used < len(tokens):
        fail('unexpected ")" in action')
    return pipeline


def _pipeline_from(tokens: list[Token], start: int, fail: Fail) -> tuple[Pipeline, int]:
    pipeline: Pipeline = []
    command: Command = []
    i = start
    while i < len(tokens):
        token = tokens[i]
        if token.kind == "close":
            break
        if token.kind == "pipe":
            if not command:
                fail("missing value before |")
            pipeline.append(command)
            command = []
        elif token.kind == "open":
            inner, following = _pipeline_from(tokens, i + 1, fail)
            if following >= len(tokens) or tokens[following].kind != "close":
                fail("unclosed (")
            command.append(PipelineOperand(inner))
            i = following
        else:
            command.append(_operand_of(token))
        i += 1
    if not command:
        fail("missing command after |" if pipeline else "missing value for the action")
    pipeline.append(command)
    return pipeline, i


def _operand_of(token: Token) -> Operand:
    if token.kind == "field":
        return FieldOperand(token.root, token.path)
    if token.kind in ("string", "number"):
        return LiteralOperand(token.value)
    if token.value == "true" or token.value == "false":
        return LiteralOperand(token.value == "true")
    if token.value == "nil":
        return NilOperand()
    return NameOperand(str(token.value))
