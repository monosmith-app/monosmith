"""The parsed form of a template."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Literal


@dataclass(frozen=True)
class FieldOperand:
    """A field of the data: `.A.B`, or `$.A` from the root."""

    root: bool
    path: tuple[str, ...]


@dataclass(frozen=True)
class LiteralOperand:
    """A string, number or boolean literal."""

    value: str | int | float | bool


@dataclass(frozen=True)
class NilOperand:
    """The literal `nil`."""


@dataclass(frozen=True)
class NameOperand:
    """A function name."""

    name: str


@dataclass(frozen=True)
class PipelineOperand:
    """A parenthesised pipeline."""

    pipeline: Pipeline


Operand = FieldOperand | LiteralOperand | NilOperand | NameOperand | PipelineOperand
"""One operand of a command: a field of the data, a literal, a parenthesised pipeline or a function name."""

Command = list[Operand]
"""Operands in a row: a function call (`eq .Kind "app"`) or a single value."""

Pipeline = list[Command]
"""Commands joined by `|`, each receiving the previous one's value as its last argument."""


@dataclass(frozen=True)
class TextNode:
    """Literal text."""

    text: str


@dataclass(frozen=True)
class ActionNode:
    """`[[ pipeline ]]`: a value printed."""

    pipeline: Pipeline


@dataclass(frozen=True)
class BlockNode:
    """`if`, `with` or `range`, with the nodes of its body and of its `else`."""

    type: Literal["if", "with", "range"]
    pipeline: Pipeline
    body: list[TemplateNode] = field(default_factory=list)
    otherwise: list[TemplateNode] = field(default_factory=list)


TemplateNode = TextNode | ActionNode | BlockNode
"""A node of a parsed template."""
