"""Turns a schema description into the text of a Python docstring."""

from __future__ import annotations

WIDTH = 100


def escape_docstring(text: str) -> str:
    """Make text safe inside a triple-quoted string: no backslash escapes, no early end."""
    escaped = text.replace("\\", "\\\\").replace('"""', '\\"\\"\\"')
    return escaped[:-1] + '\\"' if escaped.endswith('"') else escaped


def wrap_words(text: str, width: int) -> list[str]:
    """Greedy word wrap; a word longer than the width stays whole on its own line."""
    lines: list[str] = []
    line = ""
    for word in text.split(" "):
        if line and len(line) + 1 + len(word) > width:
            lines.append(line)
            line = word
        else:
            line = word if not line else f"{line} {word}"
    if line:
        lines.append(line)
    return lines


def docstring_lines(text: str | None, indent: str) -> list[str]:
    """The description as wrapped, escaped lines (paragraphs kept, a blank line between them).

    Returns an empty list for no text. The lines carry no indentation of their own: the
    caller adds the indentation and the quotes.
    """
    if text is None or text.strip() == "":
        return []
    width = WIDTH - len(indent) - 3
    lines: list[str] = []
    for i, paragraph in enumerate(escape_docstring(text).split("\n\n")):
        if i > 0:
            lines.append("")
        lines.extend(wrap_words(paragraph.replace("\n", " "), width))
    return lines


def docstring(text: str | None, indent: str = "") -> str:
    """A docstring statement at an indentation, or an empty string for no text.

    One line when it fits, otherwise the quotes on their own lines.
    """
    lines = docstring_lines(text, indent)
    if not lines:
        return ""
    if len(lines) == 1 and len(indent) + len(lines[0]) + 6 <= WIDTH:
        return f'{indent}"""{lines[0]}"""\n'
    body = "".join(f"{indent}{line}\n" if line else "\n" for line in lines)
    return f'{indent}"""\n{body}{indent}"""\n'
